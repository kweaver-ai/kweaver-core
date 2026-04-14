# Execution Factory

## Overview

The **Execution Factory** registers and runs **operators**, **tools**, and **skills** that agents invoke under policy. It bridges LLM planning to concrete business actions (HTTP tools, code runners, MCP integrations, etc.).

Typical ingress prefix:

| Prefix | Role |
| --- | --- |
| `/api/agent-operator-integration/v1` | Operator integration and tool execution surface |

**Related modules:** [Decision Agent](decision-agent.md), [Dataflow](dataflow.md) (automation and code-runner paths).

## Usage

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver execution --help
# Or tool/operator subcommands as provided by your CLI version
# kweaver tool list
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# Illustrative
# tools = client.execution_factory.list_tools()
# result = client.execution_factory.invoke(tool_id="...", input={})
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// const tools = await client.executionFactory.listTools();
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-operator-integration/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# Illustrative operator list or invoke — align with OpenAPI
curl -sk -X GET "$KWEAVER_BASE/api/agent-operator-integration/v1/operators" \
  -H "Authorization: Bearer $TOKEN"
```
