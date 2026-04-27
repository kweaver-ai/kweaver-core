#!/bin/bash

# kn-logic-property-resolver Metric 类型逻辑属性查询测试脚本

BASE_URL="${BASE_URL:-http://localhost:8080}"
API_PATH="/api/kn/logic-property-resolver"

echo "========================================"
echo "测试 kn-logic-property-resolver"
echo "========================================"
echo ""

# 测试用例 1: 成功查询 metric 类型逻辑属性
echo "【测试 1】成功查询 - metric 类型（趋势查询，带 step）"
echo "----------------------------------------"
curl -s -X POST "${BASE_URL}${API_PATH}" \
  -H "Content-Type: application/json" \
  -H "x-account-id: test_account" \
  -H "x-account-type: user" \
  -H "x-kn-id: kn_medical" \
  -d '{
    "kn_id": "kn_medical",
    "ot_id": "company",
    "query": "查询最近半年的药品上市数量",
    "unique_identities": [
      {"company_id": "company_000001"}
    ],
    "properties": ["approved_drug_count"],
    "additional_context": "时间范围：2024年6月至2024年12月，按月统计",
    "options": {
      "return_debug": true,
      "max_concurrency": 4
    }
  }' | jq '.'

echo ""
echo ""

# 测试用例 2: 即时查询（instant=true，无 step）
echo "【测试 2】成功查询 - metric 类型（即时查询，instant=true）"
echo "----------------------------------------"
curl -s -X POST "${BASE_URL}${API_PATH}" \
  -H "Content-Type: application/json" \
  -H "x-account-id: test_account" \
  -H "x-account-type: user" \
  -H "x-kn-id: kn_medical" \
  -d '{
    "kn_id": "kn_medical",
    "ot_id": "company",
    "query": "查询当前的药品上市数量",
    "unique_identities": [
      {"company_id": "company_000001"}
    ],
    "properties": ["approved_drug_count"],
    "additional_context": "即时查询，查询当前时刻的值",
    "options": {
      "return_debug": true
    }
  }' | jq '.'

echo ""
echo ""

# 测试用例 3: 缺参测试
echo "【测试 3】缺参情况 - 缺少时间范围"
echo "----------------------------------------"
curl -s -X POST "${BASE_URL}${API_PATH}" \
  -H "Content-Type: application/json" \
  -H "x-account-id: test_account" \
  -H "x-account-type: user" \
  -H "x-kn-id: kn_medical" \
  -d '{
    "kn_id": "kn_medical",
    "ot_id": "company",
    "query": "查询药品上市数量",
    "unique_identities": [
      {"company_id": "company_000001"}
    ],
    "properties": ["approved_drug_count"],
    "additional_context": "",
    "options": {
      "return_debug": true
    }
  }' | jq '.'

echo ""
echo ""

# 测试用例 4: 批量查询多个对象实例
echo "【测试 4】批量查询 - 多个对象实例"
echo "----------------------------------------"
curl -s -X POST "${BASE_URL}${API_PATH}" \
  -H "Content-Type: application/json" \
  -H "x-account-id: test_account" \
  -H "x-account-type: user" \
  -H "x-kn-id: kn_medical" \
  -d '{
    "kn_id": "kn_medical",
    "ot_id": "company",
    "query": "查询多个企业的药品上市情况",
    "unique_identities": [
      {"company_id": "company_000001"},
      {"company_id": "company_000002"},
      {"company_id": "company_000003"}
    ],
    "properties": ["approved_drug_count"],
    "additional_context": "时间范围：最近一年",
    "options": {
      "return_debug": true
    }
  }' | jq '.'

echo ""
echo ""

# 测试用例 5: 混合查询 metric + operator
echo "【测试 5】混合查询 - metric + operator"
echo "----------------------------------------"
curl -s -X POST "${BASE_URL}${API_PATH}" \
  -H "Content-Type: application/json" \
  -H "x-account-id: test_account" \
  -H "x-account-type: user" \
  -H "x-kn-id: kn_medical" \
  -d '{
    "kn_id": "kn_medical",
    "ot_id": "company",
    "query": "查询企业的药品上市情况和健康度评分",
    "unique_identities": [
      {"company_id": "company_000001"}
    ],
    "properties": ["approved_drug_count", "business_health_score"],
    "additional_context": "时间范围：最近半年",
    "options": {
      "return_debug": true
    }
  }' | jq '.'

echo ""
echo "========================================"
echo "测试完成"
echo "========================================"

