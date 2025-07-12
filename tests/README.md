## chat 测试

```bash
# 基本测试
go test -v ./tests -run TestRealAPISimpleChat
# 系统消息测试
go test -v ./tests -run TestRealAPIChatWithSystem
# 历史对话测试
go test -v ./tests -run TestRealAPIChatWithHistory
# 流式聊天测试
go test -v ./tests -run TestRealAPIStreamChat   
# 多种模型测试
go test -v ./tests -run TestRealAPIMultipleModels
# Token使用情况测试
go test -v ./tests -run TestRealAPITokenUsage
# 错误处理测试
go test -v ./tests -run TestRealAPIErrorHandling
# 上下文取消测试
go test -v ./tests -run TestRealAPIContextCancellation
# 配置验证测试
go test -v ./tests -run TestRealAPIConfigValidation
```


## embeddings 测试

```bash
# 单个文本嵌入测试
go test -v ./tests -run TestEmbeddingRealAPICreateEmbedding
# 批量文本嵌入测试
go test -v ./tests -run TestEmbeddingRealAPICreateEmbeddings
# 多种模型测试
go test -v ./tests -run TestEmbeddingRealAPIMultipleModels
# 嵌入选项测试
go test -v ./tests -run TestEmbeddingRealAPIEmbeddingOptions
# Token使用情况测试
go test -v ./tests -run TestEmbeddingRealAPITokenUsage
# 错误处理测试
go test -v ./tests -run TestEmbeddingRealAPIErrorHandling
# 输入验证测试
go test -v ./tests -run TestEmbeddingRealAPIInputValidation
# 上下文取消测试
go test -v ./tests -run TestEmbeddingRealAPIContextCancellation
```


## 音频测试

```bash
# 文本转语音测试
go test -v ./tests -run TestRealAPIAudioTranscription
```

