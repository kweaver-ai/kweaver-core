import { configureKweaver, kweaver } from './shared.js';

configureKweaver();

const agentId = process.env.AGENT_ID;

if (!agentId) {
  console.log('Skip chat stream example: set AGENT_ID.');
  process.exit(0);
}

const client = kweaver.getClient();
let prevLen = 0;

const result = await client.agents.stream(
  agentId,
  process.env.CHAT_MESSAGE ?? '请分三点说明你的能力',
  {
    onTextDelta: (fullText: string) => {
      process.stdout.write(fullText.slice(prevLen));
      prevLen = fullText.length;
    },
    onProgress: (progress: unknown) => {
      console.error('\n[progress]', JSON.stringify(progress));
    },
  },
  {
    version: process.env.AGENT_VERSION ?? 'v0',
    conversationId: process.env.CONVERSATION_ID,
  }
);

console.error('\nresult:', JSON.stringify(result, null, 2));
