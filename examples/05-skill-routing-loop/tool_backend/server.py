#!/usr/bin/env python3
"""
Mock business backend for example 05.

Three business endpoints (called by Skills):
  POST /procurement/order          - standard_replenish
  POST /mes/swap                   - substitute_swap
  POST /supplier/expedite          - supplier_expedite

One admin endpoint (called by Bonus):
  POST /admin/material-binding     - simulate "ops re-binds a material's
                                     applicable Skill in the business system"
"""
import os
import re
import sys

import mysql.connector
from flask import Flask, jsonify, request

VALID_SKILL_IDS = {
    "standard_replenish",
    "substitute_swap",
    "supplier_expedite",
    os.environ.get("STANDARD_REPLENISH_ID", ""),
    os.environ.get("SUBSTITUTE_SWAP_ID", ""),
    os.environ.get("SUPPLIER_EXPEDITE_ID", ""),
}
VALID_SKILL_IDS.discard("")

app = Flask(__name__)
PORT = int(os.environ.get("TOOL_BACKEND_PORT", "8765"))
DB_CONFIG = {
    "host": os.environ["DB_HOST"],
    "port": int(os.environ.get("DB_PORT", "3306")),
    "database": os.environ["DB_NAME"],
    "user": os.environ["DB_USER"],
    "password": os.environ["DB_PASS"],
}

# run.sh imports CSVs with --table-prefix, so the real table is ex05_<ts>_materials.
MATERIALS_TABLE = os.environ.get("MATERIALS_TABLE", "materials")
if not re.fullmatch(r"[A-Za-z0-9_]+", MATERIALS_TABLE):
    raise ValueError(f"Invalid MATERIALS_TABLE: {MATERIALS_TABLE!r}")

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

@app.post("/admin/material-binding")
def admin_set_material_binding():
    """Simulate ops re-binding a material's applicable Skill; writes to MySQL.

    body: { "sku": "MAT-002", "bound_skill_id": "standard_replenish" }
    """
    body = request.get_json(force=True)
    sku = body.get("sku")
    bound_skill_id = body.get("bound_skill_id")
    if not sku or bound_skill_id not in VALID_SKILL_IDS:
        return jsonify({"error": "invalid request"}), 400
    conn = mysql.connector.connect(**DB_CONFIG)
    try:
        cur = conn.cursor()
        cur.execute(
            f"UPDATE {MATERIALS_TABLE} SET bound_skill_id=%s WHERE sku=%s",
            (bound_skill_id, sku),
        )
        affected = cur.rowcount
        conn.commit()
    finally:
        conn.close()
    print(f"[admin] material {sku} bound_skill_id -> {bound_skill_id} ({affected} rows)", file=sys.stderr)
    return jsonify({"updated": affected, "sku": sku, "bound_skill_id": bound_skill_id})


@app.get("/healthz")
def healthz():
    return jsonify({"status": "ok"})


if __name__ == "__main__":
    print(f"[tool_backend] listening on 127.0.0.1:{PORT}", file=sys.stderr)
    app.run(host="127.0.0.1", port=PORT, debug=False)
