# 06 · World Cup database → Vega Catalog → BKN → Agent Q&A

> Load the public [Fjelstul World Cup Database](https://github.com/jfjelstul/worldcup) into MySQL `wc_*` tables, then run a single script **`./run.sh`** that scans the source through **Vega**, pushes a checked-in **BKN** (`worldcup_vega_catalog_bkn`), and stands up a Decision Agent bound to it.

[中文版](./README.zh.md)

## The path

```
                       ┌─ 1) Download CSVs   fetch 27 CSVs from jfjelstul/worldcup (cached)
                       │
                       ├─ 2) Import MySQL    kweaver ds connect + ds import-csv → wc_* tables
                       │
                       ├─ 3) Vega scan       vega catalog create + discover --wait
                       │
   ./run.sh  ─────────►├─ 4) Render BKN      map Resources → render worldcup-bkn
                       │
                       ├─ 5) Push BKN        bkn validate + push
                       │
                       ├─ 6) Upload toolbox  kweaver toolbox import + publish (idempotent by box_name)
                       │
                       └─ 7) Create Agent    agent create --config + bind KN + publish
```

Checked-in assets in this directory:
- **`worldcup-bkn.tar`** — offline BKN tree (27 object types, 29 `rel_*` edges) packaged as a tar archive; each OT ends with **`resource | {{*_RES_ID}}`** placeholders. `network.bkn` pins id `worldcup_vega_catalog_bkn`. `run.sh` extracts to `.tmp/worldcup-bkn/` before rendering.
- **`agent-worldcup.config.json`** — Agent template (Context Loader toolbox + system prompt). `run.sh` injects `data_source.knowledge_network[0].knowledge_network_id` at runtime.
- **`bkn_schema_and_query_toolbox.adp`** — Toolbox ADP (BKN schema reads + ontology instance/action query). Wired for the 43.131.23.35 k8s deployment via in-cluster ClusterIP services `bkn-backend-svc:13014` and `ontology-query-svc:13018` against the `/in/v1` internal routes; adjust `server_url` / paths and account headers if your cluster differs.

## Data source and license

CSVs come from Joshua C. Fjelstul’s **The Fjelstul World Cup Database** ([repo](https://github.com/jfjelstul/worldcup)).
- **© 2023 Joshua C. Fjelstul, Ph.D.**
- Licensed under **CC-BY-SA 4.0** — [legal text](https://creativecommons.org/licenses/by-sa/4.0/legalcode)

Keep attribution and the share-alike notice on derived data. **Pin a revision** via `WORLDCUP_REF` in `.env` (default `master`, which may move).

## Prerequisites

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://<your-platform-url>
# Use the Node SDK `kweaver` (avoid a broken /usr/local/bin/kweaver stub).
# MySQL must be reachable from the platform AND from Vega connectors.
# curl + jq + python3
```

## Quick start

```bash
cd examples/06-world-cup
cp env.sample .env
vim .env   # at minimum: DB_* and VEGA_CATALOG_NAME

# Single command runs all 7 steps end-to-end (step 6 reuses an existing toolbox if box_name matches)
./run.sh
```

`./run.sh --help` lists every flag. Common variants:

| Command | Effect |
|---------|--------|
| `./run.sh` | Run steps 1→7 |
| `./run.sh --dry-run` | Plan only, no API calls |
| `./run.sh --from 3` | Rerun from Vega scan onward (CSVs already in MySQL) |
| `./run.sh --only 7` | Only create the Agent |
| `./run.sh --no-publish` | Keep Agent private (skip publish) |

## The 27 datasets (grouped)

1. **Core entities** — `tournaments`, `confederations`, `teams`, `players`, `managers`, `referees`, `stadiums`, `matches`, `awards`
2. **Tournament mappings** — `qualified_teams`, `squads`, `manager_appointments`, `referee_appointments`
3. **Match appearances** — `team_appearances`, `player_appearances`, `manager_appearances`, `referee_appearances`
4. **In-match events** — `goals`, `penalty_kicks`, `bookings`, `substitutions`
5. **Standings / awards** — `host_countries`, `tournament_stages`, `groups`, `group_standings`, `tournament_standings`, `award_winners`

## Troubleshooting

| Symptom | What to try |
|---------|--------------|
| `download.sh` fails | Check network; verify `WORLDCUP_REF` points at a revision with `data-csv/`. |
| `kweaver auth` 401 | `kweaver auth login`; confirm business domain via `kweaver config show`. |
| `import-csv` → MySQL **Error 1118** | Step 2 pre-creates `wc_matches` / `wc_team_appearances` with VARCHAR(255) via the `mysql` CLI. Without that client installed you must pre-create them manually or relax column types. |
| Vega `discover` fails | See [`report/troubleshooting.md`](./report/troubleshooting.md); set `VEGA_CATALOG_ID` then `./run.sh --from 2`. |
| Fewer than 27 Resources | `databases` in connector config incomplete, or discover did not finish — adjust `VEGA_MYSQL_DATABASES` and rerun step 1. |
| `bkn build` returns `NoneConceptType` | Expected for all-resource KN; keep `DO_BKN_BUILD=0` and query via Vega / Context Loader tools. |
| `agent create` fails on missing LLM | Set `AGENT_LLM_ID` in `.env` to a `model_id` from `kweaver model llm list`. |

## Differences from Example 02

| | 02-csv-to-kn | 06-world-cup |
|---|--------------|--------------|
| Data | Three small HR CSVs in repo | 27 upstream CSVs (downloaded), CC-BY-SA |
| Knowledge path | `create-from-csv` | **MySQL + Vega Resource** + checked-in **`worldcup-bkn`** push |
| Entry point | Multiple scripts | Single `./run.sh` (steps 1–7, separately runnable) |

## Cleanup

`./run.sh` does **not** auto-delete the datasource, MySQL tables, Vega catalog, KN, or Agent. Remove them explicitly in Studio / CLI when no longer needed.
