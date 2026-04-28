#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

INFRA_SERVICES=(
  mariadb
  redis
  zookeeper
  kafka
  opensearch
  minio
)

APP_SERVICES=(
  kweaver-core-data-migrator
  mf-model-manager
  mf-model-api
  bkn-backend
  ontology-query
  vega-backend
  data-connection
  vega-gateway
  vega-gateway-pro
  mdl-data-model
  mdl-uniquery
  mdl-data-model-job
  agent-operator-integration
  agent-retrieval
  agent-backend
  dataflow
  flow-stream-data-pipeline
  dataflowtools
  coderunner
  doc-convert-gotenberg
  doc-convert-tika
  sandbox
  oss-gateway-backend
  otelcol-contrib
  agent-observability
  nginx
)

# Vega stack only (aligned with deploy.sh grouping: install_vega vs install_bkn).
# Does not start nginx — compose nginx depends on the full app; use ./compose.sh app up
# or ./compose.sh all up for port 8080 routing. Access Vega on container ports directly.
VEGA_SERVICES=(
  vega-backend
  data-connection
  vega-gateway
  vega-gateway-pro
)

usage() {
  cat <<'EOF'
Usage: ./compose.sh <module> <action> [docker compose args...]

Modules:
  setup             Render configs and validate Compose (delegates to ./setup.sh)
  infra             Public dependency services only
  vega              Vega-related services only (vega-backend, data-connection, gateways)
  app               Same release set as deploy.sh kweaver-core install --minimum
                    (see deploy/release-manifests/*/kweaver-core.yaml), plus nginx
  all               infra + app

Actions:
  setup run         Run ./setup.sh
  infra up          Start MariaDB/Redis/Zookeeper/Kafka/OpenSearch/MinIO
  vega up           Start infra, migrator, then Vega services (no nginx; see README)
  app up            Start migrator + all kweaver-core manifest services + nginx
  all up            Start infra first, then app
  <module> pull     Pull images for that module
  <module> down     Stop/remove services for that module
  <module> restart  Restart services for that module
  <module> status   Show service status
  <module> logs     Follow logs for that module

Examples:
  ./compose.sh setup run -p YOUR_PASSWORD -y
  ./compose.sh infra up
  ./compose.sh vega up
  ./compose.sh app pull
  ./compose.sh app up
  ./compose.sh all up
EOF
}

ensure_setup_files() {
  if [[ ! -f .env || ! -d configs/generated ]]; then
    echo "ERROR: run ./compose.sh setup run first." >&2
    exit 1
  fi
}

compose_for_services() {
  local action="$1"
  shift
  local -a services=("$@")

  case "${action}" in
    up)
      docker compose up -d "${services[@]}"
      ;;
    pull)
      docker compose pull "${services[@]}"
      ;;
    down)
      docker compose rm -sf "${services[@]}"
      ;;
    restart)
      docker compose restart "${services[@]}"
      ;;
    status)
      docker compose ps "${services[@]}"
      ;;
    logs)
      docker compose logs -f "${services[@]}"
      ;;
    *)
      echo "Unknown action: ${action}" >&2
      usage >&2
      exit 2
      ;;
  esac
}

module="${1:-}"
action="${2:-}"

if [[ -z "${module}" || "${module}" == "-h" || "${module}" == "--help" ]]; then
  usage
  exit 0
fi

shift || true
if [[ -n "${action}" ]]; then
  shift || true
fi

case "${module}:${action}" in
  setup:run)
    exec ./setup.sh "$@"
    ;;
  infra:*)
    ensure_setup_files
    compose_for_services "${action}" "${INFRA_SERVICES[@]}" "$@"
    ;;
  vega:up)
    ensure_setup_files
    compose_for_services up "${INFRA_SERVICES[@]}"
    compose_for_services up kweaver-core-data-migrator "${VEGA_SERVICES[@]}" "$@"
    ;;
  vega:pull)
    ensure_setup_files
    compose_for_services pull kweaver-core-data-migrator "${VEGA_SERVICES[@]}" "$@"
    ;;
  vega:down)
    ensure_setup_files
    compose_for_services down "${VEGA_SERVICES[@]}" "$@"
    ;;
  vega:restart)
    ensure_setup_files
    compose_for_services restart "${VEGA_SERVICES[@]}" "$@"
    ;;
  vega:status)
    ensure_setup_files
    docker compose ps "${VEGA_SERVICES[@]}" "$@"
    ;;
  vega:logs)
    ensure_setup_files
    docker compose logs -f "${VEGA_SERVICES[@]}" "$@"
    ;;
  vega:*)
    echo "Unknown vega action: ${action}" >&2
    usage >&2
    exit 2
    ;;
  app:*)
    ensure_setup_files
    compose_for_services "${action}" "${APP_SERVICES[@]}" "$@"
    ;;
  all:up)
    ensure_setup_files
    compose_for_services up "${INFRA_SERVICES[@]}"
    compose_for_services up "${APP_SERVICES[@]}" "$@"
    ;;
  all:pull)
    ensure_setup_files
    compose_for_services pull "${INFRA_SERVICES[@]}" "${APP_SERVICES[@]}" "$@"
    ;;
  all:down)
    ensure_setup_files
    docker compose down "$@"
    ;;
  all:restart)
    ensure_setup_files
    compose_for_services restart "${INFRA_SERVICES[@]}" "${APP_SERVICES[@]}" "$@"
    ;;
  all:status)
    ensure_setup_files
    docker compose ps "$@"
    ;;
  all:logs)
    ensure_setup_files
    docker compose logs -f "$@"
    ;;
  *)
    echo "Unknown command: ${module} ${action}" >&2
    usage >&2
    exit 2
    ;;
esac
