# Agent 概念指南

Agent 概念指南面向产品、前端、测试、实施和集成开发者，用来统一理解 Decision Agent 中的核心名词和行为边界。这里不重点讲接口参数或命令细节，而是解释为什么需要这些能力、它们在产品里怎么表达，以及 API、CLI、SDK 文档里相关字段如何对应。

## 推荐阅读顺序

1. [基础概念](./agent-basics.md)：先理解 Agent、个人空间、广场和几个常见 ID。
2. [发布逻辑](./publishing.md)：理解发布、更新发布信息、重新发布，以及普通 Agent、技能 Agent、API Agent 的区别。
3. [Agent 模式](./agent-modes.md)：理解默认模式、ReAct 模式和 Dolphin 模式的使用边界。
4. [运行控制](./runtime-control.md)：理解人工干预、中断恢复、终止执行和断线续连。
5. [产品术语与表达](./product-terminology.md)：统一文档、页面和接口里的名词表达。

## 与接入文档的关系

本目录解释概念；具体接入请继续阅读：

- [API 接入指南](../api/README.md)
- [CLI 用户指南](../cli/README.md)
- [TypeScript SDK 用户指南](../sdk/typescript/README.md)

如果只想快速跑通调用，可以直接看 [可运行示例](../examples/README.md)。
