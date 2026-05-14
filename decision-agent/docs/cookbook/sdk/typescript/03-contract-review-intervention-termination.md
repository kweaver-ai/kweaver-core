# 用 TypeScript SDK 处理人工干预和终止恢复

> - **难度**：⭐⭐⭐ 专家
> - **耗时**：约 35 分钟
> - **涉及模块**：`@kweaver-ai/kweaver-sdk`、`intervention`、`REST fallback`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个 SDK 视角的人工干预流程：Agent 创建、发布、普通执行用 SDK；当前 SDK 未封装的 terminate / resume 使用 REST API 兜底。

## 2. Prerequisites（前置条件）

- 已完成带 Sub-Agent 的主 Agent。
- Sub-Agent 技能配置已开启 `intervention: true`。
- 已能从响应中保留 `conversation_id`、`agent_run_id` 和中断消息 ID。

## 3. Steps（操作步骤）

### 3.1 创建带干预的主 Agent

```ts
const config = {
  input: { fields: [{ name: 'query', type: 'string' }] },
  output: { default_format: 'markdown', variables: { answer_var: 'answer' } },
  llms,
  skills: {
    agents: [
      {
        agent_key: process.env.SUB_AGENT_KEY,
        agent_version: 'latest',
        intervention: true,
        intervention_confirmation_message: '即将调用合同风险抽取 Agent，请确认是否继续。',
      },
    ],
  },
};
```

### 3.2 用 SDK 发起流式对话

```ts
const result = await client.agents.stream(mainAgentId, '请审核这段合同：<contract text>', {
  onTextDelta: (fullText) => process.stdout.write(fullText),
  onProgress: (progress) => console.error(JSON.stringify(progress)),
});
```

### 3.3 使用 REST API 兜底恢复或终止

当前 TypeScript SDK resource 尚未提供 terminate / resume 的正式方法。需要时调用 REST API：

```ts
await fetch(`${baseUrl}/api/agent-factory/v1/app/${agentKey}/chat/termination`, {
  method: 'POST',
  headers,
  body: JSON.stringify({
    conversation_id: conversationId,
    agent_run_id: agentRunId,
    interrupted_assistant_message_id: interruptedAssistantMessageId,
  }),
});
```

## 4. Expected output（期望输出）

> **判定成功的依据**：SDK 能创建和执行带干预配置的 Agent；终止和续连请求通过 REST 返回成功。

```jsonc
{
  "conversation_id": "01...",
  "agent_run_id": "01..."
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| SDK 找不到 terminate 方法 | 当前 SDK 未封装该能力 | 使用 REST API 兜底。 |
| 人工干预未触发 | 配置未写到 `skills.agents[]` | 检查 `intervention` 和 `agent_key`。 |
| 续连没有事件 | run 已结束 | 改查对话详情或重新执行。 |

## 6. See also（延伸阅读）

- 场景说明：[人工干预和终止恢复流程](../../scenario/03-contract-review-intervention-termination.md)
- API Recipe：[api/03-contract-review-intervention-termination.md](../../api/03-contract-review-intervention-termination.md)
- 用户手册：[SDK 对话](../../../user_manual/sdk/typescript/chat-streaming.md)

