# Decision Agent

## Overview

**Decision Agents** are goal-oriented systems that plan, retrieve context, call tools under policy, and iterate with feedback. Core services include **agent-factory** (orchestration APIs), **agent-executor**, **memory**, and **retrieval** integration.

Typical ingress prefixes:

| Prefix | Role |
| --- | --- |
| `/api/agent-factory/v1` | Factory APIs (templates, agents, runs) |
| `/api/agent-factory/v2`, `/api/agent-factory/v3` | Versioned factory surfaces |
| `/api/agent-app/v1` | Application-facing agent APIs |

**Related modules:** [Context Loader](context-loader.md), [Execution Factory](execution-factory.md), [BKN Engine](bkn.md), [Trace AI](trace-ai.md).

## Usage

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver agent --help
# List agents, create from template, chat, or run tasks per CLI version
# kweaver agent list
# kweaver agent chat <agent-id> --message "Hello"
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# agents = client.agent.list()
# reply = client.agent.chat(agent_id="...", message="Plan next steps")
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// const agents = await client.agent.list();
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-factory/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# Illustrative: list agents or templates
curl -sk -X GET "$KWEAVER_BASE/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $TOKEN"
```
