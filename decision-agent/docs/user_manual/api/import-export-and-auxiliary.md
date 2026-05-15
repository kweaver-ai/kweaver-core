# 导入导出与辅助 API

本页覆盖高级能力，通常不属于首次接入主路径。

## 导出 Agent

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

## 导入 Agent

导入会创建或更新 Agent，执行前请确认文件来源可信。

```bash
test -f /tmp/agent-export-result.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-inout/import" \
  -F "import_type=upsert" \
  -F "file=@/tmp/agent-export-result.json;type=application/json" | jq .
```

## 权限检查

```bash
jq -n --arg agent_id "${AGENT_ID:-}" '{
  agent_id: $agent_id
}' > /tmp/agent-permission.json

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-permission/execute" \
  -H "Content-Type: application/json" \
  -d @/tmp/agent-permission.json | jq .
```

## 用户权限状态

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent-permission/management/user-status" | jq .
```

## 产品与分类

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v3/product" | jq .
af_curl "$AF_BASE_URL/api/agent-factory/v3/category" | jq .
```
<!-- 
## 临时区文件扩展名映射

```bash
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/temp-zone/file-ext-map" | jq .
``` -->
