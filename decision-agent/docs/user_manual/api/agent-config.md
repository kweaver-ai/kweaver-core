# Agent 配置 API

Agent 的 `config` 决定输入输出、LLM、技能、知识网络、记忆、历史策略和运行模式。创建与更新时，`config` 是最容易出错的部分。

请先阅读 [Agent config 示例](./agent-config-examples.md)，按场景准备一个配置文件，例如 `/tmp/agent-config-minimal.json` 或 `/tmp/agent-config-full.json`。Agent 模式的产品含义可参考 [Agent 模式](../concepts/agent-modes.md)。

## 最小可用要求

| 配置 | 要求 |
| --- | --- |
| `input` | 必须存在，并包含输入字段定义。 |
| `output` | 必须存在，并定义默认输出格式或变量。 |
| `llms` | 至少一个 `is_default: true` 的 LLM。 |
| `mode` | 可选，支持 `default`、`dolphin`、`react`；缺省时按默认模式处理。 |
| `react_config` | 仅允许 `mode=react` 时传入。 |

## 设置 Agent Mode

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

## 配置知识网络

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

## 常见错误

| 错误 | 原因 | 处理 |
| --- | --- | --- |
| `react_config is only allowed when mode is react` | 非 React 模式传入了 `react_config` | 删除 `react_config` 或设置 `mode=react`。 |
| `mode is invalid` | `mode` 不在允许枚举内 | 使用 `default`、`dolphin`、`react`。 |
| LLM 缺失 | `llms` 为空或没有默认 LLM | 补充 `is_default: true` 的 LLM。 |
