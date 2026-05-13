# 增量流式

Decision Agent 对话接口支持三种返回方式：

| 模式 | 请求参数 | 返回形态 | 适合场景 |
| --- | --- | --- | --- |
| 非流式 | `stream: false` | 一次返回完整 JSON | 后端任务、批处理、简单脚本。 |
| 普通流式 | `stream: true, inc_stream: false` | SSE；每个 `data:` 是当前完整响应快照 | 前端直接覆盖式渲染。 |
| 增量流式 | `stream: true, inc_stream: true` | SSE；每个 `data:` 是增量变更 | 前端希望减少传输和精细更新局部状态。 |

## 请求示例

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "请用三点说明你的能力" \
  '{
  agent_id: $agent_id,
  agent_key: $agent_key,
  agent_version: $agent_version,
  query: $query,
  stream: true,
  inc_stream: true
}' > /tmp/chat-incremental-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d @/tmp/chat-incremental-stream-request.json
```

## 增量事件格式

每条 SSE 消息的 `data:` 是一个 JSON 对象：

```json
{
  "seq_id": 0,
  "key": ["message", "content"],
  "content": "新增内容",
  "action": "append"
}
```

| 字段 | 说明 |
| --- | --- |
| `seq_id` | 增量事件序号，单次对话内递增。 |
| `key` | 目标路径。字符串表示对象字段，数字表示数组下标；空数组表示根节点。 |
| `content` | 本次变更内容。`remove` 或 `end` 时可以为 `null`。 |
| `action` | 变更动作，支持 `upsert`、`append`、`remove`、`end`。 |

## Decision Agent 增量算法

Decision Agent 后端会先接收 Agent Executor 返回的完整 JSON 片段，再在服务端做 diff：

1. 将上一帧完整 JSON 作为 `oldJSON`，当前帧完整 JSON 作为 `newJSON`。
2. 从根节点递归比较对象、数组、字符串和基础类型。
3. 对新增字段发送 `upsert`，对删除字段发送 `remove`。
4. 对字符串前缀增长发送 `append`，例如旧值为 `"你好"`、新值为 `"你好世界"` 时，只发送 `"世界"`。
5. 对数组新增元素发送 `append`，对类型变化或非前缀字符串替换发送 `upsert`。
6. 对话结束时发送 `{ "key": [], "content": null, "action": "end" }`。

服务端不会要求客户端回传上一帧；客户端只需要按 `seq_id` 顺序应用增量事件。

## 客户端合并规则

客户端可以按以下规则还原完整响应：

```js
function applyIncrementalEvent(state, event) {
  if (event.action === 'end') {
    return state;
  }

  const path = event.key ?? [];
  if (path.length === 0) {
    return event.content;
  }

  let cursor = state;
  for (let i = 0; i < path.length - 1; i += 1) {
    const key = path[i];
    const nextKey = path[i + 1];
    if (cursor[key] == null) {
      cursor[key] = Number.isInteger(nextKey) ? [] : {};
    }
    cursor = cursor[key];
  }

  const lastKey = path[path.length - 1];
  if (event.action === 'remove') {
    if (Array.isArray(cursor) && Number.isInteger(lastKey)) {
      cursor.splice(lastKey, 1);
    } else {
      delete cursor[lastKey];
    }
  } else if (event.action === 'append') {
    if (typeof cursor[lastKey] === 'string') {
      cursor[lastKey] += event.content;
    } else if (Array.isArray(cursor[lastKey])) {
      cursor[lastKey].push(event.content);
    } else {
      cursor[lastKey] = event.content;
    }
  } else {
    cursor[lastKey] = event.content;
  }

  return state;
}
```

## 与 Executor 增量输出的关系

Agent Executor 内部也有 `_options.incremental_output`，它会在 Executor 内部把完整 JSON 转为增量事件。REST 用户通过 Decision Agent 对接时，主要使用 `inc_stream`；只有在直接调用 Executor 或调试 Executor 内部输出链路时，才需要关注 `_options.incremental_output`。
