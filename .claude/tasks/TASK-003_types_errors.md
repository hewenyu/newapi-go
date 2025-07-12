# [TASK-003] 实现类型定义和错误处理

- **状态**: Blocked
- **前置依赖**: TASK-001

## 1. 任务目标
定义SDK中所有API相关的数据结构和错误处理机制，确保类型安全和错误处理的一致性。

## 2. 上下文与价值
这是SDK的基础数据层，所有服务模块都将使用这些类型定义。完成后，SDK将有统一的数据结构和错误处理标准。

## 3. 输入 (Inputs)
- 文件: `docs/auto/sdk_requirements.md`（需求文档）
- 文件: `docs/coding_standards.md`（编码规范）
- 目录: `types/`（TASK-001创建的目录结构）

## 4. 输出与交付物 (Outputs & Deliverables)
- **创建**: 
  - `types/common.go` - 通用类型定义
  - `types/chat.go` - 聊天相关类型
  - `types/embeddings.go` - 嵌入相关类型
  - `types/image.go` - 图像相关类型
  - `types/audio.go` - 音频相关类型
  - `types/errors.go` - 错误类型定义
  - `types/stream.go` - 流式处理类型
- **修改**: 无
- **删除**: 无

## 5. 验收标准 (Definition of Done)
1. 所有API响应结构支持字段忽略（未定义字段不影响解析）
2. 错误类型包含完整的错误信息和上下文
3. 流式处理类型支持实时数据传输
4. 所有结构体都有适当的JSON标签
5. 所有导出类型都有完整的文档注释
6. 代码通过`go vet`和`gofmt`检查
7. 单元测试覆盖类型转换和错误处理

## 6. 实现概要 (Implementation Plan)
1. 实现`types/common.go`：
   - 定义通用的响应结构体
   - 定义分页相关结构体
   - 定义元数据结构体

2. 实现`types/errors.go`：
   - 定义自定义错误类型
   - 实现错误包装和展开
   - 定义错误码常量

3. 实现`types/chat.go`：
   - 定义聊天请求/响应结构体
   - 定义消息类型和角色
   - 定义聊天选项和参数

4. 实现`types/embeddings.go`：
   - 定义嵌入请求/响应结构体
   - 定义嵌入向量类型

5. 实现`types/image.go`：
   - 定义图像生成请求/响应结构体
   - 定义图像参数和选项

6. 实现`types/audio.go`：
   - 定义音频处理请求/响应结构体
   - 定义音频格式和参数

7. 实现`types/stream.go`：
   - 定义流式响应接口
   - 实现流式数据处理机制

## 7. 注意事项与潜在风险
- 注意：JSON标签必须支持API兼容性要求
- 注意：错误信息必须包含足够的上下文信息
- 风险：类型定义不当可能影响API兼容性
- 风险：流式处理的并发安全性需要特别关注 