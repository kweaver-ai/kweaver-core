# Agents 使用

Agent 列表和普通对话优先使用 simple API。创建、详情、发布、取消发布和删除属于管理能力，通过 `kweaver.getClient().agents` 调用。

## 列表

`kweaver.agents()` 返回已发布 Agent 列表，适合在用户应用中展示可用 Agent。

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const agents = await kweaver.agents({
  keyword: 'sales',
  limit: 10,
});

console.log(agents);
```

本地 no-auth 环境：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: 'http://127.0.0.1:13020',
  auth: false,
});

const agents = await kweaver.agents({ limit: 10 });
console.log(agents);
```

## 创建

创建 Agent 时，将 Agent 配置作为普通 JS object 传入。实际项目中也可以从用户项目自己的 JSON 文件读取配置，再传给 `client.agents.create()`。

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();

const config = {
  input: {
    fields: [
      {
        name: 'query',
        type: 'string',
      },
      {
        name: 'history',
        type: 'object',
      },
      {
        name: 'header',
        type: 'object',
      },
      {
        name: 'self_config',
        type: 'object',
      },
    ],
    augment: {
      enable: false,
      data_source: {},
    },
    rewrite: {
      enable: false,
      llm_config: {
        id: '',
        name: 'test',
        max_tokens: 1000,
        retrieval_max_tokens: 1000,
        model_type: 'llm',
        temperature: 0.5,
        top_p: 0.5,
        top_k: 0,
        frequency_penalty: 0,
        presence_penalty: 0,
      },
    },
    temp_zone_config: null,
  },
  system_prompt: '',
  is_dolphin_mode: 0,
  dolphin: '',
  pre_dolphin: [],
  post_dolphin: [],
  memory: {
    is_enabled: false,
  },
  related_question: {
    is_enabled: false,
  },
  plan_mode: {
    is_enabled: false,
  },
  data_source: {
    knowledge_network: [],
  },
  skills: {
    tools: [],
    agents: [],
    mcps: [],
  },
  llms: [
    {
      is_default: true,
      llm_config: {
        model_type: 'llm',
        temperature: 1,
        top_p: 1,
        top_k: 1,
        frequency_penalty: 0,
        presence_penalty: 0,
        max_tokens: 1000,
        retrieval_max_tokens: 32,
        id: '1951511743712858112',
        name: 'mini-gpt5',
      },
    },
  ],
  output: {
    default_format: 'markdown',
    variables: {
      answer_var: 'answer',
    },
  },
  preset_questions: [],
};

const created = await client.agents.create({
  name: `doc_sdk_agent_${Date.now()}`,
  profile: 'Decision Agent SDK user guide example',
  product_key: 'DIP',
  config,
});

console.log(created.id);
```

如果配置已经保存在业务项目自己的 JSON 文件中，可以这样读取：

```js
import fs from 'node:fs/promises';
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();
const config = JSON.parse(await fs.readFile('./agent-config.json', 'utf8'));

const created = await client.agents.create({
  name: `doc_sdk_agent_${Date.now()}`,
  profile: 'Decision Agent SDK user guide example',
  product_key: 'DIP',
  config,
});

console.log(created);
```

## 详情

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();
const detail = await client.agents.get('agent-id');

console.log(detail);
```

## 发布与取消发布

发布时直接调用 `publish`。如需业务域相关能力，以服务端或 SDK 默认配置为准。

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();

const result = await client.agents.publish('agent-id');

console.log(result);
```

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();
await client.agents.unpublish('agent-id');
```

## 删除

如果 Agent 已经发布，先取消发布，再删除。

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();

await client.agents.unpublish('agent-id');
await client.agents.delete('agent-id');
```
