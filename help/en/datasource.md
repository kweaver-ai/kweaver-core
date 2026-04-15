# Data Source Management

## Overview

**Data Source Management (DS)** handles **connection registration**, **table discovery**, **CSV import**, and **lifecycle maintenance** for external databases. It is the prerequisite for building Knowledge Networks (BKN) — first connect a database to the platform, then use `bkn create-from-ds` or `bkn create-from-csv` to turn tables into a knowledge network.

Ingress prefix (typical):

| Prefix | Role |
| --- | --- |
| `/api/builder/v1` | Data source connections, discovery, and management |

**Related modules:** [BKN Engine](bkn.md) (create knowledge networks from data sources), [VEGA Engine](vega.md) (data virtualization and query), [Dataflow](dataflow.md) (data pipelines and transformation).

## Supported Database Types

mysql, postgresql, sqlserver, oracle, clickhouse, hive, opensearch, elasticsearch, and more. Run `kweaver vega connector-type list` to see which connector types are installed on your platform.

## CLI

### Connect a Data Source

```bash
# Connect to MySQL
kweaver ds connect mysql db.example.com 3306 erp \
  --account root --password pass123
# → returns ds_id, e.g. ds-abc123

# Connect to PostgreSQL with schema and custom name
kweaver ds connect postgresql pg.example.com 5432 analytics \
  --account reader --password pass456 \
  --schema public --name "analytics-db"
```

Argument order: `<db_type> <host> <port> <database>`. Use `--account` and `--password` for credentials.

### List and Inspect Data Sources

```bash
# List all data sources
kweaver ds list

# Search by keyword
kweaver ds list --keyword "erp"

# Filter by type
kweaver ds list --type mysql

# Get details for a single data source
kweaver ds get ds-abc123
```

### Discover Tables

```bash
# List all tables in a data source
kweaver ds tables ds-abc123

# Search tables by keyword
kweaver ds tables ds-abc123 --keyword "order"
```

### Import CSV

Upload local CSV files into an existing data source (database), then use them to build a knowledge network.

```bash
# Import multiple CSV files
kweaver ds import-csv ds-abc123 --files "materials.csv,inventory.csv"

# Use glob patterns with a table prefix
kweaver ds import-csv ds-abc123 --files "*.csv" --table-prefix sc_

# Recreate existing tables (when column structure changed)
kweaver ds import-csv ds-abc123 --files "materials.csv" --recreate

# Adjust batch size for large files
kweaver ds import-csv ds-abc123 --files "big-table.csv" --batch-size 1000
```

| Parameter | Required | Default | Description |
| --- | --- | --- | --- |
| `datasource_id` | yes | — | Target data source ID |
| `--files` | yes | — | CSV file paths, comma-separated or glob |
| `--table-prefix` | no | `""` | Prefix for generated table names |
| `--batch-size` | no | 500 | Rows per write batch (1–10000) |
| `--recreate` | no | off | Send overwrite on first batch to recreate table |

### Delete a Data Source

```bash
# Delete (with confirmation prompt)
kweaver ds delete ds-abc123

# Skip confirmation
kweaver ds delete ds-abc123 --yes
```

### End-to-End Example

```bash
# 1. Connect to a database
kweaver ds connect mysql db.example.com 3306 erp \
  --account root --password pass123
# → ds-abc123

# 2. Discover tables
kweaver ds tables ds-abc123

# 3. Create a knowledge network from the data source
kweaver bkn create-from-ds ds-abc123 \
  --name "erp-supply-chain" \
  --tables "orders,products,customers" \
  --build --timeout 600

# 4. Verify the knowledge network
kweaver bkn object-type list <kn_id>
kweaver bkn search <kn_id> "overdue orders"
```

---

## Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")

# List data sources
ds_list = client.ds.list()
for ds in ds_list["data"]:
    print(ds["id"], ds["name"], ds["type"], ds["status"])

# Filter by keyword and type
ds_list_filtered = client.ds.list(keyword="erp", type="mysql")

# Get details
detail = client.ds.get("ds-abc123")
print(f"host: {detail['host']}, database: {detail['database']}, status: {detail['status']}")

# Connect a new data source
new_ds = client.ds.connect(
    type="mysql",
    host="db.example.com",
    port=3306,
    database="erp",
    account="root",
    password="pass123",
)
print(f"data source ID: {new_ds['id']}")

# Discover tables
tables = client.ds.tables("ds-abc123")
for t in tables["data"]:
    print(t["name"], t["columns"])

# Search tables by keyword
tables_filtered = client.ds.tables("ds-abc123", keyword="order")

# Import CSV
import_result = client.ds.import_csv(
    datasource_id="ds-abc123",
    files=["materials.csv", "inventory.csv"],
    table_prefix="sc_",
    batch_size=500,
)
print(f"imported tables: {import_result['tables']}")

# Delete
client.ds.delete("ds-abc123")
```

---

## TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });

// List data sources
const dsList = await client.ds.list();
dsList.data.forEach((ds) => console.log(ds.id, ds.name, ds.type, ds.status));

// Filter
const dsFiltered = await client.ds.list({ keyword: 'erp', type: 'mysql' });

// Get details
const detail = await client.ds.get('ds-abc123');
console.log('host:', detail.host, 'database:', detail.database, 'status:', detail.status);

// Connect a new data source
const newDs = await client.ds.connect({
  type: 'mysql',
  host: 'db.example.com',
  port: 3306,
  database: 'erp',
  account: 'root',
  password: 'pass123',
});
console.log('data source ID:', newDs.id);

// Discover tables
const tables = await client.ds.tables('ds-abc123');
tables.data.forEach((t) => console.log(t.name, t.columns));

// Import CSV
const importResult = await client.ds.importCsv({
  datasourceId: 'ds-abc123',
  files: ['materials.csv', 'inventory.csv'],
  tablePrefix: 'sc_',
  batchSize: 500,
});
console.log('imported tables:', importResult.tables);

// Delete
await client.ds.delete('ds-abc123');
```

---

## curl

```bash
# List data sources
curl -sk "https://<access-address>/api/builder/v1/datasources?page=1&size=20" \
  -H "Authorization: Bearer $(kweaver token)"

# Filter by type
curl -sk "https://<access-address>/api/builder/v1/datasources?type=mysql&page=1&size=20" \
  -H "Authorization: Bearer $(kweaver token)"

# Get data source details
curl -sk "https://<access-address>/api/builder/v1/datasources/ds-abc123" \
  -H "Authorization: Bearer $(kweaver token)"

# Connect a new data source
curl -sk -X POST "https://<access-address>/api/builder/v1/datasources" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "db.example.com",
    "port": 3306,
    "database": "erp",
    "account": "root",
    "password": "pass123"
  }'

# Discover tables
curl -sk "https://<access-address>/api/builder/v1/datasources/ds-abc123/tables" \
  -H "Authorization: Bearer $(kweaver token)"

# Delete a data source
curl -sk -X DELETE "https://<access-address>/api/builder/v1/datasources/ds-abc123" \
  -H "Authorization: Bearer $(kweaver token)"
```
