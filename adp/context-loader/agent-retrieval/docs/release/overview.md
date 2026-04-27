# Context Loader 交付说明

本文档是 `docs/release/` 的唯一入口。版本号以 `agent-retrieval/VERSION` 为准。

## 1. 本版交付内容

当前版本：`0.7.0`

### 持续交付能力

本版继续交付 Context Loader 的核心能力：

- MCP Server 接入能力
- 探索发现、精确查询、候选资源发现、逻辑属性解析、动态工具发现
- 索引构建相关能力

### 本版新增 / 变更

本版为 `0.7.0`，仅发布以下两个需求：

- `issue-189`：Schema 探索入口统一收口为 `search_schema`
- `issue-234`：`search_schema` 新增 `metric_types` 返回桶，并支持 `include_metric_types`

围绕上述两个需求，本版交付变化如下：

- MCP 工具面移除 `kn_search` / `kn_schema_search`，仅暴露 `search_schema`
- 新增标准 HTTP 接口 `POST /api/agent-retrieval/in/v1/kn/search_schema`
- HTTP `search_schema` 统一通过 request body 传递 `kn_id`
- `POST /api/agent-retrieval/in/v1/kn/kn_search` 保留为兼容接口，继续兼容旧请求写法，但与 `search_schema` 共用收敛后的 Schema-only logic
- `POST /api/agent-retrieval/in/v1/kn/semantic-search` 保留为 legacy 接口，维持历史输出形态

## 2. 交付清单

本版交付内容建议采用以下目录结构：

```text
docs/release/
├── overview.md                # 本版本说明文档（唯一入口）
├── toolset/                   # Context Loader 工具集快照（ADP 形式交付）
├── tool-deps/                 # Context Loader 依赖的其他工具集快照（按需导入）
├── agent-deps/                # 必选依赖：逻辑属性解析等能力依赖的 Agent
└── agent-recall-examples/     # 可选示例：用于接入参考与回放验证
```

说明如下：

- `toolset/`：本版 Context Loader 工具集快照，接入时优先导入。
- `tool-deps/`：本版依赖的其他工具集快照；如当前场景不涉及相关能力，可按需导入。
- `agent-deps/`：本版必选的 Agent 依赖；逻辑属性解析能力依赖此处的 Agent。
- `agent-recall-examples/`：本版回放样例，用于接入参考与最小验证。

## 3. 接入与使用说明

### 导入顺序

1. 导入 `toolset/`
2. 按需导入 `tool-deps/`
3. 导入 `agent-deps/`
4. 使用 `agent-recall-examples/` 做回放验证

### 本版变化与注意事项

- `search_schema` 是当前版本唯一的 MCP Schema 探索工具；`kn_search` / `kn_schema_search` 不再出现在 MCP 工具列表中
- HTTP `search_schema` 通过 request body 传递 `kn_id`；不要再使用 `x-kn-id` Header
- `kn_search` 继续作为兼容 HTTP 接口保留，但输出固定为 Schema-only 结果
- `kn_schema_search`（`/api/agent-retrieval/in/v1/kn/semantic-search`）继续作为 legacy 接口保留，维持历史 `concepts[]` 输出形态
- `search_schema` 当前版本新增 `metric_types` 返回桶，可通过 `search_scope.include_metric_types` 控制是否返回
- 其他已交付能力继续沿用；完整工具说明见 `docs/release/tool-usage-guide.md`

## 4. 最小验证

- 交付清单验证：`docs/release/` 下的实际内容与本文档描述一致
- 导入验证：`toolset/`、`tool-deps/`、`agent-deps/` 无缺失依赖
- Schema Search 验证：`search_schema` 作为唯一 MCP Schema 探索工具可用，`kn_search` / `kn_schema_search` 不再出现在 MCP 工具列表中
- HTTP 契约验证：`search_schema` 通过 request body 传递 `kn_id`，契约中不再出现 `x-kn-id`
- Metric Schema 验证：`search_schema` 返回 `metric_types`，并支持 `search_scope.include_metric_types`
- 分层验证：标准接口 `search_schema`、兼容接口 `kn_search`、legacy 接口 `semantic-search` 的文档叙事保持一致
