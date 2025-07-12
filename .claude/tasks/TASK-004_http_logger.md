# [TASK-004] 实现HTTP传输层和日志系统

- **状态**: Blocked
- **前置依赖**: TASK-002,TASK-003

## 1. 任务目标
实现HTTP传输层和日志系统，为所有API调用提供统一的网络通信和日志记录功能。

## 2. 上下文与价值
这是SDK与New-API服务通信的基础设施，所有服务模块都将使用这个传输层。完成后，SDK将具备完整的HTTP请求处理和日志记录能力。

## 3. 输入 (Inputs)
- 文件: `client/client.go`（TASK-002创建的核心客户端）
- 文件: `config/config.go`（TASK-002创建的配置管理）
- 文件: `types/common.go`（TASK-003创建的通用类型）
- 文件: `types/errors.go`（TASK-003创建的错误类型）
- 目录: `internal/transport/`（TASK-001创建的目录结构）
- 目录: `internal/utils/`（TASK-001创建的目录结构）

## 4. 输出与交付物 (Outputs & Deliverables)
- **创建**: 
  - `internal/transport/http.go` - HTTP传输层实现
  - `internal/transport/request.go` - 请求构建器
  - `internal/transport/response.go` - 响应处理器
  - `internal/transport/stream.go` - 流式处理器
  - `internal/utils/logger.go` - 日志系统
  - `internal/utils/context.go` - 上下文工具
- **修改**: `client/client.go`（集成传输层和日志系统）
- **删除**: 无

## 5. 验收标准 (Definition of Done)
1. HTTP传输层支持所有基本的HTTP操作
2. 流式处理器支持Server-Sent Events (SSE)
3. 日志系统集成zap库，支持结构化日志
4. 所有HTTP请求都支持Context传递
5. 错误处理机制完整，包含重试逻辑
6. 请求/响应都有完整的日志记录
7. 代码通过`go vet`和`gofmt`检查
8. 单元测试覆盖核心功能

## 6. 实现概要 (Implementation Plan)
1. 实现`internal/utils/logger.go`：
   - 初始化zap日志器
   - 定义日志级别和格式
   - 实现结构化日志记录

2. 实现`internal/utils/context.go`：
   - 定义上下文键值常量
   - 实现上下文工具函数
   - 处理请求超时和取消

3. 实现`internal/transport/request.go`：
   - 定义请求构建器结构
   - 实现HTTP请求构建
   - 添加认证头和通用头

4. 实现`internal/transport/response.go`：
   - 定义响应处理器结构
   - 实现HTTP响应解析
   - 处理错误响应

5. 实现`internal/transport/stream.go`：
   - 实现SSE流式处理
   - 处理流式数据解析
   - 实现流式错误处理

6. 实现`internal/transport/http.go`：
   - 定义HTTP传输层接口
   - 实现具体的HTTP客户端
   - 集成请求、响应和流式处理

7. 更新`client/client.go`：
   - 集成HTTP传输层
   - 集成日志系统
   - 提供统一的API调用接口

## 7. 注意事项与潜在风险
- 注意：流式处理需要特别注意内存管理
- 注意：日志不能记录敏感信息（如API密钥）
- 风险：网络错误处理不当可能导致程序崩溃
- 风险：并发访问可能导致竞态条件 