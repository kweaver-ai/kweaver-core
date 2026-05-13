# 对话 API

对话接口使用 `agent_key`，而不是直接使用 `agent_id`。如果手上只有 `agent_id`，先通过广场详情接口解析。

## 解析 agent_key

```bash
test -n "$AGENT_ID"
export AGENT_VERSION="${AGENT_VERSION:-v0}"

af_curl "$AF_BASE_URL/api/agent-factory/v3/agent-market/agent/$AGENT_ID/version/$AGENT_VERSION?is_visit=true" \
  | tee /tmp/agent-market-detail.json

export AGENT_KEY="$(jq -r '.key' /tmp/agent-market-detail.json)"
test -n "$AGENT_KEY"
```

## 非流式对话

接口：

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/completion
```

示例：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "请用一句话介绍你自己" \
  '{
  agent_id: $agent_id,
  agent_key: $agent_key,
  agent_version: $agent_version,
  query: $query,
  stream: false
}' > /tmp/chat-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-request.json | tee /tmp/chat-response.json

export CONVERSATION_ID="$(jq -r '.. | objects | .conversation_id? // empty' /tmp/chat-response.json | head -n 1)"
echo "$CONVERSATION_ID"
```

## 普通流式对话

普通流式对话返回 SSE。每个 `data:` 片段都是当前完整响应快照，适合希望实时展示但不想维护增量合并逻辑的客户端：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "请流式介绍你的能力" \
  '{
  agent_id: $agent_id,
  agent_key: $agent_key,
  agent_version: $agent_version,
  query: $query,
  stream: true,
  inc_stream: false
}' > /tmp/chat-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d @/tmp/chat-stream-request.json
```

增量流式对话请阅读 [增量流式](./incremental-streaming.md)。

## 继续对话

```bash
test -n "$CONVERSATION_ID"

jq -n \
  --arg query "请继续补充一个使用建议" \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg conversation_id "$CONVERSATION_ID" \
  '{
    agent_id: $agent_id,
    agent_key: $agent_key,
    agent_version: $agent_version,
    query: $query,
    conversation_id: $conversation_id,
    stream: false
  }' > /tmp/chat-followup-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-followup-request.json | jq .
```

## 对话列表

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation?page=1&size=10" | jq .
```

## 对话详情

对话详情中会包含该对话下的消息列表。

```bash
test -n "$CONVERSATION_ID"

af_curl "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID" | jq .
```

## 标记已读

```bash
test -n "$CONVERSATION_ID"

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID/mark_read" \
  -H "Content-Type: application/json" \
  -d '{"latest_read_index": 2}' | jq .
```

Debug 对话、运行期会话和缓存说明请阅读 [Debug 对话](./debug-chat.md) 与 [对话、会话与执行](./conversation-session-run.md)。
