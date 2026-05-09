# Build a knowledge network from CSV in one shot

> Start a new recipe by copying [`_TEMPLATE.md`](./_TEMPLATE.md). This page is the first concrete example showing how each section is filled in.

> - **Difficulty**: ⭐ Beginner
> - **Time**: ~ 10 minutes
> - **Modules touched**: `bkn`, `datasource`
> - **CLI version**: `kweaver >= 0.6`

## 1. Goal

**After this recipe you will have:** a knowledge network named `supply-kn` where each input CSV becomes one object type (OT), queryable via `bkn object-type query` and searchable via `bkn search` — all from a single command, with no hand-written schema.

## 2. Prerequisites

- Logged in via `kweaver auth login <platform-url>`.
- Correct business domain: `kweaver config show`; if it's wrong, run `kweaver config set-bd <uuid>`.
- A **datasource** that KWeaver can reach (the CSV files are imported into it first as the staging store).
- Your local CSV files (header on row 1, UTF-8). This recipe uses two files — `materials.csv` and `inventory.csv`, both with `material_code` and `material_name` columns.

## 3. Steps

### 3.1 Pick or create a datasource

List existing datasources first:

```bash
kweaver ds list
```

Connect a new one if none fits (MySQL example):

```bash
kweaver ds connect mysql db.example.com 3306 erp \
  --account root --password pass123
# → returns ds_id
```

> Record **`<ds_id>`** — the rest of the recipe assumes it's already known.

### 3.2 One-shot: build a KN from CSV

```bash
kweaver bkn create-from-csv <ds_id> \
  --files "materials.csv,inventory.csv" \
  --name "supply-kn" \
  --table-prefix sc_
# → Imports the CSVs, creates the dataview, the OTs, and runs the index build.
# → Returns kn_id.
```

Quick parameter reference:

| Parameter | Required | Description |
| --- | --- | --- |
| `<ds_id>` | yes | Datasource that stages the CSVs |
| `--files` | yes | Comma-separated paths or a glob (e.g. `"*.csv"`) |
| `--name` | yes | Knowledge network name |
| `--table-prefix` | no | Prefix for staging tables (avoids name clashes) |
| `--build` / `--no-build` | no | `--build` by default; pass `--no-build` to skip |
| `--timeout` | no | Build wait timeout in seconds (default 300) |

<details>
<summary>Equivalent two-step path (use this when you want to override primary/display keys)</summary>

```bash
kweaver ds import-csv <ds_id> --files "*.csv" --table-prefix sc_
kweaver bkn create-from-ds <ds_id> --name "supply-kn" --build
```

In the step-by-step path you can pass `--primary-key` / `--display-key` to `bkn object-type create` to pin the keys explicitly.

</details>

### 3.3 Verify

```bash
# List OTs — each CSV should yield one
kweaver bkn object-type list <kn_id>

# Sample query (always cap with limit to avoid wide-row JSON truncation)
kweaver bkn object-type query <kn_id> <ot_id> '{"limit":5}'

# Semantic search
kweaver bkn search <kn_id> "material"
```

## 4. Expected output

> **Success criterion**: `object-type query` returns `total > 0`, `datas[0]` contains the CSV columns you imported, and `bkn search` returns a non-empty `concepts` list.

`object-type query` should return something like:

```jsonc
{
  "total": 1234,
  "datas": [
    {
      "_instance_identity": "...",
      "material_code": "M-001",
      "material_name": "Screw",
      // ... other columns
    }
  ]
}
```

A non-empty `concepts` list from `bkn search` indicates the retrieval pipeline is healthy.

## 5. Troubleshooting

> The "Symptom" column lists **the literal output or error a reader will see**, so it can be search-matched directly.

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| `401 Unauthorized`, or response body contains `oauth info is not active` | Token expired | `kweaver auth login <platform-url>` |
| `kweaver bkn object-type list <kn_id>` prints `[]` | Wrong path or glob matched nothing | Check `--files`; use absolute paths if needed |
| `object-type query` response shows `total = 0` | Build incomplete or mapping mismatch | `kweaver bkn stats <kn_id>` to check `doc_count`; rebuild with `kweaver bkn build <kn_id> --wait --timeout 600` |
| `ds import-csv` reports `table already exists` | Staging table already there | First batch with `--recreate`: `kweaver ds import-csv <ds_id> --files "*.csv" --recreate` |
| Auto-detected primary key is not your business key | Heuristic could not infer it | Use the step-by-step path and pass `kweaver bkn object-type create ... --primary-key ... --display-key ...` |
| `bkn search` returns `HTTP 500` | The view does not support full-text search | Switch the query `condition` from `match` to `like` |

## 6. See also

- References: [BKN Engine](../manual/bkn.md) · [Data Source Management](../manual/datasource.md) · [Quick start](../quick-start.md)
- End-to-end sample project: [`examples/02-csv-to-kn/`](../../../examples/02-csv-to-kn/) in the repo
- Agent import template: [`../examples/sample-agent.import.json`](../examples/sample-agent.import.json) — bind it to your new KN once the build finishes
