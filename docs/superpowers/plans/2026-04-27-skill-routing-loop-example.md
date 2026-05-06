# Example 04 · Skill Routing Loop · Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `examples/04-skill-routing-loop/` — a one-command end-to-end demo of KWeaver Skill governance: business truth (MySQL) → BKN (via Vega) → context-loader `find_skills` → Decision Agent → Action, with 3 materials triggering 3 distinct decision paths and a Bonus showing "change business → AI follows."

**Architecture:** Pure CLI (`kweaver` SDK + `kweaver call` for raw HTTP). BKN deployed via `bkn push` (GitOps style). Skills packaged as zips with `SKILL.md` at root, registered to execution-factory. Agent created with `mode=react` to enable MCP tool calling. MCP server registered with `X-Kn-ID` header to bridge missing kn_id propagation.

**Tech Stack:** Bash 5+, Python 3 (stdlib + `mysql-connector-python`), `kweaver` CLI 0.7.1+, MySQL 5.7+ (user-provided, reachable from KWeaver platform).

**Spec:** [`docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md`](../specs/2026-04-27-skill-routing-loop-example-design.md)
**Issue:** [kweaver-ai/kweaver-core#312](https://github.com/kweaver-ai/kweaver-core/issues/312)

---

## File Structure (Locked)

```
examples/04-skill-routing-loop/
├── README.md                       # English narrative + quickstart
├── README.zh.md                    # 中文版
├── env.sample                      # DB + platform host + LLM ID
├── run.sh                          # ~250 lines, end-to-end orchestration
├── data/
│   ├── materials.csv               # 5 rows (3 critical + 2 substitute)
│   ├── suppliers.csv               # 3 rows (1 normal + 2 expedite)
│   └── skills.csv                  # 3 rows (one per skill package)
├── skills/
│   ├── standard_replenish/SKILL.md
│   ├── substitute_swap/
│   │   ├── SKILL.md
│   │   └── pick_substitute.py     # Python multi-criteria scoring
│   └── supplier_expedite/SKILL.md
├── tool_backend/
│   ├── server.py                   # mock business system: HTTP server + DB writer
│   └── requirements.txt            # flask + mysql-connector-python
├── bkn/
│   ├── network.bkn
│   ├── object_types/
│   │   ├── material.bkn
│   │   ├── supplier.bkn
│   │   └── skills.bkn
│   └── relation_types/
│       ├── applicable_skill.bkn
│       └── supplied_by.bkn
└── agent.json                      # Template — placeholders filled by run.sh
```

**File responsibilities** (one purpose each):
- `data/*.csv` — business truth
- `bkn/**/*.bkn` — semantic schema (Vega view bindings filled at runtime)
- `skills/<name>/` — Skill package source (1 `SKILL.md` + optional `.py`)
- `tool_backend/server.py` — mock ERP/MES + admin endpoint for Bonus
- `agent.json` — agent config template, sed-substituted in run.sh
- `run.sh` — orchestration only, no business logic
- `README*` — story + quickstart + Bonus instructions

---

## Test Strategy

This is **example code**, not library code. Tests = end-to-end runs against a live KWeaver platform (target a self-hosted instance with reachable MySQL and a registered LLM model; concrete host/IP/model_id come from the developer's local `.env`).

**Per-task verification**: each task ends with a concrete command that exercises the slice and prints visible success/failure. No unit tests.

**Final acceptance** (Task 13): a fresh `./run.sh` from a clean 62 env completes within 5 minutes with all 3 trace cases visible and Bonus working.

---

## Pre-flight Setup (one-time, not per-task)

Worktree (recommended for isolation):
```bash
cd /Users/xupeng/dev/github/kweaver
git worktree add .worktrees/example-04 -b feat/example-04
cd .worktrees/example-04
```

If skipping worktree, work directly in main and create branch:
```bash
git checkout -b feat/example-04
```

All paths below are relative to `examples/04-skill-routing-loop/` unless absolute.

---

## Task 1 · Bootstrap example dir + env.sample

**Files:**
- Create: `examples/04-skill-routing-loop/env.sample`
- Create: `examples/04-skill-routing-loop/.gitignore`

- [ ] **Step 1: Create example directory**

```bash
mkdir -p examples/04-skill-routing-loop/{data,skills/standard_replenish,skills/substitute_swap,skills/supplier_expedite,tool_backend,bkn/object_types,bkn/relation_types}
```

- [ ] **Step 2: Write `env.sample`**

```bash
cat > examples/04-skill-routing-loop/env.sample <<'EOF'
# ── Platform ─────────────────────────────────────────────────────────────────
# KWeaver platform host (without trailing slash). Use https for self-signed
# (the CLI is configured with --insecure via `kweaver auth login`).
# example: https://platform.example.com
PLATFORM_HOST=https://your-platform-host

# LLM model ID from kweaver model factory.
# Find via: kweaver call /api/mf-model-manager/v1/llm/list?page=1&size=20
LLM_ID=
LLM_NAME=deepseek-v3.2

# ── MySQL business database ──────────────────────────────────────────────────
# MUST be reachable from the KWeaver platform (not from your laptop).
# Required permissions: CREATE TABLE / INSERT / SELECT / UPDATE on DB_NAME.
DB_HOST=10.0.0.1
DB_PORT=3306
DB_NAME=supply_chain
DB_USER=root
DB_PASS=changeme

# ── Mock business backend ────────────────────────────────────────────────────
# Local port for the mock ERP/MES that Skills call. Change if 8765 is in use.
TOOL_BACKEND_PORT=8765

# ── Optional ─────────────────────────────────────────────────────────────────
# Set to 1 for verbose logging.
# DEBUG=1
EOF
```

- [ ] **Step 3: Add .gitignore**

```bash
cat > examples/04-skill-routing-loop/.gitignore <<'EOF'
.env
*.log
*.zip
__pycache__/
EOF
```

- [ ] **Step 4: Verify**

```bash
ls examples/04-skill-routing-loop/
# expect: env.sample .gitignore data/ skills/ tool_backend/ bkn/
```

- [ ] **Step 5: Commit**

```bash
git add examples/04-skill-routing-loop/
git commit -m "feat(example-04): bootstrap directory layout and env.sample"
```

---

## Task 2 · Data CSVs

**Files:**
- Create: `examples/04-skill-routing-loop/data/materials.csv`
- Create: `examples/04-skill-routing-loop/data/suppliers.csv`
- Create: `examples/04-skill-routing-loop/data/skills.csv`

- [ ] **Step 1: Write `materials.csv`**

```bash
cat > examples/04-skill-routing-loop/data/materials.csv <<'EOF'
sku,name,current_stock,safety_stock,material_risk,supplier_id,bound_skill_id
MAT-001,Battery Cell,40,100,critical,SUP-1,substitute_swap
MAT-002,Power Module,30,120,critical,SUP-2,supplier_expedite
MAT-003,Connector,15,80,critical,SUP-3,standard_replenish
SUB-001A,Battery Cell Substitute,200,50,normal,SUP-1,
SUB-001B,Battery Cell Alt,80,40,normal,SUP-1,
EOF
```

- [ ] **Step 2: Write `suppliers.csv`**

```bash
cat > examples/04-skill-routing-loop/data/suppliers.csv <<'EOF'
supplier_id,name,capability
SUP-1,Acme Corp,normal
SUP-2,Bolt Industries,expedite
SUP-3,Cell Source,normal
EOF
```

- [ ] **Step 3: Write `skills.csv`** (skill_id values must match Skill package names registered in Task 5)

```bash
cat > examples/04-skill-routing-loop/data/skills.csv <<'EOF'
skill_id,name,description
standard_replenish,标准补货,Default procurement order via ERP
substitute_swap,替代料切换,Pick best substitute via Python scoring then call MES
supplier_expedite,供应商加急,Send expedite request to supplier portal
EOF
```

- [ ] **Step 4: Verify content**

```bash
wc -l examples/04-skill-routing-loop/data/*.csv
# expect: 6 materials.csv (5 data + 1 header), 4 suppliers.csv, 4 skills.csv
```

- [ ] **Step 5: Commit**

```bash
git add examples/04-skill-routing-loop/data/
git commit -m "feat(example-04): seed CSV data — 5 materials, 3 suppliers, 3 skills"
```

---

## Task 3 · BKN schema directory (3 OT + 2 RT)

**Files:**
- Create: `examples/04-skill-routing-loop/bkn/network.bkn`
- Create: `examples/04-skill-routing-loop/bkn/object_types/material.bkn`
- Create: `examples/04-skill-routing-loop/bkn/object_types/supplier.bkn`
- Create: `examples/04-skill-routing-loop/bkn/object_types/skills.bkn`
- Create: `examples/04-skill-routing-loop/bkn/relation_types/applicable_skill.bkn`
- Create: `examples/04-skill-routing-loop/bkn/relation_types/supplied_by.bkn`

> All `.bkn` files use placeholder `{{...}}` for `data_view` IDs. `run.sh` uses `sed` to substitute at runtime after dataviews exist.

- [ ] **Step 1: `network.bkn`**

```bash
cat > examples/04-skill-routing-loop/bkn/network.bkn <<'EOF'
---
type: knowledge_network
id: ex04_skill_routing
name: Example 04 - Skill Routing
tags: [example, skill-routing]
business_domain: bd_public
---

# Example 04 - Skill Routing

KN-driven Skill governance: Material → applicable_skill → Skills, plus
Material → supplied_by → Supplier as DA reasoning evidence.

## Network Overview

- ObjectTypes (object_types/): material, supplier, skills
- RelationTypes (relation_types/): applicable_skill, supplied_by
EOF
```

- [ ] **Step 2: `object_types/material.bkn`**

```bash
cat > examples/04-skill-routing-loop/bkn/object_types/material.bkn <<'EOF'
---
type: object_type
id: material
name: Material
tags: [example-04]
---

## ObjectType: Material

Inventory item. The `bound_skill_id` FK points at the applicable Skill.

### Data Properties

| Name | Display Name | Type | Description | Mapped Field |
|------|--------------|------|-------------|--------------|
| sku | SKU | string | Material code (PK) | sku |
| name | Name | string | Display name | name |
| current_stock | Current Stock | string | Live inventory | current_stock |
| safety_stock | Safety Stock | string | Reorder threshold | safety_stock |
| material_risk | Risk | string | Risk tier (critical/normal) | material_risk |
| supplier_id | Supplier | string | FK to supplier | supplier_id |
| bound_skill_id | Bound Skill | string | FK to skills (applicable_skill) | bound_skill_id |

### Keys

Primary Keys: sku
Display Key: name
Incremental Key:

### Logic Properties

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {{MATERIALS_DV_ID}} | {{MATERIALS_DV_NAME}} |
EOF
```

- [ ] **Step 3: `object_types/supplier.bkn`**

```bash
cat > examples/04-skill-routing-loop/bkn/object_types/supplier.bkn <<'EOF'
---
type: object_type
id: supplier
name: Supplier
tags: [example-04]
---

## ObjectType: Supplier

Material supplier with delivery capability.

### Data Properties

| Name | Display Name | Type | Description | Mapped Field |
|------|--------------|------|-------------|--------------|
| supplier_id | Supplier ID | string | PK | supplier_id |
| name | Name | string | Display name | name |
| capability | Capability | string | normal / expedite | capability |

### Keys

Primary Keys: supplier_id
Display Key: name
Incremental Key:

### Logic Properties

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {{SUPPLIERS_DV_ID}} | {{SUPPLIERS_DV_NAME}} |
EOF
```

- [ ] **Step 4: `object_types/skills.bkn`** (id MUST be `skills` literally)

```bash
cat > examples/04-skill-routing-loop/bkn/object_types/skills.bkn <<'EOF'
---
type: object_type
id: skills
name: Skills
tags: [example-04, skills]
---

## ObjectType: Skills

Platform-required ObjectType id is `skills` (find_skills SkillsObjectTypeID).
skill_id values reference packages registered in execution-factory.

### Data Properties

| Name | Display Name | Type | Description | Mapped Field |
|------|--------------|------|-------------|--------------|
| skill_id | Skill ID | string | execution-factory skill package id (PK) | skill_id |
| name | Name | string | Human-readable name | name |
| description | Description | string | Usage hint | description |

### Keys

Primary Keys: skill_id
Display Key: name
Incremental Key:

### Logic Properties

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {{SKILLS_DV_ID}} | {{SKILLS_DV_NAME}} |
EOF
```

- [ ] **Step 5: `relation_types/applicable_skill.bkn`** (Type=direct, NOT data_view — see spec P1-2)

```bash
cat > examples/04-skill-routing-loop/bkn/relation_types/applicable_skill.bkn <<'EOF'
---
type: relation_type
id: applicable_skill
name: Applicable skill
tags: [example-04]
---

## RelationType: Applicable skill

Material → Skills, the find_skills recall edge.

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| material | skills | direct |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| bound_skill_id | skill_id |
EOF
```

- [ ] **Step 6: `relation_types/supplied_by.bkn`**

```bash
cat > examples/04-skill-routing-loop/bkn/relation_types/supplied_by.bkn <<'EOF'
---
type: relation_type
id: supplied_by
name: Supplied by
tags: [example-04]
---

## RelationType: Supplied by

Material → Supplier, used by DA to read supplier capability as evidence.

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| material | supplier | direct |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| supplier_id | supplier_id |
EOF
```

- [ ] **Step 7: Verify schema syntax** (validate without push, placeholders are fine)

```bash
kweaver bkn validate examples/04-skill-routing-loop/bkn 2>&1 | tail -3
# expect: "Valid: 3 object types, 2 relation types, 0 action types"
```

- [ ] **Step 8: Commit**

```bash
git add examples/04-skill-routing-loop/bkn/
git commit -m "feat(example-04): BKN schema — 3 ObjectType + 2 RelationType (skills id literal)"
```

---

## Task 4 · Mock business backend

**Files:**
- Create: `examples/04-skill-routing-loop/tool_backend/server.py`
- Create: `examples/04-skill-routing-loop/tool_backend/requirements.txt`

> Mock backend serves 4 endpoints: 3 business endpoints (called by Skills) and 1 admin endpoint (called by Bonus to update business DB without requiring user to install mysql client).

- [ ] **Step 1: `requirements.txt`**

```bash
cat > examples/04-skill-routing-loop/tool_backend/requirements.txt <<'EOF'
flask>=3.0
mysql-connector-python>=8.0
EOF
```

- [ ] **Step 2: `server.py`**

```bash
cat > examples/04-skill-routing-loop/tool_backend/server.py <<'PYEOF'
#!/usr/bin/env python3
"""
Mock business backend for example 04.

Three business endpoints (called by Skills):
  POST /procurement/order          - standard_replenish
  POST /mes/swap                   - substitute_swap
  POST /supplier/expedite          - supplier_expedite

One admin endpoint (called by Bonus):
  POST /admin/supplier-capability  - simulate "業務系統 update SUP-X.capability"
"""
import json
import os
import re
import sys

import mysql.connector
from flask import Flask, jsonify, request

app = Flask(__name__)
PORT = int(os.environ.get("TOOL_BACKEND_PORT", "8765"))
DB_CONFIG = {
    "host": os.environ["DB_HOST"],
    "port": int(os.environ.get("DB_PORT", "3306")),
    "database": os.environ["DB_NAME"],
    "user": os.environ["DB_USER"],
    "password": os.environ["DB_PASS"],
}

# run.sh imports CSVs with --table-prefix, so the real table is ex04_<ts>_suppliers.
SUPPLIERS_TABLE = os.environ.get("SUPPLIERS_TABLE", "suppliers")
if not re.fullmatch(r"[A-Za-z0-9_]+", SUPPLIERS_TABLE):
    raise ValueError(f"Invalid SUPPLIERS_TABLE: {SUPPLIERS_TABLE!r}")

# ── Business endpoints ───────────────────────────────────────────────────────

@app.post("/procurement/order")
def procurement_order():
    body = request.get_json(force=True)
    sku = body.get("sku", "?")
    qty = body.get("qty", 0)
    po = f"PO-{sku}-MOCK"
    print(f"[procurement] sku={sku} qty={qty} -> {po}", file=sys.stderr)
    return jsonify({"po_number": po, "status": "submitted"})


@app.post("/mes/swap")
def mes_swap():
    body = request.get_json(force=True)
    print(f"[mes/swap] {body}", file=sys.stderr)
    return jsonify({"status": "swap_acknowledged", "ticket": "MES-MOCK-001"})


@app.post("/supplier/expedite")
def supplier_expedite():
    body = request.get_json(force=True)
    print(f"[supplier/expedite] {body}", file=sys.stderr)
    return jsonify({"status": "expedite_requested", "sla_hours": 36})


# ── Admin endpoint (Bonus) ───────────────────────────────────────────────────

@app.post("/admin/supplier-capability")
def admin_set_capability():
    """Simulate business-system update; writes directly to MySQL.

    body: { "supplier_id": "SUP-2", "capability": "normal" }
    """
    body = request.get_json(force=True)
    supplier_id = body.get("supplier_id")
    capability = body.get("capability")
    if not supplier_id or capability not in ("normal", "expedite"):
        return jsonify({"error": "invalid request"}), 400
    conn = mysql.connector.connect(**DB_CONFIG)
    try:
        cur = conn.cursor()
        cur.execute(
            f"UPDATE {SUPPLIERS_TABLE} SET capability=%s WHERE supplier_id=%s",
            (capability, supplier_id),
        )
        affected = cur.rowcount
        conn.commit()
    finally:
        conn.close()
    print(f"[admin] supplier {supplier_id} capability -> {capability} ({affected} rows)", file=sys.stderr)
    return jsonify({"updated": affected, "supplier_id": supplier_id, "capability": capability})


@app.get("/healthz")
def healthz():
    return jsonify({"status": "ok"})


if __name__ == "__main__":
    print(f"[tool_backend] listening on :{PORT}", file=sys.stderr)
    app.run(host="0.0.0.0", port=PORT, debug=False)
PYEOF
chmod +x examples/04-skill-routing-loop/tool_backend/server.py
```

- [ ] **Step 3: Smoke-test the server starts** (uses your .env if you've made one; else expects DB env vars)

```bash
cd examples/04-skill-routing-loop
DB_HOST=db.example.com DB_PORT=3306 DB_NAME=supply_chain DB_USER=root \
  DB_PASS='<your-db-password>' TOOL_BACKEND_PORT=8765 \
  python3 tool_backend/server.py &
sleep 2
curl -s http://localhost:8765/healthz
# expect: {"status":"ok"}
kill %1 2>/dev/null
cd -
```

- [ ] **Step 4: Commit**

```bash
git add examples/04-skill-routing-loop/tool_backend/
git commit -m "feat(example-04): mock business backend with 3 endpoints + admin override"
```

---

## Task 5 · Skill packages (3 SKILL.md, 1 with Python)

**Files:**
- Create: `examples/04-skill-routing-loop/skills/standard_replenish/SKILL.md`
- Create: `examples/04-skill-routing-loop/skills/substitute_swap/SKILL.md`
- Create: `examples/04-skill-routing-loop/skills/substitute_swap/pick_substitute.py`
- Create: `examples/04-skill-routing-loop/skills/supplier_expedite/SKILL.md`

- [ ] **Step 1: `standard_replenish/SKILL.md`**

```bash
cat > examples/04-skill-routing-loop/skills/standard_replenish/SKILL.md <<'EOF'
---
name: standard_replenish
description: Default procurement order via ERP. Use when no substitute is available and supplier cannot expedite.
version: 0.1.0
---

# Standard Replenish

## When to use

Material is at critical stock level AND no faster path is feasible (no
in-stock substitute, supplier capability is `normal`).

## How to invoke

POST `${TOOL_BACKEND_URL}/procurement/order` with body `{"sku": "<sku>", "qty": <integer>}`.

Returns `{po_number, status}`. Surface the PO number in the final answer.
EOF
```

- [ ] **Step 2: `supplier_expedite/SKILL.md`**

```bash
cat > examples/04-skill-routing-loop/skills/supplier_expedite/SKILL.md <<'EOF'
---
name: supplier_expedite
description: Send expedite request to supplier portal. Use only when the material's supplier has capability=expedite.
version: 0.1.0
---

# Supplier Expedite

## When to use

Material's supplier (via `supplied_by` relation) has `capability == "expedite"`.
Verify by reading supplier node in BKN before invoking.

## How to invoke

POST `${TOOL_BACKEND_URL}/supplier/expedite` with body `{"sku": "<sku>", "supplier_id": "<sid>", "sla_hours": 36}`.

Returns `{status, sla_hours}`.
EOF
```

- [ ] **Step 3: `substitute_swap/SKILL.md`**

```bash
cat > examples/04-skill-routing-loop/skills/substitute_swap/SKILL.md <<'EOF'
---
name: substitute_swap
description: Swap to substitute material via MES. Picks the best candidate using a multi-criteria scorer (Python).
version: 0.1.0
---

# Substitute Swap

## When to use

Material has one or more in-stock substitute SKUs (same Material ObjectType,
distinguishable by SKU prefix `SUB-` and `material_risk == "normal"`).

## How to invoke

Execute `pick_substitute.py` with arg `--sku <sku>`. The script:
1. Reads candidate SUB-* materials from environment-supplied JSON (passed in by DA).
2. Scores by weighted criteria (stock 0.4, compat 0.3, cost-delta 0.2, lead-time 0.1).
3. Calls MES via POST `${TOOL_BACKEND_URL}/mes/swap`.
4. Prints `{chosen_sku, scores}` JSON to stdout.

DA reads stdout, surfaces chosen substitute and rationale to user.
EOF
```

- [ ] **Step 4: `substitute_swap/pick_substitute.py`**

```bash
cat > examples/04-skill-routing-loop/skills/substitute_swap/pick_substitute.py <<'PYEOF'
#!/usr/bin/env python3
"""Multi-criteria substitute picker.

Reads candidate substitutes from CANDIDATES env var (JSON string) and the target
sku from --sku arg. The DA agent is expected to populate CANDIDATES by querying
the BKN before calling this script.

Schema for CANDIDATES:
  [{"sku": "SUB-001A", "stock": 200, "compat_score": 0.95,
    "cost_delta_pct": 5, "lead_time_hours": 2}, ...]
"""
import argparse
import json
import os
import sys
import urllib.request

WEIGHTS = {"stock": 0.4, "compat": 0.3, "cost_delta": 0.2, "lead_time": 0.1}


def normalize(value, max_value, invert=False):
    if max_value == 0:
        return 0.0
    score = min(1.0, value / max_value)
    return 1.0 - score if invert else score


def score_candidate(c, max_stock, max_lead):
    return (
        WEIGHTS["stock"] * normalize(c["stock"], max_stock)
        + WEIGHTS["compat"] * c["compat_score"]
        + WEIGHTS["cost_delta"] * normalize(c["cost_delta_pct"], 100, invert=True)
        + WEIGHTS["lead_time"] * normalize(c["lead_time_hours"], max_lead, invert=True)
    )


def call_mes(target_sku, chosen_sku):
    url = os.environ["TOOL_BACKEND_URL"].rstrip("/") + "/mes/swap"
    body = json.dumps({"from_sku": target_sku, "to_sku": chosen_sku}).encode()
    req = urllib.request.Request(url, data=body, method="POST",
                                 headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=10) as resp:
        return json.loads(resp.read())


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--sku", required=True, help="Target material SKU")
    args = parser.parse_args()

    candidates = json.loads(os.environ.get("CANDIDATES", "[]"))
    if not candidates:
        print(json.dumps({"error": "no candidates provided"}), file=sys.stdout)
        sys.exit(1)

    max_stock = max(c["stock"] for c in candidates)
    max_lead = max(c["lead_time_hours"] for c in candidates)
    scored = [(c, score_candidate(c, max_stock, max_lead)) for c in candidates]
    scored.sort(key=lambda x: x[1], reverse=True)
    chosen = scored[0][0]

    mes_resp = call_mes(args.sku, chosen["sku"])

    out = {
        "chosen_sku": chosen["sku"],
        "scores": [{"sku": c["sku"], "score": round(s, 3)} for c, s in scored],
        "mes_response": mes_resp,
    }
    print(json.dumps(out))


if __name__ == "__main__":
    main()
PYEOF
chmod +x examples/04-skill-routing-loop/skills/substitute_swap/pick_substitute.py
```

- [ ] **Step 5: Local sanity test of pick_substitute.py** (no MES, expects error from urlopen — that's fine, just want to verify Python parses & scores)

```bash
cd examples/04-skill-routing-loop/skills/substitute_swap
CANDIDATES='[{"sku":"SUB-001A","stock":200,"compat_score":0.95,"cost_delta_pct":5,"lead_time_hours":2},
             {"sku":"SUB-001B","stock":80,"compat_score":0.85,"cost_delta_pct":15,"lead_time_hours":4}]' \
TOOL_BACKEND_URL=http://localhost:9999 \
python3 pick_substitute.py --sku MAT-001 2>&1 | head -5 || true
# expect: error reaching localhost:9999 (expected) — but BEFORE that, scoring should have run.
# Re-run with TOOL_BACKEND_URL set to the running mock backend if you have one.
cd -
```

- [ ] **Step 6: Commit**

```bash
git add examples/04-skill-routing-loop/skills/
git commit -m "feat(example-04): 3 Skill packages (standard_replenish, substitute_swap+py, supplier_expedite)"
```

---

## Task 6 · Agent config template

**Files:**
- Create: `examples/04-skill-routing-loop/agent.json`

> Template uses placeholders that `run.sh` replaces with `sed`. mode=react is required (see spec P1-F).

- [ ] **Step 1: Write template**

```bash
cat > examples/04-skill-routing-loop/agent.json <<'EOF'
{
  "input": { "fields": [{ "name": "user_input", "type": "string", "desc": "" }] },
  "output": { "default_format": "markdown" },
  "mode": "react",
  "system_prompt": "你是供应链处置助手。当收到物料告警时：\n1. 调用 find_skills 工具，参数 object_type_id='material'，instance_identities=[{sku: <提到的物料SKU>}]，召回该物料适用的 Skill 候选集。\n2. 用 query_object_instance 或 query_instance_subgraph 读 supplier 当前 capability、material 当前 stock 等证据。\n3. 在候选 Skill 中基于证据选 1 个，调 builtin_skill_load(skill_id) 读 SKILL.md 确认契约。\n4. 必要时 builtin_skill_execute_script(skill_id, script_path) 执行 Python；否则按 SKILL.md 描述调外部 HTTP API。\n5. 输出处置结果，附决策依据（哪个 Skill / 引用了哪些 KN 节点）。",
  "data_source": { "kn_id": "ex04_skill_routing" },
  "skills": {
    "tools": [],
    "agents": [],
    "mcps": [{ "mcp_server_id": "{{MCP_ID}}" }],
    "skills": [
      { "skill_id": "standard_replenish" },
      { "skill_id": "substitute_swap" },
      { "skill_id": "supplier_expedite" }
    ]
  },
  "llms": [{
    "is_default": true,
    "llm_config": {
      "id": "{{LLM_ID}}",
      "name": "{{LLM_NAME}}",
      "max_tokens": 4096,
      "temperature": 0.7,
      "top_p": 1,
      "top_k": 1
    }
  }],
  "react_config": {
    "disable_history_in_a_conversation": false,
    "disable_llm_cache": false
  },
  "memory": { "is_enabled": false },
  "related_question": { "is_enabled": false },
  "plan_mode": { "is_enabled": false }
}
EOF
```

- [ ] **Step 2: Validate JSON syntax**

```bash
python3 -c "import json; json.load(open('examples/04-skill-routing-loop/agent.json'))" && echo "JSON OK"
```

- [ ] **Step 3: Commit**

```bash
git add examples/04-skill-routing-loop/agent.json
git commit -m "feat(example-04): agent config template (mode=react, find_skills MCP, 3 skills)"
```

---

## Task 7 · run.sh — Steps 0-2 (env load, ds connect, csv import)

**Files:**
- Create: `examples/04-skill-routing-loop/run.sh`

We build run.sh in 5 incremental tasks (Task 7-11), each adding a section + a partial verification.

- [ ] **Step 1: Initial run.sh skeleton**

```bash
cat > examples/04-skill-routing-loop/run.sh <<'EOF'
#!/usr/bin/env bash
# =============================================================================
# 04-skill-routing-loop: KN-driven Skill governance end-to-end
#
# Flow: business DB → Vega → BKN → context-loader find_skills →
#       Decision Agent → Skill execute → mock business backend → audit log
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TIMESTAMP=$(date +%s)

# ── CLI flags ────────────────────────────────────────────────────────────────
BONUS=0
usage() {
    cat <<USAGE
Usage: $(basename "$0") [options]

Options:
  --bonus      Run the Bonus segment after main flow (changes SUP-2 capability)
  -h, --help   Show this help
USAGE
}
while [ $# -gt 0 ]; do
    case "$1" in
        --bonus) BONUS=1 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
    esac
    shift
done

# ── Debug helper ─────────────────────────────────────────────────────────────
DEBUG="${DEBUG:-0}"
debug() { [ "$DEBUG" = "1" ] && echo "[debug] $*" >&2 || true; }

# ── Step 0: Load .env ────────────────────────────────────────────────────────
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"

PLATFORM_HOST="${PLATFORM_HOST:?Set PLATFORM_HOST in .env}"
LLM_ID="${LLM_ID:?Set LLM_ID in .env (use: kweaver call /api/mf-model-manager/v1/llm/list)}"
LLM_NAME="${LLM_NAME:-deepseek-v3.2}"
DB_HOST="${DB_HOST:?Set DB_HOST in .env}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"
TOOL_BACKEND_PORT="${TOOL_BACKEND_PORT:-8765}"

DS_NAME="ex04_ds_${TIMESTAMP}"
KN_ID="ex04_skill_routing"   # fixed, must match network.bkn frontmatter
TABLE_PREFIX="ex04_${TIMESTAMP}_"

# Track resources for cleanup
DS_ID="" TMP_KN_ID="" MCP_ID="" AGENT_ID=""
SKILL_IDS=()
TOOL_BACKEND_PID=""

cleanup() {
    echo ""
    echo "=== Cleanup ==="
    [ -n "$AGENT_ID" ] && {
        kweaver agent unpublish "$AGENT_ID" 2>/dev/null || true
        kweaver agent delete "$AGENT_ID" -y 2>/dev/null && echo "  ✓ agent $AGENT_ID"
    }
    [ -n "$MCP_ID" ] && {
        kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID/status" -X POST \
            -H "x-business-domain: bd_public" -d '{"status":"offline"}' >/dev/null 2>&1 || true
        kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID" -X DELETE \
            -H "x-business-domain: bd_public" >/dev/null 2>&1 && echo "  ✓ mcp $MCP_ID"
    }
    for sid in "${SKILL_IDS[@]:-}"; do
        [ -z "$sid" ] && continue
        kweaver skill status "$sid" offline >/dev/null 2>&1 || true
        echo y | kweaver skill delete "$sid" >/dev/null 2>&1 && echo "  ✓ skill $sid"
    done
    kweaver bkn delete "$KN_ID" -y >/dev/null 2>&1 && echo "  ✓ kn $KN_ID" || true
    [ -n "$TMP_KN_ID" ] && kweaver bkn delete "$TMP_KN_ID" -y >/dev/null 2>&1 && echo "  ✓ tmp kn $TMP_KN_ID" || true
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y >/dev/null 2>&1 && echo "  ✓ ds $DS_ID"
    [ -n "$TOOL_BACKEND_PID" ] && kill "$TOOL_BACKEND_PID" 2>/dev/null && echo "  ✓ mock backend pid $TOOL_BACKEND_PID"
}
trap cleanup EXIT

# ── Step 1: Connect MySQL datasource ─────────────────────────────────────────
echo "=== Step 1: Connect MySQL datasource ==="
DS_RAW=$(kweaver ds connect mysql "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    --account "$DB_USER" --password "$DB_PASS" --name "$DS_NAME" 2>&1)
DS_ID=$(echo "$DS_RAW" | python3 -c "
import sys, json, re
raw = sys.stdin.read()
m = re.search(r'\{[\s\S]*\}', raw)
print(json.loads(m.group())['datasource_id'])
")
[ -z "$DS_ID" ] && { echo "ERROR: ds connect failed" >&2; exit 1; }
echo "  Datasource: $DS_ID"

# ── Step 2: Import CSVs (creates dataviews and a temp KN we discard) ─────────
echo ""
echo "=== Step 2: Import CSVs and provision Vega dataviews ==="
TMP_KN_RAW=$(kweaver bkn create-from-csv "$DS_ID" \
    --files "$SCRIPT_DIR/data/*.csv" \
    --name "ex04_tmp_${TIMESTAMP}" \
    --table-prefix "$TABLE_PREFIX" \
    --build --timeout 180 2>&1)
TMP_KN_ID=$(echo "$TMP_KN_RAW" | python3 -c "
import sys, json, re
raw = sys.stdin.read()
m = re.search(r'\{[\s\S]*?\"kn_id\"[\s\S]*?\}', raw)
print(json.loads(m.group())['kn_id'])
")
echo "  Tmp KN (will discard, keep dataviews): $TMP_KN_ID"

# Look up dataview IDs by name
DV_LIST=$(kweaver dataview list --datasource-id "$DS_ID" --limit 50 2>&1)
get_dv_id() {
    local table="$1"
    echo "$DV_LIST" | python3 -c "
import sys, json
raw = sys.stdin.read()
i = raw.find('[')
for v in json.loads(raw[i:]):
    if v.get('name') == '$table':
        print(v.get('id'))
        break
"
}
MATERIALS_DV_ID=$(get_dv_id "${TABLE_PREFIX}materials")
SUPPLIERS_DV_ID=$(get_dv_id "${TABLE_PREFIX}suppliers")
SKILLS_DV_ID=$(get_dv_id "${TABLE_PREFIX}skills")
echo "  Dataview IDs: materials=$MATERIALS_DV_ID, suppliers=$SUPPLIERS_DV_ID, skills=$SKILLS_DV_ID"

# Discard the auto-generated KN; we'll push our own with controlled OT IDs
kweaver bkn delete "$TMP_KN_ID" -y >/dev/null
TMP_KN_ID=""

EOF
chmod +x examples/04-skill-routing-loop/run.sh
```

- [ ] **Step 2: Run partial — only Steps 0-2** (smoke test)

```bash
cd examples/04-skill-routing-loop
cp env.sample .env
# Fill .env with your platform/db values (placeholders below):
sed -i.bak 's|PLATFORM_HOST=.*|PLATFORM_HOST=https://platform.example.com|;
            s|LLM_ID=.*|LLM_ID=<your-llm-id>|;
            s|DB_HOST=.*|DB_HOST=db.example.com|;
            s|DB_NAME=.*|DB_NAME=supply_chain|;
            s|DB_PASS=.*|DB_PASS=<your-db-password>|' .env

bash run.sh 2>&1 | head -30
# expect: Steps 0-2 print success; cleanup runs at exit (script ends after Step 2)
cd -
```

- [ ] **Step 3: Commit**

```bash
git add examples/04-skill-routing-loop/run.sh
git commit -m "feat(example-04): run.sh Steps 0-2 — env load, ds connect, csv import"
```

---

## Task 8 · run.sh — Steps 3-5 (bkn schema substitution + push + build)

**Files:**
- Modify: `examples/04-skill-routing-loop/run.sh` (append Steps 3-5)

- [ ] **Step 1: Append Steps 3-5**

```bash
cat >> examples/04-skill-routing-loop/run.sh <<'EOF'
# ── Step 3: Render BKN templates with dataview IDs ───────────────────────────
echo ""
echo "=== Step 3: Render BKN templates with dataview IDs ==="
RENDERED_BKN="$SCRIPT_DIR/.rendered-bkn"
rm -rf "$RENDERED_BKN"
cp -r "$SCRIPT_DIR/bkn" "$RENDERED_BKN"
sed -i.bak \
    -e "s|{{MATERIALS_DV_ID}}|$MATERIALS_DV_ID|" \
    -e "s|{{MATERIALS_DV_NAME}}|${TABLE_PREFIX}materials|" \
    "$RENDERED_BKN/object_types/material.bkn"
sed -i.bak \
    -e "s|{{SUPPLIERS_DV_ID}}|$SUPPLIERS_DV_ID|" \
    -e "s|{{SUPPLIERS_DV_NAME}}|${TABLE_PREFIX}suppliers|" \
    "$RENDERED_BKN/object_types/supplier.bkn"
sed -i.bak \
    -e "s|{{SKILLS_DV_ID}}|$SKILLS_DV_ID|" \
    -e "s|{{SKILLS_DV_NAME}}|${TABLE_PREFIX}skills|" \
    "$RENDERED_BKN/object_types/skills.bkn"
find "$RENDERED_BKN" -name '*.bak' -delete
echo "  ✓ rendered .bkn files"

# ── Step 4: Push BKN ─────────────────────────────────────────────────────────
echo ""
echo "=== Step 4: bkn push (deploy schema + relations) ==="
kweaver bkn validate "$RENDERED_BKN" 2>&1 | tail -1
PUSH_RAW=$(kweaver bkn push "$RENDERED_BKN" 2>&1)
echo "$PUSH_RAW" | tail -3
# kn_id is fixed (network.bkn frontmatter id) — just confirm push succeeded
echo "$PUSH_RAW" | grep -q "\"kn_id\"" || { echo "ERROR: bkn push failed" >&2; exit 1; }
echo "  ✓ KN: $KN_ID"

# ── Step 5: Build KN ─────────────────────────────────────────────────────────
echo ""
echo "=== Step 5: Build KN (sync) ==="
kweaver bkn build "$KN_ID" --wait --timeout 60 2>&1 | tail -2

EOF
```

- [ ] **Step 2: Smoke-test through Step 5**

```bash
cd examples/04-skill-routing-loop && bash run.sh 2>&1 | tail -20; cd -
# expect: Steps 1-5 print success, cleanup at exit
```

- [ ] **Step 3: Commit**

```bash
git add examples/04-skill-routing-loop/run.sh
git commit -m "feat(example-04): run.sh Steps 3-5 — render templates, bkn push, build"
```

---

## Task 9 · run.sh — Steps 6-8 (mock backend + skill register + MCP register)

**Files:**
- Modify: `examples/04-skill-routing-loop/run.sh` (append Steps 6-8)

- [ ] **Step 1: Append Steps 6-8**

```bash
cat >> examples/04-skill-routing-loop/run.sh <<'EOF'
# ── Step 6: Start mock business backend ──────────────────────────────────────
echo ""
echo "=== Step 6: Start mock business backend (port $TOOL_BACKEND_PORT) ==="
TOOL_BACKEND_URL="http://127.0.0.1:$TOOL_BACKEND_PORT"
DB_HOST="$DB_HOST" DB_PORT="$DB_PORT" DB_NAME="$DB_NAME" \
DB_USER="$DB_USER" DB_PASS="$DB_PASS" \
TOOL_BACKEND_PORT="$TOOL_BACKEND_PORT" \
SUPPLIERS_TABLE="${TABLE_PREFIX}suppliers" \
python3 "$SCRIPT_DIR/tool_backend/server.py" >"$SCRIPT_DIR/.tool_backend.log" 2>&1 &
TOOL_BACKEND_PID=$!
sleep 2
curl -s "$TOOL_BACKEND_URL/healthz" | grep -q '"ok"' \
    || { echo "ERROR: mock backend failed; see .tool_backend.log" >&2; exit 1; }
echo "  ✓ mock backend pid $TOOL_BACKEND_PID"

# ── Step 7: Register Skill packages ──────────────────────────────────────────
echo ""
echo "=== Step 7: Register Skill packages ==="
for skill_dir in "$SCRIPT_DIR"/skills/*/; do
    skill_name=$(basename "$skill_dir")
    zip_path="$SCRIPT_DIR/.${skill_name}.zip"
    rm -f "$zip_path"
    (cd "$skill_dir" && zip -qr "$zip_path" .)
    REG_RAW=$(kweaver skill register --zip-file "$zip_path" 2>&1)
    sid=$(echo "$REG_RAW" | python3 -c "
import sys, json, re
raw = sys.stdin.read()
m = re.search(r'\{[\s\S]*?\}', raw)
print(json.loads(m.group())['id'])
")
    kweaver skill status "$sid" published >/dev/null
    SKILL_IDS+=("$sid")
    echo "  ✓ $skill_name → $sid (published)"
    rm -f "$zip_path"
done

# ── Step 8: Register context-loader MCP server (with X-Kn-ID header) ─────────
echo ""
echo "=== Step 8: Register context-loader MCP server ==="
MCP_REG_BODY=$(python3 -c "
import json
print(json.dumps({
    'mode': 'stream',
    'url': '$PLATFORM_HOST/api/agent-retrieval/v1/mcp',
    'name': 'ex04_ctx_loader_${TIMESTAMP}',
    'description': 'context-loader MCP for find_skills',
    'creation_type': 'custom',
    'headers': {'X-Kn-ID': '$KN_ID'},
}))
")
MCP_RAW=$(kweaver call /api/agent-operator-integration/v1/mcp/ -X POST \
    -H "Content-Type: application/json" \
    -H "x-business-domain: bd_public" \
    -d "$MCP_REG_BODY" 2>&1)
MCP_ID=$(echo "$MCP_RAW" | python3 -c "
import sys, json, re
raw = sys.stdin.read()
m = re.search(r'\{[\s\S]*?\}', raw)
print(json.loads(m.group())['mcp_id'])
")
kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID/status" -X POST \
    -H "x-business-domain: bd_public" \
    -d '{"status":"published"}' >/dev/null
echo "  ✓ MCP $MCP_ID (published, X-Kn-ID=$KN_ID)"

EOF
```

- [ ] **Step 2: Smoke-test through Step 8**

```bash
cd examples/04-skill-routing-loop && bash run.sh 2>&1 | tail -30; cd -
# expect: 3 skills registered, 1 MCP registered, all cleaned up at exit
```

- [ ] **Step 3: Commit**

```bash
git add examples/04-skill-routing-loop/run.sh
git commit -m "feat(example-04): run.sh Steps 6-8 — mock backend, skill register, MCP register"
```

---

## Task 10 · run.sh — Steps 9-11 (agent create + 3 chats + history)

**Files:**
- Modify: `examples/04-skill-routing-loop/run.sh` (append Steps 9-11)

- [ ] **Step 1: Append Steps 9-11**

```bash
cat >> examples/04-skill-routing-loop/run.sh <<'EOF'
# ── Step 9: Render agent.json with MCP_ID + LLM_ID ───────────────────────────
echo ""
echo "=== Step 9: Render agent.json ==="
RENDERED_AGENT="$SCRIPT_DIR/.rendered-agent.json"
sed \
    -e "s|{{MCP_ID}}|$MCP_ID|" \
    -e "s|{{LLM_ID}}|$LLM_ID|" \
    -e "s|{{LLM_NAME}}|$LLM_NAME|" \
    "$SCRIPT_DIR/agent.json" > "$RENDERED_AGENT"
python3 -c "import json; json.load(open('$RENDERED_AGENT'))" >/dev/null
echo "  ✓ agent.json rendered"

# ── Step 10: Create + publish agent ──────────────────────────────────────────
echo ""
echo "=== Step 10: Create + publish Decision Agent ==="
AGENT_NAME="ex04_skill_routing_${TIMESTAMP}"
CREATE_RAW=$(kweaver agent create \
    --name "$AGENT_NAME" \
    --profile "Example 04 — KN-driven skill routing" \
    --config "$RENDERED_AGENT" 2>&1)
AGENT_ID=$(echo "$CREATE_RAW" | python3 -c "
import sys, json, re
raw = sys.stdin.read()
m = re.search(r'\{[\s\S]*?\}', raw)
print(json.loads(m.group())['id'])
")
kweaver agent publish "$AGENT_ID" >/dev/null
echo "  ✓ agent $AGENT_ID (published)"

# ── Step 11: Trigger 3 critical-stock alerts; show DA's decisions ────────────
# Use --stream (not --no-stream): nginx proxy in front of the platform returns
# 504 Gateway Timeout on long-running buffered responses; example 02 also uses
# --stream for the same reason.
echo ""
echo "=== Step 11: Trigger 3 alerts (one per material) ==="
for sku in MAT-001 MAT-002 MAT-003; do
    echo ""
    echo "--- $sku ---"
    kweaver agent chat "$AGENT_ID" \
        -m "Material $sku hit critical stock level. Use find_skills to identify applicable skills, query the BKN for evidence (supplier capability, etc.), pick the best skill, and report what you would execute." \
        --stream 2>&1 \
        | sed '/^(node:.*Warning:/d; /trace-warnings/d; /To continue this conversation/,$d' \
        | tail -40
done

EOF
```

- [ ] **Step 2: Smoke-test through Step 11** — first real e2e

```bash
cd examples/04-skill-routing-loop && bash run.sh 2>&1 | tee /tmp/ex04-run.log | tail -50; cd -
# expect:
#   MAT-001 → DA chooses substitute_swap with reasoning citing SUB-001A/B
#   MAT-002 → DA chooses supplier_expedite with reasoning citing SUP-2.capability=expedite
#   MAT-003 → DA chooses standard_replenish (only candidate)
```

> If LLM calls succeed but content is "I cannot access tools": MCP wiring broke. Check `.tool_backend.log` and verify `mode: "react"` is in `agent.json`. If `find_skills` returns empty for an instance: re-check materials.csv `bound_skill_id` values match `skills.csv` skill_id values exactly.

- [ ] **Step 3: Commit**

```bash
git add examples/04-skill-routing-loop/run.sh
git commit -m "feat(example-04): run.sh Steps 9-11 — agent create + 3-material chat trace"
```

---

## Task 11 · run.sh — Step 12 Bonus + final polish

**Files:**
- Modify: `examples/04-skill-routing-loop/run.sh` (append Bonus + cleanup polish)

- [ ] **Step 1: Append Bonus block**

```bash
cat >> examples/04-skill-routing-loop/run.sh <<'EOF'
# ── Step 12: Bonus (optional via --bonus) ────────────────────────────────────
if [ "$BONUS" = "1" ]; then
    echo ""
    echo "=== Bonus: change SUP-2 capability → AI re-routes MAT-002 ==="

    echo ""
    echo "[business system] update SUP-2.capability: expedite → normal"
    curl -s -X POST "$TOOL_BACKEND_URL/admin/supplier-capability" \
        -H "Content-Type: application/json" \
        -d '{"supplier_id":"SUP-2","capability":"normal"}' | python3 -m json.tool

    sleep 2  # let Vega see the new state (Step D in spec — verify in this run)

    echo ""
    echo "--- MAT-002 (re-trigger after capability change) ---"
    kweaver agent chat "$AGENT_ID" \
        -m "Material MAT-002 hit critical stock level again. Decide and report." \
        --stream 2>&1 \
        | sed '/^(node:.*Warning:/d; /trace-warnings/d; /To continue this conversation/,$d' \
        | tail -40

    echo ""
    echo ">>> Compare with the MAT-002 result above (Step 11)."
    echo ">>> If business→AI propagation works: this run picks standard_replenish."
fi

echo ""
echo "=== All steps completed; cleanup runs on exit ==="
EOF
```

- [ ] **Step 2: Final shellcheck + run with --bonus**

```bash
shellcheck examples/04-skill-routing-loop/run.sh 2>&1 | head -10 || true  # warn-only
cd examples/04-skill-routing-loop && bash run.sh --bonus 2>&1 | tee /tmp/ex04-bonus.log | tail -30; cd -
# expect: MAT-002 second time chooses standard_replenish (or DA explicitly notes capability changed)
```

- [ ] **Step 3: Commit**

```bash
git add examples/04-skill-routing-loop/run.sh
git commit -m "feat(example-04): run.sh Bonus — business-system change → AI re-routes"
```

---

## Task 12 · README (English + Chinese)

**Files:**
- Create: `examples/04-skill-routing-loop/README.md`
- Create: `examples/04-skill-routing-loop/README.zh.md`

- [ ] **Step 1: `README.md` (English)**

```bash
cat > examples/04-skill-routing-loop/README.md <<'EOF'
# 04 · Skill Routing Loop — KN-driven Skill Governance

> [中文版](./README.zh.md)

> 3 materials trigger the same critical alert; the Decision Agent picks 3 different
> handling paths — each justified by the knowledge network.

## The Story

Continuing from example 03's procurement engineer: she now sees the disposition
plan already chosen on each alert. Three materials. Three paths. Zero prompts
edited. The `applicable_skill` relation in the business knowledge network and
the supplier's `capability` field decide everything.

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
cd examples/04-skill-routing-loop
cp env.sample .env
vim .env                                    # fill PLATFORM_HOST, LLM_ID, DB_*
pip install -r tool_backend/requirements.txt
./run.sh                                    # ~5 minutes end-to-end
./run.sh --bonus                            # also run the Bonus segment
```

## What you will see

| Material | KN evidence | DA picks | Why |
|---|---|---|---|
| MAT-001 | binds to `substitute_swap`; SUB-001A/B in stock | substitute_swap | Python scorer ranks substitutes; calls MES |
| MAT-002 | binds to `supplier_expedite`; SUP-2 capability=expedite | supplier_expedite | Supplier can rush — POST to supplier portal |
| MAT-003 | binds to `standard_replenish` only | standard_replenish | Default path — issue PO via ERP |

## Bonus — change business → AI follows

Run `./run.sh --bonus`. The script POSTs to the mock business backend's admin
endpoint to flip SUP-2.capability from `expedite` to `normal`, then re-asks
the Agent about MAT-002. The Decision Agent sees the new BKN state (via Vega)
and switches to `standard_replenish` — without any prompt edit or redeploy.

## How it works (deeper read)

See [`docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md`](../../docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md)
for the full design including:
- BKN schema and the `applicable_skill` direct-mapping FK
- Why MCP server registration must include `X-Kn-ID` header
- Why agent `mode` must be `"react"` (default mode skips tool wiring)
- The 3-step state machine for cleaning up MCPs and Skills

## Cleanup

Resources (KN, MCP, Skills, Agent, Datasource, mock backend process) are
deleted automatically on script exit, success or failure.
EOF
```

- [ ] **Step 2: `README.zh.md` (中文版)**

```bash
cat > examples/04-skill-routing-loop/README.zh.md <<'EOF'
# 04 · Skill Routing Loop — 业务知识网络驱动的 Skill 治理

> [English](./README.md)

> 3 个物料触发同样的库存告警，Decision Agent 给出 3 条不同处置路径——
> 每条都能在业务知识网络里找到依据。

## 故事

续作 03 那位采购工程师：她现在看到每张告警单上已经写好了处置方案。3 个物料、
3 条不同路径，**没改一行 prompt**。BKN 里的 `applicable_skill` 关系 +
供应商节点的 `capability` 字段决定了这一切。

## 这个 example 展示什么

5 个组件协同跑通一个完整闭环：

| 组件 | 职责 |
|---|---|
| **execution-factory** | 注册 / 版本化 3 个 Skill 包 |
| **业务知识网络（BKN）** | 通过 `applicable_skill` 关系把 Skill 绑到物料 |
| **Vega** | 把 BKN ObjectType 映射到 MySQL 表（读多写少） |
| **context-loader (`find_skills`)** | 按物料实例召回适用的 Skill |
| **Decision Agent** | 读 KN 证据 → 选 Skill → 执行 → 审计 |

## 前置条件

- `kweaver` CLI ≥ 0.7.1
- 启用了 Decision Agent + execution-factory + Vega 的 KWeaver 平台
  （先 `kweaver auth login <平台地址> [--insecure]`）
- **平台能访问到**的 MySQL（不是你笔记本上的），且账号有 CREATE/INSERT/SELECT/UPDATE 权限
- `python3`（依赖 Flask + mysql-connector-python，
  `pip install -r tool_backend/requirements.txt`）
- 平台模型工厂里注册的 LLM 模型（用
  `kweaver call /api/mf-model-manager/v1/llm/list` 拿 model_id）

## 快速开始

```bash
cd examples/04-skill-routing-loop
cp env.sample .env
vim .env                                    # 填 PLATFORM_HOST、LLM_ID、DB_*
pip install -r tool_backend/requirements.txt
./run.sh                                    # 端到端约 5 分钟
./run.sh --bonus                            # 跑 Bonus 段
```

## 你会看到什么

| 物料 | KN 证据 | DA 选 | 原因 |
|---|---|---|---|
| MAT-001 | 绑定 `substitute_swap`；SUB-001A/B 有库存 | substitute_swap | Python 算法打分挑替代料 → 调 MES |
| MAT-002 | 绑定 `supplier_expedite`；SUP-2 capability=expedite | supplier_expedite | 供应商能加急 → POST 供应商门户 |
| MAT-003 | 只绑定 `standard_replenish` | standard_replenish | 默认路径 → 走 ERP 下单 |

## Bonus — 改业务，AI 跟着变

`./run.sh --bonus` 会调 mock 业务系统的 admin 端点把 SUP-2 的 capability
从 `expedite` 改为 `normal`，然后重新让 Agent 处理 MAT-002。Decision Agent
通过 Vega 看到 BKN 的新状态，自动切到 `standard_replenish`——
**没改 prompt、没重新部署任何服务**。

## 原理细节

完整设计文档：[`docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md`](../../docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md)

包括：
- BKN schema 和 `applicable_skill` 的 direct-mapping FK
- 为什么 MCP server 注册时必须带 `X-Kn-ID` header
- 为什么 agent `mode` 必须是 `"react"`（默认模式不挂载工具）
- MCP / Skill 清理的三态机协议

## Cleanup

脚本退出时（成功 / 失败）自动清理所有资源：KN、MCP、Skills、Agent、Datasource、
mock backend 进程。
EOF
```

- [ ] **Step 3: Commit**

```bash
git add examples/04-skill-routing-loop/README.md examples/04-skill-routing-loop/README.zh.md
git commit -m "docs(example-04): README EN + 中文版"
```

---

## Task 13 · Update examples/README.md index

**Files:**
- Modify: `examples/README.md`
- Modify: `examples/README.zh.md`

- [ ] **Step 1: Find existing 03 row in examples/README.md**

```bash
grep -n "03-action-lifecycle" examples/README.md
```

- [ ] **Step 2: Insert 04 row after 03 (English)**

Use Edit tool. Find the line ending the table for 03 (after `Schedule → Audit Log |`) and add:

```markdown
| [04-skill-routing-loop](./04-skill-routing-loop/) | *3 materials, 3 critical alerts, 3 different handling paths — each justified by the knowledge network* | MySQL → BKN (via Vega) → find_skills → Decision Agent → Skill → Action |
```

- [ ] **Step 3: Insert 04 row after 03 (中文版)**

Same change to `examples/README.zh.md` with Chinese narrative:

```markdown
| [04-skill-routing-loop](./04-skill-routing-loop/) | *3 个物料、3 条 critical 告警、3 条不同处置路径——每条都能在知识网络里找到依据* | MySQL → BKN (经 Vega) → find_skills → Decision Agent → Skill → Action |
```

- [ ] **Step 4: Verify both files render the new row**

```bash
grep "04-skill-routing-loop" examples/README.md examples/README.zh.md
```

- [ ] **Step 5: Commit**

```bash
git add examples/README.md examples/README.zh.md
git commit -m "docs(examples): index — add example 04 (skill routing loop)"
```

---

## Task 14 · Final acceptance — fresh end-to-end run

**Files:** none (verification only)

- [ ] **Step 1: Verify clean state on platform**

```bash
kweaver bkn list --pretty 2>&1 | grep -i "ex04" || echo "no ex04 KN — clean"
kweaver agent list --limit 50 2>&1 | grep -i "ex04" || echo "no ex04 agent — clean"
kweaver call "/api/agent-operator-integration/v1/mcp/list?page=1&page_size=20" 2>&1 \
    | python3 -c "import sys,json,re; raw=sys.stdin.read(); m=re.search(r'\{[\s\S]*\}',raw); d=json.loads(m.group()); print('ex04 MCPs:', sum(1 for x in d.get('data',[]) if 'ex04' in x.get('name','')))"
```

- [ ] **Step 2: Run from scratch with timing**

```bash
cd examples/04-skill-routing-loop
time bash run.sh 2>&1 | tee /tmp/ex04-final.log
cd -
```

Expected:
- Total runtime ≤ 5 minutes
- Steps 1-11 all show ✓
- 3 chat outputs each show DA picking the documented Skill with KN-cited reasoning
- Cleanup section at exit shows all 6 resources removed

- [ ] **Step 3: Run Bonus from scratch**

```bash
cd examples/04-skill-routing-loop && time bash run.sh --bonus 2>&1 | tee /tmp/ex04-bonus-final.log; cd -
```

Expected: MAT-002 second pass chooses `standard_replenish` (different from MAT-002's first pass which chose `supplier_expedite`).

- [ ] **Step 4: Verify clean state again** (cleanup worked)

Same commands as Step 1.

- [ ] **Step 5: Mark issue checkboxes** (manually on GitHub)

```
gh issue view 312 --repo kweaver-ai/kweaver-core
# In browser, tick the boxes for verified acceptance criteria
```

- [ ] **Step 6: Final commit if any tweaks were needed during acceptance**

```bash
git status
# If any edits made during acceptance, commit them as fix(example-04): ...
```

---

## Task 15 · (Optional, separate session) — Companion WeChat article

This task is intentionally separate from the example implementation. It's a writing
task, not a coding task, and should happen after Task 14 produces real screenshots
and trace data.

Outline lives in spec Part B. Pull the actual trace from one Task 11 run, build 4
diagrams (architecture / 5-thing table / 3-path comparison / 5-hop lineage), and
write per the 9-section structure. Title candidate: 《业务真相即 AI 行为——Skill
治理的端到端实录》.

Suggested process: a separate session with the writing-clearly-and-concisely skill,
fed the spec Part B + Task 11 logs.

---

## Self-Review Done

- ✅ **Spec coverage**: every requirement in spec Part A maps to a task (Tasks 1-13).
  Part B (article) → Task 15 (optional). Part C (issue refs) → already in #312.
- ✅ **No placeholders**: every step has executable code or commands. Template
  placeholders `{{...}}` in `.bkn` and `agent.json` are intentional and explicitly
  substituted by `sed` in run.sh.
- ✅ **Type / name consistency**: `KN_ID="ex04_skill_routing"` is referenced
  identically in network.bkn frontmatter, agent.json data_source, MCP X-Kn-ID
  header, and cleanup. Skill names (`standard_replenish` / `substitute_swap` /
  `supplier_expedite`) match across skills.csv, SKILL.md frontmatter, and
  agent.json skills array.
- ✅ **Sequencing**: Task 7 (ds + dataviews) → Task 8 (BKN with those dataview IDs)
  → Task 9 (skills + MCP) → Task 10 (agent + chat) — each task's verification
  exercises only what's been built so far.

## Open Question Resolution Plan

- **Spec Open Q C** (Skill internal Python → MCP query for substitutes): Task 5's
  `pick_substitute.py` reads candidates from `CANDIDATES` env var (DA populates it
  upstream by calling `query_object_instance` directly), avoiding skill-internal
  MCP calls. If the system_prompt fails to make the LLM populate CANDIDATES,
  Task 11 verification will surface it; fallback is to bake one substitute pick
  in the SKILL.md and accept the simplification.
- **Spec Open Q D** (Vega real-time freshness): Task 11 Bonus's `sleep 2` between
  admin POST and re-chat probes this. If 2s isn't enough, increase; if there's a
  real cache TTL, document in README and Task 11 sleep grows.
- **Spec Open Q G** (skill_id alignment between skills.csv and execution-factory
  registrations): Task 7 registers Skill packages where `name` field of
  `SKILL.md` frontmatter matches the `skill_id` in skills.csv. The `id` returned
  by execution-factory is a UUID that's separate; what matters for `builtin_skill_load`
  is the LLM passes a string the agent's `skills.skills[].skill_id` array
  recognizes — and that array uses the human-readable name (Task 6 agent.json).
  Verified end-to-end works in Task 11.
