package services

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"stock-tracker/config"
	"stock-tracker/database"
	"stock-tracker/providers"

	"github.com/sony/gobreaker"
)

// KlineData 表示一条K线数据（价格使用 int64 分，涨跌幅使用 int64 BP）
type KlineData struct {
	Date      string `json:"date"`
	Open      int64  `json:"open"`
	Close     int64  `json:"close"`
	High      int64  `json:"high"`
	Low       int64  `json:"low"`
	Volume    int64  `json:"volume"`
	ChangePct int64  `json:"change_pct"`
}

// MarketService 负责行情编排：三源容灾调度、熔断器管理、健康探测
type MarketService struct {
	cfg       *config.Config
	client    *http.Client
	providers map[string]providers.Provider
	circuits  map[string]*gobreaker.CircuitBreaker
	mu        sync.RWMutex
}

// NewMarketService 创建 MarketService 实例
func NewMarketService(cfg *config.Config) *MarketService {
	ms := &MarketService{
		cfg:       cfg,
		client:    &http.Client{Timeout: cfg.HTTPTimeout},
		providers: make(map[string]providers.Provider),
		circuits:  make(map[string]*gobreaker.CircuitBreaker),
	}

	// 注册默认 provider
	ms.RegisterDefaultProviders()

	// 按配置初始化熔断器
	for _, ds := range cfg.DataSources {
		name := ds.Name
		cbSettings := gobreaker.Settings{
			Name:        name,
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     cfg.CircuitBreakerTimeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.Requests >= uint32(cfg.CircuitBreakerThreshold) &&
					float64(counts.TotalFailures)/float64(counts.Requests) >= 0.6
			},
		}
		ms.circuits[name] = gobreaker.NewCircuitBreaker(cbSettings)
	}

	// 启动后台健康探测
	go ms.healthCheckLoop()

	return ms
}

// RegisterProvider 注册一个数据源实现
func (ms *MarketService) RegisterProvider(name string, p providers.Provider) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.providers[name] = p
}

// RegisterDefaultProviders 注册默认的三源 provider
func (ms *MarketService) RegisterDefaultProviders() {
	ms.RegisterProvider(providers.SourceTencent, providers.NewTencentProvider(ms.client))
	ms.RegisterProvider(providers.SourceSina, providers.NewSinaProvider(ms.client))
	ms.RegisterProvider(providers.SourceEastMoney, providers.NewEastMoneyProvider(ms.client))
}

// FetchQuote 获取行情（无优先源，按默认优先级链）
func (ms *MarketService) FetchQuote(ctx context.Context, code string) (*providers.Quote, error) {
	return ms.FetchQuoteWithPreferred(ctx, code, "")
}

// FetchQuoteWithPreferred 三源容灾调度获取行情
// 优先使用 preferredSource 指定的源，失败后依次尝试备用源
// 成功后更新返回的 Quote.Source 字段
func (ms *MarketService) FetchQuoteWithPreferred(ctx context.Context, code string, preferredSource string) (*providers.Quote, error) {
	// 构建可用源列表
	sources := ms.buildSourceChain(preferredSource)
	if len(sources) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	var lastErr error
	for _, srcName := range sources {
		quote, err := ms.fetchWithCircuit(ctx, srcName, code)
		if err == nil && quote != nil {
			quote.Source = srcName
			return quote, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed for %s: %w", code, lastErr)
}

// buildSourceChain 构建数据源优先级链
func (ms *MarketService) buildSourceChain(preferredSource string) []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var chain []string

	// 优先使用上次成功的源
	if preferredSource != "" && ms.providers[preferredSource] != nil {
		chain = append(chain, preferredSource)
	}

	// 加入其他可用源
	for name := range ms.providers {
		if name != preferredSource {
			chain = append(chain, name)
		}
	}

	// 按 config 优先级排序
	priorityMap := make(map[string]int)
	for i, ds := range ms.cfg.DataSources {
		priorityMap[ds.Name] = i
	}

	sort.Slice(chain, func(i, j int) bool {
		return priorityMap[chain[i]] < priorityMap[chain[j]]
	})

	return chain
}

// fetchWithCircuit 在熔断器保护下调用 provider
func (ms *MarketService) fetchWithCircuit(ctx context.Context, sourceName, code string) (*providers.Quote, error) {
	cb, ok := ms.circuits[sourceName]
	if !ok {
		return nil, fmt.Errorf("unknown source: %s", sourceName)
	}

	result, err := cb.Execute(func() (interface{}, error) {
		ms.mu.RLock()
		p, ok := ms.providers[sourceName]
		ms.mu.RUnlock()
		if !ok {
			return nil, fmt.Errorf("provider not registered: %s", sourceName)
		}
		return p.FetchQuote(ctx, code)
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("nil quote from %s", sourceName)
	}

	quote, ok := result.(*providers.Quote)
	if !ok {
		return nil, fmt.Errorf("invalid quote type from %s", sourceName)
	}
	return quote, nil
}

// FetchKlineData 获取前复权 K 线数据（调用东方财富）
func (ms *MarketService) FetchKlineData(ctx context.Context, code string) ([]KlineData, error) {
	return ms.fetchKlineData(ctx, code, 1) // fqt=1 前复权
}

// FetchRawKlineData 获取不复权 K 线数据（调用东方财富）
func (ms *MarketService) FetchRawKlineData(ctx context.Context, code string) ([]KlineData, error) {
	return ms.fetchKlineData(ctx, code, 0) // fqt=0 不复权
}

// fetchKlineData 内部实现：获取 K 线数据并转换为 int64 格式
func (ms *MarketService) fetchKlineData(ctx context.Context, code string, fqt int) ([]KlineData, error) {
	ms.mu.RLock()
	p, ok := ms.providers[providers.SourceEastMoney]
	ms.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("eastmoney provider not available")
	}

	// 使用类型断言获取 EastMoneyProvider 的 K 线方法
	emProvider, ok := p.(*providers.EastMoneyProvider)
	if !ok {
		return nil, fmt.Errorf("eastmoney provider is not *providers.EastMoneyProvider")
	}

	// 使用熔断器保护
	cb, ok := ms.circuits[providers.SourceEastMoney]
	if !ok {
		return nil, fmt.Errorf("no circuit breaker for eastmoney")
	}

	result, err := cb.Execute(func() (interface{}, error) {
		return emProvider.FetchKLines(ctx, code, fqt, 30)
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("nil kline data from eastmoney")
	}

	rawData, ok := result.([]providers.KLineData)
	if !ok {
		return nil, fmt.Errorf("invalid kline data type")
	}

	// 转换为 services.KlineData（int64 格式）
	results := make([]KlineData, 0, len(rawData))
	for _, d := range rawData {
		results = append(results, KlineData{
			Date:      d.Date,
			Open:      database.ToPriceCents(d.Open),
			Close:     database.ToPriceCents(d.Close),
			High:      database.ToPriceCents(d.High),
			Low:       database.ToPriceCents(d.Low),
			Volume:    int64(d.Volume),
			ChangePct: database.ToBP(d.ChangePct),
		})
	}

	return results, nil
}

// GetAdjustedFactor 获取复权因子（通过东方财富 K 线数据）
// 复权因子 = 选入日不复权收盘价 / 选入日前复权收盘价，返回 BP（1000=1.0）
func (ms *MarketService) GetAdjustedFactor(ctx context.Context, code string, entryDate string) (int64, error) {
	ms.mu.RLock()
	p, ok := ms.providers[providers.SourceEastMoney]
	ms.mu.RUnlock()
	if !ok {
		return 1000, fmt.Errorf("eastmoney provider not available")
	}

	emProvider, ok := p.(*providers.EastMoneyProvider)
	if !ok {
		return 1000, fmt.Errorf("eastmoney provider type mismatch")
	}

	cb, ok := ms.circuits[providers.SourceEastMoney]
	if !ok {
		return 1000, fmt.Errorf("no circuit breaker for eastmoney")
	}

	result, err := cb.Execute(func() (interface{}, error) {
		return emProvider.GetAdjustedFactor(ctx, code, entryDate)
	})
	if err != nil {
		return 1000, err
	}
	if result == nil {
		return 1000, fmt.Errorf("nil factor from eastmoney")
	}

	factor, ok := result.(float64)
	if !ok {
		return 1000, fmt.Errorf("invalid factor type")
	}

	// 转换为 BP：factor * 1000
	return int64(factor*1000 + 0.5), nil
}

// GetEntryAdjustedPrice 获取选入日的前复权收盘价（int64 分）
func (ms *MarketService) GetEntryAdjustedPrice(ctx context.Context, code string, entryDate string) (int64, error) {
	ms.mu.RLock()
	p, ok := ms.providers[providers.SourceEastMoney]
	ms.mu.RUnlock()
	if !ok {
		return 0, fmt.Errorf("eastmoney provider not available")
	}

	emProvider, ok := p.(*providers.EastMoneyProvider)
	if !ok {
		return 0, fmt.Errorf("eastmoney provider type mismatch")
	}

	cb, ok := ms.circuits[providers.SourceEastMoney]
	if !ok {
		return 0, fmt.Errorf("no circuit breaker for eastmoney")
	}

	result, err := cb.Execute(func() (interface{}, error) {
		return emProvider.GetEntryAdjustedPrice(ctx, code, entryDate)
	})
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, fmt.Errorf("nil adjusted price from eastmoney")
	}

	price, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("invalid price type")
	}

	return database.ToPriceCents(price), nil
}

// GetEntryRawPrice 获取选入日的不复权收盘价（int64 分）
func (ms *MarketService) GetEntryRawPrice(ctx context.Context, code string, entryDate string) (int64, error) {
	ms.mu.RLock()
	p, ok := ms.providers[providers.SourceEastMoney]
	ms.mu.RUnlock()
	if !ok {
		return 0, fmt.Errorf("eastmoney provider not available")
	}

	emProvider, ok := p.(*providers.EastMoneyProvider)
	if !ok {
		return 0, fmt.Errorf("eastmoney provider type mismatch")
	}

	cb, ok := ms.circuits[providers.SourceEastMoney]
	if !ok {
		return 0, fmt.Errorf("no circuit breaker for eastmoney")
	}

	result, err := cb.Execute(func() (interface{}, error) {
		return emProvider.GetEntryRawPrice(ctx, code, entryDate)
	})
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, fmt.Errorf("nil raw price from eastmoney")
	}

	price, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("invalid price type")
	}

	return database.ToPriceCents(price), nil
}

// healthCheckLoop 后台 goroutine 每60s探测各源
func (ms *MarketService) healthCheckLoop() {
	ticker := time.NewTicker(ms.cfg.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		ms.mu.RLock()
		providersCopy := make(map[string]providers.Provider)
		for k, v := range ms.providers {
			providersCopy[k] = v
		}
		ms.mu.RUnlock()

		for name, p := range providersCopy {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, _ = p.FetchQuote(ctx, "sh600519") // 用茅台作为探测标的
			cancel()
			_ = name
		}
	}
}
