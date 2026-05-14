# CLI 安装和运行

此处统一说明 KWeaver CLI 的安装、运行、认证、通用环境变量和本地调试连接方式。后续章节默认已经可以直接使用 `kweaver` 命令。

## 环境要求

- Node.js 22 或更高版本。
- 可访问 Decision Agent 服务地址，例如本地 `http://127.0.0.1:13020`。
- 如果需要创建 Agent，请提前准备可用的 LLM ID 和 LLM 名称。

## 使用远程 npm 仓库安装

```bash
npm install -g @kweaver-ai/kweaver-sdk
```

安装后检查入口：

```bash
kweaver --help
kweaver agent --help
```

## 使用本地下载的 SDK 项目安装

如果 SDK 项目已经下载到本地，可以从本地目录安装 CLI：

```bash
npm install -g <path-to-kweaver-sdk>/packages/typescript
```

安装完成后，使用方式仍然是：

```bash
kweaver --help
kweaver agent --help
```

## 本地 no-auth 环境

连接本地 no-auth Decision Agent 服务：

```bash
export KWEAVER_BASE_URL="${KWEAVER_BASE_URL:-http://127.0.0.1:13020}"
export KWEAVER_NO_AUTH=1

curl -fsS "$KWEAVER_BASE_URL/health/ready"
kweaver agent --help
```

如果服务开启业务域校验，可以设置默认业务域：

```bash
export KWEAVER_BUSINESS_DOMAIN="${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

也可以在单条命令上使用 `-bd <value>` 指定业务域。

## 通用环境变量

| 环境变量 | 说明 |
| --- | --- |
| `KWEAVER_BASE_URL` | Decision Agent 服务地址。 |
| `KWEAVER_NO_AUTH` | 本地 no-auth 模式下设置为 `1`，CLI 会省略鉴权 header。 |
| `KWEAVER_TOKEN` | 鉴权环境中的 access token。 |
| `KWEAVER_BUSINESS_DOMAIN` | 默认业务域。支持命令级 `-bd` 的子命令可覆盖该值。 |
| `KWEAVER_LLM_ID` | 示例创建 Agent 时使用的大模型 ID。 |
| `KWEAVER_LLM_NAME` | 示例创建 Agent 时使用的大模型名称。 |

## 通用参数

`kweaver agent` 下不同子命令支持的参数并不完全相同。常见参数如下，具体可用范围请以对应子命令章节为准：

| 参数 | 说明 |
| --- | --- |
| `-bd, --biz-domain <value>` | 覆盖业务域。当前主要用于列表、详情、模板、对话、会话和 skill 相关命令。 |
| `--pretty` | 以缩进 JSON 输出。大多数查询命令默认开启。 |
| `--compact` | 以紧凑 JSON 输出。主要用于 `sessions`、`history`、`trace`、`skill list`。 |
| `--verbose, -v` | 输出完整响应或调试请求信息。 |
| `--help, -h` | 查看命令帮助。 |

部分命令的 `--help` 中可能保留历史兼容参数。本文档按当前 CLI 实际解析和请求行为说明；如果服务需要业务域，优先使用 `kweaver config set-bd <value>` 或 `KWEAVER_BUSINESS_DOMAIN`。

## 鉴权环境

推荐通过 CLI 登录保存认证信息：

```bash
kweaver auth login https://your-kweaver.example.com
kweaver auth status
```

无浏览器或 CI 环境中，也可以通过环境变量传入服务地址和 token：

```bash
export KWEAVER_BASE_URL=https://your-kweaver.example.com
export KWEAVER_TOKEN=your-access-token
```

业务域可以持久保存到当前平台：

```bash
kweaver config set-bd <business-domain-id>
kweaver config show
```

## 示例 Makefile 目标

用户指南的可运行示例位于 [../examples/cli](../examples/cli/README.md)。示例优先使用以下目标：

| 目标 | 含义 |
| --- | --- |
| `make quick-check` | 快速检查，不创建或删除数据，通常只运行 help 和列表类查询。 |
| `make flow` | 完整流程，会创建、查询、更新、发布、取消发布并删除临时 Agent。 |
| `make smoke` | `quick-check` 的兼容别名。 |

## 准备 Agent config

创建 Agent 时需要传入 `--config <file>`。下面生成一个最小配置文件，适合本地快速验证：

```bash
: "${KWEAVER_LLM_ID:?请先设置 KWEAVER_LLM_ID}"
: "${KWEAVER_LLM_NAME:?请先设置 KWEAVER_LLM_NAME}"

export AGENT_CONFIG="${AGENT_CONFIG:-/tmp/kweaver-cli-agent-config.json}"

jq -n \
  --arg llm_id "$KWEAVER_LLM_ID" \
  --arg llm_name "$KWEAVER_LLM_NAME" \
  '{
    input: {
      fields: [
        {
          name: "query",
          type: "string"
        }
      ]
    },
    output: {
      default_format: "markdown",
      variables: {
        answer_var: "answer"
      }
    },
    llms: [
      {
        is_default: true,
        llm_config: {
          id: $llm_id,
          name: $llm_name,
          model_type: "llm",
          temperature: 1,
          top_p: 1,
          top_k: 1,
          frequency_penalty: 0,
          presence_penalty: 0,
          max_tokens: 1000,
          retrieval_max_tokens: 32
        }
      }
    ],
    memory: {
      is_enabled: false
    },
    related_question: {
      is_enabled: false
    },
    plan_mode: {
      is_enabled: false
    }
  }' > "$AGENT_CONFIG"
```

后续示例都假设已经设置 `KWEAVER_BASE_URL`，并且在需要创建 Agent 时已经设置 `AGENT_CONFIG`。
