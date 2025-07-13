package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 代理服务器配置
type Config struct {
	// NEW API配置
	NewAPIURL string // 从NEW_API环境变量获取
	NewAPIKey string // 从NEW_API_KEY环境变量获取

	// 代理服务器配置
	ServerPort  int    // 代理服务器端口，默认8080
	ServerHost  string // 代理服务器主机，默认0.0.0.0
	LogLevel    string // 日志级别，默认INFO
	EnableDebug bool   // 调试模式

	// 性能配置
	RequestTimeout   time.Duration // 请求超时时间
	MaxRequestSize   int64         // 最大请求体大小
	MaxConcurrent    int           // 最大并发数
	EnableCORS       bool          // 启用CORS
	CORSAllowOrigins []string      // CORS允许的来源
	CORSAllowMethods []string      // CORS允许的方法
	CORSAllowHeaders []string      // CORS允许的头部
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	config := &Config{
		// 默认值
		ServerPort:       8082,
		ServerHost:       "0.0.0.0",
		LogLevel:         "INFO",
		EnableDebug:      false,
		RequestTimeout:   30 * time.Second,
		MaxRequestSize:   10 * 1024 * 1024, // 10MB
		MaxConcurrent:    100,
		EnableCORS:       true,
		CORSAllowOrigins: []string{"*"},
		CORSAllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSAllowHeaders: []string{"Content-Type", "Authorization", "X-API-Key", "anthropic-version"},
	}

	// 必需的环境变量
	config.NewAPIURL = os.Getenv("NEW_API")
	if config.NewAPIURL == "" {
		return nil, fmt.Errorf("NEW_API environment variable is required")
	}

	config.NewAPIKey = os.Getenv("NEW_API_KEY")
	if config.NewAPIKey == "" {
		return nil, fmt.Errorf("NEW_API_KEY environment variable is required")
	}

	// 可选的环境变量
	if port := os.Getenv("PROXY_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.ServerPort = p
		}
	}

	if host := os.Getenv("PROXY_HOST"); host != "" {
		config.ServerHost = host
	}

	if logLevel := os.Getenv("PROXY_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if debug := os.Getenv("PROXY_DEBUG"); debug != "" {
		if d, err := strconv.ParseBool(debug); err == nil {
			config.EnableDebug = d
		}
	}

	if timeout := os.Getenv("PROXY_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			config.RequestTimeout = t
		}
	}

	if maxSize := os.Getenv("PROXY_MAX_REQUEST_SIZE"); maxSize != "" {
		if size, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			config.MaxRequestSize = size
		}
	}

	if maxConcurrent := os.Getenv("PROXY_MAX_CONCURRENT"); maxConcurrent != "" {
		if c, err := strconv.Atoi(maxConcurrent); err == nil {
			config.MaxConcurrent = c
		}
	}

	if cors := os.Getenv("PROXY_ENABLE_CORS"); cors != "" {
		if c, err := strconv.ParseBool(cors); err == nil {
			config.EnableCORS = c
		}
	}

	return config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.NewAPIURL == "" {
		return fmt.Errorf("NEW_API_URL is required")
	}

	if c.NewAPIKey == "" {
		return fmt.Errorf("NEW_API_KEY is required")
	}

	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", c.ServerPort)
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	if c.MaxRequestSize <= 0 {
		return fmt.Errorf("max request size must be positive")
	}

	if c.MaxConcurrent <= 0 {
		return fmt.Errorf("max concurrent must be positive")
	}

	return nil
}

// GetServerAddress 获取服务器地址
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

// IsDebugEnabled 检查是否启用调试模式
func (c *Config) IsDebugEnabled() bool {
	return c.EnableDebug
}

// GetLogLevel 获取日志级别
func (c *Config) GetLogLevel() string {
	return c.LogLevel
}

// Print 打印配置信息（隐藏敏感信息）
func (c *Config) Print() {
	fmt.Printf("Configuration:\n")
	fmt.Printf("  NEW_API_URL: %s\n", c.NewAPIURL)
	fmt.Printf("  NEW_API_KEY: %s****\n", c.NewAPIKey[:min(len(c.NewAPIKey), 8)])
	fmt.Printf("  Server: %s\n", c.GetServerAddress())
	fmt.Printf("  Log Level: %s\n", c.LogLevel)
	fmt.Printf("  Debug: %t\n", c.EnableDebug)
	fmt.Printf("  Request Timeout: %v\n", c.RequestTimeout)
	fmt.Printf("  Max Request Size: %d bytes\n", c.MaxRequestSize)
	fmt.Printf("  Max Concurrent: %d\n", c.MaxConcurrent)
	fmt.Printf("  CORS Enabled: %t\n", c.EnableCORS)
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
