#!/usr/bin/env python3
"""Build placeholders JSON for render_worldcup_bkn_vega_placeholders.py.

Reads Vega `vega catalog resources … --category table` JSON from stdin.

Each table resource yields worldcup.wc_<stem> in source_identifier; maps to key
<STEM_UPPER_WITH_UNDERSCORES>_RES_ID (matches worldcup-bkn-vega object_types).
"""
from __future__ import annotations

import json
import re
import sys

# Keep in sync with scripts/worldcup_dataset_stems.inc.sh
WORLD_CUP_STEMS = (
    "tournaments",
    "confederations",
    "teams",
    "players",
    "managers",
    "referees",
    "stadiums",
    "matches",
    "awards",
    "qualified_teams",
    "squads",
    "manager_appointments",
    "referee_appointments",
    "team_appearances",
    "player_appearances",
    "manager_appearances",
    "referee_appearances",
    "goals",
    "penalty_kicks",
    "bookings",
    "substitutions",
    "host_countries",
    "tournament_stages",
    "groups",
    "group_standings",
    "tournament_standings",
    "award_winners",
)


def _stem_from_sid(sid: str) -> str | None:
    sid = sid.strip()
    if not sid:
        return None
    tail = sid.rsplit(".", 1)[-1]
    if tail.startswith("wc_"):
        return tail[len("wc_") :]
    # Allow bare wc-less table names if discover ever emits them.
    return tail


def main() -> None:
    payload = json.load(sys.stdin)
    entries = payload.get("entries") or payload.get("resources") or []
    want = set(WORLD_CUP_STEMS)
    by_stem: dict[str, str] = {}

    for e in entries:
        sid = (
            str(e.get("source_identifier") or "")
            if isinstance(e.get("source_identifier"), str)
            else ""
        )
        stem = _stem_from_sid(sid)
        if not stem or stem not in want:
            continue
        rid = str(e.get("id") or "").strip()
        if not rid:
            continue
        if stem in by_stem:
            sys.stderr.write(
                f"Warning: duplicate resource for stem={stem}: {rid} overrides {by_stem[stem]}\n",
            )
        by_stem[stem] = rid

    missing = sorted(want - set(by_stem))
    if missing:
        sys.stderr.write(
            "Missing table resources for stems: {}\n(Run discover; check catalog id).\n".format(
                ", ".join(missing),
            ),
        )
        raise SystemExit(1)

    out: dict[str, str] = {}
    for stem in WORLD_CUP_STEMS:
        key = f'{stem.upper().replace("-", "_")}_RES_ID'
        out[key] = by_stem[stem]

    json.dump(out, sys.stdout, indent=2, ensure_ascii=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
