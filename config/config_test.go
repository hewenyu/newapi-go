package config

import (
	"testing"
	"time"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				APIKey:     "test-key",
				BaseURL:    "https://api.example.com",
				Timeout:    30 * time.Second,
				HTTPClient: DefaultHTTPClient(),
				UserAgent:  "test-agent",
				Debug:      false,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: &Config{
				BaseURL:    "https://api.example.com",
				Timeout:    30 * time.Second,
				HTTPClient: DefaultHTTPClient(),
				UserAgent:  "test-agent",
				Debug:      false,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				APIKey:     "test-key",
				BaseURL:    "https://api.example.com",
				Timeout:    -1 * time.Second,
				HTTPClient: DefaultHTTPClient(),
				UserAgent:  "test-agent",
				Debug:      false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()
	config, err := builder.
		WithAPIKey("test-key").
		WithBaseURL("https://api.example.com").
		WithTimeout(60 * time.Second).
		WithDebug(true).
		Build()

	if err != nil {
		t.Errorf("ConfigBuilder.Build() error = %v", err)
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected APIKey = 'test-key', got %s", config.APIKey)
	}

	if config.BaseURL != "https://api.example.com" {
		t.Errorf("Expected BaseURL = 'https://api.example.com', got %s", config.BaseURL)
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout = 60s, got %v", config.Timeout)
	}

	if !config.Debug {
		t.Errorf("Expected Debug = true, got %v", config.Debug)
	}
}

func TestConfigClone(t *testing.T) {
	original := DefaultConfig()
	original.APIKey = "original-key"
	original.Debug = true

	clone := original.Clone()

	if clone.APIKey != original.APIKey {
		t.Errorf("Clone APIKey mismatch: expected %s, got %s", original.APIKey, clone.APIKey)
	}

	// 修改原始配置，确保克隆不受影响
	original.APIKey = "modified-key"
	if clone.APIKey == original.APIKey {
		t.Errorf("Clone should not be affected by original config changes")
	}
}
