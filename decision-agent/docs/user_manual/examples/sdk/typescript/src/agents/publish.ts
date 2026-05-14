import { configureKweaver, ensureAgentId, kweaver, stateSet } from '../shared.js';

// 目标：发布 Agent。发布请求不传业务域参数。
configureKweaver();

const agentId = await ensureAgentId({ createIfMissing: true });

const client = kweaver.getClient();
const result = await client.agents.publish(agentId);
stateSet({ AGENT_VERSION: result.version });

console.log(JSON.stringify(result, null, 2));
