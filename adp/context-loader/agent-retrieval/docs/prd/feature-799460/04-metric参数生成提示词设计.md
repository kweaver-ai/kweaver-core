## 目标

为 **metric 类型逻辑属性** 生成该属性对应的 `dynamic_params[property]`，用于后续调用底层逻辑属性查询接口。

设计原则：**按 property 生成、输入尽量短、输出强约束为严格 JSON、缺参可解释可追问**。

---

## 1. 输入是什么（Input Contract）

提示词输入建议由 resolver 以“结构化 JSON”传入（上游的 `additional_context` 是 string 也没关系，但 resolver 这里给 LLM 的输入必须结构化，减少歧义）。

最小输入字段（建议固定字段名；推荐把“逻辑属性定义”整体传入，避免重复传 property）：

- **logic_property**: object（必须）
  resolver 从对象类定义中提取的“指定逻辑属性定义”，建议至少包含：
  - name / display_name / type / comment / data_source / parameters
  其中 `logic_property.name` 就是本次目标属性，无需再单独传 `property` 字段。
- **query**: string
  用户原始问题（自然语言）
- **unique_identities**: array<object>
  对象主键列表（可能批量）。用于让模型知道“对象是什么”，以及某些业务参数可能从主键推导。
- **additional_context**: string
  调用方提供的上下文（自由文本或 JSON 字符串），可能包含：
  - 已筛选条件（例如注册资本>100万）
  - 已解析时间范围（例如最近半年）
  - 对象实例关键字段（例如 company_name）
  - 用户偏好（例如按月）
- **now_ms**: number（强烈建议作为“稳定必传”字段）
  用于解析“最近/当前/截至今天”等相对时间；即时查询（instant=true）通常直接用 now_ms 作为 start/end
- **timezone**: string（强烈建议作为“稳定必传”字段）
  用于自然语言时间解析一致性

> 关键点：LLM 不是“猜参数”，而是**从 logic_property.parameters 清单中挑出所有 value_from="input" 的参数并赋值**。

---

## 2. 输出是什么（Output Contract）

为减少“中间转换”，推荐让 LLM 直接输出 **dynamic_params 的片段**（与底层属性查询接口一致）：即 **用 property 作为 key**，value 为该 property 的参数对象。

LLM 输出必须是 **严格合法 JSON**，且只能二选一：

### 2.1 成功输出（dynamic_params 片段）

```json
{
  "approved_drug_count": {
    "instant": false,
    "start": 1731460342241,
    "end": 1762996342241,
    "step": "month"
  }
}
```

### 2.2 缺参输出（用于 Agent 追问/补上下文，极简）

```json
{
  "_error": "missing approved_drug_count: start,end | ask: time_range(start/end or 最近半年)"
}
```

输出约束：

- 成功输出时：顶层必须只有一个 key（即 `logic_property.name`），其 value 只能包含该属性 **value_from="input"** 的参数键值
- 缺参输出时：只输出 `_error`，且 `_error` 里必须包含“哪个 property 缺了哪些参数”的最小信息（便于 resolver/agent 重试）

---

## 3. metric 参数生成的关键规则（必须写进指令）

> **重要说明**：instant=true 和 instant=false 的本质区别
> - **instant=true**：返回单个聚合值（如总数、平均值），即使背后统计了一段时间范围的数据
>   - start/end 定义的是**聚合统计的时间范围**（如最近 7 天的总和）
>   - 如果不传 start/end，底层默认统计最近 30 天
> - **instant=false**：返回时间序列数据（多个时间点的值），用于绘制趋势图
>   - start/end 定义的是**趋势的时间范围**
>   - step 定义每个时间点的间隔（day/week/month/quarter/year）

### 3.1 系统时间参数（必生成/必处理）

metric 固定系统项（通常 `if_system_generate:true`）：

- **instant**: boolean
- **start**: int64 ms（可选，但推荐明确指定）
- **end**: int64 ms（可选，但推荐明确指定）
- **step**: string（仅 instant=false 时需要）

定义与约束：

- **instant=true**：即时查询（当前值/最新值）
  - **禁止输出 step**
  - **start/end 可选**：
    - 如果传递 start/end：统计该时间范围内的聚合数据（如最近 7 天、最近 30 天的总和）
    - 如果不传 start/end：底层默认统计最近 30 天的数据
    - 推荐策略：明确传递 start/end，避免依赖默认值
- **instant=false**：趋势查询（时间序列/按周期）
  - **必须输出 step**
  - step 枚举固定为：`day/week/month/quarter/year`（固定 1 单位）
  - **start/end 必需**：定义趋势查询的时间范围
- start/end 单位：毫秒时间戳（int64）

### 3.2 如何识别 instant=true vs instant=false

**识别规则（按优先级）：**

1. **优先从 additional_context 中识别**
   - 如果 additional_context 包含明确的 `instant` 字段（JSON 格式），直接使用
   - 例如：`{"instant": true}` 或 `{"instant": false, "step": "month"}`

2. **从 query 关键词识别**（按优先级判断）

   **instant=true 的关键词**（即时查询/聚合值/单个结果）：
   - **高优先级（强制 instant=true）**：
     - 总数、总量、合计、总计、汇总
     - 当前、最新、现在、此刻、目前
     - 截至今天、截至目前、截至当前
     - 最新值、当前值、实时值
   - **判断逻辑**：只要包含以上关键词，即使有时间范围（如"最近一年"），也应该是 instant=true
   - **示例**：
     - "查询最近一年的药品上市**总数**" → instant=true（总数是聚合值）
     - "查询当前药品数量" → instant=true（当前是即时值）

   **instant=false 的关键词**（趋势查询/时间序列）：
   - 趋势、变化、走势、发展
   - 按日/按周/按月/按季度/按年、每日/每周/每月/每季度
   - 时间序列、历史数据
   - **判断逻辑**：必须明确包含这些词，才是 instant=false

3. **无法明确判断时的默认策略**
   - 如果 query 包含具体时间范围（如"最近半年"），默认 `instant=false`
   - 如果 query 只是简单查询（如"查询药品数量"），默认 `instant=true`

**instant=true 时的 start/end 语义：**
- start/end 定义的是**聚合统计的时间范围**，不是趋势的起止点
- 例如："查询最近 7 天的药品上市总数" → `instant=true, start=now_ms-7天, end=now_ms`
- 如果不传 start/end，底层默认统计最近 30 天

### 3.3 业务 input 参数（同样必须生成）

除系统项外，parameters 清单中所有 `value_from="input"` 的业务参数也要生成（例如 `region`、`drug_type` 等）：

- 优先从 `additional_context` 抽取（若是 JSON 字符串，按键取值；若是文本，按语义抽取）
- 其次从 `query` 抽取
- 仍无法确定则返回 `_error`，用极短文字说明"缺什么、怎么补"

---

## 4. 提示词指令怎么写（推荐结构）

建议提示词由以下段落组成（保证不同模型也能理解，且可控）：

- **角色/任务**：你是 metric 参数生成器，只做一件事：为一个 logic_property 生成 dynamic_params
- **输入说明**：输入是 JSON，字段固定
- **输出格式**：只允许输出严格 JSON；成功/缺参二选一
- **硬约束**（最重要）：只输出 value_from="input"；instant/step 互斥规则；step 枚举；start/end 毫秒；禁止多余文本
- **判定规则**：如何从 query 判断即时 vs 趋势；如何选择 step；时间推导依赖 now_ms/timezone
- **自检清单**：输出前检查 JSON、step 枚举、instant/step 组合、参数覆盖范围

---

## 5. 可直接使用的“提示词指令模板”（给 resolver 拼接）

> 下方是“指令模板”（system+user 皆可用）。resolver 运行时将真实输入 JSON 填入最后的 `INPUT_JSON`。

【角色】
你是指标（metric）逻辑属性的参数生成器。

【任务】
基于输入 JSON，为 logic_property 生成 dynamic_params 片段，只生成 logic_property.parameters 中 value_from="input" 的参数。

【输出】
只能输出严格合法 JSON，且只能二选一：
1) 成功：输出 dynamic_params 片段（顶层 key=logic_property.name）：
   { "<logic_property.name>": { "<param_name>": <param_value>, ... } }
2) 缺参：输出缺参结构（顶层 key 固定为 _error）：
   { "_error": "missing <logic_property.name>: <p1>,<p2> | ask: <一句话>" }

【硬约束】
1) 只生成 value_from="input" 的参数；不要输出 property/const 参数。
2) 对所有 value_from="input" 参数都要尝试生成：能确定就填值；不能确定就输出 `_error`。
3) instant/start/end/step 必须处理：
   - **识别 instant=true 的最高优先级规则**：
     - 如果 query 包含"总数/总量/合计/总计/汇总"，必须 instant=true
     - 即使同时包含时间范围（如"最近一年"），也是 instant=true
     - 例如："最近一年的药品总数" → instant=true, start=now_ms-1年, end=now_ms

   - instant=true（即时查询/当前值/聚合值）：
     - 禁止输出 step
     - start/end 推荐输出（定义聚合统计的时间范围）
     - 如不传 start/end，底层默认统计最近 30 天
     - 例如："最近 7 天的药品总数" → instant=true, start=now_ms-7天, end=now_ms

   - instant=false（趋势查询/时间序列）：
     - 必须明确包含"趋势/变化/走势/按XX/每XX"等关键词
     - 必须输出 step（枚举：day/week/month/quarter/year，固定 1 单位）
     - 必须输出 start/end（定义趋势的时间范围）
     - 例如："最近半年的趋势（按月）" → instant=false, step=month
4) start/end 必须是毫秒时间戳（int64）。若 query/additional_context 是自然语言时间，请基于 now_ms/timezone 推导。
5) 输出不得包含任何额外文本、解释、markdown、注释。
6) 成功输出时：顶层只能有一个 key（logic_property.name）；缺参输出时：顶层只能有 `_error`。
7) 尽量不要输出 `_error`：除非缺少 `now_ms` 且又需要推导时间范围导致无法给出 start/end，否则一律用默认策略生成可执行参数。

【判定规则：识别 instant=true vs instant=false】
1) **优先从 additional_context 中识别**：
   - 如果 additional_context 包含明确的 `instant` 字段（JSON），直接使用

2) **从 query 关键词识别（按优先级）**：
   - **最高优先级（强制 instant=true）**：
     - 包含"总数/总量/合计/总计/汇总" → 必须 instant=true（即使有时间范围）
     - 例如："最近一年的药品上市**总数**" → instant=true, start/end 为最近一年

   - **次优先级（instant=true）**：
     - 包含"当前/最新/现在/此刻/目前/截至今天/最新值/当前值/实时值"
     - 例如："查询当前药品数量" → instant=true

   - **明确的 instant=false 标志**：
     - 必须包含"趋势/变化/走势/按日/周/月/季/年/每日/周/月"
     - 例如："最近半年的药品上市趋势" → instant=false

3) **instant=true 的 start/end 语义**：
   - start/end 定义的是聚合统计的时间范围，不是趋势点
   - 例如："最近一年的药品总数" → instant=true, start=now_ms-1年, end=now_ms
   - 如果 query 只说"查询当前XX"没有时间范围，可以不输出 start/end（底层默认最近 30 天）

4) **无法判断时的默认策略**：
   - 优先考虑 instant=true（更常见）
   - 简单查询（如"查询药品数量"）→ instant=true（不输出 start/end，走底层默认）

【默认策略（强烈建议，提升稳定性）】
- 若未命中即时关键词（即不需要 instant=true），且无法从 query/additional_context 抽取明确时间范围：
  - instant=false
  - end=now_ms
  - start=now_ms-90天（粗略按 90*24*3600*1000 毫秒）
  - step="month"
- 若命中即时关键词（instant=true）但缺少 now_ms：
  - 输出 `_error`: "missing <logic_property.name>: now_ms | ask: need now_ms for instant=true"

【自检】
- JSON 可解析？
- 只包含 value_from="input" 参数？
- instant=false 是否包含 step 且枚举合法？
- instant=true 是否未输出 step？

INPUT_JSON:
{{INPUT_JSON}}

---

## 6. 测试样例数据（用于调试提示词）

说明：每个样例包含两段：
- INPUT_JSON：resolver 输入给 LLM 的结构化 JSON（可直接复制）
- 期望输出：LLM 应输出的严格 JSON（用于对比调试）

> now_ms/timezone 这里统一使用：now_ms=1762996342241，timezone=Asia/Shanghai（你也可以替换成真实运行时值）。

### 样例 1：即时查询（当前值，不指定时间范围）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "approved_drug_count",
    "display_name": "药品上市数量",
    "type": "metric",
    "comment": "药品上市数量",
    "index": false,
    "data_source": { "type": "metric", "id": "approved_drug_metric", "name": "药品上市数量" },
    "parameters": [
      { "name": "company_id", "type": "text", "value_from": "property", "value": "company_id" },
      { "name": "company_name", "type": "text", "value_from": "property", "value": "company_name" },
      { "name": "instant", "type": "BOOLEAN", "value_from": "input", "if_system_generate": true, "comment": "是否是即时查询。instant=true 即时；instant=false 趋势/范围查询。" },
      { "name": "start", "type": "INTEGER", "value_from": "input", "if_system_generate": true, "comment": "指标查询开始时间（毫秒时间戳）。instant=true 时定义聚合统计的时间范围。" },
      { "name": "end", "type": "INTEGER", "value_from": "input", "if_system_generate": true, "comment": "指标查询结束时间（毫秒时间戳）。" },
      { "name": "step", "type": "STRING", "value_from": "input", "if_system_generate": true, "comment": "趋势查询步长。instant=false 时必须，枚举：day/week/month/quarter/year（固定1单位）。" }
    ]
  },
  "query": "查询当前药品上市数量（最新值）",
  "unique_identities": [{ "company_id": "company_000001", "company_name": "某药企A" }],
  "additional_context": "now_ms=1762996342241 timezone=Asia/Shanghai",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出（不输出 start/end，走底层默认最近 30 天）：
```json
{
  "approved_drug_count": { "instant": true }
}
```

### 样例 1-B：即时查询（指定时间范围：最近 7 天）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "approved_drug_count",
    "type": "metric",
    "parameters": [
      { "name": "instant", "value_from": "input" },
      { "name": "start", "value_from": "input" },
      { "name": "end", "value_from": "input" },
      { "name": "step", "value_from": "input" }
    ]
  },
  "query": "查询最近 7 天的药品上市总数",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "now_ms=1762996342241 timezone=Asia/Shanghai",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出（instant=true，但需要 start/end 定义聚合统计的时间范围）：
```json
{
  "approved_drug_count": { "instant": true, "start": 1762391942241, "end": 1762996342241 }
}
```

### 样例 1-C：即时查询（带时间范围：最近一年的总数）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "approved_drug_count",
    "type": "metric",
    "parameters": [
      { "name": "instant", "value_from": "input" },
      { "name": "start", "value_from": "input" },
      { "name": "end", "value_from": "input" },
      { "name": "step", "value_from": "input" }
    ]
  },
  "query": "查询最近一年的药品上市总数",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出（关键词"总数"强制 instant=true，start/end 为最近一年）：
```json
{
  "approved_drug_count": { "instant": true, "start": 1731460342241, "end": 1762996342241 }
}
```

> **重点**：虽然包含"最近一年"的时间范围，但"总数"表示要一个聚合值（单个结果），所以是 instant=true，而不是 instant=false。

### 样例 2：趋势查询（最近半年，按月）⚠️ 对比样例

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "统计最近半年药品上市数量趋势（按月）",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "now_ms=1762996342241 timezone=Asia/Shanghai",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出：
```json
{
  "approved_drug_count": { "instant": false, "start": 1747048342241, "end": 1762996342241, "step": "month" }
}
```

> **对比说明**：
> - "最近一年的**总数**" → instant=true（返回单个聚合值）
> - "最近半年的**趋势**（按月）" → instant=false（返回时间序列，每个月一个值）

### 样例 3：趋势查询（最近 4 周，按周）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "近4周每周的药品上市数量变化",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "now_ms=1762996342241 timezone=Asia/Shanghai",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出：
```json
{
  "approved_drug_count": { "instant": false, "start": 1760577142241, "end": 1762996342241, "step": "week" }
}
```

### 样例 4：趋势查询（今年以来，按季度）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "今年以来药品上市数量的季度趋势",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "now_ms=1762996342241 timezone=Asia/Shanghai",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出：
```json
{
  "approved_drug_count": { "instant": false, "start": 1735660800000, "end": 1762996342241, "step": "quarter" }
}
```

### 样例 5：未给时间范围（走默认兜底：近3个月，按月）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "统计药品上市数量趋势",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出（示例，具体 ask 文字可更短）：
```json
{
  "approved_drug_count": { "instant": false, "start": 1755216342241, "end": 1762996342241, "step": "month" }
}
```

### 样例 6：produced_drug_count（只依赖 company_name 的 property 参数）

INPUT_JSON：
```json
{
  "logic_property": {
    "name": "produced_drug_count",
    "type": "metric",
    "parameters": [
      { "name": "company_name", "value_from": "property" },
      { "name": "instant", "value_from": "input" },
      { "name": "start", "value_from": "input" },
      { "name": "end", "value_from": "input" },
      { "name": "step", "value_from": "input" }
    ]
  },
  "query": "查询最新的生产药品总数",
  "unique_identities": [{ "company_id": "company_000001", "company_name": "某药企A" }],
  "additional_context": "now_ms=1762996342241 timezone=Asia/Shanghai",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出（不输出 start/end，走底层默认）：
```json
{
  "produced_drug_count": { "instant": true }
}
```

### 样例 7：多对象批量（同一属性，同一时间范围）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "对这批企业统计最近一年药品上市数量趋势（按季度）",
  "unique_identities": [
    { "company_id": "company_000001", "company_name": "某药企A" },
    { "company_id": "company_000002", "company_name": "某药企B" }
  ],
  "additional_context": "now_ms=1762996342241 timezone=Asia/Shanghai",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出：
```json
{
  "approved_drug_count": { "instant": false, "start": 1731456342241, "end": 1762996342241, "step": "quarter" }
}
```

### 样例 8：additional_context 已给定明确时间戳（模型直接复用）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "趋势查询药品上市数量",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "{\"instant\":false,\"start\":1731460342241,\"end\":1762996342241,\"step\":\"month\"}",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出：
```json
{
  "approved_drug_count": { "instant": false, "start": 1731460342241, "end": 1762996342241, "step": "month" }
}
```

### 非成功样例 A：即时查询但缺 now_ms（无法给 start/end）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" } ] },
  "query": "查询最新药品上市数量",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": ""
}
```

期望输出：
```json
{ "_error": "missing approved_drug_count: now_ms | ask: need now_ms for instant=true" }
```

### 非成功样例 B：趋势查询但缺 now_ms 且也未给 end（无法默认兜底）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "统计药品上市数量趋势（按月）",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": ""
}
```

期望输出：
```json
{ "_error": "missing approved_drug_count: now_ms | ask: need now_ms (or provide end/start in additional_context)" }
```

### 非成功样例 C：传入的逻辑属性不是 metric（提示词不处理）

INPUT_JSON：
```json
{
  "logic_property": { "name": "business_health_score", "type": "operator", "parameters": [ { "name": "include_details", "value_from": "input" } ] },
  "query": "查询企业经营健康评分",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出：
```json
{ "_error": "not_metric business_health_score | ask: use operator prompt" }
```

### 非成功样例 D：parameters 缺少系统项（无法生成可执行参数）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "company_id", "value_from": "property" } ] },
  "query": "统计最近半年药品上市数量趋势（按月）",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "",
  "now_ms": 1762996342241,
  "timezone": "Asia/Shanghai"
}
```

期望输出：
```json
{ "_error": "invalid_parameters approved_drug_count | ask: missing system params instant/start/end/step in logic_property.parameters" }
```

### 非成功样例 E：上下文自相矛盾且缺关键字段（建议直接报错让上游统一口径）

INPUT_JSON：
```json
{
  "logic_property": { "name": "approved_drug_count", "type": "metric", "parameters": [ { "name": "instant", "value_from": "input" }, { "name": "start", "value_from": "input" }, { "name": "end", "value_from": "input" }, { "name": "step", "value_from": "input" } ] },
  "query": "查询最新药品上市数量",
  "unique_identities": [{ "company_id": "company_000001" }],
  "additional_context": "{\"instant\":false,\"step\":\"month\"}"
}
```

期望输出：
```json
{ "_error": "conflict approved_drug_count | ask: query=latest needs instant=true + now_ms, or provide explicit start/end for trend" }
```


