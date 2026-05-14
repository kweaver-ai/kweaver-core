# 用 TypeScript SDK 创建带 Sub-Agent 的合同审核流程

> - **难度**：⭐⭐ 进阶
> - **耗时**：约 25 分钟
> - **涉及模块**：`@kweaver-ai/kweaver-sdk`、`agents`、`stream`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个 TypeScript SDK 版主 Agent，它通过 `config.skills.agents[]` 引用已发布 Sub-Agent，并使用流式对话查看执行进度。

## 2. Prerequisites（前置条件）

- 已创建并发布 Sub-Agent，拿到 `SUB_AGENT_KEY`。
- 已安装 TypeScript SDK。
- 已准备合同文本输入。

## 3. Steps（操作步骤）

### 3.1 配置 Sub-Agent 技能

```ts
const mainConfig = {
  input: { fields: [{ name: 'query', type: 'string' }] },
  output: { default_format: 'markdown', variables: { answer_var: 'answer' } },
  llms,
  skills: {
    agents: [
      {
        agent_key: process.env.SUB_AGENT_KEY,
        agent_version: 'latest',
        agent_timeout: 60,
      },
    ],
  },
};
```

### 3.2 创建并发布主 Agent

```ts
const client = kweaver.getClient();
const main = await client.agents.create({
  name: 'contract-review-main',
  profile: '调用 Sub-Agent 进行合同审核',
  product_key: 'DIP',
  config: mainConfig,
});

await client.agents.publish(main.id);
```

### 3.3 流式调用主 Agent

```ts
await client.agents.stream(
  main.id,
  '请审核这段合同：<contract text>',
  {
    onTextDelta: (fullText) => process.stdout.write(fullText),
    onProgress: (progress) => console.error(JSON.stringify(progress)),
  },
  { version: 'v0' }
);
```

## 4. Expected output（期望输出）

> **判定成功的依据**：`onProgress` 中出现 agent 技能步骤，最终文本包含审核结论。

```jsonc
[
  { "skill_info": { "type": "agent" }, "status": "success" }
]
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| Sub-Agent 未调用 | 主 Agent 指令未描述调用策略 | 在角色指令或 Dolphin 中显式要求调用。 |
| `agent_key is required` | `SUB_AGENT_KEY` 未设置 | 先发布子 Agent，并设置环境变量。 |
| 需要原始增量 patch | SDK stream 默认给文本和进度回调 | 使用 REST API 增量流式 Recipe。 |

## 6. See also（延伸阅读）

- 场景说明：[带 Sub-Agent 的合同审核流程](../../scenario/02-contract-review-with-sub-agent.md)
- API 增量流式：[api/02-contract-review-with-sub-agent.md](../../api/02-contract-review-with-sub-agent.md)
- SDK 示例：[chat/stream.ts](../../../user_manual/examples/sdk/typescript/src/chat/stream.ts)

