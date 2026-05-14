# 对话、会话与执行

Decision Agent 中需要区分三个概念：`conversation` 对应“对话”，`session` 对应“会话”，`run` 对应“一次执行”。普通 API 对接通常只需要关心 `conversation_id`；`session` 更多用于缓存、续连和运行期状态，暂时用不到时可以先忽略。

## 概念关系

| 概念 | 常见字段 | 中文名称 | 作用 |
| --- | --- | --- | --- |
| `conversation` | `conversation_id` | 对话 | 用户侧的一段连续对话上下文，用于继续对话、查看消息、标记已读等。 |
| `session` | `conversation_session_id` | 会话 | 某段运行期会话状态，当前主要服务于缓存、续连和运行态管理。 |
| `run` | `agent_run_id` | 一次执行 | 一次 Agent 调用或接口执行，用于运行日志、trace、终止和排障。 |

关系上，一个对话可以对应多个会话，一个会话可以包含一个或多个 run。服务端在对话时会基于 `conversation_id` 和 session `start_time` 生成 `conversation_session_id`，格式通常类似：

```text
{conversation_id}-{start_time}
```

## 对话续连

Decision Agent 后端在流式对话过程中会维护运行中内存 `SessionMap`，按 `conversation_id` 暂存响应快照、停止信号和续连状态。客户端断开后，如果服务端仍保留该对话的运行态，可以通过 resume 接口继续读取流式响应。

这里的 resume 是“断线续连”，不是人工干预后的中断恢复。人工干预恢复需要再次调用 `chat/completion` 或 `debug/completion`，并传入 `resume_interrupt_info`。完整说明见 [人工干预与终止](./intervention-termination.md)。

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/resume
```

```bash
jq -n --arg conversation_id "$CONVERSATION_ID" '{
  conversation_id: $conversation_id
}' > /tmp/chat-resume-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/resume" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-resume-request.json
```

`SessionMap` 是内存态结构，主要用于当前运行中的流式响应和断线续连；一次执行结束后，服务端会清理对应的 `conversation_id`。

## Redis 会话

除了运行中内存态，Decision Agent 后端还会在 Redis 中维护 conversation session，key 前缀为：

```text
agent-app:conversation-session:
```

默认 TTL 为 600 秒。会话管理接口可用于获取会话信息、创建会话或恢复会话生命周期：

```text
PUT /api/agent-factory/v1/conversation/session/{conversation_id}
```

获取或创建：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-latest}" \
  '{
    action: "get_info_or_create",
    agent_id: $agent_id,
    agent_version: $agent_version
  }' > /tmp/conversation-session.json

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/conversation/session/$CONVERSATION_ID" \
  -H "Content-Type: application/json" \
  -d @/tmp/conversation-session.json | jq .
```

恢复生命周期或创建：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-latest}" \
  '{
    action: "recover_lifetime_or_create",
    agent_id: $agent_id,
    agent_version: $agent_version
  }' > /tmp/conversation-session-recover.json

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/conversation/session/$CONVERSATION_ID" \
  -H "Content-Type: application/json" \
  -d @/tmp/conversation-session-recover.json | jq .
```

响应中会包含 `conversation_session_id` 和 `ttl`：

```json
{
  "conversation_session_id": "conversation-id-1710000000",
  "ttl": 600
}
```

普通对接场景里，继续对话优先使用 `conversation_id`。会话管理接口属于缓存和运行期优化能力，只有需要控制会话生命周期或对接断线续连时再使用。

## 缓存行为

缓存相关逻辑分两层：

| 位置 | 说明 |
| --- | --- |
| Decision Agent 会话管理 | 管理 Redis conversation session；部分入口可触发 Executor agent cache upsert。 |
| Agent Executor dependency cache | 普通执行可通过 `chat_option.enable_dependency_cache` 启用；默认 TTL 为 60 秒，超过 10 秒阈值后可触发缓存更新。 |

普通对话示例：

```bash
jq -n --arg query "使用缓存回答这个问题" '{
  query: $query,
  stream: false,
  chat_option: {
    enable_dependency_cache: true
  }
}' > /tmp/chat-cache-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-cache-request.json | jq .
```

Debug 执行会跳过 dependency cache 初始化。如果要验证缓存命中和刷新，请使用普通对话接口。
