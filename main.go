package main

import (
	"context"
	"embed"
	"fmt"
	"stock-tracker/config"
	"stock-tracker/database"
	"stock-tracker/services"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "StockTracker",
		Width:     1200,
		Height:    700,
		MinWidth:  900,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 250, G: 250, B: 252, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// App application struct
type App struct {
	ctx            context.Context
	config         *config.Config
	db             *database.DB
	stockService   *services.StockService
	marketService  *services.MarketService
	refreshManager *services.RefreshManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg := config.Load()
	a.config = cfg

	db, err := database.Init(cfg.DBPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	a.db = db

	a.stockService = services.NewStockService(db)
	a.marketService = services.NewMarketService(cfg)
	a.stockService.SetMarketService(a.marketService)
	a.refreshManager = services.NewRefreshManager(cfg, a.marketService, db)
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}
