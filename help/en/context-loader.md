# Context Loader

## Overview

The **Context Loader** (including **agent-retrieval** services) assembles **high-quality context** for Decision Agents: ontology-aware recall, ranking, and on-demand loading from BKN and data plane. It sits between raw data/VEGA and the agent runtime.

Typical ingress prefix:

| Prefix | Role |
| --- | --- |
| `/api/agent-retrieval/v1` | Retrieval and context assembly APIs |

**Related modules:** [BKN Engine](bkn.md), [VEGA Engine](vega.md), [Decision Agent](decision-agent.md).

## Usage

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver --help
# Subcommands for retrieval/context may be exposed under a dedicated group in newer CLI versions
# kweaver context --help
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# Illustrative: build context for a task / agent turn
# ctx = client.context_loader.build(
#     agent_id="...",
#     query="user question",
#     bkn_id="...",
# )
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// const ctx = await client.contextLoader.build({ ... });
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-retrieval/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# Illustrative retrieval request — replace body with OpenAPI contract
curl -sk -X POST "$KWEAVER_BASE/api/agent-retrieval/v1/retrieve" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"example","limit":10}'
```
