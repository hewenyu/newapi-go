package embeddings

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

// EmbeddingService 嵌入服务结构体
type EmbeddingService struct {
	transport transport.HTTPTransport
	logger    utils.Logger
	config    *EmbeddingConfig
	mu        sync.RWMutex
}

// NewEmbeddingService 创建新的嵌入服务实例
func NewEmbeddingService(transport transport.HTTPTransport, logger utils.Logger, options ...EmbeddingOption) *EmbeddingService {
	config := DefaultEmbeddingConfig()

	// 应用选项
	for _, option := range options {
		option(config)
	}

	return &EmbeddingService{
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

// CreateEmbedding 创建单个文本的嵌入向量
func (s *EmbeddingService) CreateEmbedding(ctx context.Context, text string, options ...EmbeddingOption) (*types.EmbeddingResponse, error) {
	// 验证输入
	if text == "" {
		return nil, fmt.Errorf("input text cannot be empty")
	}

	// 创建配置副本并应用选项
	config := s.getConfig()
	for _, option := range options {
		option(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid embedding config: %w", err)
	}

	// 构建请求
	req := config.ToRequest(text)

	// 验证请求参数
	if err := req.ValidateParameters(); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// 发送请求
	resp, err := s.transport.Post(ctx, "/v1/embeddings", req)
	if err != nil {
		s.logger.Error("Failed to create embedding", zap.Error(err))
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	// 解析响应
	var embeddingResp types.EmbeddingResponse
	if err := parseJSONResponse(resp, &embeddingResp); err != nil {
		s.logger.Error("Failed to parse embedding response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查API错误
	if embeddingResp.IsError() {
		apiErr := embeddingResp.GetError()
		s.logger.Error("API returned error", zap.String("error", apiErr.Message))
		return nil, fmt.Errorf("API error: %s", apiErr.Message)
	}

	s.logger.Debug("Embedding created successfully", zap.Int("count", embeddingResp.GetEmbeddingCount()))
	return &embeddingResp, nil
}

// CreateEmbeddings 创建批量文本的嵌入向量
func (s *EmbeddingService) CreateEmbeddings(ctx context.Context, texts []string, options ...EmbeddingOption) (*types.EmbeddingResponse, error) {
	// 验证输入
	if len(texts) == 0 {
		return nil, fmt.Errorf("input texts cannot be empty")
	}

	for i, text := range texts {
		if text == "" {
			return nil, fmt.Errorf("input text at index %d cannot be empty", i)
		}
	}

	// 创建配置副本并应用选项
	config := s.getConfig()
	for _, option := range options {
		option(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid embedding config: %w", err)
	}

	// 构建请求
	req := config.ToRequest(texts)

	// 验证请求参数
	if err := req.ValidateParameters(); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// 发送请求
	resp, err := s.transport.Post(ctx, "/v1/embeddings", req)
	if err != nil {
		s.logger.Error("Failed to create embeddings", zap.Error(err))
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	// 解析响应
	var embeddingResp types.EmbeddingResponse
	if err := parseJSONResponse(resp, &embeddingResp); err != nil {
		s.logger.Error("Failed to parse embeddings response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查API错误
	if embeddingResp.IsError() {
		apiErr := embeddingResp.GetError()
		s.logger.Error("API returned error", zap.String("error", apiErr.Message))
		return nil, fmt.Errorf("API error: %s", apiErr.Message)
	}

	s.logger.Debug("Embeddings created successfully", zap.Int("count", embeddingResp.GetEmbeddingCount()))
	return &embeddingResp, nil
}

// CreateEmbeddingFromTokens 从token创建嵌入向量
func (s *EmbeddingService) CreateEmbeddingFromTokens(ctx context.Context, tokens []int, options ...EmbeddingOption) (*types.EmbeddingResponse, error) {
	// 验证输入
	if len(tokens) == 0 {
		return nil, fmt.Errorf("input tokens cannot be empty")
	}

	// 创建配置副本并应用选项
	config := s.getConfig()
	for _, option := range options {
		option(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid embedding config: %w", err)
	}

	// 构建请求
	req := config.ToRequest(tokens)

	// 验证请求参数
	if err := req.ValidateParameters(); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// 发送请求
	resp, err := s.transport.Post(ctx, "/v1/embeddings", req)
	if err != nil {
		s.logger.Error("Failed to create embedding from tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to create embedding from tokens: %w", err)
	}

	// 解析响应
	var embeddingResp types.EmbeddingResponse
	if err := parseJSONResponse(resp, &embeddingResp); err != nil {
		s.logger.Error("Failed to parse embedding response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查API错误
	if embeddingResp.IsError() {
		apiErr := embeddingResp.GetError()
		s.logger.Error("API returned error", zap.String("error", apiErr.Message))
		return nil, fmt.Errorf("API error: %s", apiErr.Message)
	}

	s.logger.Debug("Embedding from tokens created successfully", zap.Int("count", embeddingResp.GetEmbeddingCount()))
	return &embeddingResp, nil
}

// UpdateConfig 更新配置
func (s *EmbeddingService) UpdateConfig(options ...EmbeddingOption) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, option := range options {
		option(s.config)
	}
}

// GetConfig 获取配置的只读副本
func (s *EmbeddingService) GetConfig() *EmbeddingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config.Clone()
}

// getConfig 获取配置的副本（内部使用）
func (s *EmbeddingService) getConfig() *EmbeddingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config.Clone()
}

// ValidateInput 验证输入
func (s *EmbeddingService) ValidateInput(input interface{}) error {
	switch v := input.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("input text cannot be empty")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("input texts cannot be empty")
		}
		for i, text := range v {
			if text == "" {
				return fmt.Errorf("input text at index %d cannot be empty", i)
			}
		}
	case []int:
		if len(v) == 0 {
			return fmt.Errorf("input tokens cannot be empty")
		}
	default:
		return fmt.Errorf("unsupported input type: %T", input)
	}

	return nil
}

// GetSupportedModels 获取支持的模型列表
func (s *EmbeddingService) GetSupportedModels() []string {
	return []string{
		"text-embedding-3-small",
		"text-embedding-3-large",
		"text-embedding-ada-002",
	}
}

// GetMaxInputLength 获取最大输入长度
func (s *EmbeddingService) GetMaxInputLength(model string) int {
	// 不同模型的最大token长度
	switch model {
	case "text-embedding-3-small", "text-embedding-3-large":
		return 8192
	case "text-embedding-ada-002":
		return 8192
	default:
		return 8192 // 默认值
	}
}

// GetDefaultDimensions 获取默认维度
func (s *EmbeddingService) GetDefaultDimensions(model string) int {
	// 不同模型的默认维度
	switch model {
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	case "text-embedding-ada-002":
		return 1536
	default:
		return 1536 // 默认值
	}
}
