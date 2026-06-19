package services

import (
	"context"
	"fmt"
	"testing"

	"stock-tracker/config"
)

func TestSearchStocks(t *testing.T) {
	cfg := config.Load()
	ms := NewMarketService(cfg)

	// 测试按名称搜索
	results, err := ms.SearchStocks(context.Background(), "贵州茅台")
	if err != nil {
		t.Fatalf("search by name failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for 贵州茅台")
	}
	if results[0].Code != "sh600519" {
		t.Fatalf("expected sh600519, got %s", results[0].Code)
	}
	if results[0].Name != "贵州茅台" {
		t.Fatalf("expected 贵州茅台, got %s", results[0].Name)
	}
	if results[0].Type != "stock" {
		t.Fatalf("expected stock type, got %s", results[0].Type)
	}
	fmt.Printf("✓ 名称搜索: %s (%s)\n", results[0].Name, results[0].Code)

	// 测试按首字母搜索
	results2, err := ms.SearchStocks(context.Background(), "gzmt")
	if err != nil {
		t.Fatalf("search by pinyin failed: %v", err)
	}
	if len(results2) == 0 {
		t.Fatal("expected results for gzmt")
	}
	found := false
	for _, r := range results2 {
		if r.Code == "sh600519" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected sh600519 in pinyin results, got %+v", results2)
	}
	fmt.Printf("✓ 首字母搜索: gzmt -> sh600519\n")

	// 测试场内基金搜索
	results3, err := ms.SearchStocks(context.Background(), "510300")
	if err != nil {
		t.Fatalf("search fund by code failed: %v", err)
	}
	if len(results3) == 0 {
		t.Fatal("expected results for 510300")
	}
	if results3[0].Type != "fund" {
		t.Fatalf("expected fund type, got %s", results3[0].Type)
	}
	if results3[0].Code != "sh510300" {
		t.Fatalf("expected sh510300, got %s", results3[0].Code)
	}
	fmt.Printf("✓ 场内基金搜索: %s (%s, %s)\n", results3[0].Name, results3[0].Code, results3[0].Type)

	// 测试深圳基金
	results4, err := ms.SearchStocks(context.Background(), "159915")
	if err != nil {
		t.Fatalf("search fund by code failed: %v", err)
	}
	if len(results4) == 0 {
		t.Fatal("expected results for 159915")
	}
	if results4[0].Code != "sz159915" {
		t.Fatalf("expected sz159915, got %s", results4[0].Code)
	}
	fmt.Printf("✓ 深圳基金搜索: %s (%s)\n", results4[0].Name, results4[0].Code)

	fmt.Println("\n所有搜索测试通过！")
}

func TestSearch000001(t *testing.T) {
	cfg := config.Load()
	ms := NewMarketService(cfg)

	// 搜索 "000001"，应该同时返回平安银行（AStock）和华夏成长混合（OTCFUND）
	results, err := ms.SearchStocks(context.Background(), "000001")
	if err != nil {
		t.Fatalf("search 000001 failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for 000001")
	}

	foundOTC := false
	foundStock := false
	var otcResult, stockResult SearchResult

	for _, r := range results {
		if r.Type == "otc_fund" {
			foundOTC = true
			otcResult = r
		}
		if r.Type == "stock" {
			foundStock = true
			stockResult = r
		}
	}

	if !foundOTC {
		t.Fatalf("expected otc_fund in results, got %+v", results)
	}
	if !foundStock {
		t.Logf("warning: no stock type found in results, got %+v", results)
	}

	if otcResult.Code != "000001" {
		t.Fatalf("expected otc_fund code 000001, got %s", otcResult.Code)
	}
	if otcResult.Name != "华夏成长混合" {
		t.Fatalf("expected 华夏成长混合, got %s", otcResult.Name)
	}

	fmt.Printf("✓ 搜索 000001: found otc_fund=%s (%s)", otcResult.Name, otcResult.Code)
	if foundStock {
		fmt.Printf(", found stock=%s (%s)", stockResult.Name, stockResult.Code)
	}
	fmt.Println()
}

func TestSearchOTCFund(t *testing.T) {
	cfg := config.Load()
	ms := NewMarketService(cfg)

	// 测试按名称搜索场外基金
	results, err := ms.SearchStocks(context.Background(), "华夏成长混合")
	if err != nil {
		t.Fatalf("search otc fund by name failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for 华夏成长混合")
	}

	found := false
	for _, r := range results {
		if r.Code == "000001" && r.Type == "otc_fund" {
			found = true
			fmt.Printf("✓ 场外基金名称搜索: %s (%s, %s)\n", r.Name, r.Code, r.Type)
			break
		}
	}
	if !found {
		t.Fatalf("expected otc_fund 000001 in results, got %+v", results)
	}

	// 测试按首字母搜索场外基金
	results2, err := ms.SearchStocks(context.Background(), "hxczhh")
	if err != nil {
		t.Fatalf("search otc fund by pinyin failed: %v", err)
	}
	found2 := false
	for _, r := range results2 {
		if r.Code == "000001" && r.Type == "otc_fund" {
			found2 = true
			fmt.Printf("✓ 场外基金首字母搜索: hxczhh -> %s (%s, %s)\n", r.Name, r.Code, r.Type)
			break
		}
	}
	if !found2 {
		t.Fatalf("expected otc_fund 000001 in pinyin results, got %+v", results2)
	}
}

func TestFetchOTCFundQuote(t *testing.T) {
	cfg := config.Load()
	ms := NewMarketService(cfg)

	quote, err := ms.FetchOTCFundQuote(context.Background(), "000001")
	if err != nil {
		t.Fatalf("fetch otc fund quote failed: %v", err)
	}
	if quote.Price <= 0 {
		t.Fatalf("expected positive NAV, got %f", quote.Price)
	}
	fmt.Printf("✓ 场外基金净值获取: 000001, NAV=%.4f, Change=%.2f%%\n", quote.Price, quote.ChangePct)
}
