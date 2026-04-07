# Agent Factory OpenAPI 自动化与 Scalar 文档使用说明

## 1. 背景与目标

本项目已经从“手写 `docs/api/agent-factory.{json,yaml,html}`”迁移到“代码注释自动生成 + OAS3 后处理 + Scalar 展示”的模式。

当前目标有两个：

1. 让接口文档尽可能从代码本身产生，避免手写 YAML 长期漂移。
2. 让最终产物和当前接受的 OpenAPI 基线保持一致，保证已有调用方、测试方式、阅读体验和页面入口不被破坏。

这套方案最终保留了两层文档：

1. `swaggo/swag` 生成的 Swagger 2.0（旧版 API 规范格式）中间产物。
2. 通过 `kin-openapi` 后处理得到的最终 OpenAPI 3.0.2 产物，以及对应的 Scalar 页面。

## 2. 当前产物概览

### 2.1 中间产物

- [docs/swagger.json](../swagger.json)
- [docs/swagger.yaml](../swagger.yaml)
- [docs/docs.go](../docs.go)

这三份文件由 `swag init` 直接生成，属于 Swagger 2.0（旧版 API 规范格式）的原始输出，不建议手工编辑。

### 2.2 最终产物

- [docs/api/agent-factory.json](./agent-factory.json)
- [docs/api/agent-factory.yaml](./agent-factory.yaml)
- [docs/api/agent-factory.html](./agent-factory.html)

这三份文件是最终对外的 OpenAPI 3.0.2 规范和静态文档页面。

### 2.3 基线文档

- [docs/api/baseline/agent-factory.json](./baseline/agent-factory.json)
- [docs/api/baseline/agent-factory.yaml](./baseline/agent-factory.yaml)
- [docs/api/baseline/agent-factory.html](./baseline/agent-factory.html)

这三份文件保存当前接受的比较基线，主要用于比较和兜底。它最初来自历史手写文档，后续会在明确的接口变更后同步更新。

### 2.4 对比报告

- [test_out/openapi_compare_report.md](../../test_out/openapi_compare_report.md)

这份报告会比较“当前自动生成的最终结果”和“当前接受的比较基线”的差异，覆盖：

1. path 数量
2. operation 数量
3. summary
4. tags
5. security
6. request body 是否存在
7. response code 集合

当前这份报告已经清零，说明自动生成结果与当前基线对齐。

## 3. 目录职责

### 3.1 生成命令入口

- [cmd/openapi-docs/main.go](../../cmd/openapi-docs/main.go)
- [cmd/openapi-docs/README.simple.md](../../cmd/openapi-docs/README.simple.md)
- [cmd/openapi-docs/README.md](../../cmd/openapi-docs/README.md)

职责：

1. `main.go` 负责子命令分发，只做入口路由。
2. `generate.go` / `compare.go` / `validate.go` 分别承载生成、对比、校验三条流程。
3. `common.go` 和 `constants.go` 提供共享工具函数和默认路径配置。
4. 这一层的职责是“接参数、调核心库、写产物”，真正的文档加工逻辑已经下沉到 `internal/openapidoc`。

### 3.2 生成逻辑核心

- [internal/openapidoc/build.go](../../internal/openapidoc/build.go)
- [internal/openapidoc/README.simple.md](../../internal/openapidoc/README.simple.md)
- [internal/openapidoc/README.md](../../internal/openapidoc/README.md)
- [internal/openapidoc/convert.go](../../internal/openapidoc/convert.go)
- [internal/openapidoc/merge.go](../../internal/openapidoc/merge.go)
- [internal/openapidoc/normalize.go](../../internal/openapidoc/normalize.go)
- [internal/openapidoc/document.go](../../internal/openapidoc/document.go)
- [internal/openapidoc/render.go](../../internal/openapidoc/render.go)
- [internal/openapidoc/compare.go](../../internal/openapidoc/compare.go)
- [internal/openapidoc/operations.go](../../internal/openapidoc/operations.go)
- [internal/openapidoc/sanitize.go](../../internal/openapidoc/sanitize.go)

职责：

1. `build.go` 里的 `BuildArtifactsFromFiles` 是当前真正的核心入口，负责把整条流水线编排起来。
2. `convert.go` 负责 Swagger 2.0（旧版 API 规范格式） -> OpenAPI 3.0.2 转换，以及路径统一改写为 `/api/agent-factory/...`。
3. `merge.go` 负责 overlay 覆盖合并，以及 baseline 缺失字段兜底。
4. `normalize.go` 负责路径参数、operationId、安全定义等归一化和最终校验前整理。
5. `document.go` 负责文档对象与 JSON / YAML / 文件之间的转换。
6. `render.go` 负责生成静态 Scalar HTML。
7. `compare.go` 和 `operations.go` 负责 compare report 与路径/接口统计。
8. `sanitize.go` 负责清洗历史 swagger/openapi 中的非标准字段形态。

### 3.2.1 测试文件拆分

现在测试也已经按模块拆分，不再集中在单个大测试文件中，常见入口包括：

- [internal/openapidoc/convert_test.go](../../internal/openapidoc/convert_test.go)
- [internal/openapidoc/merge_test.go](../../internal/openapidoc/merge_test.go)
- [internal/openapidoc/compare_test.go](../../internal/openapidoc/compare_test.go)
- [internal/openapidoc/normalize_path_test.go](../../internal/openapidoc/normalize_path_test.go)
- [internal/openapidoc/security_test.go](../../internal/openapidoc/security_test.go)
- [internal/openapidoc/render_test.go](../../internal/openapidoc/render_test.go)
- [internal/openapidoc/sanitize_test.go](../../internal/openapidoc/sanitize_test.go)
- [internal/openapidoc/validate_test.go](../../internal/openapidoc/validate_test.go)

对应原则是：

1. 每个测试文件尽量只覆盖一个主题模块。
2. 查问题时可以先按“功能模块”定位测试，而不是从一个超大测试文件里往下翻。

### 3.3 运行时文档路由

- [src/infra/server/httpserver/router_swagger.go](../../src/infra/server/httpserver/router_swagger.go)

职责：

1. 暴露 `/swagger/index.html` 作为 Scalar 主入口。
2. 暴露 `/scalar` 作为兼容入口。
3. 暴露 `/swagger/doc.json` 和 `/swagger/doc.yaml`。
4. 在 `/swagger/doc.json` 返回时，按当前请求动态重写 `servers[0].url`，保证 `Try it out` 默认命中当前服务。

### 3.4 最终文档嵌入

- [docs/api/embed.go](./embed.go)

职责：

1. 将最终 JSON/YAML/HTML 以 `embed` 方式打进 Go 二进制。
2. 避免运行时依赖工作目录去找文档文件。

### 3.5 全局文档元信息

- [main.go](../../main.go)

职责：

1. 声明 `@title`、`@version`、`@BasePath`、`@host`。
2. 提供 `BearerAuth` 安全定义，后续由后处理归一成最终的 `ApiKeyAuth` 语义。

### 3.6 代码内文档来源

当前代码内文档来源已经收敛为一类：

1. 真正写在 handler 方法上的注释。

例如：

- 真实 handler 注释：
  [src/driveradapter/api/httphandler/agenthandler/chat.go](../../src/driveradapter/api/httphandler/agenthandler/chat.go)
  [src/driveradapter/api/httphandler/conversationhandler/list.go](../../src/driveradapter/api/httphandler/conversationhandler/list.go)
  [src/driveradapter/api/httphandler/agentconfighandler/create.go](../../src/driveradapter/api/httphandler/agentconfighandler/create.go)

说明：

1. 原来的中心化根级文件 `swagger_stubs_generated.go` 已移除。
2. 原来的模块级 `swagger_doc_generated.go` 也已全部下沉并移除，真实 handler 注释已经成为主要来源。
3. `GET /v1/app/{app_key}`、`POST /v1/file/check`、`POST /v3/agent/batch-check-index-status` 已确认废弃，不再保留占位文档。

## 4. 整体生成流程

完整链路如下：

1. 在真实 handler 方法里编写 `swaggo` 注释。
2. 运行 `make gen-swag`。
3. `swag` 生成 Swagger 2.0（旧版 API 规范格式）的 `docs/swagger.json` / `docs/swagger.yaml`。
4. 运行 `go run ./cmd/openapi-docs generate` 或 `make gen-api-docs`。
5. 生成器读取 Swagger 2.0（旧版 API 规范格式）文档。
6. 用 `kin-openapi/openapi2conv` 转成 OpenAPI 3。
7. 统一把路径前缀改写为 `/api/agent-factory/...`。
8. 合并 [docs/api/overlay.yaml](./overlay.yaml)。
9. 使用基线文档做 compare 和缺失字段兜底。
10. 归一化安全定义与 operation metadata。
11. 输出最终 `agent-factory.json` / `agent-factory.yaml` / `agent-factory.html`。
12. 运行时由 `/swagger/index.html` 使用 Scalar 渲染这份规范。

## 5. 常用命令

### 5.1 只生成 Swagger 2.0（旧版 API 规范格式）中间产物

```bash
make gen-swag
```

### 5.2 生成最终 OpenAPI 3.0.2 + Scalar 静态页

```bash
make gen-api-docs
```

等价命令：

```bash
go run ./cmd/openapi-docs generate
```

### 5.3 与当前比较基线对比

```bash
make compare-api-docs
```

等价命令：

```bash
go run ./cmd/openapi-docs compare
```

### 5.4 校验最终文档

```bash
make validate-api-docs
```

等价命令：

```bash
go run ./cmd/openapi-docs validate
```

### 5.5 启动后查看页面

```bash
make localRun
```

或使用你的本地启动方式。启动后可访问：

1. `http://127.0.0.1:13020/swagger/index.html`
2. `http://127.0.0.1:13020/scalar`
3. `http://127.0.0.1:13020/swagger/doc.json`
4. `http://127.0.0.1:13020/swagger/doc.yaml`

## 6. 如何新增或修改一个接口文档

### 6.1 优先策略

新增公开接口时，优先顺序如下：

1. 直接把 Swagger 注释写在真实 handler 方法上。
2. 已废弃且不再暴露的 route，不要再用占位文档保留。
3. 只有代码注释无法表达的部分，才放到 overlay 中。

### 6.2 最小注释模板

```go
// HandlerName xxx
// @Summary      简短摘要
// @Description  较完整说明
// @Tags         标签1,标签2
// @Accept       json
// @Produce      json
// @Param        id       path   string  true  "资源ID"
// @Param        request  body   dto.Req true  "请求体"
// @Success      200      {object} dto.Res "成功"
// @Failure      400      {object} swagger.APIError "请求参数错误"
// @Failure      500      {object} swagger.APIError "服务器内部错误"
// @Router       /v3/example/{id} [post]
// @Security     BearerAuth
func (h *exampleHandler) Example(c *gin.Context) {
    ...
}
```

### 6.3 编写注释后的验证步骤

```bash
make gen-api-docs
make compare-api-docs
make validate-api-docs
```

如果只是想快速看差异：

```bash
sed -n '1,220p' test_out/openapi_compare_report.md
```

### 6.4 什么时候改 overlay

以下内容优先放在 overlay，而不是强塞进 `swaggo` 注释里：

1. 全局 `servers`
2. 全局 `securitySchemes`
3. `x-tagGroups`
4. 公共 Header 说明
5. 相关文档链接
6. 少量复杂示例
7. 需要对某个 operation 做 `security: []` 之类的覆盖

Overlay 文件位置：

- [docs/api/overlay.yaml](./overlay.yaml)

## 7. 当前文档策略

### 7.1 为什么不再保留遗留占位文档

当前已经不再保留模块级 `swagger_doc_generated.go`，也不再保留 `legacy_doc_handlers.go`。

原因是：

1. `GET /v1/app/{app_key}`、`POST /v1/file/check`、`POST /v3/agent/batch-check-index-status` 已确认没有真实路由注册，也没有业务实现入口。
2. 继续把它们保留在文档中，会让公开规范和实际可调用接口发生漂移。

因此当前策略是：

1. 真实存在的接口，文档直接挂在真实 handler 方法上。
2. 已废弃 route 从代码、文档和 compare 基线里一起移除。

### 7.2 为什么最终安全定义是 `ApiKeyAuth`

代码注释里很多地方仍写的是 `@Security BearerAuth`，这是为了兼容 `swaggo` 的输入方式。

后处理阶段会做两件事：

1. 把 `BearerAuth` 统一归一为最终文档里的 `ApiKeyAuth` 语义。
2. 去掉和根级默认安全要求完全重复的 operation 级 `security`，避免 compare 报告产生大量噪音。

所以：

1. 代码层面可以继续用 `BearerAuth`。
2. 最终对外文档层面会表现为 `ApiKeyAuth`，并与历史基线一致。

### 7.3 为什么最终路径都是 `/api/agent-factory/...`

因为实际路由注册里很多模块只声明了 `/v1/...`、`/v3/...` 的相对路径，而历史对外文档是以 `/api/agent-factory/...` 为统一前缀暴露的。

因此生成器在后处理时会统一补前缀，避免：

1. 运行时路由入口和文档路径不一致。
2. 与旧手写文档路径不一致。
3. Scalar Try it out 默认地址不正确。

## 8. Scalar 页面说明

### 8.1 运行时页面

运行时页面由 [router_swagger.go](../../src/infra/server/httpserver/router_swagger.go) 提供。

特点：

1. 使用 Scalar，而不是 Swagger UI。
2. `Try it out` 读取 `/swagger/doc.json`。
3. 返回文档时动态回填 `servers`，自动指向当前请求 host。

### 8.2 静态页面

静态页面由 [docs/api/agent-factory.html](./agent-factory.html) 提供。

特点：

1. 使用内嵌 JSON 的方式生成。
2. 不依赖运行时 `/swagger/doc.json`。
3. 可用于离线预览或静态分发。

## 9. 常见排查方式

### 9.1 `compare` 报告出现 path / operation 缺失

先检查：

1. 对应 handler 或模块文档函数是否有 `@Router`。
2. `@Router` 的 path 与 method 是否正确。
3. `swag init` 是否真的扫到了该文件。

建议命令：

```bash
rg -n "@Router" src/driveradapter/api/httphandler
```

### 9.2 报告出现 tag 不一致

先检查：

1. `@Tags` 是否和基线一致。
2. 是否少了 `模板`、`对话-internal`、`发布相关-internal`、`agent-internal` 等次级标签。
3. 是否有 overlay 对该 operation 做过覆盖。

### 9.3 报告出现 security 不一致

先检查：

1. 该接口是需要鉴权还是公开接口。
2. 公开接口是否在 overlay 中显式写了 `security: []`。
3. 代码注释是否写了 `@Security BearerAuth`。

### 9.4 报告出现 response code 不一致

先检查：

1. handler 注释中的 `@Failure` 是否补齐。
2. 是否应该保留基线里的 `401` / `403`。
3. 某些接口虽然业务代码可能返回 `404`，但历史基线未展示，也要按基线来对齐。

### 9.5 Scalar 页面能打开，但 Try it out 调错地址

先检查：

1. 请求是否走了反向代理。
2. 代理是否正确传递 `X-Forwarded-Proto` / `X-Forwarded-Host`。
3. `/swagger/doc.json` 返回的 `servers[0].url` 是否正确。

## 10. 推荐维护流程

每次新增/修改接口文档时，建议按下面流程走：

1. 修改真实 handler 注释或模块级文档函数。
2. 如需全局元数据或公开接口覆盖，修改 overlay。
3. 运行 `make gen-api-docs`。
4. 运行 `make compare-api-docs`。
5. 打开 compare 报告确认差异是否符合预期。
6. 运行 `make validate-api-docs`。
7. 启动服务后打开 `/swagger/index.html` 人工确认 Scalar 页面效果。

## 11. 后续可继续优化的方向

### 11.1 继续保持文档与真实接口一致

现在文档源已经收敛到真实 handler。后续继续维护时，重点是：

1. 新增接口时直接在真实 handler 上补注释。
2. 废弃接口时同步从路由、生成文档和 baseline 中一起移除。
3. 避免再次引入“代码里没有、文档里还在”的占位 route。

### 11.2 提升 schema 质量

当前已经做到结构、路径、标签、鉴权、响应码和基线一致，但后续还可以继续增强：

1. DTO 字段级 `example`
2. `enum`
3. `default`
4. 更细的错误模型
5. 更准确的响应 schema，而不是少量 `object` 兜底

### 11.3 减少对基线 fallback 的依赖

理想状态是：

1. `compare` 报告始终清零。
2. `generate` 时即使关闭 baseline fallback，最终文档也不会退化。

这代表代码注释和 overlay 已经足够完整。

## 12. 一句话工作原则

可以把这套方案理解成下面这条规则：

1. 接口结构尽量写回代码。
2. 全局展示能力放到 overlay。
3. 最终结果必须经过 compare 和 validate。
4. 任何变更都以最终生成结果与基线一致为准。
