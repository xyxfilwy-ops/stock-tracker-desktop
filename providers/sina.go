package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// --------------- SinaProvider ---------------

// SinaProvider 实现 Provider 接口，对接新浪财经实时行情接口。
// 接口地址: https://hq.sinajs.cn/list={market}{code}
// 编码: GBK → UTF-8
// 请求头: 必须设置 Referer: https://finance.sina.com.cn
// 返回格式: , 分隔文本，新浪不直接返回涨跌幅，需自行计算。
type SinaProvider struct {
	client *http.Client
}

// NewSinaProvider 创建新浪财经数据源实例。
func NewSinaProvider(client *http.Client) *SinaProvider {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &SinaProvider{client: client}
}

// Name 返回数据源标识。
func (p *SinaProvider) Name() string {
	return SourceSina
}

// FetchQuote 获取单只股票实时行情。
func (p *SinaProvider) FetchQuote(ctx context.Context, code string) (*Quote, error) {
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
// 接口示例: https://hq.sinajs.cn/list=sh600519,sz000001
// 注意：必须设置 Referer 请求头，否则会被拒绝。
func (p *SinaProvider) FetchQuotes(ctx context.Context, codes []string) ([]*Quote, error) {
	if len(codes) == 0 {
		return []*Quote{}, nil
	}

	// 组装批量请求参数
	var parts []string
	for _, c := range codes {
		parts = append(parts, normalizeCode(c))
	}
	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", strings.Join(parts, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// 必须设置 Referer，否则新浪会拒绝请求
	req.Header.Set("Referer", "https://finance.sina.com.cn")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrSourceUnavailable, resp.StatusCode)
	}

	// GBK → UTF-8 解码
	utf8Reader := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	bodyBytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	body := string(bodyBytes)

	return p.parseBatch(body, codes)
}

// parseBatch 解析新浪批量返回的文本数据。
// 返回格式示例:
//   var hq_str_sh600519="贵州茅台,1720.00,1725.00,1750.50,1760.00,1710.00,...";
//   var hq_str_sz000001="平安银行,12.30,12.50,12.80,12.90,12.20,...";
//
// 关键字段下标（按 , 分割）:
//   idx 0 = 股票名称
//   idx 2 = 昨收价
//   idx 3 = 当前价
//
// 新浪不直接返回涨跌幅，需自行计算: (当前价 - 昨收价) / 昨收价 * 100%
func (p *SinaProvider) parseBatch(body string, codes []string) ([]*Quote, error) {
	lines := strings.Split(body, ";")
	results := make([]*Quote, 0, len(codes))

	codeIndex := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 提取 "var hq_str_sh600519=\"...\"" 中的内容
		eqIdx := strings.Index(line, "=\"")
		if eqIdx < 0 {
			continue
		}
		dataPart := line[eqIdx+2:]
		if strings.HasSuffix(dataPart, "\"") {
			dataPart = dataPart[:len(dataPart)-1]
		}
		if dataPart == "" {
			if codeIndex < len(codes) {
				codeIndex++
			}
			continue
		}

		// 按 , 分割字段
		fields := strings.Split(dataPart, ",")
		if len(fields) < 4 {
			if codeIndex < len(codes) {
				codeIndex++
			}
			continue // 字段不足，跳过
		}

		// 提取关键字段
		name := fields[0]
		prevCloseStr := fields[2]
		priceStr := fields[3]

		prevClose, err := strconv.ParseFloat(prevCloseStr, 64)
		if err != nil {
			if codeIndex < len(codes) {
				codeIndex++
			}
			continue
		}
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			if codeIndex < len(codes) {
				codeIndex++
			}
			continue
		}

		// 新浪不直接返回涨跌幅，需自行计算
		changePct := 0.0
		if prevClose > 0 {
			changePct = (price - prevClose) / prevClose * 100.0
		}

		// 判断停牌：成交量字段为 0（idx 8 为成交量）
		isSuspended := false
		if len(fields) > 8 {
			volume, _ := strconv.ParseFloat(fields[8], 64)
			isSuspended = volume == 0 && price == prevClose
		}

		// 确定 code
		var code string
		if codeIndex < len(codes) {
			code = normalizeCode(codes[codeIndex])
			codeIndex++
		}

		raw := RawQuote{
			Code:        code,
			Name:        name,
			Price:       price,
			PrevClose:   prevClose,
			ChangePct:   changePct,
			IsSuspended: isSuspended,
		}

		q, err := normalizeQuote(raw, SourceSina)
		if err != nil {
			continue
		}
		results = append(results, q)
	}

	return results, nil
}
