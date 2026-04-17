# `cmd/openapi-docs` 简化说明

## 这个目录是做什么的

`cmd/openapi-docs` 是 Agent Factory OpenAPI 文档工具的命令入口，也是文档生成链路的“开发者工作区”。

这里集中放了三类内容：

1. CLI 代码：`main.go`、`generate.go`、`compare.go`、`validate.go`
2. 生成输入：`assets/overlay.yaml`、`assets/baseline/*`、`generated/swagger/*`
3. 开发说明：`docs/OPENAPI_AUTOMATION_GUIDE.md`、`docs/swagger_openapi_pipeline.md`

## 生成后会写到哪里

`generate` 会一次写两套产物：

- 对外目录：`docs/api/agent-factory.{json,yaml}`、`docs/api/agent-factory.html`、`docs/api/agent-factory-redoc.html`、`docs/api/favicon.png` 与 `docs/api/ui/*`
- 运行时副本：`src/infra/server/apidocs/assets/*`

## 三个子命令

### `generate`

生成最终 OpenAPI JSON / YAML / 双 HTML 页面，并同步运行时副本。

```bash
go run ./cmd/openapi-docs generate
```

### `compare`

只生成 compare report，不改最终文档。

```bash
go run ./cmd/openapi-docs compare
```

### `validate`

校验最终文档是否合法，并检查公共产物与运行时副本是否一致。

```bash
go run ./cmd/openapi-docs validate
```

## 从哪里开始看

建议顺序：

1. `main.go`
2. `generate.go`
3. `validate.go`
4. `internal/openapidoc/build.go`

## 想看更详细的说明

- 目录说明：查看 [README.md](./README.md)
- 自动化维护说明：查看 [docs/OPENAPI_AUTOMATION_GUIDE.md](./docs/OPENAPI_AUTOMATION_GUIDE.md)
- 生成链路梳理：查看 [docs/swagger_openapi_pipeline.md](./docs/swagger_openapi_pipeline.md)
