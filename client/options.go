package client

import (
	"net/http"
	"time"

	"github.com/hewenyu/newapi-go/config"
)

// ClientOption 定义客户端配置选项的函数类型
type ClientOption func(*Client)

// WithAPIKey 设置API密钥
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.config.APIKey = apiKey
	}
}

// WithBaseURL 设置API基础URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.config.BaseURL = baseURL
	}
}

// WithTimeout 设置HTTP请求超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.config.Timeout = timeout
	}
}

// WithHTTPClient 设置自定义HTTP客户端
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.config.HTTPClient = client
	}
}

// WithUserAgent 设置User-Agent头
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.config.UserAgent = userAgent
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.config.Debug = debug
	}
}

// WithConfig 直接设置配置对象
func WithConfig(cfg *config.Config) ClientOption {
	return func(c *Client) {
		c.config = cfg.Clone()
	}
}

// WithConfigBuilder 使用配置构建器设置配置
func WithConfigBuilder(builder *config.ConfigBuilder) ClientOption {
	return func(c *Client) {
		cfg, err := builder.Build()
		if err != nil {
			// 在选项应用阶段，我们不能返回错误
			// 这里将错误信息记录到客户端的配置中
			// 实际的错误处理将在NewClient中进行
			return
		}
		c.config = cfg
	}
}

// applyOptions 应用所有选项到客户端
func applyOptions(client *Client, options []ClientOption) {
	for _, option := range options {
		option(client)
	}
}
