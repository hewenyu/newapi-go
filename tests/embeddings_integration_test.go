package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hewenyu/newapi-go/client"
	"github.com/hewenyu/newapi-go/config"
	"github.com/hewenyu/newapi-go/services/embeddings"
	"github.com/hewenyu/newapi-go/types"
)

const (
	// 通用多语言模型
	multilingualModel = "BAAI/bge-m3"
	// 中文模型
	chineseModel = "BAAI/bge-large-zh-v1.5"
	// 英文模型
	englishModel = "BAAI/bge-large-en-v1.5"
)

// setupEmbeddingAPIClient 设置真实的API客户端
func setupEmbeddingAPIClient(t *testing.T) *client.Client {
	baseURL := os.Getenv("NEW_API")
	apiKey := os.Getenv("NEW_API_KEY")

	if baseURL == "" || apiKey == "" {
		t.Skip("Skipping integration test: NEW_API or NEW_API_KEY not set")
	}

	// 使用默认配置作为基础
	cfg := config.DefaultConfig()
	cfg.BaseURL = baseURL
	cfg.APIKey = apiKey
	cfg.Timeout = 30 * time.Second
	cfg.Debug = true

	c, err := client.NewClient(client.WithConfig(cfg))
	require.NoError(t, err)

	return c
}

// TestEmbeddingRealAPICreateEmbedding 测试真实API的单个文本嵌入
func TestEmbeddingRealAPICreateEmbedding(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	testText := "Hello, this is a test sentence for embedding."

	response, err := c.CreateEmbedding(ctx, testText,
		embeddings.WithModel(multilingualModel),
	)

	if err != nil {
		t.Logf("API error: %v", err)
		// 如果是模型不可用的错误，跳过测试
		if strings.Contains(err.Error(), "无可用渠道") || strings.Contains(err.Error(), "not available") {
			t.Skip("Model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	if response != nil {
		assert.NotEmpty(t, response.Data)
		assert.Equal(t, 1, len(response.Data))
		assert.NotEmpty(t, response.Data[0].Embedding)
		assert.Greater(t, len(response.Data[0].Embedding), 0)
		assert.Equal(t, multilingualModel, response.Model)

		t.Logf("Model: %s", response.Model)
		t.Logf("Embedding dimensions: %d", len(response.Data[0].Embedding))
		t.Logf("Token usage: %d", response.Usage.TotalTokens)
	}
}

// TestEmbeddingRealAPICreateEmbeddings 测试真实API的批量文本嵌入
func TestEmbeddingRealAPICreateEmbeddings(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	testTexts := []string{
		"The quick brown fox jumps over the lazy dog.",
		"Machine learning is a subset of artificial intelligence.",
		"Python is a popular programming language.",
	}

	response, err := c.CreateEmbeddings(ctx, testTexts,
		embeddings.WithModel(englishModel),
	)

	if err != nil {
		t.Logf("API error: %v", err)
		// 如果是模型不可用的错误，跳过测试
		if strings.Contains(err.Error(), "无可用渠道") || strings.Contains(err.Error(), "not available") {
			t.Skip("Model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	assert.Equal(t, len(testTexts), len(response.Data))

	for i, embedding := range response.Data {
		assert.NotEmpty(t, embedding.Embedding)
		assert.Equal(t, i, embedding.Index)
		t.Logf("Text %d embedding dimensions: %d", i, len(embedding.Embedding))
	}

	t.Logf("Model: %s", response.Model)
	t.Logf("Total embeddings: %d", len(response.Data))
	t.Logf("Token usage: %d", response.Usage.TotalTokens)
}

// TestEmbeddingRealAPIChineseEmbedding 测试真实API的中文文本嵌入
func TestEmbeddingRealAPIChineseEmbedding(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	chineseTexts := []string{
		"你好世界，这是一个测试句子。",
		"人工智能是计算机科学的一个分支。",
		"机器学习在现代技术中扮演重要角色。",
	}

	response, err := c.CreateEmbeddings(ctx, chineseTexts,
		embeddings.WithModel(chineseModel),
	)

	if err != nil {
		t.Logf("API error: %v", err)
		// 如果是模型不可用的错误，跳过测试
		if strings.Contains(err.Error(), "无可用渠道") || strings.Contains(err.Error(), "not available") {
			t.Skip("Model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	assert.Equal(t, len(chineseTexts), len(response.Data))

	for i, embedding := range response.Data {
		assert.NotEmpty(t, embedding.Embedding)
		assert.Equal(t, i, embedding.Index)
		t.Logf("Chinese text %d embedding dimensions: %d", i, len(embedding.Embedding))
	}

	t.Logf("Model: %s", response.Model)
	t.Logf("Total Chinese embeddings: %d", len(response.Data))
	t.Logf("Token usage: %d", response.Usage.TotalTokens)
}

// TestEmbeddingRealAPIMultilingualEmbedding 测试真实API的多语言文本嵌入
func TestEmbeddingRealAPIMultilingualEmbedding(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	multilingualTexts := []string{
		"Hello world in English",
		"你好世界用中文",
		"Bonjour le monde en français",
		"Hola mundo en español",
	}

	response, err := c.CreateEmbeddings(ctx, multilingualTexts,
		embeddings.WithModel(multilingualModel),
	)

	if err != nil {
		t.Logf("API error: %v", err)
		// 如果是模型不可用的错误，跳过测试
		if strings.Contains(err.Error(), "无可用渠道") || strings.Contains(err.Error(), "not available") {
			t.Skip("Model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	assert.Equal(t, len(multilingualTexts), len(response.Data))

	for i, embedding := range response.Data {
		assert.NotEmpty(t, embedding.Embedding)
		assert.Equal(t, i, embedding.Index)
		t.Logf("Multilingual text %d embedding dimensions: %d", i, len(embedding.Embedding))
	}

	t.Logf("Model: %s", response.Model)
	t.Logf("Total multilingual embeddings: %d", len(response.Data))
	t.Logf("Token usage: %d", response.Usage.TotalTokens)
}

// TestEmbeddingRealAPIMultipleModels 测试真实API的多种嵌入模型
func TestEmbeddingRealAPIMultipleModels(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	testText := "This is a test sentence for comparing different embedding models."
	models := []string{multilingualModel, chineseModel, englishModel}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			response, err := c.CreateEmbedding(ctx, testText,
				embeddings.WithModel(model),
			)

			if err != nil {
				t.Logf("Model %s failed: %v", model, err)
				return
			}

			assert.NotNil(t, response)
			assert.NotEmpty(t, response.Data)
			assert.Equal(t, model, response.Model)

			t.Logf("Model %s - Dimensions: %d, Tokens: %d",
				model,
				len(response.Data[0].Embedding),
				response.Usage.TotalTokens)
		})
	}
}

// TestEmbeddingRealAPIEmbeddingOptions 测试真实API的嵌入选项
func TestEmbeddingRealAPIEmbeddingOptions(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	testText := "Test text for embedding options."

	// 测试不同的编码格式
	response, err := c.CreateEmbedding(ctx, testText,
		embeddings.WithModel(englishModel),
		embeddings.WithEncodingFormat(types.EmbeddingEncodingFormatFloat),
		embeddings.WithUser("test-user"),
	)

	if err != nil {
		t.Logf("API error: %v", err)
		// 如果是模型不可用的错误，跳过测试
		if strings.Contains(err.Error(), "无可用渠道") || strings.Contains(err.Error(), "not available") {
			t.Skip("Model not available, skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Data)
	assert.Equal(t, englishModel, response.Model)

	t.Logf("Options test - Model: %s, Dimensions: %d",
		response.Model,
		len(response.Data[0].Embedding))
}

// TestEmbeddingRealAPITokenUsage 测试真实API的Token使用情况
func TestEmbeddingRealAPITokenUsage(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	// 测试不同长度的文本
	texts := []string{
		"Short text.",
		"This is a medium length text that contains more words and should use more tokens.",
		"This is a very long text that contains many words and sentences. It should demonstrate how token usage scales with text length. The embedding API should report the total number of tokens used for processing this text.",
	}

	for i, text := range texts {
		t.Run(fmt.Sprintf("Length_%d", i), func(t *testing.T) {
			response, err := c.CreateEmbedding(ctx, text,
				embeddings.WithModel(multilingualModel),
			)

			if err != nil {
				t.Logf("API error: %v", err)
				// 如果是模型不可用的错误，跳过测试
				if strings.Contains(err.Error(), "无可用渠道") || strings.Contains(err.Error(), "not available") {
					t.Skip("Model not available, skipping test")
				}
				t.Fatalf("Unexpected error: %v", err)
			}

			assert.NotNil(t, response)
			assert.Greater(t, response.Usage.TotalTokens, 0)

			t.Logf("Text length: %d chars, Tokens: %d",
				len(text),
				response.Usage.TotalTokens)
		})
	}
}

// TestEmbeddingRealAPIErrorHandling 测试真实API的错误处理
func TestEmbeddingRealAPIErrorHandling(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	// 测试无效模型
	response, err := c.CreateEmbedding(ctx, "Test text",
		embeddings.WithModel("invalid-embedding-model"),
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	t.Logf("Expected error for invalid model: %v", err)

	// 测试空文本
	response, err = c.CreateEmbedding(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, response)
	t.Logf("Expected error for empty text: %v", err)

	// 测试空文本数组
	response, err = c.CreateEmbeddings(ctx, []string{})
	assert.Error(t, err)
	assert.Nil(t, response)
	t.Logf("Expected error for empty text array: %v", err)
}

// TestEmbeddingRealAPIInputValidation 测试真实API的输入验证
func TestEmbeddingRealAPIInputValidation(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	// 测试输入验证函数
	err := c.ValidateEmbeddingInput("valid text")
	assert.NoError(t, err)

	err = c.ValidateEmbeddingInput([]string{"text1", "text2"})
	assert.NoError(t, err)

	err = c.ValidateEmbeddingInput("")
	assert.Error(t, err)

	err = c.ValidateEmbeddingInput([]string{})
	assert.Error(t, err)

	err = c.ValidateEmbeddingInput(123)
	assert.Error(t, err)

	t.Log("Input validation tests completed")
}

// TestEmbeddingRealAPIContextCancellation 测试真实API的上下文取消
func TestEmbeddingRealAPIContextCancellation(t *testing.T) {
	c := setupEmbeddingAPIClient(t)
	defer c.Close()

	// 创建一个可取消的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 尝试创建嵌入（应该被取消）
	response, err := c.CreateEmbedding(ctx, "This request should be cancelled",
		embeddings.WithModel(multilingualModel),
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "context")
	t.Logf("Expected context cancellation error: %v", err)
}

// BenchmarkEmbeddingRealAPICreateEmbedding 基准测试真实API的嵌入创建
func BenchmarkEmbeddingRealAPICreateEmbedding(b *testing.B) {
	baseURL := os.Getenv("NEW_API")
	apiKey := os.Getenv("NEW_API_KEY")

	if baseURL == "" || apiKey == "" {
		b.Skip("Skipping benchmark: NEW_API or NEW_API_KEY not set")
	}

	cfg := config.DefaultConfig()
	cfg.BaseURL = baseURL
	cfg.APIKey = apiKey
	cfg.Timeout = 30 * time.Second

	c, err := client.NewClient(client.WithConfig(cfg))
	require.NoError(b, err)
	defer c.Close()

	ctx := context.Background()
	testText := "Benchmark test text for embedding creation."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.CreateEmbedding(ctx, testText,
			embeddings.WithModel(multilingualModel),
		)
		if err != nil {
			b.Errorf("Benchmark failed: %v", err)
		}
	}
}
