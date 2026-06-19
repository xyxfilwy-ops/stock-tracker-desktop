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

// --------------- TencentProvider ---------------

// TencentProvider 实现 Provider 接口，对接腾讯财经实时行情接口。
// 接口地址: https://qt.gtimg.cn/q={market}{code}
// 编码: GBK → UTF-8
// 返回格式: ~ 分隔文本
type TencentProvider struct {
	client *http.Client
}

// NewTencentProvider 创建腾讯财经数据源实例。
func NewTencentProvider(client *http.Client) *TencentProvider {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &TencentProvider{client: client}
}

// Name 返回数据源标识。
func (p *TencentProvider) Name() string {
	return SourceTencent
}

// FetchQuote 获取单只股票实时行情。
func (p *TencentProvider) FetchQuote(ctx context.Context, code string) (*Quote, error) {
	quotes, err := p.FetchQuotes(ctx, []string{code})
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("%w: no data returned for %s", ErrInvalidResponse, code)
	}
	return quotes[0], nil
}

// FetchQuotes 批量获取股票实时行情，支持逗号分隔批量请求。
// 接口示例: https://qt.gtimg.cn/q=sh600519,sz000001
func (p *TencentProvider) FetchQuotes(ctx context.Context, codes []string) ([]*Quote, error) {
	if len(codes) == 0 {
		return []*Quote{}, nil
	}

	// 组装批量请求参数
	var parts []string
	for _, c := range codes {
		parts = append(parts, normalizeCode(c))
	}
	url := fmt.Sprintf("https://qt.gtimg.cn/q=%s", strings.Join(parts, ","))

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

	// GBK → UTF-8 解码
	utf8Reader := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	bodyBytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	body := string(bodyBytes)

	return p.parseBatch(body, codes)
}

// parseBatch 解析腾讯批量返回的文本数据。
// 返回格式示例:
//   v_sh600519="1~贵州茅台~600519~1725.00~1730.00~...~...";
//   v_sz000001="1~平安银行~000001~12.50~12.30~...~...";
//
// 关键字段下标（A 股布局，已验证）:
//   idx 1  = 股票名称
//   idx 3  = 当前价
//   idx 4  = 昨收价
//   idx 31 = 涨跌额
//   idx 32 = 涨跌幅%
func (p *TencentProvider) parseBatch(body string, codes []string) ([]*Quote, error) {
	lines := strings.Split(body, ";")
	results := make([]*Quote, 0, len(codes))

	codeIndex := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 提取 "v_sh600519=\"...\"" 中的内容
		// 分割 key=value
		eqIdx := strings.Index(line, "=\"")
		if eqIdx < 0 {
			continue
		}
		// dataPart 是 " 到结尾 " 之间的内容
		dataPart := line[eqIdx+2:]
		if strings.HasSuffix(dataPart, "\"") {
			dataPart = dataPart[:len(dataPart)-1]
		}

		// 按 ~ 分割字段
		fields := strings.Split(dataPart, "~")

		// 防御性解析：A 股应 >= 45 个字段
		if len(fields) < 45 {
			if codeIndex < len(codes) {
				codeIndex++
			}
			continue // 跳过无效响应
		}

		// 提取关键字段
		name := fields[1]
		priceStr := fields[3]
		prevCloseStr := fields[4]
		changePctStr := fields[32]

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			if codeIndex < len(codes) {
				codeIndex++
			}
			continue
		}
		prevClose, err := strconv.ParseFloat(prevCloseStr, 64)
		if err != nil {
			if codeIndex < len(codes) {
				codeIndex++
			}
			continue
		}
		changePct, err := strconv.ParseFloat(changePctStr, 64)
		if err != nil {
			changePct = 0
		}

		// 确定 code
		var code string
		if codeIndex < len(codes) {
			code = normalizeCode(codes[codeIndex])
			codeIndex++
		}

		raw := RawQuote{
			Code:      code,
			Name:      name,
			Price:     price,
			PrevClose: prevClose,
			ChangePct: changePct,
			// 腾讯接口不直接返回停牌状态，需通过成交量等判断
			// 这里简化为：价格为0或成交量为0时可能停牌
			IsSuspended: price == 0 || price == prevClose && changePct == 0,
		}

		q, err := normalizeQuote(raw, SourceTencent)
		if err != nil {
			continue // 跳过异常数据
		}
		results = append(results, q)
	}

	return results, nil
}
