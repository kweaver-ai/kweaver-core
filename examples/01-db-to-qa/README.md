# 01 · From Database to Intelligent Q&A

> Your database can finally answer questions — in plain language, not SQL.

## The Problem

A supply chain analyst has years of purchasing and inventory records locked in MySQL.
Every business question — "Which suppliers are most reliable?" "What's at risk of stockout?" —
means filing a request with the DBA and waiting hours for a custom query.

This example connects that database to a knowledge network and an Agent. Ask questions
in natural language, get answers grounded in your actual data.

## What This Example Does

```
MySQL Database
     │
     ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  Datasource │────▶│  Knowledge   │────▶│ Context-Loader  │
│  Connect    │     │  Network     │     │ Semantic Search  │
└─────────────┘     └──────────────┘     └─────────────────┘
                           │
                           ▼
                    ┌──────────────┐     ┌─────────────────┐
                    │   Schema     │     │   Agent Chat    │
                    │   Explore    │     │   (Q&A)         │
                    └──────────────┘     └─────────────────┘
```

0. **Seed** sample data into MySQL (`seed.sql` — fictional smart-home supply chain)
1. **Connect** a MySQL datasource to the platform
2. **Create & Build** a Knowledge Network from the datasource
3. **Explore** auto-discovered object types and properties
4. **Search** the knowledge network with natural language
5. **Chat** with an Agent to answer questions about the data

## Prerequisites

```bash
# 1. Install the KWeaver CLI
npm install -g @kweaver-ai/kweaver-sdk

# 2. Install the MySQL client (for Step 0: seed.sql runs on your machine)
#    macOS:  brew install mysql-client
#    Ubuntu: sudo apt install -y mysql-client

# 3. Authenticate to a KWeaver platform
kweaver auth login https://<platform-url>

# 4. Prepare a MySQL database reachable from the platform
#    The DB user must have CREATE TABLE / INSERT / SELECT rights.
```

## Quick Start

```bash
cp env.sample .env
# Fill in DB_HOST, DB_NAME, DB_USER, DB_PASS — see comments in env.sample
vim .env
./run.sh
```

> **Security:** `.env` is gitignored. Never commit credentials to version control.

## Configuration Notes

**`DB_HOST` vs `DB_HOST_SEED`**
Step 0 runs `mysql` on your local machine; Step 1 uses the platform's network to connect.
If your laptop uses a public IP but the platform needs a VPC internal IP, set `DB_HOST`
to the internal address and `DB_HOST_SEED` to the public one.

**`DEBUG=1`** in `.env` prints verbose output (API bodies, kweaver config). Passwords are never logged.

## Key Commands

```bash
kweaver ds connect mysql $DB_HOST $DB_PORT $DB_NAME \
  --account $DB_USER --password $DB_PASS --name "my-datasource"

kweaver bkn create-from-ds <datasource-id> --name "my-kn" --build

kweaver bkn object-type list <kn-id>
kweaver context-loader kn-search "supply chain" --kn-id <kn-id>
kweaver agent chat <agent-id> -m "What are the main suppliers?"
```

## Troubleshooting

**`ERROR 1044 Access denied`** — the DB user has no rights on `DB_NAME`. Ask your DBA to run
`GRANT ALL ON your_db.* TO 'your_user'@'%';`

## Cleanup

Resources (KN, datasource) are deleted automatically on exit. Manual cleanup:
```bash
kweaver bkn delete <kn-id> -y
kweaver ds delete <datasource-id> -y
```
