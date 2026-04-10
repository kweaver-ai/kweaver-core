.DEFAULT_GOAL := lint

SWAG_CMD := go run github.com/swaggo/swag/cmd/swag@v1.16.6
OPENAPI_DOCS_CMD := go run ./cmd/openapi-docs

ciLint:
	golangci-lint run --out-format=colored-line-number ./...
	@echo "ciLint done"

ciLintFix:
	@golangci-lint run --out-format=colored-line-number ./... --fix
	@echo "ciLintFix done"

wsl:
	@wsl --fix ./... >/dev/null 2>&1 ||true
	@echo "wsl done"

fmt:
	@gofumpt -w .
	@goimports -w ./**/*.go
	@echo "fmt done"

lint:
	$(MAKE) wsl
	$(MAKE) fmt
	$(MAKE) ciLintFix

ut:
	rm -rf coverage_report
	mkdir coverage_report
	go generate ./...
	go test -v -coverprofile=coverage_report/cover.out  ./... 2>&1 | go-junit-report > coverage_report/lint_report.xml
	gocov convert coverage_report/cover.out | gocov-xml > coverage_report/coverage.xml
	gocov convert coverage_report/cover.out | gocov-html > coverage_report/coverage.html

goGenerate:
	go generate ./...

goTest:
	go test -v $$(go list ./... | grep -vE '/[^/]*mock[^/]*($$|/)')

# 排除不需要测试的包列表:
# - *mock*: 自动生成的 mock 代码，无业务逻辑
# - boot: init() 依赖真实 DB/Redis/MQ，UT 环境无法连接
# - port/driven/idbaccess$$: 纯 interface 定义，无可执行代码
# - port/driven/ihttpaccess: 纯 interface 定义
# - port/driven/imqaccess: 纯 interface 定义
# - port/driven/iredisaccess: 纯 interface 定义
# - port/driver: 纯 interface 定义
# - infra/opentelemetry/(logs|metrics|middleware|trace): 依赖外部 OTLP Collector，UT 环境 panic
# - infra/interf: 纯 interface 定义
# - infra/common/cglobal: 纯全局变量定义
# - infra/common/constant: 纯常量定义
# - infra/cmp/opensearchcmp: 依赖 OpenSearch
# - infra/cmp/efastcmp: 依赖 eFast HTTP
# - infra/cmp/redisdlmcmp: 依赖 Redis
# - infra/server/httpserver: 服务器启动代码
# - infra/common/cconf: 配置加载基础库
UT_EXCLUDE_PATTERN := mock|boot|port/driven/idbaccess$$|port/driven/ihttpaccess|port/driven/imqaccess|port/driven/iredisaccess|port/driver|infra/opentelemetry/(logs|metrics|middleware|trace)|infra/interf|infra/common/cglobal|infra/common/constant|infra/cmp/opensearchcmp|infra/cmp/efastcmp|infra/cmp/redisdlmcmp|infra/server/httpserver|infra/common/cconf

utExclude:
	go test -v -count=1 $$(go list ./... | grep -vE '$(UT_EXCLUDE_PATTERN)')

utCoverage:
	@mkdir -p test_out
	go test -count=1 -timeout 30m \
		-coverprofile=test_out/coverage_full.out \
		-covermode=atomic \
		$$(go list ./... | grep -vE '$(UT_EXCLUDE_PATTERN)') \
		2>&1 | tee test_out/ut_full_output.log
	@go tool cover -func=test_out/coverage_full.out > test_out/coverage_by_func.txt
	@go tool cover -func=test_out/coverage_full.out | tail -1 > test_out/coverage_total.txt
	@grep -E "^(ok|FAIL)" test_out/ut_full_output.log > test_out/ut_timing.log
	@echo "Coverage report saved to test_out/"
	@cat test_out/coverage_total.txt

serviceUtAndCoverage:
	@mkdir -p test_out
	@echo "========================================="
	@echo "Running Service UT: ./src/domain/service/..."
	@echo "========================================="
	go test -timeout 30m \
		-coverprofile=test_out/coverage_service.out \
		-covermode=atomic \
		./src/domain/service/... \
		2>&1 | tee test_out/ut_service_output.log || true
	@echo ""
	@echo "========================================="
	@echo "Coverage by function:"
	@echo "========================================="
	@go tool cover -func=test_out/coverage_service.out 2>/dev/null || true
	@echo ""
	@echo "========================================="
	@echo "Total coverage:"
	@echo "========================================="
	@go tool cover -func=test_out/coverage_service.out 2>/dev/null | tail -1 || true
	@echo "========================================="
# Swagger 文档生成
gen-swag:
	$(SWAG_CMD) init -g main.go -o cmd/openapi-docs/generated/swagger --parseDependency

gen-api-docs: gen-swag
	$(OPENAPI_DOCS_CMD) generate

# 查看文档（需要先启动服务）
view-swag:
	@echo "Scalar: http://localhost:30777/scalar"
	@open http://localhost:30777/scalar

# 清理并重新生成 Swagger
clean-swag:
	rm -f cmd/openapi-docs/generated/swagger/docs.go cmd/openapi-docs/generated/swagger/swagger.json cmd/openapi-docs/generated/swagger/swagger.yaml
	$(MAKE) gen-swag

compare-api-docs: gen-swag
	$(OPENAPI_DOCS_CMD) compare

validate-api-docs:
	$(OPENAPI_DOCS_CMD) validate

add:
	goimports -w ./**/*.go
	git add .

pull:
	git pull origin $(shell git rev-parse --abbrev-ref HEAD)

push:
	git pull origin $(shell git rev-parse --abbrev-ref HEAD)
	git push origin $(shell git rev-parse --abbrev-ref HEAD)
