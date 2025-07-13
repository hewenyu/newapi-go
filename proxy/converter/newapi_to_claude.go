package converter

import (
	"encoding/json"
	"fmt"
	"time"

	claudeTypes "github.com/hewenyu/newapi-go/proxy/types"
	"github.com/hewenyu/newapi-go/types"
)

// NewAPIToClaudeConverter NewAPI到Claude转换器
type NewAPIToClaudeConverter struct {
	stopReasonMapping map[string]string
}

// NewNewAPIToClaudeConverter 创建新的转换器
func NewNewAPIToClaudeConverter() *NewAPIToClaudeConverter {
	return &NewAPIToClaudeConverter{
		stopReasonMapping: getStopReasonMapping(),
	}
}

// getStopReasonMapping 获取停止原因映射
func getStopReasonMapping() map[string]string {
	return map[string]string{
		types.FinishReasonStop:          claudeTypes.StopReasonEndTurn,
		types.FinishReasonLength:        claudeTypes.StopReasonMaxTokens,
		types.FinishReasonContentFilter: claudeTypes.StopReasonEndTurn,
		types.FinishReasonToolCalls:     claudeTypes.StopReasonToolUse,
		types.FinishReasonFunctionCall:  claudeTypes.StopReasonToolUse,
		"":                              claudeTypes.StopReasonEndTurn,
	}
}

// ConvertResponse 转换响应
func (c *NewAPIToClaudeConverter) ConvertResponse(newAPIResp *types.ChatCompletionResponse, originalModel string) (*claudeTypes.ClaudeResponse, error) {
	if newAPIResp == nil {
		return nil, fmt.Errorf("newapi response is nil")
	}

	// 检查错误
	if newAPIResp.IsError() {
		return nil, c.convertError(newAPIResp.GetError())
	}

	// 获取第一个选择
	choice := newAPIResp.GetFirstChoice()
	if choice == nil {
		return nil, fmt.Errorf("no choices in response")
	}

	// 转换响应
	claudeResp := &claudeTypes.ClaudeResponse{
		ID:           newAPIResp.ID,
		Type:         "message",
		Role:         claudeTypes.RoleAssistant,
		Content:      c.convertContent(choice.Message),
		Model:        originalModel,
		StopReason:   c.mapStopReason(choice.FinishReason),
		StopSequence: "",
		Usage: claudeTypes.Usage{
			InputTokens:  newAPIResp.Usage.PromptTokens,
			OutputTokens: newAPIResp.Usage.CompletionTokens,
		},
	}

	return claudeResp, nil
}

// convertContent 转换内容
func (c *NewAPIToClaudeConverter) convertContent(message types.ChatMessage) []claudeTypes.ContentItem {
	var content []claudeTypes.ContentItem

	switch msgContent := message.Content.(type) {
	case string:
		// 简单文本内容
		if msgContent != "" {
			content = append(content, claudeTypes.ContentItem{
				Type: claudeTypes.ContentTypeText,
				Text: msgContent,
			})
		}
	case []types.MessageContent:
		// 复杂内容数组
		for _, item := range msgContent {
			switch item.Type {
			case types.ChatMessageTypeText:
				if item.Text != "" {
					content = append(content, claudeTypes.ContentItem{
						Type: claudeTypes.ContentTypeText,
						Text: item.Text,
					})
				}
			case types.ChatMessageTypeImageURL:
				if item.ImageURL != "" {
					content = append(content, claudeTypes.ContentItem{
						Type:     claudeTypes.ContentTypeImage,
						ImageURL: item.ImageURL,
					})
				}
			}
		}
	default:
		// 尝试转换为字符串
		if str := message.GetTextContent(); str != "" {
			content = append(content, claudeTypes.ContentItem{
				Type: claudeTypes.ContentTypeText,
				Text: str,
			})
		}
	}

	// 如果没有内容，添加空文本
	if len(content) == 0 {
		content = append(content, claudeTypes.ContentItem{
			Type: claudeTypes.ContentTypeText,
			Text: "",
		})
	}

	return content
}

// mapStopReason 映射停止原因
func (c *NewAPIToClaudeConverter) mapStopReason(finishReason string) string {
	if mapped, exists := c.stopReasonMapping[finishReason]; exists {
		return mapped
	}
	return claudeTypes.StopReasonEndTurn
}

// convertError 转换错误
func (c *NewAPIToClaudeConverter) convertError(err *types.ErrorResponse) error {
	if err == nil {
		return claudeTypes.NewAPIError("unknown error")
	}

	// 根据错误类型映射
	switch err.Type {
	case types.ErrTypeInvalidRequest:
		return claudeTypes.NewInvalidRequestError(err.Message)
	case types.ErrTypeAuthentication:
		return claudeTypes.NewAuthenticationError(err.Message)
	case types.ErrTypeRateLimit:
		return claudeTypes.NewRateLimitError(err.Message)
	default:
		return claudeTypes.NewAPIError(err.Message)
	}
}

// ConvertStreamChunk 转换流式响应块
func (c *NewAPIToClaudeConverter) ConvertStreamChunk(chunk *types.ChatCompletionChunk, originalModel string) (*claudeTypes.StreamEvent, error) {
	if chunk == nil {
		return nil, fmt.Errorf("chunk is nil")
	}

	// 获取第一个选择
	if len(chunk.Choices) == 0 {
		return nil, fmt.Errorf("no choices in chunk")
	}

	choice := chunk.Choices[0]

	// 根据完成原因确定事件类型
	if choice.FinishReason != "" {
		// 发送完成事件
		return c.createMessageStopEvent(), nil
	}

	// 发送内容增量事件
	return c.createContentDeltaEvent(choice.Delta, 0), nil
}

// createMessageStartEvent 创建消息开始事件
func (c *NewAPIToClaudeConverter) createMessageStartEvent(id, model string) *claudeTypes.StreamEvent {
	messageStart := claudeTypes.MessageStartEvent{
		Type: claudeTypes.EventMessageStart,
		Message: claudeTypes.MessageContent{
			ID:      id,
			Type:    "message",
			Role:    claudeTypes.RoleAssistant,
			Content: []claudeTypes.ContentItem{},
			Model:   model,
			Usage: claudeTypes.Usage{
				InputTokens:  0,
				OutputTokens: 0,
			},
		},
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventMessageStart,
		Event: "message_start",
		Data:  c.marshalToRawMessage(messageStart),
	}
}

// createContentBlockStartEvent 创建内容块开始事件
func (c *NewAPIToClaudeConverter) createContentBlockStartEvent(index int) *claudeTypes.StreamEvent {
	contentStart := claudeTypes.ContentBlockStartEvent{
		Type:  claudeTypes.EventContentBlockStart,
		Index: index,
		ContentBlock: claudeTypes.ContentItem{
			Type: claudeTypes.ContentTypeText,
			Text: "",
		},
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventContentBlockStart,
		Event: "content_block_start",
		Data:  c.marshalToRawMessage(contentStart),
	}
}

// createContentDeltaEvent 创建内容增量事件
func (c *NewAPIToClaudeConverter) createContentDeltaEvent(delta types.ChatMessage, index int) *claudeTypes.StreamEvent {
	text := delta.GetTextContent()

	contentDelta := claudeTypes.ContentBlockDeltaEvent{
		Type:  claudeTypes.EventContentBlockDelta,
		Index: index,
		Delta: claudeTypes.ContentBlockDelta{
			Type: claudeTypes.ContentTypeText,
			Text: text,
		},
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventContentBlockDelta,
		Event: "content_block_delta",
		Data:  c.marshalToRawMessage(contentDelta),
	}
}

// createContentBlockStopEvent 创建内容块停止事件
func (c *NewAPIToClaudeConverter) createContentBlockStopEvent(index int) *claudeTypes.StreamEvent {
	contentStop := claudeTypes.ContentBlockStopEvent{
		Type:  claudeTypes.EventContentBlockStop,
		Index: index,
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventContentBlockStop,
		Event: "content_block_stop",
		Data:  c.marshalToRawMessage(contentStop),
	}
}

// createMessageDeltaEvent 创建消息增量事件
func (c *NewAPIToClaudeConverter) createMessageDeltaEvent(stopReason string, usage *types.Usage) *claudeTypes.StreamEvent {
	claudeUsage := claudeTypes.Usage{
		InputTokens:  0,
		OutputTokens: 0,
	}

	if usage != nil {
		claudeUsage.InputTokens = usage.PromptTokens
		claudeUsage.OutputTokens = usage.CompletionTokens
	}

	messageDelta := claudeTypes.MessageDeltaEvent{
		Type: claudeTypes.EventMessageDelta,
		Delta: claudeTypes.MessageDelta{
			StopReason:   c.mapStopReason(stopReason),
			StopSequence: "",
		},
		Usage: claudeUsage,
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventMessageDelta,
		Event: "message_delta",
		Data:  c.marshalToRawMessage(messageDelta),
	}
}

// createMessageStopEvent 创建消息停止事件
func (c *NewAPIToClaudeConverter) createMessageStopEvent() *claudeTypes.StreamEvent {
	messageStop := claudeTypes.MessageStopEvent{
		Type: claudeTypes.EventMessageStop,
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventMessageStop,
		Event: "message_stop",
		Data:  c.marshalToRawMessage(messageStop),
	}
}

// createPingEvent 创建Ping事件
func (c *NewAPIToClaudeConverter) createPingEvent() *claudeTypes.StreamEvent {
	ping := claudeTypes.PingEvent{
		Type: claudeTypes.EventPing,
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventPing,
		Event: "ping",
		Data:  c.marshalToRawMessage(ping),
	}
}

// createErrorEvent 创建错误事件
func (c *NewAPIToClaudeConverter) createErrorEvent(err error) *claudeTypes.StreamEvent {
	errorEvent := claudeTypes.ErrorEvent{
		Type: claudeTypes.EventError,
		ErrorDetail: claudeTypes.ErrorDetail{
			Type:    "api_error",
			Message: err.Error(),
		},
	}

	return &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventError,
		Event: "error",
		Data:  c.marshalToRawMessage(errorEvent),
	}
}

// marshalToRawMessage 转换为原始消息
func (c *NewAPIToClaudeConverter) marshalToRawMessage(data interface{}) []byte {
	if jsonData, err := json.Marshal(data); err == nil {
		return jsonData
	}
	return []byte("{}")
}

// GenerateStreamEvents 生成完整的流式事件序列
func (c *NewAPIToClaudeConverter) GenerateStreamEvents(messageID, model string) []*claudeTypes.StreamEvent {
	var events []*claudeTypes.StreamEvent

	// 1. 消息开始事件
	events = append(events, c.createMessageStartEvent(messageID, model))

	// 2. 内容块开始事件
	events = append(events, c.createContentBlockStartEvent(0))

	// 3. Ping事件
	events = append(events, c.createPingEvent())

	return events
}

// GenerateStreamEndEvents 生成流式结束事件序列
func (c *NewAPIToClaudeConverter) GenerateStreamEndEvents(stopReason string, usage *types.Usage) []*claudeTypes.StreamEvent {
	var events []*claudeTypes.StreamEvent

	// 1. 内容块停止事件
	events = append(events, c.createContentBlockStopEvent(0))

	// 2. 消息增量事件（包含使用量）
	events = append(events, c.createMessageDeltaEvent(stopReason, usage))

	// 3. 消息停止事件
	events = append(events, c.createMessageStopEvent())

	return events
}

// GenerateID 生成消息ID
func (c *NewAPIToClaudeConverter) GenerateID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// ValidateResponse 验证响应
func (c *NewAPIToClaudeConverter) ValidateResponse(resp *types.ChatCompletionResponse) error {
	if resp == nil {
		return fmt.Errorf("response is nil")
	}

	if resp.ID == "" {
		return fmt.Errorf("response ID is empty")
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("response has no choices")
	}

	return nil
}

// ValidateStreamChunk 验证流式块
func (c *NewAPIToClaudeConverter) ValidateStreamChunk(chunk *types.ChatCompletionChunk) error {
	if chunk == nil {
		return fmt.Errorf("chunk is nil")
	}

	if chunk.ID == "" {
		return fmt.Errorf("chunk ID is empty")
	}

	if len(chunk.Choices) == 0 {
		return fmt.Errorf("chunk has no choices")
	}

	return nil
}
