# 发布与广场 API

Agent 创建后需要发布，才能按发布版本对外使用。不发布，自己是可以使用的，但是不能在广场等地方展示，别人也用不了。

发布、更新发布信息和重新发布的区别可参考 [发布逻辑](../concepts/publishing.md)。

## 发布 Agent

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

## 获取发布信息

```bash
test -n "$AGENT_ID"
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/publish-info" | jq .
```

## 已发布 Agent 列表

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

## 获取广场详情

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

## 取消发布

```bash
test -n "$AGENT_ID"
af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/unpublish" | jq .
```
