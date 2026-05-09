# 🤖 Decision Agent

## 📖 Overview

**Decision Agents** are goal-oriented systems that plan, retrieve context, call tools under policy, and iterate with feedback. Core services include **agent-factory** (orchestration APIs), **agent-executor**, **memory**, and **retrieval** integration.

Typical ingress prefixes:

| Prefix | Role |
| --- | --- |
| `/api/agent-factory/v1` | Factory APIs (templates, agents, runs) |
| `/api/agent-factory/v2`, `/api/agent-factory/v3` | Versioned factory surfaces |
| `/api/agent-app/v1` | Application-facing agent APIs |

**Related modules:** [Context Loader](context-loader.md), [Execution Factory](execution-factory.md), [BKN Engine](bkn.md), [Trace AI](trace-ai.md).

> **Model configuration prerequisite**: Agents require an LLM and an Embedding model. A `--minimum` install does not include pre-configured models — complete [Install and deploy — Configure models](../install.md#configure-models) before using agents. Use `--llm-id` when creating an agent to specify the registered LLM ID.

## 🚀 Usage

Run `kweaver auth login <platform-url>` first (`-k` for self-signed TLS). The CLI examples below assume a saved session. For raw HTTP, see the **curl** section at the end.

## 💻 CLI

### Listing Agents

```bash
# List all published agents visible to the current user
kweaver agent list

# Get details for a specific agent
kweaver agent get <agent_id>

# List agents in the user's personal (unpublished) workspace
kweaver agent personal-list
```

### Templates

```bash
# List all available agent templates
kweaver agent template-list

# Get a template's full definition
kweaver agent template-get <template_id>

# Save a template's config locally for offline editing
kweaver agent template-get <template_id> --save-config ./my-agent-config.json
```

### Categories

```bash
# List agent categories (used for organizing templates and agents)
kweaver agent category-list
```

### Creating an Agent

```bash
# Create a new agent from scratch
kweaver agent create \
  --name "Order Analyst" \
  --profile "Analyze e-commerce order data, identify trends, and recommend actions" \
  --llm-id <model_id>

# Create with a pre-built config (from template-get --save-config)
kweaver agent create \
  --name "Order Analyst" \
  --profile "Analyze e-commerce order data" \
  --llm-id <model_id> \
  --config ./my-agent-config.json
```

### Updating an Agent

```bash
# Bind a knowledge network to an agent
kweaver agent update <agent_id> --knowledge-network-id <kn_id>

# Update profile and LLM
kweaver agent update <agent_id> \
  --profile "Updated analysis profile with risk assessment" \
  --llm-id <new_model_id>
```

### Publishing and Unpublishing

```bash
# Publish an agent (makes it visible to all users)
kweaver agent publish <agent_id>

# Unpublish (reverts to personal workspace)
kweaver agent unpublish <agent_id>
```

### Chatting with an Agent

```bash
# Single-turn message
kweaver agent chat <agent_id> -m "What were the top 5 orders by revenue last month?"

# Interactive mode (multi-turn conversation in the terminal)
kweaver agent chat <agent_id>
# > What's the average order value this quarter?
# Agent: Based on the data, the average order value is $142.50...
# > Break that down by region
# Agent: Here's the regional breakdown...
# > quit   # or exit, q (see kweaver agent chat --help)
```

### Session and History Management

```bash
# List conversation sessions for an agent
kweaver agent sessions <agent_id>

# Get full message history for a conversation
kweaver agent history <agent_id> <conversation_id>
```

### Trace Integration

The first argument is **agent id**, the second is **conversation id** (from chat output or `kweaver agent sessions`). If the one-line `trace` summary under `kweaver agent --help` disagrees with `kweaver agent trace --help`, **trust `trace --help`**.

```bash
# Full trace for a conversation (reasoning chain, tool calls, context)
kweaver agent trace <agent_id> <conversation_id>

# Pretty-printed trace output
kweaver agent trace <agent_id> <conversation_id> --pretty
```

### Deleting an Agent

```bash
kweaver agent delete <agent_id>
```

### End-to-End: From Template to Conversation

```bash
# 1. Browse available templates
kweaver agent template-list

# 2. Save a template config locally
kweaver agent template-get tpl-data-analyst --save-config ./analyst-config.json

# 3. Create an agent from the config
kweaver agent create \
  --name "Sales Analyst" \
  --profile "Analyze sales data and generate insights" \
  --llm-id model-gpt4 \
  --config ./analyst-config.json
# → agent_id: agt-xyz789

# 4. Bind a knowledge network
kweaver agent update agt-xyz789 --knowledge-network-id kn-ecommerce

# 5. Publish the agent
kweaver agent publish agt-xyz789

# 6. Start a conversation
kweaver agent chat agt-xyz789 -m "Show me the revenue trend for the last 6 months"

# 7. Continue interactively
kweaver agent chat agt-xyz789
# > Which product categories drove the most growth?
# > quit

# 8. Review sessions and trace (use the same agent_id; conversation_id from chat or sessions)
kweaver agent sessions agt-xyz789
kweaver agent history agt-xyz789 <conversation_id>
kweaver agent trace agt-xyz789 <conversation_id> --pretty
```

---

## Python SDK

```python
from kweaver_sdk import KWeaverClient

# Same as CLI: use ~/.kweaver/ after kweaver auth login (see kweaver-sdk for constructors)
client = KWeaverClient()

# List published agents
agents = client.agent.list()
for a in agents:
    print(a["id"], a["name"], a["status"])

# Get agent details
detail = client.agent.get("agt-xyz789")
print(detail["profile"], detail["llm_id"], detail["knowledge_network_id"])

# List personal (unpublished) agents
personal = client.agent.personal_list()

# List templates
templates = client.agent.template_list()
for t in templates:
    print(t["id"], t["name"], t["category"])

# Get a template
tpl = client.agent.template_get("tpl-data-analyst")
print(tpl["config"])

# Create an agent
agent = client.agent.create(
    name="Sales Analyst",
    profile="Analyze sales data and generate insights",
    llm_id="model-gpt4",
    config=tpl["config"],
)

# Update: bind a knowledge network
client.agent.update(agent["id"], knowledge_network_id="kn-ecommerce")

# Publish
client.agent.publish(agent["id"])

# Chat (single turn)
reply = client.agent.chat(
    agent_id=agent["id"],
    message="Show me the revenue trend for the last 6 months",
)
print(reply["content"])
print("conversation_id:", reply["conversation_id"])

# Chat (multi-turn in the same session)
reply2 = client.agent.chat(
    agent_id=agent["id"],
    message="Which product categories drove the most growth?",
    conversation_id=reply["conversation_id"],
)
print(reply2["content"])

# List sessions
sessions = client.agent.sessions(agent["id"])
for s in sessions:
    print(s["conversation_id"], s["created_at"], s["message_count"])

# Get history
history = client.agent.history(agent["id"], reply["conversation_id"])
for msg in history:
    print(msg["role"], ":", msg["content"][:100])

# Trace (agent_id and conversation_id)
trace = client.agent.trace(agent["id"], reply["conversation_id"])
for span in trace["spans"]:
    print(span["name"], span["duration_ms"], span["status"])

# Delete
client.agent.delete(agent["id"])
```

---

## TypeScript SDK

> More runnable examples ship with the `@kweaver-ai/kweaver-sdk` npm package.

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';
import type { ProgressItem } from '@kweaver-ai/kweaver-sdk';

// Auto-reads credentials from ~/.kweaver/
const client = await KWeaverClient.connect();

// List agents
const agentList = await client.agents.list({ limit: 10 });
for (const a of agentList) {
  console.log(`${a.name} (${a.id}) — ${a.description ?? ''}`);
}

// Single-turn chat
const agentId = agentList[0].id;
const reply = await client.agents.chat(agentId, 'Show me the revenue trend for the last 6 months');
console.log('Reply:', reply.text);

// Inspect the reasoning chain (progress)
if (reply.progress && reply.progress.length > 0) {
  for (const step of reply.progress) {
    console.log(`[${step.skill_info?.type ?? 'step'}] ${step.skill_info?.name ?? step.agent_name} → ${step.status}`);
  }
}

// Streaming chat (real-time output)
let prevLen = 0;
const streamResult = await client.agents.stream(
  agentId,
  'Which product categories drove the most growth?',
  {
    onTextDelta: (fullText: string) => {
      process.stdout.write(fullText.slice(prevLen));
      prevLen = fullText.length;
    },
    onProgress: (progress: ProgressItem[]) => {
      for (const p of progress) {
        if (p.skill_info?.name) {
          console.log(`[progress] ${p.skill_info.name} → ${p.status ?? ''}`);
        }
      }
    },
  },
);

// Conversation history
const conversationId = reply.conversationId;
if (conversationId) {
  const messages = await client.conversations.listMessages(conversationId, { limit: 10 });
  for (const msg of messages) {
    console.log(`[${msg.role}] ${(msg.content ?? '').slice(0, 80)}`);
  }
}

// List conversation sessions
const sessions = await client.conversations.list(agentId, { limit: 5 });
for (const s of sessions) {
  console.log(`${s.id} — ${s.created_at ?? ''}`);
}
```

---

## curl

```bash
# List published agents
curl -sk "https://<access-address>/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $(kweaver token)"

# Get agent details
curl -sk "https://<access-address>/api/agent-factory/v1/agents/agt-xyz789" \
  -H "Authorization: Bearer $(kweaver token)"

# List personal agents
curl -sk "https://<access-address>/api/agent-factory/v1/agents/personal" \
  -H "Authorization: Bearer $(kweaver token)"

# List templates
curl -sk "https://<access-address>/api/agent-factory/v1/templates" \
  -H "Authorization: Bearer $(kweaver token)"

# Get a template
curl -sk "https://<access-address>/api/agent-factory/v1/templates/tpl-data-analyst" \
  -H "Authorization: Bearer $(kweaver token)"

# List categories
curl -sk "https://<access-address>/api/agent-factory/v1/categories" \
  -H "Authorization: Bearer $(kweaver token)"

# Create an agent
curl -sk -X POST "https://<access-address>/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sales Analyst",
    "profile": "Analyze sales data and generate insights",
    "llm_id": "model-gpt4",
    "config": {}
  }'

# Update an agent (bind knowledge network)
curl -sk -X PUT "https://<access-address>/api/agent-factory/v1/agents/agt-xyz789" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"knowledge_network_id": "kn-ecommerce"}'

# Publish an agent
curl -sk -X POST "https://<access-address>/api/agent-factory/v1/agents/agt-xyz789/publish" \
  -H "Authorization: Bearer $(kweaver token)"

# Chat with an agent
curl -sk -X POST "https://<access-address>/api/agent-app/v1/agents/agt-xyz789/chat" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"message": "Show me the revenue trend for the last 6 months"}'

# Continue a conversation
curl -sk -X POST "https://<access-address>/api/agent-app/v1/agents/agt-xyz789/chat" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Which product categories drove the most growth?",
    "conversation_id": "conv-abc123"
  }'

# List sessions
curl -sk "https://<access-address>/api/agent-app/v1/agents/agt-xyz789/sessions" \
  -H "Authorization: Bearer $(kweaver token)"

# Get conversation history
curl -sk "https://<access-address>/api/agent-app/v1/agents/agt-xyz789/sessions/conv-abc123/history" \
  -H "Authorization: Bearer $(kweaver token)"

# Delete an agent
curl -sk -X DELETE "https://<access-address>/api/agent-factory/v1/agents/agt-xyz789" \
  -H "Authorization: Bearer $(kweaver token)"
```
