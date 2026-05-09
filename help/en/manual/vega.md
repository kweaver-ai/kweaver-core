# 🗄️ VEGA Engine

## 📖 Overview

**VEGA** provides **data virtualization** over heterogeneous sources: **data connections**, **models**, and **views** (including atomic and composite views). Agents and applications query through a unified SQL-oriented surface instead of wiring each source by hand.

Ingress prefix (typical):

| Prefix | Role |
| --- | --- |
| `/api/vega-backend/v1` | VEGA backend — connections, metadata, query execution |

**Related modules:** [BKN Engine](bkn.md) (semantic layer on top of data), [Context Loader](context-loader.md), [Dataflow](dataflow.md) (pipelines that land or transform data).

The **curl** section at the end of this page is **optional** — use it only if you need raw HTTP or shell scripts. If you rely on the **`kweaver` CLI** or language SDKs, you can skip it.

---

## 💻 CLI

Common flags for all `kweaver vega` subcommands: `-bd` / `--biz-domain <s>` (default from `kweaver config`), `--pretty` (pretty-print JSON, default on). Run `kweaver vega --help` for the full command tree.

### Health and diagnostics

```bash
# Reachability probe (Node CLI: GET .../catalogs?limit=1 with auth)
kweaver vega health

# Catalog count (lists up to 100 catalogs and counts entries)
kweaver vega stats

# Health probe JSON + catalog_count (same catalog list cap)
kweaver vega inspect
```

The **npm** CLI does not call `GET /health` on the vega-backend pod; it uses an authenticated **catalogs list** probe. The **Python** SDK’s `client.vega.health()` calls `GET /api/vega-backend/v1/health` when that route is exposed behind your ingress.

### Catalog management

```bash
# List catalogs (optional filters)
kweaver vega catalog list
kweaver vega catalog list --status healthy --limit 20

# Get one catalog
kweaver vega catalog get <catalog_id>

# Health status for one or more catalogs, or all
kweaver vega catalog health cat_pg001 cat_mysql002
kweaver vega catalog health --all

# Test connectivity for an existing catalog (registered in Vega)
kweaver vega catalog test-connection <catalog_id>

# Discover metadata; optional wait
kweaver vega catalog discover <catalog_id>
kweaver vega catalog discover <catalog_id> --wait

# Resources under a catalog
kweaver vega catalog resources <catalog_id>
kweaver vega catalog resources <catalog_id> --category table --limit 30

# Create / update / delete catalogs
kweaver vega catalog create \
  --name my-mysql \
  --connector-type mysql \
  --connector-config '{"host":"db.example.com","port":3306,"database":"mydb","username":"u","password":"p"}'

kweaver vega catalog update <catalog_id> --name new-name --connector-config '{"host":"..."}'

kweaver vega catalog delete <catalog_id> [<catalog_id> ...]   # prompts unless -y
kweaver vega catalog delete cat_a,cat_b -y
```

### Resource operations

There is **no** `kweaver vega resource preview` subcommand. Use **`resource query`** with a small `limit` to sample rows.

```bash
# List resources (optional filters)
kweaver vega resource list
kweaver vega resource list --catalog-id <catalog_id> --category table --limit 50

# List all resources (GET .../resources/list)
kweaver vega resource list-all [--limit N] [--offset N]

kweaver vega resource get <resource_id>

# Structured data query (POST .../resources/:id/data)
kweaver vega resource query <resource_id> \
  -d '{"limit":10,"offset":0,"need_total":true}'

# Create / update / delete resources
kweaver vega resource create \
  --catalog-id <catalog_id> \
  --name my_table \
  --category table \
  [--source-identifier <si>] [--database <db>] [-d '{"extra":"fields"}']

kweaver vega resource update <resource_id> [--name X] [--status X] [--tags t1,t2] [-d '{"k":"v"}']

kweaver vega resource delete <resource_id> [<resource_id> ...] [-y]
```

### Dataset (documents and build)

For dataset-type resources, manage indexed documents and async build jobs:

```bash
kweaver vega dataset create-docs <resource_id> -d '[{"id":"doc1",...},...]'
kweaver vega dataset update-docs <resource_id> -d '[{"id":"doc1",...},...]'
kweaver vega dataset delete-docs <resource_id> <doc_id> [<doc_id> ...]
kweaver vega dataset delete-docs-query <resource_id> -d '{"filter":...}'

kweaver vega dataset build <resource_id> [--mode full|incremental|realtime]
kweaver vega dataset build-status <resource_id> <task_id>
```

### Structured query and SQL (vega-backend)

Both commands below use **`vega-backend`** only and **do not** require `vega-calculate-coordinator` (Trino). Use them on Core-only installs with MySQL/PostgreSQL catalogs.

**Structured query** — `POST /api/vega-backend/v1/query/execute`

```bash
kweaver vega query execute -d '<json>'
```

Body highlights: `tables` (required: `resource_id` + optional `alias`), `joins` (multi-table within one catalog), `output_fields`, `filter_condition`, `sort`, `offset` / `limit` (max 10000), `need_total`. Omit `query_id` on the first page; reuse it when paging. In `joins[].on`, **`left_field` / `right_field` must match `schema_definition[].name` from `kweaver vega resource get`**. All tables must share one catalog (501 otherwise).

Common `filter_condition` operations: `==`/`eq`, `!=`/`not_eq`, `>`/`gt`, `>=`/`gte`, `<`/`lt`, `<=`/`lte`, `in`/`not_in`, `like`/`not_like` (only if the field is typed as string in schema), `range`, `null`/`not_null`, nested `and`/`or` via `sub_conditions`. Leaf nodes usually include `field`, `operation`, `value`, `value_from` (`"const"` for literals).

Single-table example:

```bash
kweaver vega query execute -d '{"tables":[{"resource_id":"res_mysql_supplier"}],"limit":5,"need_total":true}'
```

Two-table JOIN (replace IDs and field names):

```bash
kweaver vega query execute -d '{
  "tables": [
    {"resource_id":"res_a","alias":"a"},
    {"resource_id":"res_b","alias":"b"}
  ],
  "joins":[{"type":"inner","left_table_alias":"a","right_table_alias":"b","on":[{"left_field":"fk_id","right_field":"id"}]}],
  "output_fields":["a.name","b.amount"],
  "limit":10
}'
```

**Direct SQL** — `POST /api/vega-backend/v1/resources/query`

**Simple mode** (no JSON body): pass engine type and query as separate flags (quote the SQL).

```bash
kweaver vega sql --resource-type mysql --query "SELECT * FROM {{.res_mysql_supplier}} LIMIT 5"
```

**Advanced mode**: full JSON body; when `-d` is present, `--resource-type` / `--query` are ignored.

```bash
kweaver vega sql -d '<json>'
kweaver vega sql --help
```

Required in the JSON body: `query` (SQL string or OpenSearch DSL object), `resource_type` (`mysql` | `mariadb` | `postgresql` | `opensearch`). Optional: `stream_size` (100–10000), `query_timeout` (seconds 1–3600), `query_id`.

Placeholders: `{{.<resource_id>}}` or `{{<resource_id>}}` (Vega resource id) are replaced with the resource’s physical table id. You may also run **native SQL** without placeholders if table names are valid for the engine.

**Comparison**

| Approach | Entry | Depends on | Typical use |
|----------|-------|------------|---------------|
| Structured | `kweaver vega query execute` | vega-backend | Same-catalog JOINs, filter DSL |
| Direct SQL | `kweaver vega sql` | vega-backend | Complex SQL, aggregations, placeholders |
| Resource data | `kweaver vega resource query <id> -d {...}` | vega-backend | Single resource, filters, `search_after` |
| Dataview `--sql` | `kweaver dataview query ... --sql` | mdl-uniquery + **Trino** (Etrino) | Cross-engine SQL via coordinator |

TypeScript: `client.vega.executeQuery(jsonString)` and `client.vega.sqlQuery(jsonString)`.  
Python: `client.vega.query.execute(...)` and `client.vega.query.sql_query({...})`.

### Connector types

```bash
kweaver vega connector-type list
kweaver vega connector-type get mysql

kweaver vega connector-type register -d '{"name":"custom",...}'
kweaver vega connector-type update <type> -d '{...}'
kweaver vega connector-type delete <type> [-y]
kweaver vega connector-type enable <type> --enabled true
```

### Dataview operations

Data views are served by **mdl-uniquery** (not the vega-backend router). Use the **`dataview`** command group:

```bash
kweaver dataview list
kweaver dataview find --name "order"
kweaver dataview get <dataview_id>
# In --sql, FROM must be fully qualified: use meta_table_name from get; mysql_demo."sales"."orders" is illustrative
kweaver dataview query <dataview_id> --sql "SELECT order_id, amount FROM mysql_demo.\"sales\".\"orders\" WHERE status = 'active' LIMIT 10"
kweaver dataview query <dataview_id> --sql "SELECT COUNT(*) AS total FROM mysql_demo.\"sales\".\"orders\""
```

**Custom SQL (`--sql`) and Etrino**: Without `--sql`, `dataview query` uses the view’s stored definition and talks to the data source directly. With `--sql`, traffic goes through **`vega-calculate-coordinator`** (Hetu/Presto–style engine), which is **not** in the default KWeaver Core manifest. Install the **Etrino** charts: `vega-hdfs`, `vega-calculate` (includes the coordinator), and `vega-metadata`. Run `./deploy.sh etrino install` from the `deploy` directory to install Etrino only. **Use fully-qualified `catalog."schema"."table"` names for ad-hoc SQL.** See **Optional: Etrino** in [Install and deploy](../install.md).

**`dataview get` response fields (for custom `--sql`)**: The JSON from `kweaver dataview get <view_id> --pretty` includes the following; names match REST and the TypeScript / Python SDK.

| Field | Role |
|-------|------|
| **`meta_table_name`** | **Required for ad-hoc SQL**: the fully-qualified table name string (`catalog."schema"."table"`). Use it verbatim in `FROM` / `JOIN`; do not guess the catalog from the view’s logical name. |
| **`sql_str`** | The stored view SQL; cross-check table references with `meta_table_name`. |
| **`fields`** | Column metadata (`name`, `type`, etc.); **does not** include the fully-qualified table name. |

**Where `meta_table_name` / fully-qualified names come from**: With `--sql`, the engine resolves identifiers in **Trino/Hetu style** as **`catalog."schema"."table"`**:

- **catalog**: The **Vega catalog id** for that data source (same as the **id** from `kweaver vega catalog list`; often prefixed by connector family, e.g. `mysql_…`, depending on the environment).
- **schema**: The source **namespace**—for MySQL this is usually the **database** name; for PostgreSQL, often the **schema** name; exact semantics follow the connector metadata.
- **table**: Physical table (or view) name.

When a data view is created, the platform materializes this into **`meta_table_name`** (and the built-in **`sql_str`**). **Do not guess the catalog segment from the view’s logical name alone**; run **`kweaver dataview get <view_id>`** and reuse **`meta_table_name`** verbatim in `FROM` / `JOIN`, or match the table references in **`sql_str`**. Multi-table joins must stay within one data source (one catalog).

TypeScript: `client.dataviews.list()`, `client.dataviews.find(...)`, `client.dataviews.query(id, { sql: '...' })`.  
Python: `client.dataviews.list()`, `client.dataviews.query(...)`, etc.

### End-to-end example

```bash
kweaver vega health
kweaver vega connector-type list
kweaver vega catalog health --all
kweaver vega catalog discover <catalog_id> --wait
kweaver vega catalog resources <catalog_id> --category table
kweaver vega resource query <resource_id> -d '{"limit":5,"need_total":true}'
kweaver dataview find --name "orders"
kweaver dataview query <dataview_id> --sql "SELECT customer_id, SUM(amount) AS total FROM mysql_demo.\"sales\".\"orders\" GROUP BY customer_id LIMIT 10"
```

---

## 🐍 Python SDK

Use **`client.vega.*`** for vega-backend (nested resources: `catalogs`, `resources`, `query`, `connector_types`, etc.). **Catalog/resource CRUD and dataset build APIs** are not yet on the Python `VegaCatalogsResource` / `VegaResourcesResource`; use the **CLI** or **TypeScript** client below, or call the REST paths in the curl section.

```python
from kweaver import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")

# Health (GET /api/vega-backend/v1/health when exposed)
health = client.vega.health()
print(health.server_name, health.server_version)

# Composite stats (best-effort counts; see VegaPlatformStats)
stats = client.vega.stats()
print(f"catalog_count={stats.catalog_count}, data_view_count={stats.data_view_count}")

# Catalogs
catalogs = client.vega.catalogs.list(status="healthy", limit=20)
for cat in catalogs:
    print(cat.id, cat.name, getattr(cat, "health_status", None) or getattr(cat, "health_check_status", None))

cat = client.vega.catalogs.get("cat-001")
hs = client.vega.catalogs.health_status(["cat-001", "cat-002"])
ok = client.vega.catalogs.test_connection("cat-001")
client.vega.catalogs.discover("cat-001", wait=True)
resources_in_cat = client.vega.catalogs.resources("cat-001", category="table", limit=50)

# Resources
resources = client.vega.resources.list(catalog_id="cat-001", category="table", limit=50)
res = client.vega.resources.get("res_orders_001")
rows = client.vega.resources.data("res_orders_001", body={"limit": 10, "offset": 0, "need_total": True})

# Structured query + direct SQL (vega-backend)
q = client.vega.query.execute(tables=["res_orders_001"], limit=5, need_total=True)
sql_rows = client.vega.query.sql_query({
    "resource_type": "mysql",
    "query": "SELECT 1 AS one",
})

# Connector types
for ct in client.vega.connector_types.list():
    print(ct.type, getattr(ct, "name", ""))

# Data views (mdl-uniquery — separate from vega-backend)
for dv in client.dataviews.list():
    print(dv.id, dv.name)
result = client.dataviews.query(
    "dv_001",
    sql='SELECT * FROM mysql_demo."sales"."orders" LIMIT 5',  # FROM: use meta_table_name from dataview get; illustrative name here
)
```

---

## 📘 TypeScript SDK

`client.vega` uses a **flat** method surface (`listCatalogs`, `createCatalog`, …). JSON bodies for `executeQuery`, `sqlQuery`, `createResource`, `updateCatalog`, etc. are **strings** (`JSON.stringify(...)`).

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });

// Health: CLI-style probe result { status, probe, statusCode }
const health = await client.vega.health();
console.log(health);

const catalogs = await client.vega.listCatalogs({ status: 'healthy', limit: 20 });
catalogs.forEach((c: any) => console.log(c.id, c.name));

const detail = await client.vega.getCatalog('cat-001');
const healthStatus = await client.vega.catalogHealthStatus('cat-001,cat-002');
const test = await client.vega.testCatalogConnection('cat-001');
await client.vega.discoverCatalog('cat-001', { wait: true });
const catRes = await client.vega.listCatalogResources('cat-001', { category: 'table', limit: 50 });

await client.vega.createCatalog({
  name: 'my-mysql',
  connector_type: 'mysql',
  connector_config: { host: 'db.example.com', port: 3306, database: 'mydb', username: 'u', password: 'p' },
});
await client.vega.updateCatalog('cat-001', JSON.stringify({ name: 'renamed' }));
await client.vega.deleteCatalogs('cat-001');

const resources = await client.vega.listResources({ catalogId: 'cat-001', limit: 50 });
const allRes = await client.vega.listAllResources({ limit: 100 });
const res = await client.vega.getResource('res-001');
const data = await client.vega.queryResourceData('res-001', JSON.stringify({ limit: 5, need_total: true }));

await client.vega.createResource(JSON.stringify({
  catalog_id: 'cat-001', name: 't', category: 'table',
}));
await client.vega.updateResource('res-001', JSON.stringify({ status: 'active' }));
await client.vega.deleteResources('res-001');

await client.vega.createDatasetDocs('res-ds', JSON.stringify([{ id: 'd1' }]));
await client.vega.buildDataset('res-ds', 'full');
const build = await client.vega.getDatasetBuildStatus('res-ds', '<task-id>');

const structured = await client.vega.executeQuery(JSON.stringify({
  tables: [{ resource_id: 'res-001' }],
  limit: 5,
  need_total: true,
}));
const sqlOut = await client.vega.sqlQuery(JSON.stringify({
  resource_type: 'mysql',
  query: 'SELECT 1 AS one',
}));

const connectors = await client.vega.listConnectorTypes();

const dvList = await client.dataviews.list({ limit: 50 });
const dvResult = await client.dataviews.query('dv-001', {
  sql: "SELECT order_id, amount FROM mysql_demo.\"sales\".\"orders\" WHERE status = 'active' LIMIT 10", // FROM: meta_table_name from dataview get
});
```

---

## 🌐 curl

After `kweaver auth login`, use **`Authorization: Bearer $(kweaver token)`** for protected calls. Replace **`https://<access-address>`** with your deployment URL.

```bash
# Probe catalogs list (same idea as `kweaver vega health` in Node CLI)
curl -sk "https://<access-address>/api/vega-backend/v1/catalogs?limit=1" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "x-business-domain: bd_public"

# Optional: raw pod health (path is /health on vega-backend, not under /v1)
# curl -sk "https://<access-address>/health" -H "Authorization: Bearer $(kweaver token)"

# List / get catalogs
curl -sk "https://<access-address>/api/vega-backend/v1/catalogs?status=healthy&limit=20" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"
curl -sk "https://<access-address>/api/vega-backend/v1/catalogs/cat-001" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"

# Create / update / delete catalog
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/catalogs" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"name":"my","connector_type":"mysql","connector_config":{"host":"h","port":3306,"database":"d","username":"u","password":"p"}}'
curl -sk -X PUT "https://<access-address>/api/vega-backend/v1/catalogs/cat-001" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"name":"new-name"}'
curl -sk -X DELETE "https://<access-address>/api/vega-backend/v1/catalogs/cat-001" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"

# Catalog health / test-connection / discover / resources
curl -sk "https://<access-address>/api/vega-backend/v1/catalogs/cat-001/health-status" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/catalogs/cat-001/test-connection" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/catalogs/cat-001/discover" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"
curl -sk "https://<access-address>/api/vega-backend/v1/catalogs/cat-001/resources?category=table&limit=30" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"

# Resources: list, list-all, get, create, update, delete, data
curl -sk "https://<access-address>/api/vega-backend/v1/resources?catalog_id=cat-001&limit=50" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/resources" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"catalog_id":"cat-001","name":"t","category":"table"}'
curl -sk -X PUT "https://<access-address>/api/vega-backend/v1/resources/res-001" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"status":"active"}'
curl -sk -X DELETE "https://<access-address>/api/vega-backend/v1/resources/res-001" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"

curl -sk -X POST "https://<access-address>/api/vega-backend/v1/resources/res-001/data" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -H "x-http-method-override: GET" \
  -d '{"limit":10,"offset":0,"need_total":true}'

# Dataset docs (use POST override)
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/resources/res-ds/data" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -H "x-http-method-override: POST" \
  -d '[{"id":"doc1","content":"..."}]'

# Dataset build task
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/build-tasks" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"resource_id":"res-ds","mode":"full"}'
curl -sk "https://<access-address>/api/vega-backend/v1/build-tasks/<task-id>" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"

# Structured query / direct SQL
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/query/execute" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"tables":[{"resource_id":"res-001"}],"limit":5,"need_total":true}'
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/resources/query" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"resource_type":"mysql","query":"SELECT 1 AS one"}'

# Connector types
curl -sk "https://<access-address>/api/vega-backend/v1/connector-types" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"
curl -sk "https://<access-address>/api/vega-backend/v1/connector-types/mysql" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public"
# curl -sk -X POST "https://<access-address>/api/vega-backend/v1/connector-types" \
#   -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
#   -H "Content-Type: application/json" \
#   -d '<connector-type-json>'
curl -sk -X POST "https://<access-address>/api/vega-backend/v1/connector-types/mysql/enabled" \
  -H "Authorization: Bearer $(kweaver token)" -H "x-business-domain: bd_public" \
  -H "Content-Type: application/json" \
  -d '{"enabled":true}'
```

Dataview HTTP paths are defined by **mdl-uniquery**, not vega-backend; use `kweaver dataview` or the `client.dataviews` SDK.

Full details: npm package `@kweaver-ai/kweaver-sdk` and `kweaver vega --help` / `kweaver vega <subcommand> --help`.
