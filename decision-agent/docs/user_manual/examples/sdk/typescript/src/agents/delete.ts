import { configureKweaver, kweaver, stateClear } from '../shared.js';

// 目标：删除 Agent。
configureKweaver();

const agentId = process.env.AGENT_ID;
if (!agentId) {
  throw new Error('AGENT_ID is required. Set AGENT_ID, add it to .env, or run make create before deleting.');
}

const client = kweaver.getClient();
await client.agents.delete(agentId);
stateClear(['AGENT_ID', 'AGENT_KEY', 'AGENT_VERSION']);

console.log('Deleted.');
