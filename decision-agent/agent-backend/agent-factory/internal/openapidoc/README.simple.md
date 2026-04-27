# `internal/openapidoc` 简化说明

## 这个目录是做什么的

`internal/openapidoc` 是 OpenAPI 文档构建引擎。

它负责把 Swagger 输入一步步加工成最终文档产物，包括：

1. Swagger 2.0 转 OpenAPI 3
2. 路径重写
3. overlay 合并
4. baseline 补齐
5. 归一化和校验
6. 输出 JSON / YAML / HTML
7. 生成 compare report

## 最重要的入口

先看 `build.go` 里的 `BuildArtifactsFromFiles`。

这是整个目录的主入口，几乎所有文件都是它串起来的。

如果只想快速看懂这套流程，按下面顺序阅读：

1. `doc.go`
2. `build.go`
3. `convert.go`
4. `merge.go`
5. `normalize.go`
6. `document.go`
7. `compare.go`
8. `render.go`

## 每个模块是做什么的

- `build.go`
  - 主流程编排
- `convert.go`
  - Swagger 2.0 转 OpenAPI 3，补路径前缀
- `merge.go`
  - 合并 overlay，或从 baseline 补缺
- `normalize.go`
  - 做路径参数、operationId、安全配置归一化，并做最终校验
- `document.go`
  - 文档对象和 JSON / YAML / 文件之间的转换
- `compare.go`
  - 生成差异报告
- `render.go`
  - 生成静态 Scalar HTML
- `sanitize.go`
  - 清洗历史非标准字段
- `operations.go`
  - 路径数、接口数和差异统计工具

## 你只需要记住的核心点

- 真正入口是 `BuildArtifactsFromFiles`
- 生成不对先看 `build.go`
- 路径不对先看 `convert.go`
- 补齐/覆盖不对先看 `merge.go`
- 校验或安全定义不对先看 `normalize.go`
- 页面标题、favicon、静态 HTML 不对先看 `render.go`

## 想看完整版

查看 [README.md](./README.md)

