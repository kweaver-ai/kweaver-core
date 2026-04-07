# 数据类型映射规则表

## 概述

本文档以表格形式展示 flow-automation 项目中各数据库的数据类型映射规则。

## 数据类型映射对照表

### 数值类型

| 通用类型 | MySQL | PostgreSQL | Oracle | SQL Server | DM8 | KDB | 说明 |
|---------|--------|------------|--------|------------|-----|-----|------|
| TINYINT | TINYINT | SMALLINT | NUMBER(3) | TINYINT | TINYINT | TINYINT | 8位整数 |
| SMALLINT | SMALLINT | SMALLINT | NUMBER(5) | SMALLINT | SMALLINT | SMALLINT | 16位整数 |
| MEDIUMINT | MEDIUMINT | INTEGER | NUMBER | INTEGER | INTEGER | INTEGER | 24位整数 |
| INT | INT | INTEGER | NUMBER(10) | INT | INTEGER | INTEGER | 32位整数 |
| INTEGER | INT | INTEGER | NUMBER(10) | INT | INTEGER | INTEGER | 32位整数 |
| BIGINT | BIGINT | BIGINT | NUMBER(19) | BIGINT | BIGINT | BIGINT | 64位整数 |
| DECIMAL | DECIMAL | DECIMAL | NUMBER | DECIMAL | DECIMAL | DECIMAL | 定点数 |
| NUMERIC | DECIMAL | DECIMAL | NUMBER | NUMERIC | DECIMAL | DECIMAL | 定点数 |
| FLOAT | FLOAT | REAL | BINARY_FLOAT | FLOAT | FLOAT | FLOAT | 单精度浮点 |
| REAL | FLOAT | REAL | BINARY_FLOAT | REAL | REAL | FLOAT | 单精度浮点 |
| DOUBLE | DOUBLE | DOUBLE PRECISION | BINARY_DOUBLE | FLOAT | DOUBLE | DOUBLE | 双精度浮点 |
| MONEY | DECIMAL | DECIMAL | NUMBER | MONEY | DECIMAL | DECIMAL | 货币类型 |
| SMALLMONEY | DECIMAL | DECIMAL | NUMBER | SMALLMONEY | DECIMAL | DECIMAL | 小额货币 |
| BIT | BIT | BOOLEAN | NUMBER(1) | BIT | BIT | BOOLEAN | 位类型 |

### 字符串类型

| 通用类型 | MySQL | PostgreSQL | Oracle | SQL Server | DM8 | KDB | 说明 |
|---------|--------|------------|--------|------------|-----|-----|------|
| CHAR | CHAR | CHAR | NVARCHAR2 | NVARCHAR | CHAR | CHAR | 定长字符串 |
| VARCHAR | VARCHAR | VARCHAR | NVARCHAR2 | NVARCHAR | VARCHAR | VARCHAR | 变长字符串 |
| STRING | VARCHAR | VARCHAR | NVARCHAR2 | NVARCHAR | VARCHAR | VARCHAR | 字符串 |
| NVARCHAR | VARCHAR | VARCHAR | NVARCHAR2 | NVARCHAR | VARCHAR | VARCHAR | Unicode字符串 |
| TEXT | TEXT | TEXT | CLOB | NVARCHAR | TEXT | TEXT | 长文本 |
| TINYTEXT | TINYTEXT | TEXT | CLOB | NVARCHAR | TEXT | TEXT | 短文本 |
| MEDIUMTEXT | MEDIUMTEXT | TEXT | CLOB | NVARCHAR | TEXT | TEXT | 中等长度文本 |
| LONGTEXT | LONGTEXT | TEXT | CLOB | NVARCHAR | TEXT | TEXT | 长文本 |
| CLOB | TEXT | TEXT | CLOB | NVARCHAR | TEXT | TEXT | 字符大对象 |

### 二进制类型

| 通用类型 | MySQL | PostgreSQL | Oracle | SQL Server | DM8 | KDB | 说明 |
|---------|--------|------------|--------|------------|-----|-----|------|
| BINARY | BINARY | BYTEA | BLOB | VARBINARY | BLOB | BLOB | 二进制数据 |
| VARBINARY | VARBINARY | BYTEA | BLOB | VARBINARY | BLOB | BLOB | 变长二进制数据 |
| BLOB | BLOB | BYTEA | BLOB | VARBINARY | BLOB | BLOB | 二进制大对象 |
| TINYBLOB | TINYBLOB | BYTEA | BLOB | VARBINARY | BLOB | BLOB | 小二进制对象 |
| MEDIUMBLOB | MEDIUMBLOB | BYTEA | BLOB | VARBINARY | BLOB | BLOB | 中二进制对象 |
| LONGBLOB | LONGBLOB | BYTEA | BLOB | VARBINARY | BLOB | BLOB | 大二进制对象 |
| IMAGE | BLOB | BYTEA | BLOB | IMAGE | BLOB | BLOB | 图像数据 |

### 时间日期类型

| 通用类型 | MySQL | PostgreSQL | Oracle | SQL Server | DM8 | KDB | 说明 |
|---------|--------|------------|--------|------------|-----|-----|------|
| DATE | DATE | DATE | DATE | DATE | DATE | DATE | 日期 |
| TIME | TIME | TIME | DATE | TIME | TIME | TIME | 时间 |
| DATETIME | DATETIME | TIMESTAMP | TIMESTAMP | DATETIME | TIMESTAMP | TIMESTAMP | 日期时间 |
| DATETIME2 | DATETIME | TIMESTAMP | TIMESTAMP | DATETIME2 | TIMESTAMP | TIMESTAMP | 扩展日期时间 |
| TIMESTAMP | TIMESTAMP | TIMESTAMP | TIMESTAMP | DATETIME | TIMESTAMP | TIMESTAMP | 时间戳 |
| YEAR | YEAR | INTEGER | NUMBER | INT | INTEGER | INTEGER | 年份 |

### 布尔类型

| 通用类型 | MySQL | PostgreSQL | Oracle | SQL Server | DM8 | KDB | 说明 |
|---------|--------|------------|--------|------------|-----|-----|------|
| BOOL | TINYINT(1) | BOOLEAN | NUMBER | BIT | BIT | BOOLEAN | 布尔值 |
| BOOLEAN | TINYINT(1) | BOOLEAN | NUMBER | BIT | BIT | BOOLEAN | 布尔值 |
| BIT | TINYINT(1) | BOOLEAN | NUMBER | BIT | BIT | BOOLEAN | 位类型 |

### 其他类型

| 通用类型 | MySQL | PostgreSQL | Oracle | SQL Server | DM8 | KDB | 说明 |
|---------|--------|------------|--------|------------|-----|-----|------|
| JSON | JSON | JSONB | CLOB | NVARCHAR | TEXT | TEXT | JSON数据 |
| UUID | VARCHAR(36) | UUID | RAW(16) | UNIQUEIDENTIFIER | VARCHAR(36) | VARCHAR(36) | 通用唯一标识符 |
| XML | TEXT | TEXT | CLOB | XML | TEXT | TEXT | XML数据 |
| ENUM | VARCHAR | VARCHAR | NVARCHAR2 | NVARCHAR | VARCHAR | VARCHAR | 枚举类型 |
| SET | VARCHAR | VARCHAR | NVARCHAR2 | NVARCHAR | VARCHAR | VARCHAR | 集合类型 |
| SQL_VARIANT | VARCHAR | TEXT | NVARCHAR2 | SQL_VARIANT | TEXT | TEXT | SQL变体类型 |

## UNSIGNED类型处理

### 当前实现
所有数据库驱动都正确处理UNSIGNED类型：

1. **检测UNSIGNED**: 检查类型名称是否包含"UNSIGNED"
2. **提取基础类型**: 分离UNSIGNED修饰符和基础类型
3. **查找映射**: 在映射表中查找基础类型
4. **重建类型**: 重新组合映射后的类型和UNSIGNED修饰符

### 示例
```go
// 输入: "TINYINT UNSIGNED"
// 处理:
// 1. 检测到UNSIGNED
// 2. 提取基础类型: "TINYINT"
// 3. 在映射表中查找: "TINYINT" -> 对应数据库的基础类型
// 4. 重建类型: 基础类型 + " UNSIGNED"
// 输出: 正确映射的UNSIGNED类型
```

### 支持的UNSIGNED类型
- TINYINT UNSIGNED → 各数据库对应的无符号8位整数
- SMALLINT UNSIGNED → 各数据库对应的无符号16位整数
- INT UNSIGNED → 各数据库对应的无符号32位整数
- BIGINT UNSIGNED → 各数据库对应的无符号64位整数

## 数据库特性说明

### MySQL 特性
- 使用 `TINYINT(1)` 表示布尔值
- 支持多种 TEXT 类型变体 (TINYTEXT, MEDIUMTEXT, LONGTEXT)
- 支持多种 BLOB 类型变体 (TINYBLOB, MEDIUMBLOB, LONGBLOB)
- 原生 JSON 类型支持
- 不支持 Schema 概念

### PostgreSQL 特性
- 使用 `BYTEA` 存储所有二进制数据
- 使用 `JSONB` 存储 JSON 数据（性能优于 JSON）
- 原生 `BOOLEAN` 类型支持
- 支持 Schema 概念
- 没有 MEDIUMINT 和 TINYINT，使用 INTEGER 和 SMALLINT 替代

### Oracle 特性
- 使用 `NUMBER` 类型表示所有数值类型
- 使用 `NVARCHAR2` 存储 Unicode 字符串
- 支持 `BINARY_FLOAT` 和 `BINARY_DOUBLE`
- 使用 `RAW` 类型存储 UUID
- 支持 Schema 概念（用户模式）

### SQL Server 特性
- 原生 `BIT` 类型表示布尔值
- 支持货币类型 `MONEY` 和 `SMALLMONEY`
- `UNIQUEIDENTIFIER` 类型用于 UUID/GUID
- 支持 `XML` 数据类型
- 支持 Schema 概念

### DM8 (达梦数据库) 特性
- 支持 `BIT` 类型表示布尔值
- 使用 `BLOB` 存储所有二进制数据
- 使用 `TEXT` 存储 JSON 数据（无原生 JSON 类型）
- 支持 Schema 概念

### KDB 特性
- 原生 `BOOLEAN` 类型支持
- 支持 Schema 概念
- 使用 `BLOB` 存储二进制数据

## 映射规则总结

### 统一性规则
1. **数值类型**: 各数据库使用各自的标准整数类型，Oracle统一使用 NUMBER(n)
2. **字符串类型**: 各数据库使用 VARCHAR/NVARCHAR 变体，Oracle使用 NVARCHAR2
3. **二进制类型**: PostgreSQL使用 BYTEA，其他数据库使用 BLOB/VARBINARY 变体
4. **时间类型**: 大多数使用标准时间类型，个别差异在 TIMESTAMP/DATETIME 处理
5. **布尔类型**: MySQL使用 TINYINT(1)，PostgreSQL/KDB使用 BOOLEAN，其他使用 BIT/NUMBER(1)

### 兼容性考虑
- 优先选择目标数据库的原生类型和最佳实践
- 保持数据精度和范围不丢失
- 考虑查询性能和存储效率
- 遵循各数据库的标准数据类型定义
- 支持 UNSIGNED 类型修饰符的正确处理

### 特殊处理
- **长度限制**: VARCHAR/CHAR 等类型根据数据库特性设置合理的长度限制
- **精度要求**: DECIMAL/NUMBER 类型支持动态精度和小数位数设置
- **字符集**: Unicode 字符串优先使用数据库的 Unicode 类型
- **大对象**: 长文本和二进制数据自动选择相应的大对象类型
- **映射表驱动**: 所有类型映射都基于 GetDataTypeMapping() 方法返回的映射表
- **动态参数处理**: 对于需要长度、精度等参数的类型，支持动态参数设置
- **兜底机制**: 当映射表中找不到对应类型时，提供合理的默认类型
