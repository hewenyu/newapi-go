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


