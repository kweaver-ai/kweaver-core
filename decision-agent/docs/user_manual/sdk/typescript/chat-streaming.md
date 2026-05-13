# SDK 对话与流式输出

普通对话优先使用 `kweaver.chat()`。流式输出需要使用底层 client 的 `agents.stream()`。

## 非流式对话

可以在初始化时设置默认 `agentId`：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
  agentId: 'agent-id',
});

const reply = await kweaver.chat('请用一句话介绍你自己');

console.log(reply.text);
console.log(reply.conversationId);
```

也可以在每次调用时传入 `agentId`：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const reply = await kweaver.chat('请用一句话介绍你自己', {
  agentId: 'agent-id',
});

console.log(reply.text);
```

## 继续对话

将上一轮返回的 `conversationId` 传回 SDK，即可继续同一个对话。

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

console.log(second.text);
```

## 流式对话

流式对话通过底层 client 调用。`onTextDelta` 会收到当前累计文本，示例里通过 `prevLen` 只输出新增片段。

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();

let prevLen = 0;
const result = await client.agents.stream(
  'agent-id',
  '请分三点说明你的能力',
  {
    onTextDelta: (fullText) => {
      process.stdout.write(fullText.slice(prevLen));
      prevLen = fullText.length;
    },
    onProgress: (progress) => {
      for (const item of progress) {
        const name = item.skill_info?.name ?? item.agent_name ?? item.stage ?? 'step';
        const status = item.status ?? '';
        console.error('\n[progress]', name, status);
      }
    },
  },
  {
    version: 'v0',
  }
);

console.error('\nconversationId:', result.conversationId ?? '');
```

如果要在流式对话中继续已有对话，传入 `conversationId`：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});

const client = kweaver.getClient();

const result = await client.agents.stream(
  'agent-id',
  '继续补充一个例子',
  {
    onTextDelta: (fullText) => {
      process.stdout.write(fullText);
    },
  },
  {
    conversationId: 'conversation-id',
    version: 'v0',
  }
);

console.log(result);
```
