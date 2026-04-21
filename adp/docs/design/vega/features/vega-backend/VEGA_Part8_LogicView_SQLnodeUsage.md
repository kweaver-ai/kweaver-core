# SQL 节点模板使用指南

## 概述

SQL 节点支持使用 Go 模板语法编写自定义 SQL，可以引用输入节点的 SQL 并为其添加别名。

## 模板上下文

在 SQL 模板中，可以通过以下两种方式引用输入节点：

### 1. 直接引用（向后兼容）

使用 `.节点ID` 直接获取带别名的子查询 SQL：

```sql
SELECT * FROM .node1
WHERE .node1.some_field = 'value'
```

展开后：
```sql
SELECT * FROM (SELECT ... FROM table1) AS node1
WHERE (SELECT ... FROM table1) AS node1.some_field = 'value'
```

### 2. 使用模板函数（推荐）

提供三个模板函数，灵活性更高：

#### `node(nodeID)` - 返回带别名的完整 SQL

```sql
SELECT a.field1, b.field2
FROM {{ node "node1" }} a
JOIN {{ node "node2" }} b ON a.id = b.ref_id
```

展开后：
```sql
SELECT a.field1, b.field2
FROM (SELECT ... FROM table1) AS node1_alias a
JOIN (SELECT ... FROM table2) AS node2_alias b ON a.id = b.ref_id
```

#### `nodeSQL(nodeID)` - 仅返回子查询 SQL

```sql
WITH base AS ({{ nodeSQL "node1" }})
SELECT * FROM base
```

展开后：
```sql
WITH base AS (SELECT ... FROM table1)
SELECT * FROM base
```

#### `nodeAlias(nodeID)` - 返回节点别名

```sql
SELECT {{ nodeAlias "node1" }}.field1
FROM {{ node "node1" }}
```

展开后：
```sql
SELECT node1_alias.field1
FROM (SELECT ... FROM table1) AS node1_alias
```

## 别名生成规则

节点别名通过 `sanitizeAlias()` 函数自动生成：

1. **替换特殊字符**：所有非字母数字字符替换为下划线
   - `node-1` → `node_1`
   - `node.1` → `node_1`
   - `my-node_1` → `my_node_1`

2. **数字开头处理**：如果以数字开头，添加 `n_` 前缀
   - `123node` → `n_123node`

3. **长度限制**：最大 60 个字符（MySQL 标识符最大 64 字符）

## 完整示例

### 示例 1：简单查询

**节点配置：**
- node1: 资源节点，查询 users 表
- node2: 资源节点，查询 orders 表
- sql_node: 自定义 SQL 节点，输入 [node1, node2]

**模板 SQL：**
```sql
SELECT 
    u.name,
    o.order_no,
    o.amount
FROM {{ node "node1" }} u
JOIN {{ node "node2" }} o ON u.id = o.user_id
WHERE o.amount > {{ .threshold }}
```

**生成 SQL：**
```sql
SELECT 
    u.name,
    o.order_no,
    o.amount
FROM (SELECT id, name, email FROM users) AS node1 u
JOIN (SELECT id, user_id, order_no, amount FROM orders) AS node2 o 
    ON u.id = o.user_id
WHERE o.amount > 100
```

### 示例 2：多层嵌套

**模板 SQL：**
```sql
WITH base_data AS (
    SELECT * FROM {{ node "node1" }}
    WHERE status = 'active'
)
SELECT 
    bd.*,
    derived.calculated_field
FROM base_data bd
CROSS JOIN (
    SELECT COUNT(*) as calculated_field 
    FROM {{ node "node2" }}
) derived
```

### 示例 3：动态字段引用

**模板 SQL：**
```sql
SELECT 
    {{ nodeAlias "node1" }}.id,
    {{ nodeAlias "node1" }}.name,
    {{ nodeAlias "node2" }}.related_info
FROM {{ node "node1" }}
LEFT JOIN {{ node "node2" }} 
    ON {{ nodeAlias "node1" }}.id = {{ nodeAlias "node2" }}.ref_id
```

## 注意事项

1. **节点 ID 必须存在**：如果模板中引用了不存在的节点 ID，会返回错误
2. **别名唯一性**：同一 SQL 节点内，不同输入节点的别名保证唯一
3. **向后兼容**：原有的 `.nodeID` 引用方式仍然可用，会自动展开为带别名的 SQL
4. **参数化查询**：所有参数使用 `?` 或 `$1, $2` 占位符，由 Squirrel 库处理

## 错误处理

| 错误场景 | 错误信息 |
|---------|---------|
| 节点 ID 不存在 | `unknown node ID in template: xxx` |
| 模板语法错误 | `failed to parse SQL template for node xxx: ...` |
| 模板执行错误 | `failed to execute SQL template for node xxx: ...` |
| 输入节点构建失败 | `failed to build sql node input xxx: ...` |

## 最佳实践

1. **使用模板函数而非直接引用**：`{{ node "node1" }}` 比 `.node1` 更清晰
2. **为子查询添加别名**：便于在 SELECT/WHERE/JOIN 中引用
3. **避免过深嵌套**：递归深度受 `MaxRecursionDepth` 限制
4. **测试复杂模板**：使用单元测试验证生成的 SQL 是否正确
