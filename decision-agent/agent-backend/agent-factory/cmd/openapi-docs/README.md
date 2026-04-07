# `cmd/openapi-docs` 目录说明

## 快速入口

- 精简版说明：查看 [README.simple.md](./README.simple.md)
- 完整版说明：继续阅读当前文件

如果你只是想快速知道这个目录做什么、入口在哪里、应该从哪个文件开始看，优先看精简版就够了。

## 目录职责

`cmd/openapi-docs` 是 OpenAPI 文档工具的命令行入口目录，负责把 `docs/swagger.json`、`docs/api/overlay.yaml`、`docs/api/baseline/agent-factory.json` 这些输入文件串起来，产出：

- `docs/api/agent-factory.json`
- `docs/api/agent-factory.yaml`
- `docs/api/agent-factory.html`
- `test_out/openapi_compare_report.md`

这个目录本身主要负责三件事：

1. 定义命令行子命令和默认参数。
2. 把参数组装成 `internal/openapidoc` 的构建请求。
3. 控制输出文件的写入、计数校验和错误退出。

## 从哪里开始看

### 命令行入口

优先从 `main.go` 开始看。

原因：

- 所有 CLI 子命令都从这里分发。
- 读完 `main.go` 就能知道有哪些入口命令。
- 再顺着 `runGenerate`、`runCompare`、`runValidate` 往下读，很容易建立整体心智模型。

### 如果你的目标是“生成文档”

推荐阅读顺序：

1. `main.go`
2. `generate.go`
3. `constants.go`
4. `internal/openapidoc/build.go`

### 如果你的目标是“看差异报告”

推荐阅读顺序：

1. `main.go`
2. `compare.go`
3. `internal/openapidoc/compare.go`

### 如果你的目标是“看校验逻辑”

推荐阅读顺序：

1. `main.go`
2. `validate.go`
3. `internal/openapidoc/normalize.go`
4. `internal/openapidoc/operations.go`

## 子命令概览

### `generate`

作用：

- 完整生成最终 JSON、YAML、HTML 文档。
- 可选生成 compare report。
- 允许启用 baseline fallback，把生成结果里缺失但 baseline 里仍保留的内容补回去。

常用命令：

```bash
go run ./cmd/openapi-docs generate
```

### `compare`

作用：

- 只生成 compare report。
- 不写最终 JSON / YAML / HTML。
- 适合排查“这次生成和 baseline 到底差了什么”。

常用命令：

```bash
go run ./cmd/openapi-docs compare
```

### `validate`

作用：

- 校验生成后的 OpenAPI JSON 是否能被 kin-openapi 正常加载和校验。
- 校验路径数、接口数是否符合预期。
- 校验静态 HTML 是否包含 Scalar 所需标记。

常用命令：

```bash
go run ./cmd/openapi-docs validate
```

## 文件说明

### `main.go`

这是整个 CLI 的总入口文件。

核心函数：

- `main`
  - 读取 `os.Args`
  - 识别当前子命令
  - 分发到 `runGenerate`、`runCompare`、`runValidate`
  - 在失败时统一走 `exitWithError`

适合谁先看：

- 第一次接触这个工具的人
- 想快速知道“支持哪些命令”的人

### `constants.go`

这个文件保存命令行工具用到的默认路径和默认校验计数。

主要内容：

- 默认 Swagger 输入路径
- 默认 overlay 路径
- 默认 baseline 路径
- 默认输出 JSON / YAML / HTML 路径
- 默认 compare report 路径
- 默认期望路径数和接口数

适合谁先看：

- 想知道工具默认读写哪些文件的人
- 调整文档产物路径的人

### `common.go`

这个文件放的是 CLI 多个子命令共享的小工具函数。

核心函数：

- `usageError`
  - 返回统一的用法错误信息
- `optionalPath`
  - 把空串或 `-` 转成“未配置”
  - 用于跳过可选输入文件或输出文件
- `exitWithError`
  - 负责把错误打印到 stderr 并退出进程

适合谁先看：

- 想理解参数兜底和错误处理方式的人

### `generate.go`

这个文件承载“生成最终文档”的主流程。

核心函数：

- `runGenerate`
  - 解析 `generate` 子命令的 flags
  - 组装 `openapidoc.BuildOptions`
  - 调用 `openapidoc.BuildArtifactsFromFiles`
  - 把产物写到 JSON / YAML / HTML 文件
  - 在 compare report 非空时写出对比报告
  - 打印生成后的路径数与接口数

这是生成链路里最重要的 CLI 文件。

如果你想排查“为什么产物长这样”，这里是最值得先看的文件之一。

### `compare.go`

这个文件承载“只做差异分析”的流程。

核心函数：

- `runCompare`
  - 解析 `compare` 子命令的 flags
  - 调用 `BuildArtifactsFromFiles` 生成当前文档
  - 禁用 baseline fallback
  - 只写 compare report，不写最终文档产物

典型使用场景：

- 检查新生成内容是否比 baseline 少字段
- 评估某次文档改动会不会影响已有接口描述

### `validate.go`

这个文件承载“校验生成结果”的流程。

核心函数：

- `runValidate`
  - 解析 `validate` 子命令的 flags
  - 读取 OpenAPI JSON
  - 执行路径参数、operationId 等归一化
  - 调用 `ValidateOpenAPI`
  - 校验路径数、接口数
  - 校验静态 HTML 是否包含 Scalar 必需标记

典型使用场景：

- 生成后做 smoke check
- CI 或本地开发时快速判断文档有没有明显退化

### `validate_test.go`

这是 CLI 目录下目前唯一的测试文件。

核心测试：

- `TestValidateDefaultsMatchGeneratedSpec`
  - 校验 `defaultExpectPaths` 与 `defaultExpectOps` 是否和当前生成产物一致
  - 避免默认校验阈值过期

这个测试解决的是“代码默认值跟真实产物不同步”的问题。

## 调用关系

最常见的一条生成链路如下：

```text
main.go
  -> runGenerate
  -> openapidoc.BuildArtifactsFromFiles
  -> WriteFile(JSON / YAML / HTML / report)
```

校验链路如下：

```text
main.go
  -> runValidate
  -> LoadOpenAPIDocFile
  -> NormalizePathParameters / NormalizeOperationIDs
  -> ValidateOpenAPI
  -> CountPathsAndOperations
```

对比链路如下：

```text
main.go
  -> runCompare
  -> openapidoc.BuildArtifactsFromFiles
  -> BuildComparisonReport
  -> WriteFile(report)
```

## 维护建议

### 当你新增子命令时

需要同步修改：

- `main.go`
- 新的 `runXxx.go`
- 如有默认参数，补到 `constants.go`

### 当你修改默认产物路径时

需要同步检查：

- `constants.go`
- 相关 README
- 外部依赖这些路径的脚本或测试

### 当生成结果的路径数 / 接口数变化时

需要同步检查：

- `constants.go` 里的 `defaultExpectPaths` / `defaultExpectOps`
- `validate_test.go`
- `go run ./cmd/openapi-docs validate`

## 一句话总结

如果只记住一个入口，请记住：

- CLI 入口在 `main.go`
- 真正的构建主流程入口在 `generate.go` 的 `runGenerate`
- 再往下就进入 `internal/openapidoc/BuildArtifactsFromFiles`
