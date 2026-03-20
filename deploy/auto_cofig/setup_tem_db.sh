#!/usr/bin/env bash
# 在 demo 命名空间创建独立的 MariaDB 实例，导入 tem 数据库，并更新 config.env。
# 供 KWeaver 从集群内扫描使用（DS_HOST 为集群内地址）。
# 兼容在非 bash 的 shell（如 sh/dash）里误执行的场景：
if [ -z "${BASH_VERSION:-}" ]; then
  if command -v bash >/dev/null 2>&1; then
    exec bash "$0" "$@"
  fi
  echo "ERROR: 当前环境没有 bash，无法执行该脚本。" >&2
  echo "请在 Git Bash 或 WSL 中运行：bash ./setup_tem_db.sh" >&2
  exit 1
fi

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONF_YAML="${ROOT_DIR}/conf/config.yaml"
SQL_FILE="${ROOT_DIR}/auto_cofig/dump-tem.sql"
CONFIG_ENV_FILE="${ROOT_DIR}/auto_cofig/config.env"
CHART_DIR="${ROOT_DIR}/charts/proton-mariadb"

NAMESPACE_DEMO="${NAMESPACE_DEMO:-demo}"
RELEASE_NAME="${RELEASE_NAME:-demo-mariadb}"
RDS_USER_DEFAULT="adp"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: 缺少命令：$1" >&2
    exit 1
  }
}

generate_random_password() {
  local len="${1:-10}"
  openssl rand -base64 32 2>/dev/null | tr -dc 'a-zA-Z0-9' | head -c "$len"
}

escape_squote() {
  local s="${1:-}"
  printf "%s" "$s" | sed "s/'/'\\\\''/g"
}

update_env_kv() {
  local file="$1"
  local key="$2"
  local val="$3"
  local esc
  esc="$(escape_squote "$val")"
  awk -v k="$key" -v v="$esc" '
    BEGIN { found=0 }
    $0 ~ ("^" k "=") { print k "='\''" v "'\''"; found=1; next }
    { print }
    END { if (!found) print k "='\''" v "'\''" }
  ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
}

need_cmd kubectl
need_cmd awk
need_cmd grep

if [[ ! -f "$CONFIG_ENV_FILE" ]]; then
  echo "ERROR: 找不到配置文件：$CONFIG_ENV_FILE" >&2
  exit 1
fi
if [[ ! -f "$SQL_FILE" ]]; then
  echo "ERROR: 找不到 SQL 文件：$SQL_FILE" >&2
  exit 1
fi

# 创建 demo 命名空间
echo "创建命名空间：${NAMESPACE_DEMO}"
kubectl create namespace "${NAMESPACE_DEMO}" 2>/dev/null || true

# 检查是否已部署
if helm status "${RELEASE_NAME}" -n "${NAMESPACE_DEMO}" >/dev/null 2>&1; then
  echo "MariaDB 已在 ${NAMESPACE_DEMO} 中部署（release: ${RELEASE_NAME}），跳过安装。"
  echo "将从 Secret 读取凭据并更新 config.env。"
  POD_NAME="$(kubectl get pod -n "${NAMESPACE_DEMO}" -l "app.kubernetes.io/instance=${RELEASE_NAME}" --no-headers 2>/dev/null | awk 'NR==1{print $1}')"
  if [[ -z "${POD_NAME}" ]]; then
    POD_NAME="$(kubectl get pod -n "${NAMESPACE_DEMO}" -l "app=mariadb" --no-headers 2>/dev/null | awk 'NR==1{print $1}')"
  fi
  if [[ -z "${POD_NAME}" ]]; then
    echo "ERROR: 未找到 MariaDB Pod。" >&2
    exit 1
  fi
  RDS_PASSWORD="$(kubectl get secret -n "${NAMESPACE_DEMO}" "${RELEASE_NAME}-proton-mariadb-auth" -o jsonpath='{.data.mariadb-password}' 2>/dev/null | base64 -d 2>/dev/null || true)"
  if [[ -z "${RDS_PASSWORD}" ]]; then
    RDS_PASSWORD="$(kubectl get secret -n "${NAMESPACE_DEMO}" "${RELEASE_NAME}-proton-mariadb-auth" -o jsonpath='{.data.mariadb\.auth\.password}' 2>/dev/null | base64 -d 2>/dev/null || true)"
  fi
  if [[ -z "${RDS_PASSWORD}" ]]; then
    RDS_PASSWORD="$(kubectl get secret -n "${NAMESPACE_DEMO}" "${RELEASE_NAME}-proton-mariadb-auth" -o jsonpath='{.data.password}' 2>/dev/null | base64 -d 2>/dev/null || true)"
  fi
  if [[ -z "${RDS_PASSWORD}" ]]; then
    echo "ERROR: 无法从 Secret 读取密码，请手动设置或删除 release 后重试。" >&2
    exit 1
  fi
  RDS_USER="${RDS_USER:-$RDS_USER_DEFAULT}"
else
  # 生成密码
  RDS_USER="${RDS_USER:-$RDS_USER_DEFAULT}"
  RDS_PASSWORD="${RDS_PASSWORD:-$(generate_random_password 10)}"
  RDS_ROOT_PASSWORD="${RDS_ROOT_PASSWORD:-$(generate_random_password 10)}"

  need_cmd helm

  if [[ ! -d "$CHART_DIR" ]]; then
    echo "ERROR: 找不到 Helm chart：$CHART_DIR" >&2
    exit 1
  fi

  echo "在 ${NAMESPACE_DEMO} 中部署 MariaDB（release: ${RELEASE_NAME}）"
  helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
    --namespace "${NAMESPACE_DEMO}" \
    --set mariadb.namespace="${NAMESPACE_DEMO}" \
    --set mariadb.auth.rootPassword="${RDS_ROOT_PASSWORD}" \
    --set mariadb.auth.database="tem" \
    --set mariadb.auth.username="${RDS_USER}" \
    --set mariadb.auth.password="${RDS_PASSWORD}" \
    --set mariadb.persistence.enabled=false \
    --wait --timeout=300s

  echo "等待 MariaDB Pod 就绪..."
  kubectl wait --for=condition=ready pod -l "app.kubernetes.io/instance=${RELEASE_NAME}" -n "${NAMESPACE_DEMO}" --timeout=60s 2>/dev/null || true

  POD_NAME="$(kubectl get pod -n "${NAMESPACE_DEMO}" -l "app.kubernetes.io/instance=${RELEASE_NAME}" --no-headers 2>/dev/null | awk 'NR==1{print $1}')"
  if [[ -z "${POD_NAME}" ]]; then
    POD_NAME="$(kubectl get pod -n "${NAMESPACE_DEMO}" -l "app=mariadb" --no-headers 2>/dev/null | awk 'NR==1{print $1}')"
  fi
  if [[ -z "${POD_NAME}" ]]; then
    echo "ERROR: 未找到 MariaDB Pod。" >&2
    exit 1
  fi
fi

echo "使用 Pod：${POD_NAME}"

# 探测 Pod 内数据库客户端
DB_CLI="$(
  kubectl exec -n "${NAMESPACE_DEMO}" "${POD_NAME}" -- sh -lc '
    if command -v mariadb >/dev/null 2>&1; then echo mariadb; elif command -v mysql >/dev/null 2>&1; then echo mysql; else exit 1; fi
  ' 2>/dev/null || true
)"
if [[ -z "${DB_CLI}" ]]; then
  echo "ERROR: Pod 内未找到 mariadb/mysql 客户端命令。" >&2
  exit 1
fi
echo "使用客户端：${DB_CLI}"

# 创建 tem 数据库并导入（若已存在则跳过）
echo "创建数据库 tem 并导入 SQL..."
kubectl exec -n "${NAMESPACE_DEMO}" "${POD_NAME}" -- sh -lc \
  "${DB_CLI} -u '${RDS_USER}' -p'${RDS_PASSWORD}' -e \"CREATE DATABASE IF NOT EXISTS tem DEFAULT CHARACTER SET utf8 COLLATE utf8_unicode_ci;\""

kubectl exec -i -n "${NAMESPACE_DEMO}" "${POD_NAME}" -- sh -lc \
  "${DB_CLI} -u '${RDS_USER}' -p'${RDS_PASSWORD}' tem" < "${SQL_FILE}"

echo "校验数据库 tem 是否存在、表数量是否 > 0"
kubectl exec -n "${NAMESPACE_DEMO}" "${POD_NAME}" -- sh -lc \
  "${DB_CLI} -u '${RDS_USER}' -p'${RDS_PASSWORD}' -N -e \"SHOW DATABASES LIKE 'tem';\" | grep -x tem"

TABLE_COUNT="$(
  kubectl exec -n "${NAMESPACE_DEMO}" "${POD_NAME}" -- sh -lc \
    "${DB_CLI} -u '${RDS_USER}' -p'${RDS_PASSWORD}' -N -e \"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='tem';\""
)"
echo "tem 表数量：${TABLE_COUNT}"

if [[ "${TABLE_COUNT}" -eq 0 ]]; then
  echo "ERROR: tem 数据库表数量为 0，可能导入失败。" >&2
  exit 1
fi

# 写入 config.env：使用集群内地址，便于 KWeaver 扫描
DS_HOST="${RELEASE_NAME}-proton-mariadb.${NAMESPACE_DEMO}.svc.cluster.local"
if ! kubectl get svc -n "${NAMESPACE_DEMO}" "${RELEASE_NAME}-proton-mariadb" >/dev/null 2>&1; then
  DS_HOST="${RELEASE_NAME}.${NAMESPACE_DEMO}.svc.cluster.local"
  if ! kubectl get svc -n "${NAMESPACE_DEMO}" "${RELEASE_NAME}" >/dev/null 2>&1; then
    SVC_NAME="$(kubectl get svc -n "${NAMESPACE_DEMO}" --no-headers 2>/dev/null | head -1 | awk '{print $1}')"
    DS_HOST="${SVC_NAME}.${NAMESPACE_DEMO}.svc.cluster.local"
  fi
fi

echo "写入 ${CONFIG_ENV_FILE}（DS_HOST 为集群内地址）"
update_env_kv "${CONFIG_ENV_FILE}" "DS_HOST" "${DS_HOST}"
update_env_kv "${CONFIG_ENV_FILE}" "DS_PORT" "3306"
update_env_kv "${CONFIG_ENV_FILE}" "DS_DATABASE_NAME" "tem"
update_env_kv "${CONFIG_ENV_FILE}" "DS_USERNAME" "${RDS_USER}"
update_env_kv "${CONFIG_ENV_FILE}" "DS_PASSWORD" "${RDS_PASSWORD}"

echo "完成：tem 数据库已创建并导入成功。"
echo "DS_HOST=${DS_HOST}（集群内地址，供 KWeaver 扫描）"
