.DEFAULT_GOAL := factory_pub_html

OPENAPI_CLI = redocly
OPENAPI_YAML = agent-factory.yaml
HTML_OUTPUT = agent-factory.html

OPENAPI_JSON = agent-factory.json

factory_pub_html:
	$(OPENAPI_CLI) build-docs $(OPENAPI_YAML) -o $(HTML_OUTPUT);open $(HTML_OUTPUT)
	make yaml2json

preview:
	@echo "Starting API documentation preview server with hot reload..."
	@echo "Open http://127.0.0.1:8080 in your browser"
	@echo "👀 Watching for file changes - any changes will auto-reload in browser"
	@echo "Press Ctrl+C to stop the server"
	@npx @redocly/cli@1 preview-docs $(OPENAPI_YAML) --port 8080

# 将YAML转换为JSON格式
yaml2json:
	@echo "Converting YAML to JSON..."
	$(OPENAPI_CLI) bundle $(OPENAPI_YAML) --output $(OPENAPI_JSON)

# 验证OpenAPI规范
validate:
	@echo "Validating OpenAPI specification..."
	$(OPENAPI_CLI) lint $(OPENAPI_YAML)

# 清理生成的文件
clean:
	@echo "Cleaning generated files..."
	rm -f $(HTML_OUTPUT) $(OPENAPI_JSON)

.PHONY: preview yaml2json validate clean

# 使用openapi-generator生成HTML文档 (保留作为备选方案)
# generate-html:
# 	npx @openapitools/openapi-generator-cli generate \
# 	  -i $(OPENAPI_YAML) \
# 	  -g html2 \
# 	  -o html-doc

factory_pub_html2:
	openapi-generator-cli generate \
	  -i $(OPENAPI_YAML) \
	  -g html2 \
	  -o /Users/Zhuanz/Work/as/dip_ws/api-docs/.local/gen/



json2yaml:
	$(OPENAPI_CLI) bundle $(OPENAPI_JSON) -o $(OPENAPI_YAML)

