# New-API Go SDK 编码规范

## 1. 包命名规范
- 使用小写，无下划线、短横线
- 名称简短有意义，体现功能
- 示例：`client`, `config`, `chat`, `embeddings`

## 2. 变量/函数命名规范
- 使用驼峰式（CamelCase）
- 首字母大写为导出（public），小写为包内（private）
- 命名清晰表达用途，避免无意义缩写
- 示例：`NewClient()`, `chatService`, `APIKey`

## 3. 结构体与接口规范
- 优先定义具体类型（struct），只有当有多个实现时再提取 interface
- interface 命名一般以 -er 结尾，如 `ChatStreamer`, `AudioProcessor`
- 避免为接口而接口，按需定义

## 4. 代码格式规范
- 使用 `gofmt` 自动格式化
- 每行不超过 120 字符
- 语句块必须使用大括号，且左大括号不换行
- 单个文件不超过 300 行，超过时拆分为更小的文件
- 单个函数不超过 40 行，逻辑复杂应拆分成多个小函数

## 5. 错误处理规范
- 明确处理 error，不忽略
- 错误信息要具体，避免仅返回 "error occurred"
- 使用 `fmt.Errorf()` 包装错误并添加上下文信息

## 6. 注释规范
- 导出函数、类型需有注释，说明用途
- 包注释在 `package` 声明前，说明包的功能
- 复杂逻辑添加行内注释

## 7. 导入包规范
- 按标准库、第三方库、本地包分组，组间空行
- 示例：
```go
import (
    "context"
    "fmt"
    "net/http"

    "go.uber.org/zap"

    "github.com/hewenyu/newapi-go/types"
)
```

## 8. 空值判断规范
- 明确判断 nil，避免空指针
- 在函数开始处进行参数校验

## 9. 并发安全规范
- 客户端必须保证并发安全
- 使用 `sync.Mutex` 或 `sync.RWMutex` 保护共享资源
- HTTP 客户端使用 `http.Client` 的内置并发安全特性

## 10. Context 传递规范
- 所有网络请求方法必须接受 `context.Context` 作为第一个参数
- 不要在结构体中存储 context，而是通过参数传递

## 11. 测试规范
- 每个包都要有对应的测试文件
- 测试函数命名：`TestXxx`，基准测试：`BenchmarkXxx`
- 使用 `testify` 库进行断言

## 12. 文档规范
- 每个服务模块提供完整的使用示例
- 在 `examples/` 目录下提供可运行的示例代码
- README.md 包含快速开始指南 