# Decision Agent 后端用户指南

本目录面向需要理解和接入 Decision Agent 后端能力的用户，覆盖核心概念和三种接入方式。业务集成优先使用 SDK；API 文档主要作为底层 REST 参考、排障补充和 SDK 尚未覆盖场景的兜底材料。后续会继续补充 Python 和 Go SDK 文档。

| 使用方式 | 适合谁 | 入口 |
| --- | --- | --- |
| 概念 | 需要先理解 Agent、个人空间、广场、发布、模式等概念的产品、前端、测试、实施和开发者 | [concepts/README.md](./concepts/README.md) |
| API | 需要查看底层 REST 协议、排障或做自定义集成的开发者 | [api/README.md](./api/README.md) |
| CLI | 通过 `kweaver agent ...` 操作 Agent 的实施、测试、开发人员 | [cli/README.md](./cli/README.md) |
| SDK (TypeScript) | 在 TypeScript 应用或脚本中集成 Agent 能力的开发者 | [sdk/typescript/README.md](./sdk/typescript/README.md) |

本文档只覆盖后端能力，不包含前端页面操作。

## 可运行示例

完整示例放在 [examples](./examples/README.md) 目录下：

- [API 示例](./examples/api/README.md)：Shell + cURL。
- [CLI 示例](./examples/cli/README.md)：Shell + `kweaver` 命令。
- [TypeScript SDK 示例](./examples/sdk/typescript/README.md)：TypeScript + npm 包入口。

每个示例目录都有独立 Makefile。默认 `make quick-check` 只执行健康检查、入口 import/help 和列表查询；`make smoke` 作为兼容别名保留。创建、发布、删除 Agent 的完整流程需要显式运行 `make flow`。

## 本地示例约定

所有示例默认在本地运行：

```bash
cd <path-to-decision-agent>
export AF_BASE_URL="${AF_BASE_URL:-http://127.0.0.1:13020}"
export AF_BD="${AF_BD:-bd_public}"
export AF_TOKEN="${AF_TOKEN:-${KWEAVER_TOKEN:-__NO_AUTH__}}"

curl -fsS "$AF_BASE_URL/health/ready"
```

如果本地部署开启鉴权，请先设置有效 token：

```bash
read -rsp "AF_TOKEN: " AF_TOKEN
echo
export KWEAVER_TOKEN="$AF_TOKEN"
export KWEAVER_BASE_URL="$AF_BASE_URL"
export KWEAVER_BUSINESS_DOMAIN="$AF_BD"
```

如果本地是 no-auth 模式，保持默认 `AF_TOKEN=__NO_AUTH__` 即可。CLI 与 SDK 会识别这个值并省略鉴权 header；API cURL 示例中也会提供相同逻辑。

## 推荐阅读路径

1. 理解概念：先读 [Agent 概念指南](./concepts/README.md)，统一 Agent、个人空间、广场、发布和模式等表达。
2. 集成 TypeScript：读 [SDK Client 初始化](./sdk/typescript/client-setup.md)、[Agents 使用](./sdk/typescript/agents.md)、[对话与流式输出](./sdk/typescript/chat-streaming.md)。
3. 命令行验证：读 [CLI 快速开始](./cli/quick-start.md)，用一条本地流程跑通创建、发布和对话。
4. REST 参考：读 [API 生命周期](./api/agent-lifecycle.md)、[Agent 配置](./api/agent-config.md)、[对话与会话](./api/chat-and-conversation.md)。

## 核心概念

| 概念 | 说明 |
| --- | --- |
| `agent_id` | Agent 的资源 ID，创建后返回，管理接口主要使用它。 |
| `agent_key` | 对话接口使用的 app key。CLI/SDK 会先通过 `agent_id + version` 解析出 `agent_key`。 |
| `version` | Agent 版本。未发布版本常见为 `v0`；发布后会产生发布版本。 |
| `conversation_id` | 对话 ID，用于继续多轮对话、查询历史、获取 trace。 |
| `business_domain` | 业务域 header，默认 `bd_public`。 |
| `config` | Agent 核心配置，包含输入输出、LLM、技能、知识网络、模式等。 |

## 示例安全

文档示例不会写入真实 token、cookie 或浏览器 header。请求日志只用于提炼路径、参数和流程，不直接复制敏感内容。
