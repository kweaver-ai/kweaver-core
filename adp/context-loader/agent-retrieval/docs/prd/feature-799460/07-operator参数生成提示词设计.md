# operator 参数生成提示词设计

## 目标

为 **operator 类型逻辑属性** 生成该属性对应的 `dynamic_params[property]`，用于后续调用底层逻辑属性查询接口。

设计原则：**参照 metric 提示词结构，区别在于入参约束和参数类型处理**。

---

## 0. 前置：算子 Schema（OpenAPI）如何进入本次参数生成上下文

operator 类型逻辑属性的 `logic_property.parameters` 可能不够完整/不够精确（尤其是 BODY 的复杂嵌套结构、枚举、默认值、对象数组等）。
因此 **本次参数生成默认会额外提供**算子 Schema（OpenAPI/JSON Schema/示例等），下文统称 `operator_schema`。

在提示词层面我们不再描述“如何获取/是否调用工具”，只要求：
- `operator_schema` **会在本次 LLM 上下文中被动态注入/提供**
- 你需要将其作为“参数类型/嵌套结构/枚举/默认值/示例”的参考来源来生成 `dynamic_params`

> 说明：`operator_schema` 的来源可以是算子详情接口、缓存、配置或其他机制；这不影响本设计的输出约束。

---

## 1. operator 与 metric 的关键区别

| 维度 | metric | operator |
|------|--------|----------|
| **系统固定参数** | 有（instant/start/end/step） | 无 |
| **参数类型** | 基础类型为主（boolean/int/string） | 可能包含复杂类型（对象/数组） |
| **参数来源** | 主要从 query/additional_context 推导时间 | 可能从 object_instances 提取业务字段 |
| **缺参容忍度** | 时间参数有默认策略（如近3个月） | 业务参数通常无默认值，必须明确 |
| **参数定义来源** | logic_property.parameters | logic_property.parameters + operator_schema（强烈建议提供） |

---

## 2. 输入是什么（Input Contract）

与 metric 保持一致，最小输入字段：

- **logic_property**: object（必须）
  - 包含：name / display_name / type / comment / data_source / parameters
  - 其中 `logic_property.type` 必须为 `"operator"`
- **query**: string（用户原始问题）
- **unique_identities**: array<object>（对象主键列表）
- **additional_context**: string（额外上下文信息）
- **object_instances**: array<object>（可选但强烈建议）
  - operator 可能需要从对象实例中提取业务字段作为参数值
  - 例如：某个 operator 需要 `company_name` 参数，可以从 object_instances 中提取

**可选增强输入**（提升准确率）:
- **operator_schema**: object（可选）
  - 算子的 OpenAPI/schema/示例信息（用于补齐复杂嵌套结构、枚举、默认值等）
  - 运行时会动态注入/提供到本次 LLM 上下文中（推荐）
  - 若缺失：仅能依赖 logic_property.parameters 的 type/comment 推断，复杂参数准确率会明显下降

---

## 3. 输出是什么（Output Contract）

与 metric 保持一致，LLM 输出必须是 **严格合法 JSON**，且只能二选一：

### 3.1 成功输出（dynamic_params 片段）

```json
{
  "business_health_score": {
    "include_details": true,
    "lang": "zh-CN"
  }
}
```

**复杂参数示例**（对象/数组）:
```json
{
  "project_analysis": {
    "projects": [
      {"project_id": "p1", "weight": 0.6},
      {"project_id": "p2", "weight": 0.4}
    ],
    "analysis_type": "comprehensive"
  }
}
```

### 3.2 缺参输出（用于 Agent 追问/补上下文）

```json
{
  "_error": "missing business_health_score: include_details | ask: 是否需要详细评分明细(true/false)"
}
```

---

## 4. operator 参数生成的关键规则（必须写进指令）

### 4.1 只生成 value_from="input" 的参数

与 metric 一致，**不要生成** `value_from="property"` 或 `value_from="const"` 的参数。

### 4.1-B 必须输出 dynamic_params（禁止按 OpenAPI 的 in/source 分组）

operator_schema 往往是 OpenAPI 文档，参数可能分散在：
- header / path / query / cookie / requestBody（body）等

但在本系统中，**你只能输出 dynamic_params 片段**，即：
- 顶层 key：`logic_property.name`
- value：一个 **单一 JSON object**，包含所有 `value_from="input"` 的参数键值对

**硬性禁止**以下“按 OpenAPI 分组”的输出形态（示例）：
```json
{
  "business_health_score": {
    "header": {"Authorization": "xxx"},
    "query": {"include_details": true},
    "path": {"company_id": "c1"},
    "body": {"dimensions": "all"}
  }
}
```

你必须忽略 OpenAPI 的 `in` / 逻辑属性参数的 `source` 字段（Path/Query/Header/Body 等），所有入参统一落在一个 object 中。

### 4.2 参数值的类型约束

operator 参数可能是：
- **基础类型**: string / number / boolean / integer
- **对象**: `{ "key1": "value1", "key2": 123 }`
- **数组**: `["item1", "item2"]` 或 `[{ "id": 1 }, { "id": 2 }]`

必须严格遵守 `logic_property.parameters[].type` 定义的类型。

### 4.3 参数值的抽取优先级

1. **additional_context**（优先）:
   - 如果是 JSON 字符串，按键直接取值
   - 如果是自由文本，按语义抽取

2. **object_instances**（次优）:
   - 如果参数名与对象字段名匹配，直接提取
   - 例如：参数 `company_name`，从 `object_instances[0].company_name` 提取

3. **query**（兜底）:
   - 从自然语言中抽取

4. **operator_schema 示例/默认值**（最后）:
   - 如果提供了 operator_schema，可以参考其示例值

5. **仍无法确定** → 返回 `_error`

### 4.4 复杂参数的处理

**对象类型参数**:
- 该参数在 `dynamic_params[property]` 中是一个字段，其值是一个 JSON object
- 必须输出完整对象结构（不能把对象拆散成多个顶层参数；也不能只输出部分字段）
- 如果某些字段可选，可以省略（但必须保证必填字段存在）

**数组类型参数**:
- 该参数在 `dynamic_params[property]` 中是一个字段，其值是一个 JSON array
- 如果 query/additional_context 提到多个值，输出数组
- 如果只有一个值但参数定义是数组，也要输出数组格式 `["value"]`

**嵌套类型的关键约束（你强调的点）**：
- 如果 BODY 中某个参数是嵌套类型（object/array/对象数组/多层嵌套），**直接把“整个参数值”放到 dynamic_params 里**
- 不要尝试把嵌套结构“打平”为多个参数，也不要按 body/query/header 再做一层包裹

示例（正确：嵌套参数整体放入）：
```json
{
  "some_operator_property": {
    "filter": {
      "year_range": { "start": 2020, "end": 2023 },
      "regions": ["北京", "上海"]
    },
    "items": [
      { "id": "a", "weight": 0.6 },
      { "id": "b", "weight": 0.4 }
    ]
  }
}
```

### 4.5 operator_schema 的使用方式（强约束）

- operator_schema 只用于帮助你确定：
  - **参数是否必填/可选**
  - **参数类型与嵌套结构**（object/array 的字段、对象数组 item schema）
  - **枚举值/默认值/示例**
- 你的最终输出**不得**包含 operator_schema/OpenAPI 原文；只能输出 dynamic_params 或 `_error`。

---

## 5. 提示词指令模板（可直接使用）

> 下方是"指令模板"。resolver 运行时将真实输入 JSON 填入最后的 `INPUT_JSON`。

【角色】
你是算子（operator）逻辑属性的参数生成器。

【任务】
基于输入 JSON，为 logic_property 生成 dynamic_params 片段，只生成 logic_property.parameters 中 value_from="input" 的参数。

【补充上下文：operator_schema】
输入中可能包含 operator_schema（算子的 OpenAPI/schema/示例）。你必须参考它来确定复杂参数（object/array）的嵌套结构与类型约束。
但你的输出只能是 dynamic_params（或 `_error`），不得输出/复述 operator_schema 原文。

【输出】
只能输出严格合法 JSON，且只能二选一：
1) 成功：输出 dynamic_params 片段（顶层 key=logic_property.name）：
   { "<logic_property.name>": { "<param_name>": <param_value>, ... } }
2) 缺参：输出缺参结构（顶层 key 固定为 _error）：
   { "_error": "missing <logic_property.name>: <p1>,<p2> | ask: <一句话>" }

【硬约束】
1) 只生成 value_from="input" 的参数；不要输出 property/const 参数。
2) 对所有 value_from="input" 参数都要尝试生成：能确定就填值；不能确定就输出 `_error`。
3) 严格遵守参数类型约束（logic_property.parameters[].type）：
   - 基础类型：string/number/boolean/integer
   - 对象：必须输出完整对象结构（JSON object）
   - 数组：必须输出 JSON 数组格式
4) **必须输出 dynamic_params 的单一 object**：禁止按 header/query/path/body 分组；忽略 OpenAPI 的 `in` 与参数的 `source` 字段。
5) 输出不得包含任何额外文本、解释、markdown、注释。
6) 成功输出时：顶层只能有一个 key（logic_property.name）；缺参输出时：顶层只能有 `_error`。

【参数值抽取优先级】
1. additional_context（优先）：如果是 JSON 字符串，按键取值；如果是文本，按语义抽取。
2. object_instances（次优）：如果参数名与对象字段名匹配，直接提取。
3. query（兜底）：从自然语言中抽取。
4. operator_schema 示例/默认值（如果提供）。
5. 仍无法确定 → 输出 `_error`。

【复杂参数处理】
- 对象类型参数：作为一个字段整体输出（不要拆散/打平），必须输出完整对象结构（必填字段必须存在）。
- 数组类型参数：作为一个字段整体输出（不要拆散/打平），如果只有一个值但参数定义是数组，也要输出数组格式 ["value"]。
- 布尔类型参数：只能输出 true/false（不要输出字符串 "true"/"false"）。

【自检】
- JSON 可解析？
- 只包含 value_from="input" 参数？
- 参数类型与 logic_property.parameters[].type 一致？
- 对象/数组参数结构完整？

INPUT_JSON:
{{INPUT_JSON}}

---

## 6. 测试样例数据（用于调试提示词）

### 样例 1：简单布尔参数（business_health_score）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "business_health_score",
    "display_name": "企业经营健康度评分",
    "type": "operator",
    "comment": "企业经营健康度评分算子",
    "data_source": { "type": "operator", "id": "business_health_operator", "name": "企业健康度" },
    "parameters": [
      { "name": "company_id", "type": "text", "value_from": "property", "value": "company_id" },
      { "name": "include_details", "type": "BOOLEAN", "value_from": "input", "comment": "是否返回详细评分明细" },
      { "name": "lang", "type": "STRING", "value_from": "input", "comment": "返回结果语言（zh-CN/en-US）" }
    ]
  },
  "query": "查询企业健康度评分，需要详细明细，中文返回",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A" }
  ]
}
```

期望输出：
```json
{
  "business_health_score": {
    "include_details": true,
    "lang": "zh-CN"
  }
}
```

---

### 样例 2：从 object_instances 提取参数（company_name）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "company_risk_analysis",
    "type": "operator",
    "parameters": [
      { "name": "company_id", "value_from": "property" },
      { "name": "company_name", "type": "STRING", "value_from": "input", "comment": "公司名称" },
      { "name": "analysis_depth", "type": "STRING", "value_from": "input", "comment": "分析深度（shallow/deep）" }
    ]
  },
  "query": "深度分析企业风险",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A", "registered_capital": 2000000 }
  ]
}
```

期望输出：
```json
{
  "company_risk_analysis": {
    "company_name": "某药企A",
    "analysis_depth": "deep"
  }
}
```

---

### 样例 3：复杂对象参数（filter 对象）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "drug_approval_analysis",
    "type": "operator",
    "parameters": [
      { "name": "company_id", "value_from": "property" },
      {
        "name": "filter",
        "type": "OBJECT",
        "value_from": "input",
        "comment": "筛选条件对象，结构：{ drug_type: string, approval_year_start: int, approval_year_end: int }"
      }
    ]
  },
  "query": "分析该企业2020-2023年的化学药品审批情况",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A" }
  ]
}
```

期望输出：
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

---

### 样例 4：数组参数（projects 列表）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "multi_project_analysis",
    "type": "operator",
    "parameters": [
      { "name": "company_id", "value_from": "property" },
      {
        "name": "project_ids",
        "type": "ARRAY",
        "value_from": "input",
        "comment": "项目ID列表（字符串数组）"
      }
    ]
  },
  "query": "分析项目proj001和proj002",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A" }
  ]
}
```

期望输出：
```json
{
  "multi_project_analysis": {
    "project_ids": ["proj001", "proj002"]
  }
}
```

---

### 样例 5：复杂数组参数（对象数组）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "weighted_project_analysis",
    "type": "operator",
    "parameters": [
      { "name": "company_id", "value_from": "property" },
      {
        "name": "projects",
        "type": "ARRAY",
        "value_from": "input",
        "comment": "项目列表（对象数组），每个对象包含 project_id(string) 和 weight(number)"
      }
    ]
  },
  "query": "分析项目proj001（权重0.6）和proj002（权重0.4）",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A" }
  ]
}
```

期望输出：
```json
{
  "weighted_project_analysis": {
    "projects": [
      { "project_id": "proj001", "weight": 0.6 },
      { "project_id": "proj002", "weight": 0.4 }
    ]
  }
}
```

---

### 样例 6：从 additional_context（JSON 字符串）提取参数

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "business_health_score",
    "type": "operator",
    "parameters": [
      { "name": "company_id", "value_from": "property" },
      { "name": "include_details", "type": "BOOLEAN", "value_from": "input" },
      { "name": "lang", "type": "STRING", "value_from": "input" }
    ]
  },
  "query": "查询企业健康度评分",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "{\"include_details\":false,\"lang\":\"en-US\"}",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A" }
  ]
}
```

期望输出：
```json
{
  "business_health_score": {
    "include_details": false,
    "lang": "en-US"
  }
}
```

---

### 样例 7：缺参示例（无法从上下文推断）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "drug_approval_analysis",
    "type": "operator",
    "parameters": [
      { "name": "company_id", "value_from": "property" },
      { "name": "drug_type", "type": "STRING", "value_from": "input", "comment": "药品类型（化学药品/生物制品/中药）" }
    ]
  },
  "query": "分析该企业的药品审批情况",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A" }
  ]
}
```

期望输出：
```json
{
  "_error": "missing drug_approval_analysis: drug_type | ask: 请指定药品类型（化学药品/生物制品/中药）"
}
```

---

### 样例 8：多对象批量（同一参数）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "business_health_score",
    "type": "operator",
    "parameters": [
      { "name": "company_id", "value_from": "property" },
      { "name": "include_details", "type": "BOOLEAN", "value_from": "input" }
    ]
  },
  "query": "批量查询企业健康度评分（需要详细信息）",
  "unique_identities": [
    { "company_id": "company_000001" },
    { "company_id": "company_000002" }
  ],
  "additional_context": "",
  "object_instances": [
    { "company_id": "company_000001", "company_name": "某药企A" },
    { "company_id": "company_000002", "company_name": "某药企B" }
  ]
}
```

期望输出：
```json
{
  "business_health_score": {
    "include_details": true
  }
}
```

---

### 样例 9：OpenAPI(body) 复杂嵌套参数（object + 对象数组，多层嵌套）

场景：算子在 body 中定义了复杂结构参数（filter 对象、items 对象数组），要求整体作为单个字段放入 dynamic_params。

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "complex_operator_demo",
    "display_name": "复杂算子示例",
    "type": "operator",
    "comment": "演示 body 嵌套结构的 dynamic_params 生成",
    "data_source": { "type": "operator", "id": "op_complex_demo", "name": "复杂算子" },
    "parameters": [
      { "name": "company_id", "type": "STRING", "value_from": "property", "value": "company_id" },
      { "name": "filter", "type": "OBJECT", "value_from": "input", "comment": "筛选条件对象" },
      { "name": "items", "type": "ARRAY", "value_from": "input", "comment": "对象数组参数" },
      { "name": "lang", "type": "STRING", "value_from": "input", "comment": "返回语言" }
    ]
  },
  "query": "按2020-2023年、地区北京/上海过滤，并给两个加权条目，中文返回",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "object_instances": [{ "company_id": "company_000001", "company_name": "某药企A" }],
  "operator_schema": {
    "openapi_hint": "仅示例：真实场景可能是完整 OpenAPI 文档",
    "requestBody": {
      "content": {
        "application/json": {
          "schema": {
            "type": "object",
            "required": ["filter", "items", "lang"],
            "properties": {
              "filter": {
                "type": "object",
                "required": ["year_range", "regions"],
                "properties": {
                  "year_range": {
                    "type": "object",
                    "required": ["start", "end"],
                    "properties": { "start": { "type": "integer" }, "end": { "type": "integer" } }
                  },
                  "regions": { "type": "array", "items": { "type": "string" } }
                }
              },
              "items": {
                "type": "array",
                "items": {
                  "type": "object",
                  "required": ["id", "weight"],
                  "properties": { "id": { "type": "string" }, "weight": { "type": "number" } }
                }
              },
              "lang": { "type": "string", "enum": ["zh-CN", "en-US"], "default": "zh-CN" }
            }
          }
        }
      }
    }
  }
}
```

期望输出（注意：filter 与 items 必须整体嵌套放入，不打平，不包 body/query/header）：
```json
{
  "complex_operator_demo": {
    "filter": {
      "year_range": { "start": 2020, "end": 2023 },
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

---

## 7. operator_schema 增强（可选）

如果 resolver 从 operator-platform 拉取了算子的 OpenAPI 定义，可以将其作为额外输入传递给 LLM，提升参数生成准确率。

### 7.1 operator_schema 结构示例

```json
{
  "operator_id": "business_health_operator",
  "name": "企业经营健康度评分",
  "openapi": {
    "paths": {
      "/execute": {
        "post": {
          "parameters": [
            {
              "name": "include_details",
              "in": "query",
              "schema": { "type": "boolean", "default": false }
            },
            {
              "name": "lang",
              "in": "query",
              "schema": { "type": "string", "enum": ["zh-CN", "en-US"], "default": "zh-CN" }
            }
          ]
        }
      }
    }
  }
}
```

### 7.2 如何使用 operator_schema

在提示词中补充一段：

```
【operator_schema 补充信息】（如果提供）
算子 OpenAPI 定义如下，可以参考其参数类型、枚举值、默认值、示例：
{{operator_schema}}

但必须优先使用 additional_context / object_instances / query 中的明确信息，operator_schema 仅作为兜底参考。
```

### 7.3 何时拉取 operator_schema

- **不建议每次都拉取**（增加延迟和依赖）
- **建议按需拉取**：
  - 如果 operator 参数类型复杂（OBJECT/ARRAY）
  - 或者历史缺参率高
  - 则在 LLM 调用前先拉取 operator_schema 并合并到输入中

---

## 8. 与 metric 提示词的对比总结

| 维度 | metric | operator |
|------|--------|----------|
| **角色** | 指标（metric）逻辑属性的参数生成器 | 算子（operator）逻辑属性的参数生成器 |
| **固定参数** | instant/start/end/step（必须处理） | 无固定参数 |
| **参数类型** | 主要是基础类型 + 时间戳 | 可能包含复杂类型（对象/数组） |
| **抽取优先级** | additional_context > query > now_ms | additional_context > object_instances > query > operator_schema |
| **默认策略** | 时间参数有兜底（如近3个月） | 业务参数通常无默认值 |
| **缺参容忍度** | 可用默认时间范围降低缺参率 | 缺参必须明确报错 |

---

## 9. 实现建议

1. **复用 metric 提示词框架**：角色/任务/输出/硬约束结构保持一致
2. **删除 metric 特有的时间处理逻辑**：instant/start/end/step 约束
3. **增强参数类型处理**：明确对象/数组的输出格式要求
4. **补充 object_instances 抽取说明**：提高业务参数抽取准确率
5. **可选增强 operator_schema**：按需拉取，作为兜底参考


