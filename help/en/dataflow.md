# Dataflow

## Overview

**Dataflow** orchestrates **pipelines**, **scheduled jobs**, **automation**, and **code runners** across the AI Data Platform. It connects ingestion, transformation, and hand-off to agents or downstream systems.

Typical ingress prefixes (see Helm values in `adp/dataflow`):

| Prefix | Role |
| --- | --- |
| `/api/automation/v1` | Automation and scheduling |
| `/api/flow-stream-data-pipeline/v1` | Stream / pipeline control |
| `/api/coderunner` | Sandboxed or managed code execution hooks |

**Related modules:** [VEGA Engine](vega.md), [Execution Factory](execution-factory.md), [BKN Engine](bkn.md).

## Usage

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver dataflow --help
# List flows, trigger runs, or inspect history when supported
# kweaver dataflow list
# kweaver dataflow run <flow-id>
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# flows = client.dataflow.list_flows()
# run = client.dataflow.trigger(flow_id="...", inputs={})
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// const flows = await client.dataflow.listFlows();
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/automation/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# Illustrative pipeline or flow API
curl -sk -X GET "$KWEAVER_BASE/api/flow-stream-data-pipeline/v1/flows" \
  -H "Authorization: Bearer $TOKEN"
```
