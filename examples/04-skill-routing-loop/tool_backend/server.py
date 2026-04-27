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
    print(f"[tool_backend] listening on 127.0.0.1:{PORT}", file=sys.stderr)
    app.run(host="127.0.0.1", port=PORT, debug=False)
