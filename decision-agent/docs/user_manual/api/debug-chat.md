# Debug 对话

Debug 对话用于 Agent 配置页调试和排障，适合在 Agent 尚未正式对外使用前验证输入变量、知识网络、工具调用和输出结构。普通业务调用优先使用 `chat/completion`；只有需要排查配置或运行链路时再使用 debug 接口。

## 接口

```text
POST /api/agent-factory/v1/app/{agent_key}/debug/completion
```

## 与普通对话的区别

| 对比项 | 普通对话 | Debug 对话 |
| --- | --- | --- |
| 接口 | `/chat/completion` | `/debug/completion` |
| 调用类型 | Decision Agent 后端设置为 `chat` 或 `api_chat` | Decision Agent 后端设置为 `debug_chat` |
| 请求字段 | `query`、`history`、`custom_querys` 在请求体顶层 | `query`、`history`、`custom_querys` 放在 `input` 内 |
| Executor 路由 | `/api/agent-executor/v2/agent/run` | `/api/agent-executor/v2/agent/debug` |
| Executor 运行参数 | `is_debug_run=false` | `is_debug_run=true` |
| 缓存行为 | 可通过 `chat_option.enable_dependency_cache` 启用依赖缓存 | Debug 执行会跳过 dependency cache 初始化 |

Decision Agent 后端的 debug handler 会把 `input.query`、`input.history`、`input.custom_querys` 转换成内部 `ChatReq`，再将 `CallType` 设置为 `debug_chat`。Agent Executor 侧 debug 路由仍复用同一套 `run_agent` 主流程，但传入 `is_debug_run=True`，因此适合验证配置，不适合作为正式业务调用入口。

## 非流式 Debug

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "测试 debug completion" \
  '{
  agent_id: $agent_id,
  agent_version: $agent_version,
  input: {
    query: $query
  },
  stream: false
}' > /tmp/debug-chat-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-chat-request.json | jq .
```

## 普通流式 Debug

普通流式返回的是 SSE，每个 `data:` 事件都是当前完整响应快照。

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "用 debug 模式流式回答" \
  '{
  agent_id: $agent_id,
  agent_version: $agent_version,
  input: {
    query: $query
  },
  stream: true,
  inc_stream: false
}' > /tmp/debug-chat-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-chat-stream-request.json
```

## 增量流式 Debug

增量流式返回的是 diff 事件，每个 `data:` 事件形如 `{seq_id, key, content, action}`。具体合并算法与事件处理方式见 [增量流式](./incremental-streaming.md)。

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "用 debug 模式增量流式回答" \
  '{
  agent_id: $agent_id,
  agent_version: $agent_version,
  input: {
    query: $query
  },
  stream: true,
  inc_stream: true
}' > /tmp/debug-chat-inc-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-chat-inc-stream-request.json
```

## 使用建议

- Debug 接口主要面向配置验证和排障，不建议作为线上用户对话入口。
- Debug 请求体的对话输入放在 `input` 内；普通对话请求体则直接使用顶层 `query`。
- Debug 执行会跳过 dependency cache。需要验证缓存效果时，请使用普通对话并配置 `chat_option.enable_dependency_cache`。
