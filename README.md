# New-API Go SDK

一个功能完整、高性能的 New-API Go SDK，为开发者提供简单易用的接口来访问 New-API 服务。

## 功能特性

- **聊天完成** - 支持流式和非流式聊天完成
- **文本嵌入** - 高效的文本向量化处理
- **图像生成** - 支持图像生成、编辑和变化
- **音频处理** - 语音转文本和文本转语音
- **类型安全** - 完整的类型定义和错误处理
- **高性能** - 优化的HTTP传输层和连接池
- **易于使用** - 直观的API设计和丰富的示例

## 快速开始

### 安装

```bash
go get github.com/hewenyu/newapi-go
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/hewenyu/newapi-go/client"
    "github.com/hewenyu/newapi-go/config"
)

func main() {
    // 创建配置
    cfg := config.New("your-api-key")
    
    // 创建客户端
    client := client.New(cfg)
    
    // 使用聊天功能
    response, err := client.Chat().Complete(context.Background(), &types.ChatRequest{
        Model: "gpt-3.5-turbo",
        Messages: []types.Message{
            {Role: "user", Content: "Hello, world!"},
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Choices[0].Message.Content)
}
```

## 文档

- [API文档](docs/api.md)
- [使用示例](examples/)
- [配置指南](docs/configuration.md)

## 许可证

MIT License
