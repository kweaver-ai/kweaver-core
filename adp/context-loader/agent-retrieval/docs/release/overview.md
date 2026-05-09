# Context Loader 交付说明

本文档是 `docs/release/` 的唯一入口。版本号以 `agent-retrieval/VERSION` 为准。

## 1. 本版交付内容

当前版本：`0.8.0`

本版交付的是 Schema 探索阶段的“概念分组范围控制”能力。使用 `search_schema` 的 Agent / 应用，可以在请求中指定 BKN `concept_group`，让返回的 Schema 只围绕指定业务分组展开，降低跨业务范围的无关对象类、关系类和行动类干扰。

用户需要关注：

- 新增参数：`search_schema.search_scope.concept_groups`，传入 BKN 概念分组 ID 列表。
- 生效结果：`object_types`、`relation_types`、`action_types` 会按 BKN 概念分组语义收敛；不传或传空时保持原有 Schema 探索行为。
- 指标说明：`metric_types` 会透传同一组 `concept_groups`，实际过滤能力取决于 BKN metrics 侧实现。
- 兼容边界：本版不要求迁移 `/kn/kn_search` 和 `/kn/semantic-search` 接入；需要概念分组能力的新接入应使用 `search_schema`。

## 2. 交付清单

本版交付内容建议采用以下目录结构：

```text
docs/release/
├── overview.md                # 本版本说明文档（唯一入口）
├── toolset/                   # Context Loader 工具集快照（ADP 形式交付）
├── tool-deps/                 # Context Loader 依赖的其他工具集快照（镜像内固定携带，启动自动同步）
├── agent-deps/                # 必选依赖：逻辑属性解析等能力依赖的 Agent
└── agent-recall-examples/     # 可选示例：用于接入参考与回放验证
```

说明如下：

- `toolset/`：本版 Context Loader 工具集快照，接入时优先导入。
- `tool-deps/`：本版依赖的其他工具集快照；镜像启动后会自动同步到执行工厂，当前不需要人工单独导入。
- `agent-deps/`：本版必选的 Agent 依赖；逻辑属性解析能力依赖此处的 Agent。
- `agent-recall-examples/`：本版回放样例，用于接入参考与最小验证。

## 3. 接入与使用说明

### 导入顺序

1. 导入 `toolset/`
2. 等待启动阶段自动同步 `tool-deps/`
3. 导入 `agent-deps/`
4. 使用 `agent-recall-examples/` 做回放验证

### 本版变化与注意事项

- `search_schema` 是当前版本唯一的 MCP Schema 探索工具；`kn_search` / `kn_schema_search` 不再出现在 MCP 工具列表中
- HTTP `search_schema` 通过 request body 传递 `kn_id`；不要再使用 `x-kn-id` Header
- `kn_search` 继续作为兼容 HTTP 接口保留，但输出固定为 Schema-only 结果
- `kn_schema_search`（`/api/agent-retrieval/in/v1/kn/semantic-search`）继续作为 legacy 接口保留，维持历史 `concepts[]` 输出形态
- `search_schema` 继续支持 `metric_types` 返回桶，可通过 `search_scope.include_metric_types` 控制是否返回
- `search_schema` 当前版本新增 `search_scope.concept_groups`，用于按 BKN 概念分组限定对象类、关系类、行动类 Schema 召回；`metric_types` 会透传该字段，实际过滤依赖 BKN metrics 实现
- 其他已交付能力继续沿用；完整工具说明见 `docs/release/tool-usage-guide.md`

## 4. 最小验证

- 交付清单验证：`docs/release/` 下的实际内容与本文档描述一致
- 导入验证：`toolset/`、`agent-deps/` 无缺失依赖，`tool-deps/` 会在服务启动后自动同步
- Schema Search 验证：`search_schema` 作为唯一 MCP Schema 探索工具可用，`kn_search` / `kn_schema_search` 不再出现在 MCP 工具列表中
- HTTP 契约验证：`search_schema` 通过 request body 传递 `kn_id`，契约中不再出现 `x-kn-id`
- Metric Schema 验证：`search_schema` 返回 `metric_types`，并支持 `search_scope.include_metric_types`
- Concept Group 验证：`search_schema.search_scope.concept_groups` 可透传到 BKN 分组搜索，返回的关系类/行动类引用对象可被补齐
- 分层验证：标准接口 `search_schema` 是本次概念分组变更入口，兼容接口 `kn_search`、legacy 接口 `semantic-search` 保持历史契约
- `find_skills` 验证：至少验证一种召回模式（对象类级 / 实例级），确认返回结构为候选 Skill 列表或空列表
- HTTP 验证：任选一个支持 `response_format` 的接口，确认默认返回 JSON，指定 `toon` 时返回 TOON
- MCP 验证：任选一个 tool，确认默认返回 TOON 文本，显式指定 `json` 时返回 JSON 文本
