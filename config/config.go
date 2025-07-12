package config

import (
	"fmt"
	"net/http"
	"time"
)

// Config 包含SDK的所有配置选项
type Config struct {
	// APIKey 是访问API的密钥
	APIKey string
	// BaseURL 是API的基础URL
	BaseURL string
	// Timeout 是HTTP请求的超时时间
	Timeout time.Duration
	// HTTPClient 是自定义的HTTP客户端
	HTTPClient *http.Client
	// UserAgent 是请求的User-Agent头
	UserAgent string
	// Debug 是否启用调试模式
	Debug bool
}

// ConfigBuilder 是配置构建器，用于创建Config实例
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder 创建一个新的配置构建器实例
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &Config{
			BaseURL:    DefaultBaseURL,
			Timeout:    DefaultTimeout,
			HTTPClient: DefaultHTTPClient(),
			UserAgent:  DefaultUserAgent,
			Debug:      false,
		},
	}
}

// WithAPIKey 设置API密钥
func (b *ConfigBuilder) WithAPIKey(apiKey string) *ConfigBuilder {
	b.config.APIKey = apiKey
	return b
}

// WithBaseURL 设置基础URL
func (b *ConfigBuilder) WithBaseURL(baseURL string) *ConfigBuilder {
	b.config.BaseURL = baseURL
	return b
}

// WithTimeout 设置请求超时时间
func (b *ConfigBuilder) WithTimeout(timeout time.Duration) *ConfigBuilder {
	b.config.Timeout = timeout
	return b
}

// WithHTTPClient 设置自定义HTTP客户端
func (b *ConfigBuilder) WithHTTPClient(client *http.Client) *ConfigBuilder {
	b.config.HTTPClient = client
	return b
}

// WithUserAgent 设置User-Agent头
func (b *ConfigBuilder) WithUserAgent(userAgent string) *ConfigBuilder {
	b.config.UserAgent = userAgent
	return b
}

// WithDebug 设置调试模式
func (b *ConfigBuilder) WithDebug(debug bool) *ConfigBuilder {
	b.config.Debug = debug
	return b
}

// Build 构建并返回配置实例
func (b *ConfigBuilder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return b.config, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got: %v", c.Timeout)
	}

	if c.HTTPClient == nil {
		return fmt.Errorf("HTTP client is required")
	}

	if c.UserAgent == "" {
		return fmt.Errorf("user agent is required")
	}

	return nil
}

// Clone 创建配置的深拷贝
func (c *Config) Clone() *Config {
	return &Config{
		APIKey:     c.APIKey,
		BaseURL:    c.BaseURL,
		Timeout:    c.Timeout,
		HTTPClient: c.HTTPClient,
		UserAgent:  c.UserAgent,
		Debug:      c.Debug,
	}
}
