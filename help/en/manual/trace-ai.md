# 🔭 Trace AI

## 📖 Overview

**Trace AI** provides **full-chain observability**: ingest OTLP traces, export to search backends, and query spans linked to agent and platform activity via the **agent-observability** service.

Every agent conversation generates a trace capturing the full reasoning chain — from user message through context retrieval, tool invocation, and response generation. Traces are the primary mechanism for understanding *why* an agent produced a particular answer and *where* the supporting data came from.

Typical ingress prefix:

| Prefix | Role |
| --- | --- |
| `/api/agent-observability/v1` | Trace query and observability APIs |

**Related modules:** [Decision Agent](decision-agent.md), platform-wide logging and metrics pipelines (feed-ingester integration).

## 💻 CLI

`kweaver agent trace` takes **agent id** first, then **conversation id** (from chat output or `kweaver agent sessions <agent_id>`).

### Retrieving Traces

```bash
# Full trace for a conversation
kweaver agent trace <agent_id> <conversation_id>

# Pretty-printed with indentation and color
kweaver agent trace <agent_id> <conversation_id> --pretty

# Compact single-line JSON output (for piping to jq)
kweaver agent trace <agent_id> <conversation_id> --compact
```

### Trace Data Structure

A trace returned by `kweaver agent trace` contains the following top-level structure:

```json
{
  "trace_id": "tr-abc123",
  "conversation_id": "conv-xyz789",
  "agent_id": "agt-001",
  "started_at": "2026-04-14T10:30:00Z",
  "duration_ms": 4520,
  "status": "completed",
  "spans": [
    {
      "span_id": "sp-001",
      "parent_span_id": null,
      "name": "agent.turn",
      "type": "agent",
      "started_at": "2026-04-14T10:30:00Z",
      "duration_ms": 4520,
      "status": "ok",
      "attributes": {
        "user_message": "What were Q1 revenues?",
        "agent_id": "agt-001"
      },
      "children": ["sp-002", "sp-003", "sp-004"]
    },
    {
      "span_id": "sp-002",
      "parent_span_id": "sp-001",
      "name": "context.retrieval",
      "type": "retrieval",
      "duration_ms": 820,
      "status": "ok",
      "attributes": {
        "kn_id": "kn-ecommerce",
        "query": "Q1 revenues",
        "results_count": 12,
        "sources": ["ot-orders", "ot-invoices"]
      }
    },
    {
      "span_id": "sp-003",
      "parent_span_id": "sp-001",
      "name": "tool.execute",
      "type": "tool",
      "duration_ms": 1200,
      "status": "ok",
      "attributes": {
        "tool_id": "op-sql-query",
        "input": {"sql": "SELECT SUM(amount) FROM orders WHERE quarter='Q1'"},
        "output": {"total": 1250000}
      }
    },
    {
      "span_id": "sp-004",
      "parent_span_id": "sp-001",
      "name": "llm.generate",
      "type": "llm",
      "duration_ms": 2100,
      "status": "ok",
      "attributes": {
        "model": "gpt-4",
        "prompt_tokens": 1850,
        "completion_tokens": 320,
        "temperature": 0.1
      }
    }
  ]
}
```

### Evidence Chain Analysis

Traces enable **evidence chain analysis** — tracing every claim in an agent's response back to its data source:

```bash
# 1. Chat with the agent
kweaver agent chat agt-001 -m "What were Q1 revenues?"
# → conversation_id: conv-xyz789

# 2. Pull the trace (same agent id as step 1)
kweaver agent trace agt-001 conv-xyz789 --pretty

# 3. Walk the evidence chain:
#    agent.turn → context.retrieval (kn-ecommerce, 12 results from ot-orders)
#              → tool.execute (SQL query returned total=1,250,000)
#              → llm.generate (synthesized response from retrieved context + tool output)
```

### Span Types

| Span Type | Description |
| --- | --- |
| `agent` | Top-level agent turn (root span) |
| `retrieval` | Context retrieval from BKN / knowledge network |
| `tool` | Tool or operator invocation via Execution Factory |
| `llm` | LLM inference call (prompt → completion) |
| `memory` | Memory read/write operations |
| `policy` | Policy evaluation (permission checks, guardrails) |

### Interpreting Spans

Key attributes to look for in each span type:

**Retrieval spans:** `kn_id`, `query`, `results_count`, `sources` — tells you which knowledge network was queried, what the search query was, and which object types contributed results.

**Tool spans:** `tool_id`, `input`, `output`, `status` — shows exactly what was sent to each tool and what came back. If `status` is `error`, the `error` attribute contains the failure reason.

**LLM spans:** `model`, `prompt_tokens`, `completion_tokens`, `temperature` — useful for cost analysis and debugging generation quality. High token counts may indicate context bloat.

---

## 🐍 Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient()  # after kweaver auth login; see kweaver-sdk for client setup

# Full trace (agent id + conversation id)
trace = client.agent.trace("agt-001", "conv-xyz789")
print("trace_id:", trace["trace_id"])
print("duration:", trace["duration_ms"], "ms")
print("status:", trace["status"])

# Walk spans
for span in trace["spans"]:
    indent = "  " if span["parent_span_id"] else ""
    print(f"{indent}{span['name']} ({span['type']}) — {span['duration_ms']}ms — {span['status']}")
    if span["type"] == "retrieval":
        print(f"    query: {span['attributes']['query']}")
        print(f"    results: {span['attributes']['results_count']}")
        print(f"    sources: {span['attributes']['sources']}")
    elif span["type"] == "tool":
        print(f"    tool: {span['attributes']['tool_id']}")
        print(f"    input: {span['attributes']['input']}")
        print(f"    output: {span['attributes']['output']}")
    elif span["type"] == "llm":
        print(f"    model: {span['attributes']['model']}")
        print(f"    tokens: {span['attributes']['prompt_tokens']}+{span['attributes']['completion_tokens']}")

# Evidence chain: find all data sources that contributed to the response
retrieval_spans = [s for s in trace["spans"] if s["type"] == "retrieval"]
tool_spans = [s for s in trace["spans"] if s["type"] == "tool"]

print("\n--- Evidence Sources ---")
for rs in retrieval_spans:
    print(f"Knowledge Network: {rs['attributes']['kn_id']}")
    print(f"  Object Types: {rs['attributes']['sources']}")
    print(f"  Results: {rs['attributes']['results_count']}")

for ts in tool_spans:
    print(f"Tool: {ts['attributes']['tool_id']}")
    print(f"  Output: {ts['attributes']['output']}")
```

---

## 📘 TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = await KWeaverClient.connect();

// Get trace (agent id + conversation id)
const trace = await client.agent.trace('agt-001', 'conv-xyz789');
console.log('trace_id:', trace.traceId);
console.log('duration:', trace.durationMs, 'ms');

// Walk spans
for (const span of trace.spans) {
  const indent = span.parentSpanId ? '  ' : '';
  console.log(
    `${indent}${span.name} (${span.type}) — ${span.durationMs}ms — ${span.status}`,
  );

  if (span.type === 'retrieval') {
    console.log(`    query: ${span.attributes.query}`);
    console.log(`    results: ${span.attributes.resultsCount}`);
    console.log(`    sources: ${span.attributes.sources}`);
  } else if (span.type === 'tool') {
    console.log(`    tool: ${span.attributes.toolId}`);
    console.log(`    input:`, span.attributes.input);
    console.log(`    output:`, span.attributes.output);
  } else if (span.type === 'llm') {
    console.log(`    model: ${span.attributes.model}`);
    console.log(
      `    tokens: ${span.attributes.promptTokens}+${span.attributes.completionTokens}`,
    );
  }
}

// Evidence chain analysis
const retrievalSpans = trace.spans.filter((s) => s.type === 'retrieval');
const toolSpans = trace.spans.filter((s) => s.type === 'tool');

console.log('\n--- Evidence Sources ---');
retrievalSpans.forEach((rs) => {
  console.log(`Knowledge Network: ${rs.attributes.knId}`);
  console.log(`  Object Types: ${rs.attributes.sources}`);
});
toolSpans.forEach((ts) => {
  console.log(`Tool: ${ts.attributes.toolId}`);
  console.log(`  Output:`, ts.attributes.output);
});
```

---

## 🌐 curl

```bash
# Get trace for a conversation
curl -sk "https://<access-address>/api/agent-observability/v1/traces/conv-xyz789" \
  -H "Authorization: Bearer $(kweaver token)"

# Search traces by agent ID and time range
curl -sk -X POST "https://<access-address>/api/agent-observability/v1/traces/search" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agt-001",
    "since": "2026-04-01T00:00:00Z",
    "until": "2026-04-14T23:59:59Z",
    "limit": 20
  }'

# Get spans for a specific trace
curl -sk "https://<access-address>/api/agent-observability/v1/traces/tr-abc123/spans" \
  -H "Authorization: Bearer $(kweaver token)"

# Get a single span's details
curl -sk "https://<access-address>/api/agent-observability/v1/traces/tr-abc123/spans/sp-003" \
  -H "Authorization: Bearer $(kweaver token)"

# Search traces by status (find failed agent turns)
curl -sk -X POST "https://<access-address>/api/agent-observability/v1/traces/search" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "error",
    "since": "2026-04-13T00:00:00Z",
    "limit": 10
  }'

# Health check
curl -sk "https://<access-address>/api/agent-observability/v1/health" \
  -H "Authorization: Bearer $(kweaver token)"
```
