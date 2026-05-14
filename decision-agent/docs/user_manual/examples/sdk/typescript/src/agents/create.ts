import { mkdir, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';

import { buildMinimalConfig, configureKweaver, kweaver, stateSet } from '../shared.js';

// 目标：创建一个最小可用 Agent，并把响应保存到 .tmp/create-response.json。
configureKweaver();

const client = kweaver.getClient();
const name = process.env.AGENT_NAME ?? `example_sdk_agent_${new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14)}`;
const created = await client.agents.create({
  name,
  profile: 'Created by docs/user_manual/examples/sdk/typescript',
  product_key: 'DIP',
  config: buildMinimalConfig(),
});

const responsePath = resolve('.tmp/create-response.json');
await mkdir(dirname(responsePath), { recursive: true });
await writeFile(responsePath, JSON.stringify(created, null, 2));
stateSet({
  AGENT_ID: created.id,
  AGENT_VERSION: created.version ?? 'v0',
});

console.log(JSON.stringify(created, null, 2));
console.error('Saved response to sdk/typescript/.tmp/create-response.json');
