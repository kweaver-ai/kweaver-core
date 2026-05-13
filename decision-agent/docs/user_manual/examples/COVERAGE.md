# Examples 覆盖矩阵

本页说明 `docs/user_manual` 中 API、CLI、SDK 主要能力对应的可运行示例。`quick-check` 是低风险检查，不创建或删除数据；`flow` 会创建、发布、取消发布并删除临时 Agent。

## API

| 文档能力 | 示例目标 | 示例文件 |
| --- | --- | --- |
| 通用环境与健康检查 | `make -C api health` | [api/scripts/health.sh](./api/scripts/health.sh) |
| 发布到广场的 Agent 列表 | `make -C api list` | [api/scripts/list-agents.sh](./api/scripts/list-agents.sh) |
| Agent 最小配置与复杂配置 | `make -C api agent-config` | [api/scripts/agent-config.sh](./api/scripts/agent-config.sh) |
| Agent 生命周期 | `make -C api flow` | [api/scripts/agent-flow.sh](./api/scripts/agent-flow.sh) |
| 分类、产品、辅助接口 | `make -C api auxiliary` | [api/scripts/auxiliary.sh](./api/scripts/auxiliary.sh) |
| 非流式/流式/增量流式/Debug 对话 | `make -C api chat-examples` | [api/scripts/chat.sh](./api/scripts/chat.sh) |
| 对话、会话、执行相关查询 | `make -C api conversations` | [api/scripts/conversations.sh](./api/scripts/conversations.sh) |
| 导入导出 | `make -C api import-export` | [api/scripts/import-export.sh](./api/scripts/import-export.sh) |

## CLI

| 文档能力 | 示例目标 | 示例文件 |
| --- | --- | --- |
| 安装、help、全局环境 | `make -C cli help` | [cli/Makefile](./cli/Makefile) |
| `list` / `personal-list` / `category-list` / `template-list` | `make -C cli quick-check` | [cli/scripts/list-agents.sh](./cli/scripts/list-agents.sh) |
| `get` / `get-by-key` / `template-get` | `make -C cli detail-examples` | [cli/scripts/agent-optional.sh](./cli/scripts/agent-optional.sh) |
| `create` / `get` / `publish` / `unpublish` / `delete` | `make -C cli flow` | [cli/scripts/agent-flow.sh](./cli/scripts/agent-flow.sh) |
| `update --knowledge-network-id` | `make -C cli update-knowledge-network` | [cli/scripts/agent-optional.sh](./cli/scripts/agent-optional.sh) |
| `chat` / `sessions` / `history` / `trace` | `make -C cli chat-examples` | [cli/scripts/agent-optional.sh](./cli/scripts/agent-optional.sh) |
| `skill add/remove/list` | `make -C cli skill-examples` | [cli/scripts/agent-optional.sh](./cli/scripts/agent-optional.sh) |

## TypeScript SDK

| 文档能力 | 示例目标 | 示例文件 |
| --- | --- | --- |
| npm 包入口 import | `make -C sdk/typescript import-smoke` | [sdk/typescript/src/import-smoke.ts](./sdk/typescript/src/import-smoke.ts) |
| Client 初始化 | `make -C sdk/typescript client-setup` | [sdk/typescript/src/client-setup.ts](./sdk/typescript/src/client-setup.ts) |
| 发布到广场的 Agent 列表 | `make -C sdk/typescript list` | [sdk/typescript/src/list-agents.ts](./sdk/typescript/src/list-agents.ts) |
| Agent 生命周期 | `make -C sdk/typescript flow` | [sdk/typescript/src/agent-flow.ts](./sdk/typescript/src/agent-flow.ts) |
| Agent 详情 | `make -C sdk/typescript agent-detail` | [sdk/typescript/src/agent-detail.ts](./sdk/typescript/src/agent-detail.ts) |
| 非流式对话 | `make -C sdk/typescript chat` | [sdk/typescript/src/chat.ts](./sdk/typescript/src/chat.ts) |
| 流式对话 | `make -C sdk/typescript chat-stream` | [sdk/typescript/src/chat-stream.ts](./sdk/typescript/src/chat-stream.ts) |
| 对话列表与消息 | `make -C sdk/typescript conversations` | [sdk/typescript/src/conversations.ts](./sdk/typescript/src/conversations.ts) |

## 可选资源

以下目标只有在设置对应环境变量后才会运行，否则会给出跳过提示并返回成功：

| 环境变量 | 用途 |
| --- | --- |
| `KN_ID` | 绑定知识网络。 |
| `SKILL_ID` | 绑定技能。 |
| `AGENT_ID` | 查询详情、更新、删除、对话等需要已有 Agent 的示例。 |
| `AGENT_KEY` | 通过已发布应用入口发起对话。 |
| `CONVERSATION_ID` | 查询对话消息、会话状态和 trace。 |
