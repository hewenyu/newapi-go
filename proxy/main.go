package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hewenyu/newapi-go/proxy/config"
	"github.com/hewenyu/newapi-go/proxy/server"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// 创建服务器
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 运行服务器
	if err := srv.Run(); err != nil {
		log.Fatalf("Server run failed: %v", err)
	}
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("Claude API Local Proxy Server")
	fmt.Println()
	fmt.Println("Required environment variables:")
	fmt.Println("  NEW_API       - NewAPI service URL")
	fmt.Println("  NEW_API_KEY   - NewAPI service API key")
	fmt.Println()
	fmt.Println("Optional environment variables:")
	fmt.Println("  PROXY_PORT    - Server port (default: 8080)")
	fmt.Println("  PROXY_HOST    - Server host (default: 0.0.0.0)")
	fmt.Println("  PROXY_DEBUG   - Enable debug mode (default: false)")
	fmt.Println("  PROXY_TIMEOUT - Request timeout (default: 30s)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run proxy/main.go")
	fmt.Println("  ./claude-proxy")
	fmt.Println()
	fmt.Println("API Endpoints:")
	fmt.Println("  POST /v1/messages  - Claude API messages endpoint")
	fmt.Println("  GET  /health       - Health check")
	fmt.Println("  GET  /info         - Service information")
	fmt.Println()
	fmt.Println("Example client usage:")
	fmt.Println("  curl -X POST http://localhost:8080/v1/messages \\")
	fmt.Println("    -H \"Content-Type: application/json\" \\")
	fmt.Println("    -d '{")
	fmt.Println("      \"model\": \"claude-3-sonnet-20240229\",")
	fmt.Println("      \"max_tokens\": 1000,")
	fmt.Println("      \"messages\": [")
	fmt.Println("        {\"role\": \"user\", \"content\": \"Hello, world!\"}")
	fmt.Println("      ]")
	fmt.Println("    }'")
}

// init 初始化函数
func init() {
	// 检查是否需要显示帮助
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printUsage()
		os.Exit(0)
	}
}
