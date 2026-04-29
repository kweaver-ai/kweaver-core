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
