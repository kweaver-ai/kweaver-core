# MariaDB Resource AT 测试

## 概述

本目录包含 MariaDB Resource 的验收测试（AT 测试）。MariaDB Resource 是物理 Resource，通过 Catalog 的 Discovery 机制自动发现，关联到实际的 MariaDB 数据源。

> **注意**：MariaDB Resource 与 Dataset Resource 有本质区别：
> - **不允许创建**：Resource 通过 Catalog Discovery 自动发现，不可手动创建
> - **不允许删除**：物理 Resource 由数据源管理，不可删除
> - **受限更新**：仅允许修改显示名称、描述、标签等元数据，不可修改 schema 等核心字段

## 测试文件

| 文件 | 描述 |
|------|------|
| `resource_test.go` | MariaDB Resource 测试入口 |

## 测试用例清单

---

### 读取测试（MR1xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR101 | 获取存在的 MariaDB Resource | 200 OK |
| MR102 | 获取不存在的 Resource | 404 Not Found |
| MR103 | 列表查询 - 按 catalog_id 过滤 | 200 OK |
| MR104 | 列表查询 - 按 category=table 过滤 | 200 OK |
| MR105 | 列表查询 - 按 connector_type=mariadb 过滤 | 200 OK |
| MR106 | 列表分页测试 | 正确分页返回 |
| MR107 | 验证 Resource 包含完整的 schema 信息 | schema_definition 完整 |
| MR108 | 验证 Resource 包含正确的 catalog 关联 | catalog_id 正确 |

---

### 更新测试（MR2xx）

#### 允许的更新（MR201-MR205）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR201 | 更新 Resource 显示名称 | 204 No Content |
| MR202 | 更新 Resource 描述 | 204 No Content |
| MR203 | 更新 Resource 标签 | 204 No Content |
| MR204 | 同时更新名称、描述、标签 | 204 No Content |
| MR205 | 更新后验证修改生效 | 查询返回更新后的值 |

#### 禁止的更新（MR221-MR228）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR221 | 尝试修改 schema_definition | 400 Bad Request |
| MR222 | 尝试修改 connector_type | 400 Bad Request |
| MR223 | 尝试修改 category | 400 Bad Request |
| MR224 | 尝试修改 catalog_id | 400 Bad Request |
| MR225 | 尝试修改 config | 400 Bad Request |
| MR226 | 尝试修改 original_name | 400 Bad Request |
| MR227 | 更新不存在的 Resource | 404 Not Found |
| MR228 | 空更新请求体 | 400 Bad Request |

---

### 删除测试（MR3xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR301 | 删除 MariaDB Resource | 204 No Content |
| MR302 | 批量删除 MariaDB Resource | 204 No Content |

---

### 数据查询测试（MR4xx）

#### 基础查询（MR401-MR408）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR401 | 查询 Resource 数据 - 基本场景 | 200 OK，返回数据列表 |
| MR402 | 查询 Resource 数据 - 分页 | 200 OK，正确分页 |
| MR403 | 查询 Resource 数据 - 指定字段 | 200 OK，仅返回指定字段 |
| MR404 | 查询 Resource 数据 - 排序 | 200 OK，正确排序 |
| MR405 | 查询 Resource 数据 - 限制返回条数 | 200 OK，返回指定条数 |
| MR406 | 查询空表数据 | 200 OK，entries 为空 |
| MR407 | 查询不存在的 Resource 数据 | 404 Not Found |
| MR408 | 验证返回数据字段与 schema 一致 | 字段类型匹配 |

#### 过滤条件查询（MR411-MR425）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR411 | 查询 - 等于条件（eq） | 200 OK，正确过滤 |
| MR412 | 查询 - 不等于条件（neq） | 200 OK，正确过滤 |
| MR413 | 查询 - 大于条件（gt） | 200 OK，正确过滤 |
| MR414 | 查询 - 大于等于条件（gte） | 200 OK，正确过滤 |
| MR415 | 查询 - 小于条件（lt） | 200 OK，正确过滤 |
| MR416 | 查询 - 小于等于条件（lte） | 200 OK，正确过滤 |
| MR417 | 查询 - IN 条件 | 200 OK，正确过滤 |
| MR418 | 查询 - NOT IN 条件 | 200 OK，正确过滤 |
| MR419 | 查询 - LIKE 条件（模糊匹配） | 200 OK，正确过滤 |
| MR420 | 查询 - IS NULL 条件 | 200 OK，正确过滤 |
| MR421 | 查询 - IS NOT NULL 条件 | 200 OK，正确过滤 |
| MR422 | 查询 - BETWEEN 条件 | 200 OK，正确过滤 |
| MR423 | 查询 - 组合条件（AND） | 200 OK，正确过滤 |
| MR424 | 查询 - 组合条件（OR） | 200 OK，正确过滤 |
| MR425 | 查询 - 嵌套组合条件 | 200 OK，正确过滤 |

#### 查询边界测试（MR431-MR436）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR431 | 查询 - offset 超出范围 | 200 OK，entries 为空 |
| MR432 | 查询 - limit 最大值 | 200 OK |
| MR433 | 查询 - limit=0 | 400 Bad Request |
| MR434 | 查询 - 无效排序字段 | 400 Bad Request |
| MR435 | 查询 - 无效过滤字段 | 400 Bad Request |
| MR436 | 查询 - 无效过滤操作符 | 400 Bad Request |

---

### 统计查询测试（MR5xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR501 | 统计查询 - 总数统计 | 200 OK，返回 total_count |
| MR502 | 统计查询 - 分组统计 | 200 OK，返回分组结果 |
| MR503 | 统计查询 - 聚合函数（SUM/AVG/MIN/MAX） | 200 OK，返回聚合结果 |
| MR504 | 统计查询 - 带过滤条件 | 200 OK，返回过滤后统计 |

## 运行测试

```bash
# 运行所有 MariaDB Resource 测试
go test -v ./tests/at/resource/table/mairadb/...

# 运行读取测试
go test -v ./tests/at/resource/table/mairadb/... -run TestMariaDBResourceRead

# 运行更新测试
go test -v ./tests/at/resource/table/mairadb/... -run TestMariaDBResourceUpdate

# 运行数据查询测试
go test -v ./tests/at/resource/table/mairadb/... -run TestMariaDBResourceQuery

# 运行特定用例（通过用例ID）
go test -v ./tests/at/resource/table/mairadb/... -run MR101
```
