# Agent 生命周期 API

本页覆盖创建、查看、更新、删除 Agent 的核心 API。请先加载 [REST 接入环境](./index.md#最小化-no-auth-环境)。

## 创建 Agent

接口：

```text
POST /api/agent-factory/v3/agent
```

可运行示例。请先按 [Agent config 示例](./agent-config-examples.md) 准备 `/tmp/agent-config-minimal.json`：

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
| `config` | Agent 运行配置，详见 [Agent config 示例](./agent-config-examples.md)。 |

## 获取详情

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

## 更新 Agent

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

## 删除 Agent

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
