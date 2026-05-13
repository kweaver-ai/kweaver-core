import { buildMinimalConfig, configureKweaver, kweaver } from './shared.js';

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
  console.log(JSON.stringify(created, null, 2));

  const detail = await client.agents.get(agentId);
  console.log(JSON.stringify({ id: detail.id, name: detail.name, key: detail.key }, null, 2));

  const published = await client.agents.publish(agentId);
  console.log(JSON.stringify(published, null, 2));

  await client.agents.unpublish(agentId);
  await client.agents.delete(agentId);
  agentId = '';
  console.log('SDK flow finished.');
} finally {
  if (agentId) {
    await client.agents.unpublish(agentId).catch(() => undefined);
    await client.agents.delete(agentId).catch(() => undefined);
  }
}
