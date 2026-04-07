# `internal/openapidoc` 目录说明

## 快速入口

- 精简版说明：查看 [README.simple.md](./README.simple.md)
- 完整版说明：继续阅读当前文件

如果你只是想快速知道这个目录做什么、主入口在哪里、排查问题该先看哪个文件，优先看精简版就够了。

## 目录职责

`internal/openapidoc` 是 OpenAPI 文档构建核心库。

它不直接处理命令行，而是提供一套可复用的文档构建能力，主要负责：

1. 把 Swagger 2.0 输入转成 OpenAPI 3。
2. 给路径补统一前缀。
3. 合并 overlay。
4. 按需从 baseline 补齐缺失内容。
5. 做路径参数、operationId、安全配置等归一化。
6. 生成 compare report。
7. 生成 JSON / YAML / 静态 Scalar HTML。
8. 校验最终文档是否合法。

从职责上看，这个目录就是“文档构建引擎”。

## 从哪里开始看

### 对外主入口

优先从 `build.go` 里的 `BuildArtifactsFromFiles` 看起。

原因：

- 它把整个构建流程串起来了。
- 读完这个方法，就能知道各个模块在什么阶段介入。
- 其他文件基本都可以视为它调度的子模块。

### 核心阅读顺序

推荐阅读顺序：

1. `doc.go`
2. `build.go`
3. `convert.go`
4. `merge.go`
5. `normalize.go`
6. `document.go`
7. `compare.go`
8. `render.go`
9. `sanitize.go`
10. `operations.go`

### 如果你要排查某类问题

#### 文档生成结果不对

先看：

1. `build.go`
2. `convert.go`
3. `merge.go`
4. `normalize.go`

#### 某个接口在 compare report 里异常

先看：

1. `compare.go`
2. `operations.go`

#### 生成的 HTML 页面有问题

先看：

1. `render.go`
2. `build.go`

#### OpenAPI 校验失败

先看：

1. `normalize.go`
2. `sanitize.go`
3. `document.go`

## 核心主流程

完整生成链路大致如下：

```text
BuildArtifactsFromFiles
  -> 读取 Swagger 输入
  -> ConvertSwagger2ToOpenAPI3
  -> RewriteAgentFactoryPaths
  -> MergeOverlay
  -> NormalizeSecurity
  -> CloneOpenAPIDoc
  -> LoadOpenAPIDocFile(baseline)
  -> BuildComparisonReport
  -> MergeMissingFromBaseline(可选)
  -> NormalizePathParameters
  -> NormalizeOperationIDs
  -> NormalizeSecurity
  -> ValidateOpenAPI
  -> MarshalOpenAPIJSON / MarshalOpenAPIYAML
  -> RenderScalarStaticHTML
```

也就是说，这个目录内部的真实主入口不是某个单文件模板，而是 `BuildArtifactsFromFiles` 这个 orchestration 方法。

## 文件说明

### `doc.go`

这是目录级基础定义文件，主要放常量和跨文件共用的数据结构。

关键内容：

- `agentFactoryBasePath`
  - 统一的接口路径前缀 `/api/agent-factory`
- `scalarCDNURL`
  - 静态 HTML 所依赖的 Scalar CDN 地址
- `apiKeySecurityName` / `bearerSecurityName`
  - 安全方案归一化时使用的名称常量
- `BuildOptions`
  - 构建请求参数
- `BuildArtifacts`
  - 构建输出结果

阅读价值：

- 先看这里，可以快速知道整个模块的输入输出是什么。

### `build.go`

这是整个目录最重要的流程编排文件。

核心函数：

- `BuildArtifactsFromFiles`
  - 读取 Swagger 输入
  - 调用 `ConvertSwagger2ToOpenAPI3`
  - 重写接口路径
  - 合并 overlay
  - 归一化安全定义
  - 克隆生成文档作为最终文档
  - 读取 baseline 并生成 compare report
  - 根据开关决定是否做 baseline fallback
  - 做路径参数、operationId、安全定义的最终归一化
  - 做 OpenAPI 校验
  - 生成 JSON / YAML / HTML 字节流

如果只看一个函数来理解整个包，请看这个函数。

### `convert.go`

这个文件专门负责“格式转换”和“路径前缀修正”。

核心函数：

- `ConvertSwagger2ToOpenAPI3`
  - 把 Swagger 2.0 JSON 转成 OpenAPI 3
  - 补齐 `OpenAPI` 版本号
  - 确保 `Paths` 和 `Components` 不为 nil
- `RewriteAgentFactoryPaths`
  - 遍历所有路径并统一加上 `/api/agent-factory`
- `prefixAgentFactoryPath`
  - 计算单条路径应该如何补前缀

典型问题：

- “为什么路径多了 `/api/agent-factory`”
- “为什么旧 swagger 也能产出 OpenAPI 3”

### `document.go`

这个文件负责“文档对象和文件内容之间的转换”。

核心函数：

- `CloneOpenAPIDoc`
  - 深拷贝 OpenAPI 文档
- `LoadOpenAPIDocFile`
  - 从文件加载 JSON / YAML 文档
- `MarshalOpenAPIJSON`
  - 生成格式化 JSON
- `MarshalOpenAPIYAML`
  - 生成 YAML
- `WriteFile`
  - 带父目录创建的统一写文件方法
- `loadOpenAPIDoc`
  - 将原始字节流装配成文档对象
- `loadMap`
  - 先把输入解析成 map，并执行清洗
- `docToMap`
  - 文档对象转 map，便于深度合并
- `mapToDoc`
  - 标准化 map 重新转回文档对象

阅读价值：

- 这里是“文件 IO 层”和“对象层”之间的桥梁。

### `sanitize.go`

这个文件负责兼容和清洗历史 Swagger / OpenAPI 数据中的非标准结构。

核心函数：

- `sanitizeOpenAPIMap`
  - 从根节点开始清洗
- `sanitizeOpenAPIMapWithContext`
  - 按 schema/example 上下文递归修正数据
- `sanitizeOpenAPIValue`
  - 处理 map / slice 的递归遍历
- `isSchemaContextKey`
  - 判断子节点是否处在 schema 语境
- `normalizeSchemaType`
  - 修正历史类型名，例如 `int -> integer`
- `joinStringList`
  - 把数组形式的描述合并成字符串
- `sanitizeExampleForSchemaType`
  - 按 schema type 调整 example 的值类型

典型问题：

- “为什么这个历史字段读入后没有报错”
- “为什么某些非标准 schema 还能被最终文档接受”

### `merge.go`

这个文件负责两类合并逻辑：

1. overlay 覆盖合并
2. baseline 缺失补齐

核心函数：

- `MergeOverlay`
  - 用 overlay 显式覆盖当前文档
- `MergeMissingFromBaseline`
  - 只在字段缺失时，用 baseline 内容兜底
- `mergeOverwrite`
  - 递归覆盖 map 内容
- `fillMissing`
  - 递归补齐缺失字段
- `fillMissingValue`
  - 决定单个值该“保留现值”还是“用 baseline 补”
- `normalizeValue`
  - 把任意 map/slice 归一化成统一结构
- `cloneValue`
  - 深拷贝任意 map/slice 值
- `isZeroValue`
  - 判断值是否为空，用于补齐决策

典型问题：

- “为什么 overlay 能覆盖文案”
- “为什么 baseline 中老接口还能出现在最终文档里”

### `normalize.go`

这个文件负责构建末尾阶段的规范化和最终校验，是生成质量的关键文件。

核心函数：

- `ValidateOpenAPI`
  - 调用 kin-openapi 做最终校验
  - 允许部分额外 sibling 字段，兼容历史数据
- `NormalizePathParameters`
  - 确保 path 参数都声明为 `required`
  - 补齐路径里缺失的 path 参数定义
- `NormalizeOperationIDs`
  - 为重复的 `operationId` 自动生成唯一后缀
- `NormalizeSecurity`
  - 统一 `BearerAuth` / `ApiKeyAuth`
  - 去掉与全局定义重复的接口级安全声明
- `normalizeParameters`
  - 参数列表级辅助归一化
- `normalizePathParameterRef`
  - 单个 path 参数归一化
- `collectPathParameterNames`
  - 收集已声明的 path 参数名
- `extractPathParameterNames`
  - 从 URL 模板中提取参数名
- `normalizeSecurityRequirements`
  - 便于 compare 时压平安全声明
- `normalizeSecurityRequirementsRef`
  - 归一化 security requirement 本身
- `cloneSecurityScopes`
  - 显式把空 scope 表示为空数组
- `securityRequirementsEqual`
  - 比较两组安全声明是否等价
- `normalizeSecurityRequirementNames`
  - 提取并排序安全声明中的 scheme 名称

典型问题：

- “为什么某些 path 参数自动变成了 required”
- “为什么重复 operationId 不会冲突”
- “为什么 security scopes 现在是 `[]` 而不是 `null`”

### `operations.go`

这个文件提供路径数、接口数、方法集合、差集等基础统计能力。

核心函数：

- `CountPathsAndOperations`
  - 统计路径数和接口数
- `orderedOperations`
  - 按固定 HTTP 方法顺序返回接口
- `pathItemOperations`
  - 提取单条 path 的 HTTP 方法列表
- `sanitizePathForOperationID`
  - 把路径转成适合拼进 operationId 的字符串
- `pathOperationSet`
  - 把文档转成 path -> methods 映射
- `diffOperations`
  - 计算缺失的“方法 + 路径”组合
- `operationResponseCodes`
  - 提取响应状态码集合
- `sortedMapKeys`
  - 稳定排序 map key
- `difference`
  - 计算差集
- `writeMarkdownList`
  - 输出 Markdown 列表

阅读价值：

- 这里是 compare / validate / 统计输出的基础工具层。

### `compare.go`

这个文件负责把 generated 文档和 baseline 文档的差异组织成可阅读报告。

核心函数：

- `BuildComparisonReport`
  - 生成整份 Markdown 差异报告
- `diffOperationSummaries`
  - 比较 summary
- `diffOperationTags`
  - 比较 tags
- `diffOperationSecurity`
  - 比较 security
- `diffOperationRequestBodies`
  - 比较 request body 是否存在
- `diffOperationResponseCodes`
  - 比较响应状态码集合
- `diffOperationField`
  - 差异遍历模板方法
- `normalizeStringSlice`
  - 字符串切片归一化
- `stringSlicesEqual`
  - 归一化切片比较

阅读价值：

- 如果 compare report 中某条差异看不懂，基本都要回到这个文件。

### `render.go`

这个文件负责生成静态 HTML 版本的 API 文档页面。

核心函数：

- `RenderScalarStaticHTML`
  - 把 OpenAPI JSON 内嵌到 HTML 中
  - 注入 Scalar 运行所需脚本
  - 配置标题、favicon、基础页面骨架

阅读价值：

- 如果你在排查页面标题、favicon、静态文档打开方式，这里是入口。

## 测试文件说明

### `compare_test.go`

验证 compare report 是否包含数量统计、缺失操作和语义差异。

### `convert_test.go`

验证 Swagger 2.0 转 OpenAPI 3 和路径前缀改写是否正确。

### `merge_test.go`

验证 overlay 覆盖合并和 baseline fallback 补齐逻辑。

### `normalize_path_test.go`

验证路径参数补齐、path 参数 required 处理以及 operationId 去重。

### `security_test.go`

验证安全定义归一化、public override 处理，以及空 scopes 是否序列化为 `[]`。

### `sanitize_test.go`

验证历史 schema 字段和 example 的清洗兼容逻辑。

### `render_test.go`

验证生成的静态 HTML 是否包含 Scalar 所需标记、标题和 favicon。

### `validate_test.go`

验证非法文档会被 `ValidateOpenAPI` 拒绝。

### `test_helpers_test.go`

测试辅助文件，提供测试文档加载等共享能力。

## 常见排查路径

### 路径不对

先看：

- `convert.go`
- `normalize.go`

### 某个接口缺了

先看：

- `build.go`
- `merge.go`
- `compare.go`

### 文档生成了，但 HTML 页面不对

先看：

- `render.go`
- `build.go`

### 文档校验失败

先看：

- `sanitize.go`
- `normalize.go`
- `document.go`

## 一句话总结

如果只记住一个内部入口，请记住：

- `BuildArtifactsFromFiles` 是整个目录的真正主入口。

它之下再按模块去看：

- 转换看 `convert.go`
- 合并看 `merge.go`
- 归一化和校验看 `normalize.go`
- 输出看 `document.go` 和 `render.go`
- 差异分析看 `compare.go`
