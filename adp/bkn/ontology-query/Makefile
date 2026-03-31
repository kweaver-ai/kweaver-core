.DEFAULT_GOAL := help
.PHONY: help generate-mock lint test test-cover test-race ci clean-test

help:
	@echo "ontology-query - 可用命令:"
	@echo "  make test             - 运行单元测试"
	@echo "  make test-cover       - 运行单元测试 + 覆盖率"
	@echo "  make test-race        - 运行单元测试 + 竞态检测"
	@echo "  make lint             - 静态检查"
	@echo "  make ci               - CI 入口 (lint + test-cover)"
	@echo "  make generate-mock    - 重新生成 mock 文件"

generate-mock:
	cd server && go generate ./...

lint:
	cd server && golangci-lint run ./... --exclude-dirs=interfaces/mock

UT_PACKAGES = $(shell cd server && go list ./... | grep -v /mock)

test:
	cd server && I18N_MODE_UT=true go test $(UT_PACKAGES) -gcflags=all=-l -v -count=1

test-cover:
	@mkdir -p test-result
	cd server && I18N_MODE_UT=true go test $(UT_PACKAGES) -gcflags=all=-l \
		-coverprofile=../test-result/coverage.out \
		-covermode=atomic
	go tool cover -func=test-result/coverage.out
	go tool cover -html=test-result/coverage.out -o test-result/coverage.html
	@command -v gocover-cobertura >/dev/null 2>&1 && \
		gocover-cobertura < test-result/coverage.out > test-result/coverage.xml || \
		echo "WARN: gocover-cobertura not found, coverage.xml not generated"

test-race:
	cd server && I18N_MODE_UT=true go test $(UT_PACKAGES) -gcflags=all=-l -race -count=1

ci: lint test-cover

clean-test:
	rm -rf test-result
	cd server && rm -f coverage.out
