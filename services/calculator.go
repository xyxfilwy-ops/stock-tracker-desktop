package services

import "stock-tracker/database"

// CalculateDailyChange 计算今日涨跌幅（BP）
// currentPrice: 当前价（分）, prevClose: 昨收价（分）
func CalculateDailyChange(currentPrice, prevClose int64) int64 {
	return database.CalculateDailyChange(currentPrice, prevClose)
}

// CalculateAccChange 计算累计涨跌幅（BP）
// currentPrice: 当前价（分）, entryPrice: 选入价（分）, adjustFactor: 复权因子（1000=1.0）
// 前复权计算：entry_price_adjusted = entryPrice * adjustFactor / 1000
func CalculateAccChange(currentPrice, entryPrice, adjustFactor int64) int64 {
	return database.CalculateAccChange(currentPrice, entryPrice, adjustFactor)
}

// CalculateHoldingDays 计算持股天数（自然日）
// entryDate, exitDate 格式: "YYYY-MM-DD"
func CalculateHoldingDays(entryDate, exitDate string) int64 {
	days, _ := database.CalculateHoldingDays(entryDate, exitDate)
	return days
}

// FormatPrice 将分转换为元字符串，保留2位小数
// 例如: 172500 → "1725.00"
func FormatPrice(cents int64) string {
	return database.FormatPrice(cents)
}

// FormatBP 将 BP 转换为百分比字符串
// 例如: 148 → "+1.48%", -50 → "-0.50%"
func FormatBP(bp int64) string {
	return database.FormatPercent(bp)
}
