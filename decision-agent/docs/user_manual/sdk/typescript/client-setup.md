# SDK Client 初始化

SDK 推荐使用默认导出的 `kweaver` 对象完成初始化。初始化完成后，可以直接调用 simple API，也可以通过 `kweaver.getClient()` 使用更完整的资源能力。

## 使用已登录配置

如果本机已经通过 CLI 登录并保存过平台配置，可以让 SDK 读取本机配置：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({ config: true });

const agents = await kweaver.agents({ limit: 10 });
console.log(agents);
```

也可以在初始化时设置默认 Agent，后续 `kweaver.chat()` 就不需要每次传 `agentId`：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  config: true,
  agentId: 'agent-id',
});

const reply = await kweaver.chat('请用一句话介绍你自己');
console.log(reply.text);
```

## 使用显式服务地址和 Token

在服务端脚本、业务后端或 CI 环境中，通常显式传入服务地址和访问令牌：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
  businessDomain: process.env.KWEAVER_BUSINESS_DOMAIN ?? 'bd_public',
});

const agents = await kweaver.agents({ keyword: 'sales', limit: 10 });
console.log(agents);
```

如果当前业务逻辑会频繁访问同一个 Agent，可以设置默认 `agentId`：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
  agentId: 'agent-id',
});

const first = await kweaver.chat('分析这个订单是否有风险');
const second = await kweaver.chat('继续给出处理建议', {
  conversationId: first.conversationId,
});

console.log(second.text);
```

## 使用本地 no-auth 环境

本地 Decision Agent 服务如果以 no-auth 模式启动，可以关闭 SDK 的鉴权头：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: 'http://127.0.0.1:13020',
  auth: false,
  businessDomain: 'bd_public',
});

const client = kweaver.getClient();
console.log(client.base());
```

也可以结合环境变量使用：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  auth: false,
});
```

## 访问底层 Client

Simple API 覆盖常用列表、搜索和对话场景。需要创建 Agent、发布 Agent、流式对话、查询会话时，可以获取底层 client：

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
