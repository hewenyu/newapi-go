# Claude API 本地代理服务器

这是一个本地代理服务器，将Claude API格式的请求转换为NewAPI-Go SDK调用，实现Claude Code的本地转发功能。

## 功能特性

- ✅ 完全兼容Claude API格式
- ✅ 支持流式和非流式聊天响应
- ✅ 环境变量配置支持
- ✅ 自动格式转换（Claude API ↔ NewAPI-Go SDK）
- ✅ 完整的错误处理和日志记录
- ✅ CORS支持
- ✅ 健康检查和服务信息端点
- ✅ 优雅关闭和信号处理

## 快速开始

### 1. 环境变量配置

```bash
# 必需配置
export NEW_API="https://your-newapi-service.com/v1"
export NEW_API_KEY="your-api-key-here"

# 可选配置
export PROXY_PORT=8080
export PROXY_HOST=0.0.0.0
export PROXY_DEBUG=false
export PROXY_TIMEOUT=30s
```

### 2. 运行服务器

```bash
# 开发环境
go run proxy/main.go

# 编译运行
go build -o claude-proxy proxy/main.go
./claude-proxy
```

### 3. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 服务信息
curl http://localhost:8080/info
```

## API 使用示例

### 普通聊天请求

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1000,
    "messages": [
      {
        "role": "user",
        "content": "Hello, world!"
      }
    ]
  }'
```

### 流式聊天请求

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1000,
    "messages": [
      {
        "role": "user",
        "content": "Tell me a story"
      }
    ],
    "stream": true
  }'
```

### 带系统消息的请求

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1000,
    "system": "You are a helpful assistant.",
    "messages": [
      {
        "role": "user",
        "content": "What is the capital of France?"
      }
    ]
  }'
```

## 支持的模型

### 模型映射关系

本代理服务器将Claude API请求中的模型名称映射到NewAPI服务的实际模型：

#### 大模型（映射到 `gemini-2.5-pro`）
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-5-sonnet-20241022`
- `claude-3-opus`
- `claude-3-sonnet`
- `claude-3.5-sonnet`
- `opus`
- `sonnet`
- `large`
- `big`

#### 小模型（映射到 `gpt-4.1-mini`）
- `claude-3-haiku-20240307`
- `claude-3-5-haiku-20241022`
- `claude-3-haiku`
- `claude-3.5-haiku`
- `haiku`
- `small`
- `mini`

### 使用示例

```bash
# 使用大模型
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model": "sonnet", "max_tokens": 1000, "messages": [{"role": "user", "content": "Hello"}]}'

# 使用小模型
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model": "haiku", "max_tokens": 1000, "messages": [{"role": "user", "content": "Hello"}]}'
```

## 配置选项

| 环境变量 | 描述 | 默认值 | 必需 |
|---------|------|--------|------|
| `NEW_API` | NewAPI服务URL | - | ✅ |
| `NEW_API_KEY` | NewAPI服务API密钥 | - | ✅ |
| `PROXY_PORT` | 代理服务器端口 | 8080 | ❌ |
| `PROXY_HOST` | 代理服务器主机 | 0.0.0.0 | ❌ |
| `PROXY_DEBUG` | 启用调试模式 | false | ❌ |
| `PROXY_TIMEOUT` | 请求超时时间 | 30s | ❌ |
| `PROXY_MAX_REQUEST_SIZE` | 最大请求体大小 | 10MB | ❌ |
| `PROXY_MAX_CONCURRENT` | 最大并发数 | 100 | ❌ |
| `PROXY_ENABLE_CORS` | 启用CORS | true | ❌ |

## API 端点

### POST /v1/messages

Claude API兼容的聊天完成端点。

**请求格式**：
```json
{
  "model": "claude-3-sonnet-20240229",
  "max_tokens": 1000,
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "system": "You are a helpful assistant.",
  "temperature": 0.7,
  "stream": false
}
```

**响应格式**：
```json
{
  "id": "msg_123456",
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
  "usage": {
    "input_tokens": 10,
    "output_tokens": 25
  }
}
```

### GET /health

健康检查端点。

**响应格式**：
```json
{
  "status": "healthy",
  "timestamp": 1640995200,
  "version": "1.0.0",
  "uptime": "2h30m15s"
}
```

### GET /info

服务信息端点。

**响应格式**：
```json
{
  "service": "Claude API Proxy",
  "version": "1.0.0",
  "description": "Local proxy server for Claude API using NewAPI-Go SDK",
  "endpoints": [
    "POST /v1/messages",
    "GET /health",
    "GET /info"
  ],
  "supported_models": [
    "claude-3-opus-20240229",
    "claude-3-sonnet-20240229",
    "claude-3-haiku-20240307"
  ]
}
```

## 错误处理

服务器返回标准的Claude API错误格式：

```json
{
  "type": "error",
  "error": {
    "type": "invalid_request_error",
    "message": "Invalid request parameters"
  }
}
```

错误类型包括：
- `invalid_request_error` - 请求参数错误
- `authentication_error` - 认证错误
- `rate_limit_error` - 速率限制错误
- `api_error` - API服务错误

## 流式响应

启用流式模式时，服务器返回Server-Sent Events (SSE)格式的数据：

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

## 开发和调试

### 启用调试模式

```bash
export PROXY_DEBUG=true
go run proxy/main.go
```

调试模式会显示详细的请求日志和错误信息。

### 查看日志

服务器会输出请求日志：

```
2024/01/01 12:00:00 Starting Claude API Proxy server on 0.0.0.0:8080
2024/01/01 12:00:01 [POST] /v1/messages 192.168.1.100:12345 - 200 - 1.234s
```

### 性能监控

可以通过健康检查端点监控服务状态：

```bash
# 监控脚本
while true; do
  curl -s http://localhost:8080/health | jq '.status'
  sleep 5
done
```

## 部署建议

### 生产环境部署

1. 编译二进制文件：
```bash
go build -o claude-proxy proxy/main.go
```

2. 创建systemd服务：
```ini
[Unit]
Description=Claude API Proxy
After=network.target

[Service]
Type=simple
User=proxy
Environment=NEW_API=https://your-api.com/v1
Environment=NEW_API_KEY=your-key
Environment=PROXY_PORT=8080
ExecStart=/usr/local/bin/claude-proxy
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

3. 启动服务：
```bash
sudo systemctl enable claude-proxy
sudo systemctl start claude-proxy
```

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o claude-proxy proxy/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/claude-proxy .
CMD ["./claude-proxy"]
```

构建和运行：
```bash
docker build -t claude-proxy .
docker run -p 8080:8080 \
  -e NEW_API="https://your-api.com/v1" \
  -e NEW_API_KEY="your-key" \
  claude-proxy
```

## 故障排除

### 常见问题

1. **环境变量未设置**
   - 错误：`NEW_API environment variable is required`
   - 解决：确保设置了`NEW_API`和`NEW_API_KEY`环境变量

2. **端口被占用**
   - 错误：`bind: address already in use`
   - 解决：更改`PROXY_PORT`或停止占用端口的进程

3. **NewAPI连接失败**
   - 错误：`failed to create NewAPI client`
   - 解决：检查`NEW_API`URL和`NEW_API_KEY`是否正确

4. **请求超时**
   - 错误：`context deadline exceeded`
   - 解决：增加`PROXY_TIMEOUT`值

### 日志分析

启用调试模式查看详细日志：
```bash
export PROXY_DEBUG=true
./claude-proxy
```

## 许可证

本项目基于现有的newapi-go SDK构建，遵循相应的开源许可证。 