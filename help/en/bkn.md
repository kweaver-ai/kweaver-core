# BKN Engine

## Overview

The **Business Knowledge Network (BKN)** is the semantic layer of KWeaver Core. It models your domain with **object types**, **relation types**, and **action types**, stores **instances** and **relations**, and powers agents and analytics.

**Related modules:** [VEGA Engine](vega.md) (data behind views), [Context Loader](context-loader.md) (context from ontology), [Decision Agent](decision-agent.md) (uses BKN at runtime).

**Operations note:** Semantic search is handled by **bkn-backend** and **ontology-query** together. Register an embedding in the **model factory** and set the default small-model name on both sides to match the registered `model_name`. See [Model configuration — Enable BKN semantic search](model.md#enable-bkn-semantic-search).

---

## BKN Language

**BKN (Business Knowledge Network)** is a Markdown-based declarative modeling language for defining objects, relations, and actions in a business knowledge network. BKN describes model structure and semantics only — it contains no execution logic.

Full language specification: [bkn-specification](https://github.com/kweaver-ai/bkn-specification).

### Core Concepts

| Concept | Description |
|---------|-------------|
| **knowledge_network** | Top-level container for a business knowledge network |
| **object_type** | Business object type (e.g. Pod, Customer, Order) with properties and data source |
| **relation_type** | Link between two object types (e.g. "Pod belongs to Node", "Customer places Order") |
| **action_type** | Operation on an object, bindable to a tool or MCP |
| **risk_type** | Structured risk modeling for actions and objects |
| **concept_group** | Logical grouping of related object types |

### File Format

BKN files use the `.bkn` extension and UTF-8 encoding. Each file has two parts:

1. **YAML Frontmatter** — metadata wrapped in `---`
2. **Markdown Body** — definitions expressed with standard Markdown tables and headings

```markdown
---
type: object_type
id: pod
name: Pod
tags: [container, Kubernetes]
---

## ObjectType: Pod

The smallest deployable unit in Kubernetes.

### Data Properties

| Name | Display Name | Type | Description | Mapped Field |
|------|--------------|------|-------------|--------------|
| id | ID | integer | Primary key | id |
| pod_name | Pod Name | string | Pod name | pod_name |
| pod_status | Status | string | Running/Pending/Failed | pod_status |
| pod_node_name | Node | string | Node the pod runs on | pod_node_name |

### Keys

Primary Keys: id
Display Key: pod_name

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | pod_info_view | pod_info_view |
```

Heading levels are fixed: `#` for network title, `##` for type definitions (`ObjectType:` / `RelationType:` / `ActionType:`), `###` for sections within a type (Data Properties, Keys, Endpoint, etc.).

### Directory Layout

Each object, relation, action, and risk gets its own file, organized into typed subdirectories:

```
my-network/
├── network.bkn              # Root file (type: knowledge_network)
├── SKILL.md                 # Agent entry point (optional, agentskills.io standard)
├── object_types/
│   ├── customer.bkn
│   └── order.bkn
├── relation_types/
│   └── customer_places_order.bkn
├── action_types/
│   └── check_order_status.bkn
├── concept_groups/
│   └── ecommerce.bkn
└── data/                    # Optional CSV instance data
    └── customers.csv
```

### Update Model

BKN uses a **no-patch update model**:

- **Add / modify** — edit `.bkn` files and import; upserts by `(network, type, id)`
- **Delete** — use the SDK/CLI delete API explicitly; deletion is not expressed in BKN files

### SDK

Official SDKs for parsing, validating, and transforming BKN files:

| Language | Package | Install |
|----------|---------|---------|
| Python | [PyPI](https://pypi.org/project/kweaver-bkn/) | `pip install kweaver-bkn` |
| TypeScript | [npm](https://www.npmjs.com/package/@kweaver-ai/bkn) | `npm install @kweaver-ai/bkn` |
| Golang | — | `go get github.com/kweaver-ai/bkn-specification/sdk/golang` |

---

## Create BKN with an Agent

AI coding agents (Cursor, Claude Code, Codex, etc.) can generate spec-compliant BKN directories automatically via the [create-bkn](https://github.com/kweaver-ai/bkn-specification/tree/main/.cursor/skills/create-bkn) skill.

### Install the Skill

For Cursor, place the `create-bkn` skill directory under `~/.cursor/skills/` or `.cursor/skills/` in your project. Other agent environments follow their own skill-loading mechanism.

Also install the [kweaver-core](https://github.com/kweaver-ai/kweaver-sdk) skill for CLI validation and push capabilities.

### Describe Your Domain in Natural Language

Simply tell the agent what you need. For example:

> Build a supply-chain knowledge network with "Material", "Warehouse", and "Inventory" objects.
> Material and Warehouse are linked through Inventory.
> Add a "Stock Check" action bound to the Inventory object.

The agent will automatically:
1. Read the BKN specification to confirm syntax rules
2. Generate `network.bkn` and per-type `.bkn` files in subdirectories
3. Generate a `SKILL.md` index file (agent-readable network navigation)
4. Cross-check ID references, heading levels, and required fields

### Validate and Push

Once BKN files are generated, validate and push to the platform with the CLI:

```bash
# Validate: check format and referential integrity
kweaver bkn validate ./my-network/

# Push to platform (creates or updates the knowledge network)
kweaver bkn push ./my-network/

# Pull an existing network from the platform to local
kweaver bkn pull <kn_id> ./export-dir/
```

### End-to-End Example

```bash
# 1. Generate BKN directory in the agent (interactive)
#    → agent creates ./supply-chain/

# 2. Validate
kweaver bkn validate ./supply-chain/

# 3. Push to platform
kweaver bkn push ./supply-chain/

# 4. List the created network
kweaver bkn list

# 5. Build indexes
kweaver bkn build <kn_id> --wait

# 6. Verify with semantic search
kweaver bkn search <kn_id> "materials running low on stock"
```

---

## Create from Data Source

Instead of writing BKN files, you can generate a knowledge network directly from an existing data source:

```bash
# From a database: auto-discover table schemas and build indexes
kweaver bkn create-from-ds <ds_id> \
  --name "sales-network" \
  --tables orders,customers,products \
  --build --timeout 300

# From CSV files
kweaver bkn create-from-csv <ds_id> \
  --files "./data/*.csv" \
  --name "analytics-network" \
  --build
```

---

## CLI

### Knowledge Network Management

```bash
kweaver bkn list --name "order" --tag production --sort update_time --direction desc --limit 50 -v
kweaver bkn get <kn_id> --stats
kweaver bkn get <kn_id> --export
kweaver bkn pull <kn_id> ./export-dir/
```

### Build and Push

```bash
kweaver bkn build <kn_id> --wait --timeout 300
kweaver bkn validate ./my-network/
kweaver bkn push ./my-network/ --branch main
```

### Object Type CRUD

```bash
kweaver bkn object-type list <kn_id>
kweaver bkn object-type get <kn_id> <ot_id>

kweaver bkn object-type create <kn_id> \
  --name "Order" \
  --dataview-id <dv_id> \
  --primary-key "order_id" \
  --display-key "order_number" \
  --tags "commerce,core"

kweaver bkn object-type update <kn_id> <ot_id> \
  --add-property '{"name":"region","type":"string","display_name":"Region"}' \
  --update-property '{"name":"status","display_name":"Order Status"}' \
  --remove-property "legacy_field" \
  --tags "commerce,core,v2"

kweaver bkn object-type delete <kn_id> <ot_id>
```

### Query Object Instances

```bash
# Equality filter
kweaver bkn object-type query <kn_id> <ot_id> \
  '{"conditions":[{"field":"status","op":"==","value":"active"}],"limit":20}'

# Like filter
kweaver bkn object-type query <kn_id> <ot_id> \
  '{"conditions":[{"field":"name","op":"like","value":"%widget%"}],"limit":20}'

# IN filter
kweaver bkn object-type query <kn_id> <ot_id> \
  '{"conditions":[{"field":"region","op":"in","value":["US","EU"]}],"limit":20}'

# Compound: AND / OR
kweaver bkn object-type query <kn_id> <ot_id> '{
  "conditions": [
    {"field":"status","op":"==","value":"active"},
    {"field":"amount","op":">","value":1000}
  ],
  "logic": "and",
  "limit": 50
}'

# Pagination with search_after
kweaver bkn object-type query <kn_id> <ot_id> '{
  "limit": 50,
  "search_after": ["2026-04-01T00:00:00Z","order-9999"]
}'
```

### Relation Types

```bash
kweaver bkn relation-type list <kn_id>
kweaver bkn relation-type get <kn_id> <rt_id>

kweaver bkn relation-type create <kn_id> \
  --name "placed_by" \
  --source-type "Order" \
  --target-type "Customer" \
  --display-name "Placed By"

kweaver bkn relation-type delete <kn_id> <rt_id>
```

### Semantic Search

```bash
kweaver bkn search <kn_id> "overdue invoices in Q1"
kweaver bkn search <kn_id> "high-value customers" --limit 20
```

### Action Types and Execution

```bash
kweaver bkn action-type list <kn_id>
kweaver bkn action-type query <kn_id> '{"name":"calculate_risk"}'
kweaver bkn action-type execute <kn_id> <action_id> '{"input":{"customer_id":"C-1001"}}'

kweaver bkn action-log list <kn_id> --limit 20
kweaver bkn action-log get <kn_id> <log_id>
kweaver bkn action-log cancel <kn_id> <log_id>
kweaver bkn action-execution get <kn_id> <execution_id>
```

### End-to-End Example

```bash
# 1. Connect a MySQL data source
kweaver ds connect mysql db.example.com 3306 mydb --account root --password secret
# → ds_id: ds-abc123

# 2. Create a knowledge network from the data source
kweaver bkn create-from-ds ds-abc123 --name "ecommerce" --tables orders,customers --build --wait

# 3. Inspect the generated object types
kweaver bkn object-type list <kn_id>

# 4. Query customer instances
kweaver bkn object-type query <kn_id> <customer_ot_id> \
  '{"conditions":[{"field":"country","op":"==","value":"US"}],"limit":5}'

# 5. Semantic search across all types
kweaver bkn search <kn_id> "top spending customers"
```

---

## Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")

# List knowledge networks
networks = client.bkn.list_networks(name="order", sort="update_time", direction="desc", limit=50)
for kn in networks:
    print(kn["id"], kn["name"], kn["status"])

# Get details with stats
detail = client.bkn.get_network("kn-001", stats=True)
print(detail["object_type_count"], detail["instance_count"])

# Create from data source
kn = client.bkn.create_from_ds(
    ds_id="ds-abc123",
    name="ecommerce",
    tables=["orders", "customers"],
    build=True,
    timeout=300,
)

# Object-type CRUD
ot = client.bkn.create_object_type(
    kn_id="kn-001",
    name="Order",
    dataview_id="dv-xyz",
    primary_key="order_id",
    display_key="order_number",
)
client.bkn.update_object_type(
    kn_id="kn-001",
    ot_id=ot["id"],
    add_properties=[{"name": "region", "type": "string"}],
    tags=["commerce", "v2"],
)

# Query instances
results = client.bkn.query_instances(
    kn_id="kn-001",
    ot_id="ot-orders",
    conditions=[{"field": "status", "op": "==", "value": "active"}],
    limit=20,
)
for row in results["data"]:
    print(row)

# Semantic search
hits = client.bkn.search("kn-001", query="overdue invoices", limit=10)

# Execute an action
execution = client.bkn.execute_action(
    kn_id="kn-001",
    action_id="act-risk",
    input={"customer_id": "C-1001"},
)
print(execution["status"], execution["output"])
```

---

## TypeScript SDK

> Full runnable examples at [kweaver-sdk/examples](https://github.com/kweaver-ai/kweaver-sdk/tree/main/examples)

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = await KWeaverClient.connect();

const knList = await client.knowledgeNetworks.list({ limit: 50 });
for (const kn of knList) {
  console.log(`${kn.name} (${kn.id})`);
}

const knId = knList[0].id;
const detail = await client.knowledgeNetworks.get(knId, { include_statistics: true });

const objectTypes = await client.knowledgeNetworks.listObjectTypes(knId);
for (const ot of objectTypes) {
  console.log(`${ot.name} (${ot.id}) — ${ot.properties?.length ?? 0} properties`);
}

const relationTypes = await client.knowledgeNetworks.listRelationTypes(knId);
for (const rt of relationTypes) {
  console.log(`${rt.source_object_type?.name} —[${rt.name}]→ ${rt.target_object_type?.name}`);
}

const actionTypes = await client.knowledgeNetworks.listActionTypes(knId);

const otId = objectTypes[0].id;
const instances = await client.bkn.queryInstances(knId, otId, {
  page: 1,
  limit: 20,
});
console.log(instances.datas);

const identity = instances.datas[0]._instance_identity;
const properties = await client.bkn.queryProperties(knId, otId, { identity });

const rt = relationTypes[0];
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

const result = await client.bkn.semanticSearch(knId, 'overdue invoices');
for (const concept of result.concepts ?? []) {
  console.log(`${concept.concept_name} (score: ${concept.intent_score})`);
}

const atId = actionTypes[0].id;
const actionDetail = await client.bkn.queryAction(knId, atId, {});
const logs = await client.bkn.listActionLogs(knId, { atId, limit: 5 });

const buildStatus = await client.knowledgeNetworks.buildAndWait(knId, {
  timeout: 300_000,
  interval: 5_000,
});
```

---

## curl

```bash
# List knowledge networks
curl -sk "https://<access-address>/api/ontology-manager/v1/knowledge-networks?name=order&sort=update_time&direction=desc&limit=50" \
  -H "Authorization: Bearer $(kweaver token)"

# Get a single knowledge network
curl -sk "https://<access-address>/api/ontology-manager/v1/knowledge-networks/kn-001" \
  -H "Authorization: Bearer $(kweaver token)"

# Create a knowledge network
curl -sk -X POST "https://<access-address>/api/ontology-manager/v1/knowledge-networks" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ecommerce",
    "description": "E-commerce domain network",
    "ds_id": "ds-abc123",
    "tables": ["orders", "customers"]
  }'

# List object types
curl -sk "https://<access-address>/api/ontology-manager/v1/knowledge-networks/kn-001/object-types" \
  -H "Authorization: Bearer $(kweaver token)"

# Create an object type
curl -sk -X POST "https://<access-address>/api/ontology-manager/v1/knowledge-networks/kn-001/object-types" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Order",
    "dataview_id": "dv-xyz",
    "primary_key": "order_id",
    "display_key": "order_number",
    "properties": [
      {"name": "order_id", "type": "string"},
      {"name": "amount", "type": "number"},
      {"name": "status", "type": "string"}
    ]
  }'

# Query instances with conditions
curl -sk -X POST "https://<access-address>/api/ontology-query/v1/knowledge-networks/kn-001/object-types/ot-orders/instances/query" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "conditions": [
      {"field": "status", "op": "==", "value": "active"},
      {"field": "amount", "op": ">", "value": 1000}
    ],
    "logic": "and",
    "limit": 20,
    "search_after": null
  }'

# Semantic search across a network
curl -sk -X POST "https://<access-address>/api/ontology-query/v1/knowledge-networks/kn-001/search" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"query": "overdue invoices in Q1", "limit": 10}'

# List relation types
curl -sk "https://<access-address>/api/ontology-manager/v1/knowledge-networks/kn-001/relation-types" \
  -H "Authorization: Bearer $(kweaver token)"

# Execute an action type
curl -sk -X POST "https://<access-address>/api/bkn-backend/v1/knowledge-networks/kn-001/actions/act-risk/execute" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"input": {"customer_id": "C-1001"}}'

# Get action execution result
curl -sk "https://<access-address>/api/bkn-backend/v1/knowledge-networks/kn-001/action-executions/exec-001" \
  -H "Authorization: Bearer $(kweaver token)"

# Build / rebuild network indexes
curl -sk -X POST "https://<access-address>/api/bkn-backend/v1/knowledge-networks/kn-001/build" \
  -H "Authorization: Bearer $(kweaver token)"
```
