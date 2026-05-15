import { configureKweaver, ensureAgentId, kweaver } from '../shared.js';

// 目标：发起一次非流式对话。
configureKweaver();

const agentId = await ensureAgentId({ createIfMissing: true });

const reply = await kweaver.chat(
  process.env.CHAT_MESSAGE ?? '请用一句话介绍你自己',
  {
    agentId,
    conversationId: process.env.CONVERSATION_ID,
  }
);

console.log(JSON.stringify(reply, null, 2));
