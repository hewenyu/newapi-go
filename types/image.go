package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 图像尺寸常量
const (
	ImageSize256x256   = "256x256"
	ImageSize512x512   = "512x512"
	ImageSize1024x1024 = "1024x1024"
	ImageSize1792x1024 = "1792x1024"
	ImageSize1024x1792 = "1024x1792"
)

// 图像格式常量
const (
	ImageFormatURL     = "url"
	ImageFormatB64JSON = "b64_json"
)

// 图像质量常量
const (
	ImageQualityStandard = "standard"
	ImageQualityHD       = "hd"
)

// 图像风格常量
const (
	ImageStyleVivid   = "vivid"
	ImageStyleNatural = "natural"
)

// 图像编辑操作常量
const (
	ImageEditOperationInpaint   = "inpaint"
	ImageEditOperationOutpaint  = "outpaint"
	ImageEditOperationVariation = "variation"
)

// ImageGenerationRequest 图像生成请求结构体
type ImageGenerationRequest struct {
	Model          string                 `json:"model,omitempty"`
	Prompt         string                 `json:"prompt"`
	NegativePrompt string                 `json:"negative_prompt,omitempty"`
	N              int                    `json:"n,omitempty"`
	Size           string                 `json:"size,omitempty"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	User           string                 `json:"user,omitempty"`
	Quality        string                 `json:"quality,omitempty"`
	Style          string                 `json:"style,omitempty"`
	ExtraBody      map[string]interface{} `json:"-"`
}

// ImageEditRequest 图像编辑请求结构体
type ImageEditRequest struct {
	Model          string                 `json:"model,omitempty"`
	Image          string                 `json:"image"`
	Mask           string                 `json:"mask,omitempty"`
	Prompt         string                 `json:"prompt"`
	NegativePrompt string                 `json:"negative_prompt,omitempty"`
	N              int                    `json:"n,omitempty"`
	Size           string                 `json:"size,omitempty"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	User           string                 `json:"user,omitempty"`
	Operation      string                 `json:"operation,omitempty"`
	ExtraBody      map[string]interface{} `json:"-"`
}

// ImageVariationRequest 图像变换请求结构体
type ImageVariationRequest struct {
	Model          string                 `json:"model,omitempty"`
	Image          string                 `json:"image"`
	N              int                    `json:"n,omitempty"`
	Size           string                 `json:"size,omitempty"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	User           string                 `json:"user,omitempty"`
	ExtraBody      map[string]interface{} `json:"-"`
}

// ImageResponse 图像响应结构体
type ImageResponse struct {
	Created int64          `json:"created"`
	Data    []ImageData    `json:"data"`
	Error   *ErrorResponse `json:"error,omitempty"`
}

// ImageData 图像数据结构体
type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageAnalysisRequest 图像分析请求结构体
type ImageAnalysisRequest struct {
	Model     string                 `json:"model,omitempty"`
	Image     string                 `json:"image"`
	Prompt    string                 `json:"prompt,omitempty"`
	MaxTokens int                    `json:"max_tokens,omitempty"`
	Detail    string                 `json:"detail,omitempty"`
	Features  []string               `json:"features,omitempty"`
	ExtraBody map[string]interface{} `json:"-"`
}

// ImageAnalysisResponse 图像分析响应结构体
type ImageAnalysisResponse struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []ImageAnalysisChoice `json:"choices"`
	Usage   Usage                 `json:"usage"`
	Error   *ErrorResponse        `json:"error,omitempty"`
}

// ImageAnalysisChoice 图像分析选择结构体
type ImageAnalysisChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ImageUploadRequest 图像上传请求结构体
type ImageUploadRequest struct {
	File      string                 `json:"file"`
	Purpose   string                 `json:"purpose"`
	Filename  string                 `json:"filename,omitempty"`
	ExtraBody map[string]interface{} `json:"-"`
}

// ImageUploadResponse 图像上传响应结构体
type ImageUploadResponse struct {
	ID        string         `json:"id"`
	Object    string         `json:"object"`
	Bytes     int64          `json:"bytes"`
	CreatedAt int64          `json:"created_at"`
	Filename  string         `json:"filename"`
	Purpose   string         `json:"purpose"`
	Error     *ErrorResponse `json:"error,omitempty"`
}

// ImageSize 图像尺寸结构体
type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ImageMetadata 图像元数据结构体
type ImageMetadata struct {
	Format    string    `json:"format"`
	Size      ImageSize `json:"size"`
	FileSize  int64     `json:"file_size"`
	Quality   string    `json:"quality,omitempty"`
	Style     string    `json:"style,omitempty"`
	Model     string    `json:"model,omitempty"`
	Prompt    string    `json:"prompt,omitempty"`
	CreatedAt int64     `json:"created_at"`
}

// NewImageGenerationRequest 创建新的图像生成请求
func NewImageGenerationRequest(prompt string) *ImageGenerationRequest {
	return &ImageGenerationRequest{
		Prompt: prompt,
	}
}

// NewImageEditRequest 创建新的图像编辑请求
func NewImageEditRequest(image, prompt string) *ImageEditRequest {
	return &ImageEditRequest{
		Image:  image,
		Prompt: prompt,
	}
}

// NewImageVariationRequest 创建新的图像变换请求
func NewImageVariationRequest(image string) *ImageVariationRequest {
	return &ImageVariationRequest{
		Image: image,
	}
}

// NewImageAnalysisRequest 创建新的图像分析请求
func NewImageAnalysisRequest(image, prompt string) *ImageAnalysisRequest {
	return &ImageAnalysisRequest{
		Image:  image,
		Prompt: prompt,
	}
}

// ValidateParameters 验证图像生成请求参数
func (r *ImageGenerationRequest) ValidateParameters() error {
	if r.Prompt == "" {
		return NewValidationError("prompt", r.Prompt, "prompt is required", ErrCodeMissingParameter)
	}

	if len(r.Prompt) > 4000 {
		return NewValidationError("prompt", r.Prompt, "prompt is too long", ErrCodeInvalidParameter)
	}

	if r.N < 1 || r.N > 10 {
		return NewValidationError("n", r.N, "n must be between 1 and 10", ErrCodeInvalidParameter)
	}

	if r.Size != "" && !IsValidImageSize(r.Size) {
		return NewValidationError("size", r.Size, "invalid image size", ErrCodeInvalidParameter)
	}

	if r.ResponseFormat != "" && !IsValidResponseFormat(r.ResponseFormat) {
		return NewValidationError("response_format", r.ResponseFormat, "invalid response format", ErrCodeInvalidParameter)
	}

	if r.Quality != "" && !IsValidImageQuality(r.Quality) {
		return NewValidationError("quality", r.Quality, "invalid image quality", ErrCodeInvalidParameter)
	}

	if r.Style != "" && !IsValidImageStyle(r.Style) {
		return NewValidationError("style", r.Style, "invalid image style", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *ImageGenerationRequest) SetDefaults() {
	if r.N == 0 {
		r.N = 1
	}
	if r.Size == "" {
		r.Size = ImageSize1024x1024
	}
	if r.ResponseFormat == "" {
		r.ResponseFormat = ImageFormatURL
	}
	if r.Quality == "" {
		r.Quality = ImageQualityStandard
	}
	if r.Style == "" {
		r.Style = ImageStyleVivid
	}
}

// ToJSON 转换为JSON字符串
func (r *ImageGenerationRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ImageGenerationRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ValidateParameters 验证图像编辑请求参数
func (r *ImageEditRequest) ValidateParameters() error {
	if r.Image == "" {
		return NewValidationError("image", r.Image, "image is required", ErrCodeMissingParameter)
	}

	if r.Prompt == "" {
		return NewValidationError("prompt", r.Prompt, "prompt is required", ErrCodeMissingParameter)
	}

	if len(r.Prompt) > 4000 {
		return NewValidationError("prompt", r.Prompt, "prompt is too long", ErrCodeInvalidParameter)
	}

	if r.N < 1 || r.N > 10 {
		return NewValidationError("n", r.N, "n must be between 1 and 10", ErrCodeInvalidParameter)
	}

	if r.Size != "" && !IsValidImageSize(r.Size) {
		return NewValidationError("size", r.Size, "invalid image size", ErrCodeInvalidParameter)
	}

	if r.ResponseFormat != "" && !IsValidResponseFormat(r.ResponseFormat) {
		return NewValidationError("response_format", r.ResponseFormat, "invalid response format", ErrCodeInvalidParameter)
	}

	if r.Operation != "" && !IsValidEditOperation(r.Operation) {
		return NewValidationError("operation", r.Operation, "invalid edit operation", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *ImageEditRequest) SetDefaults() {
	if r.N == 0 {
		r.N = 1
	}
	if r.Size == "" {
		r.Size = ImageSize1024x1024
	}
	if r.ResponseFormat == "" {
		r.ResponseFormat = ImageFormatURL
	}
	if r.Operation == "" {
		r.Operation = ImageEditOperationInpaint
	}
}

// ToJSON 转换为JSON字符串
func (r *ImageEditRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ImageEditRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ValidateParameters 验证图像变换请求参数
func (r *ImageVariationRequest) ValidateParameters() error {
	if r.Image == "" {
		return NewValidationError("image", r.Image, "image is required", ErrCodeMissingParameter)
	}

	if r.N < 1 || r.N > 10 {
		return NewValidationError("n", r.N, "n must be between 1 and 10", ErrCodeInvalidParameter)
	}

	if r.Size != "" && !IsValidImageSize(r.Size) {
		return NewValidationError("size", r.Size, "invalid image size", ErrCodeInvalidParameter)
	}

	if r.ResponseFormat != "" && !IsValidResponseFormat(r.ResponseFormat) {
		return NewValidationError("response_format", r.ResponseFormat, "invalid response format", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *ImageVariationRequest) SetDefaults() {
	if r.N == 0 {
		r.N = 1
	}
	if r.Size == "" {
		r.Size = ImageSize1024x1024
	}
	if r.ResponseFormat == "" {
		r.ResponseFormat = ImageFormatURL
	}
}

// ToJSON 转换为JSON字符串
func (r *ImageVariationRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ImageVariationRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ValidateParameters 验证图像分析请求参数
func (r *ImageAnalysisRequest) ValidateParameters() error {
	if r.Image == "" {
		return NewValidationError("image", r.Image, "image is required", ErrCodeMissingParameter)
	}

	if r.MaxTokens < 0 {
		return NewValidationError("max_tokens", r.MaxTokens, "max_tokens must be positive", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *ImageAnalysisRequest) SetDefaults() {
	if r.MaxTokens == 0 {
		r.MaxTokens = 300
	}
	if r.Detail == "" {
		r.Detail = "auto"
	}
}

// ToJSON 转换为JSON字符串
func (r *ImageAnalysisRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ImageAnalysisRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *ImageResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *ImageResponse) GetError() *ErrorResponse {
	return r.Error
}

// GetImageCount 获取图像数量
func (r *ImageResponse) GetImageCount() int {
	return len(r.Data)
}

// GetFirstImage 获取第一张图像
func (r *ImageResponse) GetFirstImage() *ImageData {
	if len(r.Data) > 0 {
		return &r.Data[0]
	}
	return nil
}

// GetImageByIndex 根据索引获取图像
func (r *ImageResponse) GetImageByIndex(index int) *ImageData {
	if index >= 0 && index < len(r.Data) {
		return &r.Data[index]
	}
	return nil
}

// GetAllImages 获取所有图像
func (r *ImageResponse) GetAllImages() []ImageData {
	return r.Data
}

// ToJSON 转换为JSON字符串
func (r *ImageResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ImageResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *ImageAnalysisResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *ImageAnalysisResponse) GetError() *ErrorResponse {
	return r.Error
}

// GetFirstChoice 获取第一个选择
func (r *ImageAnalysisResponse) GetFirstChoice() *ImageAnalysisChoice {
	if len(r.Choices) > 0 {
		return &r.Choices[0]
	}
	return nil
}

// GetFirstContent 获取第一个内容
func (r *ImageAnalysisResponse) GetFirstContent() string {
	if choice := r.GetFirstChoice(); choice != nil {
		return choice.Message.GetTextContent()
	}
	return ""
}

// ToJSON 转换为JSON字符串
func (r *ImageAnalysisResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *ImageAnalysisResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsURL 检查是否为URL格式
func (d *ImageData) IsURL() bool {
	return d.URL != ""
}

// IsBase64 检查是否为Base64格式
func (d *ImageData) IsBase64() bool {
	return d.B64JSON != ""
}

// GetContent 获取图像内容
func (d *ImageData) GetContent() string {
	if d.IsURL() {
		return d.URL
	}
	return d.B64JSON
}

// ToJSON 转换为JSON字符串
func (d *ImageData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// FromJSON 从JSON字符串解析
func (d *ImageData) FromJSON(data []byte) error {
	return json.Unmarshal(data, d)
}

// GetArea 获取图像面积
func (s *ImageSize) GetArea() int {
	return s.Width * s.Height
}

// GetAspectRatio 获取宽高比
func (s *ImageSize) GetAspectRatio() float64 {
	if s.Height == 0 {
		return 0
	}
	return float64(s.Width) / float64(s.Height)
}

// IsSquare 检查是否为正方形
func (s *ImageSize) IsSquare() bool {
	return s.Width == s.Height
}

// IsLandscape 检查是否为横向
func (s *ImageSize) IsLandscape() bool {
	return s.Width > s.Height
}

// IsPortrait 检查是否为纵向
func (s *ImageSize) IsPortrait() bool {
	return s.Height > s.Width
}

// ParseImageSize 解析图像尺寸字符串
func ParseImageSize(sizeStr string) (*ImageSize, error) {
	parts := strings.Split(sizeStr, "x")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	var width, height int
	if _, err := fmt.Sscanf(parts[0], "%d", &width); err != nil {
		return nil, fmt.Errorf("invalid width: %s", parts[0])
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &height); err != nil {
		return nil, fmt.Errorf("invalid height: %s", parts[1])
	}

	return &ImageSize{Width: width, Height: height}, nil
}

// IsValidImageSize 检查图像尺寸是否有效
func IsValidImageSize(size string) bool {
	validSizes := []string{
		ImageSize256x256,
		ImageSize512x512,
		ImageSize1024x1024,
		ImageSize1792x1024,
		ImageSize1024x1792,
	}

	for _, validSize := range validSizes {
		if size == validSize {
			return true
		}
	}
	return false
}

// IsValidResponseFormat 检查响应格式是否有效
func IsValidResponseFormat(format string) bool {
	return format == ImageFormatURL || format == ImageFormatB64JSON
}

// IsValidImageQuality 检查图像质量是否有效
func IsValidImageQuality(quality string) bool {
	return quality == ImageQualityStandard || quality == ImageQualityHD
}

// IsValidImageStyle 检查图像风格是否有效
func IsValidImageStyle(style string) bool {
	return style == ImageStyleVivid || style == ImageStyleNatural
}

// IsValidEditOperation 检查编辑操作是否有效
func IsValidEditOperation(operation string) bool {
	validOperations := []string{
		ImageEditOperationInpaint,
		ImageEditOperationOutpaint,
		ImageEditOperationVariation,
	}

	for _, validOp := range validOperations {
		if operation == validOp {
			return true
		}
	}
	return false
}
