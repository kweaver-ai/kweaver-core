#!/usr/bin/env bash
# =============================================================================
# Download all 27 CSV datasets from the Fjelstul World Cup Database
# (https://github.com/jfjelstul/worldcup, CC-BY-SA 4.0).
#
# Total payload is a few MB. The data/ directory is gitignored so re-running
# this script is safe.
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="$SCRIPT_DIR/data"
REF="${WORLDCUP_REF:-master}"
BASE="https://raw.githubusercontent.com/jfjelstul/worldcup/${REF}/data-csv"

DATASETS=(
    tournaments
    confederations
    teams
    players
    managers
    referees
    stadiums
    matches
    awards
    qualified_teams
    squads
    manager_appointments
    referee_appointments
    team_appearances
    player_appearances
    manager_appearances
    referee_appearances
    goals
    penalty_kicks
    bookings
    substitutions
    host_countries
    tournament_stages
    groups
    group_standings
    tournament_standings
    award_winners
)

mkdir -p "$DATA_DIR"
echo "Downloading ${#DATASETS[@]} CSVs from jfjelstul/worldcup@${REF} → $DATA_DIR"

for ds in "${DATASETS[@]}"; do
    url="${BASE}/${ds}.csv"
    out="${DATA_DIR}/${ds}.csv"
    if [ -s "$out" ]; then
        printf '  %-26s (cached)\n' "$ds"
        continue
    fi
    if ! curl -fsSL "$url" -o "$out"; then
        echo "  FAIL: $ds  ($url)" >&2
        rm -f "$out"
        exit 1
    fi
    bytes=$(wc -c <"$out" | tr -d ' ')
    printf '  %-26s %s bytes\n' "$ds" "$bytes"
done

cat <<EOF

Done. Attribution (required by CC-BY-SA 4.0):
  © 2023 Joshua C. Fjelstul, Ph.D.
  https://github.com/jfjelstul/worldcup
  Licensed under CC-BY-SA 4.0 (https://creativecommons.org/licenses/by-sa/4.0/legalcode)
EOF
