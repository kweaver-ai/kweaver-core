import { configureKweaver, ensureAgentId, kweaver } from '../shared.js';

// 目标：取消发布 Agent。
configureKweaver();

const agentId = await ensureAgentId({ createIfMissing: true });

const client = kweaver.getClient();
await client.agents.unpublish(agentId);

console.log('Unpublished.');
