import { buildMinimalConfig, configureKweaver, kweaver, stateClear, stateSet } from '../shared.js';

// 目标：串起 create/get/publish/unpublish/delete 完整流程。
// 注意：该脚本会创建并删除临时 Agent，不属于 quick-check。
configureKweaver();

const client = kweaver.getClient();
const name = `example_sdk_agent_${new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14)}`;
let agentId = '';

try {
  const created = await client.agents.create({
    name,
    profile: 'Created by docs/user_manual/examples/sdk/typescript',
    product_key: 'DIP',
    config: buildMinimalConfig(),
  });

  agentId = created.id;
  stateSet({
    AGENT_ID: created.id,
    AGENT_VERSION: created.version ?? 'v0',
  });
  console.log(JSON.stringify(created, null, 2));

  const detail = await client.agents.get(agentId) as Record<string, unknown>;
  stateSet({
    AGENT_KEY: typeof detail.key === 'string' ? detail.key : undefined,
    AGENT_VERSION: typeof detail.version === 'string' ? detail.version : process.env.AGENT_VERSION,
  });
  console.log(JSON.stringify({ id: detail.id, name: detail.name, key: detail.key }, null, 2));

  const published = await client.agents.publish(agentId);
  stateSet({ AGENT_VERSION: published.version });
  console.log(JSON.stringify(published, null, 2));

  await client.agents.unpublish(agentId);
  await client.agents.delete(agentId);
  stateClear(['AGENT_ID', 'AGENT_KEY', 'AGENT_VERSION']);
  agentId = '';
  console.log('SDK flow finished.');
} finally {
  if (agentId) {
    await client.agents.unpublish(agentId).catch(() => undefined);
    await client.agents.delete(agentId).catch(() => undefined);
  }
}
