package audio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/hewenyu/newapi-go/internal/transport"
	"github.com/hewenyu/newapi-go/internal/utils"
	"github.com/hewenyu/newapi-go/types"
	"go.uber.org/zap"
)

// AudioService 音频服务结构体
type AudioService struct {
	transport transport.HTTPTransport
	logger    utils.Logger
	config    *AudioConfig
	mu        sync.RWMutex
}

// NewAudioService 创建新的音频服务实例
func NewAudioService(transport transport.HTTPTransport, logger utils.Logger, options ...AudioOption) *AudioService {
	config := DefaultAudioConfig()

	// 应用选项
	for _, option := range options {
		option(config)
	}

	return &AudioService{
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

// CreateTranscription 创建音频转录
func (s *AudioService) CreateTranscription(ctx context.Context, audioFile string, options ...AudioOption) (*types.AudioTranscriptionResponse, error) {
	// 验证文件
	if audioFile == "" {
		return nil, fmt.Errorf("audio file path cannot be empty")
	}

	// 检查文件是否存在
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file does not exist: %s", audioFile)
	}

	// 创建配置副本并应用选项
	config := s.getConfig()
	for _, option := range options {
		option(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid audio config: %w", err)
	}

	// 构建请求
	req := config.ToTranscriptionRequest(audioFile)

	// 验证请求参数
	if err := req.ValidateParameters(); err != nil {
		return nil, fmt.Errorf("invalid transcription request: %w", err)
	}

	// 发送multipart请求
	resp, err := s.postMultipartFile(ctx, "/v1/audio/transcriptions", audioFile, req)
	if err != nil {
		s.logger.Error("Failed to create transcription", zap.Error(err))
		return nil, fmt.Errorf("failed to create transcription: %w", err)
	}

	// 解析响应
	var transcriptionResp types.AudioTranscriptionResponse
	if err := parseJSONResponse(resp, &transcriptionResp); err != nil {
		s.logger.Error("Failed to parse transcription response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查API错误
	if transcriptionResp.IsError() {
		apiErr := transcriptionResp.GetError()
		s.logger.Error("API returned error", zap.String("error", apiErr.Message))
		return nil, fmt.Errorf("API error: %s", apiErr.Message)
	}

	s.logger.Debug("Audio transcription created successfully", zap.String("text", transcriptionResp.Text[:min(50, len(transcriptionResp.Text))]))
	return &transcriptionResp, nil
}

// CreateTranslation 创建音频翻译
func (s *AudioService) CreateTranslation(ctx context.Context, audioFile string, options ...AudioOption) (*types.AudioTranslationResponse, error) {
	// TODO: 实现音频翻译功能
	// 当前版本暂不支持翻译功能，因为用户指定的模型主要用于识别
	return nil, fmt.Errorf("audio translation feature is not implemented yet")
}

// CreateSpeech 创建语音合成
func (s *AudioService) CreateSpeech(ctx context.Context, text string, options ...AudioOption) (*types.AudioSpeechResponse, error) {
	// TODO: 实现语音合成功能
	// 当前版本暂不支持语音合成功能，因为用户指定的模型主要用于识别
	return nil, fmt.Errorf("speech synthesis feature is not implemented yet")
}

// postMultipartFile 发送包含文件的multipart请求
func (s *AudioService) postMultipartFile(ctx context.Context, path, filename string, req interface{}) (*http.Response, error) {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 创建multipart buffer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// 复制文件内容
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// 添加其他字段
	if err := s.addFormFields(writer, req); err != nil {
		return nil, fmt.Errorf("failed to add form fields: %w", err)
	}

	// 关闭writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// 使用transport的PostMultipart方法
	return s.transport.PostMultipart(ctx, path, writer.Boundary(), body)
}

// addFormFields 添加表单字段
func (s *AudioService) addFormFields(writer *multipart.Writer, req interface{}) error {
	switch r := req.(type) {
	case *types.AudioTranscriptionRequest:
		if r.Model != "" {
			if err := writer.WriteField("model", r.Model); err != nil {
				return err
			}
		}
		if r.Language != "" {
			if err := writer.WriteField("language", r.Language); err != nil {
				return err
			}
		}
		if r.Prompt != "" {
			if err := writer.WriteField("prompt", r.Prompt); err != nil {
				return err
			}
		}
		if r.ResponseFormat != "" {
			if err := writer.WriteField("response_format", r.ResponseFormat); err != nil {
				return err
			}
		}
		if r.Temperature != 0 {
			if err := writer.WriteField("temperature", fmt.Sprintf("%f", r.Temperature)); err != nil {
				return err
			}
		}
		for _, granularity := range r.TimestampGranularities {
			if err := writer.WriteField("timestamp_granularities[]", granularity); err != nil {
				return err
			}
		}

		// 添加额外字段
		for key, value := range r.ExtraBody {
			if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
				return err
			}
		}

	case *types.AudioTranslationRequest:
		if r.Model != "" {
			if err := writer.WriteField("model", r.Model); err != nil {
				return err
			}
		}
		if r.Prompt != "" {
			if err := writer.WriteField("prompt", r.Prompt); err != nil {
				return err
			}
		}
		if r.ResponseFormat != "" {
			if err := writer.WriteField("response_format", r.ResponseFormat); err != nil {
				return err
			}
		}
		if r.Temperature != 0 {
			if err := writer.WriteField("temperature", fmt.Sprintf("%f", r.Temperature)); err != nil {
				return err
			}
		}

		// 添加额外字段
		for key, value := range r.ExtraBody {
			if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unsupported request type: %T", req)
	}

	return nil
}

// UpdateConfig 更新配置
func (s *AudioService) UpdateConfig(options ...AudioOption) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, option := range options {
		option(s.config)
	}
}

// GetConfig 获取配置
func (s *AudioService) GetConfig() *AudioConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config.Clone()
}

// getConfig 获取配置副本
func (s *AudioService) getConfig() *AudioConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config.Clone()
}

// ValidateAudioFile 验证音频文件
func (s *AudioService) ValidateAudioFile(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filename)
	}

	// 检查文件扩展名
	ext := filepath.Ext(filename)
	validExts := []string{".mp3", ".wav", ".flac", ".m4a", ".ogg", ".webm", ".mp4", ".mpeg", ".mpga", ".oga", ".opus"}

	for _, validExt := range validExts {
		if ext == validExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported file format: %s", ext)
}

// GetSupportedFormats 获取支持的音频格式
func (s *AudioService) GetSupportedFormats() []string {
	return []string{
		types.AudioFormatMP3,
		types.AudioFormatWAV,
		types.AudioFormatFLAC,
		types.AudioFormatAAC,
		types.AudioFormatOGG,
		types.AudioFormatWEBM,
		types.AudioFormatOPUS,
	}
}

// GetMaxFileSize 获取最大文件大小（25MB）
func (s *AudioService) GetMaxFileSize() int64 {
	return 25 * 1024 * 1024 // 25MB
}

// min 辅助函数，返回两个整数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
