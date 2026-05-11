#!/usr/bin/env python3
"""
Drop every BASE TABLE in the MySQL schema named by DB_NAME (from repo .env).

Usage (from examples/06-world-cup):
  ./scripts/mysql_empty_schema.py [--dry-run | --stats]

Loads DB_HOST DB_PORT DB_NAME DB_USER DB_PASS from ./.env (same pattern as example run.sh).
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


def load_env(env_path: Path) -> dict[str, str]:
    out: dict[str, str] = {}
    raw = env_path.read_text(encoding="utf-8", errors="replace")
    for line in raw.splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        m = re.match(r"([A-Za-z_][A-Za-z0-9_]*)=(.*)", line)
        if not m:
            continue
        k, v = m.group(1), m.group(2).strip()
        if len(v) >= 2 and v[0] == v[-1] and v[0] in "\"'":
            v = v[1:-1]
        out[k] = v
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description="DROP all tables in DB_NAME (.env)")
    ap.add_argument("--dry-run", action="store_true", help="Print what would happen; no connection")
    ap.add_argument(
        "--stats",
        action="store_true",
        help="Print BASE TABLE count and names from DB_NAME only; no DROP",
    )
    args = ap.parse_args()

    cup = Path(__file__).resolve().parent.parent
    env_path = cup / ".env"
    if not env_path.is_file():
        print(f"Missing {env_path}", file=sys.stderr)
        return 2

    env = load_env(env_path)
    need = ["DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASS"]
    miss = [k for k in need if k not in env]
    if miss:
        print(f".env missing: {miss}", file=sys.stderr)
        return 2

    try:
        import pymysql  # type: ignore
    except ImportError:
        print("Install PyMySQL: pip install pymysql", file=sys.stderr)
        return 2

    port = int(env["DB_PORT"])
    conn_kwargs = dict(
        host=env["DB_HOST"],
        port=port,
        user=env["DB_USER"],
        password=env["DB_PASS"],
        database=env["DB_NAME"],
        charset="utf8mb4",
    )

    sel = """SELECT TABLE_NAME FROM information_schema.TABLES
        WHERE TABLE_SCHEMA = %s AND TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME"""

    if args.dry_run:
        print(f"Would connect {env['DB_USER']}@{env['DB_HOST']}:{port}/{env['DB_NAME']}")
        print("Would run FK off + DROP TABLE IF EXISTS for each table.")
        return 0

    if args.stats:
        conn = pymysql.connect(**conn_kwargs)
        try:
            with conn.cursor() as cur:
                cur.execute(sel, (env["DB_NAME"],))
                tables = [r[0] for r in cur.fetchall()]
            sch = env["DB_NAME"]
            host = env["DB_HOST"]
            print(f"{sch}@{host}:{port}: {len(tables)} BASE TABLE(s)")
            for t in tables:
                print(f"  - {t}")
        finally:
            conn.close()
        return 0

    conn = pymysql.connect(**conn_kwargs)
    try:
        with conn.cursor() as cur:
            cur.execute(sel, (env["DB_NAME"],))
            tables = [r[0] for r in cur.fetchall()]
            if not tables:
                print(f"No tables in {env['DB_NAME']!r}.")
                return 0

            fq = ", ".join(f"`{t}`" for t in tables)
            cur.execute("SET FOREIGN_KEY_CHECKS=0")
            cur.execute(f"DROP TABLE IF EXISTS {fq}")
            cur.execute("SET FOREIGN_KEY_CHECKS=1")
            conn.commit()
            print(f"OK: dropped {len(tables)} table(s) in {env['DB_NAME']!r} on {env['DB_HOST']!r}:{port}")
    finally:
        conn.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
