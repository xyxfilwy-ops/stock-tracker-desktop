package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// DataSource 表示一个行情数据源配置
type DataSource struct {
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Priority  int           `json:"priority"`
	Timeout   time.Duration `json:"timeout"`
	Referer   string        `json:"referer,omitempty"`
	Encoding  string        `json:"encoding"`        // "utf-8" 或 "gbk"
}

// Config 集中配置所有运行时参数
type Config struct {
	// 数据库
	DBPath string `json:"db_path"`

	// 网络超时
	HTTPTimeout     time.Duration `json:"http_timeout"`
	ProviderTimeout time.Duration `json:"provider_timeout"`

	// 并发控制
	RefreshConcurrency int `json:"refresh_concurrency"`

	// 数据源配置（优先级从高到低）
	DataSources []DataSource `json:"data_sources"`

	// 熔断器配置
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
	HealthCheckInterval     time.Duration `json:"health_check_interval"`

	// 时区
	Timezone string `json:"timezone"`
}

// defaultDataSources 返回内置的默认数据源列表
func defaultDataSources() []DataSource {
	return []DataSource{
		{
			Name:     "tencent",
			URL:      "https://qt.gtimg.cn/q=%s",
			Priority: 1,
			Timeout:  3 * time.Second,
			Encoding: "gbk",
		},
		{
			Name:     "sina",
			URL:      "https://hq.sinajs.cn/list=%s",
			Priority: 2,
			Timeout:  3 * time.Second,
			Referer:  "https://finance.sina.com.cn",
			Encoding: "gbk",
		},
		{
			Name:     "eastmoney",
			URL:      "https://push2.eastmoney.com/api/qt/stock/get",
			Priority: 3,
			Timeout:  3 * time.Second,
			Encoding: "utf-8",
		},
	}
}

// userDataDir 返回 OS 用户数据目录下的 StockTracker 子目录
func userDataDir() string {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("APPDATA")
		}
		if baseDir == "" {
			baseDir = os.Getenv("USERPROFILE")
			if baseDir != "" {
				baseDir = filepath.Join(baseDir, "AppData", "Local")
			}
		}
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err == nil {
			baseDir = filepath.Join(homeDir, "Library", "Application Support")
		}
	default: // linux and others
		homeDir, err := os.UserHomeDir()
		if err == nil {
			baseDir = filepath.Join(homeDir, ".local", "share")
		}
	}

	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}

	return filepath.Join(baseDir, "StockTracker")
}

// Load 加载配置，所有未显式设置的字段使用默认值
func Load() *Config {
	dataDir := userDataDir()

	cfg := &Config{
		DBPath:               filepath.Join(dataDir, "stocktracker.db"),
		HTTPTimeout:          10 * time.Second,
		ProviderTimeout:      3 * time.Second,
		RefreshConcurrency:   5,
		DataSources:          defaultDataSources(),
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
		HealthCheckInterval:   60 * time.Second,
		Timezone:              "Asia/Shanghai",
	}

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create data directory: %v\n", err)
	}

	return cfg
}

// GetDataSource 按名称查找数据源配置
func (c *Config) GetDataSource(name string) *DataSource {
	for i := range c.DataSources {
		if c.DataSources[i].Name == name {
			return &c.DataSources[i]
		}
	}
	return nil
}

// GetPrimaryDataSource 返回优先级最高的数据源
func (c *Config) GetPrimaryDataSource() *DataSource {
	if len(c.DataSources) == 0 {
		return nil
	}
	return &c.DataSources[0]
}
