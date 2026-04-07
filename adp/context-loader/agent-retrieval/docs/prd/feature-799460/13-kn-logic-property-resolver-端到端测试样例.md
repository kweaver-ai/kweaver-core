# kn-logic-property-resolver 端到端测试样例

## 文档说明

本文档提供 `kn-logic-property-resolver` 接口的人工测试样例，用于端到端调试。

**测试目标**：验证 Agent 能否基于 query 和 additional_context 正确生成 dynamic_params，并成功调用底层逻辑属性查询接口。

**样例结构**：
- **请求示例**：完整的 HTTP 请求体（可直接复制到 Postman/curl）
- **期望行为**：描述预期的参数生成逻辑和返回结果
- **验证要点**：重点检查的字段和值

**接口地址**：`POST /api/agent-retrieval/in/v1/kn/logic-property-resolver`

**必需 Headers**：
```
x-account-id: test_user_001
x-account-type: user
Content-Type: application/json
```

---

## 测试样例清单

- [样例 1：Metric - 即时查询（当前值）](#样例-1metric---即时查询当前值)
- [样例 2：Metric - 即时查询（带时间范围：最近一年的总数）](#样例-2metric---即时查询带时间范围最近一年的总数)
- [样例 3：Metric - 趋势查询（最近半年，按月）](#样例-3metric---趋势查询最近半年按月)
- [样例 4：Metric - 趋势查询（最近 4 周，按周）](#样例-4metric---趋势查询最近-4-周按周)
- [样例 5：Operator - 简单布尔参数](#样例-5operator---简单布尔参数)
- [样例 6：Operator - 从 object_instances 提取参数](#样例-6operator---从-object_instances-提取参数)
- [样例 7：Operator - 复杂对象参数](#样例-7operator---复杂对象参数)
- [样例 8：Operator - 数组参数](#样例-8operator---数组参数)
- [样例 9：Operator - 复杂嵌套参数（对象 + 对象数组）](#样例-9operator---复杂嵌套参数对象--对象数组)
- [样例 10：混合查询（Metric + Operator）](#样例-10混合查询metric--operator)
- [样例 11：批量查询（多对象，同属性）](#样例-11批量查询多对象同属性)
- [样例 12：缺参场景 - Metric 缺时间范围](#样例-12缺参场景---metric-缺时间范围)
- [样例 13：缺参场景 - Operator 缺必填参数](#样例-13缺参场景---operator-缺必填参数)

---

## 样例 1：Metric - 即时查询（当前值）

### 场景描述
查询企业当前的药品上市数量（最新值），不指定时间范围，走底层默认（最近 30 天）。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "查询当前药品上市数量（最新值）",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["approved_drug_count"],
  "additional_context": "{\"now_ms\":1762996342241,\"timezone\":\"Asia/Shanghai\"}",
  "options": {
    "return_debug": true,
    "max_repair_rounds": 1
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "approved_drug_count": {
    "instant": true
  }
}
```

**验证要点**：
- ✅ `instant` 必须为 `true`（关键词："当前"、"最新值"）
- ✅ 不应输出 `step`（instant=true 时禁止 step）
- ✅ 不输出 `start/end`（走底层默认最近 30 天）
- ✅ 返回结果包含单个聚合值（不是时间序列）

---

## 样例 2：Metric - 即时查询（带时间范围：最近一年的总数）

### 场景描述
查询最近一年的药品上市总数，虽然包含时间范围，但"总数"关键词强制 `instant=true`。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "查询最近一年的药品上市总数",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["approved_drug_count"],
  "additional_context": "{\"now_ms\":1762996342241,\"timezone\":\"Asia/Shanghai\"}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "approved_drug_count": {
    "instant": true,
    "start": 1731460342241,
    "end": 1762996342241
  }
}
```

**关键时间计算**：
- `now_ms`: 1762996342241（2025-12-18）
- `start`: now_ms - 1年 = 1731460342241（2024-12-18）
- `end`: now_ms = 1762996342241

**验证要点**：
- ✅ `instant` 必须为 `true`（关键词："总数"）
- ✅ 必须输出 `start/end`（定义聚合统计的时间范围）
- ✅ 不应输出 `step`
- ⚠️ **重点对比**：虽然包含"最近一年"，但因为是"总数"（聚合值），所以是 instant=true，而非趋势查询

---

## 样例 3：Metric - 趋势查询（最近半年，按月）

### 场景描述
查询最近半年的药品上市数量趋势，按月统计（返回时间序列）。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "统计最近半年药品上市数量趋势（按月）",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["approved_drug_count"],
  "additional_context": "{\"now_ms\":1762996342241,\"timezone\":\"Asia/Shanghai\"}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "approved_drug_count": {
    "instant": false,
    "start": 1747048342241,
    "end": 1762996342241,
    "step": "month"
  }
}
```

**关键时间计算**：
- `now_ms`: 1762996342241（2025-12-18）
- `start`: now_ms - 6个月 ≈ 1747048342241（2025-06-18）
- `end`: now_ms = 1762996342241
- `step`: "month"（从 query 中提取："按月"）

**验证要点**：
- ✅ `instant` 必须为 `false`（关键词："趋势"、"按月"）
- ✅ 必须输出 `step="month"`
- ✅ 必须输出 `start/end`
- ✅ 返回结果是时间序列（多个时间点的值）

---

## 样例 4：Metric - 趋势查询（最近 4 周，按周）

### 场景描述
查询最近 4 周的药品上市数量变化，按周统计。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "近4周每周的药品上市数量变化",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["approved_drug_count"],
  "additional_context": "{\"now_ms\":1762996342241,\"timezone\":\"Asia/Shanghai\"}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "approved_drug_count": {
    "instant": false,
    "start": 1760577142241,
    "end": 1762996342241,
    "step": "week"
  }
}
```

**关键时间计算**：
- `now_ms`: 1762996342241
- `start`: now_ms - 4周 = 1760577142241
- `end`: now_ms
- `step`: "week"（从 query 中提取："每周"）

**验证要点**：
- ✅ `instant` 必须为 `false`（关键词："每周"、"变化"）
- ✅ `step="week"`
- ✅ 返回 4 个数据点（每周一个）

---

## 样例 5：Operator - 简单布尔参数

### 场景描述
查询企业健康度评分，需要详细明细，中文返回。参数从 query 中抽取。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "查询企业健康度评分，需要详细明细，中文返回",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["business_health_score"],
  "additional_context": "{\"object_instances\":[{\"company_id\":\"company_000001\",\"company_name\":\"某药企A\"}]}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "business_health_score": {
    "include_details": true,
    "lang": "zh-CN"
  }
}
```

**参数来源分析**：
- `include_details`: true（从 query 中抽取："需要详细明细"）
- `lang`: "zh-CN"（从 query 中抽取："中文返回"）

**验证要点**：
- ✅ `include_details` 必须是布尔值 `true`（不是字符串 "true"）
- ✅ `lang` 正确识别为 "zh-CN"
- ✅ 不应输出 `company_id`（value_from="property" 的参数不生成）

---

## 样例 6：Operator - 从 object_instances 提取参数

### 场景描述
深度分析企业风险，参数 `company_name` 从 additional_context 中的 object_instances 提取。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "深度分析企业风险",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["company_risk_analysis"],
  "additional_context": "{\"object_instances\":[{\"company_id\":\"company_000001\",\"company_name\":\"某药企A\",\"registered_capital\":2000000}]}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "company_risk_analysis": {
    "company_name": "某药企A",
    "analysis_depth": "deep"
  }
}
```

**参数来源分析**：
- `company_name`: "某药企A"（从 additional_context 的 object_instances 提取）
- `analysis_depth`: "deep"（从 query 中抽取："深度分析"）

**验证要点**：
- ✅ 能正确从 additional_context 的 JSON 结构中提取 `company_name`
- ✅ 能从 query 的语义中推断 `analysis_depth="deep"`

---

## 样例 7：Operator - 复杂对象参数

### 场景描述
分析企业 2020-2023 年的化学药品审批情况，参数 `filter` 是一个对象（包含嵌套字段）。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "分析该企业2020-2023年的化学药品审批情况",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["drug_approval_analysis"],
  "additional_context": "{\"object_instances\":[{\"company_id\":\"company_000001\",\"company_name\":\"某药企A\"}]}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "drug_approval_analysis": {
    "filter": {
      "drug_type": "化学药品",
      "approval_year_start": 2020,
      "approval_year_end": 2023
    }
  }
}
```

**参数来源分析**：
- `filter` 是一个完整对象（不应拆散为多个参数）
- `drug_type`: "化学药品"（从 query 提取）
- `approval_year_start`: 2020（从 query 提取："2020-2023年"）
- `approval_year_end`: 2023

**验证要点**：
- ✅ `filter` 必须作为一个完整对象输出（不能打平）
- ✅ 对象内部字段类型正确（year 是 int，不是 string）
- ✅ 能正确解析 query 中的时间范围

---

## 样例 8：Operator - 数组参数

### 场景描述
分析项目 proj001 和 proj002，参数 `project_ids` 是字符串数组。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "分析项目proj001和proj002",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["multi_project_analysis"],
  "additional_context": "{\"object_instances\":[{\"company_id\":\"company_000001\",\"company_name\":\"某药企A\"}]}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "multi_project_analysis": {
    "project_ids": ["proj001", "proj002"]
  }
}
```

**参数来源分析**：
- `project_ids`: 从 query 中提取多个项目 ID，输出为数组格式

**验证要点**：
- ✅ `project_ids` 必须是数组格式（不是逗号分隔的字符串）
- ✅ 数组元素顺序与 query 中一致
- ✅ 即使只有一个值，也应输出数组格式 `["proj001"]`

---

## 样例 9：Operator - 复杂嵌套参数（对象 + 对象数组）

### 场景描述
复杂算子演示，包含嵌套对象 `filter` 和对象数组 `items`，验证 Agent 能否正确处理多层嵌套结构。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "按2020-2023年、地区北京/上海过滤，并给两个加权条目，中文返回",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["complex_operator_demo"],
  "additional_context": "{\"object_instances\":[{\"company_id\":\"company_000001\",\"company_name\":\"某药企A\"}]}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "complex_operator_demo": {
    "filter": {
      "year_range": {
        "start": 2020,
        "end": 2023
      },
      "regions": ["北京", "上海"]
    },
    "items": [
      { "id": "a", "weight": 0.6 },
      { "id": "b", "weight": 0.4 }
    ],
    "lang": "zh-CN"
  }
}
```

**参数来源分析**：
- `filter`: 嵌套对象（包含 `year_range` 对象和 `regions` 数组）
- `items`: 对象数组（每个对象包含 `id` 和 `weight`）
- `lang`: "zh-CN"（从 query 提取："中文返回"）

**验证要点**：
- ✅ 嵌套结构完整（不打平，不按 body/query/header 分组）
- ✅ 对象数组每个元素的字段类型正确（weight 是 number）
- ✅ 多层嵌套的 JSON 结构合法可解析

---

## 样例 10：混合查询（Metric + Operator）

### 场景描述
同时查询企业的药品上市数量（metric）和健康度评分（operator）。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "查询企业的药品上市情况和健康度评分",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["approved_drug_count", "business_health_score"],
  "additional_context": "{\"now_ms\":1762996342241,\"timezone\":\"Asia/Shanghai\",\"object_instances\":[{\"company_id\":\"company_000001\",\"company_name\":\"某药企A\"}]}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "approved_drug_count": {
    "instant": false,
    "start": 1755216342241,
    "end": 1762996342241,
    "step": "month"
  },
  "business_health_score": {
    "include_details": false,
    "lang": "zh-CN"
  }
}
```

**参数来源分析**：
- `approved_drug_count`: 走默认兜底策略（近 3 个月，按月）
- `business_health_score`: 使用默认参数（不需要详细明细，中文）

**验证要点**：
- ✅ 能同时处理 metric 和 operator 两种类型
- ✅ 两个属性的参数生成互不干扰
- ✅ 返回结果包含两个属性的值（数据结构不同）

---

## 样例 11：批量查询（多对象，同属性）

### 场景描述
对多个企业批量查询同一属性（药品上市数量），使用相同的参数。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "对这批企业统计最近一年药品上市数量趋势（按季度）",
  "unique_identities": [
    { "company_id": "company_000001" },
    { "company_id": "company_000002" },
    { "company_id": "company_000003" }
  ],
  "properties": ["approved_drug_count"],
  "additional_context": "{\"now_ms\":1762996342241,\"timezone\":\"Asia/Shanghai\"}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**Agent 应生成的 dynamic_params**：
```json
{
  "approved_drug_count": {
    "instant": false,
    "start": 1731456342241,
    "end": 1762996342241,
    "step": "quarter"
  }
}
```

**验证要点**：
- ✅ 三个对象使用相同的 dynamic_params（不需要为每个对象单独生成）
- ✅ 返回结果 `datas` 数组包含 3 个元素（与 unique_identities 对齐）
- ✅ 每个对象的 `company_id` 正确对应

---

## 样例 12：缺参场景 - Metric 缺时间范围

### 场景描述
查询药品上市数量趋势，但既没有提供 now_ms，也没有明确时间范围，Agent 无法生成 start/end。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "统计药品上市数量趋势（按月）",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["approved_drug_count"],
  "additional_context": "",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**应返回缺参错误**：
```json
{
  "error_code": "MISSING_INPUT_PARAMS",
  "message": "dynamic_params 缺少必需的 input 参数",
  "missing": [
    {
      "property": "approved_drug_count",
      "params": [
        {
          "name": "now_ms",
          "type": "INTEGER",
          "hint": "需要 now_ms (或在 additional_context 中提供 end/start)"
        }
      ]
    }
  ],
  "trace_id": "xxxx-xxxx-xxxx-xxxx"
}
```

**验证要点**：
- ✅ 返回 `error_code: MISSING_INPUT_PARAMS`
- ✅ `missing` 数组包含缺失参数的详细信息
- ✅ `hint` 提供明确的补参建议

---

## 样例 13：缺参场景 - Operator 缺必填参数

### 场景描述
分析企业药品审批情况，但 query 中没有提到药品类型，Agent 无法确定 `drug_type` 参数。

### 请求示例

```json
{
  "kn_id": "kn_medical",
  "ot_id": "company",
  "query": "分析该企业的药品审批情况",
  "unique_identities": [
    { "company_id": "company_000001" }
  ],
  "properties": ["drug_approval_analysis"],
  "additional_context": "{\"object_instances\":[{\"company_id\":\"company_000001\",\"company_name\":\"某药企A\"}]}",
  "options": {
    "return_debug": true
  }
}
```

### 期望行为

**应返回缺参错误**：
```json
{
  "error_code": "MISSING_INPUT_PARAMS",
  "message": "dynamic_params 缺少必需的 input 参数",
  "missing": [
    {
      "property": "drug_approval_analysis",
      "params": [
        {
          "name": "drug_type",
          "type": "STRING",
          "hint": "请指定药品类型（化学药品/生物制品/中药）"
        }
      ]
    }
  ],
  "trace_id": "xxxx-xxxx-xxxx-xxxx"
}
```

**验证要点**：
- ✅ Agent 识别到无法从 query/additional_context 推断 `drug_type`
- ✅ 返回缺参结构（而非强行猜测或使用空值）
- ✅ `hint` 提供可选值范围，便于用户补充

---

## 附录：cURL 命令模板

### 基础命令模板

```bash
curl -X POST "http://localhost:8080/api/agent-retrieval/in/v1/kn/logic-property-resolver" \
  -H "Content-Type: application/json" \
  -H "x-account-id: test_user_001" \
  -H "x-account-type: user" \
  -d '{
    "kn_id": "kn_medical",
    "ot_id": "company",
    "query": "查询企业的药品上市数量",
    "unique_identities": [
      { "company_id": "company_000001" }
    ],
    "properties": ["approved_drug_count"],
    "additional_context": "{\"now_ms\":1762996342241,\"timezone\":\"Asia/Shanghai\"}",
    "options": {
      "return_debug": true,
      "max_repair_rounds": 1,
      "max_concurrency": 4
    }
  }' | jq '.'
```

### 测试脚本示例

参考项目中的 `test-metric-logic-property.sh`，可以批量执行多个测试用例。

---

## 测试检查清单

### Metric 参数生成检查

- [ ] instant=true 场景：正确识别"当前/最新/总数"关键词
- [ ] instant=false 场景：正确识别"趋势/变化/按XX"关键词
- [ ] 时间范围计算：start/end 时间戳正确（误差±1天内）
- [ ] step 枚举：只能是 day/week/month/quarter/year
- [ ] instant/step 互斥：instant=true 时不输出 step
- [ ] 默认兜底：无明确时间时使用近 3 个月 + month

### Operator 参数生成检查

- [ ] 基础类型：布尔值是 true/false（非字符串）
- [ ] 对象参数：作为完整对象输出（不打平）
- [ ] 数组参数：输出 JSON 数组格式（非逗号分隔）
- [ ] 嵌套结构：多层嵌套保持完整（不按 body/query/header 分组）
- [ ] 参数来源：正确从 additional_context/object_instances/query 提取
- [ ] 缺参处理：无法确定时返回 _error（不强行猜测）

### 接口响应检查

- [ ] 成功响应：`datas` 数组长度与 `unique_identities` 一致
- [ ] 缺参响应：包含 `error_code`, `message`, `missing`, `trace_id`
- [ ] debug 信息：`return_debug=true` 时包含 `dynamic_params`
- [ ] 数据结构：metric 返回时序数据，operator 返回业务对象

---

## 常见问题（FAQ）

### Q1: 为什么"最近一年的总数"是 instant=true？

**A**: "总数"表示要一个聚合值（单个结果），即使统计的是一年的数据，返回的也是一个数字，而非时间序列。
- `instant=true, start/end=最近一年` → 返回：`{"value": 123}`
- `instant=false, start/end=最近一年, step=month` → 返回：12 个月的数据点

### Q2: additional_context 应该传 JSON 还是文本？

**A**: 推荐传 JSON 字符串（更精确），但也支持自由文本。
- JSON 格式：`{"now_ms":1762996342241,"instant":true}`
- 文本格式：`"时间范围：2024年6月至2024年12月，按月统计"`

### Q3: 批量查询时，每个对象能用不同的参数吗？

**A**: 本接口不支持。所有对象使用相同的 dynamic_params。如果需要为不同对象使用不同参数，需要分多次调用。

### Q4: 缺参时，能否让 Agent 自动使用默认值？

**A**: metric 的时间参数有默认策略（如近 3 个月），但 operator 的业务参数通常没有默认值，必须明确提供。

---

## 版本历史

- **v1.0**（2025-12-18）：初始版本，包含 13 个测试样例

---

**文档维护者**：Feature-799460 开发团队

