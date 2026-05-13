import { configureKweaver, kweaver } from './shared.js';

configureKweaver();

const client = kweaver.getClient();

console.log(JSON.stringify({
  getClient: typeof kweaver.getClient,
  agents: typeof client.agents,
  conversations: typeof client.conversations,
}, null, 2));
