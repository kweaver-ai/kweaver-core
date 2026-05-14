# TypeScript SDK 示例

本目录使用 TypeScript 调用 npm 包入口：

```ts
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';
```

默认按 `.env` 中的 `KWEAVER_SDK_PACKAGE_SOURCE` 选择远程 npm 包或本地 SDK 项目包，并安装到当前目录的 `node_modules`：

```bash
make install
```

可用目标：

| 目标 | 说明 |
| --- | --- |
| `make install` | 安装 TypeScript 示例依赖，并按 `KWEAVER_SDK_PACKAGE_SOURCE` 安装 SDK 包。 |
| `make import-smoke` | 验证 SDK 包入口可 import。 |
| `make client-setup` | 验证 `kweaver.configure()` 和 `kweaver.getClient()`。 |
| `make list` | 使用 `kweaver.agents()` 查询发布到广场的 Agent。 |
| `make agent-detail` | 使用 `client.agents.get()` 查询详情；优先读取环境变量，其次读取 `../../.tmp/state.env`。 |
| `make create` | 使用 `client.agents.create()` 创建 Agent，并保存 `AGENT_ID`。 |
| `make publish` | 使用 `client.agents.publish()` 发布当前 Agent。 |
| `make unpublish` | 使用 `client.agents.unpublish()` 取消发布当前 Agent。 |
| `make delete` | 使用 `client.agents.delete()` 删除当前 Agent，并清理 state。 |
| `make chat` | 使用 `kweaver.chat()` 发起非流式对话；缺少 `AGENT_ID` 时返回错误提示。 |
| `make chat-stream` | 使用 `client.agents.stream()` 发起流式对话；缺少 `AGENT_ID` 时返回错误提示。 |
| `make conversations` | 使用 `client.conversations.list()` 查询对话列表；缺少 `AGENT_ID` 时返回错误提示。 |
| `make conversation-messages` | 使用 `client.conversations.listMessages()` 查询消息；缺少必需变量时返回错误提示。 |
| `make quick-check` | 执行安装、import、Client 初始化和列表查询。 |
| `make smoke` | `quick-check` 的兼容别名。 |
| `make flow` | 使用 SDK 创建测试 Agent、获取详情、发布、取消发布并删除。 |

使用远程 npm 包：

```bash
make install KWEAVER_SDK_PACKAGE_SOURCE=remote
```

使用本地 SDK 项目包：

```bash
make install KWEAVER_SDK_PACKAGE_SOURCE=local KWEAVER_SDK_TS_DIR="<path-to-kweaver-sdk>/packages/typescript"
```

两种方式都只安装到当前示例目录，不做全局安装。`make install` 会先卸载并清理旧的 `@kweaver-ai/kweaver-sdk`，再安装当前选择的包来源。

`create` 和 `flow` 会修改本地数据，运行前需要可用模型；`.env` 已提供默认值，可按需修改：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
make flow
```

默认 `quick-check` 不依赖 `AGENT_ID`、`KN_ID`、`SKILL_ID` 等外部资源。显式运行依赖资源的目标时，缺少变量会返回错误提示。

目录结构按 SDK resource 拆分，每个子目录都有自己的 Makefile；例如 `make -C chat stream` 会执行 `src/chat/stream.ts`。
