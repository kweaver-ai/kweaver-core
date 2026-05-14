# 用 TypeScript SDK 创建合同摘要 Agent

> - **难度**：⭐ 入门
> - **耗时**：约 15 分钟
> - **涉及模块**：`@kweaver-ai/kweaver-sdk`、`agents`、`chat`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个 TypeScript 脚本可创建、发布并调用合同摘要 Agent。

## 2. Prerequisites（前置条件）

- 已安装 SDK：`npm install @kweaver-ai/kweaver-sdk`。
- 本地示例可运行：`make -C docs/user_manual/examples/sdk/typescript quick-check`。
- 已设置 `KWEAVER_LLM_ID` 和 `KWEAVER_LLM_NAME`。

## 3. Steps（操作步骤）

### 3.1 初始化 SDK

```ts
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  auth: false,
});
```

### 3.2 创建并发布

```ts
const client = kweaver.getClient();
const created = await client.agents.create({
  name: 'contract-summary',
  profile: '合同摘要 Agent',
  product_key: 'DIP',
  config,
});

await client.agents.publish(created.id);
```

### 3.3 发起对话

```ts
const reply = await kweaver.chat('请总结这段合同：<contract text>', {
  agentId: created.id,
});
```

## 4. Expected output（期望输出）

> **判定成功的依据**：`reply.conversationId` 有值，`reply.text` 包含合同摘要。

```jsonc
{
  "conversationId": "01...",
  "text": "合同摘要..."
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `baseUrl is required` | 未设置连接地址 | 设置 `KWEAVER_BASE_URL` 或显式传入 `baseUrl`。 |
| `Please set KWEAVER_LLM_ID` | 未设置模型 | 设置模型环境变量后再运行 flow。 |
| 对话失败 | Agent 未发布或版本不对 | 先调用 `client.agents.publish()`。 |

## 6. See also（延伸阅读）

- 场景说明：[创建合同摘要 Agent](../../scenario/01-contract-summary-basic.md)
- SDK 示例：[examples/sdk/typescript](../../../user_manual/examples/sdk/typescript/README.md)
- 用户手册：[SDK Client 初始化](../../../user_manual/sdk/typescript/client-setup.md)

