# 对话响应

本页说明 `chat/completion` 与 `debug/completion` 的响应结构。两者都会返回同一种 `ChatResponse`；差异主要在请求体和运行模式，详见 [Debug 对话](./debug-chat.md)。

## 非流式响应

`stream: false` 时，接口一次性返回完整 JSON：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "user_message_id": "01K...",
  "assistant_message_id": "01K...",
  "message": {
    "id": "01K...",
    "conversation_id": "01K...",
    "role": "assistant",
    "content_type": "prompt",
    "status": "",
    "reply_id": "01K...",
    "index": 2,
    "agent_info": {
      "agent_id": "01K...",
      "agent_name": "example_agent",
      "agent_status": "",
      "agent_version": "v0"
    },
    "content": {
      "final_answer": {
        "query": "请用一句话介绍你自己",
        "answer": {
          "text": "我是一个智能助手，可以理解问题并帮助完成任务。"
        },
        "selected_files": null,
        "thinking": "",
        "skill_process": null,
        "answer_type_other": null,
        "output_variables_config": {
          "answer_var": "answer",
          "doc_retrieval_var": "doc_retrieval_res",
          "graph_retrieval_var": "graph_retrieval_res",
          "related_questions_var": "related_questions",
          "other_vars": null,
          "middle_output_vars": null
        }
      },
      "middle_answer": {
        "progress": [],
        "doc_retrieval": null,
        "graph_retrieval": null,
        "other_variables": {}
      }
    },
    "ext": {
      "agent_run_id": "01K...",
      "ttft": 1234
    }
  },
  "error": null
}
```

## 顶层字段

| 字段 | 说明 |
| --- | --- |
| `conversation_id` | 对话 ID。继续对话时把它放回请求体的 `conversation_id`。 |
| `agent_run_id` | 本次 Agent 执行 ID。终止执行、恢复执行和 trace 排障时会用到。 |
| `user_message_id` | 本轮用户消息 ID。 |
| `assistant_message_id` | 本轮助手消息 ID。人工干预恢复时通常要和 `agent_run_id` 一起保存。 |
| `message` | 助手消息对象，结构与对话详情中的消息基本一致。 |
| `error` | 运行错误。成功时通常为 `null`；如果 Executor 运行中返回业务错误，应优先查看此字段。 |

## `message` 字段

| 字段 | 说明 |
| --- | --- |
| `id` | 助手消息 ID，通常与顶层 `assistant_message_id` 一致。 |
| `conversation_id` | 消息所属对话 ID。 |
| `role` | 消息角色。对话响应里的 `message` 通常是 `assistant`。 |
| `content_type` | 内容类型。`prompt` 表示最终答案在 `content.final_answer.answer.text`；`explore` 和 `other` 用于特殊输出。 |
| `reply_id` | 被回复的用户消息 ID，通常与顶层 `user_message_id` 一致。 |
| `index` | 消息在对话中的序号。 |
| `agent_info` | 本次响应使用的 Agent 基本信息和版本。 |
| `content` | 响应主体，包含最终答案和中间过程。 |
| `ext` | 扩展信息，例如 `agent_run_id`、`ttft`、`total_time`、`interrupt_info`。 |

## 最终答案

普通文本回答优先读取：

```text
message.content.final_answer.answer.text
```

如果 Agent 配置了特殊输出等，还需要关注：

| 字段 | 说明 |
| --- | --- |
| `thinking` | 思考过程或规划内容，是否返回取决于 Agent 模式和后端配置。 |
| `output_variables_config` | 输出变量映射，说明答案等变量名。 |

## 中间过程

`message.content.middle_answer` 用于展示执行轨迹：

| 字段 | 说明 |
| --- | --- |
| `progress` | Dolphin / ReAct 执行过程列表，通常包含模型阶段、工具阶段、状态、输入、输出、token 统计等。 |
| `other_variables` | `config.output.variables.other_vars` 中配置的其他输出变量。 |

## 人工干预

当工具或 Sub-Agent 需要用户确认时，响应可能在 `message.ext.interrupt_info` 中返回中断信息。前端或调用方应保存这些字段：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "assistant_message_id": "01K...",
  "interrupt_info": {
    "resume_handle": {
      "frame_id": "frame-id",
      "snapshot_id": "snapshot-id",
      "resume_token": "resume-token",
      "interrupt_type": "tool_interrupt"
    },
    "data": {
      "tool_name": "获取agent详情"
    },
    "action": "confirm",
    "modified_args": []
  }
}
```

恢复执行时，把 `resume_interrupt_info`、`agent_run_id`、`interrupted_assistant_message_id` 放回请求体。具体流程可参考 Cookbook 中的人工干预场景。

## 流式响应

`stream: true` 时返回 SSE：

```text
data: {"conversation_id":"01K...","agent_run_id":"01K...","message":{...},"error":null}
```

普通流式下，每个 `data:` 都是当前完整响应快照，客户端可以直接用最新快照覆盖旧状态。

增量流式下，每个 `data:` 是增量事件：

```json
{
  "seq_id": 1,
  "key": ["message", "content", "final_answer", "answer", "text"],
  "content": "新增文本",
  "action": "append"
}
```

增量事件的合并方式见 [增量流式](./incremental-streaming.md)。

