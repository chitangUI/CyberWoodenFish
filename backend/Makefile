# 电子木鱼后端项目 Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop deps dev-setup

# 默认目标
help: ## 显示帮助信息
	@echo "电子木鱼后端项目 - 可用命令："
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 环境变量
GO_VERSION := 1.21
APP_NAME := muyu-backend
DOCKER_IMAGE := $(APP_NAME):latest

# 构建相关
build: ## 构建应用
	@echo "构建应用..."
	go build -o bin/$(APP_NAME) .

build-linux: ## 构建Linux版本
	@echo "构建Linux版本..."
	GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-linux .

run: ## 运行应用
	@echo "启动应用..."
	go run .

test: ## 运行测试
	@echo "运行测试..."
	go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试覆盖率..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成完成: coverage.html"

# 依赖管理
deps: ## 下载依赖
	@echo "下载依赖..."
	go mod download
	go mod tidy

deps-update: ## 更新依赖
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy

# 开发环境
dev-setup: ## 设置开发环境
	@echo "设置开发环境..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

dev: ## 以开发模式运行(热重载)
	@echo "启动开发模式..."
	air

lint: ## 运行代码检查
	@echo "运行代码检查..."
	golangci-lint run

format: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	goimports -w .

# Docker相关
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## 运行Docker容器
	@echo "启动Docker容器..."
	docker-compose up -d

docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	docker-compose down

docker-logs: ## 查看Docker日志
	@echo "查看Docker日志..."
	docker-compose logs -f muyu-backend

docker-clean: ## 清理Docker资源
	@echo "清理Docker资源..."
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

# 数据库相关
db-migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	go run scripts/migrate.go

db-seed: ## 填充测试数据
	@echo "填充测试数据..."
	go run scripts/seed.go

db-reset: ## 重置数据库
	@echo "重置数据库..."
	docker-compose exec postgres psql -U muyu_user -d muyu_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# 部署相关
deploy-staging: ## 部署到测试环境
	@echo "部署到测试环境..."
	docker-compose -f docker-compose.staging.yml up -d

deploy-prod: ## 部署到生产环境
	@echo "部署到生产环境..."
	@echo "请确认生产环境配置..."
	@read -p "确认部署到生产环境? [y/N] " confirm && [ "$$confirm" = "y" ]
	docker-compose -f docker-compose.prod.yml up -d

# 监控和维护
logs: ## 查看应用日志
	@echo "查看应用日志..."
	docker-compose logs -f muyu-backend

monitor: ## 启动监控
	@echo "启动监控..."
	docker-compose -f docker-compose.monitoring.yml up -d

backup-db: ## 备份数据库
	@echo "备份数据库..."
	docker-compose exec postgres pg_dump -U muyu_user muyu_db > backups/db_backup_$(shell date +%Y%m%d_%H%M%S).sql

restore-db: ## 恢复数据库 (需要指定文件: make restore-db FILE=backup.sql)
	@echo "恢复数据库..."
	@if [ -z "$(FILE)" ]; then echo "请指定备份文件: make restore-db FILE=backup.sql"; exit 1; fi
	docker-compose exec -T postgres psql -U muyu_user -d muyu_db < $(FILE)

# 清理
clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

clean-all: clean docker-clean ## 清理所有文件和Docker资源
	@echo "清理完成"

# 生产环境健康检查
health-check: ## 检查服务健康状态
	@echo "检查服务健康状态..."
	@curl -f http://localhost:8080/health || echo "服务未运行"

# 工具
generate-docs: ## 生成API文档
	@echo "生成API文档..."
	@echo "API文档功能待实现..."

security-scan: ## 运行安全扫描
	@echo "运行安全扫描..."
	gosec ./...

# 性能测试
benchmark: ## 运行性能测试
	@echo "运行性能测试..."
	go test -bench=. -benchmem ./...

load-test: ## 运行负载测试
	@echo "运行负载测试..."
	@echo "负载测试功能待实现..."
