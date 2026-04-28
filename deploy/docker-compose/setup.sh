#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

usage() {
  cat <<'EOF'
Usage: ./setup.sh [options]

Generates configs/generated/* from configs/kweaver/**/*.tmpl and validates Compose.

Shared password (one value used for ALL three fields below):

  -p, --password=PW             Sets MARIADB_ROOT_PASSWORD, MARIADB_PASSWORD,
                                and MINIO_ROOT_PASSWORD at once. Equivalent
                                to setting the three flags below to the same
                                value. Per-field flags override this.

Per-field passwords (override the shared --password if both are given):

  --mariadb-root-password=PW    Sets MARIADB_ROOT_PASSWORD
  --mariadb-password=PW         Sets MARIADB_PASSWORD
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

if ! docker compose version >/dev/null 2>&1; then
  echo "Docker Compose v2 is required before setup can continue." >&2
  echo "" >&2
  echo "Please install the Docker Compose v2 plugin and use: docker compose" >&2
  echo "The legacy docker-compose v1 command is not supported." >&2
  echo "" >&2
  echo "Ubuntu/Debian example:" >&2
  echo "  sudo apt-get update" >&2
  echo "  sudo apt-get install -y docker-compose-plugin" >&2
  echo "" >&2
  echo "CentOS/RHEL example:" >&2
  echo "  sudo yum install -y docker-compose-plugin" >&2
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

read_secret_confirmed() {
  local prompt="$1" first="" second=""
  while true; do
    read -r -s -p "$prompt" first </dev/tty
    echo >&2
    if [[ -z "$first" ]]; then
      printf ''
      return
    fi

    read -r -s -p "Confirm password: " second </dev/tty
    echo >&2
    if [[ "$first" == "$second" ]]; then
      printf '%s' "$first"
      return
    fi

    echo "Passwords do not match; please try again." >&2
  done
}

# Shared password from env var fallback (only if --password was not given).
SHARED_PW_ENV="${PASSWORD:-}"
if ! $CLI_SHARED_PW_SET && [[ -z "$SHARED_PW_ENV" ]] && $INTERACTIVE && [[ -t 0 ]] \
    && ! $CLI_ROOT_PW_SET && ! $CLI_ADP_PW_SET && ! $CLI_MINIO_PW_SET \
    && [[ -z "${MARIADB_ROOT_PASSWORD:-}" ]] \
    && [[ -z "${MARIADB_PASSWORD:-}" ]] \
    && [[ -z "${MINIO_ROOT_PASSWORD:-}" ]]; then
  entered="$(read_secret_confirmed "Enter a single password to use for MariaDB root, MariaDB '${MARIADB_USER:-adp}', and MinIO root [Enter to skip and configure each separately]: ")"
  if [[ -n "$entered" ]]; then
    CLI_SHARED_PW="$entered"
    CLI_SHARED_PW_SET=true
  fi
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
require_env_var DIP_NAMESPACE "image path segment (usually dip)"
require_env_var KWEAVER_VERSION "KWeaver Core image tag"
require_env_var MARIADB_USER "MariaDB application user, usually adp"
require_env_var MARIADB_DATABASE "MariaDB application database, usually adp"
require_env_var MINIO_ROOT_USER "MinIO root user"
require_env_var ACCESS_HOST "public host used in generated config"
require_env_var ACCESS_PORT "public port used in generated config"

IMAGE_REGISTRY_VALUE="$(read_env_var .env IMAGE_REGISTRY)"
DIP_NAMESPACE_VALUE="$(read_env_var .env DIP_NAMESPACE)"

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
    entered="$(read_secret_confirmed "Enter ${key}${prompt_default}: ")"
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
    echo "embedded in generated configs without URL encoding when required." >&2
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

MARIADB_HOST_VAL="$(read_env_var .env MARIADB_HOST)"; MARIADB_HOST_VAL="${MARIADB_HOST_VAL:-mariadb}"
MARIADB_PORT_VAL="$(read_env_var .env MARIADB_PORT)"; MARIADB_PORT_VAL="${MARIADB_PORT_VAL:-3306}"
SANDBOX_DB_EXISTING="$(read_env_var .env SANDBOX_DATABASE_URL)"
if [[ -z "$SANDBOX_DB_EXISTING" ]]; then
  set_env_var .env SANDBOX_DATABASE_URL "mysql+aiomysql://${ADP_USER}:${ADP_PW}@${MARIADB_HOST_VAL}:${MARIADB_PORT_VAL}/sandbox"
fi

if ! python3 "${ROOT}/tools/manifest.py" check-compose; then
  echo "ERROR: docker-compose.yml is out of sync with compose-manifest.yaml." >&2
  exit 1
fi

# Backfill any tagEnv default from compose-manifest.yaml that .env is missing.
while IFS='=' read -r mkey mval; do
  [[ -z "$mkey" ]] && continue
  current="$(read_env_var .env "$mkey" || true)"
  if [[ -z "$current" ]]; then
    set_env_var .env "$mkey" "$mval"
    echo "Set ${mkey}=${mval} from compose-manifest.yaml" >&2
  fi
done < <(python3 "${ROOT}/tools/manifest.py" env-defaults)

# Warn (do not fail) when .env *_VERSION drifts from manifest — operators may
# pin a custom tag locally; we only block missing values via :? in compose.
python3 "${ROOT}/tools/manifest.py" check-env .env || \
  echo "WARN: .env *_VERSION values differ from compose-manifest.yaml (see above)." >&2

mkdir -p "${ROOT}/configs/generated"

if ! python3 "${ROOT}/tools/render_compose_configs.py"; then
  echo "ERROR: config render failed (tools/render_compose_configs.py)." >&2
  exit 1
fi

if [[ -f "${HOME}/.docker/config.json" ]] \
    && grep -q 'swr.cn-east-3.myhuaweicloud.com' "${HOME}/.docker/config.json" 2>/dev/null; then
  :
else
  echo "NOTE: Default KWeaver SWR images should pull anonymously when paths include /${DIP_NAMESPACE_VALUE}/." >&2
  echo "      If pull fails, run: docker compose config --images" >&2
  echo "      Confirm paths look like ${IMAGE_REGISTRY_VALUE}/${DIP_NAMESPACE_VALUE}/<image>:<tag> before trying docker login swr.cn-east-3.myhuaweicloud.com." >&2
fi

cd "${ROOT}"
export COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-kweaver-compose}"

docker compose config >/dev/null
ACCESS_HINT="$(read_env_var .env ACCESS_HOST)"; ACCESS_HINT="${ACCESS_HINT:-localhost}"
PORT_HINT="$(read_env_var .env KWEAVER_HTTP_PORT)"; PORT_HINT="${PORT_HINT:-8080}"
echo "Wrote configs under ${ROOT}/configs/generated/. docker compose configuration is valid."
echo "Start infra when ready: ./compose.sh infra up"
echo "Then start KWeaver services: ./compose.sh app up"
echo "HTTP entry (nginx): http://${ACCESS_HINT}:${PORT_HINT}/healthz"
