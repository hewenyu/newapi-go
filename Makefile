# New-API Go SDK Makefile

.PHONY: test lint build clean help

# 默认目标
help:
	@echo "Available commands:"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linting"
	@echo "  build    - Build the project"
	@echo "  clean    - Clean build artifacts"
	@echo "  help     - Show this help message"

# 运行测试
test:
	go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码检查
lint:
	go vet ./...
	go fmt ./...
	golint ./...

# 构建项目
build:
	go build -v ./...

# 清理构建文件
clean:
	go clean
	rm -f coverage.out coverage.html

# 整理依赖
tidy:
	go mod tidy

# 运行所有检查
check: lint test

# 安装开发依赖
install-deps:
	go install golang.org/x/lint/golint@latest 