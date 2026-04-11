# Context Loader 交付说明

本文档是 `docs/release/` 的唯一入口。版本号以 `agent-retrieval/VERSION` 为准。

## 1. 本版交付内容

当前版本：`0.6.0`

### 持续交付能力

本版继续交付 Context Loader 的核心能力：

- MCP Server 接入能力
- 探索发现、精确查询、候选资源发现、逻辑属性解析、动态工具发现
- 索引构建相关能力

### 本版新增 / 变更

本版为 `0.6.0`，新增 Skill 候选发现能力，并延续 TOON 压缩响应能力：

- 新增 `find_skills`，用于在指定 `kn_id` 和业务上下文边界内发现 Skill 候选
- 工具体系补充 `find_*` 语义，用于承接“有边界的候选资源发现”场景
- `find_skills` 支持网络级、对象类级、实例级三种召回模式，返回最小化 Skill 元数据（`skill_id`、`name`、`description`）

- 新增 `response_format`，支持 `json` 和 `toon`
- HTTP 默认仍返回 `json`
- MCP Tool 默认返回 `toon` 文本内容
- 错误响应继续保持 JSON
- 统一了 Context Loader 相关工具中的 `x-account-id`、`x-account-type` 参数口径，便于调用方接入

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

- `find_skills` 不是升级 Context Loader 后自动可用的能力；只有知识网络已完成 Skill 承接和绑定后，Skill recall 才能返回稳定结果。快速启用方式与前提说明见 `docs/release/tool-usage-guide.md` 的 `6.5 find_skills`
- `find_skills` 属于候选资源发现工具，不替代 `kn_search` / `query_object_instance`
- 若调用方尚未明确对象类或实例，建议先使用 `kn_schema_search` / `kn_search` 与 `query_*` 工具获取上下文，再调用 `find_skills`
- `find_skills` 的 `skill_query` 仅用于当前业务边界内过滤和排序，不用于跨知识网络搜索
- HTTP 接口可通过 Query 参数 `response_format` 指定响应格式，默认 `json`
- MCP Tool 可通过 argument `response_format` 指定响应格式，默认 `toon`
- 当使用 `response_format=toon` 时，成功响应返回 TOON；具体接口说明以 `docs/apis/api_private/` 为准
- 如果接入方不需要 TOON 压缩能力，保持现有默认调用方式即可。

## 4. 最小验证

- 交付清单验证：`docs/release/` 下的实际内容与本文档描述一致
- 导入验证：`toolset/`、`tool-deps/`、`agent-deps/` 无缺失依赖
- `find_skills` 验证：至少验证一种召回模式（网络级 / 对象类级 / 实例级），确认返回结构为候选 Skill 列表或空列表
- HTTP 验证：任选一个支持 `response_format` 的接口，确认默认返回 JSON，指定 `toon` 时返回 TOON
- MCP 验证：任选一个 tool，确认默认返回 TOON 文本，显式指定 `json` 时返回 JSON 文本
