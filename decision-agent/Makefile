# 静态代码检查
lint:
# 	@pip3 install pre-commit
	@pre-commit install
	@echo "pre-commit hooks installed."

	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	@echo "golangci-lint installed."
	@go install github.com/bombsimon/wsl/v4/cmd/wsl@v4.7.0
	@echo "wsl installed."
	@go install mvdan.cc/gofumpt@v0.9.2
	@echo "gofumpt installed."

	pre-commit run --all-files

# Act 本地测试 GitHub Actions
act-lint:
	act push -j agent-factory-lint -W .github/workflows/agent-factory-code-check.yml

act-ut:
	act push -j agent-factory-ut -W .github/workflows/agent-factory-code-check.yml

act-all:
	act push -W .github/workflows/agent-factory-code-check.yml
