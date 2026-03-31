# Logical Catalog AT 测试

## 概述

本目录包含 Logical Catalog 的验收测试（AT 测试）。Logical Catalog 是虚拟的元数据容器，不连接实际数据源。

## 测试文件

| 文件 | 描述 |
|------|------|
| `catalog_test.go` | Logical Catalog CRUD 测试入口 |
| `internal/helpers.go` | 辅助函数和 Payload 构建器 |

## 测试用例清单

### 创建测试（LG1xx）

#### 正向测试（LG101-LG119）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| LG101 | 创建 logical catalog - 基本场景 | 201 Created |
| LG102 | 创建 logical catalog - 完整字段 | 201 Created |
| LG103 | 创建后验证 type 为 logical | type = "logical" |
| LG104 | 创建后立即查询 | 查询返回一致数据 |
| LG105 | 创建多个 logical catalog，列表查询 | 列表返回正确 |
| LG106 | logical catalog 无 connector_type | connector_type 为空 |

#### 负向测试（LG121-LG129）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| LG121 | 重复的 catalog 名称 | 409 Conflict |
| LG122 | 缺少必填字段 - name | 400 Bad Request |
| LG123 | 空字符串 name | 400 Bad Request |
| LG124 | name 超过最大长度（255字符） | 400 Bad Request |
| LG125 | description 超过最大长度（1000字符） | 400 Bad Request |
| LG126 | 单个 tag 超过最大长度（40字符） | 400 Bad Request |
| LG127 | tags 数量超过限制（5个） | 400 Bad Request |
| LG128 | tag 包含无效字符 | 400 Bad Request |

#### 边界测试（LG131-LG139）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| LG131 | name 最大长度（255字符） | 201 Created |
| LG132 | description 最大长度（1000字符） | 201 Created |
| LG133 | name 包含中文 | 201 Created |
| LG134 | name 包含特殊字符（下划线、连字符） | 201 Created |
| LG135 | tags 为空数组 | 201 Created |
| LG136 | tags 包含最大数量（5个） | 201 Created |
| LG137 | 单个 tag 最大长度（40字符） | 201 Created |
| LG138 | description 为空字符串 | 201 Created |
| LG139 | logical catalog 不需要 connector_config | 201 Created |

---

### 读取测试（LG2xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| LG201 | 获取存在的 logical catalog | 200 OK |
| LG202 | 获取不存在的 catalog | 404 Not Found |
| LG203 | 列表查询 - 按 type 过滤 logical | 200 OK |
| LG204 | 列表分页测试 | 正确分页返回 |
| LG205 | 列表查询 - 按 name 模糊搜索 | 200 OK |
| LG206 | 列表查询 - 空结果 | 200 OK，entries 为空 |

---

### 更新测试（LG3xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| LG301 | 更新 logical catalog 名称 | 204 No Content |
| LG302 | 更新 logical catalog 描述 | 204 No Content |
| LG303 | 更新不存在的 catalog | 404 Not Found |
| LG304 | 验证 update_time 更新 | update_time 已更新 |
| LG305 | 更新 tags | 204 No Content |
| LG306 | 更新为已存在的 name | 409 Conflict |

---

### 删除测试（LG4xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| LG401 | 删除存在的 logical catalog | 204 No Content |
| LG402 | 删除不存在的 catalog | 404 Not Found |
| LG403 | 重复删除同一 catalog | 404 Not Found |
| LG404 | 删除后可以创建同名 catalog | 201 Created |

---

### Logical 特有测试（LG5xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| LG501 | logical catalog 测试连接 | 200 OK |
| LG502 | logical catalog 健康检查 | 200 OK |

## 运行测试

```bash
# 运行所有 Logical Catalog 测试
go test -v ./tests/at/catalog/logical/...

# 运行单个测试函数
go test -v ./tests/at/catalog/logical/... -run TestLogicalCatalogCreate

# 运行特定用例
go test -v ./tests/at/catalog/logical/... -run LG101
```
