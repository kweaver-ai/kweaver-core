# 对话

`kweaver.chat()` 返回的 `conversationId` 可以用于继续对话。需要查询对话列表或对话消息时，通过底层 client 的 `conversations` resource 完成。

## 创建并继续对话

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
  agentId: 'agent-id',
});

const first = await kweaver.chat('请用一句话介绍你自己');

const second = await kweaver.chat('继续补充一个使用建议', {
  conversationId: first.conversationId,
});

console.log(first.conversationId);
console.log(second.text);
```

## 对话列表

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();

const conversations = await client.conversations.list('agent-id', {
  version: 'v0',
  limit: 10,
});

console.log(conversations);
```

## 对话消息

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();

const messages = await client.conversations.listMessages(
  'agent-id',
  'conversation-id',
  {
    version: 'v0',
  }
);

console.log(messages);
```

## 与 CLI 的对应关系

| SDK | CLI |
| --- | --- |
| `client.conversations.list(agentId)` | `kweaver agent sessions <agent_id>` |
| `client.conversations.listMessages(agentId, conversationId)` | `kweaver agent history <agent_id> <conversation_id>` |
