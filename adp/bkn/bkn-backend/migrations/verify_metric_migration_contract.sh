#!/usr/bin/env bash
# TDD / CI：指标表迁移文件存在（IMPLEMENTATION_PLAN Task 2）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fail() { echo "verify_metric_migration_contract: $*" >&2; exit 1; }

M1="${ROOT}/mariadb/0.7.0/01-metric_definition.sql"
M2="${ROOT}/dm8/0.7.0/01-metric_definition.sql"
DM_INIT="${ROOT}/dm8/0.7.0/init.sql"
[[ -f "${M1}" ]] || fail "missing ${M1}"
[[ -f "${M2}" ]] || fail "missing ${M2}"
[[ -f "${DM_INIT}" ]] || fail "missing ${DM_INIT} (达梦 0.7.0 需提供 init.sql 全量初始化脚本)"
grep -q 't_metric_definition' "${M1}" || fail "mariadb migration must create t_metric_definition"
grep -q 'f_metric_type' "${M1}" || fail "mariadb migration must include f_metric_type"
grep -q 't_metric_definition' "${M2}" || fail "dm8 01-metric_definition must create t_metric_definition"
grep -q 't_metric_definition' "${DM_INIT}" || fail "dm8 init.sql must include t_metric_definition"
grep -q 't_knowledge_network' "${DM_INIT}" || fail "dm8 init.sql should be full baseline (expect t_knowledge_network)"
echo "ok: metric_definition migration 0.7.0 present (dm8 init.sql + incremental)"
