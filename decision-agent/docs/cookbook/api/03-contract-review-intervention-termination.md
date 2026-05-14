# 用 REST API 创建带人工干预和终止恢复的合同审核流程

> - **难度**：⭐⭐⭐ 专家
> - **耗时**：约 35 分钟
> - **涉及模块**：`agent-factory REST`、`intervention`、`resume`、`terminate`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个 REST API 版人工干预流程：主 Agent 调用 Sub-Agent 前暂停确认，确认后恢复；必要时可终止正在执行的一次 run。

## 2. Prerequisites（前置条件）

- 已准备主 Agent 和已发布 Sub-Agent。
- 已能从流式响应或对话详情中保留 `conversation_id`、`agent_run_id`、`interrupted_assistant_message_id`。

## 3. Steps（操作步骤）

### 3.1 开启人工干预

```jsonc
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

### 3.2 发起对话并等待干预

```bash
AGENT_ID=<main-agent-id> AGENT_KEY=<main-agent-key> \
  make -C docs/user_manual/examples/api chat-incremental-stream
```

### 3.3 人工确认后恢复

人工干预恢复通过再次调用对话接口，并传入服务端返回的 `resume_interrupt_info`：

```jsonc
{
  "query": "继续执行",
  "conversation_id": "<conversation-id>",
  "resume_interrupt_info": {
    "action": "continue"
  },
  "stream": true,
  "inc_stream": true
}
```

### 3.4 终止执行或恢复流式读取

```bash
AGENT_KEY=<main-agent-key> CONVERSATION_ID=<conversation-id> \
  AGENT_RUN_ID=<agent-run-id> INTERRUPTED_ASSISTANT_MESSAGE_ID=<message-id> \
  make -C docs/user_manual/examples/api chat-terminate

AGENT_KEY=<main-agent-key> CONVERSATION_ID=<conversation-id> \
  make -C docs/user_manual/examples/api chat-resume
```

## 4. Expected output（期望输出）

> **判定成功的依据**：干预响应中出现确认提示；终止接口返回成功；续连接口返回 SSE。

```jsonc
{
  "interrupted": true,
  "confirmation_message": "即将调用合同风险抽取 Agent，请确认是否继续。"
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| terminate 返回 400 | 缺少 `conversation_id` | 必须传入正在执行的对话 ID。 |
| resume 无数据 | run 已结束或内存会话已清理 | 只能恢复仍在运行的流式响应。 |
| 人工干预未触发 | 配置未写入 Sub-Agent 技能项 | 检查 `config.skills.agents[].intervention`。 |

## 6. See also（延伸阅读）

- 场景说明：[人工干预和终止恢复流程](../scenario/03-contract-review-intervention-termination.md)
- 用户手册：[Debug 对话](../../user_manual/api/debug-chat.md)
- 用户手册：[对话、会话与执行](../../user_manual/api/conversation-session-run.md)

