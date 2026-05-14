# Decision Agent Cookbook

Cookbook 面向已经读过 [Decision Agent 用户手册](../user_manual/README.md) 的使用者，按业务场景组织可执行路径。用户手册解释概念和接口，Cookbook 解释“怎样把这些能力串成一个业务流程”。

## 目录

| 目录 | 内容 |
| --- | --- |
| [scenario](./scenario/README.md) | 三个业务场景的总览：简单 Agent、Sub-Agent 编排、人工干预与终止恢复。 |
| [api](./api/01-contract-summary-basic.md) | 使用 REST API 完成场景。适合排障、底层集成和 SDK 尚未覆盖的能力。 |
| [cli](./cli/01-contract-summary-basic.md) | 使用 `kweaver` CLI 完成场景。当前 CLI 缺失的能力会明确链接到 REST API。 |
| [sdk/typescript](./sdk/typescript/01-contract-summary-basic.md) | 使用 TypeScript SDK 完成场景。 |

## 使用约定

- `conversation` 统一称为“对话”，`session` 统一称为“会话”，`run` 统一称为“一次执行”。
- 示例默认连接本地 Decision Agent：`KWEAVER_BASE_URL=http://127.0.0.1:13020`、`KWEAVER_NO_AUTH=1`。
- 需要真实模型时设置 `KWEAVER_LLM_ID` 和 `KWEAVER_LLM_NAME`。
- Sub-Agent 场景需要先发布子 Agent，并在主 Agent 的 `config.skills.agents[]` 中引用子 Agent 的 `agent_key`。
- 人工干预通过 `config.skills.agents[].intervention=true` 或工具技能上的同名字段启用。

