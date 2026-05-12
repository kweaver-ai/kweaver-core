# 变更日志

英文版本：[CHANGELOG.md](CHANGELOG.md)

## 0.8.0

### 功能与改进

- 为 `search_schema` 增加按概念分组限定 Schema 探索范围的能力
  - 支持通过 `search_scope.concept_groups` 将对象类、关系类和动作类 Schema 探索限定在指定 BKN 概念分组内
  - 未传概念分组时，现有 `search_schema` 行为保持不变
  - 分组范围内返回关系类和动作类时，会一并补齐其引用的对象类，让调用方拿到完整 Schema 上下文
  - 说明：指标类 Schema 请求会携带同一组概念分组范围，但是否真正按组过滤取决于 BKN metrics 侧支持
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/304
- 将 ContextLoader 标准工具集内置到服务启动流程
  - ContextLoader 启动时自动同步内置工具集到执行工厂
  - 工具集契约描述统一为 `ContextLoader 标准内置工具集；契约版本: 0.8.0`
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/306

### 兼容性

- 兼容 HTTP 路径和 legacy HTTP 路径保持不变；需要概念分组能力的新接入方应使用 `search_schema`

### 文档

- 更新 API、MCP schema、toolset 和发布文档，说明按概念分组限定 Schema 探索范围的使用方式
- 补充 ContextLoader 内置工具集交付方式和契约版本描述规则

## 0.7.0

### 功能与改进

- 新增 `search_schema`，作为 MCP 和 HTTP 调用方的标准 Schema 探索入口
  - 通过一个接口支持对象类、关系类、动作类和指标类 Schema 探索
  - HTTP `search_schema` API 使用请求体中的 `kn_id`
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/189
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/234
- MCP Schema 探索能力统一收敛到 `search_schema`，减少 Agent 选择工具时的歧义

### 兼容性

- `kn_search` 继续作为兼容 HTTP 路径保留，`semantic-search` 继续作为 legacy HTTP 路径保留

### 文档

- 更新 release overview 和工具使用文档，说明 `search_schema` 入口统一与 metric schema 召回契约
