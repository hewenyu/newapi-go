package types

import (
	"encoding/json"
	"time"
)

// BaseResponse 基础响应结构体
type BaseResponse struct {
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model,omitempty"`
	Error   *ErrorResponse  `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ErrorResponse API错误响应结构体
type ErrorResponse struct {
	Type    string      `json:"type"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Param   interface{} `json:"param,omitempty"`
}

// Usage 使用量统计结构体
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ListResponse 列表响应结构体
type ListResponse struct {
	Object  string          `json:"object"`
	Data    json.RawMessage `json:"data"`
	HasMore bool            `json:"has_more"`
	FirstID string          `json:"first_id,omitempty"`
	LastID  string          `json:"last_id,omitempty"`
}

// PaginationOptions 分页选项
type PaginationOptions struct {
	Limit  int    `json:"limit,omitempty"`
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

// Model 模型信息结构体
type Model struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Created   int64  `json:"created"`
	OwnedBy   string `json:"owned_by"`
	Root      string `json:"root,omitempty"`
	Parent    string `json:"parent,omitempty"`
	MaxTokens int    `json:"max_tokens,omitempty"`
}

// Permission 权限信息结构体
type Permission struct {
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	Created            int64       `json:"created"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
}

// RequestOptions 请求选项
type RequestOptions struct {
	Headers map[string]string `json:"headers,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// APIVersion API版本信息
type APIVersion struct {
	Version    string `json:"version"`
	MinVersion string `json:"min_version,omitempty"`
	MaxVersion string `json:"max_version,omitempty"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description,omitempty"`
	APIVersion  APIVersion `json:"api_version"`
	Timestamp   int64      `json:"timestamp"`
}

// GetCurrentTime 获取当前时间戳
func GetCurrentTime() int64 {
	return time.Now().Unix()
}

// GetCurrentTimeMs 获取当前毫秒时间戳
func GetCurrentTimeMs() int64 {
	return time.Now().UnixMilli()
}

// ToJSON 转换为JSON字符串
func (r *BaseResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *BaseResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *BaseResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *BaseResponse) GetError() *ErrorResponse {
	return r.Error
}

// ToJSON 转换为JSON字符串
func (e *ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON 从JSON字符串解析
func (e *ErrorResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// Error 实现error接口
func (e *ErrorResponse) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "unknown error"
}

// GetTotal 获取总使用量
func (u *Usage) GetTotal() int {
	return u.TotalTokens
}

// CalculateTotal 计算总使用量
func (u *Usage) CalculateTotal() {
	u.TotalTokens = u.PromptTokens + u.CompletionTokens
}

// IsEmpty 检查使用量是否为空
func (u *Usage) IsEmpty() bool {
	return u.PromptTokens == 0 && u.CompletionTokens == 0 && u.TotalTokens == 0
}

// HasValidPagination 检查分页选项是否有效
func (p *PaginationOptions) HasValidPagination() bool {
	return p.Limit > 0 || p.After != "" || p.Before != ""
}

// GetEffectiveLimit 获取有效的限制数量
func (p *PaginationOptions) GetEffectiveLimit() int {
	if p.Limit <= 0 {
		return 20 // 默认限制
	}
	if p.Limit > 100 {
		return 100 // 最大限制
	}
	return p.Limit
}

// IsValidModel 检查模型是否有效
func (m *Model) IsValidModel() bool {
	return m.ID != "" && m.Object != ""
}

// GetMaxTokens 获取最大Token数
func (m *Model) GetMaxTokens() int {
	if m.MaxTokens > 0 {
		return m.MaxTokens
	}
	return 4096 // 默认最大Token数
}

// ToJSON 转换为JSON字符串
func (m *Model) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// IsActive 检查权限是否激活
func (p *Permission) IsActive() bool {
	return p.AllowView || p.AllowSampling || p.AllowCreateEngine
}

// CanFineTune 检查是否可以微调
func (p *Permission) CanFineTune() bool {
	return p.AllowFineTuning
}

// HasTimeout 检查是否有超时设置
func (r *RequestOptions) HasTimeout() bool {
	return r.Timeout > 0
}

// GetEffectiveTimeout 获取有效的超时时间
func (r *RequestOptions) GetEffectiveTimeout() time.Duration {
	if r.Timeout > 0 {
		return r.Timeout
	}
	return 30 * time.Second // 默认30秒
}

// IsCompatible 检查API版本是否兼容
func (v *APIVersion) IsCompatible(targetVersion string) bool {
	return v.Version >= targetVersion
}

// IsValid 检查服务器信息是否有效
func (s *ServerInfo) IsValid() bool {
	return s.Name != "" && s.Version != "" && s.APIVersion.Version != ""
}

// GetAge 获取服务器信息年龄（秒）
func (s *ServerInfo) GetAge() int64 {
	return GetCurrentTime() - s.Timestamp
}
