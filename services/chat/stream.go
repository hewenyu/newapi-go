package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/hewenyu/newapi-go/internal/utils"
	"github.com/hewenyu/newapi-go/types"
	"go.uber.org/zap"
)

// ChatStreamProcessor 聊天流式处理器
type ChatStreamProcessor struct {
	stream   types.StreamResponse
	logger   utils.Logger
	mu       sync.RWMutex
	chunks   []types.ChatCompletionChunk
	finished bool
	err      error
}

// NewChatStreamProcessor 创建新的聊天流式处理器
func NewChatStreamProcessor(stream types.StreamResponse, logger utils.Logger) *ChatStreamProcessor {
	return &ChatStreamProcessor{
		stream:   stream,
		logger:   logger,
		chunks:   make([]types.ChatCompletionChunk, 0),
		finished: false,
	}
}

// Next 获取下一个流式事件
func (p *ChatStreamProcessor) Next() (*types.StreamEvent, error) {
	p.mu.RLock()
	if p.finished {
		p.mu.RUnlock()
		return nil, io.EOF
	}
	p.mu.RUnlock()

	event, err := p.stream.Next()
	if err != nil {
		p.mu.Lock()
		p.err = err
		p.finished = true
		p.mu.Unlock()
		return nil, err
	}

	// 解析聊天完成流式响应
	if event.Type == types.StreamEventTypeData {
		chunk, parseErr := p.parseChunk(event.Data)
		if parseErr != nil {
			p.logger.Warn("Failed to parse chat completion chunk", zap.Error(parseErr))
		} else {
			p.mu.Lock()
			p.chunks = append(p.chunks, *chunk)
			p.mu.Unlock()
		}
	}

	return event, nil
}

// Close 关闭流式处理器
func (p *ChatStreamProcessor) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.finished = true
	if p.stream != nil {
		return p.stream.Close()
	}
	return nil
}

// Err 获取错误
func (p *ChatStreamProcessor) Err() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.err
}

// Done 检查是否完成
func (p *ChatStreamProcessor) Done() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.finished
}

// Context 获取上下文
func (p *ChatStreamProcessor) Context() context.Context {
	if p.stream != nil {
		return p.stream.Context()
	}
	return context.Background()
}

// GetChunks 获取所有已接收的块
func (p *ChatStreamProcessor) GetChunks() []types.ChatCompletionChunk {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]types.ChatCompletionChunk, len(p.chunks))
	copy(result, p.chunks)
	return result
}

// GetLastChunk 获取最后一个块
func (p *ChatStreamProcessor) GetLastChunk() *types.ChatCompletionChunk {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.chunks) == 0 {
		return nil
	}
	return &p.chunks[len(p.chunks)-1]
}

// CollectContent 收集完整的内容
func (p *ChatStreamProcessor) CollectContent() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var content strings.Builder
	for _, chunk := range p.chunks {
		for _, choice := range chunk.Choices {
			content.WriteString(choice.GetContent())
		}
	}
	return content.String()
}

// CollectResponse 收集完整的响应
func (p *ChatStreamProcessor) CollectResponse() *types.ChatCompletionResponse {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.chunks) == 0 {
		return nil
	}

	firstChunk := p.chunks[0]

	// 构建完整响应
	response := &types.ChatCompletionResponse{
		ID:      firstChunk.ID,
		Object:  "chat.completion",
		Created: firstChunk.Created,
		Model:   firstChunk.Model,
		Choices: make([]types.ChatCompletionChoice, 0),
	}

	// 合并所有选择
	choiceMap := make(map[int]*types.ChatCompletionChoice)

	for _, chunk := range p.chunks {
		for _, chunkChoice := range chunk.Choices {
			if choice, exists := choiceMap[chunkChoice.Index]; exists {
				// 合并内容
				if choice.Message.Content == nil {
					choice.Message.Content = chunkChoice.Delta.Content
				} else if chunkChoice.Delta.Content != nil {
					if contentStr, ok := choice.Message.Content.(string); ok {
						if deltaStr, ok := chunkChoice.Delta.Content.(string); ok {
							choice.Message.Content = contentStr + deltaStr
						}
					}
				}

				// 更新结束原因
				if chunkChoice.FinishReason != "" {
					choice.FinishReason = chunkChoice.FinishReason
				}

				// 合并工具调用
				if len(chunkChoice.Delta.ToolCalls) > 0 {
					choice.Message.ToolCalls = append(choice.Message.ToolCalls, chunkChoice.Delta.ToolCalls...)
				}
			} else {
				// 新建选择
				choiceMap[chunkChoice.Index] = &types.ChatCompletionChoice{
					Index: chunkChoice.Index,
					Message: types.ChatMessage{
						Role:      chunkChoice.Delta.Role,
						Content:   chunkChoice.Delta.Content,
						ToolCalls: chunkChoice.Delta.ToolCalls,
					},
					FinishReason: chunkChoice.FinishReason,
				}
			}
		}

		// 更新使用情况
		if chunk.Usage != nil {
			response.Usage = *chunk.Usage
		}
	}

	// 转换为切片
	for i := 0; i < len(choiceMap); i++ {
		if choice, exists := choiceMap[i]; exists {
			response.Choices = append(response.Choices, *choice)
		}
	}

	return response
}

// parseChunk 解析流式块
func (p *ChatStreamProcessor) parseChunk(data json.RawMessage) (*types.ChatCompletionChunk, error) {
	var chunk types.ChatCompletionChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse chat completion chunk: %w", err)
	}
	return &chunk, nil
}

// ChatStreamReader 聊天流式读取器
type ChatStreamReader struct {
	processor *ChatStreamProcessor
	logger    utils.Logger
	buffer    chan types.ChatCompletionChunk
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.RWMutex
	finished  bool
	err       error
}

// NewChatStreamReader 创建新的聊天流式读取器
func NewChatStreamReader(processor *ChatStreamProcessor, logger utils.Logger) *ChatStreamReader {
	ctx, cancel := context.WithCancel(context.Background())

	reader := &ChatStreamReader{
		processor: processor,
		logger:    logger,
		buffer:    make(chan types.ChatCompletionChunk, 100),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 启动读取协程
	reader.wg.Add(1)
	go reader.readLoop()

	return reader
}

// Read 读取下一个块
func (r *ChatStreamReader) Read() (*types.ChatCompletionChunk, error) {
	select {
	case <-r.ctx.Done():
		return nil, r.ctx.Err()
	case chunk := <-r.buffer:
		return &chunk, nil
	}
}

// ReadWithTimeout 带超时的读取
func (r *ChatStreamReader) ReadWithTimeout(timeout time.Duration) (*types.ChatCompletionChunk, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-r.ctx.Done():
		return nil, r.ctx.Err()
	case <-timer.C:
		return nil, fmt.Errorf("read timeout after %v", timeout)
	case chunk := <-r.buffer:
		return &chunk, nil
	}
}

// Close 关闭读取器
func (r *ChatStreamReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.finished {
		return nil
	}

	r.finished = true
	r.cancel()

	// 等待读取协程结束
	r.wg.Wait()

	// 关闭缓冲区
	close(r.buffer)

	// 关闭处理器
	if r.processor != nil {
		return r.processor.Close()
	}

	return nil
}

// Err 获取错误
func (r *ChatStreamReader) Err() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.err
}

// Done 检查是否完成
func (r *ChatStreamReader) Done() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.finished
}

// readLoop 读取循环
func (r *ChatStreamReader) readLoop() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		event, err := r.processor.Next()
		if err != nil {
			if err == io.EOF {
				r.logger.Debug("Chat stream completed")
			} else {
				r.logger.Error("Chat stream error", zap.Error(err))
				r.mu.Lock()
				r.err = err
				r.mu.Unlock()
			}
			return
		}

		if event.Type == types.StreamEventTypeData {
			chunk, parseErr := r.processor.parseChunk(event.Data)
			if parseErr != nil {
				r.logger.Warn("Failed to parse chat completion chunk", zap.Error(parseErr))
				continue
			}

			select {
			case <-r.ctx.Done():
				return
			case r.buffer <- *chunk:
			}
		}
	}
}

// ChatStreamHandler 聊天流式处理器函数类型
type ChatStreamHandler func(*types.ChatCompletionChunk) error

// ProcessStream 处理流式响应
func ProcessStream(ctx context.Context, stream types.StreamResponse, handler ChatStreamHandler) error {
	logger := utils.GetLogger()
	processor := NewChatStreamProcessor(stream, logger)
	defer processor.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		event, err := processor.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("stream error: %w", err)
		}

		if event.Type == types.StreamEventTypeData {
			chunk, parseErr := processor.parseChunk(event.Data)
			if parseErr != nil {
				logger.Warn("Failed to parse chat completion chunk", zap.Error(parseErr))
				continue
			}

			if err := handler(chunk); err != nil {
				return fmt.Errorf("handler error: %w", err)
			}
		}
	}
}

// CollectStreamResponse 收集流式响应为完整响应
func CollectStreamResponse(ctx context.Context, stream types.StreamResponse) (*types.ChatCompletionResponse, error) {
	logger := utils.GetLogger()
	processor := NewChatStreamProcessor(stream, logger)
	defer processor.Close()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		_, err := processor.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("stream error: %w", err)
		}
	}

	response := processor.CollectResponse()
	if response == nil {
		return nil, fmt.Errorf("no response collected")
	}

	return response, nil
}
