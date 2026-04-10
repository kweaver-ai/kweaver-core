# Agent Factory API 文档

这个目录只保留对外可查看的最终 OpenAPI 文档产物。

## 目录里的文件

- `agent-factory.json`
  - OpenAPI 3 JSON 版本，适合工具接入
- `agent-factory.yaml`
  - OpenAPI 3 YAML 版本，适合人工阅读和分发
- `agent-factory.html`
  - 可直接打开的静态文档页面
- `favicon.png`
  - 静态文档页面图标

## 如何生成

在 `agent-backend/agent-factory` 目录下执行：

```bash
make gen-api-docs
```

## 如何校验

```bash
make validate-api-docs
```

这个命令会同时检查：

- OpenAPI 结构是否合法
- 路径数、接口数是否符合预期
- 这里的公共文档是否与运行时副本一致

## 如何查看

### 直接查看静态页面

直接打开当前目录下的 `agent-factory.html`。

### 启动服务后查看

启动 Agent Factory 后访问：

- `http://127.0.0.1:30777/swagger/index.html`
- `http://127.0.0.1:30777/swagger/doc.json`
- `http://127.0.0.1:30777/swagger/doc.yaml`

## 深入维护说明

如果你需要了解生成链路、overlay、baseline 或 Swagger 中间产物，请查看：

- `../../cmd/openapi-docs/README.simple.md`
- `../../cmd/openapi-docs/docs/OPENAPI_AUTOMATION_GUIDE.md`
