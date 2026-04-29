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

Entry shell:

```bash
TOOL_BACKEND_URL={{TOOL_BACKEND_PUBLIC_URL}} CANDIDATES='[{"sku":"SUB-001A","stock":200,"compat_score":0.95,"cost_delta_pct":5,"lead_time_hours":2},{"sku":"SUB-001B","stock":80,"compat_score":0.85,"cost_delta_pct":2,"lead_time_hours":4}]' python pick_substitute.py --sku <sku>
```

The script:
1. Reads candidate SUB-* materials from environment-supplied JSON (passed in by DA).
2. Scores by weighted criteria (stock 0.4, compat 0.3, cost-delta 0.2, lead-time 0.1).
3. Calls MES via POST `${TOOL_BACKEND_URL}/mes/swap`.
4. Prints `{chosen_sku, scores}` JSON to stdout.

DA reads stdout, surfaces chosen substitute and rationale to user.
