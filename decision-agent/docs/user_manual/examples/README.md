# Decision Agent 可运行示例

本目录提供和用户指南配套的可运行示例。示例按接入方式拆分：

| 目录 | 内容 | 默认 quick-check |
| --- | --- | --- |
| [api](./api/README.md) | Shell + cURL，直接调用 Decision Agent REST API。 | 健康检查、发布 Agent 列表 |
| [cli](./cli/README.md) | Shell + `kweaver` 命令，使用安装后的 CLI。 | 本地包安装、CLI help、发布 Agent 列表、个人空间列表 |
| [sdk/typescript](./sdk/typescript/README.md) | TypeScript，使用 npm 包入口调用 SDK。 | 本地包安装、SDK import、Client 初始化、发布 Agent 列表 |

覆盖关系见 [COVERAGE.md](./COVERAGE.md)。

默认连接本地 Decision Agent 服务：

```bash
export KWEAVER_BASE_URL="${KWEAVER_BASE_URL:-http://127.0.0.1:13020}"
export KWEAVER_NO_AUTH=1
```

CLI 与 SDK 示例默认从本地 SDK 项目安装包：

```bash
export KWEAVER_SDK_TS_DIR="<path-to-kweaver-sdk>/packages/typescript"
```

也可以使用远程 npm 包：

```bash
export SDK_PACKAGE="@kweaver-ai/kweaver-sdk"
```

推荐先运行低风险检查：

```bash
make quick-check
```

`make smoke` 作为历史兼容别名保留，等价于 `make quick-check`。

会创建或删除 Agent 的完整流程不会作为默认目标执行，请显式运行对应目录或本目录的 `make flow`。运行前通常需要提供可用的 LLM：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
```

知识网络和技能绑定示例依赖外部资源，默认不阻塞 `quick-check`。只有显式设置 `KN_ID` 或 `SKILL_ID` 后，对应目标才会实际运行。
