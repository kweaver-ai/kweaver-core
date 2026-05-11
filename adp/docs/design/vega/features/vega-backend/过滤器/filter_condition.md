# 过滤器语法规则说明


## 1. 概述

该过滤器语法用于统一描述查询条件，支持：

* 多层级组合条件
* 多种比较/匹配/范围/存在性判断
* 使用常量、字段值、用户属性作为比较值
* 前端按字段类型动态限制可选操作符

过滤器当前约束为：

* 最多支持 **3 层过滤条件**
* 每层最多支持 **10 个条件**

后续可按需放开限制。

---

## 2. 基本语法结构

过滤条件分为两类：

1. **逻辑组合节点**：`and` / `or`
2. **叶子条件节点**：具体比较或匹配操作

---

## 3. 数据结构

### 3.1 逻辑组合节点

当 `operation` 为 `and` 或 `or` 时，结构如下：

```json
{
  "operation": "and|or",
  "sub_conditions": []
}
```

语义说明：

* `and`：所有子条件都必须满足
* `or`：任意一个子条件满足即可  

---

### 3.2 普通条件节点

当 `operation` 为具体过滤操作时，结构如下：

```json
{
  "operation": "other",
  "field": "字段名称",
  "value_from": "const|field|user",
  "value": {}
}
```

字段说明：

* `operation`：过滤操作符
* `field`：被过滤的字段名
* `value_from`：右值来源，默认 `const`
* `value`：比较值。可为字符串、数字、数组等
* 对于 `exist`、`not_exist`、`null`、`not_null`、`true`、`false`、`empty`、`not_empty` 等操作，通常 **不需要 `value`**

补充字段：

* `before` 需要 `unit`
* `multi_match` 需要 `fields`，可选 `match_type`
* `knn` / `knn_vector` 需要 `limit_key`、`limit_value`  

---

## 4. 值来源（value_from）

支持以下取值：

| 值       | 含义                          |
| ------- | --------------------------- |
| `const` | 常量，`value` 即直接参与比较的值        |
| `field` | 字段，`value` 表示另一个字段名，用于字段间比较 |
| `user`  | 用户属性，`value` 表示当前用户的某个属性字段  |

说明：

* `const`：最常见
* `field`：用于“字段A 与 字段B 比较”
* `user`：用于“字段 与 当前用户属性比较”  

---

## 5. 过滤条件示例

```json
{
  "condition": {
    "operation": "and",
    "sub_conditions": [
      {
        "operation": "or",
        "sub_conditions": [
          {
            "operation": "==",
            "field": "f1",
            "value_from": "const",
            "value": "123"
          },
          {
            "operation": "!=",
            "field": "f2",
            "value_from": "field",
            "value": "f3"
          }
        ]
      },
      {
        "operation": "==",
        "field": "f4",
        "value_from": "user",
        "value": "group"
      }
    ]
  }
}
```

该示例表示：

* 外层为 `and`
* 第一部分是一个 `or` 组合：

  * `f1 == "123"`
  * 或 `f2 != f3`
* 第二部分为：

  * `f4 == 当前用户.group` 

---

## 6. 操作符总览

### 6.1 逻辑操作符

| 操作符   | 含义           |
| ----- | ------------ |
| `and` | 逻辑与，所有子条件均满足 |
| `or`  | 逻辑或，任意子条件满足  |



---

### 6.2 比较类操作符

| 操作符  | 别名              | 含义   | 右值要求 |
| ---- | --------------- | ---- | ---- |
| `==` | `eq`            | 等于   | 单值   |
| `!=` | `not_eq` | 不等于  | 单值   |
| `<`  | `lt`            | 小于   | 单值   |
| `<=` | `lte`           | 小于等于 | 单值   |
| `>`  | `gt`            | 大于   | 单值   |
| `>=` | `gte`           | 大于等于 | 单值   |

适用类型：

* `==` / `!=`：`string`、`text`、`integer`、`unsigned integer`、`float`、`decimal`、`ip`、`boolean`、`date`、`time`、`datetime`
* `<` / `<=` / `>` / `>=`：数值类和时间类字段 

---

### 6.3 集合类操作符

| 操作符           | 含义  | 右值要求  |
| ------------- | --- | ----- |
| `in`          | 属于  | 数组    |
| `not_in`      | 不属于 | 数组    |
| `contain`     | 包含  | 单值或数组 |
| `not_contain` | 不包含 | 单值或数组 |

说明：

#### `in` / `not_in`

* 右侧值必须是 **一个或多个相同类型值组成的数组**
* 判断字段值是否在给定数组中

#### `contain` / `not_contain`

* 左侧字段值应为 **数组**
* 右侧可为单值或数组
* 若右侧为数组：

  * `contain`：右侧数组中的值都应出现在左侧字段值中
  * `not_contain`：右侧数组中的值都不应出现在左侧字段值中

适用类型：

* `contain` / `not_contain`：`json`、`string`、`text`、`integer`、`unsigned integer`、`float`、`decimal`、`ip`  

---

### 6.4 字符串匹配类操作符

| 操作符          | 含义       | 右值要求  |
| ------------ | -------- | ----- |
| `like`       | 相似，查找子串  | 字符串   |
| `not_like`   | 不相似，查找子串 | 字符串   |
| `prefix`     | 以前缀开头    | 单值    |
| `not_prefix` | 不以前缀开头   | 单值    |
| `regex`      | 正则匹配     | 正则字符串 |

适用类型：

* `string`
* `text`  

---

### 6.5 空值 / 存在性 / 布尔判断

| 操作符         | 含义      | 是否需要 value |
| ----------- | ------- | ---------- |
| `exist`     | 字段存在    | 否          |
| `not_exist` | 字段不存在   | 否          |
| `empty`     | 空字符串    | 否          |
| `not_empty` | 非空字符串   | 否          |
| `null`      | 为 null  | 否          |
| `not_null`  | 不为 null | 否          |
| `true`      | 为 true  | 否          |
| `false`     | 为 false | 否          |

说明：

* `empty` / `not_empty` 主要用于字符串字段
* `null` / `not_null` / `exist` / `not_exist` 可用于多数类型
* `true` / `false` 用于布尔字段  

---

### 6.6 范围与区间操作符

| 操作符         | 含义  | 右值要求      | 边界规则                |
| ----------- | --- | --------- | ------------------- |
| `range`     | 范围内 | 长度为 2 的数组 | 左闭右开 `[a, b)`       |
| `out_range` | 范围外 | 长度为 2 的数组 | 小于 `a` 或大于等于 `b`    |
| `between`   | 介于  | 长度为 2 的数组 | 双闭区间 `[start, end]` |

说明：

#### `range`

```json
{
  "operation": "range",
  "field": "price",
  "value_from": "const",
  "value": [100, 1000]
}
```

表示：

* `100 <= price < 1000`

#### `out_range`

```json
{
  "operation": "out_range",
  "field": "price",
  "value_from": "const",
  "value": [0, 50]
}
```

表示：

* `price < 0` 或 `price >= 50`

#### `between`

```json
{
  "operation": "between",
  "field": "date",
  "value_from": "const",
  "value": ["2026-01-01", "2026-12-31"]
}
```

表示：

* `date` 位于闭区间 `[2026-01-01, 2026-12-31]` 内  

---

### 6.7 时间语义操作符

| 操作符       | 含义       | 右值要求  | 额外字段   |
| --------- | -------- | ----- | ------ |
| `before`  | 过去某段时间之前 | 单值    | `unit` |
| `current` | 当前时间单位   | 单值字符串 | 无      |

#### `before`

适用字段类型：

* `date`
* `time`
* `datetime`

支持单位：

* `year`
* `month`
* `week`
* `day`
* `hour`
* `minute`

示例：

```json
{
  "operation": "before",
  "field": "created_at",
  "value_from": "const",
  "value": 1,
  "unit": "day"
}
```

#### `current`

示例：

```json
{
  "operation": "current",
  "field": "updated_at",
  "value_from": "const",
  "value": "day"
}
```

 

---

### 6.8 全文搜索类操作符

| 操作符            | 含义      | 右值要求 | 备注          |
| -------------- | ------- | ---- | ----------- |
| `match`        | 全文匹配    | 单值   | SQL 类不支持    |
| `match_phrase` | 短语匹配    | 单值   | SQL 类不支持    |
| `multi_match`  | 多字段全文匹配 | 单值   | 需传 `fields` |

#### `multi_match` 结构

```json
{
  "operation": "multi_match",
  "field": "*",
  "value_from": "const",
  "value": "search term",
  "fields": ["title", "content", "description"],
  "match_type": "best_fields"
}
```

`match_type` 可选值：

* `best_fields`
* `most_fields`
* `cross_fields`
* `phrase`
* `phrase_prefix`
* `bool_prefix`

含义简述：

* `best_fields`：取单字段最高匹配分
* `most_fields`：累加多字段匹配分
* `cross_fields`：跨字段匹配查询词
* `phrase` / `phrase_prefix`：短语或短语前缀匹配
* `bool_prefix`：布尔 + 前缀匹配

 

---

### 6.9 向量检索类操作符

| 操作符          | 含义   | 右值要求    | 备注       |
| ------------ | ---- | ------- | -------- |
| `knn`        | 语义搜索 | 单值（字符串） | SQL 类不支持 |
| `knn_vector` | 向量查询 | 数组      | 前端不需要支持  |

#### `knn`

```json
{
  "field": "embedding",
  "operation": "knn",
  "value": "",
  "value_from": "const",
  "limit_key": "k",
  "limit_value": 3000
}
```

补充规则：

* `limit_key` 可选：

  * `k`
  * `max_distance`
  * `min_score`
* `limit_value` 取值范围：

  * `[1, 10000]`

#### `knn_vector`

```json
{
  "field": "embedding",
  "operation": "knn_vector",
  "value": [2, 3, 4, 5],
  "value_from": "const",
  "limit_key": "k",
  "limit_value": 3000
}
```

说明：

* 左侧字段必须是向量字段
* `knn_vector` 的 `value` 为向量数组
* 文档说明前端暂不需要支持 `knn_vector`  

---

## 7. 按字段类型支持的操作符

以下为文档中的字段类型与操作符映射整理。

### 7.1 all Fields

```text
match, match_phrase, multi_match
```

### 7.2 string

```text
==, !=, in, not_in, like, not_like, prefix, not_prefix, regex,
contain, not_contain, empty, not_empty, exist, not_exist, null, not_null
```

### 7.3 text

```text
==, !=, like, not_like, prefix, not_prefix, regex,
contain, not_contain, empty, not_empty, match, match_phrase, multi_match,
exist, not_exist, null, not_null
```

### 7.4 integer / float / decimal

```text
==, !=, <, <=, >, >=, in, not_in, range, out_range, between,
contain, not_contain, exist, not_exist, null, not_null
```

### 7.5 date / time / datetime

```text
==, !=, <, <=, >, >=, range, out_range, before, current, between,
exist, not_exist, null, not_null
```

### 7.6 ip

```text
==, !=, in, not_in, contain, not_contain, exist, not_exist, null, not_null
```

### 7.7 boolean

```text
==, !=, true, false, exist, not_exist, null, not_null
```

### 7.8 json

```text
contain, not_contain, exist, not_exist, null, not_null
```

### 7.9 vector

```text
knn, exist, not_exist, null, not_null
```

### 7.10 binary / point / shape / other

```text
exist, not_exist, null, not_null
```



---

## 8. 前端聚合类型映射

文档还定义了 UI 层聚合类型，便于前端统一展示操作符范围。

| 前端类型         | 映射来源                                 |
| ------------ | ------------------------------------ |
| `all Fields` | `typeOperationMapping['all Fields']` |
| `textString` | `string + text` 去重                   |
| `number`     | `integer + float + decimal` 去重       |
| `date`       | `date + time + datetime` 去重          |
| `ip`         | `ip`                                 |
| `boolean`    | `boolean`                            |
| `json`       | `json`                               |
| `vector`     | `vector`                             |
| `binary`     | `binary`                             |
| `point`      | `point`                              |
| `shape`      | `shape`                              |
| `other`      | `other`                              |

---

## 9. 编写规则建议

基于原文规则，实际使用时建议遵循以下约束：

### 9.1 逻辑节点规则

* `and` / `or` 必须带 `sub_conditions`
* `sub_conditions` 应为数组
* `and` / `or` 节点不应直接带 `field`

### 9.2 叶子节点规则

* 必须有 `operation`
* 除纯存在/布尔/空值类操作外，通常需要 `field`
* `value_from` 缺省视为 `const`

### 9.3 value 类型规则

* 比较类：单值
* `in` / `not_in`：数组
* `range` / `out_range` / `between`：长度为 2 的数组
* `contain` / `not_contain`：单值或数组
* `before`：单值 + `unit`
* `current`：字符串时间单位
* `knn_vector`：向量数组
* `multi_match`：单值 + `fields`

### 9.4 兼容性注意事项

文档中明确提到：

* `match`、`match_phrase`、`knn` 不支持 SQL 类查询
* `contain` 可能依赖 `json_array_contain` 和 etrino 逻辑
* 某些前端交互和错误提示仍有遗留问题

因此在实现 DSL 转 SQL、SQL 转执行器时，需要按后端能力做校验和降级。

---

## 10. 最小可用示例

### 10.1 单条件

```json
{
  "operation": "==",
  "field": "status",
  "value_from": "const",
  "value": "active"
}
```

### 10.2 AND 组合

```json
{
  "operation": "and",
  "sub_conditions": [
    {
      "operation": ">=",
      "field": "age",
      "value_from": "const",
      "value": 18
    },
    {
      "operation": "not_empty",
      "field": "name"
    }
  ]
}
```

### 10.3 OR 组合

```json
{
  "operation": "or",
  "sub_conditions": [
    {
      "operation": "==",
      "field": "status",
      "value_from": "const",
      "value": "active"
    },
    {
      "operation": "==",
      "field": "status",
      "value_from": "const",
      "value": "pending"
    }
  ]
}
```

---

## 11. 总结

该过滤器 DSL 的核心设计可以概括为：

* 用 `and` / `or` 表达条件树
* 用统一的叶子结构表达具体判断
* 用 `value_from` 支持常量、字段、用户属性三种右值来源
* 用字段类型映射约束可选操作符
* 同时兼容常规过滤、时间语义、全文检索和向量检索场景

---
