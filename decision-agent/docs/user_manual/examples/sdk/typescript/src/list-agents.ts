import { configureKweaver, kweaver } from './shared.js';

configureKweaver();

const limit = Number(process.env.LIMIT ?? '1');
const agents = await kweaver.agents({ limit });

console.log(JSON.stringify(agents, null, 2));

