# 04 · Skill Routing Loop — KN-driven Skill Governance

> [中文版](./README.zh.md)

> 3 materials trigger the same critical alert; the Decision Agent picks 3 different
> handling paths — each justified by the knowledge network.

## The Story

Continuing from example 03's procurement engineer: she now sees the disposition
plan already chosen on each alert. Three materials. Three paths. Zero prompts
edited. The `applicable_skill` relation in the business knowledge network is
the single source of truth — the agent picks from whatever `find_skills`
returns, nothing else.

## What this shows

Five components co-operate in a single end-to-end loop:

| Component | Role |
|---|---|
| **execution-factory** | registers and versions the 3 Skill packages |
| **business knowledge network (BKN)** | binds Skills to materials via `applicable_skill` |
| **Vega** | maps BKN ObjectTypes to MySQL tables (read-mostly) |
| **context-loader (`find_skills`)** | recalls applicable skills per material instance |
| **Decision Agent** | reads BKN evidence, picks a Skill, executes, audits |

## Prerequisites

- `kweaver` CLI ≥ 0.7.1 (`brew install kweaver-ai/tap/kweaver` or npm)
- KWeaver platform with **Decision Agent + execution-factory + Vega** enabled
  (use `kweaver auth login <platform-url> [--insecure]` first)
- A MySQL instance reachable from the KWeaver platform (NOT from your laptop)
  with CREATE/INSERT/SELECT/UPDATE on a chosen database
- `python3` (Flask + mysql-connector-python — install via
  `pip install -r tool_backend/requirements.txt`)
- An LLM model registered in the platform's model factory (find its ID via
  `kweaver call /api/mf-model-manager/v1/llm/list`)

## Quick Start

```bash
cd examples/05-skill-routing-loop
cp env.sample .env
vim .env                                    # fill PLATFORM_HOST, LLM_ID, DB_*
pip install -r tool_backend/requirements.txt
./run.sh                                    # ~5 minutes end-to-end
./run.sh --bonus                            # also run the Bonus segment
```

> **Concurrency caveat:** Do not run two instances of `./run.sh` concurrently.
> The script uses a fixed `KN_ID` (`ex05_skill_routing`); a second run's cleanup
> would delete the first run's KN.

## What you will see

| Material | KN evidence | DA picks | Why |
|---|---|---|---|
| MAT-001 | binds to `substitute_swap`; SUB-001A/B in stock | substitute_swap | Python scorer ranks substitutes; calls MES |
| MAT-002 | binds to `supplier_expedite`; SUP-2 capability=expedite | supplier_expedite | Supplier can rush — POST to supplier portal |
| MAT-003 | binds to `standard_replenish` only | standard_replenish | Default path — issue PO via ERP |

## Bonus — change business → KN rebuild → AI follows

Run `./run.sh --bonus`. The script POSTs to the mock business backend's admin
endpoint to re-bind MAT-002 from `supplier_expedite` to `standard_replenish`
(updates `materials.bound_skill_id` in MySQL, which drives the
`applicable_skill` direct-mapping FK), then triggers `kweaver bkn build` to
re-materialize the relation edges, then re-asks the Agent about MAT-002.
The Decision Agent's next `find_skills` call returns the new candidate set
and it switches to `standard_replenish` — without any prompt edit or redeploy.

> **Why the rebuild:** `applicable_skill` is a relation; its edges are
> materialized into the BKN graph at build time, not live-mapped. ObjectType
> data properties (e.g. `supplier.capability`) are read live from MySQL via
> Vega, but relation edges need a build to refresh. The rebuild step makes
> the loop explicit — business change → KN sync → AI sees new state.

## How it works (deeper read)

See [`docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md`](../../docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md)
for the full design including:
- BKN schema and the `applicable_skill` direct-mapping FK
- Why MCP server registration must include `X-Kn-ID` header
- Why agent `mode` must be `"react"` (default mode skips tool wiring)
- The 3-step state machine for cleaning up MCPs and Skills

## Troubleshooting

If you see `builtin_skill_load returned 404` in the chat trace, that's harmless.
The Decision Agent picks the correct skill from `find_skills`'s MCP metadata
and the `SKILL.md` description; the explicit `builtin_skill_load` lookup is
best-effort and the run still succeeds.

## Cleanup

Resources (KN, MCP, Skills, Agent, Datasource, mock backend process) are
deleted automatically on script exit, success or failure.
