# TypeScript SDK 用户指南

SDK 手册面向在 TypeScript 或 JavaScript 项目中集成 Decision Agent 能力的用户。文档假设用户通过 npm 安装 `@kweaver-ai/kweaver-sdk`，并通过统一入口引入 SDK：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';
```

如果需要先理解 Agent、个人空间、广场、发布和 Agent 模式，请阅读 [Agent 概念指南](../../concepts/README.md)。

## 文档索引

- [安装和运行](./install-and-run.md)
- [Client 初始化](./client-setup.md)
- [Agents 使用](./agents.md)
- [对话与流式输出](./chat-streaming.md)
- [对话](./conversations.md)

## 阅读顺序

先阅读 [安装和运行](./install-and-run.md)，完成 SDK 安装、运行方式和认证配置。随后根据使用场景阅读初始化、Agent、对话和会话章节。

普通列表和对话场景优先使用 `kweaver.agents()`、`kweaver.chat()` 这类 simple API；创建、发布、删除、流式对话、会话查询等高级能力可以通过 `kweaver.getClient()` 获取底层 client 后调用。

配套可运行示例见 [../../examples/sdk/typescript](../../examples/sdk/typescript/README.md)。示例会先安装本地 SDK 包，再通过 TypeScript 代码调用 npm 包入口；默认运行 `make quick-check`，完整流程运行 `make flow`。
