package main

import (
	"stock-tracker/config"
	"stock-tracker/database"
	"stock-tracker/services"
)

// GetStocks returns all current holdings
func (a *App) GetStocks() ([]database.Stock, error) {
	return a.stockService.GetAll()
}

// GetHistory returns all historical records
func (a *App) GetHistory() ([]database.HistoryRecord, error) {
	return a.stockService.GetHistory()
}

// AddStock adds a new stock to tracking
func (a *App) AddStock(code string) (*database.Stock, error) {
	return a.stockService.Add(code)
}

// RemoveStock removes a stock and moves it to history
func (a *App) RemoveStock(id int64) (*database.HistoryRecord, error) {
	return a.stockService.Remove(id)
}

// RefreshAll refreshes all stock quotes
func (a *App) RefreshAll() (*services.RefreshResult, error) {
	return a.refreshManager.RefreshAll(a.ctx)
}

// RefreshStock refreshes a single stock quote
func (a *App) RefreshStock(id int64) (*database.Stock, error) {
	return a.refreshManager.RefreshOne(a.ctx, id)
}

// GetMarketStatus returns current market status info
func (a *App) GetMarketStatus() *services.MarketStatus {
	return a.refreshManager.GetMarketStatus()
}

// GetConfig returns app configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}
