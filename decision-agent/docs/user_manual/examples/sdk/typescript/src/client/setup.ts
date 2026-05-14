import { configureKweaver, kweaver } from '../shared.js';

// 目标：验证 SDK Client 初始化，并确认核心 resource 可访问。
configureKweaver();

const client = kweaver.getClient();

console.log(JSON.stringify({
  getClient: typeof kweaver.getClient,
  agents: typeof client.agents,
  conversations: typeof client.conversations,
}, null, 2));
