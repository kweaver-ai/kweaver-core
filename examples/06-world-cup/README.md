# 06 · World Cup database → Vega Catalog → BKN → Agent Q&A

> Load the public [Fjelstul World Cup Database](https://github.com/jfjelstul/worldcup) into MySQL `wc_*` tables, then run a single script **`./run.sh`** that scans the source through **Vega**, pushes a checked-in **BKN** (`worldcup_vega_catalog_bkn`), builds search indexes, registers a Vega-SQL tool, and stands up a Decision Agent bound to it.

[中文版](./README.zh.md)

## The path

```
                       ┌─ 1) Download CSVs   fetch 27 CSVs from jfjelstul/worldcup (cached)
                       │
                       ├─ 2) Import MySQL    kweaver ds connect + ds import-csv → wc_* tables
                       │                     (pre-creates wc_matches / wc_team_appearances
                       │                      with VARCHAR(255) to dodge MySQL Error 1118)
                       │
                       ├─ 3) Vega scan       vega catalog create + discover --wait
                       │
   ./run.sh  ─────────►├─ 4) Render BKN      map vega Resources → render worldcup-bkn.tar
                       │
                       ├─ 5) Push BKN +      bkn validate + push (idempotent),
                       │   build indexes     then build vega resource OpenSearch datasets
                       │                     (7 entity tables get vector embedding)
                       │
                       ├─ 6) Upload toolbox  kweaver toolbox create + tool upload <OpenAPI>
                       │                     (registers `vega_sql_execute` so the agent
                       │                      can run raw SQL against MySQL)
                       │
                       └─ 7) Create Agent    agent create --config + bind KN + publish
                                             (reuses same-name agent on rerun)
```

Checked-in assets in this directory:
- **`worldcup-bkn.tar`** — offline BKN tree (27 object types, 29 `rel_*` edges) packaged as a tar archive; each OT ends with **`resource | {{*_RES_ID}}`** placeholders. `network.bkn` pins id `worldcup_vega_catalog_bkn`. `run.sh` extracts to `.tmp/worldcup-bkn/` before rendering.
- **`agent-worldcup.config.json`** — Agent template (Context Loader toolbox + system prompt). `run.sh` injects `data_source.knowledge_network[0].knowledge_network_id` and the `vega_sql_execute` tool/box ids at runtime.
- **`vega_sql_execute.openapi.json`** — OpenAPI 3.0 spec for the SQL-execute tool. Step 6 uploads it via `kweaver tool upload` (the OpenAPI parser path; sidesteps the 0.7.0 `kweaver toolbox import` bug that stored `api_spec` as null).

## Data source and license

CSVs come from Joshua C. Fjelstul’s **The Fjelstul World Cup Database** ([repo](https://github.com/jfjelstul/worldcup)).
- **© 2023 Joshua C. Fjelstul, Ph.D.**
- Licensed under **CC-BY-SA 4.0** — [legal text](https://creativecommons.org/licenses/by-sa/4.0/legalcode)

Keep attribution and the share-alike notice on derived data. **Pin a revision** via `WORLDCUP_REF` in `.env` (default `master`, which may move).

## First-time setup checklist

`run.sh` only automates the 7 steps above. On a fresh machine + fresh cluster you still need these one-shot platform tasks:

1. **Install the KWeaver Core platform** (k8s + bkn-backend + ontology-query + vega-backend + agent-* + mf-model-* + opensearch + minio + mariadb). Use `deploy/onboard.sh` from the repo root — see [deploy/README.md](../../deploy/README.md). Recommended `0.8.0+` (fixes the `_score` resource-path bug and the toolbox-import write bug).
2. **Authenticate the CLI**: `kweaver auth login https://<your-platform-url>` (writes `~/.kweaver/`).
3. **Register an LLM model** (agent chat needs it):
   ```bash
   kweaver model llm add --body-file <llm.json>   # see `kweaver model llm --template`
   ```
   Note the returned `model_id` and put it in `.env` as `AGENT_LLM_ID`.
4. **Register an embedding model** (vector index needs it; keyword-only build works without it):
   ```bash
   kweaver model small add --name text-embedding-v4-cn \
     --type embedding --batch-size 10 --max-tokens 512 --embedding-dim 1024 \
     --model-config-file <emb.json>
   ```
   `EMBEDDING_MODEL_NAME` in `.env` is resolved to `model_id` at runtime (default `text-embedding-v4-cn`).
5. **Wire BKN to the embedding model** (only needed for KN-level semantic search; this example doesn't rely on it):
   ```bash
   sudo bash deploy/onboard.sh --enable-bkn-search \
     --bkn-embedding-name=text-embedding-v4-cn
   ```
6. **Stage MySQL** — create the `worldcup` database and a user that is reachable from the kweaver platform (k8s pod network). Fill `DB_HOST / DB_PORT / DB_NAME / DB_USER / DB_PASS` in `.env`.

Once those are done, `./run.sh` is fully turnkey and idempotent.

## Prerequisites (per-shell)

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://<your-platform-url>
# Use the Node SDK `kweaver` (avoid a broken /usr/local/bin/kweaver stub).
# MySQL must be reachable from the platform AND from Vega connectors.
# curl + jq + python3 + (optional) `mysql` CLI for the wide-table pre-create
```

## Quick start

```bash
cd examples/06-world-cup
cp env.sample .env
vim .env   # at minimum: DB_* and AGENT_LLM_ID

# Single command runs all 7 steps end-to-end; every step is idempotent on rerun.
./run.sh
```

`./run.sh --help` lists every flag. Common variants:

| Command | Effect |
|---------|--------|
| `./run.sh` | Run steps 1→7 |
| `./run.sh --dry-run` | Plan only, no API calls |
| `./run.sh --from 3` | Rerun from Vega scan onward (CSVs already in MySQL) |
| `./run.sh --only 5` | Only run step 5 (BKN push + index build) |
| `./run.sh --only 7` | Only run agent create/update |
| `./run.sh --no-publish` | Keep Agent private (skip publish) |
| `./run.sh --no-reuse` | Always create a new Agent, ignore same-name reuse |

## The agent's 3 tools

After step 7 the agent has three complementary read tools — the LLM picks per question:

| Tool | Source toolbox | Use when |
|---|---|---|
| **`search_schema`** | platform built-in (`contextloader工具集_070`) | Explore object types / relations / look up `data_source.id` of a wc_* resource |
| **`query_object_instance`** | platform built-in | Equality / `in` / range condition on one OT, e.g. `family_name='Sun' AND given_name='Wen'`. Reads from the OpenSearch dataset built in step 5 — fast for value lookups. |
| **`vega_sql_execute`** | uploaded by step 6 from `vega_sql_execute.openapi.json` | Raw MySQL SQL — needed for `ORDER BY`, `GROUP BY`, multi-table `JOIN`, `COUNT(*)`. Reference resources by `{{<resource_id>}}` placeholder; `resource_id` comes from `search_schema`. |

Typical LLM workflow: `search_schema` → pick OT + grab `data_source.id` → `query_object_instance` for filtered lookups OR `vega_sql_execute` for aggregations.

## Example questions

Once `./run.sh` finishes, chat with the agent:

```bash
kweaver agent chat <AGENT_ID> -m '<your question>' --stream
```

`<AGENT_ID>` is written back to `.env` after step 7.

### Q1 · Messi's World Cup awards
**Prompt**: `梅西获得过哪些奖项？`

**Agent path**: `search_schema(award)` → `query_object_instance(award_winners, family_name='Messi' AND given_name='Lionel')` → returns:
- 2014 FIFA Men's World Cup — **Golden Ball**
- 2022 FIFA Men's World Cup — **Golden Ball** + **Silver Boot**

### Q2 · Sun Wen's per-tournament goals + all-time rank
**Prompt**: `孙雯每届女足世界杯进球数？历史射手榜排第几？`

**Agent path**: `vega_sql_execute(SELECT tournament_name, COUNT(*) FROM goals WHERE family_name='Sun' AND given_name='Wen' GROUP BY tournament_name)` + a second SQL for the ranking → returns:

| Tournament | Goals |
|---|---|
| 1991 Women's WC | 1 |
| 1995 Women's WC | 2 |
| **1999 Women's WC** | **7** (also won Golden Ball + Golden Boot) |
| 2003 Women's WC | 1 |
| **Total** | **11** (tied 5th on all-time women's WC scorers) |

### Q3 · Last three men's World Cup winners
**Prompt**: `近三届男足世界杯冠军？`

**Agent path**: `vega_sql_execute(SELECT year, host_country, winner FROM tournaments WHERE tournament_name LIKE '%Men%' ORDER BY CAST(year AS UNSIGNED) DESC LIMIT 3)` → returns:
- 2022 Qatar → Argentina
- 2018 Russia → France
- 2014 Brazil → Germany

Every number comes from the 27 `wc_*` tables you imported in step 2 — zero LLM hallucination.

## The 27 datasets (grouped)

1. **Core entities** — `tournaments`, `confederations`, `teams`, `players`, `managers`, `referees`, `stadiums`, `matches`, `awards`
2. **Tournament mappings** — `qualified_teams`, `squads`, `manager_appointments`, `referee_appointments`
3. **Match appearances** — `team_appearances`, `player_appearances`, `manager_appearances`, `referee_appearances`
4. **In-match events** — `goals`, `penalty_kicks`, `bookings`, `substitutions`
5. **Standings / awards** — `host_countries`, `tournament_stages`, `groups`, `group_standings`, `tournament_standings`, `award_winners`

## Troubleshooting

| Symptom | What to try |
|---------|--------------|
| Step 1 download fails | Check network; verify `WORLDCUP_REF` points at a revision with `data-csv/`. |
| `kweaver auth` 401 | `kweaver auth login`; confirm business domain via `kweaver config show`. |
| `import-csv` → MySQL **Error 1118** | Step 2 pre-creates `wc_matches` / `wc_team_appearances` with VARCHAR(255) via the local `mysql` CLI. Without that client installed you must pre-create them manually or relax column types. |
| Vega `discover` fails | Set `VEGA_CATALOG_ID` then `./run.sh --from 4`. |
| Fewer than 27 Resources | `databases` in connector config incomplete, or discover didn't finish — adjust `VEGA_MYSQL_DATABASES` and rerun step 3. |
| Step 5 index build skipped for a large table | Expected on platform 0.7.0 — `vega-backend` has a batch-sync cursor bug on tables `>2× batch_size` (8 event tables). Agent falls back to `vega_sql_execute` for those. Fixed in 0.8.0+. |
| Step 5 `embedding model … not registered` | Either register one (see setup checklist) or set `DO_INDEX=0` / `EMBEDDING_MODEL_NAME=` (empty) — script then builds keyword-only indexes. |
| `agent create` fails on missing LLM | Set `AGENT_LLM_ID` in `.env` to a `model_id` from `kweaver model llm list`. |

## Differences from Example 02

| | 02-csv-to-kn | 06-world-cup |
|---|--------------|--------------|
| Data | Three small HR CSVs in repo | 27 upstream CSVs (downloaded), CC-BY-SA |
| Knowledge path | `create-from-csv` | **MySQL + Vega Resource** + checked-in **`worldcup-bkn.tar`** push |
| Tooling | none | OpenAPI-uploaded `vega_sql_execute` so the agent can run raw SQL |
| Entry point | Multiple scripts | Single `./run.sh` (steps 1–7, separately runnable, idempotent) |

## Cleanup

`./run.sh` does **not** auto-delete the datasource, MySQL tables, Vega catalog, KN, toolbox, or Agent. Remove them explicitly in Studio / CLI when no longer needed.
