# Agent Factory OpenAPI 自动化维护说明

## 目标

这套链路把 OpenAPI 文档拆成了三层：

1. Swagger 2.0 中间产物
2. OpenAPI 3 最终对外产物
3. 运行时嵌入副本

这样做的目的有两个：

- `docs/api` 只保留普通用户真正会查看的最终文档
- 生成链路所需的 overlay、baseline、中间产物都收口到 `cmd/openapi-docs`

## 目录分工

### 公共文档

- `../../docs/api/agent-factory.json`
- `../../docs/api/agent-factory.yaml`
- `../../docs/api/agent-factory.html`
- `../../docs/api/agent-factory-redoc.html`
- `../../docs/api/favicon.png`

这些文件是对外文档目录中的最终结果。

其中：

- `docs/api/*.html` 对外分发时默认使用 `cdn.jsdmirror.com`
- 不再要求同步 `docs/api/ui/*`

### 生成输入

- `../generated/swagger/swagger.json`
- `../generated/swagger/swagger.yaml`
- `../generated/swagger/docs.go`
- `../assets/overlay.yaml`
- `../assets/baseline/agent-factory.json`
- `../../src/infra/server/apidocs/assets/favicon.png`

这些文件是开发和生成链路内部使用的输入。

### 运行时副本

- `../../src/infra/server/apidocs/assets/agent-factory.json`
- `../../src/infra/server/apidocs/assets/agent-factory.yaml`
- `../../src/infra/server/apidocs/assets/agent-factory.html`
- `../../src/infra/server/apidocs/assets/agent-factory-redoc.html`
- `../../src/infra/server/apidocs/assets/favicon.png`
- `../../src/infra/server/apidocs/assets/ui/*`

服务运行时通过 `src/infra/server/apidocs/embed.go` 将这组文件打进二进制。

## 生成流程

1. 在 handler 和类型上维护 `swaggo` 注释
2. 执行 `make gen-swag`
3. Swagger 中间产物写入 `cmd/openapi-docs/generated/swagger`
4. 执行 `make gen-api-docs`
5. `cmd/openapi-docs generate` 读取 Swagger / overlay / baseline / 运行时 favicon / 本地 UI 资源
6. 生成最终 OpenAPI 3 JSON、YAML、Scalar HTML、Redoc HTML
7. 同步写入公共文档目录和运行时副本目录
8. 执行 `make validate-api-docs` 做结构校验、HTML 依赖策略校验和副本一致性校验

## 常用命令

### 生成 Swagger 中间产物

```bash
make gen-swag
```

### 生成最终文档

```bash
make gen-api-docs
```

### 生成差异报告

```bash
make compare-api-docs
```

### 校验最终文档

```bash
make validate-api-docs
```

## 维护规则

### 为什么 favicon 还保留两份

`favicon.png` 现在最少保留两份：

1. `../../docs/api/favicon.png`
   - 给用户直接打开静态文档时使用
2. `../../src/infra/server/apidocs/assets/favicon.png`
   - 给运行时 `go:embed` 使用

之前额外的 `cmd/openapi-docs/assets/favicon.png` 已移除，不再保留第三份。

### 什么时候改代码注释

优先修改真实 handler 上的 `swaggo` 注释。接口定义、参数、响应结构的主来源应当是代码。

### 什么时候改 overlay

只有代码注释不适合表达、或者必须做显式覆盖时，才修改 `../assets/overlay.yaml`。

### 什么时候改 baseline

当你确认最终对外接口变化是预期行为，并且 compare report 中的差异需要被接受时，再更新 `../assets/baseline/*`。

### 什么时候需要看运行时副本

当 `/scalar/doc.json`、`/scalar/doc.yaml`、`/scalar/index.html` 表现异常时，优先检查：

1. `make gen-api-docs` 是否已执行
2. `../../src/infra/server/apidocs/assets/*.json`、`*.yaml`、`favicon.png` 是否与 `../../docs/api/*` 一致
3. `make validate-api-docs` 是否通过

当 `/redoc/index.html` 或运行中的 `/scalar` 页面异常时，也优先检查上述三项，并确认 `../../src/infra/server/apidocs/assets/ui/*` 是否已经同步。

当离线分发的 `docs/api/*.html` 异常时，优先检查：

1. 当前网络是否可访问 `cdn.jsdmirror.com`
2. `make validate-api-docs` 是否通过
3. HTML 中引用的 CDN URL 是否仍为版本固定的预期地址

## 相关文档

- 目录入口：查看 [../README.md](../README.md)
- 生成链路：查看 [swagger_openapi_pipeline.md](./swagger_openapi_pipeline.md)
- 对外文档使用说明：查看 [../../docs/api/README.md](../../docs/api/README.md)
