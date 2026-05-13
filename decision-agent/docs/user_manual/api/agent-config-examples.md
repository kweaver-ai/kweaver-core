# Agent config 示例

Agent 的 `config` 是创建和更新 Agent 时最重要的字段。本页给出两个可直接保存为 JSON 文件的示例：最小化配置和较复杂配置。

## 最小化配置

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

## 较复杂配置

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

## 参数说明

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
