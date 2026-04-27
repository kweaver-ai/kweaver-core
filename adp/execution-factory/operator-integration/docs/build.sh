#!/bin/bash
# 安装swagger npm install swagger-cli -g
# swagger-cli bundle apis/api_private/index.yaml --outfile apis/api_private/index.json --type json
# redoc-cli bundle -o apis/api_private/index.html apis/api_private/index.json

# 算子
swagger-cli bundle apis/api_public/operator.yaml --outfile apis/api_public/operator.json --type json
redoc-cli bundle -o apis/api_public/operator.html apis/api_public/operator.json
swagger-cli bundle apis/api_private/operator.yaml --outfile apis/api_private/operator.json --type json
redoc-cli bundle -o apis/api_private/operator.html apis/api_private/operator.json

# 工具
swagger-cli bundle apis/api_public/toolbox.yaml --outfile apis/api_public/toolbox.json --type json
redoc-cli bundle -o apis/api_public/toolbox.html apis/api_public/toolbox.json
swagger-cli bundle apis/api_private/toolbox.yaml --outfile apis/api_private/toolbox.json --type json
redoc-cli bundle -o apis/api_private/toolbox.html apis/api_private/toolbox.json

# mcp
swagger-cli bundle apis/api_public/mcp.yaml --outfile apis/api_public/mcp.json --type json
redoc-cli bundle -o apis/api_public/mcp.html apis/api_public/mcp.json
swagger-cli bundle apis/api_private/mcp.yaml --outfile apis/api_private/mcp.json --type json
redoc-cli bundle -o apis/api_private/mcp.html apis/api_private/mcp.json

# 导入导出
swagger-cli bundle apis/api_public/impex.yaml --outfile apis/api_public/impex.json --type json
redoc-cli bundle -o apis/api_public/impex.html apis/api_public/impex.json
# pandoc apis/readme.md -o apis/readme.html --metadata title="算子平台接口文档"