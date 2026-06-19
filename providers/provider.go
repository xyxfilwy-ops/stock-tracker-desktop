package providers

import (
	"context"
	"fmt"
	"math"
)

// --------------- 数据源标识 ---------------

const (
	SourceTencent  = "tencent"
	SourceSina     = "sina"
	SourceEastMoney = "eastmoney"
)

// --------------- Provider 接口 ---------------

// Provider 定义了所有行情数据源必须实现的统一接口。
// 每个具体实现负责协议适配、编码转换、字段映射，
// 最终通过 normalizeQuote 输出统一格式的 Quote。
type Provider interface {
	// Name 返回数据源标识名称（如 "tencent" / "sina" / "eastmoney"）
	Name() string

	// FetchQuote 获取单只股票的实时行情。
	// ctx 携带超时控制，上层设置 3s 超时。
	// code 格式为 "sh600519" 或 "sz000001"
	FetchQuote(ctx context.Context, code string) (*Quote, error)

	// FetchQuotes 批量获取行情（可选实现，用于刷新阶段优化）。
	// 若源不支持批量，可回退为逐只调用。
	FetchQuotes(ctx context.Context, codes []string) ([]*Quote, error)
}

// --------------- Quote 结构体（归一化后）---------------

// Quote 是经过统一校验和归一化后的行情数据结构，
// 无论数据来自腾讯、新浪还是东方财富，最终都转换为 Quote。
type Quote struct {
	Code        string  // 股票代码，如 "sh600519"
	Name        string  // 股票名称，如 "贵州茅台"
	Price       float64 // 当前价（元）
	PrevClose   float64 // 昨收价（元），用于计算日涨跌幅
	ChangePct   float64 // 涨跌幅（%），如 1.48 表示 +1.48%
	Source      string  // 数据来源标识（tencent/sina/eastmoney）
	IsSuspended bool    // 是否停牌
}

// --------------- RawQuote 结构体（归一化前）---------------

// RawQuote 是各数据源解析后的原始数值结构，
// 在交给 normalizeQuote 进行统一校验之前使用。
type RawQuote struct {
	Code        string
	Name        string
	Price       float64
	PrevClose   float64
	ChangePct   float64
	IsSuspended bool
	IsIPO       bool // 新股首日标记
}

// --------------- 公共错误定义 ---------------

var (
	ErrInvalidPrice     = fmt.Errorf("invalid price: must be > 0")
	ErrAbnormalChange   = fmt.Errorf("abnormal change percentage")
	ErrInvalidResponse  = fmt.Errorf("invalid response from data source")
	ErrSuspended        = fmt.Errorf("stock is suspended")
	ErrSourceUnavailable = fmt.Errorf("data source unavailable")
)

// --------------- normalizeQuote ---------------

// normalizeQuote 对 RawQuote 进行统一校验和归一化，
// 无论数据来自哪个源，最终都输出符合业务规则的 Quote。
//
// 校验规则（按 PRD 6.4 规范）：
//   1. 值域校验：价格必须 > 0
//   2. 新股首日保护：昨收价 < 0.01 时标记为 IPO 模式
//   3. 涨跌幅合理性：|ChangePct| <= 1000%（新股首日可放宽）
//   4. 停牌检测：停牌时价格保持停牌前值，涨跌幅置零
func normalizeQuote(raw RawQuote, source string) (*Quote, error) {
	// 1. 值域校验：价格必须为正数
	if raw.Price <= 0 {
		return nil, fmt.Errorf("%w: price=%v", ErrInvalidPrice, raw.Price)
	}

	// 2. 新股首日保护：昨收价几乎为 0，说明是新股上市首日
	// 此时不校验昨收价，ChangePct 可能极大（新股无涨跌幅限制）
	if raw.PrevClose < 0.01 && raw.Price > 0 {
		raw.IsIPO = true
		// 新股首日，若 ChangePct 未提供，设为 0
		if raw.ChangePct == 0 {
			// 保持 0，前端显示 "--" 或 "上市首日"
		}
	} else {
		// 非新股，昨收价也必须 > 0
		if raw.PrevClose <= 0 {
			return nil, fmt.Errorf("%w: prev_close=%v", ErrInvalidPrice, raw.PrevClose)
		}
	}

	// 3. 涨跌幅合理性校验
	// 正常交易日：|ChangePct| <= 1000%
	// 新股首日：不限制（北交所新股首日可能涨 200%+）
	absChange := math.Abs(raw.ChangePct)
	if !raw.IsIPO && absChange > 1000 {
		return nil, fmt.Errorf("%w: change_pct=%.2f%%", ErrAbnormalChange, raw.ChangePct)
	}
	// 即便是新股，超过 5000% 也判定为异常数据
	if raw.IsIPO && absChange > 5000 {
		return nil, fmt.Errorf("%w: IPO change_pct=%.2f%%", ErrAbnormalChange, raw.ChangePct)
	}

	// 4. 停牌处理
	q := &Quote{
		Code:        raw.Code,
		Name:        raw.Name,
		Price:       raw.Price,
		PrevClose:   raw.PrevClose,
		ChangePct:   raw.ChangePct,
		Source:      source,
		IsSuspended: raw.IsSuspended,
	}

	if q.IsSuspended {
		// 停牌时：价格保持停牌前值（不变），涨跌幅置 0
		q.ChangePct = 0
	}

	return q, nil
}

// --------------- 辅助工具函数 ---------------

// normalizeCode 将用户输入的股票代码标准化为 "sh600519" 或 "sz000001" 格式。
// 支持：
//   - "600519" → "sh600519"
//   - "000001" → "sz000001"
//   - "sh600519" → "sh600519"（不变）
//   - "SH600519" → "sh600519"
func normalizeCode(code string) string {
	if len(code) == 0 {
		return ""
	}

	// 统一小写
	c := code
	// 已有 sh/sz 前缀，统一小写
	if len(c) >= 2 && (c[0] == 's' || c[0] == 'S') {
		if c[1] == 'h' || c[1] == 'H' {
			return "sh" + c[2:]
		}
		if c[1] == 'z' || c[1] == 'Z' {
			return "sz" + c[2:]
		}
	}

	// 纯数字，根据代码规则判断市场
	if len(code) >= 6 {
		prefix := code[0]
		switch prefix {
		case '6', '5': // 上海（主板、科创板）
			return "sh" + code
		case '0', '3', '2', '4': // 深圳（主板、创业板、中小板、B股）
			return "sz" + code
		}
	}

	// 无法识别，原样返回
	return code
}

// extractNumericCode 从 "sh600519" 中提取 "600519"
func extractNumericCode(code string) string {
	if len(code) > 2 {
		return code[2:]
	}
	return code
}

// extractMarketPrefix 从 "sh600519" 中提取 "sh" 或 "sz"
func extractMarketPrefix(code string) string {
	if len(code) >= 2 {
		return code[:2]
	}
	return ""
}
