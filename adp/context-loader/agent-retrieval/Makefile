.DEFAULT_GOAL := help
.PHONY: help generate-mock lint test test-cover ci

PKG_LIST_CMD = go list ./server/... | grep -v '/server/tests/' | grep -v '/server/mocks'
SHELL_STRICT = set -eu; set -o pipefail 2>/dev/null || true;
TEST_RESULT_DIR = test-result

help:
	@echo "可用目标:"
	@echo "  generate-mock\t生成mock代码"
	@echo "  lint\t\t运行代码检查"
	@echo "  test\t\t执行单元测试"
	@echo "  test-cover\t执行单元测试并生成覆盖率报告"
	@echo "  ci\t\t执行CI检查（lint + test-cover）"

generate-mock:
	@echo "==== 生成mock代码 ===="
	go generate ./...

lint:
	@echo "==== 代码检查 ===="
	@$(SHELL_STRICT) \
	pkgs="$$( $(PKG_LIST_CMD) )"; \
	if [ -z "$$pkgs" ]; then \
		echo "ERROR: no packages found for lint"; \
		exit 1; \
	fi; \
	go vet $$pkgs

test:
	@echo "==== 执行测试 ===="
	@$(SHELL_STRICT) \
	pkgs="$$( $(PKG_LIST_CMD) | tr ' ' '\n' | grep -v '/server$$' )"; \
	if [ -z "$$pkgs" ]; then \
		echo "ERROR: no packages found for test"; \
		exit 1; \
	fi; \
	go test $$pkgs -v -count=1

test-cover:
	@echo "==== 执行测试并生成覆盖率报告 ===="
	@mkdir -p $(TEST_RESULT_DIR)
	@rm -f $(TEST_RESULT_DIR)/coverage.out
	@$(SHELL_STRICT) \
	pkgs="$$( $(PKG_LIST_CMD) | tr ' ' '\n' | grep -v '/server$$' )"; \
	if [ -z "$$pkgs" ]; then \
		echo "ERROR: no packages found for test-cover"; \
		exit 1; \
	fi; \
	first_profile=1; \
	for pkg in $$pkgs; do \
		profile="$(TEST_RESULT_DIR)/$$(echo $$pkg | tr '/.' '__').out"; \
		go test $$pkg -count=1 -coverprofile=$$profile -covermode=atomic || exit $$?; \
		if [ $$first_profile -eq 1 ]; then \
			cat $$profile > $(TEST_RESULT_DIR)/coverage.out; \
			first_profile=0; \
		else \
			tail -n +2 $$profile >> $(TEST_RESULT_DIR)/coverage.out; \
		fi; \
	done
	go tool cover -func=$(TEST_RESULT_DIR)/coverage.out
	go tool cover -html=$(TEST_RESULT_DIR)/coverage.out -o $(TEST_RESULT_DIR)/coverage.html
	@command -v gocover-cobertura >/dev/null 2>&1 && \
		gocover-cobertura < $(TEST_RESULT_DIR)/coverage.out > $(TEST_RESULT_DIR)/coverage.xml || \
		(echo "WARN: gocover-cobertura not found, coverage.xml not generated; install for CI compliance")

ci: lint test-cover
