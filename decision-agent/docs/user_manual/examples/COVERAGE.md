# Examples 覆盖矩阵

本页说明 `docs/user_manual` 中 API、CLI、SDK 主要能力对应的可运行示例。`quick-check` 是低风险检查，不创建或删除数据；`flow` 会创建、发布、取消发布并删除临时 Agent。

## API

| 文档能力 | 示例目标 | 示例文件 |
| --- | --- | --- |
| 通用环境与健康检查 | `make -C api health` | [api/health/ready.sh](./api/health/ready.sh) |
| 发布到广场的 Agent 列表 | `make -C api list` | [api/agents/list-published.sh](./api/agents/list-published.sh) |
| Agent 最小配置 | `make -C api agent-config-minimal` | [api/agent-config/minimal.sh](./api/agent-config/minimal.sh) |
| Agent 复杂配置 | `make -C api agent-config-advanced` | [api/agent-config/advanced.sh](./api/agent-config/advanced.sh) |
| Agent 生命周期 | `make -C api flow` | [api/agents/flow.sh](./api/agents/flow.sh) |
| 分类、产品、辅助接口 | `make -C api auxiliary` | [api/auxiliary](./api/auxiliary/) |
| 非流式对话 | `make -C api chat-non-stream` | [api/chat/non-stream.sh](./api/chat/non-stream.sh) |
| 普通流式对话 | `make -C api chat-stream` | [api/chat/stream.sh](./api/chat/stream.sh) |
| 增量流式对话 | `make -C api chat-incremental-stream` | [api/chat/incremental-stream.sh](./api/chat/incremental-stream.sh) |
| Debug 对话 | `make -C api chat-debug` | [api/chat/debug.sh](./api/chat/debug.sh) |
| 终止与续连 | `make -C api chat-terminate` / `make -C api chat-resume` | [api/chat](./api/chat/) |
| 对话列表、详情、更新标题、标记已读、删除 | `make -C api conversations` | [api/conversations](./api/conversations/) |
| 会话缓存管理 | `make -C api conversation-session` | [api/session/manage.sh](./api/session/manage.sh) |
| 导入导出 | `make -C api import-export` | [api/import-export](./api/import-export/) |

## CLI

| 文档能力 | 示例目标 | 示例文件 |
| --- | --- | --- |
| 安装、help、全局环境 | `make -C cli help` | [cli/install/help.sh](./cli/install/help.sh) |
| `list` / `personal-list` / `category-list` / `template-list` | `make -C cli quick-check` | [cli/agents](./cli/agents/) |
| `get` / `get-by-key` / `template-get` | `make -C cli detail-examples` | [cli/agents](./cli/agents/) |
| `create` / `get` / `update` / `publish` / `unpublish` / `delete` | `make -C cli create get update publish unpublish delete` / `make -C cli flow` | [cli/agents](./cli/agents/) |
| `update --knowledge-network-id` | `make -C cli update-knowledge-network` | [cli/agents/update-knowledge-network.sh](./cli/agents/update-knowledge-network.sh) |
| 单轮对话、多轮对话、`sessions` / `history` / `trace` | `make -C cli chat-examples` | [cli/chat](./cli/chat/) |
| `skill add/remove/list` | `make -C cli skill-examples` | [cli/skills](./cli/skills/) |

## TypeScript SDK

| 文档能力 | 示例目标 | 示例文件 |
| --- | --- | --- |
| npm 包入口 import | `make -C sdk/typescript import-smoke` | [sdk/typescript/src/client/import-smoke.ts](./sdk/typescript/src/client/import-smoke.ts) |
| Client 初始化 | `make -C sdk/typescript client-setup` | [sdk/typescript/src/client/setup.ts](./sdk/typescript/src/client/setup.ts) |
| 发布到广场的 Agent 列表 | `make -C sdk/typescript list` | [sdk/typescript/src/agents/list.ts](./sdk/typescript/src/agents/list.ts) |
| Agent 生命周期 | `make -C sdk/typescript create publish unpublish delete` / `make -C sdk/typescript flow` | [sdk/typescript/src/agents](./sdk/typescript/src/agents/) |
| Agent 详情 | `make -C sdk/typescript agent-detail` | [sdk/typescript/src/agents/detail.ts](./sdk/typescript/src/agents/detail.ts) |
| 非流式对话 | `make -C sdk/typescript chat` | [sdk/typescript/src/chat/non-stream.ts](./sdk/typescript/src/chat/non-stream.ts) |
| 流式对话 | `make -C sdk/typescript chat-stream` | [sdk/typescript/src/chat/stream.ts](./sdk/typescript/src/chat/stream.ts) |
| 对话列表 | `make -C sdk/typescript conversations` | [sdk/typescript/src/conversations/list.ts](./sdk/typescript/src/conversations/list.ts) |
| 对话消息 | `make -C sdk/typescript conversation-messages` | [sdk/typescript/src/conversations/messages.ts](./sdk/typescript/src/conversations/messages.ts) |

## 可选资源

创建类示例会把生成的运行状态写入 `.tmp/state.env`。显式运行依赖资源的目标时，如果缺少必要变量，会给出错误提示并返回非 0 状态：

| 环境变量 | 用途 |
| --- | --- |
| `KN_ID` | 绑定知识网络。 |
| `SKILL_ID` | 绑定技能。 |
| `AGENT_ID` | 查询详情、更新、删除、对话等需要已有 Agent 的示例。 |
| `AGENT_KEY` | 通过已发布应用入口发起对话。 |
| `CONVERSATION_ID` | 查询对话消息、会话状态和 trace。 |
