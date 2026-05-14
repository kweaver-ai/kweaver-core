# 运行控制

运行控制用于处理 Agent 执行过程中的人工确认、恢复、终止和断线续连。这里先解释产品和运行时概念；具体 REST 请求请阅读 [人工干预与终止](../api/intervention-termination.md)。

## 核心概念

| 概念 | 说明 | 常见字段或接口 |
| --- | --- | --- |
| 人工干预 | 工具或 Sub-Agent 执行前需要用户确认，Executor 会暂停当前 run 并返回中断信息。 | `interrupt_info`、`intervention` |
| 中断恢复 | 用户确认或跳过后，继续之前被暂停的一次执行。 | `resume_interrupt_info`、`agent_run_id`、`interrupted_assistant_message_id` |
| 终止执行 | 主动停止正在运行的一次执行，必要时同步取消被中断的助手消息。 | `/chat/termination` |
| 断线续连 | 流式连接断开但服务端 run 仍在继续时，重新接入并读取后续事件。 | `/chat/resume` |

人工干预恢复和断线续连容易混淆：

- 人工干预恢复是继续一个被 Dolphin / Executor 暂停的 run，需要再次调用 `chat/completion` 或 `debug/completion`，并传入 `resume_interrupt_info`。
- 断线续连只是恢复读取还在运行的流式响应，使用 `/chat/resume`，请求体只需要 `conversation_id`。

## 人工干预

人工干预通常配置在工具技能或 Sub-Agent 技能上。执行到该工具或 Sub-Agent 时，Dolphin SDK 根据 `intervention` 配置生成中断事件，Agent Executor 将中断信息返回给 Decision Agent。

常见配置字段：

| 字段 | 说明 |
| --- | --- |
| `intervention` | 是否启用人工干预。 |
| `intervention_confirmation_message` | 展示给用户的确认提示。 |

典型场景是合同审核主 Agent 调用合同风险抽取 Sub-Agent 前，先让用户确认是否继续调用。

## 中断恢复

当响应中出现 `message.ext.interrupt_info` 时，客户端应保存三类信息：

- `conversation_id`：对话 ID，用于继续同一段对话。
- `agent_run_id`：被暂停的一次执行 ID。
- `assistant_message_id`：被中断的助手消息 ID，恢复时作为 `interrupted_assistant_message_id` 传回。
- `interrupt_info.handle` 和 `interrupt_info.data`：恢复执行所需的句柄和中断详情。

这里有一个容易混淆的字段：Dolphin 外层中断事件可能用 `tool_confirmation` 表示“工具确认事件”，但真正传回 REST 恢复接口的句柄是 `interrupt_info.handle`，其中 `handle.interrupt_type` 当前为 `tool_interrupt`。接入方应原样保存并回传 `handle`，不要把它改写成外层事件类型。

用户确认后，客户端再次调用对话接口，把 `interrupt_info.handle` 转成 `resume_interrupt_info.resume_handle`，并设置 `action`：

| `action` | 含义 |
| --- | --- |
| `confirm` | 确认继续执行。 |
| `skip` | 跳过当前被中断的工具或 Sub-Agent。 |

如果允许用户修改工具参数，可以通过 `modified_args` 传回修改后的参数；不需要修改时传空数组或省略。

恢复请求通常不需要再传新的 `query`。后端会根据 `conversation_id`、`agent_run_id`、`interrupted_assistant_message_id` 和 `resume_interrupt_info` 回到暂停点继续执行。

## 终止执行

终止执行用于用户主动停止一个正在运行的 run。它会尽量通知 Agent Executor 终止对应的 `agent_run_id`，同时关闭 Decision Agent 侧的流式 stop channel。

如果传入 `interrupted_assistant_message_id`，后端还会把对应助手消息标记为 cancelled。这个字段适合在人工干预等待中用户选择终止时使用。

## 断线续连

流式对话过程中，Decision Agent 会按 `conversation_id` 在内存中保存当前响应快照和续连信号。如果客户端网络断开，而服务端 run 仍在继续，可以调用 `/chat/resume` 重新读取流式增量事件。

断线续连有两个边界：

- 它依赖运行中的内存态 `SessionMap`，run 结束后通常无法继续读取。
- 它不会恢复人工干预；人工干预必须走 `resume_interrupt_info`。

## 建议接入顺序

1. 普通对话先保存 `conversation_id`、`agent_run_id` 和 `assistant_message_id`。
2. 如果响应里有 `interrupt_info`，展示确认提示并保存 `handle`、`data`。
3. 用户确认后调用原对话接口恢复执行。
4. 用户取消或任务不再需要时调用终止接口。
5. 只有流式连接断开且 run 仍在继续时，才调用 `/chat/resume`。
