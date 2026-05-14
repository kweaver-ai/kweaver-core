<!-- 请勿直接编辑：本文件由 docs/user_manual/api 下的 `make aggregate` 生成。 -->
<!-- 来源文件：index.md, agent-lifecycle.md, agent-config.md, agent-config-examples.md, publish-and-market.md, chat-and-conversation.md, chat-response.md, incremental-streaming.md, intervention-termination.md, debug-chat.md, conversation-session-run.md, import-export-and-auxiliary.md。 -->

# Decision Agent REST 接入指南（聚合版）

> 本文件由脚本生成，请不要直接修改本文件；如需调整内容，请修改分文件文档后运行 `make -C docs/user_manual/api aggregate`。

## 目录

- [Decision Agent REST 接入指南](#decision-agent-rest-接入指南)
- [Agent 生命周期 API](#agent-生命周期-api)
- [Agent 配置 API](#agent-配置-api)
- [Agent config 示例](#agent-config-示例)
- [发布与广场 API](#发布与广场-api)
- [对话 API](#对话-api)
- [对话响应](#对话响应)
- [增量流式](#增量流式)
- [人工干预与终止](#人工干预与终止)
- [Debug 对话](#debug-对话)
- [对话、会话与执行](#对话会话与执行)
- [导入导出与辅助 API](#导入导出与辅助-api)


<!-- 来源：index.md -->

## Decision Agent REST 接入指南

Decision Agent REST 接入指南适合需要查看底层 REST 协议、排障或做自定义集成的开发者。业务接入优先使用 SDK；当前重点是 TypeScript SDK，后续会继续补充 Python 和 Go SDK 文档。示例默认从 `decision-agent` 目录运行。

如果需要先统一产品概念和术语，请阅读 [Agent 概念指南](../concepts/README.md)。

### 整体流程

一个典型接入场景可以这样串起来：先创建一个 Agent，并用最小化 `config` 跑通保存；随后按业务需要更新 `config`，例如绑定知识网络或调整 LLM；确认配置可用后发布到广场；发布后通过广场详情接口获取 `agent_key`；业务系统使用 `agent_key` 发起非流式、普通流式或增量流式对话；遇到配置页调试或排障场景时使用 Debug 对话；本地验证完成后可删除测试 Agent。

推荐阅读顺序：

1. [Agent 生命周期](#agent-生命周期-api)：创建、查看、更新和删除 Agent。
2. [Agent 配置](#agent-配置-api) 与 [Agent config 示例](#agent-config-示例)：准备可创建、可更新的 `config`。
3. [发布与广场](#发布与广场-api)：发布 Agent 并解析 `agent_key`。
4. [对话](#对话-api)、[对话响应](#对话响应) 与 [增量流式](#增量流式)：完成实际调用并理解返回结构。
5. [人工干预与终止](#人工干预与终止)：处理人工确认、中断恢复、终止执行和断线续连。
6. [Debug 对话](#debug-对话) 和 [对话、会话与执行](#对话会话与执行)：排障和理解运行期概念。

配套可运行示例见 [../examples/api](../examples/api/README.md)。示例使用 Shell + cURL，并提供 `make quick-check` 与 `make flow`。

### 概念关系

| 英文字段/概念 | 中文名称 | 说明 |
| --- | --- | --- |
| `conversation` / `conversation_id` | 对话 | 面向用户的多轮上下文容器。继续同一个 `conversation_id` 会复用历史上下文。 |
| `session` / `conversation_session_id` | 会话 | 运行期会话，主要用于缓存、续连和临时状态。普通 REST 对接通常可以忽略。 |
| `run` / `agent_run_id` | 一次执行 | 一次 Agent 接口调用或一次 Agent 执行过程，用于追踪、中断、恢复和观测。 |

一个对话可以对应多个会话，一个会话可以包含一个或多个 run。文档后续统一使用“对话”对应 `conversation`，“会话”对应 `session`，“一次执行”对应 `run`。

### 最小化 no-auth 环境

本地 no-auth 调试时，只需要服务地址和基础请求函数：

```bash
cd <path-to-decision-agent>
export AF_BASE_URL="${AF_BASE_URL:-http://127.0.0.1:13020}"

af_curl() {
  curl -fsS -H "X-Language: zh-cn" "$@"
}

af_curl "$AF_BASE_URL/health/ready"
```

后续 API 示例都假设当前 shell 已定义 `af_curl`。

### 完整环境

如果部署开启鉴权或业务域校验，可以使用完整环境：

```bash
cd <path-to-decision-agent>
export AF_BASE_URL="${AF_BASE_URL:-http://127.0.0.1:13020}"
export AF_BD="${AF_BD:-bd_public}"
export AF_TOKEN="${AF_TOKEN:-${KWEAVER_TOKEN:-__NO_AUTH__}}"

af_curl() {
  if [ "$AF_TOKEN" = "__NO_AUTH__" ]; then
    curl -fsS -H "X-Business-Domain: $AF_BD" -H "X-Language: zh-cn" "$@"
  else
    curl -fsS \
      -H "Authorization: Bearer $AF_TOKEN" \
      -H "Token: $AF_TOKEN" \
      -H "X-Business-Domain: $AF_BD" \
      -H "X-Language: zh-cn" \
      "$@"
  fi
}

af_curl "$AF_BASE_URL/health/ready"
```

### 文档索引

- [Agent 生命周期](#agent-生命周期-api)
- [Agent 配置](#agent-配置-api)
- [Agent config 示例](#agent-config-示例)
- [发布与广场](#发布与广场-api)
- [对话](#对话-api)
- [对话响应](#对话响应)
- [增量流式](#增量流式)
- [人工干预与终止](#人工干预与终止)
- [Debug 对话](#debug-对话)
- [对话、会话与执行](#对话会话与执行)
- [导入导出与辅助接口](#导入导出与辅助-api)

### API 版本

| 前缀 | 主要用途 |
| --- | --- |
| `/api/agent-factory/v3` | Agent 管理、发布、个人空间、广场、权限等管理接口。 |
| `/api/agent-factory/v1` | 对话执行、会话管理、run 管理等运行接口。 |


<!-- 来源：agent-lifecycle.md -->

## Agent 生命周期 API

本页覆盖创建、查看、更新、删除 Agent 的核心 API。请先加载 [REST 接入环境](./index.md#最小化-no-auth-环境)。

### 创建 Agent

接口：

```text
POST /api/agent-factory/v3/agent
```

可运行示例。请先按 [Agent config 示例](#agent-config-示例) 准备 `/tmp/agent-config-minimal.json`：

```bash
export AGENT_NAME="doc_api_agent_$(date +%Y%m%d%H%M%S)"

jq -n \
  --arg name "$AGENT_NAME" \
  --arg profile "Decision Agent API user guide local example" \
  --argjson config "$(cat /tmp/agent-config-minimal.json)" \
  '{
    name: $name,
    profile: $profile,
    avatar_type: 1,
    avatar: "icon-dip-agent-default",
    product_key: "DIP",
    config: $config
  }' > /tmp/agent-create.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent" \
  -H "Content-Type: application/json" \
  -d @/tmp/agent-create.json | tee /tmp/agent-create-response.json

export AGENT_ID="$(jq -r '.id' /tmp/agent-create-response.json)"
echo "$AGENT_ID"
```

关键参数：

| 字段 | 说明 |
| --- | --- |
| `name` | Agent 名称，建议不超过 50 字符。 |
| `profile` | Agent 描述，建议说明用途。 |
| `product_key` | 产品标识，本地示例使用 `DIP`。 |
| `config` | Agent 运行配置，详见 [Agent config 示例](#agent-config-示例)。 |

### 获取详情

接口：

```text
GET /api/agent-factory/v3/agent/{agent_id}
GET /api/agent-factory/v3/agent/by-key/{key}
```

示例：

```bash
test -n "$AGENT_ID"
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" | jq .
```

### 更新 Agent

接口：

```text
PUT /api/agent-factory/v3/agent/{agent_id}
```

更新接口通常需要带完整配置。最稳妥方式是先读取详情，再改需要的字段：

```bash
test -n "$AGENT_ID"

af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  | jq '.name = (.name + "_updated") | .profile = "Updated by local API example"' \
  > /tmp/agent-update.json

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  -H "Content-Type: application/json" \
  -d @/tmp/agent-update.json | jq .
```

### 删除 Agent

接口：

```text
DELETE /api/agent-factory/v3/agent/{agent_id}
```

示例：

```bash
test -n "$AGENT_ID"
af_curl -X DELETE "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID"
```

删除是破坏性操作。建议只对本地示例创建的 Agent 执行。


<!-- 来源：agent-config.md -->

## Agent 配置 API

Agent 的 `config` 决定输入输出、LLM、技能、知识网络、记忆、历史策略和运行模式。创建与更新时，`config` 是最容易出错的部分。

请先阅读 [Agent config 示例](#agent-config-示例)，按场景准备一个配置文件，例如 `/tmp/agent-config-minimal.json` 或 `/tmp/agent-config-full.json`。Agent 模式的产品含义可参考 [Agent 模式](../concepts/agent-modes.md)。

### 最小可用要求

| 配置 | 要求 |
| --- | --- |
| `input` | 必须存在，并包含输入字段定义。 |
| `output` | 必须存在，并定义默认输出格式或变量。 |
| `llms` | 至少一个 `is_default: true` 的 LLM。 |
| `mode` | 可选，支持 `default`、`dolphin`、`react`；缺省时按默认模式处理。 |
| `react_config` | 仅允许 `mode=react` 时传入。 |

### 设置 Agent Mode

默认模式：

```bash
jq '.mode = "default"' /tmp/agent-config-minimal.json > /tmp/agent-config-default.json
```

React 模式：

```bash
jq '.mode = "react" | .react_config = {
  disable_history_in_a_conversation: false,
  disable_llm_cache: false
}' /tmp/agent-config-minimal.json > /tmp/agent-config-react.json
```

创建 React Agent：

```bash
export AGENT_NAME="doc_api_react_agent_$(date +%Y%m%d%H%M%S)"

jq -n \
  --arg name "$AGENT_NAME" \
  --argjson config "$(cat /tmp/agent-config-react.json)" \
  '{
    name: $name,
    profile: "React mode local example",
    avatar_type: 1,
    avatar: "icon-dip-agent-default",
    product_key: "DIP",
    config: $config
  }' > /tmp/agent-create-react.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent/react" \
  -H "Content-Type: application/json" \
  -d @/tmp/agent-create-react.json | jq .
```

### 配置知识网络

如果已有知识网络 ID，可以写入 `config.data_source.knowledge_network`：

```bash
: "${KN_ID:?请先导出要绑定的知识网络 ID 到 KN_ID}"

jq --arg kn "$KN_ID" '
  .data_source.knowledge_network = [
    {
      knowledge_network_id: $kn
    }
  ]
' /tmp/agent-config-minimal.json > /tmp/agent-config-with-kn.json
```

### 常见错误

| 错误 | 原因 | 处理 |
| --- | --- | --- |
| `react_config is only allowed when mode is react` | 非 React 模式传入了 `react_config` | 删除 `react_config` 或设置 `mode=react`。 |
| `mode is invalid` | `mode` 不在允许枚举内 | 使用 `default`、`dolphin`、`react`。 |
| LLM 缺失 | `llms` 为空或没有默认 LLM | 补充 `is_default: true` 的 LLM。 |


<!-- 来源：agent-config-examples.md -->

## Agent config 示例

Agent 的 `config` 是创建和更新 Agent 时最重要的字段。本页给出两个可直接保存为 JSON 文件的示例：最小化配置和较复杂配置。

### 最小化配置

适合本地 no-auth 快速验证或只验证 Agent 创建链路。保存为 `/tmp/agent-config-minimal.json`：

```json
{
  "input": {
    "fields": [
      {
        "name": "query",
        "type": "string"
      }
    ]
  },
  "output": {
    "default_format": "markdown",
    "variables": {
      "answer_var": "answer"
    }
  },
  "llms": [
    {
      "is_default": true,
      "llm_config": {
        "id": "your-llm-id",
        "name": "your-llm-name",
        "model_type": "llm",
        "temperature": 1,
        "top_p": 1,
        "top_k": 1,
        "frequency_penalty": 0,
        "presence_penalty": 0,
        "max_tokens": 1000,
        "retrieval_max_tokens": 32
      }
    }
  ],
  "memory": {
    "is_enabled": false
  },
  "related_question": {
    "is_enabled": false
  },
  "plan_mode": {
    "is_enabled": false
  }
}
```

### 较复杂配置

适合需要系统提示词、知识网络、技能占位和 React 模式开关的场景。保存为 `/tmp/agent-config-full.json`：

```json
{
  "mode": "react",
  "input": {
    "fields": [
      {
        "name": "query",
        "type": "string",
        "desc": "用户问题"
      },
      {
        "name": "history",
        "type": "object",
        "desc": "历史上下文"
      }
    ],
    "rewrite": {
      "enable": false,
      "llm_config": {
        "id": "your-llm-id",
        "name": "your-llm-name",
        "model_type": "llm",
        "temperature": 0.5,
        "top_p": 0.8,
        "top_k": 1,
        "frequency_penalty": 0,
        "presence_penalty": 0,
        "max_tokens": 1000,
        "retrieval_max_tokens": 1000
      }
    },
    "augment": {
      "enable": false,
      "data_source": {}
    }
  },
  "system_prompt": "你是一个专业的业务助手，请基于用户问题给出清晰、可执行的回答。",
  "react_config": {
    "disable_history_in_a_conversation": false,
    "disable_llm_cache": false
  },
  "data_source": {
    "knowledge_network": [
      {
        "knowledge_network_id": "your-knowledge-network-id"
      }
    ]
  },
  "skills": {
    "tools": [],
    "agents": [],
    "mcps": []
  },
  "llms": [
    {
      "is_default": true,
      "llm_config": {
        "id": "your-llm-id",
        "name": "your-llm-name",
        "model_type": "llm",
        "temperature": 1,
        "top_p": 1,
        "top_k": 1,
        "frequency_penalty": 0,
        "presence_penalty": 0,
        "max_tokens": 1000,
        "retrieval_max_tokens": 32
      }
    }
  ],
  "output": {
    "default_format": "markdown",
    "variables": {
      "answer_var": "answer"
    }
  },
  "preset_questions": [
    {
      "question": "请介绍一下你的能力"
    }
  ],
  "memory": {
    "is_enabled": false
  },
  "related_question": {
    "is_enabled": false
  },
  "plan_mode": {
    "is_enabled": false
  },
  "conversation_history_config": {
    "strategy": "count",
    "count_params": {
      "round_limit": 5
    }
  }
}
```

### 参数说明

| 字段 | 说明 |
| --- | --- |
| `input.fields` | Agent 入参定义。常用 `query` 表示用户问题；字段类型支持 `string`、`object`、`file`。 |
| `output.variables.answer_var` | 回答变量名，Dolphin 模式下尤其重要。 |
| `llms` | LLM 配置列表，至少需要一个 `is_default=true` 的模型。 |
| `mode` | Agent 模式。可用值包括 `default`、`dolphin`、`react`；不传时按默认模式处理。 |
| `system_prompt` | 系统提示词，用于约束 Agent 的角色、语气和回答边界。 |
| `react_config` | React 模式专属配置，只能在 `mode=react` 时传入。 |
| `data_source.knowledge_network` | 绑定的知识网络列表。公开对接只推荐配置 `knowledge_network_id`；没有知识网络时可省略或传空数组。 |
| `skills` | 技能、子 Agent、MCP 配置。没有时可传空数组。 |
| `preset_questions` | 预设问题列表，用于前端或产品展示。 |
| `memory` | 记忆开关。简单示例通常关闭。 |
| `related_question` | 相关问题推荐开关。简单示例通常关闭。 |
| `plan_mode` | 规划模式开关；不能与 Dolphin 模式同时启用。 |
| `conversation_history_config` | 对话历史策略配置，可用于控制历史消息数量、轮数、时间窗口或 token 策略。 |


<!-- 来源：publish-and-market.md -->

## 发布与广场 API

Agent 创建后需要发布，才能按发布版本对外使用。不发布，自己是可以使用的，但是不能在广场等地方展示，别人也用不了。

发布、更新发布信息和重新发布的区别可参考 [发布逻辑](../concepts/publishing.md)。

### 发布 Agent

接口：

```text
POST /api/agent-factory/v3/agent/{agent_id}/publish
```

示例：

```bash
test -n "$AGENT_ID"

jq -n '{
  category_ids: [],
  description: "Published by local API example",
  publish_to_where: ["square"],
  publish_to_bes: ["skill_agent"],
  pms_control: null
}' > /tmp/agent-publish.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/publish" \
  -H "Content-Type: application/json" \
  -d @/tmp/agent-publish.json | tee /tmp/agent-publish-response.json
```

关键参数：

| 字段 | 说明 |
| --- | --- |
| `category_ids` | 发布分类，可为空。 |
| `publish_to_where` | 发布目标，常用 `square`。 |
| `publish_to_bes` | 发布形态，CLI 默认使用 `skill_agent`。 |
| `pms_control` | 权限控制配置，简单场景为 `null`。 |

### 获取发布信息

```bash
test -n "$AGENT_ID"
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/publish-info" | jq .
```

### 已发布 Agent 列表

接口：

```text
POST /api/agent-factory/v3/published/agent
```

示例：

```bash
jq -n '{
  offset: 0,
  limit: 10,
  is_to_square: 1
}' > /tmp/published-agent-list.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/published/agent" \
  -H "Content-Type: application/json" \
  -d @/tmp/published-agent-list.json | jq .
```

### 获取广场详情

对话前通常需要解析 `agent_id` 到 `agent_key`：

```bash
test -n "$AGENT_ID"
export AGENT_VERSION="${AGENT_VERSION:-v0}"

af_curl "$AF_BASE_URL/api/agent-factory/v3/agent-market/agent/$AGENT_ID/version/$AGENT_VERSION" \
  | tee /tmp/agent-market-detail.json

export AGENT_KEY="$(jq -r '.key' /tmp/agent-market-detail.json)"
echo "$AGENT_KEY"
```

这里的接口路径仍使用历史命名 `agent-market`，但产品和文档展示统一称为“广场”。

### 取消发布

```bash
test -n "$AGENT_ID"
af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/unpublish" | jq .
```


<!-- 来源：chat-and-conversation.md -->

## 对话 API

对话接口使用 `agent_key`，而不是直接使用 `agent_id`。如果手上只有 `agent_id`，先通过广场详情接口解析。

### 解析 agent_key

```bash
test -n "$AGENT_ID"
export AGENT_VERSION="${AGENT_VERSION:-v0}"

af_curl "$AF_BASE_URL/api/agent-factory/v3/agent-market/agent/$AGENT_ID/version/$AGENT_VERSION?is_visit=true" \
  | tee /tmp/agent-market-detail.json

export AGENT_KEY="$(jq -r '.key' /tmp/agent-market-detail.json)"
test -n "$AGENT_KEY"
```

### 非流式对话

接口：

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/completion
```

示例：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "请用一句话介绍你自己" \
  '{
  agent_id: $agent_id,
  agent_key: $agent_key,
  agent_version: $agent_version,
  query: $query,
  stream: false
}' > /tmp/chat-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-request.json | tee /tmp/chat-response.json

export CONVERSATION_ID="$(jq -r '.. | objects | .conversation_id? // empty' /tmp/chat-response.json | head -n 1)"
echo "$CONVERSATION_ID"
```

响应字段说明见 [对话响应](#对话响应)。

### 普通流式对话

普通流式对话返回 SSE。每个 `data:` 片段都是当前完整响应快照，适合希望实时展示但不想维护增量合并逻辑的客户端：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "请流式介绍你的能力" \
  '{
  agent_id: $agent_id,
  agent_key: $agent_key,
  agent_version: $agent_version,
  query: $query,
  stream: true,
  inc_stream: false
}' > /tmp/chat-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d @/tmp/chat-stream-request.json
```

增量流式对话请阅读 [增量流式](#增量流式)。如果流式执行中断开连接但服务端 run 仍在继续，可以通过 [人工干预与终止](#人工干预与终止) 中的断线续连说明继续读取。

### 继续对话

```bash
test -n "$CONVERSATION_ID"

jq -n \
  --arg query "请继续补充一个使用建议" \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg conversation_id "$CONVERSATION_ID" \
  '{
    agent_id: $agent_id,
    agent_key: $agent_key,
    agent_version: $agent_version,
    query: $query,
    conversation_id: $conversation_id,
    stream: false
  }' > /tmp/chat-followup-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-followup-request.json | jq .
```

### 对话列表

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation?page=1&size=10" | jq .
```

### 对话详情

对话详情中会包含该对话下的消息列表。

```bash
test -n "$CONVERSATION_ID"

af_curl "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID" | jq .
```

### 标记已读

```bash
test -n "$CONVERSATION_ID"

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID/mark_read" \
  -H "Content-Type: application/json" \
  -d '{"latest_read_index": 2}' | jq .
```

Debug 对话、人工干预、终止执行、运行期会话和缓存说明请阅读 [Debug 对话](#debug-对话)、[人工干预与终止](#人工干预与终止) 与 [对话、会话与执行](#对话会话与执行)。


<!-- 来源：chat-response.md -->

## 对话响应

本页说明 `chat/completion` 与 `debug/completion` 的响应结构。两者都会返回同一种 `ChatResponse`；差异主要在请求体和运行模式，详见 [Debug 对话](#debug-对话)。

### 非流式响应

`stream: false` 时，接口一次性返回完整 JSON：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "user_message_id": "01K...",
  "assistant_message_id": "01K...",
  "message": {
    "id": "01K...",
    "conversation_id": "01K...",
    "role": "assistant",
    "content_type": "prompt",
    "status": "",
    "reply_id": "01K...",
    "index": 2,
    "agent_info": {
      "agent_id": "01K...",
      "agent_name": "example_agent",
      "agent_status": "",
      "agent_version": "v0"
    },
    "content": {
      "final_answer": {
        "query": "请用一句话介绍你自己",
        "answer": {
          "text": "我是一个智能助手，可以理解问题并帮助完成任务。"
        },
        "selected_files": null,
        "thinking": "",
        "skill_process": null,
        "answer_type_other": null,
        "output_variables_config": {
          "answer_var": "answer",
          "doc_retrieval_var": "doc_retrieval_res",
          "graph_retrieval_var": "graph_retrieval_res",
          "related_questions_var": "related_questions",
          "other_vars": null,
          "middle_output_vars": null
        }
      },
      "middle_answer": {
        "progress": [],
        "doc_retrieval": null,
        "graph_retrieval": null,
        "other_variables": {}
      }
    },
    "ext": {
      "agent_run_id": "01K...",
      "ttft": 1234
    }
  },
  "error": null
}
```

### 顶层字段

| 字段 | 说明 |
| --- | --- |
| `conversation_id` | 对话 ID。继续对话时把它放回请求体的 `conversation_id`。 |
| `agent_run_id` | 本次 Agent 执行 ID。终止执行、恢复执行和 trace 排障时会用到。 |
| `user_message_id` | 本轮用户消息 ID。 |
| `assistant_message_id` | 本轮助手消息 ID。人工干预恢复时通常要和 `agent_run_id` 一起保存。 |
| `message` | 助手消息对象，结构与对话详情中的消息基本一致。 |
| `error` | 运行错误。成功时通常为 `null`；如果 Executor 运行中返回业务错误，应优先查看此字段。 |

### `message` 字段

| 字段 | 说明 |
| --- | --- |
| `id` | 助手消息 ID，通常与顶层 `assistant_message_id` 一致。 |
| `conversation_id` | 消息所属对话 ID。 |
| `role` | 消息角色。对话响应里的 `message` 通常是 `assistant`。 |
| `content_type` | 内容类型。`prompt` 表示最终答案在 `content.final_answer.answer.text`；`explore` 和 `other` 用于特殊输出。 |
| `reply_id` | 被回复的用户消息 ID，通常与顶层 `user_message_id` 一致。 |
| `index` | 消息在对话中的序号。 |
| `agent_info` | 本次响应使用的 Agent 基本信息和版本。 |
| `content` | 响应主体，包含最终答案和中间过程。 |
| `ext` | 扩展信息，例如 `agent_run_id`、`ttft`、`total_time`、`interrupt_info`。 |

### 最终答案

普通文本回答优先读取：

```text
message.content.final_answer.answer.text
```

如果 Agent 配置了特殊输出等，还需要关注：

| 字段 | 说明 |
| --- | --- |
| `thinking` | 思考过程或规划内容，是否返回取决于 Agent 模式和后端配置。 |
| `output_variables_config` | 输出变量映射，说明答案等变量名。 |

### 中间过程

`message.content.middle_answer` 用于展示执行轨迹：

| 字段 | 说明 |
| --- | --- |
| `progress` | Dolphin / ReAct 执行过程列表，通常包含模型阶段、工具阶段、状态、输入、输出、token 统计等。 |
| `other_variables` | `config.output.variables.other_vars` 中配置的其他输出变量。 |

### 人工干预

当工具或 Sub-Agent 需要用户确认时，响应可能在 `message.ext.interrupt_info` 中返回中断信息。前端或调用方应保存这些字段：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "assistant_message_id": "01K...",
  "interrupt_info": {
    "handle": {
      "frame_id": "frame-id",
      "snapshot_id": "snapshot-id",
      "resume_token": "resume-token",
      "interrupt_type": "tool_interrupt",
      "current_block": 0,
      "restart_block": false
    },
    "data": {
      "tool_name": "获取agent详情",
      "tool_description": "",
      "tool_args": [
        {
          "key": "key",
          "value": "DocQA_Agent",
          "type": "str"
        }
      ],
      "interrupt_config": {
        "requires_confirmation": true,
        "confirmation_message": "是否确认参数并继续执行?"
      }
    }
  }
}
```

Dolphin 对外层中断事件可能使用 `tool_confirmation` 表示“工具确认事件”，但恢复句柄里的 `interrupt_info.handle.interrupt_type` 是 `tool_interrupt`。恢复执行时，不要手写这个字段；应把响应中的 `interrupt_info.handle` 原样放入请求体的 `resume_interrupt_info.resume_handle`，并把 `interrupt_info.data` 放入 `resume_interrupt_info.data`。同时传回 `agent_run_id` 和 `interrupted_assistant_message_id`。完整流程见 [人工干预与终止](#人工干预与终止)。

### 流式响应

`stream: true` 时返回 SSE：

```text
data: {"conversation_id":"01K...","agent_run_id":"01K...","message":{...},"error":null}
```

普通流式下，每个 `data:` 都是当前完整响应快照，客户端可以直接用最新快照覆盖旧状态。

增量流式下，每个 `data:` 是增量事件：

```json
{
  "seq_id": 1,
  "key": ["message", "content", "final_answer", "answer", "text"],
  "content": "新增文本",
  "action": "append"
}
```

增量事件的合并方式见 [增量流式](#增量流式)。


<!-- 来源：incremental-streaming.md -->

## 增量流式

Decision Agent 对话接口支持三种返回方式：

| 模式 | 请求参数 | 返回形态 | 适合场景 |
| --- | --- | --- | --- |
| 非流式 | `stream: false` | 一次返回完整 JSON | 后端任务、批处理、简单脚本。 |
| 普通流式 | `stream: true, inc_stream: false` | SSE；每个 `data:` 是当前完整响应快照 | 前端直接覆盖式渲染。 |
| 增量流式 | `stream: true, inc_stream: true` | SSE；每个 `data:` 是增量变更 | 前端希望减少传输和精细更新局部状态。 |

### 请求示例

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_key "$AGENT_KEY" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "请用三点说明你的能力" \
  '{
  agent_id: $agent_id,
  agent_key: $agent_key,
  agent_version: $agent_version,
  query: $query,
  stream: true,
  inc_stream: true
}' > /tmp/chat-incremental-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d @/tmp/chat-incremental-stream-request.json
```

### 增量事件格式

每条 SSE 消息的 `data:` 是一个 JSON 对象：

```json
{
  "seq_id": 0,
  "key": ["message", "content"],
  "content": "新增内容",
  "action": "append"
}
```

| 字段 | 说明 |
| --- | --- |
| `seq_id` | 增量事件序号，单次对话内递增。 |
| `key` | 目标路径。字符串表示对象字段，数字表示数组下标；空数组表示根节点。 |
| `content` | 本次变更内容。`remove` 或 `end` 时可以为 `null`。 |
| `action` | 变更动作，支持 `upsert`、`append`、`remove`、`end`。 |

### Decision Agent 增量算法

Decision Agent 后端会先接收 Agent Executor 返回的完整 JSON 片段，再在服务端做 diff：

1. 将上一帧完整 JSON 作为 `oldJSON`，当前帧完整 JSON 作为 `newJSON`。
2. 从根节点递归比较对象、数组、字符串和基础类型。
3. 对新增字段发送 `upsert`，对删除字段发送 `remove`。
4. 对字符串前缀增长发送 `append`，例如旧值为 `"你好"`、新值为 `"你好世界"` 时，只发送 `"世界"`。
5. 对数组新增元素发送 `append`，对类型变化或非前缀字符串替换发送 `upsert`。
6. 对话结束时发送 `{ "key": [], "content": null, "action": "end" }`。

服务端不会要求客户端回传上一帧；客户端只需要按 `seq_id` 顺序应用增量事件。

### 客户端合并规则

客户端可以按以下规则还原完整响应：

```js
function applyIncrementalEvent(state, event) {
  if (event.action === 'end') {
    return state;
  }

  const path = event.key ?? [];
  if (path.length === 0) {
    return event.content;
  }

  let cursor = state;
  for (let i = 0; i < path.length - 1; i += 1) {
    const key = path[i];
    const nextKey = path[i + 1];
    if (cursor[key] == null) {
      cursor[key] = Number.isInteger(nextKey) ? [] : {};
    }
    cursor = cursor[key];
  }

  const lastKey = path[path.length - 1];
  if (event.action === 'remove') {
    if (Array.isArray(cursor) && Number.isInteger(lastKey)) {
      cursor.splice(lastKey, 1);
    } else {
      delete cursor[lastKey];
    }
  } else if (event.action === 'append') {
    if (typeof cursor[lastKey] === 'string') {
      cursor[lastKey] += event.content;
    } else if (Array.isArray(cursor[lastKey])) {
      cursor[lastKey].push(event.content);
    } else {
      cursor[lastKey] = event.content;
    }
  } else {
    cursor[lastKey] = event.content;
  }

  return state;
}
```

### 与 Executor 增量输出的关系

Agent Executor 内部也有 `_options.incremental_output`，它会在 Executor 内部把完整 JSON 转为增量事件。REST 用户通过 Decision Agent 对接时，主要使用 `inc_stream`；只有在直接调用 Executor 或调试 Executor 内部输出链路时，才需要关注 `_options.incremental_output`。


<!-- 来源：intervention-termination.md -->

## 人工干预与终止

本页说明人工干预、中断恢复、终止执行和断线续连的 REST 接入方式。概念边界可先阅读 [运行控制](../concepts/runtime-control.md)。

### 能力边界

| 能力 | 使用场景 | 接口 |
| --- | --- | --- |
| 人工干预 | 工具或 Sub-Agent 执行前需要用户确认。 | `chat/completion` 或 `debug/completion` 返回 `interrupt_info` |
| 中断恢复 | 用户确认或跳过被中断的工具 / Sub-Agent 后继续执行。 | 再次调用 `chat/completion` 或 `debug/completion` |
| 终止执行 | 用户主动停止正在运行的 run。 | `POST /api/agent-factory/v1/app/{agent_key}/chat/termination` |
| 断线续连 | 流式连接断开，但服务端 run 仍在继续。 | `POST /api/agent-factory/v1/app/{agent_key}/chat/resume` |

`/chat/resume` 只用于流式断线续连，不用于人工干预恢复。人工干预恢复必须把 `resume_interrupt_info` 传回对话接口。

### 开启人工干预

Sub-Agent 技能可以在 `config.skills.agents[]` 中开启人工干预：

```json
{
  "skills": {
    "agents": [
      {
        "agent_key": "<sub-agent-key>",
        "agent_version": "latest",
        "intervention": true,
        "intervention_confirmation_message": "即将调用合同风险抽取 Agent，请确认是否继续。"
      }
    ]
  }
}
```

工具技能也可以使用同名字段：

```json
{
  "skills": {
    "tools": [
      {
        "tool_id": "<tool-id>",
        "intervention": true,
        "intervention_confirmation_message": "即将调用外部工具，请确认是否继续。"
      }
    ]
  }
}
```

`intervention_confirmation_message` 会进入中断详情的 `interrupt_config.confirmation_message`，客户端可以直接展示给用户。

### 发起对话并捕获中断

普通对话请求与其他对话一致。为了后续恢复和终止，建议保存响应里的 `conversation_id`、`agent_run_id` 和 `assistant_message_id`：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v0}" \
  --arg query "请审核这段合同，并在需要调用风险抽取 Agent 前让我确认。" \
  '{
    agent_id: $agent_id,
    agent_version: $agent_version,
    query: $query,
    stream: false,
    executor_version: "v2",
    chat_option: {
      is_need_history: true,
      is_need_doc_retrival_post_process: true,
      is_need_progress: true,
      enable_dependency_cache: true
    }
  }' > /tmp/chat-intervention-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-intervention-request.json | tee /tmp/chat-intervention-response.json
```

上面用 `stream: false` 是为了让示例能直接保存完整 JSON。如果前端或页面使用 `stream: true, inc_stream: true`，人工干预字段仍在合并后的响应快照 `message.ext.interrupt_info` 中。

发生人工干预时，响应的 `message.ext.interrupt_info` 中会包含恢复句柄和中断详情：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "assistant_message_id": "01K...",
  "message": {
    "ext": {
      "interrupt_info": {
        "handle": {
          "frame_id": "frame-id",
          "snapshot_id": "snapshot-id",
          "resume_token": "resume-token",
          "interrupt_type": "tool_interrupt",
          "current_block": 0,
          "restart_block": false
        },
        "data": {
          "tool_name": "获取agent详情",
          "tool_description": "",
          "tool_args": [
            {
              "key": "key",
              "value": "DocQA_Agent",
              "type": "str"
            }
          ],
          "interrupt_config": {
            "requires_confirmation": true,
            "confirmation_message": "是否确认参数并继续执行?"
          }
        }
      }
    }
  }
}
```

注意字段名称：响应里是 `interrupt_info.handle`；恢复请求里才使用 `resume_interrupt_info.resume_handle`。Dolphin 侧的外层事件类型会用 `tool_confirmation` 表示工具确认事件，但传给恢复接口的句柄字段 `interrupt_info.handle.interrupt_type` 是 `tool_interrupt`，调用方应直接复用服务端返回的 `handle`。

### 确认后恢复执行

用户确认继续后，再次调用原对话接口。把响应中的 `interrupt_info.handle` 原样放入 `resume_interrupt_info.resume_handle`，把 `interrupt_info.data` 原样放入 `resume_interrupt_info.data`：

```bash
test -n "$CONVERSATION_ID"
test -n "$AGENT_RUN_ID"
test -n "$INTERRUPTED_ASSISTANT_MESSAGE_ID"

jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v0}" \
  --arg conversation_id "$CONVERSATION_ID" \
  --arg agent_run_id "$AGENT_RUN_ID" \
  --arg interrupted_assistant_message_id "$INTERRUPTED_ASSISTANT_MESSAGE_ID" \
  --argjson interrupt_info "$(jq '.message.ext.interrupt_info' /tmp/chat-intervention-response.json)" \
  '{
    agent_id: $agent_id,
    agent_version: $agent_version,
    conversation_id: $conversation_id,
    agent_run_id: $agent_run_id,
    interrupted_assistant_message_id: $interrupted_assistant_message_id,
    resume_interrupt_info: {
      resume_handle: $interrupt_info.handle,
      action: "confirm",
      modified_args: [],
      data: $interrupt_info.data
    },
    stream: true,
    inc_stream: true,
    executor_version: "v2",
    chat_option: {
      is_need_history: true,
      is_need_doc_retrival_post_process: true,
      is_need_progress: true,
      enable_dependency_cache: true
    }
  }' > /tmp/chat-resume-interrupt-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-resume-interrupt-request.json
```

恢复请求通常不需要再传新的 `query`，后端会根据 `conversation_id`、`agent_run_id` 和 `resume_interrupt_info` 回到被中断的执行位置继续运行。

如果用户选择跳过当前工具或 Sub-Agent，把 `action` 改为 `skip`：

```json
{
  "resume_interrupt_info": {
    "resume_handle": {
      "frame_id": "frame-id",
      "snapshot_id": "snapshot-id",
      "resume_token": "resume-token",
      "interrupt_type": "tool_interrupt",
      "current_block": 0,
      "restart_block": false
    },
    "action": "skip",
    "modified_args": [],
    "data": {
      "tool_name": "获取agent详情",
      "tool_description": "",
      "tool_args": [
        {
          "key": "key",
          "value": "DocQA_Agent",
          "type": "str"
        }
      ],
      "interrupt_config": {
        "requires_confirmation": true,
        "confirmation_message": "是否确认参数并继续执行?"
      }
    }
  }
}
```

如果允许用户修改参数，可以在 `modified_args` 中传入修改后的参数：

```json
{
  "modified_args": [
    {
      "key": "contract_text",
      "value": "用户确认后的合同文本"
    }
  ]
}
```

### Debug 对话中的恢复

Debug 对话同样支持 `resume_interrupt_info`。区别是输入内容放在 `input` 内：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v0}" \
  --arg conversation_id "$CONVERSATION_ID" \
  --arg agent_run_id "$AGENT_RUN_ID" \
  --arg interrupted_assistant_message_id "$INTERRUPTED_ASSISTANT_MESSAGE_ID" \
  --argjson interrupt_info "$(jq '.message.ext.interrupt_info' /tmp/debug-intervention-response.json)" \
  '{
    agent_id: $agent_id,
    agent_version: $agent_version,
    input: {},
    conversation_id: $conversation_id,
    agent_run_id: $agent_run_id,
    interrupted_assistant_message_id: $interrupted_assistant_message_id,
    resume_interrupt_info: {
      resume_handle: $interrupt_info.handle,
      action: "confirm",
      modified_args: [],
      data: $interrupt_info.data
    },
    stream: true,
    inc_stream: true
  }' > /tmp/debug-resume-interrupt-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-resume-interrupt-request.json
```

### 终止执行

终止接口用于停止正在运行的一次执行：

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/termination
```

请求体：

```json
{
  "conversation_id": "01K...",
  "agent_run_id": "01K...",
  "interrupted_assistant_message_id": "01K..."
}
```

字段说明：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `conversation_id` | 是 | 用于定位 Decision Agent 侧的流式 stop channel。 |
| `agent_run_id` | 否，推荐传 | 用于通知 Agent Executor 终止对应 run。 |
| `interrupted_assistant_message_id` | 否 | 传入后，后端会尝试把对应助手消息标记为 cancelled。 |

示例：

```bash
test -n "$CONVERSATION_ID"

jq -n \
  --arg conversation_id "$CONVERSATION_ID" \
  --arg agent_run_id "${AGENT_RUN_ID:-}" \
  --arg interrupted_assistant_message_id "${INTERRUPTED_ASSISTANT_MESSAGE_ID:-}" \
  '{
    conversation_id: $conversation_id,
    agent_run_id: $agent_run_id,
    interrupted_assistant_message_id: $interrupted_assistant_message_id
  }' > /tmp/chat-termination-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/termination" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-termination-request.json
```

成功时通常返回 `204` 或空成功响应。

### 断线续连

断线续连接口只用于继续读取运行中的流式响应：

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/resume
```

```bash
jq -n --arg conversation_id "$CONVERSATION_ID" '{
  conversation_id: $conversation_id
}' > /tmp/chat-stream-resume-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/resume" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-stream-resume-request.json
```

该接口依赖运行中的内存态会话。如果 run 已经结束，或服务端已清理 `SessionMap`，续连会失败或没有可读取的数据。

### 常见问题

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 人工干预没有触发 | `intervention` 没有配置到实际会被调用的工具或 Sub-Agent 上 | 检查 `config.skills.tools[]` 或 `config.skills.agents[]`。 |
| 恢复时报参数错误 | 把响应字段 `interrupt_info.handle` 直接写成了 `interrupt_info.resume_handle` | 恢复请求中应写 `resume_interrupt_info.resume_handle: interrupt_info.handle`。 |
| 恢复后像重新开始 | 缺少 `conversation_id`、`agent_run_id` 或 `interrupted_assistant_message_id` | 从被中断响应中保存并传回这些字段。 |
| `/chat/resume` 不能恢复人工干预 | 混淆了断线续连和中断恢复 | 人工干预恢复应再次调用 `chat/completion` 或 `debug/completion`。 |
| 终止接口返回错误 | run 已结束，或没有找到对应 stop channel | 如果只需要取消被中断消息，传入 `interrupted_assistant_message_id`；否则重新检查当前 run 状态。 |


<!-- 来源：debug-chat.md -->

## Debug 对话

Debug 对话用于 Agent 配置页调试和排障，适合在 Agent 尚未正式对外使用前验证输入变量、知识网络、工具调用和输出结构。普通业务调用优先使用 `chat/completion`；只有需要排查配置或运行链路时再使用 debug 接口。

Debug 对话的返回结构与普通对话一致，字段说明见 [对话响应](#对话响应)。

### 接口

```text
POST /api/agent-factory/v1/app/{agent_key}/debug/completion
```

### 与普通对话的区别

| 对比项 | 普通对话 | Debug 对话 |
| --- | --- | --- |
| 接口 | `/chat/completion` | `/debug/completion` |
| 调用类型 | Decision Agent 后端设置为 `chat` 或 `api_chat` | Decision Agent 后端设置为 `debug_chat` |
| 请求字段 | `query`、`history`、`custom_querys` 在请求体顶层 | `query`、`history`、`custom_querys` 放在 `input` 内 |
| Executor 路由 | `/api/agent-executor/v2/agent/run` | `/api/agent-executor/v2/agent/debug` |
| Executor 运行参数 | `is_debug_run=false` | `is_debug_run=true` |
| 缓存行为 | 可通过 `chat_option.enable_dependency_cache` 启用依赖缓存 | Debug 执行会跳过 dependency cache 初始化 |

Decision Agent 后端的 debug handler 会把 `input.query`、`input.history`、`input.custom_querys` 转换成内部 `ChatReq`，再将 `CallType` 设置为 `debug_chat`。Agent Executor 侧 debug 路由仍复用同一套 `run_agent` 主流程，但传入 `is_debug_run=True`，因此适合验证配置，不适合作为正式业务调用入口。

### 非流式 Debug

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "测试 debug completion" \
  '{
  agent_id: $agent_id,
  agent_version: $agent_version,
  input: {
    query: $query
  },
  stream: false
}' > /tmp/debug-chat-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-chat-request.json | jq .
```

### 普通流式 Debug

普通流式返回的是 SSE，每个 `data:` 事件都是当前完整响应快照。

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "用 debug 模式流式回答" \
  '{
  agent_id: $agent_id,
  agent_version: $agent_version,
  input: {
    query: $query
  },
  stream: true,
  inc_stream: false
}' > /tmp/debug-chat-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-chat-stream-request.json
```

### 增量流式 Debug

增量流式返回的是 diff 事件，每个 `data:` 事件形如 `{seq_id, key, content, action}`。具体合并算法与事件处理方式见 [增量流式](#增量流式)。

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "$AGENT_VERSION" \
  --arg query "用 debug 模式增量流式回答" \
  '{
  agent_id: $agent_id,
  agent_version: $agent_version,
  input: {
    query: $query
  },
  stream: true,
  inc_stream: true
}' > /tmp/debug-chat-inc-stream-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/debug-chat-inc-stream-request.json
```

### 使用建议

- Debug 接口主要面向配置验证和排障，不建议作为线上用户对话入口。
- Debug 请求体的对话输入放在 `input` 内；普通对话请求体则直接使用顶层 `query`。
- Debug 执行会跳过 dependency cache。需要验证缓存效果时，请使用普通对话并配置 `chat_option.enable_dependency_cache`。
- Debug 对话也支持人工干预恢复，请把 `resume_interrupt_info`、`agent_run_id`、`interrupted_assistant_message_id` 放在请求体顶层；具体示例见 [人工干预与终止](#人工干预与终止)。


<!-- 来源：conversation-session-run.md -->

## 对话、会话与执行

Decision Agent 中需要区分三个概念：`conversation` 对应“对话”，`session` 对应“会话”，`run` 对应“一次执行”。普通 API 对接通常只需要关心 `conversation_id`；`session` 更多用于缓存、续连和运行期状态，暂时用不到时可以先忽略。

### 概念关系

| 概念 | 常见字段 | 中文名称 | 作用 |
| --- | --- | --- | --- |
| `conversation` | `conversation_id` | 对话 | 用户侧的一段连续对话上下文，用于继续对话、查看消息、标记已读等。 |
| `session` | `conversation_session_id` | 会话 | 某段运行期会话状态，当前主要服务于缓存、续连和运行态管理。 |
| `run` | `agent_run_id` | 一次执行 | 一次 Agent 调用或接口执行，用于运行日志、trace、终止和排障。 |

关系上，一个对话可以对应多个会话，一个会话可以包含一个或多个 run。服务端在对话时会基于 `conversation_id` 和 session `start_time` 生成 `conversation_session_id`，格式通常类似：

```text
{conversation_id}-{start_time}
```

### 对话续连

Decision Agent 后端在流式对话过程中会维护运行中内存 `SessionMap`，按 `conversation_id` 暂存响应快照、停止信号和续连状态。客户端断开后，如果服务端仍保留该对话的运行态，可以通过 resume 接口继续读取流式响应。

这里的 resume 是“断线续连”，不是人工干预后的中断恢复。人工干预恢复需要再次调用 `chat/completion` 或 `debug/completion`，并传入 `resume_interrupt_info`。完整说明见 [人工干预与终止](#人工干预与终止)。

```text
POST /api/agent-factory/v1/app/{agent_key}/chat/resume
```

```bash
jq -n --arg conversation_id "$CONVERSATION_ID" '{
  conversation_id: $conversation_id
}' > /tmp/chat-resume-request.json

af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/resume" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-resume-request.json
```

`SessionMap` 是内存态结构，主要用于当前运行中的流式响应和断线续连；一次执行结束后，服务端会清理对应的 `conversation_id`。

### Redis 会话

除了运行中内存态，Decision Agent 后端还会在 Redis 中维护 conversation session，key 前缀为：

```text
agent-app:conversation-session:
```

默认 TTL 为 600 秒。会话管理接口可用于获取会话信息、创建会话或恢复会话生命周期：

```text
PUT /api/agent-factory/v1/conversation/session/{conversation_id}
```

获取或创建：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-latest}" \
  '{
    action: "get_info_or_create",
    agent_id: $agent_id,
    agent_version: $agent_version
  }' > /tmp/conversation-session.json

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/conversation/session/$CONVERSATION_ID" \
  -H "Content-Type: application/json" \
  -d @/tmp/conversation-session.json | jq .
```

恢复生命周期或创建：

```bash
jq -n \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-latest}" \
  '{
    action: "recover_lifetime_or_create",
    agent_id: $agent_id,
    agent_version: $agent_version
  }' > /tmp/conversation-session-recover.json

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/conversation/session/$CONVERSATION_ID" \
  -H "Content-Type: application/json" \
  -d @/tmp/conversation-session-recover.json | jq .
```

响应中会包含 `conversation_session_id` 和 `ttl`：

```json
{
  "conversation_session_id": "conversation-id-1710000000",
  "ttl": 600
}
```

普通对接场景里，继续对话优先使用 `conversation_id`。会话管理接口属于缓存和运行期优化能力，只有需要控制会话生命周期或对接断线续连时再使用。

### 缓存行为

缓存相关逻辑分两层：

| 位置 | 说明 |
| --- | --- |
| Decision Agent 会话管理 | 管理 Redis conversation session；部分入口可触发 Executor agent cache upsert。 |
| Agent Executor dependency cache | 普通执行可通过 `chat_option.enable_dependency_cache` 启用；默认 TTL 为 60 秒，超过 10 秒阈值后可触发缓存更新。 |

普通对话示例：

```bash
jq -n --arg query "使用缓存回答这个问题" '{
  query: $query,
  stream: false,
  chat_option: {
    enable_dependency_cache: true
  }
}' > /tmp/chat-cache-request.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d @/tmp/chat-cache-request.json | jq .
```

Debug 执行会跳过 dependency cache 初始化。如果要验证缓存命中和刷新，请使用普通对话接口。


<!-- 来源：import-export-and-auxiliary.md -->

## 导入导出与辅助 API

本页覆盖高级能力，通常不属于首次接入主路径。

### 导出 Agent

```bash
test -n "$AGENT_ID"

jq -n --arg agent_id "$AGENT_ID" '{
  agent_ids: [$agent_id]
}' > /tmp/agent-export.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-inout/export" \
  -H "Content-Type: application/json" \
  -d @/tmp/agent-export.json \
  -o /tmp/agent-export-result.json
```

### 导入 Agent

导入会创建或更新 Agent，执行前请确认文件来源可信。

```bash
test -f /tmp/agent-export-result.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-inout/import" \
  -F "import_type=upsert" \
  -F "file=@/tmp/agent-export-result.json;type=application/json" | jq .
```

### 权限检查

```bash
jq -n --arg agent_id "${AGENT_ID:-}" '{
  agent_id: $agent_id
}' > /tmp/agent-permission.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-permission/execute" \
  -H "Content-Type: application/json" \
  -d @/tmp/agent-permission.json | jq .
```

### 用户权限状态

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent-permission/management/user-status" | jq .
```

### 产品与分类

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v3/product" | jq .
af_curl "$AF_BASE_URL/api/agent-factory/v3/category" | jq .
```
<!-- 
### 临时区文件扩展名映射

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/temp-zone/file-ext-map" | jq .
``` -->
