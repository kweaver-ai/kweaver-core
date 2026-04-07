.default: memory_public_html

# 定义变量
OPENAPI_CLI = redocly
OPENAPI_YAML = agent-memory.yaml
HTML_OUTPUT = agent-memory.html
OPENAPI_JSON = agent-memory.json

# 默认目标
.PHONY: all
all: memory_public_html

# 生成HTML文档
.PHONY: memory_public_html
memory_public_html:
	$(OPENAPI_CLI) build-docs $(OPENAPI_YAML) -o $(HTML_OUTPUT);open $(HTML_OUTPUT)
	make yaml2json

yaml2json:
	$(OPENAPI_CLI) bundle $(OPENAPI_YAML) -o $(OPENAPI_JSON)

json2yaml:
	$(OPENAPI_CLI) bundle $(OPENAPI_JSON) -o $(OPENAPI_YAML)

preview:
	@echo "正在启动 API 文档预览服务器(支持热重载)..."
	@echo "浏览器将自动打开 http://127.0.0.1:8080"
	@echo " 监听文件变化 - 任何修改都会自动在浏览器中重新加载"
	@echo "按 Ctrl+C 停止服务器"
	@open http://127.0.0.1:8080 &
	@npx --yes @redocly/cli@1 preview-docs $(OPENAPI_YAML) --port 8080