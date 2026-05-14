# 人工干预与终止

本页说明人工干预、中断恢复、终止执行和断线续连的 REST 接入方式。概念边界可先阅读 [运行控制](../concepts/runtime-control.md)。

## 能力边界

| 能力 | 使用场景 | 接口 |
| --- | --- | --- |
| 人工干预 | 工具或 Sub-Agent 执行前需要用户确认。 | `chat/completion` 或 `debug/completion` 返回 `interrupt_info` |
| 中断恢复 | 用户确认或跳过被中断的工具 / Sub-Agent 后继续执行。 | 再次调用 `chat/completion` 或 `debug/completion` |
| 终止执行 | 用户主动停止正在运行的 run。 | `POST /api/agent-factory/v1/app/{agent_key}/chat/termination` |
| 断线续连 | 流式连接断开，但服务端 run 仍在继续。 | `POST /api/agent-factory/v1/app/{agent_key}/chat/resume` |

`/chat/resume` 只用于流式断线续连，不用于人工干预恢复。人工干预恢复必须把 `resume_interrupt_info` 传回对话接口。

## 开启人工干预

Sub-Agent 技能可以在 `config.skills.agents[]` 中开启人工干预：

```json
{
  "skills": {
    "agents": [
      {
        "agent_key": "<sub-agent-key>",
        "agent_version": "latest",
        "intervention": true,
        "intervention_confirmation_message": "即将调用合同风险抽取 Agent，请确认是否继续。"
      }
    ]
  }
}
```

工具技能也可以使用同名字段：

```json
{
  "skills": {
    "tools": [
      {
        "tool_id": "<tool-id>",
        "intervention": true,
        "intervention_confirmation_message": "即将调用外部工具，请确认是否继续。"
      }
    ]
  }
}
```

`intervention_confirmation_message` 会进入中断详情的 `interrupt_config.confirmation_message`，客户端可以直接展示给用户。

## 发起对话并捕获中断

普通对话请求与其他对话一致。为了后续恢复和终止，建议保存响应里的 `conversation_id`、`agent_run_id` 和 `assistant_message_id`：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v0}" \
  --arg query "请审核这段合同，并在需要调用风险抽取 Agent 前让我确认。" \
  '{
    agent_id: $agent_id,
    agent_version: $agent_version,
    query: $query,
    stream: false,
    executor_version: "v2",
    chat_option: {
      is_need_history: true,
      is_need_doc_retrival_post_process: true,
      is_need_progress: true,
      enable_dependency_cache: true
    }
  }' > /tmp/chat-intervention-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-intervention-request.json | tee /tmp/chat-intervention-response.json
```

上面用 `stream: false` 是为了让示例能直接保存完整 JSON。如果前端或页面使用 `stream: true, inc_stream: true`，人工干预字段仍在合并后的响应快照 `message.ext.interrupt_info` 中。

发生人工干预时，响应的 `message.ext.interrupt_info` 中会包含恢复句柄和中断详情：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "assistant_message_id": "01K...",
  "message": {
    "ext": {
      "interrupt_info": {
        "handle": {
          "frame_id": "frame-id",
          "snapshot_id": "snapshot-id",
          "resume_token": "resume-token",
          "interrupt_type": "tool_interrupt",
          "current_block": 0,
          "restart_block": false
        },
        "data": {
          "tool_name": "获取agent详情",
          "tool_description": "",
          "tool_args": [
            {
              "key": "key",
              "value": "DocQA_Agent",
              "type": "str"
            }
          ],
          "interrupt_config": {
            "requires_confirmation": true,
            "confirmation_message": "是否确认参数并继续执行?"
          }
        }
      }
    }
  }
}
```

注意字段名称：响应里是 `interrupt_info.handle`；恢复请求里才使用 `resume_interrupt_info.resume_handle`。Dolphin 侧的外层事件类型会用 `tool_confirmation` 表示工具确认事件，但传给恢复接口的句柄字段 `interrupt_info.handle.interrupt_type` 是 `tool_interrupt`，调用方应直接复用服务端返回的 `handle`。

## 确认后恢复执行

用户确认继续后，再次调用原对话接口。把响应中的 `interrupt_info.handle` 原样放入 `resume_interrupt_info.resume_handle`，把 `interrupt_info.data` 原样放入 `resume_interrupt_info.data`：

```bash
test -n "$CONVERSATION_ID"
test -n "$AGENT_RUN_ID"
test -n "$INTERRUPTED_ASSISTANT_MESSAGE_ID"

jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v0}" \
  --arg conversation_id "$CONVERSATION_ID" \
  --arg agent_run_id "$AGENT_RUN_ID" \
  --arg interrupted_assistant_message_id "$INTERRUPTED_ASSISTANT_MESSAGE_ID" \
  --argjson interrupt_info "$(jq '.message.ext.interrupt_info' /tmp/chat-intervention-response.json)" \
  '{
    agent_id: $agent_id,
    agent_version: $agent_version,
    conversation_id: $conversation_id,
    agent_run_id: $agent_run_id,
    interrupted_assistant_message_id: $interrupted_assistant_message_id,
    resume_interrupt_info: {
      resume_handle: $interrupt_info.handle,
      action: "confirm",
      modified_args: [],
      data: $interrupt_info.data
    },
    stream: true,
    inc_stream: true,
    executor_version: "v2",
    chat_option: {
      is_need_history: true,
      is_need_doc_retrival_post_process: true,
      is_need_progress: true,
      enable_dependency_cache: true
    }
  }' > /tmp/chat-resume-interrupt-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-resume-interrupt-request.json
```

恢复请求通常不需要再传新的 `query`，后端会根据 `conversation_id`、`agent_run_id` 和 `resume_interrupt_info` 回到被中断的执行位置继续运行。

如果用户选择跳过当前工具或 Sub-Agent，把 `action` 改为 `skip`：

```json
{
  "resume_interrupt_info": {
    "resume_handle": {
      "frame_id": "frame-id",
      "snapshot_id": "snapshot-id",
      "resume_token": "resume-token",
      "interrupt_type": "tool_interrupt",
      "current_block": 0,
      "restart_block": false
    },
    "action": "skip",
    "modified_args": [],
    "data": {
      "tool_name": "获取agent详情",
      "tool_description": "",
      "tool_args": [
        {
          "key": "key",
          "value": "DocQA_Agent",
          "type": "str"
        }
      ],
      "interrupt_config": {
        "requires_confirmation": true,
        "confirmation_message": "是否确认参数并继续执行?"
      }
    }
  }
}
```

如果允许用户修改参数，可以在 `modified_args` 中传入修改后的参数：

```json
{
  "modified_args": [
    {
      "key": "contract_text",
      "value": "用户确认后的合同文本"
    }
  ]
}
```

## Debug 对话中的恢复

Debug 对话同样支持 `resume_interrupt_info`。区别是输入内容放在 `input` 内：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v0}" \
  --arg conversation_id "$CONVERSATION_ID" \
  --arg agent_run_id "$AGENT_RUN_ID" \
  --arg interrupted_assistant_message_id "$INTERRUPTED_ASSISTANT_MESSAGE_ID" \
  --argjson interrupt_info "$(jq '.message.ext.interrupt_info' /tmp/debug-intervention-response.json)" \
  '{
    agent_id: $agent_id,
    agent_version: $agent_version,
    input: {},
    conversation_id: $conversation_id,
    agent_run_id: $agent_run_id,
    interrupted_assistant_message_id: $interrupted_assistant_message_id,
    resume_interrupt_info: {
      resume_handle: $interrupt_info.handle,
      action: "confirm",
      modified_args: [],
      data: $interrupt_info.data
    },
    stream: true,
    inc_stream: true
  }' > /tmp/debug-resume-interrupt-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-resume-interrupt-request.json
```

## 终止执行

终止接口用于停止正在运行的一次执行：

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/termination
```

请求体：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "interrupted_assistant_message_id": "01K..."
}
```

字段说明：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `conversation_id` | 是 | 用于定位 Decision Agent 侧的流式 stop channel。 |
| `agent_run_id` | 否，推荐传 | 用于通知 Agent Executor 终止对应 run。 |
| `interrupted_assistant_message_id` | 否 | 传入后，后端会尝试把对应助手消息标记为 cancelled。 |

示例：

```bash
test -n "$CONVERSATION_ID"

jq -n \
  --arg conversation_id "$CONVERSATION_ID" \
  --arg agent_run_id "${AGENT_RUN_ID:-}" \
  --arg interrupted_assistant_message_id "${INTERRUPTED_ASSISTANT_MESSAGE_ID:-}" \
  '{
    conversation_id: $conversation_id,
    agent_run_id: $agent_run_id,
    interrupted_assistant_message_id: $interrupted_assistant_message_id
  }' > /tmp/chat-termination-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/termination" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-termination-request.json
```

成功时通常返回 `204` 或空成功响应。

## 断线续连

断线续连接口只用于继续读取运行中的流式响应：

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/resume
```

```bash
jq -n --arg conversation_id "$CONVERSATION_ID" '{
  conversation_id: $conversation_id
}' > /tmp/chat-stream-resume-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/resume" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-stream-resume-request.json
```

该接口依赖运行中的内存态会话。如果 run 已经结束，或服务端已清理 `SessionMap`，续连会失败或没有可读取的数据。

## 常见问题

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 人工干预没有触发 | `intervention` 没有配置到实际会被调用的工具或 Sub-Agent 上 | 检查 `config.skills.tools[]` 或 `config.skills.agents[]`。 |
| 恢复时报参数错误 | 把响应字段 `interrupt_info.handle` 直接写成了 `interrupt_info.resume_handle` | 恢复请求中应写 `resume_interrupt_info.resume_handle: interrupt_info.handle`。 |
| 恢复后像重新开始 | 缺少 `conversation_id`、`agent_run_id` 或 `interrupted_assistant_message_id` | 从被中断响应中保存并传回这些字段。 |
| `/chat/resume` 不能恢复人工干预 | 混淆了断线续连和中断恢复 | 人工干预恢复应再次调用 `chat/completion` 或 `debug/completion`。 |
| 终止接口返回错误 | run 已结束，或没有找到对应 stop channel | 如果只需要取消被中断消息，传入 `interrupted_assistant_message_id`；否则重新检查当前 run 状态。 |
