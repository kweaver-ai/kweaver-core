import { configureKweaver, kweaver } from '../shared.js';

// 目标：查询广场 Agent 列表。
configureKweaver();

const limit = Number(process.env.LIMIT ?? '1');
const agents = await kweaver.agents({ limit });

console.log(JSON.stringify(agents, null, 2));
