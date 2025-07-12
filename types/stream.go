package types

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// 流式事件类型常量
const (
	StreamEventTypeData      = "data"
	StreamEventTypeError     = "error"
	StreamEventTypeComplete  = "complete"
	StreamEventTypeClose     = "close"
	StreamEventTypeKeepAlive = "keep-alive"
)

// 流式数据类型常量
const (
	StreamDataTypeChat      = "chat"
	StreamDataTypeEmbedding = "embedding"
	StreamDataTypeImage     = "image"
	StreamDataTypeAudio     = "audio"
)

// 流式连接状态常量
const (
	StreamStateConnecting = "connecting"
	StreamStateConnected  = "connected"
	StreamStateStreaming  = "streaming"
	StreamStateCompleted  = "completed"
	StreamStateError      = "error"
	StreamStateClosed     = "closed"
)

// StreamEvent 流式事件结构体
type StreamEvent struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data,omitempty"`
	ID        string          `json:"id,omitempty"`
	Event     string          `json:"event,omitempty"`
	Retry     int             `json:"retry,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

// StreamResponse 流式响应接口
type StreamResponse interface {
	// Next 获取下一个事件
	Next() (*StreamEvent, error)
	// Close 关闭流
	Close() error
	// Err 获取错误
	Err() error
	// Done 检查是否完成
	Done() bool
	// Context 获取上下文
	Context() context.Context
}

// StreamReader 流式读取器
type StreamReader struct {
	reader    *bufio.Reader
	ctx       context.Context
	cancel    context.CancelFunc
	err       error
	done      bool
	mutex     sync.RWMutex
	events    chan *StreamEvent
	closed    bool
	state     string
	startTime time.Time
}

// StreamWriter 流式写入器
type StreamWriter struct {
	writer io.Writer
	mutex  sync.Mutex
	closed bool
}

// StreamProcessor 流式处理器
type StreamProcessor struct {
	reader   StreamResponse
	handlers map[string]func(*StreamEvent) error
	mutex    sync.RWMutex
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// StreamStats 流式统计信息
type StreamStats struct {
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	EventCount    int           `json:"event_count"`
	BytesReceived int64         `json:"bytes_received"`
	BytesSent     int64         `json:"bytes_sent"`
	ErrorCount    int           `json:"error_count"`
	State         string        `json:"state"`
}

// StreamConfig 流式配置
type StreamConfig struct {
	BufferSize        int           `json:"buffer_size"`
	Timeout           time.Duration `json:"timeout"`
	RetryAttempts     int           `json:"retry_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`
	KeepAliveInterval time.Duration `json:"keep_alive_interval"`
	MaxEventSize      int           `json:"max_event_size"`
	EnableCompression bool          `json:"enable_compression"`
}

// NewStreamReader 创建新的流式读取器
func NewStreamReader(reader io.Reader, ctx context.Context) *StreamReader {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	return &StreamReader{
		reader:    bufio.NewReader(reader),
		ctx:       ctx,
		cancel:    cancel,
		events:    make(chan *StreamEvent, 100),
		state:     StreamStateConnecting,
		startTime: time.Now(),
	}
}

// NewStreamWriter 创建新的流式写入器
func NewStreamWriter(writer io.Writer) *StreamWriter {
	return &StreamWriter{
		writer: writer,
	}
}

// NewStreamProcessor 创建新的流式处理器
func NewStreamProcessor(reader StreamResponse) *StreamProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamProcessor{
		reader:   reader,
		handlers: make(map[string]func(*StreamEvent) error),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Next 获取下一个事件
func (r *StreamReader) Next() (*StreamEvent, error) {
	r.mutex.RLock()
	if r.done || r.closed {
		r.mutex.RUnlock()
		return nil, io.EOF
	}
	r.mutex.RUnlock()

	select {
	case <-r.ctx.Done():
		return nil, r.ctx.Err()
	case event := <-r.events:
		return event, nil
	default:
		return r.readEvent()
	}
}

// Close 关闭流
func (r *StreamReader) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	r.state = StreamStateClosed
	r.cancel()
	close(r.events)

	return nil
}

// Err 获取错误
func (r *StreamReader) Err() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.err
}

// Done 检查是否完成
func (r *StreamReader) Done() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.done
}

// Context 获取上下文
func (r *StreamReader) Context() context.Context {
	return r.ctx
}

// GetState 获取流状态
func (r *StreamReader) GetState() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.state
}

// GetStats 获取统计信息
func (r *StreamReader) GetStats() *StreamStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := &StreamStats{
		StartTime: r.startTime,
		State:     r.state,
	}

	if r.done || r.closed {
		stats.EndTime = time.Now()
		stats.Duration = stats.EndTime.Sub(stats.StartTime)
	}

	return stats
}

// readEvent 读取事件
func (r *StreamReader) readEvent() (*StreamEvent, error) {
	var event *StreamEvent
	var eventData strings.Builder

	for {
		select {
		case <-r.ctx.Done():
			return nil, r.ctx.Err()
		default:
		}

		line, err := r.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				r.mutex.Lock()
				r.done = true
				r.state = StreamStateCompleted
				r.mutex.Unlock()
				return nil, io.EOF
			}

			r.mutex.Lock()
			r.err = err
			r.state = StreamStateError
			r.mutex.Unlock()
			return nil, err
		}

		line = strings.TrimSpace(line)

		// 空行表示事件结束
		if line == "" {
			if eventData.Len() > 0 {
				event = r.parseEvent(eventData.String())
				if event != nil {
					r.mutex.Lock()
					r.state = StreamStateStreaming
					r.mutex.Unlock()
					return event, nil
				}
			}
			continue
		}

		// 忽略注释行
		if strings.HasPrefix(line, ":") {
			continue
		}

		eventData.WriteString(line)
		eventData.WriteString("\n")
	}
}

// parseEvent 解析事件
func (r *StreamReader) parseEvent(data string) *StreamEvent {
	event := &StreamEvent{
		Timestamp: time.Now().UnixMilli(),
	}

	lines := strings.Split(strings.TrimSpace(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")
			if dataStr == "[DONE]" {
				event.Type = StreamEventTypeComplete
				return event
			}

			event.Type = StreamEventTypeData
			event.Data = json.RawMessage(dataStr)
		} else if strings.HasPrefix(line, "event: ") {
			event.Event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "id: ") {
			event.ID = strings.TrimPrefix(line, "id: ")
		} else if strings.HasPrefix(line, "retry: ") {
			fmt.Sscanf(strings.TrimPrefix(line, "retry: "), "%d", &event.Retry)
		}
	}

	return event
}

// WriteEvent 写入事件
func (w *StreamWriter) WriteEvent(event *StreamEvent) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.closed {
		return fmt.Errorf("stream writer is closed")
	}

	var buffer bytes.Buffer

	if event.ID != "" {
		buffer.WriteString(fmt.Sprintf("id: %s\n", event.ID))
	}

	if event.Event != "" {
		buffer.WriteString(fmt.Sprintf("event: %s\n", event.Event))
	}

	if event.Retry > 0 {
		buffer.WriteString(fmt.Sprintf("retry: %d\n", event.Retry))
	}

	if event.Data != nil {
		buffer.WriteString(fmt.Sprintf("data: %s\n", string(event.Data)))
	}

	buffer.WriteString("\n")

	_, err := w.writer.Write(buffer.Bytes())
	return err
}

// WriteData 写入数据
func (w *StreamWriter) WriteData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	event := &StreamEvent{
		Type: StreamEventTypeData,
		Data: jsonData,
	}

	return w.WriteEvent(event)
}

// WriteError 写入错误
func (w *StreamWriter) WriteError(err error) error {
	errorData, jsonErr := json.Marshal(map[string]string{
		"error": err.Error(),
	})
	if jsonErr != nil {
		return jsonErr
	}

	event := &StreamEvent{
		Type:  StreamEventTypeError,
		Event: "error",
		Data:  errorData,
	}

	return w.WriteEvent(event)
}

// WriteComplete 写入完成信号
func (w *StreamWriter) WriteComplete() error {
	event := &StreamEvent{
		Type:  StreamEventTypeComplete,
		Event: "complete",
		Data:  json.RawMessage("\"[DONE]\""),
	}

	return w.WriteEvent(event)
}

// Close 关闭写入器
func (w *StreamWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	return w.WriteComplete()
}

// AddHandler 添加事件处理器
func (p *StreamProcessor) AddHandler(eventType string, handler func(*StreamEvent) error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.handlers[eventType] = handler
}

// RemoveHandler 移除事件处理器
func (p *StreamProcessor) RemoveHandler(eventType string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.handlers, eventType)
}

// Start 启动处理器
func (p *StreamProcessor) Start() error {
	p.mutex.Lock()
	if p.running {
		p.mutex.Unlock()
		return fmt.Errorf("processor is already running")
	}
	p.running = true
	p.mutex.Unlock()

	go p.process()
	return nil
}

// Stop 停止处理器
func (p *StreamProcessor) Stop() error {
	p.mutex.Lock()
	if !p.running {
		p.mutex.Unlock()
		return nil
	}
	p.running = false
	p.mutex.Unlock()

	p.cancel()
	return nil
}

// process 处理事件
func (p *StreamProcessor) process() {
	defer func() {
		p.mutex.Lock()
		p.running = false
		p.mutex.Unlock()
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		event, err := p.reader.Next()
		if err != nil {
			if err == io.EOF {
				return
			}

			// 处理错误
			p.handleEvent(&StreamEvent{
				Type:  StreamEventTypeError,
				Event: "error",
				Data:  json.RawMessage(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
			})
			continue
		}

		if event != nil {
			p.handleEvent(event)
		}
	}
}

// handleEvent 处理事件
func (p *StreamProcessor) handleEvent(event *StreamEvent) {
	p.mutex.RLock()
	handler, exists := p.handlers[event.Type]
	p.mutex.RUnlock()

	if exists && handler != nil {
		if err := handler(event); err != nil {
			// 处理处理器错误
			p.mutex.RLock()
			errorHandler, hasErrorHandler := p.handlers[StreamEventTypeError]
			p.mutex.RUnlock()

			if hasErrorHandler && errorHandler != nil {
				errorEvent := &StreamEvent{
					Type:  StreamEventTypeError,
					Event: "handler_error",
					Data:  json.RawMessage(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
				}
				errorHandler(errorEvent)
			}
		}
	}
}

// IsValidStreamEvent 检查流式事件是否有效
func IsValidStreamEvent(event *StreamEvent) bool {
	return event != nil && event.Type != ""
}

// ParseStreamData 解析流式数据
func ParseStreamData(data json.RawMessage, target interface{}) error {
	return json.Unmarshal(data, target)
}

// CreateStreamEvent 创建流式事件
func CreateStreamEvent(eventType string, data interface{}) (*StreamEvent, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &StreamEvent{
		Type:      eventType,
		Data:      jsonData,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// DefaultStreamConfig 默认流式配置
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		BufferSize:        4096,
		Timeout:           30 * time.Second,
		RetryAttempts:     3,
		RetryDelay:        time.Second,
		KeepAliveInterval: 30 * time.Second,
		MaxEventSize:      1024 * 1024, // 1MB
		EnableCompression: false,
	}
}

// ValidateStreamConfig 验证流式配置
func ValidateStreamConfig(config *StreamConfig) error {
	if config.BufferSize <= 0 {
		return NewValidationError("buffer_size", config.BufferSize, "buffer size must be positive", ErrCodeInvalidParameter)
	}

	if config.Timeout <= 0 {
		return NewValidationError("timeout", config.Timeout, "timeout must be positive", ErrCodeInvalidParameter)
	}

	if config.RetryAttempts < 0 {
		return NewValidationError("retry_attempts", config.RetryAttempts, "retry attempts must be non-negative", ErrCodeInvalidParameter)
	}

	if config.RetryDelay < 0 {
		return NewValidationError("retry_delay", config.RetryDelay, "retry delay must be non-negative", ErrCodeInvalidParameter)
	}

	if config.MaxEventSize <= 0 {
		return NewValidationError("max_event_size", config.MaxEventSize, "max event size must be positive", ErrCodeInvalidParameter)
	}

	return nil
}
