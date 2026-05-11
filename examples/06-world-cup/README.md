# 06 · World Cup database → Knowledge Network → Agent Q&A

> Turn the public [Fjelstul World Cup Database](https://github.com/jfjelstul/worldcup) into a queryable KWeaver knowledge network, then ask an Agent analytic questions spanning matches, goals, and players.

[中文版](./README.zh.md)

## The problem

Historical World Cup facts live in many related tables — tournaments, matches, squads, goals, awards — not in a single report. Exploring “who scored the most career goals?” or “how did host nations perform?” usually means juggling joins and spreadsheets.

This example pulls **all 27 CSV datasets** into one knowledge network (`kweaver bkn create-from-csv`), samples a few object types (`matches`, `players`, `goals`), optionally runs semantic schema retrieval via Context Loader, and ends with **Agent chat** seeded with schema context and sampled rows.

## What this example does

```
CSV on disk (download.sh)
       │
       ▼
MySQL backing store  ◀── kweaver ds connect
       │
       ▼
Knowledge Network      ◀── kweaver bkn create-from-csv (27 tables → OTs + build)
       │
       ├── object-type query (samples)
       ├── context-loader kn-search (schema only, optional)
       └── agent chat (with schema + row snippets in the prompt)
```

0. **Download** 27 CSVs from `jfjelstul/worldcup` (not committed; `data/` is gitignored).
1. **Connect** a MySQL datasource the platform can reach.
2. **Import** every CSV and **build** one KN (`--table-prefix wc_`).
3. **List** object types and **query** sample instances for `matches`, `players`, `goals`.
4. **Search** schema text via Context Loader when the deployment supports it.
5. **Chat** with an existing Agent (or the first agent from `kweaver agent list`) using a factual prompt.

### Data source and license

The CSVs come from Joshua C. Fjelstul’s **The Fjelstul World Cup Database** ([repository](https://github.com/jfjelstul/worldcup)).

- **© 2023 Joshua C. Fjelstul, Ph.D.**
- Licensed under **CC-BY-SA 4.0** — [legal text](https://creativecommons.org/licenses/by-sa/4.0/legalcode)

You must keep attribution and the share-alike notice when redistributing derived data or documentation. This README satisfies the attribution requirement for our script + tutorial use.

**Pinning a revision:** set `WORLDCUP_REF` in `.env` (see `env.sample`) to a branch, tag, or commit SHA used in the raw GitHub URLs (for example `v1.2.0` if that tag exists). Default is `master`, which may move over time.

### The 27 datasets (grouped)

Aligned with the upstream project’s structure:

1. **Core entities** — `tournaments`, `confederations`, `teams`, `players`, `managers`, `referees`, `stadiums`, `matches`, `awards`
2. **Tournament mappings** — `qualified_teams`, `squads`, `manager_appointments`, `referee_appointments`
3. **Match appearances** — `team_appearances`, `player_appearances`, `manager_appearances`, `referee_appearances`
4. **In-match events** — `goals`, `penalty_kicks`, `bookings`, `substitutions`
5. **Standings / awards outcomes** — `host_countries`, `tournament_stages`, `groups`, `group_standings`, `tournament_standings`, `award_winners`

## Prerequisites

```bash
# 1. KWeaver CLI
npm install -g @kweaver-ai/kweaver-sdk

# 2. Authenticate (after install.md / quick-start on your deployment)
kweaver auth login https://<your-platform-url>

# 3. MySQL reachable from the platform — empty schema is OK; CSV import creates tables

# 4. curl (for download.sh)
```

Checked-in offline BKN with **Vega `resource`** placeholders: **[`worldcup-bkn-vega/`](./worldcup-bkn-vega)** (~**29** relation edges, semantic `rel_*` ids). Run **`kweaver bkn validate ./worldcup-bkn-vega`** before push (needs writable **`TMPDIR`** on some hosts — see [`worldcup-bkn-vega/README.zh.md`](./worldcup-bkn-vega/README.zh.md)).

(Optional CSV backup-only rsync lives in `./upload-data.sh`; it does **not** load data into MySQL.)

### Vega Catalog path (no Dataview)

Set **`VEGA_CATALOG_NAME`** (and optional **`VEGA_MYSQL_*`**) in `.env`, then run **`./run-branch-vega.sh`** (`catalog create` + **`discover --wait`**). Next: fill **`worldcup-bkn-vega/`** Resource placeholders or run **`kweaver bkn object-type create … --dataview-id <vega-resource-uuid>`** per table. Use the Node **`kweaver`** from **`@kweaver-ai/kweaver-sdk`** — ensure **`which kweaver`** does not resolve to a broken **`/usr/local/bin/kweaver`** stub. **`./run-branch-vega.sh --dry-run`** prints the plan only. Docs: **[WORKFLOW-BRANCH-VEGA.md](./WORKFLOW-BRANCH-VEGA.md)** / **[WORKFLOW-BRANCH-VEGA.zh.md](./WORKFLOW-BRANCH-VEGA.zh.md)**.

## Quick start

```bash
cd examples/06-world-cup
./download.sh
cp env.sample .env
vim .env   # DB_HOST / DB_NAME / DB_USER / DB_PASS (+ optional AGENT_ID, WORLDCUP_REF)
./run.sh
```

> **MySQL:** `create-from-csv` / `import-csv` defaults can hit Error 1118 on wide CSVs; `run.sh` slims `matches` / `team_appearances` before import by default (`SLIM_WIDE_CSV_FOR_MYSQL` in `.env`; see Troubleshooting tables in this README).

### Agent provisioning

The script **does not** create an Agent. Either:

- set `AGENT_ID` in `.env`, or  
- rely on `kweaver agent list` (first agent is used),

or import a template such as [`help/en/examples/sample-agent.import.json`](../../help/en/examples/sample-agent.import.json) in Studio / your admin flow, bind the KN after the tutorial run, configure an LLM, then re-run `./run.sh` with `AGENT_ID` set.

## Key commands

| Command | Role |
|---------|------|
| `./download.sh` | Fetch `data-csv/*.csv` from `jfjelstul/worldcup` via `raw.githubusercontent.com` |
| `kweaver ds connect mysql ...` | Register staging datasource |
| `kweaver bkn create-from-csv <ds> --files 'data/*.csv' --table-prefix wc_ --build` | Import + build one KN covering all CSVs |
| `kweaver bkn object-type list \| query` | Inspect OTs and row samples |
| `kweaver context-loader kn-search '...' --kn-id … --only-schema` | Optional schema-only semantic context |
| `kweaver agent chat <id> -m '...'` | Natural-language answer with facts in the prompt |

## Differences from Example 02 (CSV → KN)

| | 02-csv-to-kn | 06-world-cup |
|---|--------------|--------------|
| Data | Three small HR CSVs in repo | 27 upstream CSVs (downloaded), CC-BY-SA |
| Table prefix | Optional default | `wc_` to avoid clashes |
| Flow | Includes subgraph traversal + export | Trimmed to list/query + optional CL search + agent chat |
| Build timeout | 300s | 600s (many tables / rows) |

## Troubleshooting

| Symptom | What to try |
|---------|--------------|
| `curl` failures in `download.sh` | Check network; verify `WORLDCUP_REF` points at a revision that contains `data-csv/*.csv`. |
| `401` / oauth errors | `kweaver auth login` again; verify business domain (`kweaver config show`). |
| `create-from-csv` fails or times out | Increase `--timeout` in `run.sh` temporarily; or run two-step (`ds import-csv` + `bkn create-from-ds --build`). |
| Fewer than 27 object types after import | See platform logs / partial import errors; rerun with `--no-build`, fix tables, then `kweaver bkn build <kn_id>`. |
| `object-type query` returns `total = 0` | Wait for indexing; `kweaver bkn stats <kn_id>` / `kweaver bkn build <kn_id> --wait`. |
| Context Loader step skips | Deployment may omit Context Loader; script continues. |
| No Agent for chat step | Import or create an Agent; set `AGENT_ID`. |

## Cleanup

`run.sh` registers `trap cleanup EXIT`: the created **KN and datasource are deleted on exit**. Agents are **not** removed.
