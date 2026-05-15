# Decision Agent 可运行示例

本目录提供和用户指南配套的可运行示例。示例按接入方式拆分：

| 目录 | 内容 | 默认 quick-check |
| --- | --- | --- |
| [api](./api/README.md) | Shell + cURL，直接调用 Decision Agent REST API。 | 健康检查、发布 Agent 列表 |
| [cli](./cli/README.md) | Shell + `kweaver` 命令，使用安装后的 CLI。 | SDK 包安装、CLI help、发布 Agent 列表、个人空间列表 |
| [sdk/typescript](./sdk/typescript/README.md) | TypeScript，使用 npm 包入口调用 SDK。 | SDK 包安装、SDK import、Client 初始化、发布 Agent 列表 |

覆盖关系见 [COVERAGE.md](./COVERAGE.md)。

## 环境变量

本目录提供两个环境文件：

- `.env.example`：可提交的模板，列出所有示例会用到的变量和默认值。
- `.env`：本地可修改配置，已加入 `.gitignore`，不会提交。

首次使用可以直接编辑 `.env`。默认连接本地 Decision Agent 服务：

```bash
KWEAVER_BASE_URL=http://127.0.0.1:13020
KWEAVER_NO_AUTH=1
KWEAVER_BUSINESS_DOMAIN=bd_public
KWEAVER_LLM_ID=xxx
KWEAVER_LLM_NAME=deepseek-v3
```

CLI 创建 Agent 还会读取本地配置文件：

```bash
docs/user_manual/examples/cli/agents/agent-config.json
```

该文件由 [agent-config.json.example](./cli/agents/agent-config.json.example) 初始化，并已加入 `.gitignore`。可以按本地模型、知识网络和技能情况修改它。

CLI 与 SDK 示例通过 `KWEAVER_SDK_PACKAGE_SOURCE` 选择包来源。`.env.example` 默认使用远程 npm 包，不依赖本地 SDK 仓库：

```bash
KWEAVER_SDK_PACKAGE_SOURCE=remote
KWEAVER_SDK_REMOTE_PACKAGE=@kweaver-ai/kweaver-sdk
```

如需使用本地 SDK 项目包，可以在 `.env` 或当前 shell 中改为：

```bash
KWEAVER_SDK_PACKAGE_SOURCE=local
KWEAVER_SDK_TS_DIR=../../../../../kweaver-sdk/packages/typescript
```

远程和本地两种方式都会安装到当前示例目录的 `node_modules`，不会执行全局安装。每次 `make install` 会先卸载并清理旧的 `@kweaver-ai/kweaver-sdk`，再安装当前选择的包来源，方便在远程包和本地包之间切换。

变量优先级为：当前 shell 中显式设置的环境变量 > `.env` > `.tmp/state.env` > 脚本默认值。

## 临时状态

创建 Agent 的示例会把生成的 `AGENT_ID`、`AGENT_VERSION` 等写入 `.tmp/state.env`。后续 `get`、`update`、`publish`、`chat` 等示例会自动读取这个文件，因此通常不需要手工复制 `AGENT_ID`。

如果缺少必需变量，示例会打印明确的 `Error:` 并返回非 0 状态。对需要 `AGENT_ID` 的脚本，如果在交互式终端中运行，会询问是否立即创建一个临时 Agent；非交互环境不会隐式创建，也不会静默成功。

推荐先运行低风险检查：

```bash
make quick-check
```

`make smoke` 作为历史兼容别名保留，等价于 `make quick-check`。

会创建或删除 Agent 的完整流程不会作为默认目标执行，请显式运行对应目录或本目录的 `make flow`。运行前通常需要提供可用的 LLM；`.env` 已给出默认值，可按环境修改：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
```

知识网络和技能绑定示例依赖外部资源，不进入默认 `quick-check`。显式运行对应目标时，如果没有设置 `KN_ID` 或 `SKILL_ID`，示例会返回错误提示。
