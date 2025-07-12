package client

import (
	"fmt"
	"sync"
	"time"

	"github.com/hewenyu/newapi-go/config"
	"github.com/hewenyu/newapi-go/internal/transport"
	"github.com/hewenyu/newapi-go/internal/utils"
)

// Client 是SDK的核心客户端结构
type Client struct {
	// config 存储客户端配置
	config *config.Config
	// transport HTTP传输层
	transport transport.HTTPTransport
	// logger 日志器
	logger utils.Logger
	// mu 用于保护客户端的并发安全
	mu sync.RWMutex
}

// NewClient 创建一个新的客户端实例
func NewClient(options ...ClientOption) (*Client, error) {
	// 创建客户端实例并设置默认配置
	client := &Client{
		config: config.DefaultConfig(),
		logger: utils.GetLogger(),
	}

	// 应用所有选项
	applyOptions(client, options)

	// 验证配置的有效性
	if err := client.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid client configuration: %w", err)
	}

	// 初始化HTTP传输层
	client.transport = transport.NewHTTPClient(
		client.config.BaseURL,
		client.config.APIKey,
		transport.WithTimeout(client.config.Timeout),
		transport.WithMiddleware(transport.LoggingMiddleware),
	)

	client.logger.Info("Client initialized successfully")

	return client, nil
}

// GetConfig 获取客户端配置的只读副本
func (c *Client) GetConfig() *config.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config.Clone()
}

// UpdateConfig 更新客户端配置
func (c *Client) UpdateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证新配置
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 保存旧配置用于回滚
	oldConfig := c.config

	// 更新配置
	c.config = cfg.Clone()

	// 重新初始化传输层
	if c.transport != nil {
		c.transport.Close()
	}

	c.transport = transport.NewHTTPClient(
		c.config.BaseURL,
		c.config.APIKey,
		transport.WithTimeout(c.config.Timeout),
		transport.WithMiddleware(transport.LoggingMiddleware),
	)

	// 如果初始化失败，回滚配置
	if c.transport == nil {
		c.config = oldConfig
		c.transport = transport.NewHTTPClient(
			oldConfig.BaseURL,
			oldConfig.APIKey,
			transport.WithTimeout(oldConfig.Timeout),
			transport.WithMiddleware(transport.LoggingMiddleware),
		)
		return fmt.Errorf("failed to initialize transport with new config")
	}

	c.logger.Info("Client configuration updated successfully")

	return nil
}

// Close 关闭客户端并清理资源
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport != nil {
		if err := c.transport.Close(); err != nil {
			utils.LogError(nil, err, "Failed to close transport")
			return fmt.Errorf("failed to close transport: %w", err)
		}
	}

	// 同步日志
	if err := c.logger.Sync(); err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	c.logger.Info("Client closed successfully")

	return nil
}

// GetAPIKey 获取API密钥
func (c *Client) GetAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config.APIKey
}

// GetBaseURL 获取基础URL
func (c *Client) GetBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config.BaseURL
}

// IsDebugMode 检查是否处于调试模式
func (c *Client) IsDebugMode() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config.Debug
}

// GetTimeout 获取超时时间
func (c *Client) GetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config.Timeout
}

// GetHTTPTransport 获取HTTP传输层
func (c *Client) GetHTTPTransport() transport.HTTPTransport {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.transport
}

// GetLogger 获取日志器
func (c *Client) GetLogger() utils.Logger {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.logger
}

// SetLogger 设置日志器
func (c *Client) SetLogger(logger utils.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger = logger
}

// SetTimeout 设置超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.Timeout = timeout
	if c.transport != nil {
		c.transport.SetTimeout(timeout)
	}
}

// SetRetryPolicy 设置重试策略
func (c *Client) SetRetryPolicy(policy transport.RetryPolicy) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport != nil {
		c.transport.SetRetryPolicy(policy)
	}
}

// IsHealthy 检查客户端健康状态
func (c *Client) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 检查基本配置
	if c.config == nil || c.transport == nil {
		return false
	}

	// 检查配置有效性
	if err := c.config.Validate(); err != nil {
		return false
	}

	return true
}

// String 返回客户端的字符串表示
func (c *Client) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("NewAPIClient{BaseURL: %s, Debug: %t, Timeout: %v}",
		c.config.BaseURL, c.config.Debug, c.config.Timeout)
}
