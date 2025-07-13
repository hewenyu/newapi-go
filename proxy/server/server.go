package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hewenyu/newapi-go/client"
	"github.com/hewenyu/newapi-go/proxy/config"
)

// Server 代理服务器
type Server struct {
	config         *config.Config
	httpServer     *http.Server
	newAPIClient   *client.Client
	messageHandler *MessageHandler
	healthHandler  *HealthHandler
	infoHandler    *InfoHandler
	mu             sync.RWMutex
	running        bool
}

// NewServer 创建新的代理服务器
func NewServer(cfg *config.Config) (*Server, error) {
	// 创建NewAPI客户端
	newAPIClient, err := client.NewClient(
		client.WithAPIKey(cfg.NewAPIKey),
		client.WithBaseURL(cfg.NewAPIURL),
		client.WithTimeout(cfg.RequestTimeout),
		client.WithDebug(cfg.IsDebugEnabled()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create NewAPI client: %w", err)
	}

	// 创建处理器
	messageHandler := NewMessageHandler(cfg, newAPIClient)
	healthHandler := NewHealthHandler(cfg)
	infoHandler := NewInfoHandler(cfg)

	// 创建服务器
	server := &Server{
		config:         cfg,
		newAPIClient:   newAPIClient,
		messageHandler: messageHandler,
		healthHandler:  healthHandler,
		infoHandler:    infoHandler,
	}

	// 创建HTTP服务器
	mux := server.setupRoutes()
	handler := server.withMiddleware(mux)
	server.httpServer = &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      handler,
		ReadTimeout:  cfg.RequestTimeout,
		WriteTimeout: cfg.RequestTimeout,
		IdleTimeout:  cfg.RequestTimeout * 2,
	}

	return server, nil
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Claude API路由
	mux.HandleFunc("/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			MethodNotAllowedHandler(w, r)
			return
		}
		s.messageHandler.HandleMessage(w, r)
	})

	// 健康检查路由
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			MethodNotAllowedHandler(w, r)
			return
		}
		s.healthHandler.HandleHealth(w, r)
	})

	// 信息路由
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			MethodNotAllowedHandler(w, r)
			return
		}
		s.infoHandler.HandleInfo(w, r)
	})

	// 默认路由
	mux.HandleFunc("/", NotFoundHandler)

	return mux
}

// withMiddleware 添加中间件
func (s *Server) withMiddleware(handler http.Handler) http.Handler {
	// 应用中间件链
	wrapped := handler

	// 添加CORS中间件
	if s.config.EnableCORS {
		wrapped = s.corsMiddleware(wrapped)
	}

	// 添加日志中间件
	wrapped = s.loggingMiddleware(wrapped)

	// 添加恢复中间件
	wrapped = s.recoveryMiddleware(wrapped)

	// 添加速率限制中间件
	wrapped = s.rateLimitMiddleware(wrapped)

	return wrapped
}

// corsMiddleware CORS中间件
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头部
		for _, origin := range s.config.CORSAllowOrigins {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			break // 只设置第一个，实际应用中可能需要更复杂的逻辑
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, anthropic-version")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 日志中间件
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应写入器包装器
		ww := &responseWriter{ResponseWriter: w}

		// 处理请求
		next.ServeHTTP(ww, r)

		// 记录日志
		duration := time.Since(start)
		if s.config.IsDebugEnabled() {
			log.Printf("[%s] %s %s - %d - %v",
				r.Method, r.URL.Path, r.RemoteAddr,
				ww.statusCode, duration)
		}
	})
}

// recoveryMiddleware 恢复中间件
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)

				// 发送错误响应
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware 速率限制中间件
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	// 这是一个简单的速率限制实现
	// 实际生产环境可能需要更复杂的实现
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 这里可以实现基于IP的速率限制
		// 目前暂时跳过
		next.ServeHTTP(w, r)
	})
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	// 打印配置信息
	s.config.Print()

	log.Printf("Starting Claude API Proxy server on %s", s.config.GetServerAddress())

	// 启动HTTP服务器
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Printf("Server started successfully")
	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is not running")
	}
	s.running = false
	s.mu.Unlock()

	log.Printf("Stopping server...")

	// 关闭HTTP服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		return err
	}

	// 关闭NewAPI客户端
	if s.newAPIClient != nil {
		if err := s.newAPIClient.Close(); err != nil {
			log.Printf("NewAPI client close error: %v", err)
		}
	}

	log.Printf("Server stopped")
	return nil
}

// IsRunning 检查服务器是否运行
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetAddress 获取服务器地址
func (s *Server) GetAddress() string {
	return s.config.GetServerAddress()
}

// Run 运行服务器（带信号处理）
func (s *Server) Run() error {
	// 启动服务器
	if err := s.Start(); err != nil {
		return err
	}

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-sigChan
	log.Printf("Received signal: %s", sig)

	// 创建关闭上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止服务器
	return s.Stop(ctx)
}

// WaitForReady 等待服务器准备就绪
func (s *Server) WaitForReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server to be ready")
		case <-ticker.C:
			if s.IsRunning() {
				// 尝试访问健康检查端点
				resp, err := http.Get(fmt.Sprintf("http://%s/health", s.GetAddress()))
				if err == nil && resp.StatusCode == http.StatusOK {
					resp.Body.Close()
					return nil
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
		}
	}
}
