package services

import "time"

// MarketStatus 表示市场状态
type MarketStatus struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	IsTrading bool   `json:"isTrading"`
}

// TradingCalendar 处理 A 股交易日判断
type TradingCalendar struct {
	loc *time.Location
}

// NewTradingCalendar 创建交易日历实例
func NewTradingCalendar() *TradingCalendar {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载时区失败，回退到固定偏移 +8
		loc = time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return &TradingCalendar{loc: loc}
}

// GetMarketStatus 返回当前市场状态
// 简化版：按周一到五判断，暂不考虑节假日
func (tc *TradingCalendar) GetMarketStatus() *MarketStatus {
	now := time.Now().In(tc.loc)
	weekday := now.Weekday()
	hour := now.Hour()
	minute := now.Minute()

	// 周末休市
	if weekday == time.Saturday || weekday == time.Sunday {
		return &MarketStatus{
			Status:    "closed",
			Message:   "周末休市",
			IsTrading: false,
		}
	}

	// 将时间转换为分钟数，便于判断
	timeVal := hour*60 + minute

	switch {
	case timeVal >= 9*60+15 && timeVal < 9*60+25:
		// 9:15-9:25 集合竞价
		return &MarketStatus{
			Status:    "auction",
			Message:   "开盘集合竞价",
			IsTrading: false,
		}
	case timeVal >= 9*60+30 && timeVal < 11*60+30:
		// 9:30-11:30 上午交易
		return &MarketStatus{
			Status:    "trading",
			Message:   "上午交易中",
			IsTrading: true,
		}
	case timeVal >= 11*60+30 && timeVal < 13*60+0:
		// 11:30-13:00 午休
		return &MarketStatus{
			Status:    "lunch",
			Message:   "午休时间",
			IsTrading: false,
		}
	case timeVal >= 13*60+0 && timeVal < 14*60+57:
		// 13:00-14:57 下午交易
		return &MarketStatus{
			Status:    "trading",
			Message:   "下午交易中",
			IsTrading: true,
		}
	case timeVal >= 14*60+57 && timeVal < 15*60+0:
		// 14:57-15:00 收盘集合
		return &MarketStatus{
			Status:    "closing",
			Message:   "收盘集合竞价",
			IsTrading: true,
		}
	default:
		// 其他时间（盘前、盘后）
		return &MarketStatus{
			Status:    "closed",
			Message:   "非交易时段",
			IsTrading: false,
		}
	}
}
