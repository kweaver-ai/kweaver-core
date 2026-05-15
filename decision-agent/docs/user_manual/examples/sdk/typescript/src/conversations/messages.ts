import { configureKweaver, ensureAgentId, kweaver } from '../shared.js';

// 目标：查询指定对话的消息列表。
configureKweaver();

const agentId = await ensureAgentId({ createIfMissing: true });
const conversationId = process.env.CONVERSATION_ID;

if (!conversationId) {
  throw new Error('CONVERSATION_ID is required. Set CONVERSATION_ID or add it to .env before listing messages.');
}

const client = kweaver.getClient();
const messages = await client.conversations.listMessages(agentId, conversationId, {
  version: process.env.AGENT_VERSION ?? 'v0',
});

console.log(JSON.stringify(messages, null, 2));
