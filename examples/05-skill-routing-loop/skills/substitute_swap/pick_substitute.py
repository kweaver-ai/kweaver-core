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
        candidates = [
            {"sku": "SUB-001A", "stock": 200, "compat_score": 0.95, "cost_delta_pct": 5, "lead_time_hours": 2},
            {"sku": "SUB-001B", "stock": 80, "compat_score": 0.85, "cost_delta_pct": 2, "lead_time_hours": 4},
        ]
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
