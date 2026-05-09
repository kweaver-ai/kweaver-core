#!/usr/bin/env python3
"""
kweaver `ds import-csv` maps every CSV column to VARCHAR(512); ~35+ columns then
exceeds MySQL's max inline row size (Error 1118) under utf8mb4.

Trim denormalised text columns from the two widest Fjelstul tables — the same facts
remain joinable via ids (tournaments, teams, stadiums, …).

 Idempotent: first run copies each target to `<name>.csv.wide_backup` before overwrite.
 Restore: mv data/matches.csv.wide_backup data/matches.csv (and team_appearances).
"""

from __future__ import annotations

import argparse
import csv
import shutil
from pathlib import Path


MATCHES_KEEP = (
    "key_id",
    "tournament_id",
    "match_id",
    "group_stage",
    "knockout_stage",
    "replayed",
    "replay",
    "match_date",
    "match_time",
    "stadium_id",
    "home_team_id",
    "home_team_code",
    "away_team_id",
    "away_team_code",
    "score",
    "home_team_score",
    "away_team_score",
    "home_team_score_margin",
    "away_team_score_margin",
    "extra_time",
    "penalty_shootout",
    "score_penalties",
    "home_team_score_penalties",
    "away_team_score_penalties",
    "home_team_win",
    "away_team_win",
    "draw",
)

TEAM_APP_KEEP = (
    "key_id",
    "tournament_id",
    "match_id",
    "group_stage",
    "knockout_stage",
    "replayed",
    "replay",
    "match_date",
    "match_time",
    "stadium_id",
    "team_id",
    "team_code",
    "opponent_id",
    "opponent_code",
    "home_team",
    "away_team",
    "goals_for",
    "goals_against",
    "goal_differential",
    "extra_time",
    "penalty_shootout",
    "penalties_for",
    "penalties_against",
    "win",
    "lose",
    "draw",
)


def slim_file(path: Path, keep: tuple[str, ...]) -> None:
    if not path.is_file():
        return
    bak = path.parent / f"{path.stem}.csv.wide_backup"
    if not bak.is_file():
        shutil.copy2(path, bak)

    src = bak
    with src.open(newline="", encoding="utf-8") as fp:
        r = csv.DictReader(fp)
        fieldnames_in = r.fieldnames or []
        missing = [c for c in keep if c not in fieldnames_in]
        if missing:
            raise SystemExit(f"{path}: missing columns {missing}")
        rows = [{k: row.get(k, "") for k in keep} for row in r]

    with path.open("w", newline="", encoding="utf-8") as fp:
        w = csv.DictWriter(fp, fieldnames=list(keep), lineterminator="\n")
        w.writeheader()
        w.writerows(rows)


def main() -> None:
    ap = argparse.ArgumentParser(description="Slim widest World Cup CSVs for MySQL import-csv.")
    ap.add_argument("--data-dir", type=Path, required=True)
    args = ap.parse_args()
    d = args.data_dir.resolve()
    slim_file(d / "matches.csv", MATCHES_KEEP)
    slim_file(d / "team_appearances.csv", TEAM_APP_KEEP)
    print("Slimmed matches.csv and team_appearances.csv for MySQL row-size limit (.wide_backup if first run).")


if __name__ == "__main__":
    main()
