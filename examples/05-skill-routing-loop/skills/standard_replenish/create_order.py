#!/usr/bin/env python3
import json
import os
import sys
import urllib.request


def main():
    sku = os.environ.get("SKU", "MAT-003")
    qty = int(os.environ.get("QTY", "65"))
    base = os.environ["TOOL_BACKEND_URL"].rstrip("/")
    body = json.dumps({"sku": sku, "qty": qty}).encode()
    req = urllib.request.Request(
        base + "/procurement/order",
        data=body,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=10) as resp:
        result = json.loads(resp.read())
    print(json.dumps({"skill": "standard_replenish", "request": {"sku": sku, "qty": qty}, "response": result}))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(json.dumps({"skill": "standard_replenish", "error": str(exc)}), file=sys.stderr)
        raise
