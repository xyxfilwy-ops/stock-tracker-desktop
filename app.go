package main

import (
	"fmt"
	"stock-tracker/config"
	"stock-tracker/database"
	"stock-tracker/services"
)

// GetStocks returns all current holdings
func (a *App) GetStocks() ([]database.Stock, error) {
	if a.stockService == nil {
		return nil, fmt.Errorf("stock service not initialized")
	}
	return a.stockService.GetAll()
}

// GetHistory returns all historical records
func (a *App) GetHistory() ([]database.HistoryRecord, error) {
	if a.stockService == nil {
		return nil, fmt.Errorf("stock service not initialized")
	}
	return a.stockService.GetHistory()
}

// AddStock adds a new stock to tracking
func (a *App) AddStock(code string) (*database.Stock, error) {
	if a.stockService == nil {
		return nil, fmt.Errorf("stock service not initialized")
	}
	return a.stockService.Add(code)
}

// RemoveStock removes a stock and moves it to history
func (a *App) RemoveStock(id int64) (*database.HistoryRecord, error) {
	if a.stockService == nil {
		return nil, fmt.Errorf("stock service not initialized")
	}
	return a.stockService.Remove(id)
}

// RefreshAll refreshes all stock quotes
func (a *App) RefreshAll() (*services.RefreshResult, error) {
	if a.refreshManager == nil {
		return nil, fmt.Errorf("refresh manager not initialized")
	}
	return a.refreshManager.RefreshAll(a.ctx)
}

// RefreshStock refreshes a single stock quote
func (a *App) RefreshStock(id int64) (*database.Stock, error) {
	if a.refreshManager == nil {
		return nil, fmt.Errorf("refresh manager not initialized")
	}
	return a.refreshManager.RefreshOne(a.ctx, id)
}

// GetMarketStatus returns current market status info
func (a *App) GetMarketStatus() *services.MarketStatus {
	if a.refreshManager == nil {
		return nil
	}
	return a.refreshManager.GetMarketStatus()
}

// ClearHistory clears all historical records
func (a *App) ClearHistory() error {
	if a.stockService == nil {
		return fmt.Errorf("stock service not initialized")
	}
	return a.stockService.ClearHistory()
}
func (a *App) GetConfig() *config.Config {
	return a.config
}
