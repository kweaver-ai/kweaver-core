# Trace AI

## 概述

**Trace AI** 提供**全链路可观测**：接收 OTLP 链路、导出到检索后端，并通过 **agent-observability** 服务查询与智能体及平台活动关联的 Span。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/agent-observability/v1` | 链路查询与可观测 API |

**相关模块：** [Decision Agent](decision-agent.md)，以及平台日志、指标管道（如 feed-ingester）。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver trace --help
# 或按版本提供的 observability 子命令
# kweaver trace query --trace-id <id>
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# spans = client.trace_ai.search(trace_id="...")
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// const spans = await client.traceAi.search({ traceId: '...' });
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-observability/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 示例：链路检索 — 以部署环境 OpenAPI 为准
curl -sk -X POST "$KWEAVER_BASE/api/agent-observability/v1/traces/search" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"limit":20}'
```
