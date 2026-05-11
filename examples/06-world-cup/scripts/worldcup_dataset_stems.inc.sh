# shellcheck shell=bash
# Source-only: CSV basename stems == logical MySQL tables wc_<stem> (27 total).
# Keep in sync with upstream https://github.com/jfjelstul/worldcup data-csv layout.
WORLD_CUP_DATASET_STEMS=(
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
