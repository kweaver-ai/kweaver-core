import { configureKweaver, kweaver } from './shared.js';

configureKweaver();

const agentId = process.env.AGENT_ID;

if (!agentId) {
  console.log('Skip chat example: set AGENT_ID.');
  process.exit(0);
}

const reply = await kweaver.chat(
  process.env.CHAT_MESSAGE ?? '请用一句话介绍你自己',
  {
    agentId,
    version: process.env.AGENT_VERSION ?? 'v0',
    conversationId: process.env.CONVERSATION_ID,
  }
);

console.log(JSON.stringify(reply, null, 2));
