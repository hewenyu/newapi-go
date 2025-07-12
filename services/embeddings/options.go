package embeddings

import (
	"github.com/hewenyu/newapi-go/types"
)

// EmbeddingOption 嵌入选项函数类型
type EmbeddingOption func(*EmbeddingConfig)

// EmbeddingConfig 嵌入配置结构体
type EmbeddingConfig struct {
	Model          string                 `json:"model"`
	EncodingFormat string                 `json:"encoding_format,omitempty"`
	Dimensions     int                    `json:"dimensions,omitempty"`
	User           string                 `json:"user,omitempty"`
	ExtraBody      map[string]interface{} `json:"-"`
}

// DefaultEmbeddingConfig 创建默认嵌入配置
func DefaultEmbeddingConfig() *EmbeddingConfig {
	return &EmbeddingConfig{
		Model:          "text-embedding-3-small",
		EncodingFormat: types.EmbeddingEncodingFormatFloat,
		Dimensions:     0, // 0表示使用模型默认维度
	}
}

// WithModel 设置嵌入模型
func WithModel(model string) EmbeddingOption {
	return func(c *EmbeddingConfig) {
		c.Model = model
	}
}

// WithEncodingFormat 设置编码格式
func WithEncodingFormat(format string) EmbeddingOption {
	return func(c *EmbeddingConfig) {
		c.EncodingFormat = format
	}
}

// WithDimensions 设置嵌入维度
func WithDimensions(dimensions int) EmbeddingOption {
	return func(c *EmbeddingConfig) {
		c.Dimensions = dimensions
	}
}

// WithUser 设置用户标识
func WithUser(user string) EmbeddingOption {
	return func(c *EmbeddingConfig) {
		c.User = user
	}
}

// WithExtraBody 设置额外的请求体参数
func WithExtraBody(extraBody map[string]interface{}) EmbeddingOption {
	return func(c *EmbeddingConfig) {
		if c.ExtraBody == nil {
			c.ExtraBody = make(map[string]interface{})
		}
		for k, v := range extraBody {
			c.ExtraBody[k] = v
		}
	}
}

// ToRequest 将配置转换为嵌入请求
func (c *EmbeddingConfig) ToRequest(input interface{}) *types.EmbeddingRequest {
	req := &types.EmbeddingRequest{
		Input:          input,
		Model:          c.Model,
		EncodingFormat: c.EncodingFormat,
		Dimensions:     c.Dimensions,
		User:           c.User,
		ExtraBody:      c.ExtraBody,
	}

	// 设置默认值
	req.SetDefaults()

	return req
}

// Validate 验证配置
func (c *EmbeddingConfig) Validate() error {
	if c.Model == "" {
		return types.NewValidationError("model", c.Model, "model is required", types.ErrCodeMissingParameter)
	}

	if c.EncodingFormat != "" {
		switch c.EncodingFormat {
		case types.EmbeddingEncodingFormatFloat, types.EmbeddingEncodingFormatBase64:
			// 有效格式
		default:
			return types.NewValidationError("encoding_format", c.EncodingFormat, "invalid encoding format", types.ErrCodeInvalidParameter)
		}
	}

	if c.Dimensions < 0 {
		return types.NewValidationError("dimensions", c.Dimensions, "dimensions must be non-negative", types.ErrCodeInvalidParameter)
	}

	return nil
}

// Clone 克隆配置
func (c *EmbeddingConfig) Clone() *EmbeddingConfig {
	cloned := &EmbeddingConfig{
		Model:          c.Model,
		EncodingFormat: c.EncodingFormat,
		Dimensions:     c.Dimensions,
		User:           c.User,
	}

	if c.ExtraBody != nil {
		cloned.ExtraBody = make(map[string]interface{})
		for k, v := range c.ExtraBody {
			cloned.ExtraBody[k] = v
		}
	}

	return cloned
}
