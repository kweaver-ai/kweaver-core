# MCP 工具 Schema 配置

工具元信息与 JSON Schema 统一在此目录维护，便于扩展与 LLM 理解。

## 文件规范

| 文件 | 说明 |
|------|------|
| `tools_meta.json` | 工具元信息（name、description），新增工具在此添加条目 |
| `{tool_key}.json` | 工具 Schema（含 `input_schema` 与 `output_schema` 两个键） |

## 新增工具步骤

1. 在 `tools_meta.json` 中添加 `{tool_key}: { "name": "...", "description": "..." }`
2. 添加 `{tool_key}.json`，包含 `input_schema` 与 `output_schema` 两个 JSON Schema 对象
3. 在 `app.go` 中注册工具（`loadToolMeta` 与 `loadToolSchemas` 已统一封装，无需修改 `schemas.go`）

## 描述参考

描述建议参考 `docs/releases/v5.0.4/tool-usage-guide.md` 中的「工具总览」与「工具参考」章节。
