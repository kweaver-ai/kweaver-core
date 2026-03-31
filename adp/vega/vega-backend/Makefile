.DEFAULT_GOAL := help
.PHONY: help generate-mock lint test test-cover test-race test-at test-at-catalog test-at-verbose test-at-catalog-verbose test-at-setup ci clean-test

help:
	@echo "vega-backend - 可用命令:"
	@echo "  make test             - 运行单元测试"
	@echo "  make test-cover       - 运行单元测试 + 覆盖率"
	@echo "  make lint             - 静态检查"
	@echo "  make ci               - CI 入口 (lint + test-cover)"
	@echo "  make test-at          - 运行所有 AT 测试"
	@echo "  make test-at-catalog  - 运行 Catalog AT 测试"
	@echo "  make test-at-setup    - 检查 AT 测试配置"

generate-mock:
	cd server && go generate ./...

lint:
	cd server && golangci-lint run ./... --exclude-dirs=interfaces/mock,tests

UT_PACKAGES = $(shell cd server && go list ./... | grep -v /tests/ | grep -v /mock)

test:
	cd server && go test $(UT_PACKAGES) -gcflags=all=-l -v -count=1

test-cover:
	@mkdir -p test-result
	cd server && go test $(UT_PACKAGES) -gcflags=all=-l \
		-coverprofile=../test-result/coverage.out \
		-covermode=atomic
	go tool cover -func=test-result/coverage.out
	go tool cover -html=test-result/coverage.out -o test-result/coverage.html
	@command -v gocover-cobertura >/dev/null 2>&1 && \
		gocover-cobertura < test-result/coverage.out > test-result/coverage.xml || \
		echo "WARN: gocover-cobertura not found, coverage.xml not generated"

test-race:
	cd server && go test $(UT_PACKAGES) -gcflags=all=-l -race -count=1

# === AT Tests (需要运行中的服务 + 外部依赖) ===

test-at: test-at-setup
	@echo "运行 AT 测试..."
	cd tests && go test -v ./at/... -timeout 5m

test-at-catalog: test-at-setup
	@echo "运行 Catalog AT 测试..."
	cd tests && go test -v ./at/catalog/... -timeout 2m

test-at-verbose: test-at-setup
	@echo "运行 AT 测试（详细输出）..."
	cd tests && go test -v -count=1 ./at/... -timeout 5m

test-at-catalog-verbose: test-at-setup
	@echo "运行 Catalog AT 测试（详细输出）..."
	cd tests && go test -v -count=1 ./at/catalog/... -timeout 2m

test-at-setup:
	@if [ ! -f tests/at/testdata/test-config.yaml ]; then \
		echo "错误: 测试配置文件不存在"; \
		echo "请复制: cp tests/at/testdata/test-config.yaml.example tests/at/testdata/test-config.yaml"; \
		exit 1; \
	fi
	@echo "测试配置文件已就绪"

ci: lint test-cover

clean-test:
	rm -rf test-result
	cd server && rm -f coverage.out
