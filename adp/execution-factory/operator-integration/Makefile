.DEFAULT_GOAL := help
.PHONY: help generate-mock lint test test-cover test-race test-integration test-at test-performance ci

help:
	@echo "operator-integration - available commands:"
	@echo "  make test             - run unit tests"
	@echo "  make test-cover       - run unit tests with coverage output"
	@echo "  make test-race        - run unit tests with race detection"
	@echo "  make lint             - run golangci-lint"
	@echo "  make ci               - run lint and unit test coverage"
	@echo "  make generate-mock    - regenerate mocks"
	@echo "  make test-integration - integration test placeholder"
	@echo "  make test-at          - acceptance test placeholder"
	@echo "  make test-performance - performance test placeholder"

UT_PACKAGES := $(shell go list ./... | \
	grep -v /server/tests | \
	grep -v /server/mocks)

generate-mock:
	go generate ./...

lint:
	golangci-lint run --exclude-dirs=server/tests,server/mocks --timeout=10m ./...

test:
	go test $(UT_PACKAGES) -gcflags=all=-l -v -count=1

test-cover:
	@mkdir -p test-result
	go test $(UT_PACKAGES) -gcflags=all=-l -count=1 \
		-coverprofile=test-result/coverage.out \
		-covermode=atomic
	go tool cover -func=test-result/coverage.out
	go tool cover -html=test-result/coverage.out -o test-result/coverage.html
	@command -v gocover-cobertura >/dev/null 2>&1 && gocover-cobertura < test-result/coverage.out > test-result/coverage.xml || \
		(echo "WARN: gocover-cobertura not found, coverage.xml not generated; install for CI compliance")

test-race:
	go test $(UT_PACKAGES) -gcflags=all=-l -race -count=1

test-integration:
	@echo "No integration test target is configured for this module yet."

test-at:
	@echo "No acceptance test target is configured for this module yet."

test-performance:
	@echo "No performance test target is configured for this module yet."

ci: lint test-cover
