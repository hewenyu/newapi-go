# Claude API 本地代理服务器设计文档

## 项目概述

本项目旨在创建一个本地代理服务器，将Claude API格式的请求转换为NewAPI-Go SDK调用，实现Claude Code的本地转发功能。

## 核心目标

1. **API兼容性**: 完全兼容Claude API格式，无需修改现有客户端代码
2. **环境配置**: 支持NEW_API和NEW_API_KEY环境变量配置
3. **流式支持**: 支持流式聊天响应，实现实时对话
4. **错误处理**: 完善的错误处理和日志记录机制
5. **性能优化**: 高效的请求处理和内存管理

## 架构设计

### 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Claude Code   │───▶│  Local Proxy    │───▶│   NEW API       │
│   (Client)      │    │   Server        │    │   Service       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 组件架构

```
proxy/
├── main.go                    # 主程序入口
├── server/                    # HTTP服务器组件
│   ├── server.go             # 服务器实现
│   ├── handlers.go           # 请求处理器
│   ├── middleware.go         # 中间件
│   └── routes.go             # 路由定义
├── converter/                 # 格式转换器
│   ├── claude_to_newapi.go   # Claude -> NewAPI转换
│   ├── newapi_to_claude.go   # NewAPI -> Claude转换
│   └── stream_converter.go   # 流式转换
├── config/                   # 配置管理
│   └── config.go             # 环境变量配置
├── types/                    # 类型定义
│   └── claude.go             # Claude API类型
└── utils/                    # 工具函数
    ├── logger.go             # 日志工具
    └── validator.go          # 验证工具
```

## API映射设计

### Claude API 端点映射

| Claude API | 本地代理 | NewAPI-Go方法 |
|-----------|----------|---------------|
| POST /v1/messages | POST /v1/messages | client.CreateChatCompletion |
| POST /v1/messages (stream) | POST /v1/messages (stream) | client.CreateChatCompletionStream |

### 请求格式转换

#### Claude API请求格式
```json
{
  "model": "claude-3-sonnet-20240229",
  "max_tokens": 1000,
  "messages": [
    {
      "role": "user",
      "content": "Hello, world!"
    }
  ],
  "stream": false
}
```

#### NewAPI-Go SDK调用
```go
response, err := client.CreateChatCompletion(ctx, []types.ChatMessage{
    types.NewUserMessage("Hello, world!"),
}, chat.WithModel("claude-3-sonnet-20240229"), chat.WithMaxTokens(1000))
```

### 响应格式转换

#### NewAPI-Go SDK响应
```go
type ChatCompletionResponse struct {
    ID      string                 `json:"id"`
    Object  string                 `json:"object"`
    Created int64                  `json:"created"`
    Model   string                 `json:"model"`
    Choices []ChatCompletionChoice `json:"choices"`
    Usage   Usage                  `json:"usage"`
}
```

#### Claude API响应格式
```json
{
  "id": "msg_01XFDUDYJgAACzvnptvVoYEL",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "Hello! How can I help you today?"
    }
  ],
  "model": "claude-3-sonnet-20240229",
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 10,
    "output_tokens": 25
  }
}
```

## 核心功能实现

### 1. 配置管理

```go
type Config struct {
    NewAPIURL    string // 从NEW_API环境变量获取
    NewAPIKey    string // 从NEW_API_KEY环境变量获取
    ServerPort   int    // 代理服务器端口，默认8080
    LogLevel     string // 日志级别，默认INFO
    EnableDebug  bool   // 调试模式
}
```

### 2. HTTP服务器

- **监听端口**: 默认8080，可通过环境变量配置
- **路由处理**: 支持Claude API标准路由
- **中间件**: 日志记录、CORS支持、错误处理
- **健康检查**: 提供服务健康状态检查端点

### 3. 格式转换器

#### Claude -> NewAPI转换器
- 消息格式转换
- 参数映射和验证
- 模型名称映射

#### NewAPI -> Claude转换器
- 响应格式转换
- 错误格式标准化
- 使用量信息映射

#### 流式转换器
- SSE格式处理
- 实时数据转换
- 连接管理

### 4. 请求处理器

#### 消息处理器
```go
func (h *MessageHandler) HandleMessage(w http.ResponseWriter, r *http.Request) {
    // 1. 解析Claude API请求
    // 2. 转换为NewAPI-Go调用
    // 3. 执行SDK调用
    // 4. 转换响应格式
    // 5. 返回Claude API响应
}
```

#### 流式处理器
```go
func (h *MessageHandler) HandleStreamMessage(w http.ResponseWriter, r *http.Request) {
    // 1. 设置SSE响应头
    // 2. 转换请求格式
    // 3. 创建流式调用
    // 4. 实时转换和发送数据
    // 5. 处理连接关闭
}
```

## 错误处理机制

### 错误类型映射

| NewAPI-Go错误 | Claude API错误 | HTTP状态码 |
|---------------|----------------|-----------|
| 认证错误 | authentication_error | 401 |
| 参数错误 | invalid_request_error | 400 |
| 限流错误 | rate_limit_error | 429 |
| 服务错误 | api_error | 500 |

### 错误响应格式

```json
{
  "type": "error",
  "error": {
    "type": "invalid_request_error",
    "message": "Invalid request parameters"
  }
}
```

## 流式处理实现

### SSE格式

```
event: message_start
data: {"type": "message_start", "message": {...}}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {...}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {...}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: message_stop
data: {"type": "message_stop"}
```

### 流式转换逻辑

1. 接收NewAPI-Go SDK流式响应
2. 解析每个数据块
3. 转换为Claude API SSE格式
4. 实时发送给客户端
5. 处理连接断开和错误

## 性能优化

### 连接池管理
- HTTP客户端连接池
- NewAPI-Go SDK客户端复用
- 连接超时和重试机制

### 内存管理
- 流式数据缓冲控制
- 请求响应数据及时释放
- 垃圾回收优化

### 并发处理
- Goroutine池管理
- 请求限流和排队
- 资源竞争避免

## 配置示例

### 环境变量配置

```bash
# 必需配置
export NEW_API="https://api.example.com/v1"
export NEW_API_KEY="your-api-key-here"

# 可选配置
export PROXY_PORT=8080
export PROXY_LOG_LEVEL=INFO
export PROXY_DEBUG=false
```

### 启动命令

```bash
# 开发环境
go run proxy/main.go

# 生产环境
./claude-proxy

# Docker运行
docker run -p 8080:8080 \
  -e NEW_API="https://api.example.com/v1" \
  -e NEW_API_KEY="your-key" \
  claude-proxy
```

## 使用示例

### 客户端调用

```bash
# 普通聊天
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-claude-key" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1000,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 流式聊天
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-claude-key" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1000,
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

### 集成到Claude Code

将Claude Code的API端点配置为本地代理服务器地址：

```
API_BASE_URL=http://localhost:8080
```

## 测试策略

### 单元测试
- 格式转换器测试
- 配置管理测试
- 工具函数测试

### 集成测试
- 端到端API测试
- 流式处理测试
- 错误处理测试

### 性能测试
- 并发请求测试
- 内存使用测试
- 响应时间测试

## 部署建议

### 开发环境
- 直接运行Go程序
- 使用环境变量配置
- 启用调试模式

### 生产环境
- 编译为可执行文件
- 使用进程管理器(如systemd)
- 配置日志轮转
- 监控和告警

### 容器化部署
- 提供Dockerfile
- 支持Docker Compose
- Kubernetes部署配置

## 安全考虑

### API密钥管理
- 环境变量存储
- 密钥轮转支持
- 访问权限控制

### 网络安全
- HTTPS支持
- 请求验证
- 防护DDoS攻击

### 数据安全
- 请求日志脱敏
- 内存数据清理
- 传输加密

## 监控和日志

### 日志记录
- 结构化日志格式
- 请求响应记录
- 错误和性能日志

### 监控指标
- 请求量和响应时间
- 错误率和成功率
- 系统资源使用

### 健康检查
- 服务状态检查
- 依赖服务检查
- 自动恢复机制

## 后续扩展

### 功能扩展
- 支持更多Claude API端点
- 添加缓存机制
- 实现负载均衡

### 性能优化
- 请求合并和批处理
- 智能路由选择
- 预测性缓存

### 运维增强
- 自动扩缩容
- 故障自愈
- 性能调优 