package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// --------------- EastMoneyProvider ---------------

// EastMoneyProvider 实现 Provider 接口，对接东方财富实时行情接口。
// 实时行情接口: https://push2.eastmoney.com/api/qt/stock/get
// 编码: UTF-8
// 返回格式: JSON
//
// 前复权K线接口: https://push2his.eastmoney.com/api/qt/stock/kline/get
// 用于选入时计算复权因子。
type EastMoneyProvider struct {
	client *http.Client
}

// NewEastMoneyProvider 创建东方财富数据源实例。
func NewEastMoneyProvider(client *http.Client) *EastMoneyProvider {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &EastMoneyProvider{client: client}
}

// Name 返回数据源标识。
func (p *EastMoneyProvider) Name() string {
	return SourceEastMoney
}

// FetchQuote 获取单只股票实时行情。
func (p *EastMoneyProvider) FetchQuote(ctx context.Context, code string) (*Quote, error) {
	quotes, err := p.FetchQuotes(ctx, []string{code})
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("%w: no data returned for %s", ErrInvalidResponse, code)
	}
	return quotes[0], nil
}

// FetchQuotes 批量获取股票实时行情。
// 东方财富接口不支持真正的批量请求，此处逐只获取后聚合。
func (p *EastMoneyProvider) FetchQuotes(ctx context.Context, codes []string) ([]*Quote, error) {
	if len(codes) == 0 {
		return []*Quote{}, nil
	}

	results := make([]*Quote, 0, len(codes))
	for _, code := range codes {
		q, err := p.fetchSingle(ctx, code)
		if err != nil {
			continue // 单只失败不影响其他
		}
		results = append(results, q)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%w: all %d requests failed", ErrSourceUnavailable, len(codes))
	}
	return results, nil
}

// fetchSingle 获取单只股票实时行情。
// 接口: https://push2.eastmoney.com/api/qt/stock/get
// 参数:
//   secid={market_id}.{code}（沪=1, 深=0）
//   fields=f43,f44,f45,f46,f47,f48,f58,f100,f168,f170
//
// 字段说明:
//   f43  = 当前价（可能为整数，需配合 f100 精度判断）
//   f44  = 最高价
//   f45  = 最低价
//   f46  = 开盘价
//   f47  = 昨收价（可能为整数）
//   f48  = 涨跌幅（小数，如 0.0029 = 0.29%）
//   f58  = 股票名称
//   f100 = 价格小数位数（精度）
//   f168 = 停牌状态: 0=正常, 1=停牌
//   f170 = 涨跌幅（备用）
func (p *EastMoneyProvider) fetchSingle(ctx context.Context, code string) (*Quote, error) {
	ncode := normalizeCode(code)
	secid := toEastMoneySecID(ncode)

	url := fmt.Sprintf(
		"https://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f58,f100,f168,f170",
		secid,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrSourceUnavailable, resp.StatusCode)
	}

	var emResp struct {
		Data struct {
			F43  json.Number `json:"f43"`  // 当前价
			F44  json.Number `json:"f44"`  // 最高价
			F45  json.Number `json:"f45"`  // 最低价
			F46  json.Number `json:"f46"`  // 开盘价
			F47  json.Number `json:"f47"`  // 昨收价
			F48  json.Number `json:"f48"`  // 涨跌幅（小数）
			F58  string      `json:"f58"`  // 股票名称
			F100 json.Number `json:"f100"` // 价格小数位数
			F168 json.Number `json:"f168"` // 停牌状态: 0=正常, 1=停牌
			F170 json.Number `json:"f170"` // 涨跌幅（备用）
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emResp); err != nil {
		return nil, fmt.Errorf("%w: decode error: %v", ErrInvalidResponse, err)
	}

	data := emResp.Data
	if data.F58 == "" && data.F43 == "" {
		return nil, fmt.Errorf("%w: empty data", ErrInvalidResponse)
	}

	// 解析精度
	precision := 2
	if data.F100 != "" {
		if p, err := data.F100.Int64(); err == nil {
			precision = int(p)
		}
	}
	divisor := math.Pow(10, float64(precision))

	// 解析价格（精度处理：值 > 10000 时除以 10^precision）
	price := parseEastMoneyPrice(data.F43, divisor)
	prevClose := parseEastMoneyPrice(data.F47, divisor)

	// 解析涨跌幅（f48 或 f170，小数形式，如 0.0029 = 0.29%）
	changePct := 0.0
	if data.F48 != "" {
		if v, err := data.F48.Float64(); err == nil {
			changePct = v * 100.0 // 转换为百分比
		}
	}
	if changePct == 0 && data.F170 != "" {
		if v, err := data.F170.Float64(); err == nil {
			changePct = v * 100.0
		}
	}

	// 解析停牌状态
	isSuspended := false
	if data.F168 != "" {
		if v, err := data.F168.Int64(); err == nil && v == 1 {
			isSuspended = true
		}
	}

	name := data.F58
	if name == "" {
		name = ncode // fallback
	}

	raw := RawQuote{
		Code:        ncode,
		Name:        name,
		Price:       price,
		PrevClose:   prevClose,
		ChangePct:   changePct,
		IsSuspended: isSuspended,
	}

	q, err := normalizeQuote(raw, SourceEastMoney)
	if err != nil {
		return nil, err
	}
	return q, nil
}

// FetchOTCFundQuote 获取场外基金最新净值（单位净值，元）
// 接口: https://fund.eastmoney.com/f10/F10DataApi.aspx?type=lsjz&code={code}&page=1&per=1
func (p *EastMoneyProvider) FetchOTCFundQuote(ctx context.Context, code string) (*Quote, error) {
	url := fmt.Sprintf(
		"https://fund.eastmoney.com/f10/F10DataApi.aspx?type=lsjz&code=%s&page=1&per=1",
		code,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://fund.eastmoney.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrSourceUnavailable, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %v", err)
	}
	body := string(bodyBytes)

	// 解析 HTML 表格内容
	// content 中是一个 HTML table，包含 thead 和 tbody
	// tbody 中的 td 依次是: 净值日期, 单位净值, 累计净值, 日增长率, 申购状态, 赎回状态, 分红送配
	contentStart := strings.Index(body, "content:\"")
	contentEnd := strings.Index(body, "\",records")
	if contentStart < 0 || contentEnd < 0 || contentEnd <= contentStart {
		return nil, fmt.Errorf("%w: cannot parse fund data", ErrInvalidResponse)
	}
	content := body[contentStart+9 : contentEnd]

	// 提取所有 <td> 内容
	var tdValues []string
	re := regexp.MustCompile(`<td[^>]*>([^<]*)</td>`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			val := strings.TrimSpace(m[1])
			val = strings.ReplaceAll(val, "&nbsp;", "")
			val = strings.TrimSpace(val)
			if val != "" {
				tdValues = append(tdValues, val)
			}
		}
	}

	if len(tdValues) < 4 {
		return nil, fmt.Errorf("%w: cannot parse NAV from HTML (found %d td values)", ErrInvalidResponse, len(tdValues))
	}

	// tbody 中的 td 顺序：日期, 单位净值, 累计净值, 日增长率, 申购状态, 赎回状态, 分红送配
	navDate := tdValues[0]
	navStr := tdValues[1]
	changePctStr := tdValues[3]

	if navDate == "" || navStr == "" {
		return nil, fmt.Errorf("%w: cannot parse NAV from HTML", ErrInvalidResponse)
	}

	nav, err := strconv.ParseFloat(navStr, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid NAV %s", ErrInvalidResponse, navStr)
	}

	// 解析日增长率（如 "2.84%" → 2.84）
	changePct := 0.0
	if changePctStr != "" {
		changePctStr = strings.TrimSuffix(changePctStr, "%")
		changePctStr = strings.TrimSpace(changePctStr)
		changePct, _ = strconv.ParseFloat(changePctStr, 64)
	}

	// 场外基金：昨收净值 = 当前净值 / (1 + 日增长率/100)
	prevNav := nav
	if changePct != 0 {
		prevNav = nav / (1 + changePct/100)
	}

	name := code

	return &Quote{
		Code:      code,
		Name:      name,
		Price:     nav,
		PrevClose: prevNav,
		ChangePct: changePct,
		Source:    SourceEastMoney,
	}, nil
}

// parseEastMoneyPrice 解析东方财富的价格字段，处理精度问题。
// 当 f43/f47 值 > 10000 时，说明返回了整数形式，需除以 10^precision 归一化。
func parseEastMoneyPrice(val json.Number, divisor float64) float64 {
	if val == "" || val == "-" {
		return 0
	}
	v, err := val.Float64()
	if err != nil {
		return 0
	}
	if v > 10000 {
		v = v / divisor
	}
	return v
}

// toEastMoneySecID 将 "sh600519" 转换为东方财富 secid 格式 "1.600519"（沪=1, 深=0）。
func toEastMoneySecID(code string) string {
	ncode := normalizeCode(code)
	numeric := extractNumericCode(ncode)
	prefix := extractMarketPrefix(ncode)

	var marketID string
	switch prefix {
	case "sh":
		marketID = "1"
	case "sz":
		marketID = "0"
	default:
		// 默认上海
		marketID = "1"
	}

	return marketID + "." + numeric
}

// --------------- 前复权 K 线接口 ---------------

// KLineData 表示单条K线数据。
type KLineData struct {
	Date      string  // 日期，如 "2026-06-19"
	Open      float64 // 开盘价
	Close     float64 // 收盘价
	High      float64 // 最高价
	Low       float64 // 最低价
	Volume    float64 // 成交量
	ChangePct float64 // 涨跌幅
}

// FetchKLines 获取股票前复权（或不复权）K线数据。
// 接口: https://push2his.eastmoney.com/api/qt/stock/kline/get
// 参数:
//   secid={market_id}.{code}
//   klt=101   // K线周期：101=日K
//   fqt=1     // 复权类型：1=前复权, 0=不复权, 2=后复权
//   lmt=30    // 返回条数
//
// 返回 klines 数组，每条以逗号分隔:
//   "2026-06-01,1715.00,1730.00,1700.00,1725.00,1720.00,0.0029,..."
//   列: 日期, 开盘价, 收盘价, 最高价, 最低价, 成交量, 涨跌幅, ...
func (p *EastMoneyProvider) FetchKLines(ctx context.Context, code string, fqt int, lmt int) ([]KLineData, error) {
	ncode := normalizeCode(code)
	secid := toEastMoneySecID(ncode)

	url := fmt.Sprintf(
		"https://push2his.eastmoney.com/api/qt/stock/kline/get?secid=%s&klt=101&fqt=%d&lmt=%d",
		secid, fqt, lmt,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrSourceUnavailable, resp.StatusCode)
	}

	var emResp struct {
		Data struct {
			KLines []string `json:"klines"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emResp); err != nil {
		return nil, fmt.Errorf("%w: decode error: %v", ErrInvalidResponse, err)
	}

	return p.parseKLines(emResp.Data.KLines)
}

// parseKLines 解析 klines 数组。
// 每条格式: "2026-06-01,开盘价,收盘价,最高价,最低价,成交量,涨跌幅,..."
func (p *EastMoneyProvider) parseKLines(lines []string) ([]KLineData, error) {
	results := make([]KLineData, 0, len(lines))

	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			continue
		}

		date := fields[0]
		open, _ := strconv.ParseFloat(fields[1], 64)
		close_, _ := strconv.ParseFloat(fields[2], 64)
		high, _ := strconv.ParseFloat(fields[3], 64)
		low, _ := strconv.ParseFloat(fields[4], 64)
		volume, _ := strconv.ParseFloat(fields[5], 64)
		changePct := 0.0
		if len(fields) > 6 {
			changePct, _ = strconv.ParseFloat(fields[6], 64)
			changePct = changePct * 100.0 // 转换为百分比
		}

		results = append(results, KLineData{
			Date:      date,
			Open:      open,
			Close:     close_,
			High:      high,
			Low:       low,
			Volume:    volume,
			ChangePct: changePct,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%w: no valid kline data", ErrInvalidResponse)
	}
	return results, nil
}

// GetAdjustedFactor 计算复权因子。
// 复权因子 = 选入日不复权收盘价 / 选入日前复权收盘价
// 用于选入时一次性计算并存储到 stocks 表的 adjust_factor 字段。
func (p *EastMoneyProvider) GetAdjustedFactor(ctx context.Context, code string, entryDate string) (float64, error) {
	// 1. 获取前复权数据（fqt=1）
	adjKLines, err := p.FetchKLines(ctx, code, 1, 30)
	if err != nil {
		return 0, fmt.Errorf("fetch adjusted klines failed: %w", err)
	}

	// 2. 获取不复权数据（fqt=0）
	rawKLines, err := p.FetchKLines(ctx, code, 0, 30)
	if err != nil {
		return 0, fmt.Errorf("fetch raw klines failed: %w", err)
	}

	// 3. 从两组数据中查找选入日期的收盘价
	var adjClose, rawClose float64
	for _, k := range adjKLines {
		if k.Date == entryDate {
			adjClose = k.Close
			break
		}
	}
	for _, k := range rawKLines {
		if k.Date == entryDate {
			rawClose = k.Close
			break
		}
	}

	if adjClose <= 0 || rawClose <= 0 {
		return 0, fmt.Errorf("%w: cannot find entry date %s in klines", ErrInvalidResponse, entryDate)
	}

	// 4. 计算复权因子
	factor := rawClose / adjClose
	return factor, nil
}

// GetEntryAdjustedPrice 获取选入日的前复权收盘价。
// 用于选入时写入 stocks 表的 entry_price 字段。
func (p *EastMoneyProvider) GetEntryAdjustedPrice(ctx context.Context, code string, entryDate string) (float64, error) {
	adjKLines, err := p.FetchKLines(ctx, code, 1, 30)
	if err != nil {
		return 0, fmt.Errorf("fetch adjusted klines failed: %w", err)
	}

	for _, k := range adjKLines {
		if k.Date == entryDate {
			if k.Close > 0 {
				return k.Close, nil
			}
			return 0, fmt.Errorf("%w: adjusted close is zero", ErrInvalidPrice)
		}
	}

	return 0, fmt.Errorf("%w: cannot find entry date %s in klines", ErrInvalidResponse, entryDate)
}

// GetEntryRawPrice 获取选入日的不复权收盘价。
// 用于选入时写入 stocks 表的 raw_price 字段。
func (p *EastMoneyProvider) GetEntryRawPrice(ctx context.Context, code string, entryDate string) (float64, error) {
	rawKLines, err := p.FetchKLines(ctx, code, 0, 30)
	if err != nil {
		return 0, fmt.Errorf("fetch raw klines failed: %w", err)
	}

	for _, k := range rawKLines {
		if k.Date == entryDate {
			if k.Close > 0 {
				return k.Close, nil
			}
			return 0, fmt.Errorf("%w: raw close is zero", ErrInvalidPrice)
		}
	}

	return 0, fmt.Errorf("%w: cannot find entry date %s in klines", ErrInvalidResponse, entryDate)
}
