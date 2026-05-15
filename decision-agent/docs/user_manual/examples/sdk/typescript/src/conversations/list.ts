import { configureKweaver, ensureAgentId, kweaver } from '../shared.js';

// 目标：查询 Agent 对话列表。
// 说明：当前 TypeScript SDK 的对话资源会返回数组；如后端返回 {entries,total}，
// 请优先关注后续 SDK 修复版本，或使用 REST API 文档中的 conversation list 示例核对原始响应。
configureKweaver();

const agentId = await ensureAgentId({ createIfMissing: true });

const client = kweaver.getClient();
const conversations = await client.conversations.list(agentId, {
  version: process.env.AGENT_VERSION ?? 'v0',
  limit: Number(process.env.LIMIT ?? '10'),
});

console.log(JSON.stringify(conversations, null, 2));
