package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hewenyu/newapi-go/internal/utils"
	"github.com/hewenyu/newapi-go/types"
)

// StreamEvent 流式事件结构体
type StreamEvent struct {
	Event string `json:"event,omitempty"`
	Data  string `json:"data,omitempty"`
	ID    string `json:"id,omitempty"`
	Retry int    `json:"retry,omitempty"`
}

// StreamProcessor 流式处理器
type StreamProcessor struct {
	reader    io.ReadCloser
	scanner   *bufio.Scanner
	ctx       context.Context
	cancel    context.CancelFunc
	eventChan chan StreamEvent
	errorChan chan error
	done      chan bool
}

// NewStreamProcessor 创建新的流式处理器
func NewStreamProcessor(ctx context.Context, reader io.ReadCloser) *StreamProcessor {
	ctx, cancel := context.WithCancel(ctx)

	return &StreamProcessor{
		reader:    reader,
		scanner:   bufio.NewScanner(reader),
		ctx:       ctx,
		cancel:    cancel,
		eventChan: make(chan StreamEvent, 100),
		errorChan: make(chan error, 10),
		done:      make(chan bool, 1),
	}
}

// Start 启动流式处理
func (sp *StreamProcessor) Start() {
	go sp.processStream()
}

// Events 获取事件通道
func (sp *StreamProcessor) Events() <-chan StreamEvent {
	return sp.eventChan
}

// Errors 获取错误通道
func (sp *StreamProcessor) Errors() <-chan error {
	return sp.errorChan
}

// Done 获取完成通道
func (sp *StreamProcessor) Done() <-chan bool {
	return sp.done
}

// Close 关闭流式处理器
func (sp *StreamProcessor) Close() error {
	sp.cancel()
	return sp.reader.Close()
}

// processStream 处理流式数据
func (sp *StreamProcessor) processStream() {
	defer func() {
		close(sp.eventChan)
		close(sp.errorChan)
		close(sp.done)
		sp.reader.Close()
	}()

	var event StreamEvent
	var lines []string

	for sp.scanner.Scan() {
		select {
		case <-sp.ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(sp.scanner.Text())

		// 空行表示事件结束
		if line == "" {
			if len(lines) > 0 {
				event = sp.parseEvent(lines)
				if event.Data != "" || event.Event != "" {
					sp.eventChan <- event
					utils.LogStreamEvent(sp.ctx, event.Event, event.Data)
				}
				lines = nil
			}
			continue
		}

		lines = append(lines, line)
	}

	// 处理最后一个事件
	if len(lines) > 0 {
		event = sp.parseEvent(lines)
		if event.Data != "" || event.Event != "" {
			sp.eventChan <- event
			utils.LogStreamEvent(sp.ctx, event.Event, event.Data)
		}
	}

	// 检查扫描错误
	if err := sp.scanner.Err(); err != nil {
		sp.errorChan <- types.NewStreamError(types.ErrTypeAPIError, types.ErrCodeStreamError,
			fmt.Sprintf("stream scan error: %v", err))
	}

	sp.done <- true
}

// parseEvent 解析事件
func (sp *StreamProcessor) parseEvent(lines []string) StreamEvent {
	var event StreamEvent

	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if event.Data == "" {
				event.Data = data
			} else {
				event.Data += "\n" + data
			}
		} else if strings.HasPrefix(line, "event: ") {
			event.Event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "id: ") {
			event.ID = strings.TrimPrefix(line, "id: ")
		} else if strings.HasPrefix(line, "retry: ") {
			// 忽略解析错误
			fmt.Sscanf(strings.TrimPrefix(line, "retry: "), "%d", &event.Retry)
		}
	}

	return event
}

// StreamReader 流式读取器接口
type StreamReader interface {
	Read() (interface{}, error)
	Close() error
	Err() error
}

// JSONStreamReader JSON流式读取器
type JSONStreamReader struct {
	processor *StreamProcessor
	ctx       context.Context
}

// NewJSONStreamReader 创建JSON流式读取器
func NewJSONStreamReader(ctx context.Context, reader io.ReadCloser) *JSONStreamReader {
	processor := NewStreamProcessor(ctx, reader)
	processor.Start()

	return &JSONStreamReader{
		processor: processor,
		ctx:       ctx,
	}
}

// Read 读取下一个JSON对象
func (jr *JSONStreamReader) Read() (interface{}, error) {
	for {
		select {
		case <-jr.ctx.Done():
			return nil, jr.ctx.Err()
		case event := <-jr.processor.Events():
			if event.Data == "" {
				continue
			}

			// 跳过特殊事件
			if event.Data == "[DONE]" {
				return nil, io.EOF
			}

			// 解析JSON数据
			var data interface{}
			if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
				return nil, types.NewStreamError(types.ErrTypeAPIError, types.ErrCodeParseError,
					fmt.Sprintf("failed to parse JSON: %v", err))
			}

			return data, nil
		case err := <-jr.processor.Errors():
			return nil, err
		case <-jr.processor.Done():
			return nil, io.EOF
		}
	}
}

// Close 关闭读取器
func (jr *JSONStreamReader) Close() error {
	return jr.processor.Close()
}

// Err 获取错误
func (jr *JSONStreamReader) Err() error {
	select {
	case err := <-jr.processor.Errors():
		return err
	default:
		return nil
	}
}

// ChatStreamReader 聊天流式读取器
type ChatStreamReader struct {
	jsonReader *JSONStreamReader
}

// NewChatStreamReader 创建聊天流式读取器
func NewChatStreamReader(ctx context.Context, reader io.ReadCloser) *ChatStreamReader {
	return &ChatStreamReader{
		jsonReader: NewJSONStreamReader(ctx, reader),
	}
}

// Read 读取聊天流式数据
func (cr *ChatStreamReader) Read() (interface{}, error) {
	return cr.jsonReader.Read()
}

// Close 关闭读取器
func (cr *ChatStreamReader) Close() error {
	return cr.jsonReader.Close()
}

// Err 获取错误
func (cr *ChatStreamReader) Err() error {
	return cr.jsonReader.Err()
}

// StreamBuffer 流式缓冲区
type StreamBuffer struct {
	buffer []byte
	pos    int
}

// NewStreamBuffer 创建流式缓冲区
func NewStreamBuffer(size int) *StreamBuffer {
	return &StreamBuffer{
		buffer: make([]byte, size),
		pos:    0,
	}
}

// Write 写入数据
func (sb *StreamBuffer) Write(data []byte) (int, error) {
	if sb.pos+len(data) > len(sb.buffer) {
		// 扩展缓冲区
		newSize := len(sb.buffer) * 2
		if newSize < sb.pos+len(data) {
			newSize = sb.pos + len(data)
		}
		newBuffer := make([]byte, newSize)
		copy(newBuffer, sb.buffer[:sb.pos])
		sb.buffer = newBuffer
	}

	copy(sb.buffer[sb.pos:], data)
	sb.pos += len(data)
	return len(data), nil
}

// Read 读取数据
func (sb *StreamBuffer) Read(p []byte) (int, error) {
	if sb.pos == 0 {
		return 0, io.EOF
	}

	n := copy(p, sb.buffer[:sb.pos])
	return n, nil
}

// ReadLine 读取一行
func (sb *StreamBuffer) ReadLine() (string, error) {
	if sb.pos == 0 {
		return "", io.EOF
	}

	data := sb.buffer[:sb.pos]
	for i, b := range data {
		if b == '\n' {
			line := string(data[:i])
			// 移除已读取的数据
			copy(sb.buffer, data[i+1:])
			sb.pos -= i + 1
			return strings.TrimSpace(line), nil
		}
	}

	// 没有找到换行符，返回所有数据
	line := string(data)
	sb.pos = 0
	return strings.TrimSpace(line), nil
}

// Reset 重置缓冲区
func (sb *StreamBuffer) Reset() {
	sb.pos = 0
}

// Len 获取缓冲区长度
func (sb *StreamBuffer) Len() int {
	return sb.pos
}

// Bytes 获取缓冲区数据
func (sb *StreamBuffer) Bytes() []byte {
	return sb.buffer[:sb.pos]
}

// String 获取缓冲区字符串
func (sb *StreamBuffer) String() string {
	return string(sb.buffer[:sb.pos])
}

// StreamOptions 流式选项
type StreamOptions struct {
	BufferSize int           `json:"buffer_size,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty"`
	MaxEvents  int           `json:"max_events,omitempty"`
	KeepAlive  bool          `json:"keep_alive,omitempty"`
	Retry      bool          `json:"retry,omitempty"`
	MaxRetries int           `json:"max_retries,omitempty"`
	RetryDelay time.Duration `json:"retry_delay,omitempty"`
}

// DefaultStreamOptions 默认流式选项
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		BufferSize: 4096,
		Timeout:    60 * time.Second,
		MaxEvents:  1000,
		KeepAlive:  true,
		Retry:      true,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// Validate 验证流式选项
func (so *StreamOptions) Validate() error {
	if so.BufferSize <= 0 {
		return types.NewValidationError("buffer_size", so.BufferSize,
			"buffer size must be positive", types.ErrCodeInvalidParameter)
	}

	if so.Timeout <= 0 {
		return types.NewValidationError("timeout", so.Timeout,
			"timeout must be positive", types.ErrCodeInvalidParameter)
	}

	if so.MaxEvents <= 0 {
		return types.NewValidationError("max_events", so.MaxEvents,
			"max events must be positive", types.ErrCodeInvalidParameter)
	}

	if so.MaxRetries < 0 {
		return types.NewValidationError("max_retries", so.MaxRetries,
			"max retries must be non-negative", types.ErrCodeInvalidParameter)
	}

	if so.RetryDelay < 0 {
		return types.NewValidationError("retry_delay", so.RetryDelay,
			"retry delay must be non-negative", types.ErrCodeInvalidParameter)
	}

	return nil
}
