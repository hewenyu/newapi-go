package types

import (
	"encoding/json"
	"fmt"
)

// 嵌入编码格式常量
const (
	EmbeddingEncodingFormatFloat  = "float"
	EmbeddingEncodingFormatBase64 = "base64"
)

// EmbeddingRequest 嵌入请求结构体
type EmbeddingRequest struct {
	Input          interface{}            `json:"input"`
	Model          string                 `json:"model"`
	EncodingFormat string                 `json:"encoding_format,omitempty"`
	Dimensions     int                    `json:"dimensions,omitempty"`
	User           string                 `json:"user,omitempty"`
	ExtraBody      map[string]interface{} `json:"-"`
}

// EmbeddingResponse 嵌入响应结构体
type EmbeddingResponse struct {
	Object string         `json:"object"`
	Data   []Embedding    `json:"data"`
	Model  string         `json:"model"`
	Usage  Usage          `json:"usage"`
	Error  *ErrorResponse `json:"error,omitempty"`
}

// Embedding 嵌入向量结构体
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingInput 嵌入输入类型
type EmbeddingInput struct {
	Text   string `json:"text,omitempty"`
	Tokens []int  `json:"tokens,omitempty"`
}

// EmbeddingBatch 批量嵌入请求结构体
type EmbeddingBatch struct {
	Inputs []EmbeddingInput `json:"inputs"`
	Model  string           `json:"model"`
	Config EmbeddingConfig  `json:"config,omitempty"`
}

// EmbeddingConfig 嵌入配置结构体
type EmbeddingConfig struct {
	Dimensions     int    `json:"dimensions,omitempty"`
	EncodingFormat string `json:"encoding_format,omitempty"`
	Normalize      bool   `json:"normalize,omitempty"`
	Truncate       bool   `json:"truncate,omitempty"`
}

// EmbeddingStats 嵌入统计信息
type EmbeddingStats struct {
	TotalVectors    int     `json:"total_vectors"`
	TotalDimensions int     `json:"total_dimensions"`
	AvgMagnitude    float64 `json:"avg_magnitude"`
	MinMagnitude    float64 `json:"min_magnitude"`
	MaxMagnitude    float64 `json:"max_magnitude"`
}

// EmbeddingComparison 嵌入比较结果
type EmbeddingComparison struct {
	Embedding1        []float64 `json:"embedding1"`
	Embedding2        []float64 `json:"embedding2"`
	CosineSimilarity  float64   `json:"cosine_similarity"`
	DotProduct        float64   `json:"dot_product"`
	EuclideanDistance float64   `json:"euclidean_distance"`
}

// NewEmbeddingRequest 创建新的嵌入请求
func NewEmbeddingRequest(input interface{}, model string) *EmbeddingRequest {
	return &EmbeddingRequest{
		Input: input,
		Model: model,
	}
}

// NewEmbeddingRequestFromText 从文本创建嵌入请求
func NewEmbeddingRequestFromText(text, model string) *EmbeddingRequest {
	return NewEmbeddingRequest(text, model)
}

// NewEmbeddingRequestFromTexts 从多个文本创建嵌入请求
func NewEmbeddingRequestFromTexts(texts []string, model string) *EmbeddingRequest {
	return NewEmbeddingRequest(texts, model)
}

// NewEmbeddingRequestFromTokens 从Token创建嵌入请求
func NewEmbeddingRequestFromTokens(tokens []int, model string) *EmbeddingRequest {
	return NewEmbeddingRequest(tokens, model)
}

// ValidateParameters 验证请求参数
func (r *EmbeddingRequest) ValidateParameters() error {
	if r.Model == "" {
		return NewValidationError("model", r.Model, "model is required", ErrCodeMissingParameter)
	}
	if r.Input == nil {
		return NewValidationError("input", r.Input, "input is required", ErrCodeMissingParameter)
	}

	// 验证输入类型
	switch input := r.Input.(type) {
	case string:
		if input == "" {
			return NewValidationError("input", r.Input, "input text cannot be empty", ErrCodeInvalidParameter)
		}
	case []string:
		if len(input) == 0 {
			return NewValidationError("input", r.Input, "input array cannot be empty", ErrCodeInvalidParameter)
		}
		for i, text := range input {
			if text == "" {
				return NewValidationError(fmt.Sprintf("input[%d]", i), text, "input text cannot be empty", ErrCodeInvalidParameter)
			}
		}
	case []int:
		if len(input) == 0 {
			return NewValidationError("input", r.Input, "input tokens cannot be empty", ErrCodeInvalidParameter)
		}
	default:
		return NewValidationError("input", r.Input, "invalid input type", ErrCodeInvalidParameter)
	}

	// 验证编码格式
	if r.EncodingFormat != "" {
		switch r.EncodingFormat {
		case EmbeddingEncodingFormatFloat, EmbeddingEncodingFormatBase64:
			// 有效格式
		default:
			return NewValidationError("encoding_format", r.EncodingFormat, "invalid encoding format", ErrCodeInvalidParameter)
		}
	}

	// 验证维度
	if r.Dimensions < 0 {
		return NewValidationError("dimensions", r.Dimensions, "dimensions must be positive", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *EmbeddingRequest) SetDefaults() {
	if r.EncodingFormat == "" {
		r.EncodingFormat = EmbeddingEncodingFormatFloat
	}
}

// GetInputCount 获取输入数量
func (r *EmbeddingRequest) GetInputCount() int {
	switch input := r.Input.(type) {
	case string:
		return 1
	case []string:
		return len(input)
	case []int:
		return 1
	default:
		return 0
	}
}

// GetInputTexts 获取输入文本列表
func (r *EmbeddingRequest) GetInputTexts() []string {
	switch input := r.Input.(type) {
	case string:
		return []string{input}
	case []string:
		return input
	default:
		return nil
	}
}

// IsBase64Encoding 检查是否为Base64编码
func (r *EmbeddingRequest) IsBase64Encoding() bool {
	return r.EncodingFormat == EmbeddingEncodingFormatBase64
}

// ToJSON 转换为JSON字符串
func (r *EmbeddingRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *EmbeddingRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *EmbeddingResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *EmbeddingResponse) GetError() *ErrorResponse {
	return r.Error
}

// GetEmbeddingCount 获取嵌入向量数量
func (r *EmbeddingResponse) GetEmbeddingCount() int {
	return len(r.Data)
}

// GetFirstEmbedding 获取第一个嵌入向量
func (r *EmbeddingResponse) GetFirstEmbedding() *Embedding {
	if len(r.Data) > 0 {
		return &r.Data[0]
	}
	return nil
}

// GetEmbeddingByIndex 根据索引获取嵌入向量
func (r *EmbeddingResponse) GetEmbeddingByIndex(index int) *Embedding {
	if index >= 0 && index < len(r.Data) {
		return &r.Data[index]
	}
	return nil
}

// GetAllEmbeddings 获取所有嵌入向量
func (r *EmbeddingResponse) GetAllEmbeddings() [][]float64 {
	embeddings := make([][]float64, len(r.Data))
	for i, embedding := range r.Data {
		embeddings[i] = embedding.Embedding
	}
	return embeddings
}

// ToJSON 转换为JSON字符串
func (r *EmbeddingResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *EmbeddingResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// GetDimensions 获取向量维度
func (e *Embedding) GetDimensions() int {
	return len(e.Embedding)
}

// GetMagnitude 获取向量模长
func (e *Embedding) GetMagnitude() float64 {
	var sum float64
	for _, val := range e.Embedding {
		sum += val * val
	}
	return sum
}

// Normalize 归一化向量
func (e *Embedding) Normalize() {
	magnitude := e.GetMagnitude()
	if magnitude > 0 {
		for i := range e.Embedding {
			e.Embedding[i] /= magnitude
		}
	}
}

// CosineSimilarity 计算余弦相似度
func (e *Embedding) CosineSimilarity(other *Embedding) float64 {
	if len(e.Embedding) != len(other.Embedding) {
		return 0.0
	}

	var dotProduct, magA, magB float64
	for i := range e.Embedding {
		dotProduct += e.Embedding[i] * other.Embedding[i]
		magA += e.Embedding[i] * e.Embedding[i]
		magB += other.Embedding[i] * other.Embedding[i]
	}

	if magA == 0 || magB == 0 {
		return 0.0
	}

	return dotProduct / (magA * magB)
}

// DotProduct 计算点积
func (e *Embedding) DotProduct(other *Embedding) float64 {
	if len(e.Embedding) != len(other.Embedding) {
		return 0.0
	}

	var dotProduct float64
	for i := range e.Embedding {
		dotProduct += e.Embedding[i] * other.Embedding[i]
	}

	return dotProduct
}

// EuclideanDistance 计算欧几里得距离
func (e *Embedding) EuclideanDistance(other *Embedding) float64 {
	if len(e.Embedding) != len(other.Embedding) {
		return 0.0
	}

	var sum float64
	for i := range e.Embedding {
		diff := e.Embedding[i] - other.Embedding[i]
		sum += diff * diff
	}

	return sum
}

// IsValid 检查嵌入向量是否有效
func (e *Embedding) IsValid() bool {
	return len(e.Embedding) > 0
}

// ToJSON 转换为JSON字符串
func (e *Embedding) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON 从JSON字符串解析
func (e *Embedding) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// ValidateConfig 验证配置
func (c *EmbeddingConfig) ValidateConfig() error {
	if c.Dimensions < 0 {
		return NewValidationError("dimensions", c.Dimensions, "dimensions must be positive", ErrCodeInvalidParameter)
	}

	if c.EncodingFormat != "" {
		switch c.EncodingFormat {
		case EmbeddingEncodingFormatFloat, EmbeddingEncodingFormatBase64:
			// 有效格式
		default:
			return NewValidationError("encoding_format", c.EncodingFormat, "invalid encoding format", ErrCodeInvalidParameter)
		}
	}

	return nil
}

// SetDefaults 设置默认配置
func (c *EmbeddingConfig) SetDefaults() {
	if c.EncodingFormat == "" {
		c.EncodingFormat = EmbeddingEncodingFormatFloat
	}
}

// CompareEmbeddings 比较两个嵌入向量
func CompareEmbeddings(emb1, emb2 *Embedding) *EmbeddingComparison {
	if emb1 == nil || emb2 == nil {
		return nil
	}

	return &EmbeddingComparison{
		Embedding1:        emb1.Embedding,
		Embedding2:        emb2.Embedding,
		CosineSimilarity:  emb1.CosineSimilarity(emb2),
		DotProduct:        emb1.DotProduct(emb2),
		EuclideanDistance: emb1.EuclideanDistance(emb2),
	}
}

// CalculateStats 计算嵌入统计信息
func CalculateStats(embeddings []Embedding) *EmbeddingStats {
	if len(embeddings) == 0 {
		return nil
	}

	stats := &EmbeddingStats{
		TotalVectors:    len(embeddings),
		TotalDimensions: len(embeddings[0].Embedding),
		MinMagnitude:    embeddings[0].GetMagnitude(),
		MaxMagnitude:    embeddings[0].GetMagnitude(),
	}

	var totalMagnitude float64
	for _, embedding := range embeddings {
		magnitude := embedding.GetMagnitude()
		totalMagnitude += magnitude

		if magnitude < stats.MinMagnitude {
			stats.MinMagnitude = magnitude
		}
		if magnitude > stats.MaxMagnitude {
			stats.MaxMagnitude = magnitude
		}
	}

	stats.AvgMagnitude = totalMagnitude / float64(len(embeddings))

	return stats
}
