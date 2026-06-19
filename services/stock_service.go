package services

import (
	"context"
	"database/sql"
	"fmt"
	"stock-tracker/database"
	"stock-tracker/providers"
	"strings"
	"time"
)

// StockService 负责股票管理业务：选入、调出、查询
type StockService struct {
	db            *database.DB
	stockRepo     *database.StockRepository
	historyRepo   *database.HistoryRepository
	marketService *MarketService
}

// NewStockService 创建 StockService 实例
func NewStockService(db *database.DB) *StockService {
	return &StockService{
		db:          db,
		stockRepo:   database.NewStockRepository(db),
		historyRepo: database.NewHistoryRepository(db),
	}
}

// SetMarketService 设置 MarketService（用于行情获取）
// 在 main.go 的 startup 中调用，解决循环依赖
func (ss *StockService) SetMarketService(ms *MarketService) {
	ss.marketService = ms
}

// GetAll 返回所有当前持仓
func (ss *StockService) GetAll() ([]database.Stock, error) {
	return ss.stockRepo.GetAll()
}

// GetHistory 返回所有历史记录
func (ss *StockService) GetHistory() ([]database.HistoryRecord, error) {
	return ss.historyRepo.GetAll()
}

// Add 选入股票/基金
// 1. 判断是否为场外基金（纯6位数字，不带前缀）
// 2. 标准化代码
// 3. 检查是否已存在
// 4. 获取实时行情（场外基金用净值接口，场内用行情接口）
// 5. 计算复权因子（场外基金不需要）
// 6. 写入 stocks 表
// 7. 返回新建的记录
func (ss *StockService) Add(code string) (*database.Stock, error) {
	if ss.marketService == nil {
		return nil, fmt.Errorf("market service not initialized")
	}

	// 判断是否为场外基金
	isOTCFund := isOTCFundCode(code)

	// 1. 标准化代码
	var normalizedCode string
	if isOTCFund {
		// 场外基金：直接使用原始代码，不带前缀
		normalizedCode = strings.TrimSpace(code)
	} else {
		normalizedCode = normalizeCode(code)
	}
	if normalizedCode == "" {
		return nil, fmt.Errorf("无效的股票代码: %s", code)
	}

	// 2. 检查是否已存在
	existing, err := ss.stockRepo.GetByCode(normalizedCode)
	if err != nil {
		return nil, fmt.Errorf("查询股票失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("该股票已在持仓中: %s", normalizedCode)
	}

	ctx := context.Background()
	now := time.Now()
	entryDate := now.Format("2006-01-02")
	var entryPrice int64
	var rawPrice int64
	var adjustFactor int64 = 1000
	var quote *providers.Quote

	if isOTCFund {
		// 场外基金：获取最新净值
		q, err := ss.marketService.FetchOTCFundQuote(ctx, normalizedCode)
		if err != nil {
			return nil, fmt.Errorf("获取场外基金净值失败: %w", err)
		}
		quote = q
		entryPrice = database.ToPriceCents(q.Price)
		rawPrice = entryPrice
		adjustFactor = 1000
	} else {
		// 场内股票/基金：经容灾链获取实时行情
		q, err := ss.marketService.FetchQuoteWithPreferred(ctx, normalizedCode, "")
		if err != nil {
			return nil, fmt.Errorf("获取行情失败: %w", err)
		}
		quote = q

		// 获取前复权和不复权 K 线数据
		adjKlines, err := ss.marketService.FetchKlineData(ctx, normalizedCode)
		if err == nil {
			for _, k := range adjKlines {
				if k.Date == entryDate {
					entryPrice = k.Close
					break
				}
			}
		}

		rawKlines, err := ss.marketService.FetchRawKlineData(ctx, normalizedCode)
		if err == nil {
			for _, k := range rawKlines {
				if k.Date == entryDate {
					rawPrice = k.Close
					break
				}
			}
		}

		// 计算复权因子
		if entryPrice > 0 && rawPrice > 0 {
			adjustFactor = rawPrice * 1000 / entryPrice
		}

		// 如果 K 线数据不可用，使用当前价作为降级
		currentPrice := database.ToPriceCents(quote.Price)
		if entryPrice == 0 {
			entryPrice = currentPrice
		}
		if rawPrice == 0 {
			rawPrice = currentPrice
		}
	}

	// 场外基金名称补全：净值接口不返回名称，从搜索接口获取
	if isOTCFund && quote.Name == normalizedCode {
		searchResults, _ := ss.marketService.SearchStocks(ctx, normalizedCode)
		for _, r := range searchResults {
			if r.Code == normalizedCode && r.Type == "otc_fund" && r.Name != "" {
				quote.Name = r.Name
				break
			}
		}
	}

	// 计算日涨跌幅和累计涨跌幅
	currentPrice := database.ToPriceCents(quote.Price)
	prevClose := database.ToPriceCents(quote.PrevClose)
	dailyChange := database.CalculateDailyChange(currentPrice, prevClose)
	accChange := database.CalculateAccChange(currentPrice, entryPrice, adjustFactor)

	// 停牌判断（场外基金不判断）
	status := database.StockStatusNormal
	if !isOTCFund && quote.IsSuspended {
		status = database.StockStatusSuspended
	}

	timeStr := now.Format("15:04:05")
	datetimeStr := now.Format("2006-01-02 15:04:05")

	// 写入数据库
	stock := &database.Stock{
		Code:         normalizedCode,
		Name:         quote.Name,
		EntryDate:    entryDate,
		EntryTime:    timeStr,
		EntryPrice:   entryPrice,
		RawPrice:     rawPrice,
		AdjustFactor: adjustFactor,
		CurrentPrice: currentPrice,
		PrevClose:    prevClose,
		DailyChange:  dailyChange,
		AccChange:    accChange,
		DataSource:   quote.Source,
		Status:       status,
		LastUpdate:   sql.NullString{String: datetimeStr, Valid: true},
		CreatedAt:    datetimeStr,
		UpdatedAt:    datetimeStr,
	}

	created, err := ss.stockRepo.Create(stock)
	if err != nil {
		return nil, fmt.Errorf("写入股票失败: %w", err)
	}

	return created, nil
}

// isOTCFundCode 判断代码是否为场外基金格式（纯6位数字，不带 sh/sz 前缀）
func isOTCFundCode(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Remove 调出股票
// 1. 获取 stock 记录
// 2. 经容灾链获取当前价
// 3. 计算持股天数和区间涨跌幅
// 4. 从 stocks 删除，写入 history 表
// 5. 返回 history 记录
func (ss *StockService) Remove(id int64) (*database.HistoryRecord, error) {
	if ss.marketService == nil {
		return nil, fmt.Errorf("market service not initialized")
	}

	// 1. 获取 stock 记录
	stock, err := ss.stockRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查询股票失败: %w", err)
	}
	if stock == nil {
		return nil, fmt.Errorf("股票不存在: %d", id)
	}

	ctx := context.Background()

	// 2. 经容灾链获取当前价（优先使用同一数据源）
	quote, err := ss.marketService.FetchQuoteWithPreferred(ctx, stock.Code, stock.DataSource)
	if err != nil {
		// 降级：使用数据库中的当前价
		quote = &providers.Quote{
			Price:   database.ToPriceYuan(stock.CurrentPrice),
			Source:  stock.DataSource,
		}
	}

	// 3. 计算持股天数和区间涨跌幅
	now := time.Now()
	exitDate := now.Format("2006-01-02")
	holdingDays := CalculateHoldingDays(stock.EntryDate, exitDate)
	exitPrice := database.ToPriceCents(quote.Price)
	totalReturn := CalculateAccChange(exitPrice, stock.EntryPrice, stock.AdjustFactor)

	datetimeStr := now.Format("2006-01-02 15:04:05")

	// 4. 写入 history 表
	history := &database.HistoryRecord{
		Code:        stock.Code,
		Name:        stock.Name,
		EntryDate:   stock.EntryDate,
		EntryPrice:  stock.EntryPrice,
		ExitDate:    exitDate,
		ExitPrice:   exitPrice,
		HoldingDays: holdingDays,
		TotalReturn: totalReturn,
		DataSource:  quote.Source,
		CreatedAt:   datetimeStr,
	}

	createdHistory, err := ss.historyRepo.Create(history)
	if err != nil {
		return nil, fmt.Errorf("写入历史记录失败: %w", err)
	}

	// 5. 从 stocks 删除
	err = ss.stockRepo.Delete(id)
	if err != nil {
		return nil, fmt.Errorf("删除股票失败: %w", err)
	}

	return createdHistory, nil
}

// ClearHistory 清空所有历史记录
func (ss *StockService) ClearHistory() error {
	if ss.historyRepo == nil {
		return fmt.Errorf("history repository not initialized")
	}
	return ss.historyRepo.ClearAll()
}

// normalizeCode 标准化股票代码，自动补全 sh/sz 前缀
// 支持：600519 → sh600519, 000001 → sz000001, sh600519 → sh600519, 300001.SZ → sz300001
func normalizeCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.ToLower(code)

	// 去除可能的后缀，如 .SH, .SZ, .ss, .zs
	code = strings.TrimSuffix(code, ".sh")
	code = strings.TrimSuffix(code, ".sz")
	code = strings.TrimSuffix(code, ".ss")
	code = strings.TrimSuffix(code, ".zs")

	// 如果已有前缀，直接返回
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") {
		return code
	}

	// 纯数字，根据首位判断市场
	if len(code) == 6 {
		firstChar := code[0]
		// 6 开头为上海，0/3 开头为深圳
		if firstChar == '6' || firstChar == '5' || firstChar == '9' {
			return "sh" + code
		}
		return "sz" + code
	}

	// 其他情况，原样返回
	return code
}
