# Trace AI

## Overview

**Trace AI** provides **full-chain observability**: ingest OTLP traces, export to search backends, and query spans linked to agent and platform activity via the **agent-observability** service.

Typical ingress prefix:

| Prefix | Role |
| --- | --- |
| `/api/agent-observability/v1` | Trace query and observability APIs |

**Related modules:** [Decision Agent](decision-agent.md), platform-wide logging and metrics pipelines (feed-ingester integration).

## Usage

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver trace --help
# Or observability subcommands when available
# kweaver trace query --trace-id <id>
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# spans = client.trace_ai.search(trace_id="...")
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// const spans = await client.traceAi.search({ traceId: '...' });
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-observability/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# Illustrative trace search — use your deployment OpenAPI
curl -sk -X POST "$KWEAVER_BASE/api/agent-observability/v1/traces/search" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"limit":20}'
```
