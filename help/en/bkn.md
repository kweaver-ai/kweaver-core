# BKN Engine

## Overview

The **Business Knowledge Network (BKN)** is the semantic layer of KWeaver Core. It models your domain with **object types**, **relation types**, and **action types**, stores **instances** and **relations**, and powers agents and analytics.

Typical services (ingress paths may vary by release):

| Prefix | Role |
| --- | --- |
| `/api/bkn-backend/v1` | BKN backend APIs (management and runtime integration) |
| `/api/ontology-manager/v1` | Ontology / schema management |
| `/api/ontology-query/v1` | Graph and ontology-oriented queries |

**Related modules:** [VEGA Engine](vega.md) (data behind views), [Context Loader](context-loader.md) (context from ontology), [Decision Agent](decision-agent.md) (uses BKN at runtime).

## Usage

Set a base URL and token for HTTP examples:

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"   # omit or adjust when auth is disabled (--minimum)
```

Exact request paths and bodies depend on your Core version; use your cluster OpenAPI or the `kweaver` CLI as the source of truth.

### CLI

```bash
kweaver bkn list
kweaver bkn --help
# Push or sync BKN definitions from files (see kweaver-sdk / create-bkn skill)
# kweaver bkn push ./my-network.bkn
```

### Python SDK

Install from [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) (package name and APIs follow the published SDK).

```python
# Illustrative — replace with the client API from your SDK version
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# client.auth.login(...)
networks = client.bkn.list_networks()
print(networks)
```

### TypeScript SDK

```typescript
// Illustrative — replace with the client API from your SDK version
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// await client.auth.login(...);
const networks = await client.bkn.listNetworks();
console.log(networks);
```

### curl

```bash
# Example: probe ontology-manager (adjust path to match your OpenAPI)
curl -sk "$KWEAVER_BASE/api/ontology-manager/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# Example: ontology-query (replace with a real query endpoint from docs)
curl -sk -X POST "$KWEAVER_BASE/api/ontology-query/v1/query" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```
