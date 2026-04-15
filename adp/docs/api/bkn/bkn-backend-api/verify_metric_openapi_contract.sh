#!/usr/bin/env bash
# TDD / CI: BKN 原生指标 OpenAPI 契约自检（见 IMPLEMENTATION_PLAN Task 1）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
METRICS_YAML="${ROOT}/bkn-metrics.yaml"
OQ_YAML="${ROOT}/../ontology-query-ai/ontology-query.yaml"

fail() { echo "verify_metric_openapi_contract: $*" >&2; exit 1; }

[[ -f "${METRICS_YAML}" ]] || fail "missing ${METRICS_YAML}"
[[ -f "${OQ_YAML}" ]] || fail "missing ${OQ_YAML}"

grep -qE '/api/bkn-backend/v1/knowledge-networks/\{kn_id\}/metrics' "${METRICS_YAML}" \
  || fail "bkn-metrics.yaml: expected metrics collection path"
grep -q 'metric_type' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected metric_type in schema"
grep -q 'ListMetrics:' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected ListMetrics schema (entries/total_count)"
grep -q 'entries:' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected entries field"
grep -q 'total_count:' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected total_count field"
grep -q 'ReqMetrics:' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected ReqMetrics batch schema"
grep -q 'x-http-method-override' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected POST override header"
grep -q '/metrics/validation' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected metrics validation path"
grep -q 'MetricSearchResponse:' "${METRICS_YAML}" || fail "bkn-metrics.yaml: expected MetricSearchResponse"

grep -qE '/api/ontology-query/v1/knowledge-networks/\{kn_id\}/metrics/\{metric_id\}/data' "${OQ_YAML}" \
  || fail "ontology-query.yaml: expected POST .../metrics/{metric_id}/data"
grep -qE '/api/ontology-query/v1/knowledge-networks/\{kn_id\}/metrics/dry-run' "${OQ_YAML}" \
  || fail "ontology-query.yaml: expected POST .../metrics/dry-run"
grep -q 'MetricData:' "${OQ_YAML}" || fail "ontology-query.yaml: expected components.schemas.MetricData"
grep -q 'MetricQueryRequestBody:' "${OQ_YAML}" || fail "ontology-query.yaml: expected MetricQueryRequestBody"
grep -q 'MetricDryRun:' "${OQ_YAML}" || fail "ontology-query.yaml: expected MetricDryRun"

echo "ok: metric OpenAPI contract files present and key paths/schemas referenced"
