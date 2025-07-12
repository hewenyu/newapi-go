package types

import (
	"encoding/json"
	"fmt"
)

// 聊天角色常量
const (
	ChatRoleSystem    = "system"
	ChatRoleUser      = "user"
	ChatRoleAssistant = "assistant"
	ChatRoleFunction  = "function"
	ChatRoleTool      = "tool"
)

// 聊天消息类型常量
const (
	ChatMessageTypeText         = "text"
	ChatMessageTypeImageURL     = "image_url"
	ChatMessageTypeImageBase64  = "image_base64"
	ChatMessageTypeAudio        = "audio"
	ChatMessageTypeVideo        = "video"
	ChatMessageTypeToolCall     = "tool_call"
	ChatMessageTypeToolResponse = "tool_response"
)

// 工具调用类型常量
const (
	ToolCallTypeFunction = "function"
	ToolCallTypeBuiltin  = "builtin"
	ToolCallTypePlugin   = "plugin"
)

// 聊天完成选择结束原因常量
const (
	FinishReasonStop          = "stop"
	FinishReasonLength        = "length"
	FinishReasonContentFilter = "content_filter"
	FinishReasonToolCalls     = "tool_calls"
	FinishReasonFunctionCall  = "function_call"
)

// ChatMessage 聊天消息结构体
type ChatMessage struct {
	Role         string          `json:"role"`
	Content      interface{}     `json:"content"`
	Name         string          `json:"name,omitempty"`
	ToolCalls    []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID   string          `json:"tool_call_id,omitempty"`
	FunctionCall *FunctionCall   `json:"function_call,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// MessageContent 消息内容结构体
type MessageContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

// ToolCall 工具调用结构体
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用结构体
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionRequest 聊天完成请求结构体
type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []ChatMessage          `json:"messages"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	N                int                    `json:"n,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Stop             interface{}            `json:"stop,omitempty"`
	PresencePenalty  float64                `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64                `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float64     `json:"logit_bias,omitempty"`
	User             string                 `json:"user,omitempty"`
	Functions        []ChatFunction         `json:"functions,omitempty"`
	FunctionCall     interface{}            `json:"function_call,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	ResponseFormat   *ChatResponseFormat    `json:"response_format,omitempty"`
	Seed             int                    `json:"seed,omitempty"`
	LogProbs         bool                   `json:"logprobs,omitempty"`
	TopLogProbs      int                    `json:"top_logprobs,omitempty"`
	ExtraBody        map[string]interface{} `json:"-"`
}

// ChatCompletionResponse 聊天完成响应结构体
type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             Usage                  `json:"usage"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
	Error             *ErrorResponse         `json:"error,omitempty"`
}

// ChatCompletionChoice 聊天完成选择结构体
type ChatCompletionChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason"`
	LogProbs     *LogProbs    `json:"logprobs,omitempty"`
}

// ChatCompletionChunk 聊天完成流式响应块结构体
type ChatCompletionChunk struct {
	ID                string                      `json:"id"`
	Object            string                      `json:"object"`
	Created           int64                       `json:"created"`
	Model             string                      `json:"model"`
	Choices           []ChatCompletionChunkChoice `json:"choices"`
	Usage             *Usage                      `json:"usage,omitempty"`
	SystemFingerprint string                      `json:"system_fingerprint,omitempty"`
}

// ChatCompletionChunkChoice 聊天完成流式选择结构体
type ChatCompletionChunkChoice struct {
	Index        int         `json:"index"`
	Delta        ChatMessage `json:"delta"`
	FinishReason string      `json:"finish_reason"`
	LogProbs     *LogProbs   `json:"logprobs,omitempty"`
}

// ChatFunction 聊天函数结构体
type ChatFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Tool 工具结构体
type Tool struct {
	Type     string       `json:"type"`
	Function ChatFunction `json:"function"`
}

// ChatResponseFormat 聊天响应格式结构体
type ChatResponseFormat struct {
	Type   string `json:"type"`
	Schema string `json:"schema,omitempty"`
}

// LogProbs 日志概率结构体
type LogProbs struct {
	Tokens        []string                     `json:"tokens"`
	TokenLogprobs []float64                    `json:"token_logprobs"`
	TopLogprobs   []map[string]float64         `json:"top_logprobs"`
	TextOffset    []int                        `json:"text_offset"`
	Content       []ChatCompletionTokenLogprob `json:"content,omitempty"`
}

// ChatCompletionTokenLogprob Token日志概率结构体
type ChatCompletionTokenLogprob struct {
	Token       string       `json:"token"`
	Logprob     float64      `json:"logprob"`
	Bytes       []int        `json:"bytes,omitempty"`
	TopLogprobs []TopLogprob `json:"top_logprobs,omitempty"`
}

// TopLogprob 顶级日志概率结构体
type TopLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// NewChatMessage 创建新的聊天消息
func NewChatMessage(role, content string) ChatMessage {
	return ChatMessage{
		Role:    role,
		Content: content,
	}
}

// NewSystemMessage 创建系统消息
func NewSystemMessage(content string) ChatMessage {
	return NewChatMessage(ChatRoleSystem, content)
}

// NewUserMessage 创建用户消息
func NewUserMessage(content string) ChatMessage {
	return NewChatMessage(ChatRoleUser, content)
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage(content string) ChatMessage {
	return NewChatMessage(ChatRoleAssistant, content)
}

// NewFunctionMessage 创建函数消息
func NewFunctionMessage(name, content string) ChatMessage {
	return ChatMessage{
		Role:    ChatRoleFunction,
		Name:    name,
		Content: content,
	}
}

// NewToolMessage 创建工具消息
func NewToolMessage(toolCallID, content string) ChatMessage {
	return ChatMessage{
		Role:       ChatRoleTool,
		ToolCallID: toolCallID,
		Content:    content,
	}
}

// IsValidRole 检查角色是否有效
func (m *ChatMessage) IsValidRole() bool {
	switch m.Role {
	case ChatRoleSystem, ChatRoleUser, ChatRoleAssistant, ChatRoleFunction, ChatRoleTool:
		return true
	default:
		return false
	}
}

// GetTextContent 获取文本内容
func (m *ChatMessage) GetTextContent() string {
	switch content := m.Content.(type) {
	case string:
		return content
	case []MessageContent:
		for _, c := range content {
			if c.Type == ChatMessageTypeText {
				return c.Text
			}
		}
	}
	return ""
}

// HasToolCalls 检查是否有工具调用
func (m *ChatMessage) HasToolCalls() bool {
	return len(m.ToolCalls) > 0
}

// HasFunctionCall 检查是否有函数调用
func (m *ChatMessage) HasFunctionCall() bool {
	return m.FunctionCall != nil
}

// ToJSON 转换为JSON字符串
func (m *ChatMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从JSON字符串解析
func (m *ChatMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// ValidateParameters 验证请求参数
func (r *ChatCompletionRequest) ValidateParameters() error {
	if r.Model == "" {
		return NewValidationError("model", r.Model, "model is required", ErrCodeMissingParameter)
	}
	if len(r.Messages) == 0 {
		return NewValidationError("messages", r.Messages, "messages cannot be empty", ErrCodeMissingParameter)
	}
	if r.MaxTokens < 0 {
		return NewValidationError("max_tokens", r.MaxTokens, "max_tokens must be positive", ErrCodeInvalidParameter)
	}
	if r.Temperature < 0 || r.Temperature > 2 {
		return NewValidationError("temperature", r.Temperature, "temperature must be between 0 and 2", ErrCodeInvalidParameter)
	}
	if r.TopP < 0 || r.TopP > 1 {
		return NewValidationError("top_p", r.TopP, "top_p must be between 0 and 1", ErrCodeInvalidParameter)
	}
	if r.N < 1 {
		return NewValidationError("n", r.N, "n must be at least 1", ErrCodeInvalidParameter)
	}
	if r.PresencePenalty < -2 || r.PresencePenalty > 2 {
		return NewValidationError("presence_penalty", r.PresencePenalty, "presence_penalty must be between -2 and 2", ErrCodeInvalidParameter)
	}
	if r.FrequencyPenalty < -2 || r.FrequencyPenalty > 2 {
		return NewValidationError("frequency_penalty", r.FrequencyPenalty, "frequency_penalty must be between -2 and 2", ErrCodeInvalidParameter)
	}

	// 验证消息
	for i, msg := range r.Messages {
		if !msg.IsValidRole() {
			return NewValidationError(fmt.Sprintf("messages[%d].role", i), msg.Role, "invalid role", ErrCodeInvalidParameter)
		}
		if msg.Content == nil && !msg.HasToolCalls() && !msg.HasFunctionCall() {
			return NewValidationError(fmt.Sprintf("messages[%d].content", i), msg.Content, "content cannot be empty", ErrCodeMissingParameter)
		}
	}

	return nil
}

// SetDefaults 设置默认值
func (r *ChatCompletionRequest) SetDefaults() {
	if r.Temperature == 0 {
		r.Temperature = 1.0
	}
	if r.TopP == 0 {
		r.TopP = 1.0
	}
	if r.N == 0 {
		r.N = 1
	}
}

// IsStream 检查是否为流式请求
func (r *ChatCompletionRequest) IsStream() bool {
	return r.Stream
}

// GetMaxTokens 获取最大Token数
func (r *ChatCompletionRequest) GetMaxTokens() int {
	if r.MaxTokens > 0 {
		return r.MaxTokens
	}
	return 4096 // 默认值
}

// ToJSON 转换为JSON字符串
func (r *ChatCompletionRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ChatCompletionRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *ChatCompletionResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *ChatCompletionResponse) GetError() *ErrorResponse {
	return r.Error
}

// GetFirstChoice 获取第一个选择
func (r *ChatCompletionResponse) GetFirstChoice() *ChatCompletionChoice {
	if len(r.Choices) > 0 {
		return &r.Choices[0]
	}
	return nil
}

// GetFirstMessage 获取第一个消息
func (r *ChatCompletionResponse) GetFirstMessage() *ChatMessage {
	if choice := r.GetFirstChoice(); choice != nil {
		return &choice.Message
	}
	return nil
}

// GetFirstContent 获取第一个内容
func (r *ChatCompletionResponse) GetFirstContent() string {
	if msg := r.GetFirstMessage(); msg != nil {
		return msg.GetTextContent()
	}
	return ""
}

// ToJSON 转换为JSON字符串
func (r *ChatCompletionResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ChatCompletionResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsFinished 检查是否完成
func (c *ChatCompletionChoice) IsFinished() bool {
	return c.FinishReason != ""
}

// IsToolCall 检查是否为工具调用
func (c *ChatCompletionChoice) IsToolCall() bool {
	return c.FinishReason == FinishReasonToolCalls
}

// IsFunctionCall 检查是否为函数调用
func (c *ChatCompletionChoice) IsFunctionCall() bool {
	return c.FinishReason == FinishReasonFunctionCall
}

// GetContent 获取内容
func (c *ChatCompletionChoice) GetContent() string {
	return c.Message.GetTextContent()
}

// IsFinished 检查是否完成
func (c *ChatCompletionChunkChoice) IsFinished() bool {
	return c.FinishReason != ""
}

// GetContent 获取内容
func (c *ChatCompletionChunkChoice) GetContent() string {
	return c.Delta.GetTextContent()
}

// IsValidTool 检查工具是否有效
func (t *Tool) IsValidTool() bool {
	return t.Type != "" && t.Function.Name != ""
}

// IsValidFunction 检查函数是否有效
func (f *ChatFunction) IsValidFunction() bool {
	return f.Name != ""
}

// IsValidToolCall 检查工具调用是否有效
func (tc *ToolCall) IsValidToolCall() bool {
	return tc.ID != "" && tc.Type != "" && tc.Function.Name != ""
}

// IsValidFunctionCall 检查函数调用是否有效
func (fc *FunctionCall) IsValidFunctionCall() bool {
	return fc.Name != ""
}
