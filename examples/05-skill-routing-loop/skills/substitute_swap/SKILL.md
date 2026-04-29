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
