#!/usr/bin/env python3
"""
Replace {{*_RES_ID}} placeholders in worldcup-bkn-vega with Vega resource UUIDs.

Input: JSON object mapping placeholder name → uuid, e.g.:
  {"MATCHES_RES_ID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", ...}

Paths default to sibling dirs under examples/06-world-cup/.

Example:
  kweaver vega catalog resources <cid> --category table --limit 500 > /tmp/res.json
  # Build mapping by hand from names wc_* → MATCHES_RES_ID style keys, or use jq.
  python3 scripts/render_worldcup_bkn_vega_placeholders.py --mapping mappings.json \\
    --src worldcup-bkn-vega --dst .rendered-bkn-vega
"""
from __future__ import annotations

import argparse
import json
import re
import shutil
from pathlib import Path

PLACEHOLDER_RE = re.compile(r"\{\{([A-Z0-9_]+)_RES_ID\}\}")


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--mapping",
        required=True,
        type=Path,
        help="JSON object: PLACEHOLDER_NAME (e.g. MATCHES_RES_ID) -> uuid string",
    )
    ap.add_argument(
        "--src",
        type=Path,
        default=Path("worldcup-bkn-vega"),
        help="Source BKN tree (with {{*_RES_ID}})",
    )
    ap.add_argument(
        "--dst",
        type=Path,
        default=Path(".rendered-bkn-vega"),
        help="Output directory (fresh copy + substitutions)",
    )
    args = ap.parse_args()

    mapping: dict[str, str] = json.loads(args.mapping.read_text(encoding="utf-8"))

    root = Path(__file__).resolve().parents[1]
    src = (root / args.src).resolve()
    dst = (root / args.dst).resolve()
    if dst.exists():
        shutil.rmtree(dst)
    shutil.copytree(src, dst)

    missing: set[str] = set()

    def sub_text(text: str) -> str:
        def repl(m: re.Match[str]) -> str:
            key = f"{m.group(1)}_RES_ID"
            uid = mapping.get(key)
            if not uid:
                missing.add(key)
                return m.group(0)
            return uid

        return PLACEHOLDER_RE.sub(repl, text)

    for path in sorted(dst.rglob("*.bkn")):
        old = path.read_text(encoding="utf-8")
        new = sub_text(old)
        path.write_text(new, encoding="utf-8")

    if missing:
        raise SystemExit(
            "Missing UUIDs for placeholders: {}\n(fill {})".format(
                ", ".join(sorted(missing)),
                args.mapping,
            )
        )
    print(f"Wrote {dst} ({len(mapping)} mapping keys supplied)")


if __name__ == "__main__":
    main()
