import { configureKweaver, kweaver } from './shared.js';

configureKweaver();

const agentId = process.env.AGENT_ID;

if (!agentId) {
  console.log('Skip agent detail example: set AGENT_ID.');
  process.exit(0);
}

const client = kweaver.getClient();
const detail = await client.agents.get(agentId);

console.log(JSON.stringify(detail, null, 2));
