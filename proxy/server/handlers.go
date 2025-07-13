package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hewenyu/newapi-go/client"
	"github.com/hewenyu/newapi-go/proxy/config"
	"github.com/hewenyu/newapi-go/proxy/converter"
	claudeTypes "github.com/hewenyu/newapi-go/proxy/types"
	"github.com/hewenyu/newapi-go/types"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	config                  *config.Config
	newAPIClient            *client.Client
	claudeToNewAPI          *converter.ClaudeToNewAPIConverter
	newAPIToClaudeConverter *converter.NewAPIToClaudeConverter
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(cfg *config.Config, newAPIClient *client.Client) *MessageHandler {
	return &MessageHandler{
		config:                  cfg,
		newAPIClient:            newAPIClient,
		claudeToNewAPI:          converter.NewClaudeToNewAPIConverter(),
		newAPIToClaudeConverter: converter.NewNewAPIToClaudeConverter(),
	}
}

// HandleMessage 处理消息请求
func (h *MessageHandler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 解析Claude API请求
	claudeReq, err := h.parseClaudeRequest(r)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// 验证请求
	if err := h.claudeToNewAPI.ValidateRequest(claudeReq); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// 检查是否为流式请求
	if claudeReq.Stream {
		h.handleStreamMessage(w, r, claudeReq)
		return
	}

	// 处理普通请求
	h.handleNormalMessage(w, r, claudeReq)
}

// handleNormalMessage 处理普通消息
func (h *MessageHandler) handleNormalMessage(w http.ResponseWriter, r *http.Request, claudeReq *claudeTypes.ClaudeRequest) {
	// 转换请求
	messages, options, err := h.claudeToNewAPI.ConvertRequest(claudeReq)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// 转换上下文
	ctx := h.claudeToNewAPI.ConvertContext(r.Context(), claudeReq)

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, h.config.RequestTimeout)
	defer cancel()

	// 调用NewAPI-Go SDK
	response, err := h.newAPIClient.CreateChatCompletion(ctx, messages, options...)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// 转换响应
	claudeResp, err := h.newAPIToClaudeConverter.ConvertResponse(response, claudeReq.Model)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// 发送响应
	h.sendJSONResponse(w, http.StatusOK, claudeResp)
}

// handleStreamMessage 处理流式消息
func (h *MessageHandler) handleStreamMessage(w http.ResponseWriter, r *http.Request, claudeReq *claudeTypes.ClaudeRequest) {
	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

	// 获取Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.sendErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("streaming not supported"))
		return
	}

	// 转换请求
	messages, options, err := h.claudeToNewAPI.ConvertRequest(claudeReq)
	if err != nil {
		h.sendStreamError(w, flusher, err)
		return
	}

	// 转换上下文
	ctx := h.claudeToNewAPI.ConvertContext(r.Context(), claudeReq)

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, h.config.RequestTimeout)
	defer cancel()

	// 调用NewAPI-Go SDK流式接口
	stream, err := h.newAPIClient.CreateChatCompletionStream(ctx, messages, options...)
	if err != nil {
		h.sendStreamError(w, flusher, err)
		return
	}
	defer stream.Close()

	// 生成消息ID
	messageID := h.newAPIToClaudeConverter.GenerateID()

	// 发送初始事件
	startEvents := h.newAPIToClaudeConverter.GenerateStreamEvents(messageID, claudeReq.Model)
	for _, event := range startEvents {
		h.sendStreamEvent(w, flusher, event)
	}

	// 处理流式数据
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 获取下一个事件
		event, err := stream.Next()
		if err != nil {
			if err == io.EOF {
				// 流结束
				break
			}
			h.sendStreamError(w, flusher, err)
			return
		}

		// 处理事件
		if err := h.processStreamEvent(w, flusher, event); err != nil {
			h.sendStreamError(w, flusher, err)
			return
		}
	}

	// 发送结束事件
	endEvents := h.newAPIToClaudeConverter.GenerateStreamEndEvents("", nil)
	for _, event := range endEvents {
		h.sendStreamEvent(w, flusher, event)
	}
}

// processStreamEvent 处理流式事件
func (h *MessageHandler) processStreamEvent(w http.ResponseWriter, flusher http.Flusher, event *types.StreamEvent) error {
	// 解析事件数据
	if event.Type == types.StreamEventTypeData {
		// 处理流式数据
		var chunk types.ChatCompletionChunk
		if err := json.Unmarshal(event.Data, &chunk); err != nil {
			return err
		}

		// 转换为Claude格式
		claudeEvent, err := h.newAPIToClaudeConverter.ConvertStreamChunk(&chunk, "")
		if err != nil {
			return err
		}

		// 发送转换后的事件
		h.sendStreamEvent(w, flusher, claudeEvent)
	}

	return nil
}

// sendStreamEvent 发送流式事件
func (h *MessageHandler) sendStreamEvent(w http.ResponseWriter, flusher http.Flusher, event *claudeTypes.StreamEvent) {
	if event.Event != "" {
		fmt.Fprintf(w, "event: %s\n", event.Event)
	}
	if event.Data != nil {
		fmt.Fprintf(w, "data: %s\n", string(event.Data))
	}
	fmt.Fprintf(w, "\n")
	flusher.Flush()
}

// sendStreamError 发送流式错误
func (h *MessageHandler) sendStreamError(w http.ResponseWriter, flusher http.Flusher, err error) {
	errorEvent := &claudeTypes.ErrorEvent{
		Type: claudeTypes.EventError,
		ErrorDetail: claudeTypes.ErrorDetail{
			Type:    "api_error",
			Message: err.Error(),
		},
	}

	streamEvent := &claudeTypes.StreamEvent{
		Type:  claudeTypes.EventError,
		Event: "error",
		Data:  h.marshalToRawMessage(errorEvent),
	}

	h.sendStreamEvent(w, flusher, streamEvent)
}

// parseClaudeRequest 解析Claude API请求
func (h *MessageHandler) parseClaudeRequest(r *http.Request) (*claudeTypes.ClaudeRequest, error) {
	// 检查Content-Type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf("content-type must be application/json")
	}

	// 检查请求体大小
	if r.ContentLength > h.config.MaxRequestSize {
		return nil, fmt.Errorf("request body too large")
	}

	// 读取请求体
	body, err := io.ReadAll(io.LimitReader(r.Body, h.config.MaxRequestSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// 解析JSON
	var claudeReq claudeTypes.ClaudeRequest
	if err := json.Unmarshal(body, &claudeReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// 设置默认值
	claudeReq.SetDefaults()

	return &claudeReq, nil
}

// sendJSONResponse 发送JSON响应
func (h *MessageHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// sendErrorResponse 发送错误响应
func (h *MessageHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	// 转换为Claude错误格式
	var claudeErr *claudeTypes.ClaudeError
	if ce, ok := err.(*claudeTypes.ClaudeError); ok {
		claudeErr = ce
	} else {
		claudeErr = claudeTypes.NewAPIError(err.Error())
	}

	h.sendJSONResponse(w, statusCode, claudeErr)
}

// marshalToRawMessage 转换为原始消息
func (h *MessageHandler) marshalToRawMessage(data interface{}) []byte {
	if jsonData, err := json.Marshal(data); err == nil {
		return jsonData
	}
	return []byte("{}")
}

// HealthHandler 健康检查处理器
type HealthHandler struct {
	config *config.Config
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		config: cfg,
	}
}

// HandleHealth 处理健康检查
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now()).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// InfoHandler 信息处理器
type InfoHandler struct {
	config *config.Config
}

// NewInfoHandler 创建信息处理器
func NewInfoHandler(cfg *config.Config) *InfoHandler {
	return &InfoHandler{
		config: cfg,
	}
}

// HandleInfo 处理信息请求
func (h *InfoHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"service":     "Claude API Proxy",
		"version":     "1.0.0",
		"description": "Local proxy server for Claude API using NewAPI-Go SDK",
		"endpoints": []string{
			"POST /v1/messages",
			"GET /health",
			"GET /info",
		},
		"supported_models": []string{
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
			"claude-3-5-sonnet-20241022",
			"claude-3-5-haiku-20241022",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// NotFoundHandler 404处理器
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	claudeErr := claudeTypes.NewInvalidRequestError(fmt.Sprintf("endpoint not found: %s", r.URL.Path))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(claudeErr)
}

// MethodNotAllowedHandler 方法不允许处理器
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	claudeErr := claudeTypes.NewInvalidRequestError(fmt.Sprintf("method not allowed: %s", r.Method))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(claudeErr)
}
