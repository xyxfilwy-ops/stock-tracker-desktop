package services

import (
	"fmt"
	"os"
	"testing"

	"stock-tracker/config"
	"stock-tracker/database"
)

func TestAddOTCFund(t *testing.T) {
	// 使用临时文件作为测试数据库
	tmpFile, err := os.CreateTemp("", "stock_test_*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// 初始化数据库
	db, err := database.Init(tmpFile.Name())
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer db.Close()

	// 创建 StockService
	ss := NewStockService(db)

	// 创建 MarketService（需要配置）
	cfg := config.Load()
	ms := NewMarketService(cfg)
	ss.SetMarketService(ms)

	// 调用 Add 选入场外基金
	stock, err := ss.Add("000001")
	if err != nil {
		t.Fatalf("Add OTC fund failed: %v", err)
	}

	// 验证返回结果
	if stock == nil {
		t.Fatal("expected non-nil stock")
	}
	if stock.Code != "000001" {
		t.Fatalf("expected code 000001, got %s", stock.Code)
	}
	// 场外基金净值接口不返回名称，Name 默认为代码
	if stock.Name != "000001" && stock.Name != "华夏成长混合" {
		t.Fatalf("expected Name to be 000001 or 华夏成长混合, got %s", stock.Name)
	}
	if stock.EntryPrice <= 0 {
		t.Fatalf("expected positive entry price, got %d", stock.EntryPrice)
	}

	fmt.Printf("✓ Add OTC fund: Code=%s, Name=%s, EntryPrice=%d分, DataSource=%s\n",
		stock.Code, stock.Name, stock.EntryPrice, stock.DataSource)

	// 验证数据库中确实有一条记录
	all, err := ss.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 stock, got %d", len(all))
	}
	if all[0].Code != "000001" {
		t.Fatalf("expected db code 000001, got %s", all[0].Code)
	}

	fmt.Printf("✓ Database verified: %d stock(s), Code=%s\n", len(all), all[0].Code)
}
