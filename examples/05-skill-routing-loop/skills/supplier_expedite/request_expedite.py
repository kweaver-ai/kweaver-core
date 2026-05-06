#!/usr/bin/env python3
import json
import os
import sys
import urllib.request


def main():
    sku = os.environ.get("SKU", "MAT-002")
    supplier_id = os.environ.get("SUPPLIER_ID", "SUP-2")
    sla_hours = int(os.environ.get("SLA_HOURS", "36"))
    base = os.environ["TOOL_BACKEND_URL"].rstrip("/")
    body = json.dumps({"sku": sku, "supplier_id": supplier_id, "sla_hours": sla_hours}).encode()
    req = urllib.request.Request(
        base + "/supplier/expedite",
        data=body,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=10) as resp:
        result = json.loads(resp.read())
    print(json.dumps({"skill": "supplier_expedite", "request": {"sku": sku, "supplier_id": supplier_id, "sla_hours": sla_hours}, "response": result}))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(json.dumps({"skill": "supplier_expedite", "error": str(exc)}), file=sys.stderr)
        raise
