package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

// StockStatus 表示股票状态
type StockStatus string

const (
	StockStatusNormal    StockStatus = "normal"
	StockStatusSuspended StockStatus = "suspended"
	StockStatusIPO       StockStatus = "ipo"
	StockStatusExRights  StockStatus = "ex-rights"
)

// Scan implements sql.Scanner for StockStatus
func (s *StockStatus) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		*s = ""
	case string:
		*s = StockStatus(v)
	case []byte:
		*s = StockStatus(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into StockStatus", src)
	}
	return nil
}

// Value implements driver.Valuer for StockStatus
func (s StockStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Stock 表示当前持仓表的一条记录
// 所有价格字段使用整数分（100 = 1.00元），涨跌幅使用整数 BP（100 = 1.00%）
type Stock struct {
	ID             int64          `json:"id"`
	Code           string         `json:"code"`           // 股票代码，如 "sh600519"
	Name           string         `json:"name"`           // 股票名称
	EntryDate      string         `json:"entryDate"`     // 选入日期，格式 "YYYY-MM-DD"
	EntryTime      string         `json:"entryTime"`     // 选入时间，格式 "HH:MM:SS"
	EntryPrice     int64          `json:"entryPrice"`    // 选入价（前复权，分）
	RawPrice       int64          `json:"rawPrice"`      // 选入日实际收盘价（不复权，分）
	AdjustFactor   int64          `json:"adjustFactor"`  // 复权因子（BP，1000=1.0）
	CurrentPrice   int64          `json:"currentPrice"`  // 当前价（分）
	PrevClose      int64          `json:"prevClose"`     // 昨收价（分）
	DailyChange    int64          `json:"dailyChange"`   // 今日涨跌幅（BP）
	AccChange      int64          `json:"accChange"`     // 累计涨跌幅（BP，基于前复权）
	DataSource     string         `json:"dataSource"`    // 当前使用的数据源标识
	Status         StockStatus    `json:"status"`         // 状态：normal/suspended/ipo/ex-rights
	LastUpdate     sql.NullString `json:"lastUpdate"`    // 最后更新时间，格式 "YYYY-MM-DD HH:MM:SS"
	CreatedAt      string         `json:"createdAt"`     // 创建时间，格式 "YYYY-MM-DD HH:MM:SS"
	UpdatedAt      string         `json:"updatedAt"`     // 更新时间，格式 "YYYY-MM-DD HH:MM:SS"
}

// HistoryRecord 表示已调出股票的历史记录
type HistoryRecord struct {
	ID           int64  `json:"id"`
	Code         string `json:"code"`          // 股票代码
	Name         string `json:"name"`          // 股票名称（调出时的名称）
	EntryDate    string `json:"entryDate"`    // 选入日期
	EntryPrice   int64  `json:"entryPrice"`   // 选入价（前复权，分）
	ExitDate     string `json:"exitDate"`     // 调出日期
	ExitPrice    int64  `json:"exitPrice"`    // 调出价（分）
	HoldingDays  int64  `json:"holdingDays"`  // 持股天数
	TotalReturn  int64  `json:"totalReturn"`  // 区间收益（BP，基于前复权）
	DataSource   string `json:"dataSource"`   // 数据源标识
	CreatedAt    string `json:"createdAt"`    // 记录创建时间
}

// DailySnapshot 表示每日收盘价快照（用于趋势分析）
type DailySnapshot struct {
	ID       int64 `json:"id"`
	StockID  int64 `json:"stockId"`  // 关联 stocks.id
	Date     string `json:"date"`      // 日期，格式 "YYYY-MM-DD"
	Price    int64  `json:"price"`     // 收盘价（分）
	ChangeBP int64  `json:"changeBP"` // 当日涨跌幅（BP）
}

// ToPriceYuan 将整数分转换为元（float64，展示用）
func ToPriceYuan(cents int64) float64 {
	return float64(cents) / 100.0
}

// ToPriceCents 将元转换为整数分（存储用）
func ToPriceCents(yuan float64) int64 {
	return int64(yuan * 100 + 0.5)
}

// ToPercent 将整数 BP 转换为百分比（float64，展示用）
func ToPercent(bp int64) float64 {
	return float64(bp) / 100.0
}

// ToBP 将百分比转换为整数 BP（存储用）
func ToBP(pct float64) int64 {
	return int64(pct*100 + 0.5)
}

// FormatPrice 将分格式化为展示字符串（如 172500 → "1725.00"）
func FormatPrice(cents int64) string {
	if cents < 0 {
		return "-" + FormatPrice(-cents)
	}
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// FormatPercent 将 BP 格式化为百分比字符串（如 148 → "+1.48%"）
func FormatPercent(bp int64) string {
	sign := ""
	if bp > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.2f%%", sign, ToPercent(bp))
}

// CalculateHoldingDays 计算持股天数（自然日）
func CalculateHoldingDays(entryDate, exitDate string) (int64, error) {
	entry, err := time.Parse("2006-01-02", entryDate)
	if err != nil {
		return 0, fmt.Errorf("parse entry_date: %w", err)
	}
	exit, err := time.Parse("2006-01-02", exitDate)
	if err != nil {
		return 0, fmt.Errorf("parse exit_date: %w", err)
	}
	return int64(exit.Sub(entry).Hours() / 24), nil
}

// CalculateTotalReturn 计算区间收益（BP，基于前复权）
func CalculateTotalReturn(entryPrice, exitPrice, adjustFactor int64) int64 {
	// entry_price_adjusted = entryPrice * adjustFactor / 1000
	adjustedEntry := entryPrice * adjustFactor / 1000
	if adjustedEntry == 0 {
		return 0
	}
	return (exitPrice - adjustedEntry) * 10000 / adjustedEntry
}

// CalculateAccChange 计算累计涨跌幅（BP）
func CalculateAccChange(currentPrice, entryPrice, adjustFactor int64) int64 {
	return CalculateTotalReturn(entryPrice, currentPrice, adjustFactor)
}

// CalculateDailyChange 计算日涨跌幅（BP）
func CalculateDailyChange(currentPrice, prevClose int64) int64 {
	if prevClose == 0 {
		return 0
	}
	return (currentPrice - prevClose) * 10000 / prevClose
}
