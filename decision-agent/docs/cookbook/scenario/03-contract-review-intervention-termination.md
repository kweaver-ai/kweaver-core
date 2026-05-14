# 创建带人工干预和终止恢复的合同审核流程

> - **难度**：⭐⭐⭐ 专家
> - **耗时**：约 35 分钟
> - **涉及模块**：`agent`、`sub-agent`、`intervention`、`resume`、`terminate`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个主合同审核 Agent，它调用 Sub-Agent 前需要人工确认；用户可以在执行中终止任务，也可以在人工确认后恢复执行。

## 2. Prerequisites（前置条件）

- 已准备场景二中的 Sub-Agent。
- 已理解 `conversation_id`、`agent_run_id`、`interrupted_assistant_message_id` 的来源。
- 已了解当前 CLI 不直接覆盖 terminate / resume / 对话 CRUD，需要使用 API 或 SDK 底层能力补齐。

## 3. Steps（操作步骤）

### 3.1 开启 Sub-Agent 人工干预

在主 Agent 配置中给 Sub-Agent 技能打开 `intervention`：

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

### 3.2 触发干预并恢复

调用主 Agent 后，当响应中出现干预提示，客户端展示给用户确认。确认后把用户选择写入 `resume_interrupt_info`，再次调用对话接口恢复执行。

### 3.3 终止执行或断线续连

终止执行使用 `/chat/termination`。如果只是流式连接断开、后端 run 仍在继续，则使用 `/chat/resume` 恢复读取。

## 4. Expected output（期望输出）

> **判定成功的依据**：干预前响应包含确认提示；确认后执行继续；终止接口返回 204 或空成功响应。

```jsonc
{
  "message": {
    "ext": {
      "interrupted": true,
      "confirmation_message": "即将调用合同风险抽取 Agent，请确认是否继续。"
    }
  }
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 没有出现人工干预 | `intervention` 未写在被调用的 Sub-Agent 技能上 | 检查 `config.skills.agents[].intervention`。 |
| 恢复后重新开始执行 | `resume_interrupt_info` 缺少必要上下文 | 从中断响应中保留 run/message 信息并传回。 |
| terminate 无效 | 缺少 `conversation_id` 或 run 已结束 | 用最新流式响应或对话详情中的 run 信息重试。 |

## 6. See also（延伸阅读）

- API Recipe：[api/03-contract-review-intervention-termination.md](../api/03-contract-review-intervention-termination.md)
- CLI Recipe：[cli/03-contract-review-intervention-termination.md](../cli/03-contract-review-intervention-termination.md)
- SDK Recipe：[sdk/typescript/03-contract-review-intervention-termination.md](../sdk/typescript/03-contract-review-intervention-termination.md)
- 用户手册：[对话、会话与执行](../../user_manual/api/conversation-session-run.md)

