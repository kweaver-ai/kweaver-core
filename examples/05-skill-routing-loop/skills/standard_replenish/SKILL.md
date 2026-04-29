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

Entry shell:

```bash
TOOL_BACKEND_URL={{TOOL_BACKEND_PUBLIC_URL}} SKU=<sku> QTY=<integer> python create_order.py
```

The script POSTs `${TOOL_BACKEND_URL}/procurement/order` with body
`{"sku": "<sku>", "qty": <integer>}`.

Returns `{po_number, status}`. Surface the PO number in the final answer.
