# VEGA Engine

## Overview

**VEGA** provides **data virtualization** over heterogeneous sources: **data connections**, **models**, and **views** (including atomic and composite views). Agents and applications query through a unified SQL-oriented surface instead of wiring each source by hand.

Ingress prefix (typical):

| Prefix | Role |
| --- | --- |
| `/api/vega-backend/v1` | VEGA backend — connections, metadata, query execution |

**Related modules:** [BKN Engine](bkn.md) (semantic layer on top of data), [Context Loader](context-loader.md), [Dataflow](dataflow.md) (pipelines that land or transform data).

## Usage

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver vega --help
# List connections, catalogs, or run ad-hoc queries when supported by CLI version
# kweaver vega connection list
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# Illustrative — method names depend on SDK release
# connections = client.vega.list_connections()
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// const connections = await client.vega.listConnections();
```

### curl

```bash
# Example health or metadata (adjust to OpenAPI)
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# Example SQL over a registered view (payload is illustrative)
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/query" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sql":"SELECT 1"}'
```
