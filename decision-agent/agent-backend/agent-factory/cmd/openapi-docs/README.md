# `cmd/openapi-docs` 目录说明

## 快速入口

- 精简版：查看 [README.simple.md](./README.simple.md)
- 自动化维护说明：查看 [docs/OPENAPI_AUTOMATION_GUIDE.md](./docs/OPENAPI_AUTOMATION_GUIDE.md)
- 生成链路梳理：查看 [docs/swagger_openapi_pipeline.md](./docs/swagger_openapi_pipeline.md)

## 目录职责

`cmd/openapi-docs` 负责把 Swagger 中间产物、overlay、baseline 串起来，生成最终 OpenAPI 文档，并同步运行时嵌入副本。

默认输入：

- `generated/swagger/swagger.json`
- `assets/overlay.yaml`
- `assets/baseline/agent-factory.json`
- `../../src/infra/server/apidocs/assets/favicon.png`

默认输出：

- `../../docs/api/agent-factory.json`
- `../../docs/api/agent-factory.yaml`
- `../../docs/api/agent-factory.html`
- `../../docs/api/agent-factory-redoc.html`
- `../../docs/api/favicon.png`
- `../../src/infra/server/apidocs/assets/*`
- `../../test_out/openapi_compare_report.md`

## 目录结构

```text
cmd/openapi-docs/
├── assets/
│   ├── baseline/          # compare / fallback 基线
│   └── overlay.yaml       # 手工覆盖层
├── docs/
│   ├── OPENAPI_AUTOMATION_GUIDE.md
│   └── swagger_openapi_pipeline.md
├── generated/swagger/     # swag 生成的 Swagger 2.0 中间产物
├── main.go
├── generate.go
├── compare.go
├── validate.go
└── constants.go
```

## 子命令概览

### `generate`

职责：

1. 读取 Swagger 2.0 中间产物
2. 交给 `internal/openapidoc` 转换成最终 OpenAPI 3
3. 写入 `docs/api` 对外产物
4. 同步写入 `src/infra/server/apidocs/assets` 运行时副本
5. 可选写 compare report

常用命令：

```bash
go run ./cmd/openapi-docs generate
```

### `compare`

职责：

1. 构建当前文档
2. 与 baseline 做差异比较
3. 只写 report，不写最终文档

常用命令：

```bash
go run ./cmd/openapi-docs compare
```

### `validate`

职责：

1. 校验最终 OpenAPI JSON 结构是否合法
2. 检查路径数、接口数
3. 检查公共 HTML 标记
4. 检查 `docs/api/*.html` 是否使用公共 CDN、`src/infra/server/apidocs/assets/*.html` 是否继续使用本地资源
5. 检查 `docs/api` 与 `src/infra/server/apidocs/assets` 的 JSON/YAML/favicon 是否一致

常用命令：

```bash
go run ./cmd/openapi-docs validate
```

## 最重要的文件

- `main.go`
  - CLI 分发入口
- `generate.go`
  - 生成最终产物并同步双份输出
- `compare.go`
  - 生成 compare report
- `validate.go`
  - 校验最终文档并检查双份副本一致性
- `artifacts.go`
  - 处理公共目录与运行时目录的同步写入/一致性校验
- `constants.go`
  - 默认输入输出路径

## 阅读建议

如果你是第一次接触这套链路，建议按下面顺序阅读：

1. `main.go`
2. `generate.go`
3. `artifacts.go`
4. `validate.go`
5. `internal/openapidoc/build.go`

如果你主要想理解原理，再继续看：

- [docs/OPENAPI_AUTOMATION_GUIDE.md](./docs/OPENAPI_AUTOMATION_GUIDE.md)
- [docs/swagger_openapi_pipeline.md](./docs/swagger_openapi_pipeline.md)
