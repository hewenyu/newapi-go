package audio

import (
	"github.com/hewenyu/newapi-go/types"
)

// AudioOption 音频选项类型
type AudioOption func(*AudioConfig)

// AudioConfig 音频配置结构体
type AudioConfig struct {
	// 转录相关配置
	TranscriptionModel          string
	TranscriptionLanguage       string
	TranscriptionResponseFormat string
	TranscriptionPrompt         string
	TranscriptionTemperature    float64
	TimestampGranularities      []string

	// 翻译相关配置
	TranslationModel          string
	TranslationResponseFormat string
	TranslationPrompt         string
	TranslationTemperature    float64

	// 语音合成相关配置
	SpeechModel          string
	SpeechVoice          string
	SpeechResponseFormat string
	SpeechSpeed          float64

	// 通用配置
	ExtraBody map[string]interface{}
}

// DefaultAudioConfig 返回默认音频配置
func DefaultAudioConfig() *AudioConfig {
	return &AudioConfig{
		TranscriptionModel:          types.AudioModelWhisper1,
		TranscriptionLanguage:       types.AudioLanguageAuto,
		TranscriptionResponseFormat: types.AudioResponseFormatJSON,
		TranscriptionPrompt:         "",
		TranscriptionTemperature:    0.0,
		TimestampGranularities:      []string{},

		TranslationModel:          types.AudioModelWhisper1,
		TranslationResponseFormat: types.AudioResponseFormatJSON,
		TranslationPrompt:         "",
		TranslationTemperature:    0.0,

		SpeechModel:          types.AudioModelTTS1,
		SpeechVoice:          types.AudioVoiceAlloy,
		SpeechResponseFormat: types.AudioFormatMP3,
		SpeechSpeed:          1.0,

		ExtraBody: make(map[string]interface{}),
	}
}

// WithTranscriptionModel 设置转录模型
func WithTranscriptionModel(model string) AudioOption {
	return func(config *AudioConfig) {
		config.TranscriptionModel = model
	}
}

// WithTranscriptionLanguage 设置转录语言
func WithTranscriptionLanguage(language string) AudioOption {
	return func(config *AudioConfig) {
		config.TranscriptionLanguage = language
	}
}

// WithTranscriptionResponseFormat 设置转录响应格式
func WithTranscriptionResponseFormat(format string) AudioOption {
	return func(config *AudioConfig) {
		config.TranscriptionResponseFormat = format
	}
}

// WithTranscriptionPrompt 设置转录提示词
func WithTranscriptionPrompt(prompt string) AudioOption {
	return func(config *AudioConfig) {
		config.TranscriptionPrompt = prompt
	}
}

// WithTranscriptionTemperature 设置转录温度
func WithTranscriptionTemperature(temperature float64) AudioOption {
	return func(config *AudioConfig) {
		config.TranscriptionTemperature = temperature
	}
}

// WithTimestampGranularities 设置时间戳粒度
func WithTimestampGranularities(granularities []string) AudioOption {
	return func(config *AudioConfig) {
		config.TimestampGranularities = granularities
	}
}

// WithTranslationModel 设置翻译模型
func WithTranslationModel(model string) AudioOption {
	return func(config *AudioConfig) {
		config.TranslationModel = model
	}
}

// WithTranslationResponseFormat 设置翻译响应格式
func WithTranslationResponseFormat(format string) AudioOption {
	return func(config *AudioConfig) {
		config.TranslationResponseFormat = format
	}
}

// WithTranslationPrompt 设置翻译提示词
func WithTranslationPrompt(prompt string) AudioOption {
	return func(config *AudioConfig) {
		config.TranslationPrompt = prompt
	}
}

// WithTranslationTemperature 设置翻译温度
func WithTranslationTemperature(temperature float64) AudioOption {
	return func(config *AudioConfig) {
		config.TranslationTemperature = temperature
	}
}

// WithSpeechModel 设置语音合成模型
func WithSpeechModel(model string) AudioOption {
	return func(config *AudioConfig) {
		config.SpeechModel = model
	}
}

// WithSpeechVoice 设置语音合成声音
func WithSpeechVoice(voice string) AudioOption {
	return func(config *AudioConfig) {
		config.SpeechVoice = voice
	}
}

// WithSpeechResponseFormat 设置语音合成响应格式
func WithSpeechResponseFormat(format string) AudioOption {
	return func(config *AudioConfig) {
		config.SpeechResponseFormat = format
	}
}

// WithSpeechSpeed 设置语音合成速度
func WithSpeechSpeed(speed float64) AudioOption {
	return func(config *AudioConfig) {
		config.SpeechSpeed = speed
	}
}

// WithExtraBody 设置额外的请求参数
func WithExtraBody(key string, value interface{}) AudioOption {
	return func(config *AudioConfig) {
		if config.ExtraBody == nil {
			config.ExtraBody = make(map[string]interface{})
		}
		config.ExtraBody[key] = value
	}
}

// Validate 验证音频配置
func (c *AudioConfig) Validate() error {
	// 验证转录模型
	if c.TranscriptionModel != "" && !types.IsValidAudioModel(c.TranscriptionModel) {
		return types.NewValidationError("transcription_model", c.TranscriptionModel, "invalid transcription model", types.ErrCodeInvalidParameter)
	}

	// 验证转录语言
	if c.TranscriptionLanguage != "" && !types.IsValidAudioLanguage(c.TranscriptionLanguage) {
		return types.NewValidationError("transcription_language", c.TranscriptionLanguage, "invalid transcription language", types.ErrCodeInvalidParameter)
	}

	// 验证转录响应格式
	if c.TranscriptionResponseFormat != "" && !types.IsValidAudioResponseFormat(c.TranscriptionResponseFormat) {
		return types.NewValidationError("transcription_response_format", c.TranscriptionResponseFormat, "invalid transcription response format", types.ErrCodeInvalidParameter)
	}

	// 验证转录温度
	if c.TranscriptionTemperature < 0 || c.TranscriptionTemperature > 1 {
		return types.NewValidationError("transcription_temperature", c.TranscriptionTemperature, "transcription temperature must be between 0 and 1", types.ErrCodeInvalidParameter)
	}

	// 验证翻译模型
	if c.TranslationModel != "" && !types.IsValidAudioModel(c.TranslationModel) {
		return types.NewValidationError("translation_model", c.TranslationModel, "invalid translation model", types.ErrCodeInvalidParameter)
	}

	// 验证翻译响应格式
	if c.TranslationResponseFormat != "" && !types.IsValidAudioResponseFormat(c.TranslationResponseFormat) {
		return types.NewValidationError("translation_response_format", c.TranslationResponseFormat, "invalid translation response format", types.ErrCodeInvalidParameter)
	}

	// 验证翻译温度
	if c.TranslationTemperature < 0 || c.TranslationTemperature > 1 {
		return types.NewValidationError("translation_temperature", c.TranslationTemperature, "translation temperature must be between 0 and 1", types.ErrCodeInvalidParameter)
	}

	// 验证语音合成模型
	if c.SpeechModel != "" && !types.IsValidTTSModel(c.SpeechModel) {
		return types.NewValidationError("speech_model", c.SpeechModel, "invalid speech model", types.ErrCodeInvalidParameter)
	}

	// 验证语音合成声音
	if c.SpeechVoice != "" && !types.IsValidAudioVoice(c.SpeechVoice) {
		return types.NewValidationError("speech_voice", c.SpeechVoice, "invalid speech voice", types.ErrCodeInvalidParameter)
	}

	// 验证语音合成响应格式
	if c.SpeechResponseFormat != "" && !types.IsValidAudioFormat(c.SpeechResponseFormat) {
		return types.NewValidationError("speech_response_format", c.SpeechResponseFormat, "invalid speech response format", types.ErrCodeInvalidParameter)
	}

	// 验证语音合成速度
	if c.SpeechSpeed < 0.25 || c.SpeechSpeed > 4.0 {
		return types.NewValidationError("speech_speed", c.SpeechSpeed, "speech speed must be between 0.25 and 4.0", types.ErrCodeInvalidParameter)
	}

	return nil
}

// Clone 克隆音频配置
func (c *AudioConfig) Clone() *AudioConfig {
	clone := &AudioConfig{
		TranscriptionModel:          c.TranscriptionModel,
		TranscriptionLanguage:       c.TranscriptionLanguage,
		TranscriptionResponseFormat: c.TranscriptionResponseFormat,
		TranscriptionPrompt:         c.TranscriptionPrompt,
		TranscriptionTemperature:    c.TranscriptionTemperature,
		TimestampGranularities:      make([]string, len(c.TimestampGranularities)),

		TranslationModel:          c.TranslationModel,
		TranslationResponseFormat: c.TranslationResponseFormat,
		TranslationPrompt:         c.TranslationPrompt,
		TranslationTemperature:    c.TranslationTemperature,

		SpeechModel:          c.SpeechModel,
		SpeechVoice:          c.SpeechVoice,
		SpeechResponseFormat: c.SpeechResponseFormat,
		SpeechSpeed:          c.SpeechSpeed,

		ExtraBody: make(map[string]interface{}),
	}

	// 深拷贝切片
	copy(clone.TimestampGranularities, c.TimestampGranularities)

	// 深拷贝map
	for k, v := range c.ExtraBody {
		clone.ExtraBody[k] = v
	}

	return clone
}

// ToTranscriptionRequest 转换为转录请求
func (c *AudioConfig) ToTranscriptionRequest(filename string) *types.AudioTranscriptionRequest {
	req := &types.AudioTranscriptionRequest{
		File:                   filename,
		Model:                  c.TranscriptionModel,
		Language:               c.TranscriptionLanguage,
		ResponseFormat:         c.TranscriptionResponseFormat,
		Prompt:                 c.TranscriptionPrompt,
		Temperature:            c.TranscriptionTemperature,
		TimestampGranularities: c.TimestampGranularities,
		ExtraBody:              c.ExtraBody,
	}

	req.SetDefaults()
	return req
}

// ToTranslationRequest 转换为翻译请求
func (c *AudioConfig) ToTranslationRequest(filename string) *types.AudioTranslationRequest {
	req := &types.AudioTranslationRequest{
		File:           filename,
		Model:          c.TranslationModel,
		ResponseFormat: c.TranslationResponseFormat,
		Prompt:         c.TranslationPrompt,
		Temperature:    c.TranslationTemperature,
		ExtraBody:      c.ExtraBody,
	}

	req.SetDefaults()
	return req
}

// ToSpeechRequest 转换为语音合成请求
func (c *AudioConfig) ToSpeechRequest(input string) *types.AudioSpeechRequest {
	req := &types.AudioSpeechRequest{
		Model:          c.SpeechModel,
		Input:          input,
		Voice:          c.SpeechVoice,
		ResponseFormat: c.SpeechResponseFormat,
		Speed:          c.SpeechSpeed,
		ExtraBody:      c.ExtraBody,
	}

	req.SetDefaults()
	return req
}
