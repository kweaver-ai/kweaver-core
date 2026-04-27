.DEFAULT_GOAL := gen-swag

SWAG ?= swag
SWAG_OUT := docs/swagger
IMAGE ?= swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability:local
CHART_PATH := charts/agent-observability
GOCACHE ?= /tmp/go-build

test:
	GOCACHE=$(GOCACHE) go test ./...

docker-build:
	docker build -t $(IMAGE) .

helm-lint:
	helm lint $(CHART_PATH)

helm-package:
	mkdir -p dist
	helm package $(CHART_PATH) --destination dist

gen-swag:
	$(SWAG) init -g main.go -o $(SWAG_OUT) --parseDependency --parseInternal

view-swag:
	@echo "Swagger JSON: http://localhost:8080/swagger/swagger.json"
	@echo "Swagger YAML: http://localhost:8080/swagger/swagger.yaml"

clean-swag:
	rm -rf $(SWAG_OUT)
	$(MAKE) gen-swag
