<!-- 请勿直接编辑：本文件由 docs/user_manual/cli 下的 `make aggregate` 生成。 -->
<!-- 来源文件：index.md, install-and-run.md, quick-start.md, agent-lifecycle.md, chat-session-trace.md, local-debug-flow.md。 -->

# CLI 用户指南（聚合版）

> 本文件由脚本生成，请不要直接修改本文件；如需调整内容，请修改分文件文档后运行 `make -C docs/user_manual/cli aggregate`。

## 目录

- [CLI 用户指南](#cli-用户指南)
- [CLI 安装和运行](#cli-安装和运行)
- [CLI 快速开始](#cli-快速开始)
- [Agent 生命周期命令](#agent-生命周期命令)
- [CLI 对话与 Trace](#cli-对话与-trace)
- [CLI 本地调试流程](#cli-本地调试流程)


<!-- 来源：index.md -->

## CLI 用户指南

CLI 手册面向已经安装 `@kweaver-ai/kweaver-sdk` 后使用 `kweaver ...` 命令的用户。普通用户不需要进入 SDK 源码目录，也不需要使用源码调试入口。

如果需要先理解 Agent、个人空间、广场、发布和 Agent 模式，请阅读 [Agent 概念指南](../concepts/README.md)。

### 阅读顺序

1. [安装和运行](#cli-安装和运行)：完成 CLI 安装、连接配置和认证。
2. [快速开始](#cli-快速开始)：跑通创建、详情、发布、对话、清理。
3. [Agent 生命周期命令](#agent-生命周期命令)：查看 Agent、个人空间、模板、创建、更新、删除。
4. [对话与 Trace](#cli-对话与-trace)：单轮对话、继续对话、流式输出、对话记录和 trace。
5. [本地调试流程](#cli-本地调试流程)：仅面向 SDK/CLI 开发者。

配套可运行示例见 [../examples/cli](../examples/cli/README.md)。示例会先安装本地 SDK 包，再通过 `kweaver ...` 命令执行 `quick-check` 或完整流程。

### 最小环境

```bash
export KWEAVER_BASE_URL="${KWEAVER_BASE_URL:-http://127.0.0.1:13020}"
export KWEAVER_NO_AUTH=1

curl -fsS "$KWEAVER_BASE_URL/health/ready"
kweaver agent --help
```

如果需要指定业务域，可以设置默认值或在单条命令中使用 `-bd`：

```bash
export KWEAVER_BUSINESS_DOMAIN="${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
kweaver config show
```

### 文档索引

- [安装和运行](#cli-安装和运行)
- [快速开始](#cli-快速开始)
- [Agent 生命周期命令](#agent-生命周期命令)
- [对话与 Trace](#cli-对话与-trace)
- [本地调试流程](#cli-本地调试流程)


<!-- 来源：install-and-run.md -->

## CLI 安装和运行

此处统一说明 KWeaver CLI 的安装、运行、认证、通用环境变量和本地调试连接方式。后续章节默认已经可以直接使用 `kweaver` 命令。

### 环境要求

- Node.js 22 或更高版本。
- 可访问 Decision Agent 服务地址，例如本地 `http://127.0.0.1:13020`。
- 如果需要创建 Agent，请提前准备可用的 LLM ID 和 LLM 名称。

### 使用远程 npm 仓库安装

```bash
npm install -g @kweaver-ai/kweaver-sdk
```

安装后检查入口：

```bash
kweaver --help
kweaver agent --help
```

### 使用本地下载的 SDK 项目安装

如果 SDK 项目已经下载到本地，可以从本地目录安装 CLI：

```bash
npm install -g <path-to-kweaver-sdk>/packages/typescript
```

安装完成后，使用方式仍然是：

```bash
kweaver --help
kweaver agent --help
```

### 本地 no-auth 环境

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

### 通用环境变量

| 环境变量 | 说明 |
| --- | --- |
| `KWEAVER_BASE_URL` | Decision Agent 服务地址。 |
| `KWEAVER_NO_AUTH` | 本地 no-auth 模式下设置为 `1`，CLI 会省略鉴权 header。 |
| `KWEAVER_TOKEN` | 鉴权环境中的 access token。 |
| `KWEAVER_BUSINESS_DOMAIN` | 默认业务域。支持命令级 `-bd` 的子命令可覆盖该值。 |
| `KWEAVER_LLM_ID` | 示例创建 Agent 时使用的大模型 ID。 |
| `KWEAVER_LLM_NAME` | 示例创建 Agent 时使用的大模型名称。 |

### 通用参数

`kweaver agent` 下不同子命令支持的参数并不完全相同。常见参数如下，具体可用范围请以对应子命令章节为准：

| 参数 | 说明 |
| --- | --- |
| `-bd, --biz-domain <value>` | 覆盖业务域。当前主要用于列表、详情、模板、对话、会话和 skill 相关命令。 |
| `--pretty` | 以缩进 JSON 输出。大多数查询命令默认开启。 |
| `--compact` | 以紧凑 JSON 输出。主要用于 `sessions`、`history`、`trace`、`skill list`。 |
| `--verbose, -v` | 输出完整响应或调试请求信息。 |
| `--help, -h` | 查看命令帮助。 |

部分命令的 `--help` 中可能保留历史兼容参数。本文档按当前 CLI 实际解析和请求行为说明；如果服务需要业务域，优先使用 `kweaver config set-bd <value>` 或 `KWEAVER_BUSINESS_DOMAIN`。

### 鉴权环境

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

### 示例 Makefile 目标

用户指南的可运行示例位于 [../examples/cli](../examples/cli/README.md)。示例优先使用以下目标：

| 目标 | 含义 |
| --- | --- |
| `make quick-check` | 快速检查，不创建或删除数据，通常只运行 help 和列表类查询。 |
| `make flow` | 完整流程，会创建、查询、更新、发布、取消发布并删除临时 Agent。 |
| `make smoke` | `quick-check` 的兼容别名。 |

### 准备 Agent config

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


<!-- 来源：quick-start.md -->

## CLI 快速开始

本模块用 CLI 跑通创建、详情、发布、对话。请先完成 [CLI 安装和运行](#cli-安装和运行)，并准备 `AGENT_CONFIG`。

### 1. 创建 Agent

```bash
export AGENT_NAME="doc_cli_agent_$(date +%Y%m%d%H%M%S)"

kweaver agent create \
  --name "$AGENT_NAME" \
  --profile "Decision Agent CLI user guide local example" \
  --product-key dip \
  --config "$AGENT_CONFIG" \
  --pretty | tee /tmp/kweaver-cli-agent-create.json

export AGENT_ID="$(jq -r '.id' /tmp/kweaver-cli-agent-create.json)"
echo "$AGENT_ID"
```

### 2. 获取详情

```bash
test -n "$AGENT_ID"

kweaver agent get "$AGENT_ID" \
  --pretty | tee /tmp/kweaver-cli-agent-detail.json
```

### 3. 发布

```bash
test -n "$AGENT_ID"

kweaver agent publish "$AGENT_ID" \
  | tee /tmp/kweaver-cli-agent-publish.json
```

### 4. 非流式对话

发布后按版本对话。如果还没有发布版本，可先使用 `--version v0` 测试未发布版本。

```bash
test -n "$AGENT_ID"
export AGENT_VERSION="${AGENT_VERSION:-v0}"

kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  -m "请用一句话介绍你自己" \
  --no-stream \
  --verbose | tee /tmp/kweaver-cli-chat.json
```

### 5. 对话列表

```bash
kweaver agent sessions "$AGENT_ID" \
  --limit 10 \
  --pretty
```

`sessions` 是 CLI 的历史命令名，这里查询的是当前 Agent 的对话列表。

### 6. 清理

```bash
test -n "$AGENT_ID"

kweaver agent unpublish "$AGENT_ID" || true

kweaver agent delete "$AGENT_ID" -y
```


<!-- 来源：agent-lifecycle.md -->

## Agent 生命周期命令

请先完成 [CLI 安装和运行](#cli-安装和运行)。本页覆盖 Agent 管理、个人空间、模板、发布和技能绑定相关子命令。完整可运行脚本见 [../examples/cli](../examples/cli/README.md)。

### `kweaver agent list`

用途：查询已发布到广场的 Agent 列表。

```bash
kweaver agent list --name "" --limit 10 --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--name <text>` | 否 | 空字符串 | 按名称过滤。 |
| `--offset <n>` | 否 | `0` | 分页偏移量。小于 0 或非法值会回退为 `0`。 |
| `--limit <n>` | 否 | `30` | 返回条数。小于 1 或非法值会回退为 `30`。 |
| `--category-id <id>` | 否 | 空字符串 | 按分类过滤。 |
| `--custom-space-id <id>` | 否 | 空字符串 | 按自定义空间过滤。 |
| `--is-to-square <0\|1>` | 否 | `1` | 是否查询广场 Agent。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |
| `--simple` | 否 | 无 | 兼容参数，当前不会改变输出。 |

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v3/published/agent" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}" \
  -H "Content-Type: text/plain;charset=UTF-8" \
  -d '{"offset":0,"limit":10,"category_id":"","name":"","custom_space_id":"","is_to_square":1}'
```

对应示例：`make -C docs/user_manual/examples/cli list`。

### `kweaver agent personal-list`

用途：查询个人空间中的 Agent 列表。

```bash
kweaver agent personal-list --size 10 --publish-status "" --publish-to-be "" --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--name <text>` | 否 | 空字符串 | 按名称过滤。 |
| `--pagination-marker <str>` | 否 | 空字符串 | 下一页分页游标。 |
| `--publish-status <status>` | 否 | 空字符串 | 发布状态过滤，可用 `unpublished`、`published`、`published_edited`。 |
| `--publish-to-be <value>` | 否 | 空字符串 | 发布类型过滤，例如 `skill_agent`、`api_agent`。 |
| `--size <n>` | 否 | `48` | 返回条数。小于 1 或非法值会回退为 `48`。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/personal-space/agent-list?size=10" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`make -C docs/user_manual/examples/cli personal-list`。

### `kweaver agent category-list`

用途：查询 Agent 分类。

```bash
kweaver agent category-list --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/category" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`make -C docs/user_manual/examples/cli category-list`。

### `kweaver agent template-list`

用途：查询已发布 Agent 模板列表。

```bash
kweaver agent template-list --name "" --size 10 --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--category-id <id>` | 否 | 空字符串 | 按分类过滤。 |
| `--name <text>` | 否 | 空字符串 | 按模板名称过滤。 |
| `--pagination-marker <str>` | 否 | 空字符串 | 下一页分页游标。 |
| `--size <n>` | 否 | `48` | 返回条数。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/published/agent-tpl?size=10" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`make -C docs/user_manual/examples/cli template-list`。

### `kweaver agent template-get`

用途：获取已发布 Agent 模板详情，可把模板 `config` 保存到本地文件。

```bash
kweaver agent template-get "$TEMPLATE_ID" --save-config /tmp/kweaver-template-config.json --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<template_id>` | 是 | 无 | 模板 ID。 |
| `--save-config <path>` | 否 | 无 | 保存模板 config，CLI 会在文件名中追加时间戳。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/published/agent-tpl/$TEMPLATE_ID" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`TEMPLATE_ID=<template-id> make -C docs/user_manual/examples/cli detail-examples`。

### `kweaver agent get`

用途：按 `agent_id` 获取 Agent 详情。

```bash
kweaver agent get "$AGENT_ID" --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `--save-config <path>` | 否 | 无 | 保存当前 Agent config，CLI 会在文件名中追加时间戳。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应；默认只输出摘要。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli detail-examples`。

### `kweaver agent get-by-key`

用途：按 `agent_key` 获取 Agent 信息。

```bash
kweaver agent get-by-key "$AGENT_KEY"
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<key>` | 是 | 无 | Agent key。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/by-key/$AGENT_KEY" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_KEY=<agent-key> make -C docs/user_manual/examples/cli detail-examples`。

### `kweaver agent create`

用途：创建 Agent。推荐通过 `--config <path>` 传入完整 config。

```bash
kweaver agent create \
  --name "$AGENT_NAME" \
  --profile "Created from CLI lifecycle doc" \
  --product-key dip \
  --mode default \
  --config "$AGENT_CONFIG" \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--name <text>` | 是 | 无 | Agent 名称。 |
| `--profile <text>` | 是 | 无 | Agent 描述。 |
| `--config <json\|path>` | 否 | 无 | 完整 config JSON 字符串或文件路径；推荐使用文件路径。 |
| `--key <text>` | 否 | 自动生成 | Agent 唯一 key。 |
| `--product-key <text>` | 否 | `dip` | 产品 key。 |
| `--system-prompt <text>` | 否 | 空字符串 | 未使用 `--config` 时写入 config 的系统提示词。 |
| `--llm-id <id>` | 条件必填 | 无 | 未使用 `--config` 时用于生成默认 LLM 配置。 |
| `--llm-max-tokens <n>` | 否 | `4096` | 未使用 `--config` 时的大模型最大 token。 |
| `--mode <mode>` | 否 | `default` | Agent 模式：`default`、`dolphin`、`react`。会覆盖 `config.mode`。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v3/agent" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}" \
  -H "Content-Type: application/json" \
  -d '{"name":"...","profile":"...","product_key":"dip","config":{"mode":"default","llms":[{"is_default":true,"llm_config":{"id":"...","name":"..."}}]}}'
```

对应示例：`make -C docs/user_manual/examples/cli create`。

### `kweaver agent update`

用途：更新已有 Agent 的基础信息或 config。修改已发布 Agent 后，个人空间通常会变为“已发布编辑中”，需要重新发布才会生成新的发布版本。

```bash
kweaver agent update "$AGENT_ID" \
  --name "${AGENT_NAME}_updated" \
  --profile "Updated from CLI lifecycle doc" \
  --mode default
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `--name <text>` | 否 | 保持原值 | 更新 Agent 名称。 |
| `--profile <text>` | 否 | 保持原值 | 更新 Agent 描述。 |
| `--system-prompt <text>` | 否 | 保持原值 | 更新 `config.system_prompt`。 |
| `--knowledge-network-id <id>` | 否 | 无 | 将 `config.data_source.knowledge_network` 设置为只包含该知识网络 ID。 |
| `--config-path <path>` | 否 | 无 | 从文件读取完整 config。 |
| `--mode <mode>` | 否 | 原 config 或 `default` | Agent 模式：`default`、`dolphin`、`react`。 |

背后 API：

```bash
curl -X PUT "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"...","profile":"...","product_key":"dip","config":{"mode":"default"}}'
```

知识网络绑定是可选场景，本地默认不运行：

```bash
AGENT_ID=<agent-id> KN_ID=<knowledge-network-id> make -C docs/user_manual/examples/cli update-knowledge-network
```

基础信息更新示例：

```bash
make -C docs/user_manual/examples/cli update
```

### `kweaver agent publish`

用途：发布 Agent。再次发布会生成新版本；只调整发布元信息请使用 API 的更新发布信息能力。

```bash
kweaver agent publish "$AGENT_ID"
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `--category-id <id>` | 否 | 空数组 | 发布分类。 |

当前 CLI 发布默认请求体会发布到广场，并发布为 `skill_agent`。如果需要更精细地选择发布类型，请使用 REST API 或 SDK。

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/publish" \
  -H "Content-Type: application/json" \
  -d '{"category_ids":[],"description":"","publish_to_where":["square"],"publish_to_bes":["skill_agent"],"pms_control":null}'
```

对应示例：`make -C docs/user_manual/examples/cli publish`。

### `kweaver agent unpublish`

用途：取消发布 Agent。

```bash
kweaver agent unpublish "$AGENT_ID"
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |

背后 API：

```bash
curl -X PUT "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/unpublish"
```

对应示例：`make -C docs/user_manual/examples/cli unpublish`。

### `kweaver agent delete`

用途：删除 Agent。

```bash
kweaver agent delete "$AGENT_ID" -y
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `-y, --yes` | 否 | 关闭 | 跳过删除确认。 |

背后 API：

```bash
curl -X DELETE "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID"
```

对应示例：`make -C docs/user_manual/examples/cli delete`。

### `kweaver agent skill`

用途：管理 Agent config 中绑定的 skill。默认本地环境没有 `SKILL_ID`，相关示例不会进入 `quick-check`。

```bash
kweaver agent skill list "$AGENT_ID"
kweaver agent skill add "$AGENT_ID" "$SKILL_ID" --strict
kweaver agent skill remove "$AGENT_ID" "$SKILL_ID"
```

参数：

| 命令 | 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `skill list` | `<agent-id>` | 是 | 无 | Agent ID。 |
| `skill list` | `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `skill list` | `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |
| `skill list` | `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `skill add` | `<agent-id>` | 是 | 无 | Agent ID。 |
| `skill add` | `<skill-id>...` | 是 | 无 | 一个或多个 skill ID。 |
| `skill add` | `--strict` | 否 | 关闭 | 如果 skill 未发布则报错；默认只警告并继续。 |
| `skill add` | `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `skill remove` | `<agent-id>` | 是 | 无 | Agent ID。 |
| `skill remove` | `<skill-id>...` | 是 | 无 | 一个或多个 skill ID。 |
| `skill remove` | `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |

背后 API：CLI 会先 `GET /api/agent-factory/v3/agent/{agent_id}` 读取当前 config，然后修改 `config.skills.skills`，最后 `PUT /api/agent-factory/v3/agent/{agent_id}` 写回。

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID"
curl -X PUT "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"...","profile":"...","product_key":"dip","config":{"skills":{"skills":[{"skill_id":"..."}]}}}'
```

条件示例：

```bash
AGENT_ID=<agent-id> SKILL_ID=<skill-id> make -C docs/user_manual/examples/cli skill-examples
```


<!-- 来源：chat-session-trace.md -->

## CLI 对话与 Trace

请先完成 [CLI 安装和运行](#cli-安装和运行)，并准备 `AGENT_ID`。完整可运行脚本见 [../examples/cli](../examples/cli/README.md)。

### `kweaver agent chat`

用途：与 Agent 对话。省略 `-m/--message` 时进入交互式模式；带 `-m` 时发送单轮非交互消息。

```bash
test -n "$AGENT_ID"
export AGENT_VERSION="${AGENT_VERSION:-v0}"

kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  -m "请用一句话介绍你自己" \
  --no-stream \
  --verbose | tee /tmp/kweaver-agent-chat.json
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。CLI 会先解析对应版本的 `agent_key`。 |
| `-m, --message <text>` | 否 | 无 | 单轮消息；省略时进入交互式对话。 |
| `--version <value>` / `-version <value>` | 否 | `v0` | 用于解析 `agent_key` 的 Agent 版本。 |
| `--conversation-id <id>` | 否 | 无 | 继续已有对话。 |
| `-cid <id>` | 否 | 无 | `--conversation-id` 短别名。 |
| `--conversation_id <id>` | 否 | 无 | 兼容别名。 |
| `-conversation-id <id>` | 否 | 无 | 兼容别名。 |
| `-conversation_id <id>` | 否 | 无 | 兼容别名。 |
| `--session-id <id>` | 否 | 无 | 历史兼容别名，含义同 `conversation_id`。 |
| `--stream` | 否 | 交互式默认开启 | 启用流式输出。 |
| `--no-stream` | 否 | 非交互默认关闭流式 | 禁用流式输出。 |
| `--tui` | 否 | 关闭 | 启动实验性 OpenTUI 界面，交互式使用。 |
| `--verbose, -v, --debug` | 否 | 关闭 | 输出请求细节或调试信息。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |

背后 API：CLI 先通过广场详情接口解析 `agent_key`，再调用对话接口。

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent-market/agent/$AGENT_ID/version/$AGENT_VERSION?is_visit=true"

curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d '{"query":"请用一句话介绍你自己","conversation_id":"","stream":false}'
```

对应示例：

```bash
AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli chat-examples
```

### 继续对话

从上一步输出或 `sessions` 命令中取得 `conversation_id`：

```bash
: "${CONVERSATION_ID:?请先从上一步 chat 输出或 sessions 结果导出 CONVERSATION_ID}"

kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  --conversation-id "$CONVERSATION_ID" \
  -m "请继续补充一个使用建议" \
  --no-stream
```

### 流式输出

```bash
kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  -m "请分三点说明你的能力" \
  --stream
```

### 交互式对话

```bash
kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION"
```

退出方式：输入 `exit`、`quit` 或 `q`。

### `kweaver agent sessions`

用途：查询 Agent 的对话列表。`sessions` 是 CLI 的历史命令名，这里查询的是对话列表，不是底层运行期 session。

```bash
kweaver agent sessions "$AGENT_ID" \
  --limit 10 \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。CLI 会先解析 `agent_key`。 |
| `--limit <n>` | 否 | `30` | 返回对话数。小于 1 或非法值会回退为 `30`。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation?page=1&size=10" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli chat-examples`。

### `kweaver agent history`

用途：查询某个对话的消息历史。

```bash
test -n "$CONVERSATION_ID"

kweaver agent history "$AGENT_ID" "$CONVERSATION_ID" \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 条件必填 | 无 | 推荐传入。用于解析 `agent_key`。 |
| `<conversation_id>` | 是 | 无 | 对话 ID。也兼容只传 `conversation_id` 的历史形式。 |
| `--limit <n>` | 否 | `30` | 兼容参数；当前消息详情接口不依赖该参数。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_ID=<agent-id> CONVERSATION_ID=<conversation-id> make -C docs/user_manual/examples/cli chat-examples`。

### `kweaver agent trace`

用途：查看一次对话中的执行链路、工具调用与中间进度。这个能力依赖当前环境暴露的 trace 路由；如果后端未开放对应路由（或 trace-ai 服务不可用），CLI 会返回 404，此时可先使用 `history` 查看对话消息。

```bash
test -n "$CONVERSATION_ID"

kweaver agent trace "$AGENT_ID" "$CONVERSATION_ID" \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `<conversation_id>` | 是 | 无 | 对话 ID。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v1/observability/agent/$AGENT_ID/conversation/$CONVERSATION_ID/session" \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"'"$AGENT_ID"'","start_time":1,"end_time":<now_plus_one_day_ms>,"page":1,"size":50}'
```

对应示例：`AGENT_ID=<agent-id> CONVERSATION_ID=<conversation-id> make -C docs/user_manual/examples/cli chat-examples`。


<!-- 来源：local-debug-flow.md -->

## CLI 本地调试流程

本页适合 SDK/CLI 开发者在本地源码目录中验证 CLI 与本地 Decision Agent 服务的联动。普通 CLI 用户请优先阅读 [CLI 安装和运行](#cli-安装和运行)，并直接使用 `kweaver ...` 命令。

### 健康检查

```bash
cd <path-to-kweaver-sdk>/packages/typescript

curl -fsS "${KWEAVER_BASE_URL:-http://127.0.0.1:13020}/health/ready"
```

### 本地源码入口

```bash
node --import tsx src/cli.ts agent --help
node --import tsx src/cli.ts agent create --help
node --import tsx src/cli.ts agent chat --help
```


### 构建后入口

```bash
npm run build
node bin/kweaver.js agent --help
node bin/kweaver.js agent chat --help
```

如果需要模拟全局安装后的用户体验，可以在用户项目或临时目录中安装本地包：

```bash
npm install -g <path-to-kweaver-sdk>/packages/typescript
kweaver agent --help
```
