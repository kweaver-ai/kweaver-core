# 02 · From CSV Files to Knowledge Network

> Scattered spreadsheets, connected. No SQL required.

## The Problem

An HR director has employee, department, and project data split across three spreadsheets.
Understanding "who reports to whom" or "which projects are at risk of being understaffed"
means manual VLOOKUP chaining across files — tedious and error-prone.

This example imports those files into a knowledge network. Relationships are discovered
automatically. You can traverse the org chart, run semantic queries, and ask an Agent
business questions about your people and projects.

## What This Example Does

```
CSV Files (local)
     │
     ▼
┌─────────────────────┐     ┌──────────────┐
│  bkn create-from-csv │────▶│  Knowledge   │
│  (import + build)    │     │  Network     │
└─────────────────────┘     └──────┬───────┘
                                   │
              ┌────────────────────┼───────────────────┐
              ▼                    ▼                   ▼
       ┌────────────┐     ┌──────────────┐    ┌───────────────┐
       │   Schema   │     │   Subgraph   │    │  Agent Q&A    │
       │  Explore   │     │  Traversal   │    │               │
       └────────────┘     └──────────────┘    └───────────────┘
```

0. **Connect** a MySQL datasource (backing store for the imported tables)
1. **Import** CSV files and build a Knowledge Network — one command
2. **Explore** auto-discovered object types and properties
3. **Query** object instances
4. **Traverse** the network with subgraph queries (depth 2)
5. **Search** the knowledge network semantically via context-loader
6. **Export** the KN definition to a portable file
7. **Chat** with an Agent to answer business questions

### Sample Data

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

# 3. Prepare a MySQL database reachable from the platform
#    (the script creates tables automatically — no manual SQL needed)
```

## Quick Start

```bash
cp env.sample .env
# Fill in DB_HOST, DB_NAME, DB_USER, DB_PASS — see comments in env.sample
vim .env
./run.sh
```

> **Security:** `.env` is gitignored. Never commit credentials to version control.

### Using Your Own CSV Files

Replace the files in `data/` with your own CSVs. Requirements:
- First row must be a header
- File name becomes the table (and object type) name
- All columns are imported; numeric columns are detected automatically

## Key Commands

| Command | What it does |
|---------|-------------|
| `kweaver ds connect mysql ...` | Register MySQL as backing datasource |
| `kweaver bkn create-from-csv <ds-id> --files data/*.csv --build` | Import CSVs and build KN in one step |
| `kweaver bkn object-type list <kn-id>` | List auto-discovered object types |
| `kweaver bkn object-type query <kn-id> <ot-id> --limit 5` | Query instances |
| `kweaver bkn subgraph <kn-id> <instance-id> --depth 2` | Network traversal |
| `kweaver context-loader kn-search "..." --only-schema` | Semantic schema search |
| `kweaver bkn export <kn-id>` | Export KN definition |
| `kweaver agent chat <agent-id> -m "..."` | Chat with schema context |

## Differences from Example 01

| | 01-db-to-qa | 02-csv-to-kn |
|---|---|---|
| Data source | Existing MySQL database | Local CSV files |
| Ingestion | `ds connect` + `create-from-ds` | `create-from-csv` (one step) |
| Schema setup | Write SQL seed file | Just bring CSVs |
| Network feature | Semantic search + Q&A | Subgraph traversal + export |
| Data domain | Supply chain (BOM, orders) | HR (employees, projects) |
