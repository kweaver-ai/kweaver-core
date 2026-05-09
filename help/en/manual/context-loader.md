# 📚 Context Loader

## 📖 Overview

The **Context Loader** (including **agent-retrieval** services) assembles **high-quality context** for Decision Agents: ontology-aware recall, ranking, and on-demand loading from BKN and data plane. It sits between raw data/VEGA and the agent runtime.

The Context Loader is also exposed as an **MCP server**, providing **MCP tools** that coding agents and LLM-based applications can use directly.

Typical ingress prefix:

| Prefix | Role |
| --- | --- |
| `/api/agent-retrieval/v1` | Retrieval and context assembly APIs |

**Related modules:** [BKN Engine](bkn.md), [VEGA Engine](vega.md), [Decision Agent](decision-agent.md).

---

## 🔌 MCP integration

The Context Loader exposes a standard [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server over Streamable HTTP transport. AI coding tools (Cursor, Claude Desktop, Cline, etc.) and custom agents can call all Context Loader capabilities directly via the MCP protocol.

### Endpoint URL

```
https://<access-address>/api/agent-retrieval/v1/mcp
```

### Configure in Cursor

Create `.cursor/mcp.json` in your project root (or globally at `~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "kweaver-context-loader": {
      "url": "https://<access-address>/api/agent-retrieval/v1/mcp",
      "headers": {
        "Authorization": "Bearer <token>"
      }
    }
  }
}
```

Get a token with `kweaver token`. Once saved, Cursor will auto-discover the MCP tools exposed by Context Loader, and the agent can call them directly in conversation.

### Configure in Claude Desktop

Edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "kweaver-context-loader": {
      "url": "https://<access-address>/api/agent-retrieval/v1/mcp",
      "headers": {
        "Authorization": "Bearer <token>"
      }
    }
  }
}
```

### Available MCP Tools

Once configured, MCP clients can discover and call these tools:

| Tool | Layer | Description |
|------|-------|-------------|
| `kn_search` | L1 | Semantic search across schema and instances |
| `kn_schema_search` | L1 | Search schema metadata only (discover candidate concepts) |
| `query_object_instance` | L2 | Query object instances with conditions |
| `query_instance_subgraph` | L2 | Query the relation subgraph around an instance |
| `get_logic_properties` | L3 | Get computed / derived property values |
| `get_action_info` | L3 | Get action type definition and parameter schema |

Every tool call requires the common parameter `kn_id` (knowledge network ID). Use `kweaver bkn list` to find it.

### Verify with CLI

You can verify the MCP server is working without configuring a full MCP client:

```bash
# Set the knowledge network
kweaver context-loader config set --kn-id kn_abc123

# List MCP tools
kweaver context-loader tools
```

---

## 💻 CLI

### Configuration Management

```bash
# Set active KWeaver server configuration
kweaver config set <alias> --url https://<access-address> --token <token>

# Use a saved configuration
kweaver config use <alias>

# List all saved configurations
kweaver config list

# Show current active configuration
kweaver config show

# Remove a saved configuration
kweaver config remove <alias>
```

### MCP Integration

The Context Loader exposes MCP (Model Context Protocol) endpoints for tool-aware LLM applications.

```bash
# List available MCP tools
kweaver mcp tools
```

### Knowledge Network Search

```bash
# Semantic search across a knowledge network
kweaver kn-search <kn_id> "quarterly revenue trends"
kweaver kn-search <kn_id> "customer churn risk factors" --limit 15
```

### Instance Queries

```bash
# Query a specific object instance by ID
kweaver query-object-instance <kn_id> <ot_id> <instance_id>

# Query the subgraph around an instance (neighbors, relations)
kweaver query-instance-subgraph <kn_id> <ot_id> <instance_id>
kweaver query-instance-subgraph <kn_id> <ot_id> <instance_id> --depth 2

# Get computed / logic properties for an instance
kweaver get-logic-properties <kn_id> <ot_id> <instance_id>
```

### Action Information

```bash
# Get full action type definition and parameter schema
kweaver get-action-info <kn_id> <action_id>
```

### End-to-End Example

```bash
# 1. Configure and authenticate
kweaver config set prod --url https://kweaver.example.com --token eyJ...
kweaver config use prod

# 2. Search the knowledge network for context
kweaver kn-search kn-001 "high-priority orders this month"

# 3. Drill into a specific instance
kweaver query-object-instance kn-001 ot-orders ord-5521

# 4. Explore its neighborhood
kweaver query-instance-subgraph kn-001 ot-orders ord-5521 --depth 2

# 5. Check logic properties
kweaver get-logic-properties kn-001 ot-orders ord-5521

# 6. Get action info before executing
kweaver get-action-info kn-001 act-escalate

# 7. List available MCP tools for agent integration
kweaver mcp tools
```

---

## Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")

# Semantic search across a knowledge network
results = client.context_loader.kn_search(
    kn_id="kn-001",
    query="quarterly revenue trends",
    limit=10,
)
for hit in results["items"]:
    print(hit["score"], hit["object_type"], hit["display_value"])

# Query a specific object instance
instance = client.context_loader.query_object_instance(
    kn_id="kn-001",
    ot_id="ot-orders",
    instance_id="ord-5521",
)
print(instance["properties"])

# Query the subgraph around an instance
subgraph = client.context_loader.query_instance_subgraph(
    kn_id="kn-001",
    ot_id="ot-orders",
    instance_id="ord-5521",
    depth=2,
)
for node in subgraph["nodes"]:
    print(node["id"], node["type"], node["display_value"])
for edge in subgraph["edges"]:
    print(edge["source"], "->", edge["target"], edge["relation_type"])

# Get logic (computed) properties
logic_props = client.context_loader.get_logic_properties(
    kn_id="kn-001",
    ot_id="ot-orders",
    instance_id="ord-5521",
)
for prop in logic_props:
    print(prop["name"], "=", prop["value"])

# Get action type information
action_info = client.context_loader.get_action_info(
    kn_id="kn-001",
    action_id="act-escalate",
)
print(action_info["name"], action_info["parameters"])

# List MCP tools
tools = client.context_loader.list_mcp_tools()
for tool in tools:
    print(tool["name"], tool["description"])
```

---

## TypeScript SDK

> More runnable examples ship with the `@kweaver-ai/kweaver-sdk` npm package.

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

// Auto-reads credentials from ~/.kweaver/
const client = await KWeaverClient.connect();

// Initialize Context Loader — requires the MCP endpoint URL and a knowledge network ID
const { baseUrl } = client.base();
const mcpUrl = `${baseUrl}/api/agent-retrieval/v1/mcp`;
const knId = 'kn-001';
const cl = client.contextLoader(mcpUrl, knId);

// Layer 1: Schema search — discover types by natural language
const schemaResults = await cl.schemaSearch({ query: 'quarterly revenue trends', max_concepts: 5 });
console.log('Schema hits:', schemaResults);

// Layer 2: Instance query via MCP
const otId = 'ot-orders';
const mcpInstances = await cl.queryInstances({
  ot_id: otId,
  limit: 20,
});
console.log('Instances:', mcpInstances);

// Direct Client API queries (more flexible than MCP layer)
const directInstances = await client.bkn.queryInstances(knId, otId, {
  page: 1,
  limit: 20,
});

// Subgraph traversal — expand along relation types
const relationTypes = await client.knowledgeNetworks.listRelationTypes(knId);
const rt = relationTypes.find(r => r.source_object_type?.id && r.target_object_type?.id);
if (rt) {
  const subgraph = await client.bkn.querySubgraph(knId, {
    relation_type_paths: [{
      relation_types: [{
        relation_type_id: rt.id,
        source_object_type_id: rt.source_object_type?.id,
        target_object_type_id: rt.target_object_type?.id,
      }],
    }],
    limit: 5,
  });
  console.log('Subgraph:', subgraph);
}
```

---

## curl

```bash
# Health check
curl -sk "https://<access-address>/api/agent-retrieval/v1/health" \
  -H "Authorization: Bearer $(kweaver token)"

# Semantic search across a knowledge network
curl -sk -X POST "https://<access-address>/api/agent-retrieval/v1/knowledge-networks/kn-001/search" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"query": "quarterly revenue trends", "limit": 10}'

# Query a specific object instance
curl -sk "https://<access-address>/api/agent-retrieval/v1/knowledge-networks/kn-001/object-types/ot-orders/instances/ord-5521" \
  -H "Authorization: Bearer $(kweaver token)"

# Query the subgraph around an instance
curl -sk -X POST "https://<access-address>/api/agent-retrieval/v1/knowledge-networks/kn-001/object-types/ot-orders/instances/ord-5521/subgraph" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"depth": 2}'

# Get logic properties
curl -sk "https://<access-address>/api/agent-retrieval/v1/knowledge-networks/kn-001/object-types/ot-orders/instances/ord-5521/logic-properties" \
  -H "Authorization: Bearer $(kweaver token)"

# Get action type information
curl -sk "https://<access-address>/api/agent-retrieval/v1/knowledge-networks/kn-001/actions/act-escalate/info" \
  -H "Authorization: Bearer $(kweaver token)"

# List MCP tools
curl -sk "https://<access-address>/api/agent-retrieval/v1/mcp/tools" \
  -H "Authorization: Bearer $(kweaver token)"
```
