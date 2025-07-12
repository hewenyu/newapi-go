package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hewenyu/newapi-go/client"
	"github.com/hewenyu/newapi-go/config"
	"github.com/hewenyu/newapi-go/services/audio"
	"github.com/hewenyu/newapi-go/types"
)

const (
	defaultAudioModel = "FunAudioLLM/SenseVoiceSmall"
	testAudioFile     = "audio/300_oral5.wav"
	expectedText      = "好久不见。你还记得咱们大学那会儿吗？你听到的是开源项目 T T S List。那可是风华正茂的岁月啊！还记得咱俩爬那个山顶看日出吗？当时许的愿望，我到现在还记得呢。"
)

// setupRealAPIClientForAudio 设置真实的API客户端（音频测试专用）
func setupRealAPIClientForAudio(t testing.TB) *client.Client {
	baseURL := os.Getenv("NEW_API")
	apiKey := os.Getenv("NEW_API_KEY")

	if baseURL == "" || apiKey == "" {
		t.Skip("Skipping integration test: NEW_API or NEW_API_KEY not set")
	}

	// 使用默认配置作为基础
	cfg := config.DefaultConfig()
	cfg.BaseURL = baseURL
	cfg.APIKey = apiKey
	cfg.Timeout = 60 * time.Second // 音频处理可能需要更长时间
	cfg.Debug = true

	c, err := client.NewClient(client.WithConfig(cfg))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return c
}

// checkAudioFile 检查音频文件是否存在
func checkAudioFile(t testing.TB, filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("Skipping audio test: test file %s does not exist", filename)
	}
}

// TestRealAPIAudioTranscription 测试真实API的音频转录
func TestRealAPIAudioTranscription(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	// 检查音频文件
	absPath, err := filepath.Abs(testAudioFile)
	require.NoError(t, err)
	checkAudioFile(t, absPath)

	ctx := context.Background()

	// 测试音频转录
	response, err := c.CreateTranscription(ctx, absPath,
		audio.WithTranscriptionModel(defaultAudioModel),
		audio.WithTranscriptionLanguage("zh"),
		audio.WithTranscriptionResponseFormat("json"),
		audio.WithTranscriptionTemperature(0.0),
	)

	if err != nil {
		t.Logf("Audio transcription error: %v", err)
		// 如果是模型不可用的错误，跳过测试
		if strings.Contains(err.Error(), "无可用渠道") ||
			strings.Contains(err.Error(), "not available") ||
			strings.Contains(err.Error(), "model not found") {
			t.Skip("Audio model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Text)

	t.Logf("Transcription result: %s", response.Text)
	t.Logf("Expected text: %s", expectedText)

	// 验证转录结果包含预期的关键词
	transcribedText := strings.ToLower(response.Text)
	keywords := []string{"好久不见", "大学", "开源项目", "山顶", "日出"}

	foundKeywords := 0
	for _, keyword := range keywords {
		if strings.Contains(transcribedText, strings.ToLower(keyword)) {
			foundKeywords++
			t.Logf("Found keyword: %s", keyword)
		}
	}

	// 至少应该找到一半的关键词
	assert.GreaterOrEqual(t, foundKeywords, len(keywords)/2,
		"Should find at least half of the expected keywords")
}

// TestRealAPIAudioTranscriptionWithVerboseResponse 测试详细响应格式
func TestRealAPIAudioTranscriptionWithVerboseResponse(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	// 检查音频文件
	absPath, err := filepath.Abs(testAudioFile)
	require.NoError(t, err)
	checkAudioFile(t, absPath)

	ctx := context.Background()

	// 测试详细响应格式
	response, err := c.CreateTranscription(ctx, absPath,
		audio.WithTranscriptionModel(defaultAudioModel),
		audio.WithTranscriptionLanguage("zh"),
		audio.WithTranscriptionResponseFormat("verbose_json"),
		audio.WithTranscriptionTemperature(0.0),
	)

	if err != nil {
		t.Logf("Audio transcription error: %v", err)
		if strings.Contains(err.Error(), "无可用渠道") ||
			strings.Contains(err.Error(), "not available") ||
			strings.Contains(err.Error(), "model not found") {
			t.Skip("Audio model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Text)

	t.Logf("Transcription result: %s", response.Text)
	t.Logf("Language: %s", response.Language)
	t.Logf("Duration: %.2f seconds", response.Duration)

	// 验证详细信息
	if response.Language != "" {
		assert.Contains(t, []string{"zh", "chinese", "zh-cn"}, strings.ToLower(response.Language))
	}
	if response.Duration > 0 {
		assert.Greater(t, response.Duration, 0.0)
	}
}

// TestRealAPIAudioTranscriptionWithPrompt 测试使用提示词的音频转录
func TestRealAPIAudioTranscriptionWithPrompt(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	// 检查音频文件
	absPath, err := filepath.Abs(testAudioFile)
	require.NoError(t, err)
	checkAudioFile(t, absPath)

	ctx := context.Background()

	// 使用提示词来改善转录结果
	prompt := "这是一段关于大学生活回忆的对话，包含开源项目等技术词汇。"

	response, err := c.CreateTranscription(ctx, absPath,
		audio.WithTranscriptionModel(defaultAudioModel),
		audio.WithTranscriptionLanguage("zh"),
		audio.WithTranscriptionResponseFormat("json"),
		audio.WithTranscriptionPrompt(prompt),
		audio.WithTranscriptionTemperature(0.0),
	)

	if err != nil {
		t.Logf("Audio transcription error: %v", err)
		if strings.Contains(err.Error(), "无可用渠道") ||
			strings.Contains(err.Error(), "not available") ||
			strings.Contains(err.Error(), "model not found") {
			t.Skip("Audio model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Text)

	t.Logf("Transcription with prompt result: %s", response.Text)
}

// TestRealAPIAudioTranslation 测试音频翻译（目前未实现）
func TestRealAPIAudioTranslation(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	// 检查音频文件
	absPath, err := filepath.Abs(testAudioFile)
	require.NoError(t, err)
	checkAudioFile(t, absPath)

	ctx := context.Background()

	// 测试音频翻译（预期会失败，因为未实现）
	_, err = c.CreateTranslation(ctx, absPath,
		audio.WithTranslationModel(defaultAudioModel),
		audio.WithTranslationResponseFormat("json"),
	)

	// 应该返回"未实现"错误
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	t.Logf("Translation error (expected): %v", err)
}

// TestRealAPIAudioSpeech 测试语音合成（目前未实现）
func TestRealAPIAudioSpeech(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	ctx := context.Background()

	// 测试语音合成（预期会失败，因为未实现）
	_, err := c.CreateSpeech(ctx, "Hello world",
		audio.WithSpeechModel("tts-1"),
		audio.WithSpeechVoice("alloy"),
		audio.WithSpeechResponseFormat("mp3"),
	)

	// 应该返回"未实现"错误
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	t.Logf("Speech synthesis error (expected): %v", err)
}

// TestRealAPIAudioFileValidation 测试音频文件验证
func TestRealAPIAudioFileValidation(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	// 测试有效的音频文件
	absPath, err := filepath.Abs(testAudioFile)
	require.NoError(t, err)

	if _, err := os.Stat(absPath); err == nil {
		err = c.ValidateAudioFile(absPath)
		assert.NoError(t, err)
	}

	// 测试无效的文件
	err = c.ValidateAudioFile("nonexistent.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file does not exist")

	// 测试无效的文件格式
	err = c.ValidateAudioFile("test.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file format")
}

// TestRealAPIAudioServiceInfo 测试音频服务信息
func TestRealAPIAudioServiceInfo(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	// 测试支持的格式
	formats := c.GetSupportedAudioFormats()
	assert.NotEmpty(t, formats)
	assert.Contains(t, formats, types.AudioFormatMP3)
	assert.Contains(t, formats, types.AudioFormatWAV)
	t.Logf("Supported formats: %v", formats)

	// 测试最大文件大小
	maxSize := c.GetMaxAudioFileSize()
	assert.Equal(t, int64(25*1024*1024), maxSize) // 25MB
	t.Logf("Max file size: %d bytes (%.2f MB)", maxSize, float64(maxSize)/(1024*1024))
}

// TestRealAPIAudioErrorHandling 测试音频API错误处理
func TestRealAPIAudioErrorHandling(t *testing.T) {
	c := setupRealAPIClientForAudio(t)
	defer c.Close()

	ctx := context.Background()

	// 测试空文件路径
	_, err := c.CreateTranscription(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audio file path cannot be empty")

	// 测试不存在的文件
	_, err = c.CreateTranscription(ctx, "nonexistent.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audio file does not exist")
}

// BenchmarkRealAPIAudioTranscription 基准测试音频转录
func BenchmarkRealAPIAudioTranscription(b *testing.B) {
	c := setupRealAPIClientForAudio(b)
	defer c.Close()

	// 检查音频文件
	absPath, err := filepath.Abs(testAudioFile)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		b.Skip("Test audio file does not exist")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.CreateTranscription(ctx, absPath,
			audio.WithTranscriptionModel(defaultAudioModel),
			audio.WithTranscriptionLanguage("zh"),
			audio.WithTranscriptionResponseFormat("json"),
		)
		if err != nil {
			b.Logf("Benchmark error: %v", err)
		}
	}
}
