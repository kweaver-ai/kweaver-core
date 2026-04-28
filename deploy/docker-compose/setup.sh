#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

usage() {
  cat <<'EOF'
Usage: ./setup.sh [options]

Generates configs/generated/config.yaml from .env and validates Compose.

Shared password (one value used for ALL three fields below):

  -p, --password=PW             Sets MARIADB_ROOT_PASSWORD, MARIADB_PASSWORD,
                                and MINIO_ROOT_PASSWORD at once. Equivalent
                                to setting the three flags below to the same
                                value. Per-field flags override this.

Per-field passwords (override the shared --password if both are given):

  --mariadb-root-password=PW    Sets MARIADB_ROOT_PASSWORD
  --mariadb-password=PW         Sets MARIADB_PASSWORD (also rewrites
                                SANDBOX_DATABASE_URL to embed this password)
  --minio-root-password=PW      Sets MINIO_ROOT_PASSWORD

Environment variables are honored as fallbacks:
  PASSWORD=secret ./setup.sh                 # shared
  MARIADB_PASSWORD=secret ./setup.sh         # per-field

Other options:
  -y, --non-interactive         Do not prompt; keep existing values when no
                                override is provided
  -h, --help                    Show this help

Resolution order per field (highest wins):
  per-field CLI flag  >  per-field env var
    >  shared --password CLI flag  >  shared PASSWORD env var
    >  current .env value
    >  interactive prompt (TTY only)
    >  error exit if still empty
EOF
}

INTERACTIVE=true
CLI_SHARED_PW=""
CLI_SHARED_PW_SET=false
CLI_ROOT_PW=""
CLI_ADP_PW=""
CLI_MINIO_PW=""
CLI_ROOT_PW_SET=false
CLI_ADP_PW_SET=false
CLI_MINIO_PW_SET=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    -p|--password)             CLI_SHARED_PW="${2:-}"; CLI_SHARED_PW_SET=true; shift 2;;
    --password=*)              CLI_SHARED_PW="${1#*=}"; CLI_SHARED_PW_SET=true; shift;;
    --mariadb-root-password=*) CLI_ROOT_PW="${1#*=}"; CLI_ROOT_PW_SET=true; shift;;
    --mariadb-root-password)   CLI_ROOT_PW="${2:-}";  CLI_ROOT_PW_SET=true; shift 2;;
    --mariadb-password=*)      CLI_ADP_PW="${1#*=}";  CLI_ADP_PW_SET=true;  shift;;
    --mariadb-password)        CLI_ADP_PW="${2:-}";   CLI_ADP_PW_SET=true;  shift 2;;
    --minio-root-password=*)   CLI_MINIO_PW="${1#*=}"; CLI_MINIO_PW_SET=true; shift;;
    --minio-root-password)     CLI_MINIO_PW="${2:-}";  CLI_MINIO_PW_SET=true; shift 2;;
    -y|--non-interactive)      INTERACTIVE=false; shift;;
    -h|--help)                 usage; exit 0;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2;;
  esac
done

# Shared password from env var fallback (only if --password was not given).
SHARED_PW_ENV="${PASSWORD:-}"
if ! $CLI_SHARED_PW_SET && [[ -z "$SHARED_PW_ENV" ]] && $INTERACTIVE && [[ -t 0 ]] \
    && ! $CLI_ROOT_PW_SET && ! $CLI_ADP_PW_SET && ! $CLI_MINIO_PW_SET \
    && [[ -z "${MARIADB_ROOT_PASSWORD:-}" ]] \
    && [[ -z "${MARIADB_PASSWORD:-}" ]] \
    && [[ -z "${MINIO_ROOT_PASSWORD:-}" ]]; then
  read -r -s -p "Enter a single password to use for MariaDB root, MariaDB '${MARIADB_USER:-adp}', and MinIO root [Enter to skip and configure each separately]: " entered </dev/tty
  echo >&2
  if [[ -n "$entered" ]]; then
    CLI_SHARED_PW="$entered"
    CLI_SHARED_PW_SET=true
  fi
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "Docker Compose v2 CLI is required (\"docker compose\")." >&2
  echo "The legacy \"docker-compose\" v1 command is not supported." >&2
  exit 1
fi

compose_version="$(docker compose version --short 2>/dev/null || true)"
if [[ -n "$compose_version" ]]; then
  python3 - "$compose_version" <<'PY'
import re
import sys

raw = sys.argv[1]
m = re.search(r"(\d+)\.(\d+)\.(\d+)", raw)
if not m:
    sys.exit(0)

version = tuple(map(int, m.groups()))
if version < (2, 17, 0):
    print(
        f"WARNING: Docker Compose {raw} detected. "
        "Recommended: v2.17.0+ (v2.20.0+ preferred).",
        file=sys.stderr,
    )
PY
fi

if [[ ! -f .env ]] && [[ ! -f .env.example ]]; then
  echo "Missing .env and .env.example." >&2
  exit 1
fi

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "Created .env from .env.example."
fi

# Update KEY=VAL in a file in place. Value is treated as a literal.
set_env_var() {
  local file="$1" key="$2" val="$3"
  python3 - "$file" "$key" "$val" <<'PY'
import re, sys
from pathlib import Path
path, key, val = sys.argv[1], sys.argv[2], sys.argv[3]
text = Path(path).read_text(encoding="utf-8")
pattern = re.compile(rf"^{re.escape(key)}=.*$", re.M)
new_line = f"{key}={val}"
if pattern.search(text):
    text = pattern.sub(new_line, text, count=1)
else:
    text = text.rstrip("\n") + f"\n{new_line}\n"
Path(path).write_text(text, encoding="utf-8")
PY
}

read_env_var() {
  local file="$1" key="$2"
  grep -E "^${key}=" "$file" 2>/dev/null | head -1 | cut -d= -f2-
}

require_env_var() {
  local key="$1" hint="$2" val
  val="$(read_env_var .env "$key")"
  if [[ -z "$val" ]]; then
    echo "ERROR: ${key} is empty in .env." >&2
    echo "Set ${key}${hint:+ (${hint})} and re-run this script." >&2
    exit 1
  fi
}

require_env_var IMAGE_REGISTRY "container image registry"
require_env_var KWEAVER_VERSION "KWeaver Core image tag"
require_env_var MARIADB_USER "MariaDB application user, usually adp"
require_env_var MARIADB_DATABASE "MariaDB application database, usually adp"
require_env_var MINIO_ROOT_USER "MinIO root user"
require_env_var ACCESS_HOST "public host used in generated config"
require_env_var ACCESS_PORT "public port used in generated config"

# Resolve final value for one password key.
# Args: KEY  CLI_VAL  CLI_SET(true|false)
resolve_pw() {
  local key="$1" cli_val="$2" cli_set="$3"
  local current env_val
  current="$(read_env_var .env "$key")"
  env_val="${!key:-}"

  if [[ "$cli_set" == "true" ]]; then
    printf '%s' "$cli_val"
    return
  fi
  if [[ -n "$env_val" ]]; then
    printf '%s' "$env_val"
    return
  fi
  if [[ "$CLI_SHARED_PW_SET" == "true" ]]; then
    printf '%s' "$CLI_SHARED_PW"
    return
  fi
  if [[ -n "$SHARED_PW_ENV" ]]; then
    printf '%s' "$SHARED_PW_ENV"
    return
  fi
  if $INTERACTIVE && [[ -t 0 ]]; then
    local entered="" prompt_default=""
    if [[ -n "$current" ]]; then
      prompt_default=" [Enter to keep current '${current}']"
    else
      prompt_default=" [required]"
    fi
    read -r -s -p "Enter ${key}${prompt_default}: " entered </dev/tty
    echo >&2
    if [[ -n "$entered" ]]; then
      printf '%s' "$entered"
      return
    fi
  fi
  printf '%s' "$current"
}

ROOT_PW="$(resolve_pw MARIADB_ROOT_PASSWORD "$CLI_ROOT_PW" "$CLI_ROOT_PW_SET")"
ADP_PW="$(resolve_pw MARIADB_PASSWORD       "$CLI_ADP_PW"  "$CLI_ADP_PW_SET")"
MINIO_PW="$(resolve_pw MINIO_ROOT_PASSWORD  "$CLI_MINIO_PW" "$CLI_MINIO_PW_SET")"

# Refuse to render if any required password is still empty.
UNFILLED=()
[[ -z "$ROOT_PW"  ]] && UNFILLED+=("MARIADB_ROOT_PASSWORD (--mariadb-root-password)")
[[ -z "$ADP_PW"   ]] && UNFILLED+=("MARIADB_PASSWORD (--mariadb-password)")
[[ -z "$MINIO_PW" ]] && UNFILLED+=("MINIO_ROOT_PASSWORD (--minio-root-password)")

if (( ${#UNFILLED[@]} > 0 )); then
  echo "ERROR: the following secrets are still empty:" >&2
  for f in "${UNFILLED[@]}"; do echo "  - $f" >&2; done
  echo "Provide them via the matching --... flag, the same-name env var," >&2
  echo "the shared -p / --password flag (or PASSWORD env var), or by editing .env" >&2
  echo "and re-running this script." >&2
  exit 1
fi

validate_secret() {
  local key="$1" val="$2"
  if [[ ! "$val" =~ ^[A-Za-z0-9_-]+$ ]]; then
    echo "ERROR: ${key} contains unsupported characters." >&2
    echo "Use only [A-Za-z0-9_-]. These values are written to .env and" >&2
    echo "embedded in SANDBOX_DATABASE_URL without URL encoding." >&2
    exit 1
  fi
}

ADP_USER="$(read_env_var .env MARIADB_USER)"
ADP_USER="${ADP_USER:-adp}"

validate_secret MARIADB_ROOT_PASSWORD "$ROOT_PW"
validate_secret MARIADB_USER "$ADP_USER"
validate_secret MARIADB_PASSWORD "$ADP_PW"
validate_secret MINIO_ROOT_PASSWORD "$MINIO_PW"

# Write back only after validation, so a bad CLI/env value never corrupts .env.
set_env_var .env MARIADB_ROOT_PASSWORD "$ROOT_PW"
set_env_var .env MARIADB_PASSWORD      "$ADP_PW"
set_env_var .env MINIO_ROOT_PASSWORD   "$MINIO_PW"

# Always rebuild SANDBOX_DATABASE_URL so it stays consistent with MARIADB_USER/PASSWORD.
set_env_var .env SANDBOX_DATABASE_URL \
  "mysql+aiomysql://${ADP_USER}:${ADP_PW}@mariadb:3306/sandbox"

mkdir -p "${ROOT}/configs/generated"

export TEMPLATE="${ROOT}/configs/kweaver/config.yaml.template"
export OUT="${ROOT}/configs/generated/config.yaml"
export ENV_FILE="${ROOT}/.env"

python3 - <<'PY'
import os
from pathlib import Path

def read_env(path: Path):
    values = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key] = value
    return values

env = read_env(Path(os.environ["ENV_FILE"]))
tpl = Path(os.environ["TEMPLATE"]).read_text(encoding="utf-8")
rep = {
    "__IMAGE_REGISTRY__": env.get("IMAGE_REGISTRY", ""),
    "__ACCESS_HOST__": env.get("ACCESS_HOST", "localhost"),
    "__ACCESS_PORT__": env.get("ACCESS_PORT", "8080"),
    "__MARIADB_USER__": env.get("MARIADB_USER", ""),
    "__MARIADB_PASSWORD__": env.get("MARIADB_PASSWORD", ""),
    "__MARIADB_DATABASE__": env.get("MARIADB_DATABASE", ""),
}
for k, v in rep.items():
    tpl = tpl.replace(k, v)
Path(os.environ["OUT"]).write_text(tpl, encoding="utf-8")
PY

export KWEAVER_CONFIG_FILE="${KWEAVER_CONFIG_FILE:-./configs/generated/config.yaml}"

docker compose config >/dev/null
echo "Wrote ${OUT}. docker compose configuration is valid."
echo "Start stack when ready (optional): docker compose up -d"
echo "Sandbox control plane: http://localhost:${SANDBOX_HTTP_PORT:-8001}"
