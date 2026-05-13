import { configureKweaver, kweaver } from './shared.js';

configureKweaver();

const agentId = process.env.AGENT_ID;

if (!agentId) {
  console.log('Skip conversations example: set AGENT_ID.');
  process.exit(0);
}

const client = kweaver.getClient();
const conversations = await client.conversations.list(agentId, {
  version: process.env.AGENT_VERSION ?? 'v0',
  limit: Number(process.env.LIMIT ?? '10'),
});

console.log(JSON.stringify(conversations, null, 2));

const conversationId = process.env.CONVERSATION_ID;
if (!conversationId) {
  console.log('Skip message list example: set CONVERSATION_ID.');
  process.exit(0);
}

const messages = await client.conversations.listMessages(agentId, conversationId, {
  version: process.env.AGENT_VERSION ?? 'v0',
});

console.log(JSON.stringify(messages, null, 2));
