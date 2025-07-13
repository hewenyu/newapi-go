package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Claude API 常量定义
const (
	ClaudeAPIVersion = "2023-06-01"
	DefaultModel     = "claude-3-sonnet-20240229"
)

// 消息角色常量
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

// 内容类型常量
const (
	ContentTypeText  = "text"
	ContentTypeImage = "image"
)

// 停止原因常量
const (
	StopReasonEndTurn      = "end_turn"
	StopReasonMaxTokens    = "max_tokens"
	StopReasonStopSequence = "stop_sequence"
	StopReasonToolUse      = "tool_use"
)

// 流式事件类型常量
const (
	EventMessageStart      = "message_start"
	EventMessageDelta      = "message_delta"
	EventMessageStop       = "message_stop"
	EventContentBlockStart = "content_block_start"
	EventContentBlockDelta = "content_block_delta"
	EventContentBlockStop  = "content_block_stop"
	EventPing              = "ping"
	EventError             = "error"
)

// ClaudeMessage Claude API消息结构
type ClaudeMessage struct {
	Role    string        `json:"role"`
	Content []ContentItem `json:"content"`
}

// ContentItem 内容项
type ContentItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Source   *Image `json:"source,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// Image 图像信息
type Image struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// ClaudeRequest Claude API请求结构
type ClaudeRequest struct {
	Model         string          `json:"model"`
	MaxTokens     int             `json:"max_tokens"`
	Messages      []ClaudeMessage `json:"messages"`
	System        string          `json:"system,omitempty"`
	Temperature   float64         `json:"temperature,omitempty"`
	TopP          float64         `json:"top_p,omitempty"`
	TopK          int             `json:"top_k,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
	Stream        bool            `json:"stream,omitempty"`
	Metadata      *Metadata       `json:"metadata,omitempty"`
}

// Metadata 元数据
type Metadata struct {
	UserID string `json:"user_id,omitempty"`
}

// ClaudeResponse Claude API响应结构
type ClaudeResponse struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	Content      []ContentItem `json:"content"`
	Model        string        `json:"model"`
	StopReason   string        `json:"stop_reason,omitempty"`
	StopSequence string        `json:"stop_sequence,omitempty"`
	Usage        Usage         `json:"usage"`
}

// Usage 使用量信息
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ClaudeError Claude API错误结构
type ClaudeError struct {
	Type        string      `json:"type"`
	ErrorDetail ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// 流式响应相关结构

// StreamEvent 流式事件
type StreamEvent struct {
	Type  string          `json:"type"`
	Event string          `json:"event,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// MessageStartEvent 消息开始事件
type MessageStartEvent struct {
	Type    string         `json:"type"`
	Message MessageContent `json:"message"`
}

// MessageContent 消息内容
type MessageContent struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	Content      []ContentItem `json:"content"`
	Model        string        `json:"model"`
	StopReason   string        `json:"stop_reason,omitempty"`
	StopSequence string        `json:"stop_sequence,omitempty"`
	Usage        Usage         `json:"usage"`
}

// MessageDeltaEvent 消息增量事件
type MessageDeltaEvent struct {
	Type  string       `json:"type"`
	Delta MessageDelta `json:"delta"`
	Usage Usage        `json:"usage,omitempty"`
}

// MessageDelta 消息增量
type MessageDelta struct {
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// MessageStopEvent 消息停止事件
type MessageStopEvent struct {
	Type string `json:"type"`
}

// ContentBlockStartEvent 内容块开始事件
type ContentBlockStartEvent struct {
	Type         string      `json:"type"`
	Index        int         `json:"index"`
	ContentBlock ContentItem `json:"content_block"`
}

// ContentBlockDeltaEvent 内容块增量事件
type ContentBlockDeltaEvent struct {
	Type  string            `json:"type"`
	Index int               `json:"index"`
	Delta ContentBlockDelta `json:"delta"`
}

// ContentBlockDelta 内容块增量
type ContentBlockDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ContentBlockStopEvent 内容块停止事件
type ContentBlockStopEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// PingEvent Ping事件
type PingEvent struct {
	Type string `json:"type"`
}

// ErrorEvent 错误事件
type ErrorEvent struct {
	Type        string      `json:"type"`
	ErrorDetail ErrorDetail `json:"error"`
}

// 辅助函数

// NewTextContent 创建文本内容
func NewTextContent(text string) ContentItem {
	return ContentItem{
		Type: ContentTypeText,
		Text: text,
	}
}

// NewImageContent 创建图像内容
func NewImageContent(imageURL string) ContentItem {
	return ContentItem{
		Type:     ContentTypeImage,
		ImageURL: imageURL,
	}
}

// NewUserMessage 创建用户消息
func NewUserMessage(text string) ClaudeMessage {
	return ClaudeMessage{
		Role:    RoleUser,
		Content: []ContentItem{NewTextContent(text)},
	}
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage(text string) ClaudeMessage {
	return ClaudeMessage{
		Role:    RoleAssistant,
		Content: []ContentItem{NewTextContent(text)},
	}
}

// GetTextContent 获取消息的文本内容
func (m *ClaudeMessage) GetTextContent() string {
	for _, content := range m.Content {
		if content.Type == ContentTypeText {
			return content.Text
		}
	}
	return ""
}

// HasTextContent 检查是否包含文本内容
func (m *ClaudeMessage) HasTextContent() bool {
	for _, content := range m.Content {
		if content.Type == ContentTypeText && content.Text != "" {
			return true
		}
	}
	return false
}

// HasImageContent 检查是否包含图像内容
func (m *ClaudeMessage) HasImageContent() bool {
	for _, content := range m.Content {
		if content.Type == ContentTypeImage {
			return true
		}
	}
	return false
}

// Validate 验证请求参数
func (r *ClaudeRequest) Validate() error {
	if r.Model == "" {
		return &ClaudeError{
			Type: "error",
			ErrorDetail: ErrorDetail{
				Type:    "invalid_request_error",
				Message: "model is required",
			},
		}
	}

	if r.MaxTokens <= 0 {
		return &ClaudeError{
			Type: "error",
			ErrorDetail: ErrorDetail{
				Type:    "invalid_request_error",
				Message: "max_tokens must be positive",
			},
		}
	}

	if len(r.Messages) == 0 {
		return &ClaudeError{
			Type: "error",
			ErrorDetail: ErrorDetail{
				Type:    "invalid_request_error",
				Message: "messages cannot be empty",
			},
		}
	}

	// 验证消息格式
	for i, msg := range r.Messages {
		if msg.Role != RoleUser && msg.Role != RoleAssistant {
			return &ClaudeError{
				Type: "error",
				ErrorDetail: ErrorDetail{
					Type:    "invalid_request_error",
					Message: fmt.Sprintf("invalid role at message %d: %s", i, msg.Role),
				},
			}
		}

		if len(msg.Content) == 0 {
			return &ClaudeError{
				Type: "error",
				ErrorDetail: ErrorDetail{
					Type:    "invalid_request_error",
					Message: fmt.Sprintf("message %d content cannot be empty", i),
				},
			}
		}
	}

	return nil
}

// SetDefaults 设置默认值
func (r *ClaudeRequest) SetDefaults() {
	if r.Model == "" {
		r.Model = DefaultModel
	}
	if r.Temperature == 0 {
		r.Temperature = 1.0
	}
	if r.TopP == 0 {
		r.TopP = 1.0
	}
}

// ToJSON 转换为JSON
func (r *ClaudeRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON解析
func (r *ClaudeRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ToJSON 转换为JSON
func (r *ClaudeResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON解析
func (r *ClaudeResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// GetFirstTextContent 获取第一个文本内容
func (r *ClaudeResponse) GetFirstTextContent() string {
	for _, content := range r.Content {
		if content.Type == ContentTypeText {
			return content.Text
		}
	}
	return ""
}

// GenerateID 生成消息ID
func GenerateID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// Error 实现error接口
func (e *ClaudeError) Error() string {
	return e.ErrorDetail.Message
}

// NewClaudeError 创建Claude错误
func NewClaudeError(errorType, message string) *ClaudeError {
	return &ClaudeError{
		Type: "error",
		ErrorDetail: ErrorDetail{
			Type:    errorType,
			Message: message,
		},
	}
}

// NewInvalidRequestError 创建无效请求错误
func NewInvalidRequestError(message string) *ClaudeError {
	return NewClaudeError("invalid_request_error", message)
}

// NewAuthenticationError 创建认证错误
func NewAuthenticationError(message string) *ClaudeError {
	return NewClaudeError("authentication_error", message)
}

// NewAPIError 创建API错误
func NewAPIError(message string) *ClaudeError {
	return NewClaudeError("api_error", message)
}

// NewRateLimitError 创建限流错误
func NewRateLimitError(message string) *ClaudeError {
	return NewClaudeError("rate_limit_error", message)
}
