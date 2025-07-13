package converter

import (
	"context"
	"fmt"
	"strings"

	claudeTypes "github.com/hewenyu/newapi-go/proxy/types"
	"github.com/hewenyu/newapi-go/services/chat"
	"github.com/hewenyu/newapi-go/types"
)

// ClaudeToNewAPIConverter Claude到NewAPI转换器
type ClaudeToNewAPIConverter struct {
	modelMapping map[string]string
}

// NewClaudeToNewAPIConverter 创建新的转换器
func NewClaudeToNewAPIConverter() *ClaudeToNewAPIConverter {
	return &ClaudeToNewAPIConverter{
		modelMapping: getModelMapping(),
	}
}

// getModelMapping 获取模型映射
func getModelMapping() map[string]string {
	return map[string]string{
		// Claude 3 系列
		"claude-3-opus-20240229":     "claude-3-opus-20240229",
		"claude-3-sonnet-20240229":   "claude-3-sonnet-20240229",
		"claude-3-haiku-20240307":    "claude-3-haiku-20240307",
		"claude-3-5-sonnet-20241022": "claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022":  "claude-3-5-haiku-20241022",

		// 默认映射
		"claude-3-opus":     "claude-3-opus-20240229",
		"claude-3-sonnet":   "claude-3-sonnet-20240229",
		"claude-3-haiku":    "claude-3-haiku-20240307",
		"claude-3.5-sonnet": "claude-3-5-sonnet-20241022",
		"claude-3.5-haiku":  "claude-3-5-haiku-20241022",

		// 简化映射
		"opus":   "claude-3-opus-20240229",
		"sonnet": "claude-3-sonnet-20240229",
		"haiku":  "claude-3-haiku-20240307",
	}
}

// ConvertRequest 转换请求
func (c *ClaudeToNewAPIConverter) ConvertRequest(claudeReq *claudeTypes.ClaudeRequest) ([]types.ChatMessage, []chat.ChatOption, error) {
	if claudeReq == nil {
		return nil, nil, fmt.Errorf("claude request is nil")
	}

	// 验证请求
	if err := claudeReq.Validate(); err != nil {
		return nil, nil, err
	}

	// 转换消息
	messages, err := c.convertMessages(claudeReq.Messages, claudeReq.System)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// 转换选项
	options, err := c.convertOptions(claudeReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert options: %w", err)
	}

	return messages, options, nil
}

// convertMessages 转换消息
func (c *ClaudeToNewAPIConverter) convertMessages(claudeMessages []claudeTypes.ClaudeMessage, systemMessage string) ([]types.ChatMessage, error) {
	var messages []types.ChatMessage

	// 添加系统消息
	if systemMessage != "" {
		messages = append(messages, types.NewSystemMessage(systemMessage))
	}

	// 转换对话消息
	for _, claudeMsg := range claudeMessages {
		newAPIMsg, err := c.convertMessage(claudeMsg)
		if err != nil {
			return nil, err
		}
		messages = append(messages, newAPIMsg)
	}

	return messages, nil
}

// convertMessage 转换单个消息
func (c *ClaudeToNewAPIConverter) convertMessage(claudeMsg claudeTypes.ClaudeMessage) (types.ChatMessage, error) {
	// 转换角色
	var role string
	switch claudeMsg.Role {
	case claudeTypes.RoleUser:
		role = types.ChatRoleUser
	case claudeTypes.RoleAssistant:
		role = types.ChatRoleAssistant
	case claudeTypes.RoleSystem:
		role = types.ChatRoleSystem
	default:
		return types.ChatMessage{}, fmt.Errorf("unsupported role: %s", claudeMsg.Role)
	}

	// 转换内容
	content, err := c.convertContent(claudeMsg.Content)
	if err != nil {
		return types.ChatMessage{}, err
	}

	return types.ChatMessage{
		Role:    role,
		Content: content,
	}, nil
}

// convertContent 转换内容
func (c *ClaudeToNewAPIConverter) convertContent(claudeContent []claudeTypes.ContentItem) (interface{}, error) {
	if len(claudeContent) == 0 {
		return "", fmt.Errorf("content cannot be empty")
	}

	// 如果只有一个文本内容，直接返回文本
	if len(claudeContent) == 1 && claudeContent[0].Type == claudeTypes.ContentTypeText {
		return claudeContent[0].Text, nil
	}

	// 多个内容或包含图像，转换为消息内容数组
	var messageContents []types.MessageContent
	for _, item := range claudeContent {
		switch item.Type {
		case claudeTypes.ContentTypeText:
			messageContents = append(messageContents, types.MessageContent{
				Type: types.ChatMessageTypeText,
				Text: item.Text,
			})
		case claudeTypes.ContentTypeImage:
			// 支持图像URL
			if item.ImageURL != "" {
				messageContents = append(messageContents, types.MessageContent{
					Type:     types.ChatMessageTypeImageURL,
					ImageURL: item.ImageURL,
				})
			}
		default:
			return nil, fmt.Errorf("unsupported content type: %s", item.Type)
		}
	}

	return messageContents, nil
}

// convertOptions 转换选项
func (c *ClaudeToNewAPIConverter) convertOptions(claudeReq *claudeTypes.ClaudeRequest) ([]chat.ChatOption, error) {
	var options []chat.ChatOption

	// 模型映射
	if claudeReq.Model != "" {
		model := c.mapModel(claudeReq.Model)
		options = append(options, chat.WithModel(model))
	}

	// 最大Token数
	if claudeReq.MaxTokens > 0 {
		options = append(options, chat.WithMaxTokens(claudeReq.MaxTokens))
	}

	// 温度
	if claudeReq.Temperature > 0 {
		options = append(options, chat.WithTemperature(claudeReq.Temperature))
	}

	// Top-P
	if claudeReq.TopP > 0 {
		options = append(options, chat.WithTopP(claudeReq.TopP))
	}

	// 停止序列
	if len(claudeReq.StopSequences) > 0 {
		options = append(options, chat.WithStop(claudeReq.StopSequences))
	}

	// 流式处理
	if claudeReq.Stream {
		options = append(options, chat.WithStream(true))
	}

	return options, nil
}

// mapModel 映射模型名称
func (c *ClaudeToNewAPIConverter) mapModel(claudeModel string) string {
	// 首先检查直接映射
	if mapped, exists := c.modelMapping[claudeModel]; exists {
		return mapped
	}

	// 尝试模糊匹配
	claudeModel = strings.ToLower(claudeModel)
	for pattern, mapped := range c.modelMapping {
		if strings.Contains(claudeModel, strings.ToLower(pattern)) {
			return mapped
		}
	}

	// 如果没有映射，返回原始模型名
	return claudeModel
}

// ConvertContext 转换上下文
func (c *ClaudeToNewAPIConverter) ConvertContext(ctx context.Context, claudeReq *claudeTypes.ClaudeRequest) context.Context {
	// 可以在这里添加请求相关的上下文信息
	if claudeReq.Metadata != nil && claudeReq.Metadata.UserID != "" {
		ctx = context.WithValue(ctx, "user_id", claudeReq.Metadata.UserID)
	}

	return ctx
}

// ValidateRequest 验证请求
func (c *ClaudeToNewAPIConverter) ValidateRequest(claudeReq *claudeTypes.ClaudeRequest) error {
	if claudeReq == nil {
		return claudeTypes.NewInvalidRequestError("request is nil")
	}

	if claudeReq.Model == "" {
		return claudeTypes.NewInvalidRequestError("model is required")
	}

	if claudeReq.MaxTokens <= 0 {
		return claudeTypes.NewInvalidRequestError("max_tokens must be positive")
	}

	if claudeReq.MaxTokens > 200000 {
		return claudeTypes.NewInvalidRequestError("max_tokens cannot exceed 200000")
	}

	if len(claudeReq.Messages) == 0 {
		return claudeTypes.NewInvalidRequestError("messages cannot be empty")
	}

	if claudeReq.Temperature < 0 || claudeReq.Temperature > 1 {
		return claudeTypes.NewInvalidRequestError("temperature must be between 0 and 1")
	}

	if claudeReq.TopP < 0 || claudeReq.TopP > 1 {
		return claudeTypes.NewInvalidRequestError("top_p must be between 0 and 1")
	}

	if claudeReq.TopK < 0 {
		return claudeTypes.NewInvalidRequestError("top_k must be non-negative")
	}

	// 验证消息
	for i, msg := range claudeReq.Messages {
		if err := c.validateMessage(msg, i); err != nil {
			return err
		}
	}

	return nil
}

// validateMessage 验证消息
func (c *ClaudeToNewAPIConverter) validateMessage(msg claudeTypes.ClaudeMessage, index int) error {
	if msg.Role != claudeTypes.RoleUser && msg.Role != claudeTypes.RoleAssistant {
		return claudeTypes.NewInvalidRequestError(fmt.Sprintf("invalid role at message %d: %s", index, msg.Role))
	}

	if len(msg.Content) == 0 {
		return claudeTypes.NewInvalidRequestError(fmt.Sprintf("message %d content cannot be empty", index))
	}

	for j, content := range msg.Content {
		if err := c.validateContent(content, index, j); err != nil {
			return err
		}
	}

	return nil
}

// validateContent 验证内容
func (c *ClaudeToNewAPIConverter) validateContent(content claudeTypes.ContentItem, msgIndex, contentIndex int) error {
	switch content.Type {
	case claudeTypes.ContentTypeText:
		if content.Text == "" {
			return claudeTypes.NewInvalidRequestError(fmt.Sprintf("text content cannot be empty at message %d, content %d", msgIndex, contentIndex))
		}
	case claudeTypes.ContentTypeImage:
		if content.ImageURL == "" && content.Source == nil {
			return claudeTypes.NewInvalidRequestError(fmt.Sprintf("image content must have either image_url or source at message %d, content %d", msgIndex, contentIndex))
		}
	default:
		return claudeTypes.NewInvalidRequestError(fmt.Sprintf("unsupported content type: %s at message %d, content %d", content.Type, msgIndex, contentIndex))
	}

	return nil
}

// GetSupportedModels 获取支持的模型列表
func (c *ClaudeToNewAPIConverter) GetSupportedModels() []string {
	var models []string
	for model := range c.modelMapping {
		models = append(models, model)
	}
	return models
}

// IsModelSupported 检查模型是否支持
func (c *ClaudeToNewAPIConverter) IsModelSupported(model string) bool {
	_, exists := c.modelMapping[model]
	return exists
}
