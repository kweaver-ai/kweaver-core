.default: pri_html

# 定义变量
OPENAPI_CLI = redocly
OPENAPI_YAML = agent-app.yaml
HTML_OUTPUT = agent-app.html

OPENAPI_JSON = agent-app.json

# 默认目标
.PHONY: all
all: pri_html

# 生成HTML文档
.PHONY: pri_html
pri_html:
	$(OPENAPI_CLI) build-docs $(OPENAPI_YAML) -o $(HTML_OUTPUT);open $(HTML_OUTPUT)
	make yaml2json

yaml2json:
	$(OPENAPI_CLI) bundle $(OPENAPI_YAML) -o $(OPENAPI_JSON)

json2yaml:
	$(OPENAPI_CLI) bundle $(OPENAPI_JSON) -o $(OPENAPI_YAML)

preview:
	@echo "Starting API documentation preview server with hot reload..."
	@echo "Open http://127.0.0.1:8080 in your browser"
	@echo "👀 Watching for file changes - any changes will auto-reload in browser"
	@echo "Press Ctrl+C to stop the server"
	@npx @redocly/cli@1 preview-docs $(OPENAPI_YAML) --port 8080

