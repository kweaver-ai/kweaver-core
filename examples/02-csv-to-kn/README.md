# 02 - From CSV Files to Knowledge Network

End-to-end example: import local CSV files, build a Knowledge Network, explore
the graph, run semantic search, and chat with an Agent.

**No SQL knowledge required** — bring your own CSV files and let KWeaver
discover the schema and relationships automatically.

## What This Example Does

```
CSV Files (local)
     │
     ▼
┌─────────────────────┐     ┌──────────────┐
│  bkn create-from-csv │────▶│  Knowledge   │
│  (import + build)    │     │   Network    │
└─────────────────────┘     └──────┬───────┘
                                   │
              ┌────────────────────┼───────────────────┐
              ▼                    ▼                   ▼
       ┌────────────┐     ┌──────────────┐    ┌───────────────┐
       │   Schema   │     │   Subgraph   │    │  Agent Q&A    │
       │  Explore   │     │  Traversal   │    │ (with context)│
       └────────────┘     └──────────────┘    └───────────────┘
```

0. **Connect** a MySQL datasource (backing store for the imported tables)
1. **Import** CSV files and build a Knowledge Network — one command
2. **Explore** auto-discovered object types and their properties
3. **Query** object instances
4. **Traverse** the graph with subgraph queries (depth 2)
5. **Search** the knowledge schema semantically via context-loader
6. **Export** the KN definition to a portable file
7. **Chat** with an Agent to answer business questions

### Sample Data

The `data/` directory contains a fictional HR dataset:

| File | Contents |
|------|----------|
| `departments.csv` | 5 departments with budget and headcount |
| `employees.csv` | 16 employees with role, level, salary, manager |
| `projects.csv` | 8 projects with status, budget, owner |

## Prerequisites

```bash
# 1. Install the KWeaver CLI
npm install -g @kweaver-ai/kweaver-sdk

# 2. Authenticate to a KWeaver platform
kweaver auth login https://<platform-url>

# 3. MySQL database reachable from the platform
#    (used as backing store — the script creates the tables automatically)
```

## Quick Start

```bash
cd examples/02-csv-to-kn
cp env.sample .env
vim .env          # set DB_HOST, DB_NAME, DB_USER, DB_PASS
./run.sh
```

### Using Your Own CSV Files

Replace the files in `data/` with your own CSVs — KWeaver will discover the
schema automatically. Requirements:

- First row must be a header
- File name becomes the table (and object type) name
- All columns are imported; numeric columns are detected automatically

## What You'll See

```
=== Step 1: Connect datasource ===
  Datasource: d7abc...

=== Step 2: Import CSVs and build Knowledge Network ===
  Files: departments.csv employees.csv projects.csv
  Knowledge Network: d7def...
  Auto-discovered object types (3):
    - departments
    - employees
    - projects

=== Step 3: Explore schema ===
  Object types (3):
    - departments  (5 properties)  id=...
    - employees    (9 properties)  id=...
    - projects     (8 properties)  id=...

=== Step 5: Subgraph traversal ===
  Starting from instance: emp_001
  Graph: 12 nodes, 8 edges

=== Step 7: Export Knowledge Network ===
  Exported: 3 object types, 0 relation types
```

## Key CLI Commands Used

| Command | What it does |
|---------|-------------|
| `kweaver ds connect mysql ...` | Register MySQL as backing datasource |
| `kweaver bkn create-from-csv <ds-id> --files data/*.csv --build` | Import CSVs and build KN in one step |
| `kweaver bkn object-type list <kn-id>` | List auto-discovered object types |
| `kweaver bkn object-type query <kn-id> <ot-id> --limit 5` | Query instances |
| `kweaver bkn subgraph <kn-id> <instance-id> --depth 2` | Graph traversal |
| `kweaver context-loader kn-search "..." --kn-id <kn-id> --only-schema` | Semantic search |
| `kweaver bkn export <kn-id>` | Export KN definition |
| `kweaver agent chat <agent-id> -m "..."` | Chat with schema context |

## Differences from Example 01

| | 01-db-to-qa | 02-csv-to-kn |
|---|---|---|
| Data source | Existing MySQL database | Local CSV files |
| Ingestion | `ds connect` + `create-from-ds` | `create-from-csv` (one step) |
| Schema setup | Write SQL seed file | Just bring CSVs |
| Graph feature | Semantic search + Q&A | Subgraph traversal + export |
| Data domain | Supply chain (BOM, orders) | HR (employees, projects) |
