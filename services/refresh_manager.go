package services

import (
	"context"
	"fmt"
	"net/http"
	"stock-tracker/config"
	"stock-tracker/database"
	"sync"
	"time"

	"github.com/sony/gobreaker"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RefreshDetail 单只股票的刷新结果
type RefreshDetail struct {
	ID    int64  `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

// RefreshResult 批量刷新结果
type RefreshResult struct {
	Updated int             `json:"updated"`
	Failed  int             `json:"failed"`
	Total   int             `json:"total"`
	Details []RefreshDetail `json:"details"`
}

// RefreshManager 负责批量刷新持仓股票行情
type RefreshManager struct {
	cfg           *config.Config
	marketService *MarketService
	db            *database.DB
	stockRepo     *database.StockRepository
	historyRepo   *database.HistoryRepository
	sem           chan struct{} // 信号量，并发度 5
	client        *http.Client
	circuit       map[string]*gobreaker.CircuitBreaker // 预留字段，实际熔断器在 marketService 中管理
}

// NewRefreshManager 创建 RefreshManager 实例
func NewRefreshManager(cfg *config.Config, marketService *MarketService, db *database.DB) *RefreshManager {
	return &RefreshManager{
		cfg:           cfg,
		marketService: marketService,
		db:            db,
		stockRepo:     database.NewStockRepository(db),
		historyRepo:   database.NewHistoryRepository(db),
		sem:           make(chan struct{}, cfg.RefreshConcurrency),
		client:        &http.Client{Timeout: cfg.HTTPTimeout},
		circuit:       make(map[string]*gobreaker.CircuitBreaker),
	}
}

// RefreshAll 批量刷新所有持仓
// 检测是否交易时段，非交易时段返回提示
// 使用 semaphore + WaitGroup 并发刷新，并发度由配置控制
func (rm *RefreshManager) RefreshAll(ctx context.Context) (*RefreshResult, error) {
	// 检测是否交易时段
	status := rm.GetMarketStatus()
	if !status.IsTrading {
		return nil, fmt.Errorf("当前非交易时段：%s", status.Message)
	}

	stocks, err := rm.stockRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("读取持仓失败: %w", err)
	}

	result := &RefreshResult{
		Total:   len(stocks),
		Details: make([]RefreshDetail, 0, len(stocks)),
	}

	if len(stocks) == 0 {
		return result, nil
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, s := range stocks {
		wg.Add(1)
		go func(stock database.Stock) {
			defer wg.Done()

			// 信号量控制并发
			select {
			case rm.sem <- struct{}{}:
				defer func() { <-rm.sem }()
			case <-ctx.Done():
				mu.Lock()
				result.Failed++
				result.Details = append(result.Details, RefreshDetail{
					ID:    stock.ID,
					Code:  stock.Code,
					Name:  stock.Name,
					Error: ctx.Err().Error(),
				})
				mu.Unlock()
				return
			}

			// 刷新单只股票
			updated, err := rm.RefreshOne(ctx, stock.ID)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.Failed++
				result.Details = append(result.Details, RefreshDetail{
					ID:    stock.ID,
					Code:  stock.Code,
					Name:  stock.Name,
					Error: err.Error(),
				})
				runtime.EventsEmit(ctx, "refresh:progress", map[string]interface{}{
					"current": result.Updated + result.Failed,
					"total":   result.Total,
					"code":    stock.Code,
					"success": false,
				})
				return
			}

			result.Updated++
			result.Details = append(result.Details, RefreshDetail{
				ID:   updated.ID,
				Code: updated.Code,
				Name: updated.Name,
			})
			runtime.EventsEmit(ctx, "refresh:progress", map[string]interface{}{
				"current": result.Updated + result.Failed,
				"total":   result.Total,
				"code":    stock.Code,
				"success": true,
			})
		}(s)
	}

	wg.Wait()

	// 错误分层处理：全部失败时返回 error，部分失败不返回 error
	if result.Failed == result.Total && result.Total > 0 {
		// 全部失败：banner-level 错误
		return result, fmt.Errorf("全部刷新失败")
	}

	return result, nil
}

// RefreshOne 刷新单只股票
// 经容灾链获取行情，数据归一化后更新 stocks 表
func (rm *RefreshManager) RefreshOne(ctx context.Context, id int64) (*database.Stock, error) {
	stock, err := rm.stockRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查询股票失败: %w", err)
	}
	if stock == nil {
		return nil, fmt.Errorf("股票不存在: %d", id)
	}

	// 经容灾链获取行情，优先使用 stocks 表记录的数据源
	quote, err := rm.marketService.FetchQuoteWithPreferred(ctx, stock.Code, stock.DataSource)
	if err != nil {
		return nil, fmt.Errorf("获取行情失败: %w", err)
	}

	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05")

	// 价格转换为 int64 分
	currentPrice := database.ToPriceCents(quote.Price)
	prevClose := database.ToPriceCents(quote.PrevClose)

	// 计算日涨跌幅：优先取 API 直供的 ChangePct，否则自行计算
	var dailyChange int64
	if !quote.IsSuspended {
		if quote.ChangePct != 0 {
			dailyChange = database.ToBP(quote.ChangePct)
		} else {
			dailyChange = database.CalculateDailyChange(currentPrice, prevClose)
		}
	}

	// 计算累计涨跌幅（基于前复权）
	accChange := database.CalculateAccChange(currentPrice, stock.EntryPrice, stock.AdjustFactor)

	// 更新名称（如果 API 返回的名称与数据库不一致）
	if quote.Name != "" && quote.Name != stock.Name {
		_ = rm.stockRepo.UpdateName(stock.ID, quote.Name)
	}

	// 状态判断
	status := database.StockStatusNormal
	if quote.IsSuspended {
		status = database.StockStatusSuspended
	}

	// 更新数据库（含数据源信息，实现同一源锁定）
	err = rm.stockRepo.UpdateAfterRefresh(
		stock.ID,
		currentPrice,
		prevClose,
		dailyChange,
		accChange,
		quote.Source,
		string(status),
		nowStr,
		nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("更新数据库失败: %w", err)
	}

	// 重新读取以返回最新数据
	return rm.stockRepo.GetByID(id)
}

// GetMarketStatus 返回当前市场状态
func (rm *RefreshManager) GetMarketStatus() *MarketStatus {
	tc := NewTradingCalendar()
	return tc.GetMarketStatus()
}
