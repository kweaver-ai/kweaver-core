# dynamic_params 生成规则库

## 1. 概述

本文档定义了 `kn-logic-property-resolver` 工具中 `dynamic_params` 生成的完整规则集，涵盖参数来源规则、类型特定规则、抽取优先级、校验规则等核心约束。

**适用范围**：metric 和 operator 两种逻辑属性类型的 dynamic_params 生成

**核心目标**：确保大模型生成的 dynamic_params 严格符合业务规则和类型约束

---

## 2. 核心概念

### 2.1 参数来源（value_from）

每个参数通过 `value_from` 字段定义其值的来源方式：

| value_from | 说明 | 示例 |
|-----------|------|------|
| `input` | 由大模型从 query/additional_context 中抽取生成 | `instant`, `start`, `end`, `step` |
| `property` | 从对象实例的数据属性中获取 | `company_id`, `registered_capital` |
| `const` | 使用固定的常量值 | 预定义的默认值 |

**规则**：dynamic_params **只包含** `value_from="input"` 的参数，不包含 property 和 const 参数。

### 2.2 系统生成标记（if_system_generate）

标记参数是否必须由系统（大模型）生成：

- `true`：必须生成（如 metric 的 instant/start/end/step）
- `false` 或不设置：可选参数，根据上下文决定是否生成

### 2.3 逻辑属性类型（type）

| type | 说明 | 典型示例 |
|------|------|----------|
| `metric` | 指标属性（时序数据） | 药品上市数量、销售额趋势 |
| `operator` | 算子属性（复杂计算） | 健康度评分、风险等级 |

---

## 3. 参数来源规则

### 3.1 通用规则

**规则 3.1.1**：dynamic_params 只包含 `value_from="input"` 的参数

```json
// LogicPropertyDef 示例
{
  "name": "approved_drug_count",
  "type": "metric",
  "parameters": [
    { "name": "instant", "value_from": "input", "if_system_generate": true },
    { "name": "start", "value_from": "input", "if_system_generate": true },
    { "name": "end", "value_from": "input", "if_system_generate": true },
    { "name": "step", "value_from": "input", "if_system_generate": true },
    { "name": "company_id", "value_from": "property" },  // 不生成
    { "name": "metric_id", "value_from": "const", "value": "metric_001" }  // 不生成
  ]
}

// dynamic_params 输出（只包含 input 参数）
{
  "approved_drug_count": {
    "instant": false,
    "start": 1762996342241,
    "end": 1762996342241,
    "step": "month"
  }
}
```

**规则 3.1.2**：对所有 `value_from="input"` 的参数都必须尝试生成

- 能确定值就填入
- 不能确定则输出 `_error`（缺参错误）

### 3.2 参数值抽取优先级

**规则 3.2.1**：参数值抽取优先级（从高到低）

1. **additional_context**：优先从补充上下文中抽取
2. **query**：从用户问题中抽取
3. **schema 示例/默认值**：从 operator schema 或参数定义中获取
4. **推断生成**：基于上下文逻辑推断（如时间范围）

**规则 3.2.2**：additional_context 推荐格式

```json
{
  "company_id": "company_000001",
  "registered_capital": 2000000,
  "filter": "registered_capital > 1000000",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai",
  "step": "month",
  "instant": true
}
```

或自由文本格式：
```
company_id=company_000001，registered_capital=2000000；registered_capital>1000000；now_ms=1762996342241，timezone=Asia/Shanghai；趋势查询step=month
```

---

## 4. Metric 类型规则

### 4.1 系统固定参数

**规则 4.1.1**：metric 必须生成以下系统参数（`if_system_generate:true`）

| 参数名 | 类型 | 说明 | 是否必填 |
|--------|------|------|----------|
| `instant` | BOOLEAN | 即时查询标志 | 必填 |
| `start` | INTEGER | 时间范围开始（毫秒时间戳） | instant=false 时必填 |
| `end` | INTEGER | 时间范围结束（毫秒时间戳） | instant=false 时必填 |
| `step` | STRING | 时间步长 | instant=false 时必填 |

**规则 4.1.2**：instant 参数定义

- `instant=true`：即时查询（单点值）
  - 只需生成 `instant: true`
  - 不需要 start/end/step
- `instant=false`：趋势查询（时序数据）
  - 必须生成 `instant: false`
  - 必须生成 `start` 和 `end`
  - 必须生成 `step`

**规则 4.1.3**：step 参数枚举约束

step 只能取以下枚举值（固定 1 个单位）：

| 值 | 说明 | 示例场景 |
|----|------|----------|
| `day` | 天 | 日趋势 |
| `week` | 周 | 周趋势 |
| `month` | 月 | 月趋势 |
| `quarter` | 季度 | 季度趋势 |
| `year` | 年 | 年趋势 |

**禁止**：使用 `step=2`、`step=7d` 等非标准值

### 4.2 Metric 参数生成规则

**规则 4.2.1**：时间参数生成逻辑

1. **即时查询（instant=true）**
   ```json
   {
     "approved_drug_count": {
       "instant": true
     }
   }
   ```

2. **趋势查询（instant=false）**
   ```json
   {
     "approved_drug_count": {
       "instant": false,
       "start": 1762996342241 - 90*24*60*60*1000,
       "end": 1762996342241,
       "step": "month"
     }
   }
   ```

**规则 4.2.2**：时间范围推断规则

当 query 中包含时间描述时，按以下规则推断：

| query 描述 | 推断规则 | 示例 |
|-----------|----------|------|
| "当前/现在/今天" | instant=true | `{"instant": true}` |
| "最近N天/周/月/季/年" | instant=false, step=对应单位, start=now-N*单位, end=now | "最近3个月" → `{"instant": false, "start": now-90d, "end": now, "step": "month"}` |
| "N天/周/月/季/年前" | instant=false, step=对应单位, start=now-N*单位, end=now-N*单位 | "3个月前" → `{"instant": false, "start": now-90d, "end": now-90d, "step": "month"}` |
| "从X到Y" | instant=false, step=根据X-Y跨度推断, start=X, end=Y | "从2024年1月到3月" → `{"instant": false, "start": 2024-01-01, "end": 2024-03-31, "step": "month"}` |

**规则 4.2.3**：now_ms 默认行为

- 若调用方不传 `now_ms`，由服务端取当前时间戳（毫秒）
- 在 debug 信息中返回实际使用的 `now_ms` 值

### 4.3 Metric 业务参数规则

**规则 4.3.1**：除系统参数外，还需生成所有 `value_from="input"` 的业务参数

```json
// LogicPropertyDef 示例（带业务参数）
{
  "name": "drug_sales_by_region",
  "type": "metric",
  "parameters": [
    { "name": "instant", "value_from": "input", "if_system_generate": true },
    { "name": "start", "value_from": "input", "if_system_generate": true },
    { "name": "end", "value_from": "input", "if_system_generate": true },
    { "name": "step", "value_from": "input", "if_system_generate": true },
    { "name": "region", "value_from": "input", "type": "STRING" },
    { "name": "drug_category", "value_from": "input", "type": "STRING" }
  ]
}

// dynamic_params 输出
{
  "drug_sales_by_region": {
    "instant": false,
    "start": 1762996342241,
    "end": 1762996342241,
    "step": "month",
    "region": "华东地区",
    "drug_category": "抗肿瘤药"
  }
}
```

**规则 4.3.2**：业务参数缺参处理

- 若无法从上下文确定业务参数值，输出缺参错误：
  ```json
  {
    "_error": "missing drug_sales_by_region: region,drug_category | ask: 请明确查询的地区和药品类别"
  }
  ```

---

## 5. Operator 类型规则

### 5.1 参数来源与结构

**规则 5.1.1**：operator 参数可能来自 Path/Query/Body，但统一输出为 dynamic_params[property]

```json
// LogicPropertyDef 示例
{
  "name": "business_health_score",
  "type": "operator",
  "parameters": [
    { "name": "include_details", "type": "BOOLEAN", "value_from": "input" },
    { "name": "lang", "type": "STRING", "value_from": "input" },
    { "name": "filter", "type": "OBJECT", "value_from": "input" },
    { "name": "items", "type": "ARRAY", "value_from": "input" }
  ]
}

// dynamic_params 输出（统一格式，不区分 Path/Query/Body）
{
  "business_health_score": {
    "include_details": true,
    "lang": "zh-CN",
    "filter": {
      "year_range": [2023, 2024],
      "regions": ["华东", "华南"]
    },
    "items": ["revenue", "profit", "growth"]
  }
}
```

**规则 5.1.2**：禁止按 header/query/path/body 分组输出

- 错误示例：
  ```json
  {
    "business_health_score": {
      "query": { "lang": "zh-CN" },
      "body": { "include_details": true, "filter": {...} }
    }
  }
  ```
- 正确示例：
  ```json
  {
    "business_health_score": {
      "include_details": true,
      "lang": "zh-CN",
      "filter": {...}
    }
  }
  ```

### 5.2 参数类型约束

**规则 5.2.1**：严格遵守参数类型定义（logic_property.parameters[].type）

| 类型 | 说明 | 示例 |
|------|------|------|
| `STRING` | 字符串 | `"zh-CN"`, `"华东地区"` |
| `INTEGER` | 整数 | `2024`, `90` |
| `NUMBER` | 数字（含小数） | `3.14`, `0.95` |
| `BOOLEAN` | 布尔值 | `true`, `false` |
| `OBJECT` | JSON 对象 | `{"year": 2024, "region": "华东"}` |
| `ARRAY` | JSON 数组 | `[1, 2, 3]`, `["a", "b"]` |

**规则 5.2.2**：OBJECT 类型必须输出完整结构

```json
// 参数定义
{
  "name": "filter",
  "type": "OBJECT",
  "value_from": "input"
}

// 正确输出
{
  "filter": {
    "year_range": [2023, 2024],
    "regions": ["华东", "华南"]
  }
}

// 错误输出（不完整）
{
  "filter": "year_range=2023,2024"
}
```

**规则 5.2.3**：ARRAY 类型必须输出 JSON 数组格式

```json
// 参数定义
{
  "name": "items",
  "type": "ARRAY",
  "value_from": "input"
}

// 正确输出
{
  "items": ["revenue", "profit", "growth"]
}

// 错误输出（非数组）
{
  "items": "revenue,profit,growth"
}
```

### 5.3 Operator Schema 辅助

**规则 5.3.1**：推荐使用 operator schema 提升复杂参数生成准确率

当 operator 参数包含复杂结构（OBJECT/ARRAY）时，应从 `logic_property.data_source.id` 获取 `operator_id`，调用以下接口获取 schema：

```
GET /api/agent-operator-integration/internal-v1/operator/market/{operator_id}
```

**规则 5.3.2**：operator schema 包含以下信息

- OpenAPI 元数据
- 参数 schema（类型、必填项、枚举值等）
- 示例值

**规则 5.3.3**：使用 schema 时的注意事项

- schema 仅作为参考，**不得输出/复述 schema 原文**
- 必须输出符合 dynamic_params 格式的结果
- 优先从 additional_context/query 中抽取值，schema 示例作为兜底

### 5.4 Operator 参数生成示例

**示例 5.4.1**：简单参数（基础类型）

```json
// 输入
{
  "logic_property": {
    "name": "is_high_risk_company",
    "type": "operator",
    "parameters": [
      { "name": "threshold", "type": "NUMBER", "value_from": "input" },
      { "name": "include_reason", "type": "BOOLEAN", "value_from": "input" }
    ]
  },
  "query": "判断是否为高风险企业，阈值0.8，包含原因",
  "additional_context": ""
}

// 输出
{
  "is_high_risk_company": {
    "threshold": 0.8,
    "include_reason": true
  }
}
```

**示例 5.4.2**：复杂参数（OBJECT + ARRAY）

```json
// 输入
{
  "logic_property": {
    "name": "business_health_score",
    "type": "operator",
    "parameters": [
      { "name": "include_details", "type": "BOOLEAN", "value_from": "input" },
      { "name": "lang", "type": "STRING", "value_from": "input" },
      { "name": "filter", "type": "OBJECT", "value_from": "input" },
      { "name": "items", "type": "ARRAY", "value_from": "input" }
    ]
  },
  "operator_schema": {
    "requestBody": {
      "content": {
        "application/json": {
          "schema": {
            "type": "object",
            "required": ["include_details", "lang", "filter", "items"],
            "properties": {
              "include_details": {
                "type": "boolean",
                "description": "是否返回详细评分明细"
              },
              "lang": {
                "type": "string",
                "enum": ["zh-CN", "en-US"],
                "description": "返回结果语言"
              },
              "filter": {
                "type": "object",
                "required": ["year_range", "regions"],
                "properties": {
                  "year_range": {
                    "type": "array",
                    "items": { "type": "integer" }
                  },
                  "regions": {
                    "type": "array",
                    "items": { "type": "string" }
                  }
                }
              },
              "items": {
                "type": "array",
                "items": { "type": "string" },
                "description": "评分维度"
              }
            }
          }
        }
      }
    }
  },
  "query": "评估企业健康度，2023-2024年，华东和华南地区，中文",
  "additional_context": "items=revenue,profit,growth"
}

// 输出
{
  "business_health_score": {
    "include_details": false,
    "lang": "zh-CN",
    "filter": {
      "year_range": [2023, 2024],
      "regions": ["华东", "华南"]
    },
    "items": ["revenue", "profit", "growth"]
  }
}
```

---

## 6. 输出格式规则

### 6.1 JSON 格式约束

**规则 6.1.1**：必须输出严格合法的 JSON

- 禁止输出任何额外文本、解释、markdown、注释
- 禁止输出 JSON 之外的任何内容

**规则 6.1.2**：输出格式二选一

**选项 1：成功输出**（dynamic_params）
```json
{
  "<logic_property.name>": {
    "<param_name>": <param_value>,
    ...
  }
}
```

**选项 2：缺参输出**（错误信息）
```json
{
  "_error": "missing <logic_property.name>: <p1>,<p2> | ask: <一句话>"
}
```

**规则 6.1.3**：成功输出时顶层只能有一个 key

- key 必须是 logic_property.name
- value 是该属性的所有参数键值对

**规则 6.1.4**：缺参输出时顶层只能有 `_error`

```json
{
  "_error": "missing approved_drug_count: start,end | ask: 请明确查询的时间范围"
}
```

### 6.2 缺参错误格式

**规则 6.2.1**：缺参错误必须包含以下信息

- 缺失的属性名（logic_property.name）
- 缺失的参数列表（用逗号分隔）
- 补参建议（一句话说明如何补充）

**规则 6.2.2**：缺参错误示例

```json
{
  "_error": "missing approved_drug_count: start,end,step | ask: 请明确查询的时间范围和步长"
}
```

```json
{
  "_error": "missing business_health_score: filter,items | ask: 请在 additional_context 中补充筛选条件和评分维度"
}
```

---

## 7. 校验规则

### 7.1 JSON 合法性校验

**规则 7.1.1**：输出必须是合法的 JSON

- 检查 JSON 解析是否成功
- 失败则返回错误

### 7.2 Metric 参数校验

**规则 7.2.1**：instant 参数校验

- 必须是 BOOLEAN 类型
- 不能为 null 或 undefined

**规则 7.2.2**：start/end 参数校验

- 必须是 INTEGER 类型（毫秒时间戳）
- instant=false 时必填
- start 必须 <= end

**规则 7.2.3**：step 参数校验

- 必须是 STRING 类型
- 必须是枚举值之一：`day/week/month/quarter/year`
- instant=false 时必填

**规则 7.2.4**：instant/step 组合校验

```javascript
// 校验逻辑
if (instant === false) {
  if (!start || !end || !step) {
    return "instant=false 时必须提供 start, end, step";
  }
  if (start > end) {
    return "start 必须 <= end";
  }
  if (!['day', 'week', 'month', 'quarter', 'year'].includes(step)) {
    return "step 必须是 day/week/month/quarter/year 之一";
  }
}
```

### 7.3 Operator 参数校验

**规则 7.3.1**：参数完整性校验

- 必须包含所有 `value_from="input"` 的参数
- 缺少任何参数则返回缺参错误

**规则 7.3.2**：参数类型校验

- 检查参数值类型是否与定义匹配
- 不匹配则返回错误

**规则 7.3.3**：参数结构校验

- OBJECT 类型：必须是 JSON 对象
- ARRAY 类型：必须是 JSON 数组

---

## 8. 并发策略

### 8.1 并发粒度

**规则 8.1.1**：按 property 并发生成

- 每个 property 的 dynamic_params 独立生成
- 共享上下文（query/unique_identities/additional_context/now_ms/timezone）
- 最终合并成完整的 dynamic_params

### 8.2 并发控制

**规则 8.2.1**：并发上限（max_concurrency）

- 默认值：4
- 可配置范围：3-8
- 选择依据：平衡 LLM 网关限流（如 10 QPS）与延迟收益

**规则 8.2.2**：失败处理

- 任一属性缺参/生成失败 → 整体失败
- 返回缺参清单，交给 Agent 补充 additional_context 后重试
- 不支持部分成功

### 8.3 重试机制

**规则 8.3.1**：LLM 调用失败重试

- 可重试错误：HTTP 429/5xx/网络超时
- 最多重试 2 次（含首次共 3 次调用）
- 指数退避：100ms → 200ms

**规则 8.3.2**：不可重试错误

- HTTP 4xx（除 429）
- JSON 解析失败
- 缺参错误

**规则 8.3.3**：其他依赖失败不重试

- ontology-manager 调用失败 → 直接返回错误
- ontology-query 调用失败 → 透传错误给上游

---

## 9. 完整示例

### 9.1 Metric 示例

**场景**：查询最近 3 个月的药品数量趋势

**输入**：
```json
{
  "logic_property": {
    "name": "approved_drug_count",
    "type": "metric",
    "parameters": [
      { "name": "instant", "value_from": "input", "if_system_generate": true },
      { "name": "start", "value_from": "input", "if_system_generate": true },
      { "name": "end", "value_from": "input", "if_system_generate": true },
      { "name": "step", "value_from": "input", "if_system_generate": true }
    ]
  },
  "query": "最近3个月的药品数量趋势",
  "additional_context": "now_ms=1762996342241"
}
```

**输出**：
```json
{
  "approved_drug_count": {
    "instant": false,
    "start": 1760998342241,
    "end": 1762996342241,
    "step": "month"
  }
}
```

**说明**：
- instant=false（趋势查询）
- start = now_ms - 90天（3个月）
- end = now_ms
- step = month（月度步长）

### 9.2 Operator 示例

**场景**：评估企业健康度

**输入**：
```json
{
  "logic_property": {
    "name": "business_health_score",
    "type": "operator",
    "parameters": [
      { "name": "include_details", "type": "BOOLEAN", "value_from": "input" },
      { "name": "lang", "type": "STRING", "value_from": "input" },
      { "name": "filter", "type": "OBJECT", "value_from": "input" },
      { "name": "items", "type": "ARRAY", "value_from": "input" }
    ]
  },
  "query": "评估企业健康度，2023-2024年，华东和华南地区，中文",
  "additional_context": "items=revenue,profit,growth"
}
```

**输出**：
```json
{
  "business_health_score": {
    "include_details": false,
    "lang": "zh-CN",
    "filter": {
      "year_range": [2023, 2024],
      "regions": ["华东", "华南"]
    },
    "items": ["revenue", "profit", "growth"]
  }
}
```

### 9.3 缺参示例

**场景**：查询药品数量，但缺少时间范围

**输入**：
```json
{
  "logic_property": {
    "name": "approved_drug_count",
    "type": "metric",
    "parameters": [
      { "name": "instant", "value_from": "input", "if_system_generate": true },
      { "name": "start", "value_from": "input", "if_system_generate": true },
      { "name": "end", "value_from": "input", "if_system_generate": true },
      { "name": "step", "value_from": "input", "if_system_generate": true }
    ]
  },
  "query": "药品数量",
  "additional_context": ""
}
```

**输出**：
```json
{
  "_error": "missing approved_drug_count: start,end,step | ask: 请明确查询的时间范围和步长，或在 additional_context 中补充 now_ms"
}
```

---

## 10. 常见问题

### Q1：为什么 dynamic_params 只包含 value_from="input" 的参数？

**A**：property 和 const 参数不需要大模型生成：
- property 参数：从对象实例的数据属性中直接获取
- const 参数：使用预定义的常量值
- 只有 input 参数需要大模型从上下文中抽取

### Q2：instant=true 时还需要生成 start/end/step 吗？

**A**：不需要。instant=true 表示即时查询（单点值），只需要生成 `instant: true` 即可。

### Q3：step 可以是 "2month" 或 "7day" 吗？

**A**：不可以。step 只能是 `day/week/month/quarter/year` 之一，固定 1 个单位。

### Q4：operator 参数的 Path/Query/Body 如何处理？

**A**：统一输出为 dynamic_params[property] 的键值对，不区分来源位置。

### Q5：缺参时如何处理？

**A**：输出 `_error` 字段，包含缺失的参数列表和补参建议，交给 Agent 补充 additional_context 后重试。

### Q6：并发生成时，某个 property 失败怎么办？

**A**：任一 property 失败则整体失败，返回缺参清单，不支持部分成功。

---

## 11. 附录

### 11.1 数据结构定义

**LogicPropertyDef**
```go
type LogicPropertyDef struct {
    Name        string              `json:"name"`
    DisplayName string              `json:"display_name,omitempty"`
    Type        LogicPropertyType   `json:"type"` // metric 或 operator
    Comment     string              `json:"comment,omitempty"`
    DataSource  map[string]any      `json:"data_source,omitempty"`
    Parameters  []PropertyParameter `json:"parameters,omitempty"`
}
```

**PropertyParameter**
```go
type PropertyParameter struct {
    Name             string `json:"name"`
    Type             string `json:"type"`
    ValueFrom        string `json:"value_from"` // "input", "property", "const"
    Value            any    `json:"value,omitempty"`
    IfSystemGenerate bool   `json:"if_system_generate,omitempty"`
    Comment          string `json:"comment,omitempty"`
}
```

### 11.2 相关文档

- 《03-方案设计.md》：整体方案设计
- 《04-metric参数生成提示词设计.md》：metric 提示词设计
- 《07-operator参数生成提示词设计.md》：operator 提示词设计
- 《05-并发策略与超时配置.md》：并发策略详情
- 《18-dynamic_params生成规则与大模型理解指南.md》：大模型理解策略

### 11.3 代码实现参考

- `/root/codes/agent-retrieval/server/interfaces/driven_ontology_manager.go`：LogicPropertyDef 和 PropertyParameter 定义
- `/root/codes/agent-retrieval/server/interfaces/kn_logic_property_resolver.go`：LLMPromptInput 和 LLMPromptOutput 定义
- `/root/codes/agent-retrieval/server/interfaces/driven_agent.go`：MetricDynamicParamsGeneratorReq 定义

---

## 12. 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2025-01-XX | 初稿，完整定义 dynamic_params 生成规则 |
