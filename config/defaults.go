package config

import (
	"net/http"
	"time"
)

const (
	// DefaultBaseURL 是默认的API基础URL
	DefaultBaseURL = "https://api.newapi.com"
	// DefaultTimeout 是默认的HTTP请求超时时间
	DefaultTimeout = 30 * time.Second
	// DefaultUserAgent 是默认的User-Agent头
	DefaultUserAgent = "newapi-go-sdk/1.0.0"
)

// DefaultHTTPClient 创建并返回一个默认配置的HTTP客户端
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: DefaultTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

// DefaultConfig 创建并返回一个默认配置实例
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeout,
		HTTPClient: DefaultHTTPClient(),
		UserAgent:  DefaultUserAgent,
		Debug:      false,
	}
}
