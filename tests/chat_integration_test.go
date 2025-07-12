package tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hewenyu/newapi-go/client"
	"github.com/hewenyu/newapi-go/config"
	"github.com/hewenyu/newapi-go/services/chat"
	"github.com/hewenyu/newapi-go/types"
)

const (
	defaultModel = "gpt-4.1-mini"
)

// setupRealAPIClient 设置真实的API客户端
func setupRealAPIClient(t *testing.T) *client.Client {
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

// TestRealAPISimpleChat 测试真实API的简单聊天
func TestRealAPISimpleChat(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	response, err := c.SimpleChat(ctx, "Hello, please say 'Hi' back to me.",
		chat.WithModel(defaultModel),
		chat.WithMaxTokens(50),
		chat.WithTemperature(0.7),
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
		assert.NotEmpty(t, response.ID)
		assert.NotEmpty(t, response.Choices)
		if len(response.Choices) > 0 {
			assert.NotEmpty(t, response.Choices[0].Message.GetTextContent())
			t.Logf("Response ID: %s", response.ID)
			t.Logf("Response Content: %s", response.Choices[0].Message.GetTextContent())
		}
	}
}

// TestRealAPIChatWithSystem 测试真实API的系统消息聊天
func TestRealAPIChatWithSystem(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	response, err := c.ChatWithSystem(ctx,
		"You are a helpful assistant. Always respond in English.",
		"What is the capital of France?",
		chat.WithModel(defaultModel),
		chat.WithMaxTokens(100),
		chat.WithTemperature(0.3),
	)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Choices)
	assert.Contains(t, response.Choices[0].Message.GetTextContent(), "Paris")

	t.Logf("Response: %s", response.Choices[0].Message.GetTextContent())
}

// TestRealAPIChatWithHistory 测试真实API的历史对话
func TestRealAPIChatWithHistory(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	// 构建历史对话
	history := []types.ChatMessage{
		types.NewUserMessage("My name is John."),
		types.NewAssistantMessage("Hello John! Nice to meet you."),
	}

	response, err := c.ChatWithHistory(ctx,
		"What is my name?",
		history,
		chat.WithModel(defaultModel),
		chat.WithMaxTokens(50),
		chat.WithTemperature(0.5),
	)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Choices)
	assert.Contains(t, response.Choices[0].Message.GetTextContent(), "John")

	t.Logf("Response: %s", response.Choices[0].Message.GetTextContent())
}

// TestRealAPIStreamChat 测试真实API的流式聊天
func TestRealAPIStreamChat(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	stream, err := c.SimpleChatStream(ctx,
		"Count from 1 to 5, each number on a new line.",
		chat.WithModel(defaultModel),
		chat.WithMaxTokens(100),
		chat.WithTemperature(0.5),
	)

	assert.NoError(t, err)
	assert.NotNil(t, stream)
	defer stream.Close()

	// 读取流式数据
	eventCount := 0
	var fullContent string

	for {
		event, err := stream.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			assert.NoError(t, err)
			break
		}

		if event.Type == types.StreamEventTypeData {
			eventCount++
			t.Logf("Event %d: %s", eventCount, string(event.Data))
		}
	}

	assert.Greater(t, eventCount, 0)
	t.Logf("Total events: %d", eventCount)
	t.Logf("Full content: %s", fullContent)
}

// TestRealAPIMultipleModels 测试真实API的多种模型
func TestRealAPIMultipleModels(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	models := []string{defaultModel}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			response, err := c.SimpleChat(ctx,
				"Say hello",
				chat.WithModel(model),
				chat.WithMaxTokens(20),
				chat.WithTemperature(0.5),
			)

			if err != nil {
				t.Logf("Model %s failed: %v", model, err)
				return
			}

			assert.NotNil(t, response)
			assert.NotEmpty(t, response.Choices)
			t.Logf("Model %s response: %s", model, response.Choices[0].Message.GetTextContent())
		})
	}
}

// TestRealAPITokenUsage 测试真实API的Token使用情况
func TestRealAPITokenUsage(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	response, err := c.SimpleChat(ctx,
		"Explain what is machine learning in one sentence.",
		chat.WithModel(defaultModel),
		chat.WithMaxTokens(100),
		chat.WithTemperature(0.7),
	)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Greater(t, response.Usage.TotalTokens, 0)
	assert.Greater(t, response.Usage.PromptTokens, 0)
	assert.Greater(t, response.Usage.CompletionTokens, 0)

	t.Logf("Token usage - Prompt: %d, Completion: %d, Total: %d",
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
		response.Usage.TotalTokens)
}

// TestRealAPIErrorHandling 测试真实API的错误处理
func TestRealAPIErrorHandling(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	// 测试无效模型
	response, err := c.SimpleChat(ctx,
		"Hello",
		chat.WithModel("invalid-model-name"),
		chat.WithMaxTokens(50),
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	t.Logf("Expected error for invalid model: %v", err)
}

// TestRealAPIConfigValidation 测试真实API的配置验证
func TestRealAPIConfigValidation(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx := context.Background()

	// 测试无效的温度参数
	response, err := c.SimpleChat(ctx,
		"Hello",
		chat.WithModel(defaultModel),
		chat.WithTemperature(3.0), // 无效的温度值
		chat.WithMaxTokens(50),
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "temperature must be between 0.0 and 2.0")
	t.Logf("Expected error for invalid temperature: %v", err)
}

// TestRealAPIContextCancellation 测试真实API的上下文取消
func TestRealAPIContextCancellation(t *testing.T) {
	c := setupRealAPIClient(t)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	response, err := c.SimpleChat(ctx,
		"Tell me a very long story about artificial intelligence.",
		chat.WithModel(defaultModel),
		chat.WithMaxTokens(1000),
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "context")
	t.Logf("Expected context cancellation error: %v", err)
}

// BenchmarkRealAPISimpleChat 性能测试
func BenchmarkRealAPISimpleChat(b *testing.B) {
	baseURL := os.Getenv("NEW_API")
	apiKey := os.Getenv("NEW_API_KEY")

	if baseURL == "" || apiKey == "" {
		b.Skip("Skipping benchmark: NEW_API or NEW_API_KEY not set")
	}

	cfg := &config.Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
		Debug:   false,
	}

	c, err := client.NewClient(client.WithConfig(cfg))
	require.NoError(b, err)
	defer c.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.SimpleChat(ctx,
			"Hello",
			chat.WithModel(defaultModel),
			chat.WithMaxTokens(10),
			chat.WithTemperature(0.5),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}
