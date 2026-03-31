.DEFAULT_GOAL := mcp_2_html

OPENAPI_CLI = redocly

# --- MPC START ---
mcp_2_html:
	$(OPENAPI_CLI) build-docs mcp.yaml -o mcp.html;open mcp.html
	make mcp_yaml2json

mcp_yaml2json:
	$(OPENAPI_CLI) bundle mcp.yaml -o mcp.json

mcp_json2yaml:
	$(OPENAPI_CLI) bundle mcp.json -o mcp.yaml

# --- MPC END ---


# --- TOOLBOX START ---
toolbox_2_html:
	$(OPENAPI_CLI) build-docs toolbox.yaml -o toolbox.html;open toolbox.html

toolbox_yaml2json:
	$(OPENAPI_CLI) bundle toolbox.yaml -o toolbox.json

toolbox_json2yaml:
	$(OPENAPI_CLI) bundle toolbox.json -o toolbox.yaml

# --- TOOLBOX END ---

# --- PREVIEW COMMANDS ---
preview_mcp:
	@echo "Starting MCP API documentation preview server with hot reload..."
	@echo "Open http://127.0.0.1:8080 in your browser"
	@echo "👀 Watching for file changes - any changes will auto-reload in browser"
	@echo "Press Ctrl+C to stop the server"
	@npx @redocly/cli@1 preview-docs mcp.yaml --port 8080

preview_toolbox:
	@echo "Starting Toolbox API documentation preview server with hot reload..."
	@echo "Open http://127.0.0.1:8081 in your browser"
	@echo "👀 Watching for file changes - any changes will auto-reload in browser"
	@echo "Press Ctrl+C to stop the server"
	@npx @redocly/cli@1 preview-docs toolbox.yaml --port 8081

# --- PREVIEW END ---
