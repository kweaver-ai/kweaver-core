#!/usr/bin/env bash
# =============================================================================
# Create (or reuse) a Decision Agent bound to this example's business KN.
#
# Agent wiring:
#   kweaver agent create → agent update --knowledge-network-id → agent publish
#
# Prerequisites:
#   - Logged in: kweaver auth login <url>
#   - Business domain: kweaver config show  (see kweaver-core SKILL)
#   - KN already on platform: e.g. bkn push of worldcup-bkn-vega (or KN from ./run.sh)
#
# Required in .env (or environment):
#   AGENT_LLM_ID   From Studio or: kweaver model llm list
#
# KN resolution (first match wins):
#   WC_KN_ID / KN_ID env, else `id:` from worldcup-bkn-vega/network.bkn
# =============================================================================
set -euo pipefail
[ "${WC_TRACE:-0}" = 1 ] && set -x || true

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_BKN="$SCRIPT_DIR/worldcup-bkn-vega/network.bkn"

usage() {
    cat <<'EOF'
Usage: ./create-agent-from-kn.sh [--dry-run] [--no-publish] [--reuse | --no-reuse]

Reads ./.env when present (optional KWEAVER_BASE_URL override, AGENT_*, WC_KN_ID, KN_ID).

Environment (see env.sample block «create-agent-from-kn»):
  AGENT_LLM_ID          Required for create (unless REUSE updates only).
  AGENT_NAME            Default: 世界杯数据分析助手
  AGENT_PROFILE         One-line description (max 500 chars).
  AGENT_SYSTEM_PROMPT   Optional. Shown to the LLM as system prompt.
  WC_KN_ID / KN_ID      Override KN id (else parsed from worldcup-bkn-vega/network.bkn).
  REUSE_AGENT_BY_NAME   true → if personal-list has same AGENT_NAME, only bind KN + publish.

Flags:
  --dry-run         Print planned commands, no API calls.
  --no-publish      Skip `agent publish` (agent stays in private space).
  --reuse           Force REUSE_AGENT_BY_NAME=1 for this run.
  --no-reuse        Force REUSE_AGENT_BY_NAME=0 for this run.

Env:
  SKIP_WRITE_ENV=1  Do not modify .env (default: write AGENT_ID after success).

After success, updates ./.env: sets or replaces AGENT_ID= (UTF-8). Disable with SKIP_WRITE_ENV=1.

Note on verification:
  kweaver agent list only lists PUBLISHED (square) agents — empty [] is normal if you
  used --no-publish or publish did not run. Private / draft agents:
    kweaver agent personal-list [--name <keyword>]
  Or: kweaver agent get <AGENT_ID>

Discover LLM id:
  kweaver model llm list
EOF
}

DRY_RUN=0
DO_PUBLISH=1
FORCE_REUSE=""
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        --dry-run) DRY_RUN=1; shift ;;
        --no-publish) DO_PUBLISH=0; shift ;;
        --reuse) FORCE_REUSE=1; shift ;;
        --no-reuse) FORCE_REUSE=0; shift ;;
        *) echo "Unknown arg: $1" >&2; usage >&2; exit 2 ;;
    esac
done

set -a
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"
set +a

# Default: ~/.kweaver active platform. Optional KWEAVER_BASE_URL → --base-url.
if [ -n "${KWEAVER_BASE_URL:-}" ]; then
    KWEAV=(kweaver --base-url "${KWEAVER_BASE_URL}")
else
    KWEAV=(kweaver)
fi

AGENT_NAME="${AGENT_NAME:-世界杯数据分析助手}"
AGENT_PROFILE="${AGENT_PROFILE:-基于 Fjelstul 世界杯知识网络回答问题；多表推理时用关系与 *_name 字段，不确定时请说明假设}"
AGENT_SYSTEM_PROMPT="${AGENT_SYSTEM_PROMPT:-你是世界杯数据分析助手。知识来自网络 worldcup_fjelstul_bkn：赛程、球队、球员、进球、小组积分、奖项等。优先用对象类查询与关系，引用具体届次、球队、球员名称。}"
REUSE_AGENT_BY_NAME="${REUSE_AGENT_BY_NAME:-false}"
if [ "$FORCE_REUSE" = 1 ]; then REUSE_AGENT_BY_NAME=true; fi
if [ "$FORCE_REUSE" = 0 ]; then REUSE_AGENT_BY_NAME=false; fi

read_kn_id() {
    if [ -n "${WC_KN_ID:-}" ]; then
        echo "$WC_KN_ID"; return
    fi
    if [ -n "${KN_ID:-}" ]; then
        echo "$KN_ID"; return
    fi
    if [ ! -f "$NETWORK_BKN" ]; then
        echo "Error: missing $NETWORK_BKN and no WC_KN_ID/KN_ID" >&2
        exit 1
    fi
    awk '/^id:/ {print $2; exit}' "$NETWORK_BKN"
}

require_jq() {
    if ! command -v jq >/dev/null 2>&1; then
        echo "Error: jq is required (brew install jq)" >&2
        exit 1
    fi
}

find_agent_id_by_name() {
    local name="$1"
    local raw
    raw=$("${KWEAV[@]}" agent personal-list --name "$name" --size 48 2>/dev/null) || true
    printf '%s' "$raw" | jq -r --arg n "$name" '
        (if type == "array" then . elif .entries then .entries else .data // .items // [] end)
        | if type == "array" then . else [] end
        | map(select(.name == $n))
        | sort_by(.updated_at // .created_at // 0)
        | reverse
        | .[0].id // empty
    ' 2>/dev/null | head -1
}

extract_agent_id() {
    python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("id") or d.get("agent_id") or "")' 2>/dev/null
}

# First model_id from model factory (same business domain as CLI). Empty if none.
resolve_first_llm_id() {
    local raw
    raw=$("${KWEAV[@]}" model llm list --limit 20 --json 2>/dev/null) || true
    printf '%s' "$raw" | jq -r '.data[0].model_id // empty' 2>/dev/null | head -1
}

# Upsert AGENT_ID= in ./.env (removes duplicate AGENT_ID lines). Pass agent id as $1.
write_agent_id_to_env() {
    local agent_id="$1"
    local env_file="$SCRIPT_DIR/.env"
    if [ "${SKIP_WRITE_ENV:-0}" = 1 ] || \
       [ "${SKIP_WRITE_ENV:-}" = true ]; then
        echo "=== skip .env update (SKIP_WRITE_ENV) ==="
        return 0
    fi
    _WC_PATCH_AGENT_ID="$agent_id" ENV_FILE="$env_file" python3 <<'PY'
import os
import pathlib
import re

agent_id = os.environ["_WC_PATCH_AGENT_ID"]
env_file = pathlib.Path(os.environ["ENV_FILE"])

newline = "\n"
if env_file.is_file():
    text = env_file.read_text(encoding="utf-8", errors="replace")
    if not text.endswith("\n"):
        text += newline
    lines = text.splitlines(keepends=True)
    out = []
    found = False
    pat = re.compile(r"^\s*AGENT_ID=")
    for line in lines:
        if pat.match(line):
            if not found:
                out.append(f"AGENT_ID={agent_id}{newline}")
                found = True
        else:
            out.append(line)
    if not found:
        out.append(f"AGENT_ID={agent_id}{newline}")
    env_file.write_text("".join(out), encoding="utf-8")
    action = "updated"
else:
    body = (
        f"# Written by create-agent-from-kn.sh{newline}"
        f"AGENT_ID={agent_id}{newline}"
    )
    env_file.write_text(body, encoding="utf-8")
    action = "created"
print(f"  .env {action}: {env_file} (AGENT_ID={agent_id})")
PY
}

KN_ID=$(read_kn_id)
[ -z "$KN_ID" ] && { echo "Error: could not resolve knowledge network id" >&2; exit 1; }

require_jq

if [ "$DRY_RUN" = 1 ]; then
    cat <<EOF
Planned:
  KN_ID=$KN_ID (env WC_KN_ID / KN_ID override worldcup-bkn-vega/network.bkn)
  REUSE_AGENT_BY_NAME=$REUSE_AGENT_BY_NAME  AGENT_NAME=$AGENT_NAME
  DO_PUBLISH=$DO_PUBLISH
  create: kweaver agent create --name ... --profile ... --llm-id <AGENT_LLM_ID> [--system-prompt ...]
      or from template: AGENT_TEMPLATE_ID + template-get --save-config
  bind:   kweaver agent update <id> --knowledge-network-id $KN_ID
  publish: [optional] kweaver agent publish <id>
  then: update ./.env AGENT_ID= (unless SKIP_WRITE_ENV=1)
EOF
    [ -z "${AGENT_LLM_ID:-}" ] && [ -z "${AGENT_TEMPLATE_ID:-}" ] && \
        echo "  (AGENT_LLM_ID / AGENT_TEMPLATE_ID empty — create will fail unless REUSE finds an agent)" >&2
    exit 0
fi

# Optional: pick first registered LLM when .env leaves AGENT_LLM_ID empty.
if [ -z "${AGENT_LLM_ID:-}" ] && [ -z "${AGENT_TEMPLATE_ID:-}" ]; then
    _auto_llm=$(resolve_first_llm_id)
    if [ -n "$_auto_llm" ]; then
        AGENT_LLM_ID="$_auto_llm"
        echo "=== Auto-selected AGENT_LLM_ID=$AGENT_LLM_ID (first entry from kweaver model llm list) ==="
    fi
    unset _auto_llm
fi

AGENT_ID=""
if [ "$REUSE_AGENT_BY_NAME" = true ] || [ "$REUSE_AGENT_BY_NAME" = 1 ]; then
    AGENT_ID=$(find_agent_id_by_name "$AGENT_NAME")
    if [ -n "$AGENT_ID" ]; then
        echo "=== Reusing existing agent name='$AGENT_NAME' AGENT_ID=$AGENT_ID ==="
    fi
fi

if [ -z "$AGENT_ID" ]; then
    if [ -n "${AGENT_TEMPLATE_ID:-}" ]; then
        echo "=== template-get $AGENT_TEMPLATE_ID → temp config ==="
        TPLDIR=$(mktemp -d -t wc_agent_tpl.XXXXXX)
        if ! "${KWEAV[@]}" agent template-get "$AGENT_TEMPLATE_ID" --save-config "$TPLDIR/"; then
            rm -rf "$TPLDIR"
            exit 1
        fi
        CONFIG_FILE=$(find "$TPLDIR" -maxdepth 1 -type f -name '*.json' | head -1)
        if [ -z "${CONFIG_FILE:-}" ] || [ ! -f "$CONFIG_FILE" ]; then
            echo "Error: template --save-config did not yield a .json under $TPLDIR" >&2
            rm -rf "$TPLDIR"
            exit 1
        fi
        echo "=== agent create (from template) ==="
        AGENT_JSON=$("${KWEAV[@]}" agent create \
            --name "$AGENT_NAME" \
            --profile "$AGENT_PROFILE" \
            --config "$CONFIG_FILE")
        rm -rf "$TPLDIR"
    else
        if [ -z "${AGENT_LLM_ID:-}" ]; then
            cat <<'ERR' >&2
Error: no LLM available for agent create.

  1) Register at least one LLM in Model Factory (Studio / kweaver model llm add — see help/en/model.md).
  2) Or set AGENT_LLM_ID in .env to a model_id from: kweaver model llm list
  3) Or set AGENT_TEMPLATE_ID if your platform exposes agent templates.

This platform returned an empty list from: kweaver model llm list --json
ERR
            exit 1
        fi
        echo "=== agent create ==="
        if [ -n "${AGENT_SYSTEM_PROMPT:-}" ]; then
            AGENT_JSON=$("${KWEAV[@]}" agent create \
                --name "$AGENT_NAME" \
                --profile "$AGENT_PROFILE" \
                --llm-id "$AGENT_LLM_ID" \
                --system-prompt "$AGENT_SYSTEM_PROMPT")
        else
            AGENT_JSON=$("${KWEAV[@]}" agent create \
                --name "$AGENT_NAME" \
                --profile "$AGENT_PROFILE" \
                --llm-id "$AGENT_LLM_ID")
        fi
    fi
    AGENT_ID=$(echo "$AGENT_JSON" | extract_agent_id)
    [ -z "$AGENT_ID" ] && { echo "Error: agent create failed or could not parse id:" >&2; echo "$AGENT_JSON" >&2; exit 1; }
    echo "  Created AGENT_ID=$AGENT_ID"
fi

echo "=== agent update — bind knowledge network $KN_ID ==="
"${KWEAV[@]}" agent update "$AGENT_ID" --knowledge-network-id "$KN_ID"

if [ "$DO_PUBLISH" = 1 ]; then
    echo "=== agent publish ==="
    "${KWEAV[@]}" agent publish "$AGENT_ID"
else
    echo "=== skip publish (DO_PUBLISH=0 / --no-publish) ==="
fi

echo ""
write_agent_id_to_env "$AGENT_ID"

echo ""
echo "=== Verify (agent exists on platform) ==="
if ! OUT=$("${KWEAV[@]}" agent get "$AGENT_ID" 2>&1); then
    echo "  Warning: agent get failed — creation may not have persisted. Output:" >&2
    echo "$OUT" >&2
else
    echo "  agent get $AGENT_ID — OK"
fi

echo ""
echo "=== Done ==="
echo "  AGENT_ID=$AGENT_ID"
echo "  KN_ID=$KN_ID"
echo "  Chat: ${KWEAV[*]} agent chat $AGENT_ID -m '列出近五届世界杯冠军'"
if [ "$DO_PUBLISH" != 1 ]; then
    echo ""
    echo "  Hint: you skipped publish (--no-publish). This agent will NOT appear in"
    echo "        kweaver agent list (published only). Use:"
    echo "          kweaver agent personal-list --name '$AGENT_NAME'"
else
    echo ""
    echo "  Hint: kweaver agent list shows published agents only. If still empty, wait a"
    echo "        few seconds or run: kweaver agent personal-list --name '$AGENT_NAME'"
fi
