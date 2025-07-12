# [TASK-006] 实现Embeddings服务

- **状态**: Blocked
- **前置依赖**: TASK-004

## 1. 任务目标
实现Embeddings服务模块，支持文本嵌入向量生成功能。

## 2. 上下文与价值
Embeddings是SDK的重要功能，用于生成文本的向量表示。完成后，用户可以通过SDK获取文本的嵌入向量。

## 3. 输入 (Inputs)
- 文件: `client/client.go`（TASK-004更新的核心客户端）
- 文件: `types/embeddings.go`（TASK-003创建的嵌入类型）
- 文件: `internal/transport/http.go`（TASK-004创建的HTTP传输层）
- 目录: `services/embeddings/`（TASK-001创建的目录结构）

## 4. 输出与交付物 (Outputs & Deliverables)
- **创建**: 
  - `services/embeddings/embeddings.go` - Embeddings服务实现
  - `services/embeddings/options.go` - Embeddings选项配置
  - `services/embeddings/embeddings_test.go` - 单元测试
- **修改**: `client/client.go`（添加Embeddings服务方法）
- **删除**: 无

## 5. 验收标准 (Definition of Done)
1. 支持单个文本嵌入生成
2. 支持批量文本嵌入生成
3. 支持不同的嵌入模型选择
4. 支持嵌入维度配置
5. 返回标准化的嵌入向量
6. 代码通过`go vet`和`gofmt`检查
7. 单元测试覆盖所有主要功能

## 6. 实现概要 (Implementation Plan)
1. 实现`services/embeddings/options.go`：
   - 定义EmbeddingOption函数类型
   - 实现各种嵌入选项（WithModel、WithDimensions等）
   - 定义嵌入参数结构

2. 实现`services/embeddings/embeddings.go`：
   - 定义EmbeddingService结构体
   - 实现单个文本嵌入方法
   - 实现批量文本嵌入方法
   - 实现嵌入请求构建和响应解析

3. 实现`services/embeddings/embeddings_test.go`：
   - 测试单个文本嵌入功能
   - 测试批量文本嵌入功能
   - 测试错误处理
   - 测试参数配置

4. 更新`client/client.go`：
   - 添加Embeddings服务实例
   - 实现Embeddings方法的代理

## 7. 注意事项与潜在风险
- 注意：嵌入向量可能很大，需要考虑内存使用
- 注意：批量处理需要合理的分片大小
- 风险：模型不支持的文本长度可能导致请求失败
- 风险：向量维度不匹配可能导致后续处理问题 