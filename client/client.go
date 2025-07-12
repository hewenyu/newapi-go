package client

import (
	"fmt"
	"sync"

	"github.com/hewenyu/newapi-go/config"
)

// Client 是SDK的核心客户端结构
type Client struct {
	// config 存储客户端配置
	config *config.Config
	// mu 用于保护客户端的并发安全
	mu sync.RWMutex
}

// NewClient 创建一个新的客户端实例
func NewClient(options ...ClientOption) (*Client, error) {
	// 创建客户端实例并设置默认配置
	client := &Client{
		config: config.DefaultConfig(),
	}

	// 应用所有选项
	applyOptions(client, options)

	// 验证配置的有效性
	if err := client.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid client configuration: %w", err)
	}

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

	c.config = cfg.Clone()
	return nil
}

// Close 关闭客户端并清理资源
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 当前没有需要清理的资源
	// 在后续实现中可能需要关闭HTTP连接池等
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

// String 返回客户端的字符串表示
func (c *Client) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("NewAPIClient{BaseURL: %s, Debug: %t}",
		c.config.BaseURL, c.config.Debug)
}
