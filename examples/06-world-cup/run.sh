#!/usr/bin/env bash
# =============================================================================
# 06-world-cup · KWeaver Core end-to-end (single script):
#
#   Step 1  Download CSVs      : jfjelstul/worldcup → data/ (skips cached files)
#   Step 2  Import to MySQL    : kweaver ds connect + ds import-csv --recreate
#   Step 3  Vega scan          : vega catalog create + discover --wait
#   Step 4  Render BKN         : map table Resources → render worldcup-bkn
#   Step 5  Push BKN           : kweaver bkn validate + push (build skipped by default)
#   Step 6  Upload toolbox     : kweaver toolbox import + publish (idempotent by box_name)
#   Step 7  Create Agent       : agent create --config <rendered tpl> + bind KN + publish
#
# Prerequisites:
#   - kweaver CLI logged in (`kweaver auth login`) and node `kweaver` resolvable
#     (avoid the broken `/usr/local/bin/kweaver` python stub).
#   - MySQL reachable from this host AND from the kweaver platform / Vega connectors.
#   - curl + python3 + jq
#
# Common usage:
#   ./run.sh                   # run all steps
#   ./run.sh --from 3          # rerun from Vega scan onward (CSVs already in MySQL)
#   ./run.sh --only 1          # only download CSVs
#   ./run.sh --dry-run         # print plan only
#   ./run.sh --no-publish      # create agent in private space (skip publish)
# =============================================================================
set -euo pipefail
[ "${WC_TRACE:-0}" = 1 ] && set -x || true

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BKN_ARCHIVE="$SCRIPT_DIR/worldcup-bkn.tar"
BKN_EXTRACT_DIR="$SCRIPT_DIR/.tmp/worldcup-bkn"
NETWORK_BKN="$BKN_EXTRACT_DIR/network.bkn"
RENDERED_DIR="$SCRIPT_DIR/.rendered-bkn-vega"
MAPPING_TMP="$SCRIPT_DIR/.vega-bkn-mapping.json"
AGENT_TEMPLATE="$SCRIPT_DIR/agent-worldcup.config.json"
VEGA_OPENAPI_SPEC="$SCRIPT_DIR/vega_sql_execute.openapi.json"
VEGA_TOOLBOX_NAME="${VEGA_TOOLBOX_NAME:-wc_vega_query}"
VEGA_TOOLBOX_SVC_URL="${VEGA_TOOLBOX_SVC_URL:-http://vega-backend-svc:13014}"
DATA_DIR="$SCRIPT_DIR/data"

usage() {
    cat <<'EOF'
Usage: ./run.sh [options]

Options:
  -h, --help          Show this help.
  --dry-run           Print the plan for the selected steps, no API calls.
  --from N            Start from step N (1..7). Default 1.
  --only N            Run only step N (1..7).
  --no-publish        Step 7: skip `agent publish` (agent stays private).
  --no-reuse          Step 7: ignore REUSE_AGENT_BY_NAME and always create.

Steps:
  1  Download CSVs   — fetch 27 CSV files from jfjelstul/worldcup (skips cached)
  2  Import MySQL    — kweaver ds connect + ds import-csv → wc_* tables
  3  Vega scan       — vega catalog create + discover --wait
  4  Render BKN      — map Resources → render worldcup-bkn
  5  Push BKN        — validate + push (resource-backed KN; build skipped by default)
  6  Upload toolbox  — toolbox import + publish (idempotent; DO_TOOLBOX=0 disables)
  7  Create Agent    — agent create --config + bind KN + (optional) publish

Env (see env.sample):
  WORLDCUP_REF                    Git ref for jfjelstul/worldcup; default master.
  SKIP_DOWNLOAD=0                 Set 1 to skip step 1 even if CSVs are missing.
  DB_HOST / DB_PORT / DB_NAME     MySQL coordinates (required for steps 2-3).
  DB_USER / DB_PASS               MySQL account for kweaver ds connect.
  DS_ID                           Existing datasource id; skip ds connect in step 2.
  SKIP_IMPORT=0                   Set 1 to skip step 2 (MySQL already populated).
  VEGA_CATALOG_NAME               Catalog display name (required for step 3).
  VEGA_CATALOG_ID                 Existing catalog UUID (skip create in step 3).
  VEGA_SKIP_CREATE=1              Skip catalog create (must set VEGA_CATALOG_ID).
  VEGA_MYSQL_*                    Override DB_* for connector when Vega's network
                                  view differs (HOST/PORT/USER/PASS/DATABASES).
  BKN_PUSH_BRANCH=main            Branch for `kweaver bkn push`.
  DO_INDEX=1                      Build vector indexes on 7 entity tables after push (default). Set =0 to skip.
  EMBEDDING_MODEL_NAME=text-embedding-v4-cn   Model_name registered via mf-model-manager
                                  (resolved to model_id at runtime). Unset / not found → keyword only.
  DO_TOOLBOX=1                    Step 6: set 0 to skip toolbox import + publish.
  FORCE_TOOLBOX_REIMPORT=0        Step 6: set 1 to delete the existing same-name toolbox
                                  and re-import (use after editing the .adp template).
  AGENT_NAME                      Default: 世界杯数据分析助手.
  AGENT_PROFILE                   Short description (<=500 chars).
  REUSE_AGENT_BY_NAME=true        Default: reuse same-name agent (re-bind KN, no recreate). Set =false to always create.
  AGENT_LLM_ID                    Optional: override llm id in agent template.
  KWEAVER_BASE_URL                Optional: pin a specific platform; default uses ~/.kweaver.
EOF
}

FROM=1
ONLY=""
DRY_RUN=0
DO_PUBLISH=1
FORCE_NO_REUSE=0
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        --dry-run) DRY_RUN=1; shift ;;
        --from) FROM="${2:?--from needs a number 1..7}"; shift 2 ;;
        --only) ONLY="${2:?--only needs a number 1..7}"; shift 2 ;;
        --no-publish) DO_PUBLISH=0; shift ;;
        --no-reuse) FORCE_NO_REUSE=1; shift ;;
        *) echo "Unknown arg: $1" >&2; usage >&2; exit 2 ;;
    esac
done

case "$FROM" in 1|2|3|4|5|6|7) ;; *) echo "--from must be 1..7" >&2; exit 2 ;; esac
if [ -n "$ONLY" ]; then case "$ONLY" in 1|2|3|4|5|6|7) ;; *) echo "--only must be 1..7" >&2; exit 2 ;; esac; fi

set -a
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"
set +a

if [ -n "${KWEAVER_BASE_URL:-}" ]; then
    KWEAV=(kweaver --base-url "${KWEAVER_BASE_URL}")
else
    KWEAV=(kweaver)
fi

require_jq() {
    command -v jq >/dev/null 2>&1 || { echo "Error: jq is required (brew install jq)" >&2; exit 1; }
}

# Strip CLI warnings preceding a final JSON blob.
_extract_cli_json() {
    python3 "$SCRIPT_DIR/scripts/extract_trailing_json.py"
}

kn_id_from_rendered() {
    awk '/^id:/ {print $2; exit}' "$RENDERED_DIR/network.bkn" 2>/dev/null || true
}

kn_id_from_template() {
    awk '/^id:/ {print $2; exit}' "$NETWORK_BKN" 2>/dev/null || true
}

# ─── Step 1: Download CSVs ──────────────────────────────────────────────────
step_1_download() {
    echo "=== [1/7] Download CSVs ===" >&2

    if [ "${SKIP_DOWNLOAD:-0}" = 1 ]; then
        echo "  skipped (SKIP_DOWNLOAD=1)" >&2
        return 0
    fi

    # shellcheck source=scripts/worldcup_dataset_stems.inc.sh
    source "$SCRIPT_DIR/scripts/worldcup_dataset_stems.inc.sh"
    local ref="${WORLDCUP_REF:-master}"
    local base="https://raw.githubusercontent.com/jfjelstul/worldcup/${ref}/data-csv"
    local missing=0

    for stem in "${WORLD_CUP_DATASET_STEMS[@]}"; do
        [ -s "$DATA_DIR/${stem}.csv" ] || { missing=1; break; }
    done

    if [ "$missing" = 0 ]; then
        echo "  all ${#WORLD_CUP_DATASET_STEMS[@]} CSVs already cached in $DATA_DIR — skipping download" >&2
        return 0
    fi

    if [ "$DRY_RUN" = 1 ]; then
        echo "  plan: download ${#WORLD_CUP_DATASET_STEMS[@]} CSVs from jfjelstul/worldcup@${ref} → $DATA_DIR" >&2
        return 0
    fi

    mkdir -p "$DATA_DIR"
    echo "  Downloading ${#WORLD_CUP_DATASET_STEMS[@]} CSVs from jfjelstul/worldcup@${ref}" >&2
    for stem in "${WORLD_CUP_DATASET_STEMS[@]}"; do
        local out="$DATA_DIR/${stem}.csv"
        if [ -s "$out" ]; then
            printf '    %-26s (cached)\n' "$stem" >&2
            continue
        fi
        if ! curl -fsSL "${base}/${stem}.csv" -o "$out"; then
            echo "  FAIL: $stem  (${base}/${stem}.csv)" >&2
            rm -f "$out"
            exit 1
        fi
        printf '    %-26s %s bytes\n' "$stem" "$(wc -c <"$out" | tr -d ' ')" >&2
    done
    echo "  Done. CC-BY-SA 4.0 © 2023 Joshua C. Fjelstul, Ph.D." >&2
}

# ─── Step 2: Import to MySQL ─────────────────────────────────────────────────
_write_ds_id_to_env() {
    local ds_id="$1"
    [ "${SKIP_WRITE_ENV:-0}" = 1 ] && return 0
    _WC_PATCH_DS_ID="$ds_id" ENV_FILE="$SCRIPT_DIR/.env" python3 - <<'PY'
import os, pathlib, re
ds_id    = os.environ["_WC_PATCH_DS_ID"]
env_file = pathlib.Path(os.environ["ENV_FILE"])
nl = "\n"
if env_file.is_file():
    text = env_file.read_text(encoding="utf-8", errors="replace")
    if not text.endswith("\n"): text += nl
    pat = re.compile(r"^\s*DS_ID=")
    out, found = [], False
    for line in text.splitlines(keepends=True):
        if pat.match(line):
            if not found:
                out.append(f"DS_ID={ds_id}{nl}")
                found = True
        else:
            out.append(line)
    if not found: out.append(f"DS_ID={ds_id}{nl}")
    env_file.write_text("".join(out), encoding="utf-8")
else:
    env_file.write_text(f"DS_ID={ds_id}{nl}", encoding="utf-8")
print(f"  .env DS_ID={ds_id}", flush=True)
PY
}

step_2_import() {
    echo "=== [2/7] Import to MySQL ===" >&2

    if [ "${SKIP_IMPORT:-0}" = 1 ]; then
        echo "  skipped (SKIP_IMPORT=1)" >&2
        return 0
    fi

    local ds_id="${DS_ID:-}"

    if [ "$DRY_RUN" = 1 ]; then
        if [ -n "$ds_id" ]; then
            echo "  plan: reuse DS_ID=$ds_id" >&2
        else
            echo "  plan: kweaver ds connect mysql ${DB_HOST:-<DB_HOST>} ${DB_PORT:-3306} ${DB_NAME:-<DB_NAME>} --account <DB_USER> --name worldcup_ds_<ts>" >&2
        fi
        echo "  plan: kweaver ds import-csv <DS_ID> --files data/*.csv --table-prefix wc_ --recreate" >&2
        return 0
    fi

    # Ensure CSVs exist.
    local csv_count
    csv_count=$(find "$DATA_DIR" -maxdepth 1 -name "*.csv" 2>/dev/null | wc -l | tr -d ' ')
    [ "${csv_count:-0}" -lt 27 ] && {
        echo "Error: only $csv_count CSV(s) found in $DATA_DIR — run step 1 first." >&2
        exit 1
    }

    if [ -z "$ds_id" ]; then
        local ds_name="worldcup_ds_$(date +%s)"
        echo "  Connecting datasource: $ds_name" >&2
        local connect_out
        connect_out="$("${KWEAV[@]}" ds connect mysql \
            "${DB_HOST:?Set DB_HOST in .env}" \
            "${DB_PORT:-3306}" \
            "${DB_NAME:?Set DB_NAME in .env}" \
            --account "${DB_USER:?Set DB_USER in .env}" \
            --password "${DB_PASS:?Set DB_PASS in .env}" \
            --name "$ds_name" 2>&1)"
        ds_id="$(printf '%s' "$connect_out" | _extract_cli_json | \
            python3 -c 'import sys,json; d=json.load(sys.stdin); r=d[0] if isinstance(d,list) else d; print(r.get("datasource_id") or r.get("id",""))' 2>/dev/null)"
        [ -z "$ds_id" ] && { echo "Error: kweaver ds connect returned no id." >&2; echo "$connect_out" >&2; exit 1; }
        echo "  DS_ID=$ds_id" >&2
        DS_ID="$ds_id"; export DS_ID
        _write_ds_id_to_env "$ds_id"
    else
        echo "  reusing DS_ID=$ds_id" >&2
    fi

    # Auto-skip if all 27 wc_* tables already exist in MySQL.
    if [ "${FORCE_IMPORT:-0}" != 1 ]; then
        local n_tables
        n_tables="$(_count_wc_tables)"
        if [ "${n_tables:-0}" -ge 27 ]; then
            echo "  skip import — $n_tables wc_* tables already in $DB_NAME (FORCE_IMPORT=1 to redo)" >&2
            return 0
        fi
    fi

    # Pre-create wide tables (wc_matches, wc_team_appearances) with VARCHAR(255)
    # to dodge MySQL Error 1118 — import-csv hardcodes VARCHAR(512) which exceeds
    # the 65,535-byte row size limit (strict mode) for these 36/37-column tables.
    _pre_create_wide_tables

    echo "  Importing ${csv_count} CSVs → wc_* tables (--recreate) …" >&2
    "${KWEAV[@]}" ds import-csv "$ds_id" \
        --files "$DATA_DIR/*.csv" \
        --table-prefix wc_ \
        --recreate
    echo "  Import done." >&2
}

_count_wc_tables() {
    command -v mysql >/dev/null 2>&1 || { echo 0; return; }
    MYSQL_PWD="${DB_PASS}" mysql \
        -h "${DB_HOST}" -P "${DB_PORT:-3306}" -u "${DB_USER}" -N -B "${DB_NAME}" \
        -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${DB_NAME}' AND table_name LIKE 'wc\_%';" \
        2>/dev/null | tr -d ' \n' || echo 0
}

_pre_create_wide_tables() {
    command -v mysql >/dev/null 2>&1 || {
        echo "  warn: mysql client not found — skipping wide-table pre-create; import may hit Error 1118." >&2
        return 0
    }
    echo "  Pre-creating wide tables (wc_matches, wc_team_appearances) with VARCHAR(255) …" >&2
    MYSQL_PWD="${DB_PASS}" mysql \
        -h "${DB_HOST}" -P "${DB_PORT:-3306}" -u "${DB_USER}" "${DB_NAME}" \
        --default-character-set=utf8mb4 <<'SQL' 2>&1 | grep -v -E '^Warning|using a password' >&2 || true
DROP TABLE IF EXISTS wc_matches;
CREATE TABLE wc_matches (
  key_id VARCHAR(255), tournament_id VARCHAR(255), tournament_name VARCHAR(255),
  match_id VARCHAR(255), match_name VARCHAR(255), stage_name VARCHAR(255),
  group_name VARCHAR(255), group_stage VARCHAR(255), knockout_stage VARCHAR(255),
  replayed VARCHAR(255), replay VARCHAR(255), match_date VARCHAR(255),
  match_time VARCHAR(255), stadium_id VARCHAR(255), stadium_name VARCHAR(255),
  city_name VARCHAR(255), country_name VARCHAR(255), home_team_id VARCHAR(255),
  home_team_name VARCHAR(255), home_team_code VARCHAR(255), away_team_id VARCHAR(255),
  away_team_name VARCHAR(255), away_team_code VARCHAR(255), score VARCHAR(255),
  home_team_score VARCHAR(255), away_team_score VARCHAR(255),
  home_team_score_margin VARCHAR(255), away_team_score_margin VARCHAR(255),
  extra_time VARCHAR(255), penalty_shootout VARCHAR(255), score_penalties VARCHAR(255),
  home_team_score_penalties VARCHAR(255), away_team_score_penalties VARCHAR(255),
  result VARCHAR(255), home_team_win VARCHAR(255), away_team_win VARCHAR(255),
  draw VARCHAR(255)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS wc_team_appearances;
CREATE TABLE wc_team_appearances (
  key_id VARCHAR(255), tournament_id VARCHAR(255), tournament_name VARCHAR(255),
  match_id VARCHAR(255), match_name VARCHAR(255), stage_name VARCHAR(255),
  group_name VARCHAR(255), group_stage VARCHAR(255), knockout_stage VARCHAR(255),
  replayed VARCHAR(255), replay VARCHAR(255), match_date VARCHAR(255),
  match_time VARCHAR(255), stadium_id VARCHAR(255), stadium_name VARCHAR(255),
  city_name VARCHAR(255), country_name VARCHAR(255), team_id VARCHAR(255),
  team_name VARCHAR(255), team_code VARCHAR(255), opponent_id VARCHAR(255),
  opponent_name VARCHAR(255), opponent_code VARCHAR(255), home_team VARCHAR(255),
  away_team VARCHAR(255), goals_for VARCHAR(255), goals_against VARCHAR(255),
  goal_differential VARCHAR(255), extra_time VARCHAR(255), penalty_shootout VARCHAR(255),
  penalties_for VARCHAR(255), penalties_against VARCHAR(255), result VARCHAR(255),
  win VARCHAR(255), lose VARCHAR(255), draw VARCHAR(255)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4;
SQL
}

# ─── Step 3: Vega scan ──────────────────────────────────────────────────────
build_connector_config() {
    local host port user pass dbs port_num
    host="${VEGA_MYSQL_HOST:-${DB_HOST:?Set DB_HOST in .env}}"
    port="${VEGA_MYSQL_PORT:-${DB_PORT:-3306}}"
    user="${VEGA_MYSQL_USER:-${DB_USER:?Set DB_USER in .env}}"
    pass="${VEGA_MYSQL_PASS:-${DB_PASS:?Set DB_PASS in .env}}"
    dbs="${VEGA_MYSQL_DATABASES:-}"
    if [ -z "$dbs" ]; then
        dbs="$(jq -cn --arg d "${DB_NAME:?Set DB_NAME in .env}" '[$d]')"
    fi
    port_num=$((10#${port}))
    jq -cn \
        --arg host "$host" \
        --argjson port "$port_num" \
        --arg username "$user" \
        --arg password "$pass" \
        --argjson databases "$dbs" \
        '{host:$host,port:$port,username:$username,password:$password,databases:$databases}'
}

resolve_catalog_id_by_name() {
    local name="$1" raw
    raw="$("${KWEAV[@]}" vega catalog list --limit 200 2>&1)"
    printf '%s' "$raw" | _extract_cli_json | jq -r --arg n "$name" '.entries[]? | select(.name == $n) | .id' | head -1
}

step_3_vega_scan() {
    echo "=== [3/7] Vega scan ===" >&2
    local catalog_id="${VEGA_CATALOG_ID:-}"

    if [ "${VEGA_SKIP_CREATE:-0}" = "1" ]; then
        [ -z "$catalog_id" ] && { echo "VEGA_SKIP_CREATE=1 requires VEGA_CATALOG_ID" >&2; exit 2; }
    fi

    if [ "$DRY_RUN" = 1 ]; then
        if [ "${VEGA_SKIP_CREATE:-0}" = "1" ]; then
            echo "  plan: skip create, reuse VEGA_CATALOG_ID=$catalog_id" >&2
        else
            echo "  plan: vega catalog create --name \"${VEGA_CATALOG_NAME:?Set VEGA_CATALOG_NAME in .env}\"" >&2
        fi
        echo "  plan: vega catalog discover <id> --wait" >&2
        return 0
    fi

    if [ "${VEGA_SKIP_CREATE:-0}" != "1" ] && [ -z "$catalog_id" ]; then
        local name="${VEGA_CATALOG_NAME:?Set VEGA_CATALOG_NAME in .env}"
        local conn create_out
        conn="$(build_connector_config)"
        echo "  Creating catalog: $name" >&2
        if ! create_out="$("${KWEAV[@]}" vega catalog create --name "$name" --connector-type mysql --connector-config "$conn" 2>&1)"; then
            echo "$create_out" >&2
            catalog_id="$(resolve_catalog_id_by_name "$name")"
            [ -z "$catalog_id" ] && { echo "Error: vega catalog create failed and name not found" >&2; exit 1; }
            echo "  Reusing existing catalog by name: $catalog_id" >&2
        else
            catalog_id="$(printf '%s' "$create_out" | _extract_cli_json | jq -r '.id // .data.id // empty' 2>/dev/null | head -1)"
            [ -z "$catalog_id" ] && catalog_id="$(resolve_catalog_id_by_name "$name")"
        fi
    fi
    [ -z "$catalog_id" ] && { echo "Error: empty catalog id" >&2; exit 1; }

    echo "  catalog_id=$catalog_id" >&2

    # Auto-skip discover if catalog already has all 27 table resources.
    if [ "${FORCE_DISCOVER:-0}" != 1 ]; then
        local n_resources
        n_resources="$("${KWEAV[@]}" vega resource list --catalog-id "$catalog_id" --category table --limit 50 2>/dev/null \
            | _extract_cli_json | jq '.entries | length' 2>/dev/null || echo 0)"
        if [ "${n_resources:-0}" -ge 27 ]; then
            echo "  skip discover — $n_resources table resources already present (FORCE_DISCOVER=1 to redo)" >&2
            VEGA_CATALOG_ID="$catalog_id"; export VEGA_CATALOG_ID
            return 0
        fi
    fi

    echo "  Running discover --wait …" >&2
    if ! "${KWEAV[@]}" vega catalog discover "$catalog_id" --wait; then
        echo "Discover failed — re-run later, or set VEGA_CATALOG_ID and --from 4." >&2
        exit 1
    fi

    VEGA_CATALOG_ID="$catalog_id"
    export VEGA_CATALOG_ID
}

# ─── Step 4: Render BKN ─────────────────────────────────────────────────────
resolve_catalog_id_or_die() {
    if [ -n "${VEGA_CATALOG_ID:-}" ]; then
        printf '%s' "$VEGA_CATALOG_ID"; return
    fi
    local name="${VEGA_CATALOG_NAME:?Set VEGA_CATALOG_NAME or VEGA_CATALOG_ID in .env}"
    local id
    id="$(resolve_catalog_id_by_name "$name")"
    [ -z "$id" ] && { echo "Error: cannot resolve catalog id for name $name" >&2; exit 1; }
    printf '%s' "$id"
}

step_4_render_bkn() {
    echo "=== [4/7] Render BKN ===" >&2

    if [ "$DRY_RUN" = 1 ]; then
        local hint="${VEGA_CATALOG_ID:-${VEGA_CATALOG_NAME:-<unresolved>}}"
        echo "  plan: resolve catalog (id|name: $hint)" >&2
        echo "  plan: vega catalog resources <id> --category table --limit 500" >&2
        echo "  plan: map_vega_table_resources.py → $MAPPING_TMP" >&2
        echo "  plan: render_worldcup_bkn_vega_placeholders.py → $RENDERED_DIR" >&2
        return 0
    fi

    local catalog_id
    catalog_id="$(resolve_catalog_id_or_die)"
    echo "  catalog_id=$catalog_id" >&2

    local raw body nres
    raw="$("${KWEAV[@]}" vega catalog resources "$catalog_id" --category table --limit 500 2>&1)"
    body="$(printf '%s' "$raw" | _extract_cli_json)" || { echo "Error parsing vega catalog resources output." >&2; exit 1; }
    nres="$(printf '%s' "$body" | jq '.entries | length')"
    echo "  table resources returned: $nres" >&2

    if ! printf '%s' "$body" | python3 "$SCRIPT_DIR/scripts/map_vega_table_resources.py" >"$MAPPING_TMP"; then
        echo "Error: could not map all 27 wc_* stems (see stderr)." >&2
        exit 1
    fi
    local keys
    keys="$(jq 'length' "$MAPPING_TMP")"
    echo "  mapped placeholders: $keys (need 27)" >&2
    [ "${keys:-0}" -lt 27 ] && { echo "Error: incomplete mapping (<27)." >&2; exit 1; }

    _extract_bkn_archive
    python3 "$SCRIPT_DIR/scripts/render_worldcup_bkn_vega_placeholders.py" \
        --mapping "$MAPPING_TMP" \
        --src "$BKN_EXTRACT_DIR" \
        --dst "$RENDERED_DIR"
    echo "  rendered → $RENDERED_DIR" >&2
}

_extract_bkn_archive() {
    [ -f "$BKN_ARCHIVE" ] || {
        echo "Error: BKN archive not found at $BKN_ARCHIVE." >&2
        exit 1
    }
    if [ -f "$BKN_EXTRACT_DIR/network.bkn" ] && \
       [ "$BKN_EXTRACT_DIR/network.bkn" -nt "$BKN_ARCHIVE" ]; then
        echo "  reusing extracted BKN tree at $BKN_EXTRACT_DIR" >&2
        return 0
    fi
    rm -rf "$BKN_EXTRACT_DIR"
    mkdir -p "$(dirname "$BKN_EXTRACT_DIR")"
    tar xf "$BKN_ARCHIVE" -C "$(dirname "$BKN_EXTRACT_DIR")"
    [ -f "$BKN_EXTRACT_DIR/network.bkn" ] || {
        echo "Error: extracted tree missing network.bkn (expected $BKN_EXTRACT_DIR/network.bkn)." >&2
        exit 1
    }
    echo "  extracted BKN tree → $BKN_EXTRACT_DIR" >&2
}

# ─── Step 5: Push BKN ───────────────────────────────────────────────────────
step_5_push_bkn() {
    echo "=== [5/7] Push BKN ===" >&2

    if [ "$DRY_RUN" = 1 ]; then
        echo "  plan: kweaver bkn validate $RENDERED_DIR" >&2
        echo "  plan: kweaver bkn push $RENDERED_DIR --branch ${BKN_PUSH_BRANCH:-main}" >&2
        echo "  plan: create vega vector build-tasks for 7 entity resources (embedding_fields)" >&2
        return 0
    fi

    [ -d "$RENDERED_DIR" ] || { echo "Error: $RENDERED_DIR missing — run step 4 first." >&2; exit 1; }
    local kn_id
    kn_id="$(kn_id_from_rendered)"
    [ -z "$kn_id" ] && { echo "Error: cannot read id from $RENDERED_DIR/network.bkn" >&2; exit 1; }

    mkdir -p "$SCRIPT_DIR/.tmp"
    export TMPDIR="${TMPDIR:-$SCRIPT_DIR/.tmp}"

    # Auto-skip push if KN already exists (re-push currently triggers an OpenSearch
    # mapping conflict on the BKN backend). Set FORCE_PUSH=1 to push anyway.
    local skip_push=0
    if [ "${FORCE_PUSH:-0}" != 1 ] && "${KWEAV[@]}" bkn get "$kn_id" >/dev/null 2>&1; then
        echo "  skip push — KN '$kn_id' already exists (FORCE_PUSH=1 to redo)" >&2
        skip_push=1
    fi

    if [ "$skip_push" != 1 ]; then
        "${KWEAV[@]}" bkn validate "$RENDERED_DIR"
        echo "  push → $kn_id" >&2
        "${KWEAV[@]}" bkn push "$RENDERED_DIR" --branch "${BKN_PUSH_BRANCH:-main}"
    fi

    # Build vega resource-level vector indexes (OpenSearch datasets) for the 7
    # high-value entity tables (awards/tournaments/teams/stadiums/managers/referees/players)
    # so the LLM can do fuzzy / cross-language name matching at query time.
    if [ "${DO_INDEX:-1}" = 1 ]; then
        _build_vega_indexes
    fi
    echo "  KN=$kn_id" >&2
}

# Lookup embedding model_id by name (registered via mf-model-manager / onboard.sh)
_lookup_embedding_model_id() {
    local model_name="${1:-text-embedding-v4-cn}"
    "${KWEAV[@]}" model small list 2>/dev/null | _extract_cli_json | \
        jq -r --arg n "$model_name" '
            (.data // .entries // [])[] |
            select(.model_name == $n and .model_type == "embedding") |
            .model_id' 2>/dev/null | head -1
}

# Tables that benefit from vector embedding (LLM gets fuzzy/cross-language name match).
# Format: "table_name:embedding_fields_csv"
_vector_index_tables() {
    cat <<'EOF'
awards:award_name
tournaments:tournament_name,host_country
teams:team_name
stadiums:stadium_name,city_name
managers:family_name,given_name
referees:family_name,given_name,country_name
players:family_name,given_name
EOF
}

_build_vega_indexes() {
    [ -f "$MAPPING_TMP" ] || { echo "  warn: $MAPPING_TMP missing — step 4 must run first; skipping indexes." >&2; return 0; }

    local emodel
    emodel="$(_lookup_embedding_model_id "${EMBEDDING_MODEL_NAME:-text-embedding-v4-cn}")"
    if [ -z "$emodel" ]; then
        echo "  warn: embedding model '${EMBEDDING_MODEL_NAME:-text-embedding-v4-cn}' not registered;" >&2
        echo "        vector tables will be built keyword-only (no embedding)." >&2
    else
        echo "  embedding model_id=$emodel" >&2
    fi

    # Build a temp tsv: table\tresource_id\tembedding_fields(or blank)
    local plan; plan="$(mktemp -t wc_index_plan.XXXXXX.tsv)"
    EMODEL="$emodel" MAPPING="$MAPPING_TMP" python3 - >"$plan" <<'PY'
import json, os
with open(os.environ["MAPPING"]) as f:
    mapping = json.load(f)
# mapping looks like {"TOURNAMENTS_RES_ID": "d831...", ...}; convert key → table name
table_to_rid = {}
SUFFIX = "_res_id"
for placeholder, rid in mapping.items():
    tbl = placeholder.lower()
    if tbl.endswith(SUFFIX):
        tbl = tbl[:-len(SUFFIX)]
    table_to_rid[tbl] = rid
vec = {
    "awards": "award_name",
    "tournaments": "tournament_name,host_country",
    "teams": "team_name",
    "stadiums": "stadium_name,city_name",
    "managers": "family_name,given_name",
    "referees": "family_name,given_name,country_name",
    "players": "family_name,given_name",
}
# 0.7.0 vega-backend has a cursor-advance bug in batch sync: tables with rows > 2*batch_size
# loop forever inserting the same batch. List them so we can warn but still attempt.
runaway_risk = {
    "bookings", "goals", "substitutions", "player_appearances",
    "players", "squads", "manager_appearances", "team_appearances",
}
for tbl, rid in sorted(table_to_rid.items()):
    if tbl not in vec:
        continue
    ef = vec[tbl]
    flag = "RISK" if tbl in runaway_risk else "OK"
    print(f"{tbl}\t{rid}\t{ef}\t{flag}")
PY

    # Pre-fetch ALL build tasks once (0.7.0's list endpoint ignores ?resource_id=)
    local all_tasks_json
    all_tasks_json="$("${KWEAV[@]}" call "/api/vega-backend/v1/resources/buildtask?limit=500" 2>/dev/null | _extract_cli_json 2>/dev/null || echo '{"entries":[]}')"
    [ -z "$all_tasks_json" ] && all_tasks_json='{"entries":[]}'

    # For each resource, check if a build task already exists; if not, create + start one.
    # Tolerate failures (warn-not-fail), since the 0.7.0 batch-sync cursor bug can stall
    # large tables. Agent can still answer via vega_sql_execute (step 7's tool).
    local created=0 reused=0 skipped=0
    while IFS=$'\t' read -r tbl rid ef risk; do
        [ -z "$rid" ] && continue
        local existing
        existing="$(printf '%s' "$all_tasks_json" | jq -r --arg r "$rid" \
            '(.entries // [])[] | select(.resource_id == $r) | "\(.id)\t\(.status)"' 2>/dev/null | head -1 || true)"
        if [ -n "$existing" ]; then
            local existing_status="${existing#*$'\t'}"
            case "$existing_status" in
                completed|running)
                    printf "  %-25s reuse (status=%s)\n" "$tbl" "$existing_status" >&2
                    reused=$((reused+1))
                    continue ;;
            esac
        fi

        # Build body
        local body
        if [ -n "$ef" ] && [ -n "$emodel" ]; then
            body="$(jq -cn --arg ef "$ef" --arg em "$emodel" \
                '{mode:"batch",build_key_fields:"key_id",embedding_fields:$ef,embedding_model:$em,model_dimensions:1024}')"
        else
            body='{"mode":"batch","build_key_fields":"key_id"}'
        fi

        local tid="" create_raw
        create_raw="$("${KWEAV[@]}" call "/api/vega-backend/v1/resources/buildtask/$rid" -X POST -d "$body" 2>/dev/null || true)"
        if [ -n "$create_raw" ]; then
            tid="$(printf '%s' "$create_raw" | _extract_cli_json 2>/dev/null | jq -r '.task_id // empty' 2>/dev/null | head -1 || true)"
        fi
        if [ -z "$tid" ]; then
            # Likely "构建任务已存在" (existing stopped/failed task) — keep moving.
            printf "  %-25s ⊘ create_skipped (existing or API error)\n" "$tbl" >&2
            skipped=$((skipped+1))
            continue
        fi

        # Start the task
        "${KWEAV[@]}" call "/api/vega-backend/v1/resources/buildtask/$rid/$tid/status" \
            -X PUT -d '{"status":"running","execute_type":"full"}' >/dev/null 2>&1 || true
        local kind="keyword"; [ -n "$ef" ] && [ -n "$emodel" ] && kind="vector"
        local warn=""; [ "$risk" = "RISK" ] && warn="  (⚠ may loop on 0.7.0 — agent fallback to SQL)"
        printf "  %-25s create+start (%s)%s\n" "$tbl" "$kind" "$warn" >&2
        created=$((created+1))
    done <"$plan"
    rm -f "$plan"

    echo "  indexes: $created created, $reused reused, $skipped skipped (DO_INDEX=0 to disable; FORCE_INDEX=1 to recreate)" >&2
}

# ─── Step 6: Upload toolbox (OpenAPI) ──────────────────────────────────────
_write_box_id_to_env() {
    local box_id="$1" tool_id="${2:-}"
    [ "${SKIP_WRITE_ENV:-0}" = 1 ] && return 0
    _WC_PATCH_BOX_ID="$box_id" _WC_PATCH_TOOL_ID="$tool_id" ENV_FILE="$SCRIPT_DIR/.env" python3 - <<'PY'
import os, pathlib, re
box_id   = os.environ["_WC_PATCH_BOX_ID"]
tool_id  = os.environ.get("_WC_PATCH_TOOL_ID", "")
env_file = pathlib.Path(os.environ["ENV_FILE"])
nl = "\n"
updates = {"TOOLBOX_BOX_ID": box_id}
if tool_id: updates["VEGA_TOOL_ID"] = tool_id
text = env_file.read_text(encoding="utf-8", errors="replace") if env_file.is_file() else ""
if text and not text.endswith("\n"): text += nl
out, seen = [], set()
for line in text.splitlines(keepends=True):
    matched = False
    for k, v in updates.items():
        if re.match(rf"^\s*{re.escape(k)}=", line):
            if k not in seen:
                out.append(f"{k}={v}{nl}")
                seen.add(k)
            matched = True
            break
    if not matched: out.append(line)
for k, v in updates.items():
    if k not in seen: out.append(f"{k}={v}{nl}")
env_file.write_text("".join(out), encoding="utf-8")
print(f"  .env TOOLBOX_BOX_ID={box_id}" + (f"  VEGA_TOOL_ID={tool_id}" if tool_id else ""), flush=True)
PY
}

find_toolbox_id_by_name() {
    local name="$1" raw
    raw="$("${KWEAV[@]}" toolbox list --keyword "$name" --limit 50 2>/dev/null || true)"
    printf '%s' "$raw" | _extract_cli_json | jq -r --arg n "$name" '
        (.entries // .data // .items // [])
        | if type == "array" then . else [] end
        | map(select((.box_name // .name) == $n))
        | sort_by(.updated_at // .created_at // 0) | reverse
        | (.[0].box_id // .[0].id // empty)
    ' 2>/dev/null | head -1
}

step_6_toolbox() {
    echo "=== [6/7] Upload toolbox ===" >&2
    if [ "${DO_TOOLBOX:-1}" != 1 ]; then
        echo "  skipped (DO_TOOLBOX=0)" >&2
        return 0
    fi
    [ -f "$VEGA_OPENAPI_SPEC" ] || { echo "Error: $VEGA_OPENAPI_SPEC missing" >&2; exit 1; }

    if [ "$DRY_RUN" = 1 ]; then
        echo "  plan: kweaver toolbox create --name $VEGA_TOOLBOX_NAME (reuse if found)" >&2
        echo "  plan: kweaver tool upload --toolbox <box-id> $VEGA_OPENAPI_SPEC" >&2
        echo "  plan: kweaver tool enable + toolbox publish" >&2
        return 0
    fi

    # NOTE: this example uses `kweaver tool upload` (OpenAPI parser) instead of
    # `kweaver toolbox import` (.adp via impex). On platform 0.7.0 the impex path
    # has a write-path bug (api_spec stored as null); fix is in 0.8.0 commit
    # e4aac398. The OpenAPI parser path is unaffected.

    local box_id
    box_id="$(find_toolbox_id_by_name "$VEGA_TOOLBOX_NAME")"

    if [ "${FORCE_TOOLBOX_REIMPORT:-0}" = 1 ] && [ -n "$box_id" ]; then
        echo "  FORCE_TOOLBOX_REIMPORT=1 → deleting existing $box_id" >&2
        "${KWEAV[@]}" toolbox delete "$box_id" -y || {
            echo "  warn: toolbox delete failed (likely 0.7.0 batch-delete bug); continuing — will reuse." >&2
        }
        box_id="$(find_toolbox_id_by_name "$VEGA_TOOLBOX_NAME")"
    fi

    if [ -z "$box_id" ]; then
        echo "  Creating toolbox $VEGA_TOOLBOX_NAME → $VEGA_TOOLBOX_SVC_URL" >&2
        local create_out
        create_out="$("${KWEAV[@]}" toolbox create \
            --name "$VEGA_TOOLBOX_NAME" \
            --service-url "$VEGA_TOOLBOX_SVC_URL" \
            --description "Vega backend SQL execute (worldcup example)" 2>&1)" || {
            echo "$create_out" >&2
            echo "Error: toolbox create failed." >&2
            exit 1
        }
        box_id="$(printf '%s' "$create_out" | _extract_cli_json | jq -r '.box_id // .id // .data.box_id // empty' 2>/dev/null | head -1)"
        [ -z "$box_id" ] && { echo "Error: could not resolve box_id from create output:" >&2; echo "$create_out" >&2; exit 1; }
        echo "  created box_id=$box_id" >&2
    else
        echo "  reusing existing toolbox $box_id" >&2
    fi

    # Find existing vega_sql_execute tool, or upload it
    local tool_id
    tool_id="$("${KWEAV[@]}" tool list --toolbox "$box_id" 2>/dev/null | _extract_cli_json | \
        jq -r '.tools[]? | select(.name == "vega_sql_execute") | .tool_id' 2>/dev/null | head -1)"

    if [ -z "$tool_id" ]; then
        echo "  uploading $VEGA_OPENAPI_SPEC" >&2
        local upload_out
        upload_out="$("${KWEAV[@]}" tool upload --toolbox "$box_id" "$VEGA_OPENAPI_SPEC" 2>&1)" || {
            echo "$upload_out" >&2
            echo "Error: tool upload failed." >&2
            exit 1
        }
        tool_id="$(printf '%s' "$upload_out" | _extract_cli_json | jq -r '.success_ids[0] // empty' 2>/dev/null | head -1)"
        [ -z "$tool_id" ] && { echo "Error: could not resolve tool_id from upload output:" >&2; echo "$upload_out" >&2; exit 1; }
        echo "  uploaded tool_id=$tool_id" >&2
    else
        echo "  reusing existing tool_id=$tool_id" >&2
    fi

    "${KWEAV[@]}" tool enable --toolbox "$box_id" "$tool_id" >/dev/null 2>&1 || true

    local pub_rc=0 pub_out
    pub_out="$("${KWEAV[@]}" toolbox publish "$box_id" 2>&1)" || pub_rc=$?
    if [ "$pub_rc" -ne 0 ]; then
        case "$pub_out" in
            *ToolBoxStatusInvalid*|*"can not be transition to published"*|*"already"*)
                echo "  (already published)" >&2
                ;;
            *)
                echo "$pub_out" >&2
                echo "Error: toolbox publish failed." >&2
                exit "$pub_rc"
                ;;
        esac
    fi

    _write_box_id_to_env "$box_id" "$tool_id"
    TOOLBOX_BOX_ID="$box_id"; VEGA_TOOL_ID="$tool_id"
    export TOOLBOX_BOX_ID VEGA_TOOL_ID
    echo "  TOOLBOX_BOX_ID=$box_id  VEGA_TOOL_ID=$tool_id" >&2
}

# ─── Step 7: Create Agent ───────────────────────────────────────────────────
extract_agent_id() {
    python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("id") or d.get("agent_id") or "")' 2>/dev/null
}

find_agent_id_by_name() {
    local name="$1" raw
    raw="$("${KWEAV[@]}" agent personal-list --name "$name" --size 48 2>/dev/null)" || true
    printf '%s' "$raw" | jq -r --arg n "$name" '
        (if type == "array" then . elif .entries then .entries else .data // .items // [] end)
        | if type == "array" then . else [] end
        | map(select(.name == $n))
        | sort_by(.updated_at // .created_at // 0)
        | reverse
        | .[0].id // empty
    ' 2>/dev/null | head -1
}

write_agent_id_to_env() {
    local agent_id="$1"
    [ "${SKIP_WRITE_ENV:-0}" = 1 ] && { echo "  skip .env update (SKIP_WRITE_ENV=1)" >&2; return 0; }
    _WC_PATCH_AGENT_ID="$agent_id" ENV_FILE="$SCRIPT_DIR/.env" python3 - <<'PY'
import os, pathlib, re
agent_id = os.environ["_WC_PATCH_AGENT_ID"]
env_file = pathlib.Path(os.environ["ENV_FILE"])
nl = "\n"
if env_file.is_file():
    text = env_file.read_text(encoding="utf-8", errors="replace")
    if not text.endswith("\n"): text += nl
    pat = re.compile(r"^\s*AGENT_ID=")
    out, found = [], False
    for line in text.splitlines(keepends=True):
        if pat.match(line):
            if not found:
                out.append(f"AGENT_ID={agent_id}{nl}")
                found = True
        else:
            out.append(line)
    if not found: out.append(f"AGENT_ID={agent_id}{nl}")
    env_file.write_text("".join(out), encoding="utf-8")
else:
    env_file.write_text(f"AGENT_ID={agent_id}{nl}", encoding="utf-8")
print(f"  .env AGENT_ID={agent_id}")
PY
}

render_agent_config() {
    local kn_id="$1" out="$2"
    [ -f "$AGENT_TEMPLATE" ] || { echo "Error: $AGENT_TEMPLATE missing" >&2; exit 1; }

    # If TOOLBOX_BOX_ID/VEGA_TOOL_ID not set (step 6 skipped), drop the
    # vega_sql_execute tool from the template so the agent doesn't bind a
    # placeholder id. Otherwise substitute placeholders.
    local tmp_in="$AGENT_TEMPLATE"
    if [ -n "${VEGA_TOOL_ID:-}" ] && [ -n "${TOOLBOX_BOX_ID:-}" ]; then
        tmp_in="$(mktemp -t wc_agent_tpl.XXXXXX.json)"
        jq --arg vt "$VEGA_TOOL_ID" --arg vb "$TOOLBOX_BOX_ID" '
            (.skills.tools[] | select(.tool_id == "__VEGA_TOOL_ID__")) |= (
                .tool_id = $vt | .tool_box_id = $vb
            )
        ' "$AGENT_TEMPLATE" >"$tmp_in"
    else
        echo "  warn: VEGA_TOOL_ID/TOOLBOX_BOX_ID not set — dropping vega_sql_execute from agent" >&2
        tmp_in="$(mktemp -t wc_agent_tpl.XXXXXX.json)"
        jq '.skills.tools |= map(select(.tool_id != "__VEGA_TOOL_ID__"))' "$AGENT_TEMPLATE" >"$tmp_in"
    fi

    local filter='.data_source.knowledge_network = [{"knowledge_network_id": $kn, "object_types": null}]'
    if [ -n "${AGENT_LLM_ID:-}" ]; then
        filter+=' | (.llms[0].llm_config.id) = $llm'
        jq --arg kn "$kn_id" --arg llm "$AGENT_LLM_ID" "$filter" "$tmp_in" >"$out"
    else
        jq --arg kn "$kn_id" "$filter" "$tmp_in" >"$out"
    fi
    [ "$tmp_in" != "$AGENT_TEMPLATE" ] && rm -f "$tmp_in"
}

step_7_agent() {
    echo "=== [7/7] Create Agent ===" >&2
    local kn_id
    kn_id="$(kn_id_from_rendered)"
    [ -z "$kn_id" ] && kn_id="$(kn_id_from_template)"
    [ -z "$kn_id" ] && { echo "Error: no KN id (check $NETWORK_BKN)" >&2; exit 1; }

    local agent_name="${AGENT_NAME:-世界杯数据分析助手}"
    local agent_profile="${AGENT_PROFILE:-基于 Fjelstul 世界杯知识网络回答问题；多表推理时用关系与 *_name 字段，不确定时请说明假设}"
    local reuse="${REUSE_AGENT_BY_NAME:-true}"
    [ "$FORCE_NO_REUSE" = 1 ] && reuse=false

    if [ "$DRY_RUN" = 1 ]; then
        echo "  plan: KN_ID=$kn_id  AGENT_NAME=$agent_name  REUSE=$reuse  PUBLISH=$DO_PUBLISH" >&2
        echo "  plan: render $AGENT_TEMPLATE → temp config (substitute knowledge_network_id)" >&2
        echo "  plan: kweaver agent create --name … --profile … --config <tmp>" >&2
        echo "  plan: kweaver agent update <id> --knowledge-network-id $kn_id" >&2
        [ "$DO_PUBLISH" = 1 ] && echo "  plan: kweaver agent publish <id>" >&2
        return 0
    fi

    local agent_id=""
    if [ "$reuse" = true ] || [ "$reuse" = 1 ]; then
        agent_id="$(find_agent_id_by_name "$agent_name")"
        [ -n "$agent_id" ] && echo "  reusing AGENT_ID=$agent_id (matched name='$agent_name')" >&2
    fi

    if [ -z "$agent_id" ]; then
        local tmp_cfg
        tmp_cfg="$(mktemp -t wc_agent_cfg.XXXXXX.json)"
        render_agent_config "$kn_id" "$tmp_cfg"
        echo "  agent create (config rendered for KN=$kn_id)" >&2
        local create_out
        create_out="$("${KWEAV[@]}" agent create \
            --name "$agent_name" \
            --profile "$agent_profile" \
            --config "$tmp_cfg")"
        rm -f "$tmp_cfg"
        agent_id="$(printf '%s' "$create_out" | extract_agent_id)"
        [ -z "$agent_id" ] && { echo "Error: agent create returned no id:" >&2; echo "$create_out" >&2; exit 1; }
        echo "  created AGENT_ID=$agent_id" >&2
    fi

    echo "  bind KN $kn_id" >&2
    "${KWEAV[@]}" agent update "$agent_id" --knowledge-network-id "$kn_id"

    if [ "$DO_PUBLISH" = 1 ]; then
        echo "  publish" >&2
        "${KWEAV[@]}" agent publish "$agent_id"
    else
        echo "  skip publish (--no-publish)" >&2
    fi

    write_agent_id_to_env "$agent_id"

    echo "" >&2
    echo "  Done. AGENT_ID=$agent_id  KN=$kn_id" >&2
    echo "  Chat: ${KWEAV[*]} agent chat $agent_id -m '列出近五届世界杯冠军'" >&2
}

# ─── Driver ─────────────────────────────────────────────────────────────────
require_jq

run_step() {
    local n="$1"
    if [ -n "$ONLY" ]; then
        [ "$ONLY" = "$n" ] || return 0
    else
        [ "$n" -ge "$FROM" ] || return 0
    fi
    case "$n" in
        1) step_1_download ;;
        2) step_2_import ;;
        3) step_3_vega_scan ;;
        4) step_4_render_bkn ;;
        5) step_5_push_bkn ;;
        6) step_6_toolbox ;;
        7) step_7_agent ;;
    esac
}

for n in 1 2 3 4 5 6 7; do
    run_step "$n"
done
