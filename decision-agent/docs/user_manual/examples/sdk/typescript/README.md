# TypeScript SDK 示例

本目录使用 TypeScript 调用 npm 包入口：

```ts
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';
```

默认从本地 SDK 项目安装：

```bash
make install
```

可用目标：

| 目标 | 说明 |
| --- | --- |
| `make install` | 安装 TypeScript 示例依赖，并安装本地 SDK 包。 |
| `make import-smoke` | 验证 SDK 包入口可 import。 |
| `make client-setup` | 验证 `kweaver.configure()` 和 `kweaver.getClient()`。 |
| `make list` | 使用 `kweaver.agents()` 查询发布到广场的 Agent。 |
| `make agent-detail` | 使用 `client.agents.get()` 查询详情；没有 `AGENT_ID` 时自动跳过。 |
| `make chat` | 使用 `kweaver.chat()` 发起非流式对话；没有 `AGENT_ID` 时自动跳过。 |
| `make chat-stream` | 使用 `client.agents.stream()` 发起流式对话；没有 `AGENT_ID` 时自动跳过。 |
| `make conversations` | 使用 `client.conversations.*` 查询对话和消息；缺少环境变量时自动跳过。 |
| `make quick-check` | 执行安装、import、Client 初始化和列表查询。 |
| `make smoke` | `quick-check` 的兼容别名。 |
| `make flow` | 使用 SDK 创建测试 Agent、获取详情、发布、取消发布并删除。 |

如需覆盖 SDK 包位置：

```bash
make install KWEAVER_SDK_TS_DIR="<path-to-kweaver-sdk>/packages/typescript"
```

如需直接使用远程 npm 包：

```bash
make install SDK_PACKAGE="@kweaver-ai/kweaver-sdk"
```

`flow` 会修改本地数据，运行前需要设置：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
make flow
```

默认 `quick-check` 不依赖 `AGENT_ID`、`KN_ID`、`SKILL_ID` 等外部资源。
