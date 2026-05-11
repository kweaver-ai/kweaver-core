# Vega Catalog → Resources → BKN OTs (no Dataview)

Use this path when MySQL tables (e.g. `wc_*`) already exist and you want **Vega Catalog** + **discovered/manually created table Resources**, then attach each object type with:

`kweaver bkn object-type create <kn_id> --name … --dataview-id <vega_resource_id> --primary-key … --display-key …`

CLI flag **`--dataview-id`** is overloaded: here it must be the **Vega resource UUID**, not an mdl dataview UUID. See `kweaver vega --help` and Vega/BKN docs (`Resource` binds to OT `data_source.type = resource`).

Resource-backed object types typically query Vega live and often skip offline `bkn build` (deployment-dependent).

**Full step-by-step (Chinese):** [WORKFLOW-BRANCH-VEGA.zh.md](./WORKFLOW-BRANCH-VEGA.zh.md)

Checked-in offline BKN tree: **[`worldcup-bkn-vega/`](./worldcup-bkn-vega)** — each OT ends with **`resource | {{*_RES_ID}}`** placeholders; **`network.bkn`** uses **`worldcup_vega_catalog_bkn`**. Optionally fill UUIDs with [`scripts/render_worldcup_bkn_vega_placeholders.py`](./scripts/render_worldcup_bkn_vega_placeholders.py).

Optional helper: [`./run-branch-vega.sh`](./run-branch-vega.sh) (catalog create + `discover --wait` from `.env`; OT loop remains manual).
