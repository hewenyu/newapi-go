package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/hewenyu/newapi-go/internal/transport"
	"github.com/hewenyu/newapi-go/internal/utils"
	"github.com/hewenyu/newapi-go/types"
	"go.uber.org/zap"
)

// ChatService 聊天服务结构体
type ChatService struct {
	transport transport.HTTPTransport
	logger    utils.Logger
	config    *ChatConfig
	mu        sync.RWMutex
}

// NewChatService 创建新的聊天服务实例
func NewChatService(transport transport.HTTPTransport, logger utils.Logger, options ...ChatOption) *ChatService {
	config := DefaultChatConfig()

	// 应用选项
	for _, option := range options {
		option(config)
	}

	return &ChatService{
		transport: transport,
		logger:    logger,
		config:    config,
	}
}

// parseJSONResponse 解析JSON响应
func parseJSONResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// CreateChatCompletion 创建聊天完成
func (s *ChatService) CreateChatCompletion(ctx context.Context, messages []types.ChatMessage, options ...ChatOption) (*types.ChatCompletionResponse, error) {
	// 验证输入
	if len(messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty")
	}

	// 创建配置副本并应用选项
	config := s.getConfig()
	for _, option := range options {
		option(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid chat config: %w", err)
	}

	// 构建请求
	req := config.ToRequest(messages)

	// 确保不是流式请求
	req.Stream = false

	// 发送请求
	resp, err := s.transport.Post(ctx, "/v1/chat/completions", req)
	if err != nil {
		s.logger.Error("Failed to create chat completion", zap.Error(err))
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// 解析响应
	var chatResp types.ChatCompletionResponse
	if err := parseJSONResponse(resp, &chatResp); err != nil {
		s.logger.Error("Failed to parse chat completion response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查API错误
	if chatResp.IsError() {
		apiErr := chatResp.GetError()
		s.logger.Error("API returned error", zap.String("error", apiErr.Message))
		return nil, fmt.Errorf("API error: %s", apiErr.Message)
	}

	s.logger.Debug("Chat completion created successfully", zap.String("id", chatResp.ID))
	return &chatResp, nil
}

// CreateChatCompletionStream 创建流式聊天完成
func (s *ChatService) CreateChatCompletionStream(ctx context.Context, messages []types.ChatMessage, options ...ChatOption) (types.StreamResponse, error) {
	// 验证输入
	if len(messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty")
	}

	// 创建配置副本并应用选项
	config := s.getConfig()
	for _, option := range options {
		option(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid chat config: %w", err)
	}

	// 构建请求
	req := config.ToRequest(messages)

	// 确保是流式请求
	req.Stream = true

	// 发送流式请求
	streamReader, err := s.transport.PostStream(ctx, "/v1/chat/completions", req)
	if err != nil {
		s.logger.Error("Failed to create chat completion stream", zap.Error(err))
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}

	// 创建适配器来桥接transport.StreamReader和types.StreamResponse
	adapter := &streamReaderAdapter{
		reader: streamReader,
		ctx:    ctx,
	}

	// 创建流式处理器
	streamProcessor := NewChatStreamProcessor(adapter, s.logger)

	s.logger.Debug("Chat completion stream created successfully")
	return streamProcessor, nil
}

// streamReaderAdapter 适配器，将transport.StreamReader适配为types.StreamResponse
type streamReaderAdapter struct {
	reader transport.StreamReader
	ctx    context.Context
}

// Next 获取下一个事件
func (a *streamReaderAdapter) Next() (*types.StreamEvent, error) {
	data, err := a.reader.Read()
	if err != nil {
		return nil, err
	}

	// 将data转换为JSON
	jsonData, marshalErr := json.Marshal(data)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal stream data: %w", marshalErr)
	}

	return &types.StreamEvent{
		Type: types.StreamEventTypeData,
		Data: json.RawMessage(jsonData),
	}, nil
}

// Close 关闭流
func (a *streamReaderAdapter) Close() error {
	return a.reader.Close()
}

// Err 获取错误
func (a *streamReaderAdapter) Err() error {
	return a.reader.Err()
}

// Done 检查是否完成
func (a *streamReaderAdapter) Done() bool {
	return false // transport.StreamReader没有Done方法，返回false
}

// Context 获取上下文
func (a *streamReaderAdapter) Context() context.Context {
	return a.ctx
}

// ChatWithHistory 带历史记录的聊天
func (s *ChatService) ChatWithHistory(ctx context.Context, userMessage string, history []types.ChatMessage, options ...ChatOption) (*types.ChatCompletionResponse, error) {
	// 构建消息列表
	messages := make([]types.ChatMessage, 0, len(history)+1)
	messages = append(messages, history...)
	messages = append(messages, types.NewUserMessage(userMessage))

	return s.CreateChatCompletion(ctx, messages, options...)
}

// ChatWithHistoryStream 带历史记录的流式聊天
func (s *ChatService) ChatWithHistoryStream(ctx context.Context, userMessage string, history []types.ChatMessage, options ...ChatOption) (types.StreamResponse, error) {
	// 构建消息列表
	messages := make([]types.ChatMessage, 0, len(history)+1)
	messages = append(messages, history...)
	messages = append(messages, types.NewUserMessage(userMessage))

	return s.CreateChatCompletionStream(ctx, messages, options...)
}

// SimpleChat 简单聊天
func (s *ChatService) SimpleChat(ctx context.Context, userMessage string, options ...ChatOption) (*types.ChatCompletionResponse, error) {
	messages := []types.ChatMessage{
		types.NewUserMessage(userMessage),
	}

	return s.CreateChatCompletion(ctx, messages, options...)
}

// SimpleChatStream 简单流式聊天
func (s *ChatService) SimpleChatStream(ctx context.Context, userMessage string, options ...ChatOption) (types.StreamResponse, error) {
	messages := []types.ChatMessage{
		types.NewUserMessage(userMessage),
	}

	return s.CreateChatCompletionStream(ctx, messages, options...)
}

// ChatWithSystem 带系统消息的聊天
func (s *ChatService) ChatWithSystem(ctx context.Context, systemMessage, userMessage string, options ...ChatOption) (*types.ChatCompletionResponse, error) {
	messages := []types.ChatMessage{
		types.NewSystemMessage(systemMessage),
		types.NewUserMessage(userMessage),
	}

	return s.CreateChatCompletion(ctx, messages, options...)
}

// ChatWithSystemStream 带系统消息的流式聊天
func (s *ChatService) ChatWithSystemStream(ctx context.Context, systemMessage, userMessage string, options ...ChatOption) (types.StreamResponse, error) {
	messages := []types.ChatMessage{
		types.NewSystemMessage(systemMessage),
		types.NewUserMessage(userMessage),
	}

	return s.CreateChatCompletionStream(ctx, messages, options...)
}

// UpdateConfig 更新配置
func (s *ChatService) UpdateConfig(options ...ChatOption) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, option := range options {
		option(s.config)
	}
}

// GetConfig 获取配置副本
func (s *ChatService) GetConfig() *ChatConfig {
	return s.getConfig()
}

// getConfig 获取配置副本（内部使用）
func (s *ChatService) getConfig() *ChatConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config.Clone()
}

// ValidateMessage 验证消息
func (s *ChatService) ValidateMessage(message types.ChatMessage) error {
	if !message.IsValidRole() {
		return fmt.Errorf("invalid message role: %s", message.Role)
	}

	if message.GetTextContent() == "" && len(message.ToolCalls) == 0 && message.FunctionCall == nil {
		return fmt.Errorf("message must have content, tool calls, or function call")
	}

	return nil
}

// ValidateMessages 验证消息列表
func (s *ChatService) ValidateMessages(messages []types.ChatMessage) error {
	if len(messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	for i, message := range messages {
		if err := s.ValidateMessage(message); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}

	return nil
}

// BuildConversation 构建对话
func (s *ChatService) BuildConversation(systemMessage string, userMessages []string) []types.ChatMessage {
	messages := make([]types.ChatMessage, 0, len(userMessages)+1)

	if systemMessage != "" {
		messages = append(messages, types.NewSystemMessage(systemMessage))
	}

	for _, userMessage := range userMessages {
		messages = append(messages, types.NewUserMessage(userMessage))
	}

	return messages
}

// ExtractAssistantMessages 提取助手消息
func (s *ChatService) ExtractAssistantMessages(messages []types.ChatMessage) []types.ChatMessage {
	var assistantMessages []types.ChatMessage

	for _, message := range messages {
		if message.Role == types.ChatRoleAssistant {
			assistantMessages = append(assistantMessages, message)
		}
	}

	return assistantMessages
}

// GetLastAssistantMessage 获取最后一条助手消息
func (s *ChatService) GetLastAssistantMessage(messages []types.ChatMessage) *types.ChatMessage {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == types.ChatRoleAssistant {
			return &messages[i]
		}
	}

	return nil
}

// CountTokens 计算Token数量（简单估算）
func (s *ChatService) CountTokens(messages []types.ChatMessage) int {
	totalTokens := 0

	for _, message := range messages {
		// 简单估算：每个字符约0.25个token
		content := message.GetTextContent()
		totalTokens += len(content) / 4

		// 角色和结构的开销
		totalTokens += 10
	}

	return totalTokens
}

// TruncateMessages 截断消息以适应Token限制
func (s *ChatService) TruncateMessages(messages []types.ChatMessage, maxTokens int) []types.ChatMessage {
	if len(messages) == 0 {
		return messages
	}

	// 保留系统消息
	var systemMessages []types.ChatMessage
	var otherMessages []types.ChatMessage

	for _, message := range messages {
		if message.Role == types.ChatRoleSystem {
			systemMessages = append(systemMessages, message)
		} else {
			otherMessages = append(otherMessages, message)
		}
	}

	// 计算系统消息的Token数量
	systemTokens := s.CountTokens(systemMessages)
	availableTokens := maxTokens - systemTokens

	if availableTokens <= 0 {
		return systemMessages
	}

	// 从最新的消息开始保留
	result := make([]types.ChatMessage, 0, len(messages))
	result = append(result, systemMessages...)

	currentTokens := 0
	for i := len(otherMessages) - 1; i >= 0; i-- {
		messageTokens := s.CountTokens([]types.ChatMessage{otherMessages[i]})
		if currentTokens+messageTokens > availableTokens {
			break
		}

		result = append([]types.ChatMessage{otherMessages[i]}, result...)
		currentTokens += messageTokens
	}

	return result
}
