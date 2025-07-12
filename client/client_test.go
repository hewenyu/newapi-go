package client

import (
	"testing"
	"time"

	"github.com/hewenyu/newapi-go/config"
)

func TestNewClient(t *testing.T) {
	// 测试使用默认配置创建客户端
	client, err := NewClient()
	if err == nil {
		t.Errorf("Expected error for missing API key, got nil")
	}

	// 测试使用API密钥创建客户端
	client, err = NewClient(WithAPIKey("test-key"))
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}

	if client.GetAPIKey() != "test-key" {
		t.Errorf("Expected APIKey = 'test-key', got %s", client.GetAPIKey())
	}
}

func TestClientOptions(t *testing.T) {
	client, err := NewClient(
		WithAPIKey("test-key"),
		WithBaseURL("https://api.example.com"),
		WithTimeout(60*time.Second),
		WithDebug(true),
	)

	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}

	if client.GetAPIKey() != "test-key" {
		t.Errorf("Expected APIKey = 'test-key', got %s", client.GetAPIKey())
	}

	if client.GetBaseURL() != "https://api.example.com" {
		t.Errorf("Expected BaseURL = 'https://api.example.com', got %s", client.GetBaseURL())
	}

	if !client.IsDebugMode() {
		t.Errorf("Expected Debug = true, got %v", client.IsDebugMode())
	}

	cfg := client.GetConfig()
	if cfg.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout = 60s, got %v", cfg.Timeout)
	}
}

func TestClientWithConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Debug = true

	client, err := NewClient(WithConfig(cfg))
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}

	if client.GetAPIKey() != "test-key" {
		t.Errorf("Expected APIKey = 'test-key', got %s", client.GetAPIKey())
	}

	if !client.IsDebugMode() {
		t.Errorf("Expected Debug = true, got %v", client.IsDebugMode())
	}
}

func TestClientUpdateConfig(t *testing.T) {
	client, err := NewClient(WithAPIKey("test-key"))
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}

	newCfg := config.DefaultConfig()
	newCfg.APIKey = "new-key"
	newCfg.Debug = true

	err = client.UpdateConfig(newCfg)
	if err != nil {
		t.Errorf("UpdateConfig() error = %v", err)
	}

	if client.GetAPIKey() != "new-key" {
		t.Errorf("Expected APIKey = 'new-key', got %s", client.GetAPIKey())
	}

	if !client.IsDebugMode() {
		t.Errorf("Expected Debug = true, got %v", client.IsDebugMode())
	}
}

func TestClientConcurrency(t *testing.T) {
	client, err := NewClient(WithAPIKey("test-key"))
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}

	// 测试并发读取
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = client.GetAPIKey()
			_ = client.GetBaseURL()
			_ = client.IsDebugMode()
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 测试并发写入
	for i := 0; i < 5; i++ {
		go func() {
			newCfg := config.DefaultConfig()
			newCfg.APIKey = "concurrent-key"
			_ = client.UpdateConfig(newCfg)
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestClientString(t *testing.T) {
	client, err := NewClient(WithAPIKey("test-key"))
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}

	str := client.String()
	if str == "" {
		t.Errorf("Expected non-empty string representation")
	}
}
