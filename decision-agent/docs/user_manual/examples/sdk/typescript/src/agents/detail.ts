import { configureKweaver, ensureAgentId, kweaver, stateSet } from '../shared.js';

// 目标：按 agent_id 获取 Agent 详情。
configureKweaver();

const agentId = await ensureAgentId({ createIfMissing: true });

const client = kweaver.getClient();
const detail = await client.agents.get(agentId);
stateSet({
  AGENT_KEY: detail.key,
  AGENT_VERSION: detail.version ?? process.env.AGENT_VERSION,
});

console.log(JSON.stringify(detail, null, 2));
