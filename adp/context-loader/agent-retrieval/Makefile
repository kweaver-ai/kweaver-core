.DEFAULT_GOAL := help
.PHONY: help generate-mock lint test test-cover helm-template ktctl preview build

help:
	@echo "可用目标:"
	@echo "  generate-mock\t生成mock代码"
	@echo "  lint\t\t运行代码检查"
	@echo "  test\t\t执行单元测试"
	@echo "  test-cover\t执行单元测试并检查覆盖率"
	@echo "  helm-template\t生成Helm模板"
	@echo "  ktctl\t\t连接远程集群"
	@echo "  preview\t预览API文档"
	@echo "  build\t\t构建并运行主程序"

generate-mock:
	@echo "==== 生成mock代码 ===="
	go generate ./...

lint:
	@echo "==== 代码检查 ===="
	golangci-lint run ./... --exclude-dirs=server/tests

test:
	@echo "==== 执行测试 ===="
	go test $(shell go list ./... | grep -v /server/tests/ | grep -v /server/mocks) -gcflags=all=-l -v

test-cover:
	@echo "==== 执行测试并检查覆盖率 ===="
	go test $(shell go list ./... | grep -v /server/tests/ | grep -v /server/mocks) -gcflags=all=-l -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'

helm-template:
	@echo "==== 生成Helm模板 ===="
	@if ! command -v helm &>/dev/null; then \
		echo "请先安装helm"; exit 1; \
	fi
	helm template ./helm/agent-retrieval -n agent-retrieval -f./helm/agent-retrieval/values.yaml

ktctl:
	@echo "==== 连接远程集群 ===="
	.ktctl/dev.sh

preview:
	@echo "==== 预览API文档 ===="
	cd docs && ./preview.sh

build:
	@echo "==== 检查是否已经连接远程环境 ===="
	@# 获取将连接远程环境地址
	Host=$(shell grep '^Host="' .ktctl/dev.sh | awk -F\" '{print $2}')
	@if ! pgrep -f 'ktctl connect' &>/dev/null; then \
		echo "未连接到远程环境 $(Host)"; exit 1; \
	else \
		echo "已经连接远程环境 $(Host)"; \
	fi
	@echo "==== 构建并运行 ===="
	@mkdir -p /sysvol/config || exit 1
	@mkdir -p /sysvol/secret || exit 1
	@echo "复制配置文件..."
	@cp -f ./server/infra/config/agent-retrieval.yaml /sysvol/config/
	@cp -f ./server/infra/config/observability.yaml /sysvol/config/
	@cp -f ./server/infra/config/agent-retrieval-secret.yaml /sysvol/secret/
	@echo "启动服务..."
	go run ./server/main.go