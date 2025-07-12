package types

import (
	"encoding/json"
	"strings"
)

// 音频格式常量
const (
	AudioFormatMP3  = "mp3"
	AudioFormatWAV  = "wav"
	AudioFormatFLAC = "flac"
	AudioFormatAAC  = "aac"
	AudioFormatOGG  = "ogg"
	AudioFormatWEBM = "webm"
	AudioFormatOPUS = "opus"
)

// 音频转录模型常量
const (
	AudioModelWhisper1   = "whisper-1"
	AudioModelWhisper2   = "whisper-2"
	AudioModelSenseVoice = "FunAudioLLM/SenseVoiceSmall"
)

// 音频语音合成模型常量
const (
	AudioModelTTS1   = "tts-1"
	AudioModelTTS1HD = "tts-1-hd"
)

// 音频语音常量
const (
	AudioVoiceAlloy   = "alloy"
	AudioVoiceEcho    = "echo"
	AudioVoiceFable   = "fable"
	AudioVoiceOnyx    = "onyx"
	AudioVoiceNova    = "nova"
	AudioVoiceShimmer = "shimmer"
)

// 音频语言常量
const (
	AudioLanguageAuto = "auto"
	AudioLanguageEN   = "en"
	AudioLanguageZH   = "zh"
	AudioLanguageJA   = "ja"
	AudioLanguageKO   = "ko"
	AudioLanguageFR   = "fr"
	AudioLanguageDE   = "de"
	AudioLanguageES   = "es"
	AudioLanguageRU   = "ru"
	AudioLanguageIT   = "it"
	AudioLanguagePT   = "pt"
)

// 音频响应格式常量
const (
	AudioResponseFormatJSON        = "json"
	AudioResponseFormatText        = "text"
	AudioResponseFormatSRT         = "srt"
	AudioResponseFormatVTT         = "verbose_json"
	AudioResponseFormatVerboseJSON = "verbose_json"
)

// AudioTranscriptionRequest 音频转录请求结构体
type AudioTranscriptionRequest struct {
	File                   string                 `json:"file"`
	Model                  string                 `json:"model"`
	Language               string                 `json:"language,omitempty"`
	Prompt                 string                 `json:"prompt,omitempty"`
	ResponseFormat         string                 `json:"response_format,omitempty"`
	Temperature            float64                `json:"temperature,omitempty"`
	TimestampGranularities []string               `json:"timestamp_granularities,omitempty"`
	ExtraBody              map[string]interface{} `json:"-"`
}

// AudioTranscriptionResponse 音频转录响应结构体
type AudioTranscriptionResponse struct {
	Text     string         `json:"text"`
	Language string         `json:"language,omitempty"`
	Duration float64        `json:"duration,omitempty"`
	Segments []AudioSegment `json:"segments,omitempty"`
	Words    []AudioWord    `json:"words,omitempty"`
	Error    *ErrorResponse `json:"error,omitempty"`
}

// AudioTranslationRequest 音频翻译请求结构体
type AudioTranslationRequest struct {
	File           string                 `json:"file"`
	Model          string                 `json:"model"`
	Prompt         string                 `json:"prompt,omitempty"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	Temperature    float64                `json:"temperature,omitempty"`
	ExtraBody      map[string]interface{} `json:"-"`
}

// AudioTranslationResponse 音频翻译响应结构体
type AudioTranslationResponse struct {
	Text     string         `json:"text"`
	Language string         `json:"language,omitempty"`
	Duration float64        `json:"duration,omitempty"`
	Segments []AudioSegment `json:"segments,omitempty"`
	Words    []AudioWord    `json:"words,omitempty"`
	Error    *ErrorResponse `json:"error,omitempty"`
}

// AudioSpeechRequest 音频语音合成请求结构体
type AudioSpeechRequest struct {
	Model          string                 `json:"model"`
	Input          string                 `json:"input"`
	Voice          string                 `json:"voice"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	Speed          float64                `json:"speed,omitempty"`
	ExtraBody      map[string]interface{} `json:"-"`
}

// AudioSpeechResponse 音频语音合成响应结构体
type AudioSpeechResponse struct {
	AudioContent []byte         `json:"audio_content,omitempty"`
	ContentType  string         `json:"content_type,omitempty"`
	Error        *ErrorResponse `json:"error,omitempty"`
}

// AudioSegment 音频片段结构体
type AudioSegment struct {
	ID               int         `json:"id"`
	Seek             int         `json:"seek"`
	Start            float64     `json:"start"`
	End              float64     `json:"end"`
	Text             string      `json:"text"`
	Tokens           []int       `json:"tokens"`
	Temperature      float64     `json:"temperature"`
	AvgLogprob       float64     `json:"avg_logprob"`
	CompressionRatio float64     `json:"compression_ratio"`
	NoSpeechProb     float64     `json:"no_speech_prob"`
	Words            []AudioWord `json:"words,omitempty"`
}

// AudioWord 音频单词结构体
type AudioWord struct {
	Word        string  `json:"word"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
	Probability float64 `json:"probability,omitempty"`
}

// AudioMetadata 音频元数据结构体
type AudioMetadata struct {
	Duration   float64 `json:"duration"`
	SampleRate int     `json:"sample_rate"`
	Channels   int     `json:"channels"`
	Format     string  `json:"format"`
	Bitrate    int     `json:"bitrate,omitempty"`
	FileSize   int64   `json:"file_size"`
	Language   string  `json:"language,omitempty"`
}

// AudioProcessingOptions 音频处理选项
type AudioProcessingOptions struct {
	NoiseReduction      bool   `json:"noise_reduction,omitempty"`
	VolumeNormalization bool   `json:"volume_normalization,omitempty"`
	SilenceRemoval      bool   `json:"silence_removal,omitempty"`
	TargetSampleRate    int    `json:"target_sample_rate,omitempty"`
	TargetBitrate       int    `json:"target_bitrate,omitempty"`
	TargetFormat        string `json:"target_format,omitempty"`
}

// NewAudioTranscriptionRequest 创建新的音频转录请求
func NewAudioTranscriptionRequest(file, model string) *AudioTranscriptionRequest {
	return &AudioTranscriptionRequest{
		File:  file,
		Model: model,
	}
}

// NewAudioTranslationRequest 创建新的音频翻译请求
func NewAudioTranslationRequest(file, model string) *AudioTranslationRequest {
	return &AudioTranslationRequest{
		File:  file,
		Model: model,
	}
}

// NewAudioSpeechRequest 创建新的音频语音合成请求
func NewAudioSpeechRequest(model, input, voice string) *AudioSpeechRequest {
	return &AudioSpeechRequest{
		Model: model,
		Input: input,
		Voice: voice,
	}
}

// ValidateParameters 验证音频转录请求参数
func (r *AudioTranscriptionRequest) ValidateParameters() error {
	if r.File == "" {
		return NewValidationError("file", r.File, "file is required", ErrCodeMissingParameter)
	}
	if r.Model == "" {
		return NewValidationError("model", r.Model, "model is required", ErrCodeMissingParameter)
	}

	// 验证模型
	if !IsValidAudioModel(r.Model) {
		return NewValidationError("model", r.Model, "invalid model", ErrCodeInvalidParameter)
	}

	// 验证语言
	if r.Language != "" && !IsValidAudioLanguage(r.Language) {
		return NewValidationError("language", r.Language, "invalid language", ErrCodeInvalidParameter)
	}

	// 验证响应格式
	if r.ResponseFormat != "" && !IsValidAudioResponseFormat(r.ResponseFormat) {
		return NewValidationError("response_format", r.ResponseFormat, "invalid response format", ErrCodeInvalidParameter)
	}

	// 验证温度
	if r.Temperature < 0 || r.Temperature > 1 {
		return NewValidationError("temperature", r.Temperature, "temperature must be between 0 and 1", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *AudioTranscriptionRequest) SetDefaults() {
	if r.Model == "" {
		r.Model = AudioModelWhisper1
	}
	if r.ResponseFormat == "" {
		r.ResponseFormat = AudioResponseFormatJSON
	}
	if r.Temperature == 0 {
		r.Temperature = 0.0
	}
}

// ToJSON 转换为JSON字符串
func (r *AudioTranscriptionRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *AudioTranscriptionRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ValidateParameters 验证音频翻译请求参数
func (r *AudioTranslationRequest) ValidateParameters() error {
	if r.File == "" {
		return NewValidationError("file", r.File, "file is required", ErrCodeMissingParameter)
	}
	if r.Model == "" {
		return NewValidationError("model", r.Model, "model is required", ErrCodeMissingParameter)
	}

	// 验证模型
	if !IsValidAudioModel(r.Model) {
		return NewValidationError("model", r.Model, "invalid model", ErrCodeInvalidParameter)
	}

	// 验证响应格式
	if r.ResponseFormat != "" && !IsValidAudioResponseFormat(r.ResponseFormat) {
		return NewValidationError("response_format", r.ResponseFormat, "invalid response format", ErrCodeInvalidParameter)
	}

	// 验证温度
	if r.Temperature < 0 || r.Temperature > 1 {
		return NewValidationError("temperature", r.Temperature, "temperature must be between 0 and 1", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *AudioTranslationRequest) SetDefaults() {
	if r.Model == "" {
		r.Model = AudioModelWhisper1
	}
	if r.ResponseFormat == "" {
		r.ResponseFormat = AudioResponseFormatJSON
	}
	if r.Temperature == 0 {
		r.Temperature = 0.0
	}
}

// ToJSON 转换为JSON字符串
func (r *AudioTranslationRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *AudioTranslationRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ValidateParameters 验证音频语音合成请求参数
func (r *AudioSpeechRequest) ValidateParameters() error {
	if r.Model == "" {
		return NewValidationError("model", r.Model, "model is required", ErrCodeMissingParameter)
	}
	if r.Input == "" {
		return NewValidationError("input", r.Input, "input is required", ErrCodeMissingParameter)
	}
	if r.Voice == "" {
		return NewValidationError("voice", r.Voice, "voice is required", ErrCodeMissingParameter)
	}

	// 验证模型
	if !IsValidTTSModel(r.Model) {
		return NewValidationError("model", r.Model, "invalid model", ErrCodeInvalidParameter)
	}

	// 验证语音
	if !IsValidAudioVoice(r.Voice) {
		return NewValidationError("voice", r.Voice, "invalid voice", ErrCodeInvalidParameter)
	}

	// 验证响应格式
	if r.ResponseFormat != "" && !IsValidAudioFormat(r.ResponseFormat) {
		return NewValidationError("response_format", r.ResponseFormat, "invalid response format", ErrCodeInvalidParameter)
	}

	// 验证语速
	if r.Speed < 0.25 || r.Speed > 4.0 {
		return NewValidationError("speed", r.Speed, "speed must be between 0.25 and 4.0", ErrCodeInvalidParameter)
	}

	// 验证输入长度
	if len(r.Input) > 4096 {
		return NewValidationError("input", r.Input, "input text is too long", ErrCodeInvalidParameter)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *AudioSpeechRequest) SetDefaults() {
	if r.Model == "" {
		r.Model = AudioModelTTS1
	}
	if r.Voice == "" {
		r.Voice = AudioVoiceAlloy
	}
	if r.ResponseFormat == "" {
		r.ResponseFormat = AudioFormatMP3
	}
	if r.Speed == 0 {
		r.Speed = 1.0
	}
}

// ToJSON 转换为JSON字符串
func (r *AudioSpeechRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *AudioSpeechRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *AudioTranscriptionResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *AudioTranscriptionResponse) GetError() *ErrorResponse {
	return r.Error
}

// GetSegmentCount 获取片段数量
func (r *AudioTranscriptionResponse) GetSegmentCount() int {
	return len(r.Segments)
}

// GetWordCount 获取单词数量
func (r *AudioTranscriptionResponse) GetWordCount() int {
	return len(r.Words)
}

// ToJSON 转换为JSON字符串
func (r *AudioTranscriptionResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *AudioTranscriptionResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *AudioTranslationResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *AudioTranslationResponse) GetError() *ErrorResponse {
	return r.Error
}

// ToJSON 转换为JSON字符串
func (r *AudioTranslationResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON 从JSON字符串解析
func (r *AudioTranslationResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// IsError 检查是否包含错误
func (r *AudioSpeechResponse) IsError() bool {
	return r.Error != nil
}

// GetError 获取错误信息
func (r *AudioSpeechResponse) GetError() *ErrorResponse {
	return r.Error
}

// HasContent 检查是否有音频内容
func (r *AudioSpeechResponse) HasContent() bool {
	return len(r.AudioContent) > 0
}

// GetContentSize 获取音频内容大小
func (r *AudioSpeechResponse) GetContentSize() int {
	return len(r.AudioContent)
}

// GetDuration 获取片段持续时间
func (s *AudioSegment) GetDuration() float64 {
	return s.End - s.Start
}

// GetWordCount 获取片段单词数量
func (s *AudioSegment) GetWordCount() int {
	return len(s.Words)
}

// GetTokenCount 获取片段Token数量
func (s *AudioSegment) GetTokenCount() int {
	return len(s.Tokens)
}

// GetDuration 获取单词持续时间
func (w *AudioWord) GetDuration() float64 {
	return w.End - w.Start
}

// GetBitrate 获取比特率
func (m *AudioMetadata) GetBitrate() int {
	if m.Bitrate > 0 {
		return m.Bitrate
	}
	// 估算比特率
	if m.Duration > 0 {
		return int(float64(m.FileSize*8) / m.Duration)
	}
	return 0
}

// IsHighQuality 检查是否为高质量音频
func (m *AudioMetadata) IsHighQuality() bool {
	return m.SampleRate >= 44100 && m.GetBitrate() >= 128000
}

// IsValidAudioFormat 检查音频格式是否有效
func IsValidAudioFormat(format string) bool {
	validFormats := []string{
		AudioFormatMP3,
		AudioFormatWAV,
		AudioFormatFLAC,
		AudioFormatAAC,
		AudioFormatOGG,
		AudioFormatWEBM,
		AudioFormatOPUS,
	}

	for _, validFormat := range validFormats {
		if strings.ToLower(format) == validFormat {
			return true
		}
	}
	return false
}

// IsValidAudioModel 检查音频模型是否有效
func IsValidAudioModel(model string) bool {
	validModels := []string{
		AudioModelWhisper1,
		AudioModelWhisper2,
		AudioModelSenseVoice,
	}

	for _, validModel := range validModels {
		if model == validModel {
			return true
		}
	}
	return false
}

// IsValidTTSModel 检查TTS模型是否有效
func IsValidTTSModel(model string) bool {
	validModels := []string{
		AudioModelTTS1,
		AudioModelTTS1HD,
	}

	for _, validModel := range validModels {
		if model == validModel {
			return true
		}
	}
	return false
}

// IsValidAudioVoice 检查音频语音是否有效
func IsValidAudioVoice(voice string) bool {
	validVoices := []string{
		AudioVoiceAlloy,
		AudioVoiceEcho,
		AudioVoiceFable,
		AudioVoiceOnyx,
		AudioVoiceNova,
		AudioVoiceShimmer,
	}

	for _, validVoice := range validVoices {
		if voice == validVoice {
			return true
		}
	}
	return false
}

// IsValidAudioLanguage 检查音频语言是否有效
func IsValidAudioLanguage(language string) bool {
	validLanguages := []string{
		AudioLanguageAuto,
		AudioLanguageEN,
		AudioLanguageZH,
		AudioLanguageJA,
		AudioLanguageKO,
		AudioLanguageFR,
		AudioLanguageDE,
		AudioLanguageES,
		AudioLanguageRU,
		AudioLanguageIT,
		AudioLanguagePT,
	}

	for _, validLang := range validLanguages {
		if language == validLang {
			return true
		}
	}
	return false
}

// IsValidAudioResponseFormat 检查音频响应格式是否有效
func IsValidAudioResponseFormat(format string) bool {
	validFormats := []string{
		AudioResponseFormatJSON,
		AudioResponseFormatText,
		AudioResponseFormatSRT,
		AudioResponseFormatVTT,
		AudioResponseFormatVerboseJSON,
	}

	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	return false
}
