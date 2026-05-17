#!/usr/bin/env python3
"""Read stdin; print the last complete top-level JSON object or array."""
from __future__ import annotations

import json
import sys


def main() -> None:
    t = sys.stdin.read()
    depth = 0
    in_str = False
    esc = False
    chunk_start: int | None = None
    last_span: tuple[int, int] | None = None
    for i, ch in enumerate(t):
        if in_str:
            if esc:
                esc = False
            elif ch == "\\":
                esc = True
            elif ch == '"':
                in_str = False
            continue
        if ch == '"':
            in_str = True
            continue
        if ch in "{[":
            if depth == 0:
                chunk_start = i
            depth += 1
            continue
        if ch in "}]":
            depth -= 1
            if depth == 0 and chunk_start is not None:
                last_span = (chunk_start, i + 1)
            continue

    if not last_span:
        sys.stderr.write(t[:8000])
        raise SystemExit("No complete JSON object/array found in stdin")

    blob = t[last_span[0] : last_span[1]]
    json.loads(blob)
    sys.stdout.write(blob)


if __name__ == "__main__":
    main()
