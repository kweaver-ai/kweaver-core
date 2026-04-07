# `cmd/openapi-docs` 简化说明

## 这个目录是做什么的

`cmd/openapi-docs` 是 OpenAPI 文档工具的命令行入口目录。

它主要负责三件事：

1. 接收命令行参数。
2. 调用 `internal/openapidoc` 完成真正的文档构建。
3. 把最终结果写到 `docs/api` 和 `test_out`。

## 最重要的入口

- 先看 `main.go`
- 生成文档看 `generate.go`
- 做差异报告看 `compare.go`
- 做结果校验看 `validate.go`

如果只想快速建立心智模型，按下面顺序看就够了：

1. `main.go`
2. `generate.go`
3. `compare.go`
4. `validate.go`
5. `internal/openapidoc/build.go`

## 三个子命令分别做什么

### `generate`

生成最终产物：

- `docs/api/agent-factory.json`
- `docs/api/agent-factory.yaml`
- `docs/api/agent-factory.html`

常用命令：

```bash
go run ./cmd/openapi-docs generate
```

### `compare`

只生成对比报告，不写最终文档。

常用命令：

```bash
go run ./cmd/openapi-docs compare
```

### `validate`

校验生成结果是否合法，并检查路径数、接口数、HTML 标记是否正常。

常用命令：

```bash
go run ./cmd/openapi-docs validate
```

## 每个文件快速认识

- `main.go`
  - 命令分发入口
- `generate.go`
  - 生成 JSON / YAML / HTML 的主流程
- `compare.go`
  - 生成 compare report
- `validate.go`
  - 校验产物
- `constants.go`
  - 默认输入输出路径和默认计数
- `common.go`
  - 共享的小工具函数

## 你只需要记住的核心点

- CLI 入口在 `main.go`
- 真正的构建核心在 `internal/openapidoc/BuildArtifactsFromFiles`
- 平时最常用的是 `generate` 和 `validate`

## 想看完整版

查看 [README.md](./README.md)

