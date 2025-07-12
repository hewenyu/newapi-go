package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hewenyu/newapi-go/config"
	"github.com/hewenyu/newapi-go/internal/transport"
	"github.com/hewenyu/newapi-go/internal/utils"
	"github.com/hewenyu/newapi-go/services/chat"
	"github.com/hewenyu/newapi-go/types"
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
	// chatService 聊天服务
	chatService *chat.ChatService
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

	// 初始化聊天服务
	client.chatService = chat.NewChatService(client.transport, client.logger)

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

	// 重新初始化聊天服务
	c.chatService = chat.NewChatService(c.transport, c.logger)

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

	// 更新聊天服务的日志器
	if c.chatService != nil {
		c.chatService = chat.NewChatService(c.transport, c.logger)
	}
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

// GetChatService 获取聊天服务
func (c *Client) GetChatService() *chat.ChatService {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.chatService
}

// ========== 聊天服务代理方法 ==========

// CreateChatCompletion 创建聊天完成
func (c *Client) CreateChatCompletion(ctx context.Context, messages []types.ChatMessage, options ...chat.ChatOption) (*types.ChatCompletionResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.CreateChatCompletion(ctx, messages, options...)
}

// CreateChatCompletionStream 创建流式聊天完成
func (c *Client) CreateChatCompletionStream(ctx context.Context, messages []types.ChatMessage, options ...chat.ChatOption) (types.StreamResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.CreateChatCompletionStream(ctx, messages, options...)
}

// SimpleChat 简单聊天
func (c *Client) SimpleChat(ctx context.Context, message string, options ...chat.ChatOption) (*types.ChatCompletionResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.SimpleChat(ctx, message, options...)
}

// SimpleChatStream 简单流式聊天
func (c *Client) SimpleChatStream(ctx context.Context, message string, options ...chat.ChatOption) (types.StreamResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.SimpleChatStream(ctx, message, options...)
}

// ChatWithSystem 带系统消息的聊天
func (c *Client) ChatWithSystem(ctx context.Context, systemMessage, userMessage string, options ...chat.ChatOption) (*types.ChatCompletionResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.ChatWithSystem(ctx, systemMessage, userMessage, options...)
}

// ChatWithSystemStream 带系统消息的流式聊天
func (c *Client) ChatWithSystemStream(ctx context.Context, systemMessage, userMessage string, options ...chat.ChatOption) (types.StreamResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.ChatWithSystemStream(ctx, systemMessage, userMessage, options...)
}

// ChatWithHistory 带历史记录的聊天
func (c *Client) ChatWithHistory(ctx context.Context, userMessage string, history []types.ChatMessage, options ...chat.ChatOption) (*types.ChatCompletionResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.ChatWithHistory(ctx, userMessage, history, options...)
}

// ChatWithHistoryStream 带历史记录的流式聊天
func (c *Client) ChatWithHistoryStream(ctx context.Context, userMessage string, history []types.ChatMessage, options ...chat.ChatOption) (types.StreamResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}

	return c.chatService.ChatWithHistoryStream(ctx, userMessage, history, options...)
}

// ValidateMessage 验证消息
func (c *Client) ValidateMessage(message types.ChatMessage) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return fmt.Errorf("chat service not initialized")
	}

	return c.chatService.ValidateMessage(message)
}

// ValidateMessages 验证消息列表
func (c *Client) ValidateMessages(messages []types.ChatMessage) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return fmt.Errorf("chat service not initialized")
	}

	return c.chatService.ValidateMessages(messages)
}

// BuildConversation 构建对话
func (c *Client) BuildConversation(systemMessage string, userMessages []string) []types.ChatMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return nil
	}

	return c.chatService.BuildConversation(systemMessage, userMessages)
}

// CountTokens 计算Token数量
func (c *Client) CountTokens(messages []types.ChatMessage) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return 0
	}

	return c.chatService.CountTokens(messages)
}

// TruncateMessages 截断消息
func (c *Client) TruncateMessages(messages []types.ChatMessage, maxTokens int) []types.ChatMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chatService == nil {
		return messages
	}

	return c.chatService.TruncateMessages(messages, maxTokens)
}
