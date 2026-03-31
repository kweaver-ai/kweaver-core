# OpenSearch Catalog AT 测试

## 概述

本目录包含 OpenSearch Catalog 的验收测试（AT 测试）。OpenSearch Catalog 是物理 Catalog，连接到实际的 OpenSearch 数据源。

> **注意**：通用字段测试（name/description/tags 边界验证）已在 `catalog/logical` 中覆盖，此处仅测试 OpenSearch 特有功能。

## 测试文件

| 文件 | 描述 |
|------|------|
| `catalog_test.go` | OpenSearch Catalog CRUD 测试入口 |
| `builder.go` | OpenSearch Payload 构建器 |

## 测试用例清单

### 创建测试（OS1xx）

#### 正向测试（OS101-OS119）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| OS101 | 创建 OpenSearch catalog - 基本场景 | 201 Created |
| OS102 | 创建后验证 connector_type 为 opensearch | connector_type = "opensearch" |
| OS103 | 创建后验证 type 为 physical | type = "physical" |
| OS104 | 创建 OpenSearch catalog - 完整字段 | 201 Created |
| OS105 | 创建带 SSL 配置的 catalog（SSL 禁用） | 201 Created |
| OS106 | 创建后立即查询 | 查询返回一致数据 |
| OS107 | OpenSearch 连接测试成功 | 200 OK |
| OS108 | 获取 OpenSearch catalog 健康状态 | 200 OK |

#### connector_config 负向测试（OS121-OS129）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| OS121 | 缺少 host 字段 | 400 Bad Request |
| OS122 | 缺少 port 字段 | 400 Bad Request |
| OS123 | 缺少认证信息（无 username） | 400 Bad Request |
| OS124 | 空用户名 | 400 Bad Request |
| OS125 | 错误凭证 | 400 Bad Request |
| OS126 | 无效端口（非数字） | 400 Bad Request |
| OS127 | 超出范围端口（65536） | 400 Bad Request |
| OS128 | 负数端口 | 400 Bad Request |
| OS129 | 无效 host | 400 Bad Request |

#### 边界测试（OS131-OS139）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| OS131 | port 边界值（1） | 201 Created 或 400 |
| OS132 | port 边界值（65535） | 201 Created 或 400 |
| OS133 | host 为 IP 地址 | 201 Created |
| OS134 | host 为域名 | 201 Created 或 400 |
| OS135 | password 为空 | 400 Bad Request |
| OS136 | 使用 HTTPS 协议 | 201 Created 或 400 |
| OS137 | 使用 HTTP 协议 | 201 Created |

---

### 读取测试（OS2xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| OS201 | 获取存在的 OpenSearch catalog | 200 OK |
| OS202 | 列表查询 - 按 type 过滤 physical | 200 OK |
| OS203 | 列表查询 - 按 connector_type 过滤 opensearch | 200 OK |
| OS204 | 查询 catalog - 验证所有字段返回 | 200 OK |
| OS205 | 验证 connector_config.password 不返回 | password 字段不存在 |

---

### 更新测试（OS3xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| OS301 | 整体更新 connector_config | 204 No Content |
| OS302 | 更新 connector_config 后连接测试 | 200 OK |
| OS303 | 更新 host 为无效地址 | 400 Bad Request |
| OS304 | 更新 port 为无效值 | 400 Bad Request |
| OS305 | 更新 password | 204 No Content |

---

### 删除测试（OS4xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| OS401 | 删除 OpenSearch catalog 后健康状态不可查 | 404 Not Found |
| OS402 | 删除 OpenSearch catalog 后不能测试连接 | 404 Not Found |

---

### OpenSearch 特有测试（OS5xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| OS501 | SSL/TLS 连接测试（跳过验证） | 201 Created 或 400 |
| OS502 | 自定义 index pattern 选项 | 201 Created |
| OS503 | 连接超时选项测试 | 201 Created |
| OS504 | 多节点配置测试（如支持） | 201 Created 或 400 |

## 运行测试

```bash
# 运行所有 OpenSearch Catalog 测试
go test -v ./tests/at/catalog/physical/opensearch/...

# 运行创建测试
go test -v ./tests/at/catalog/physical/opensearch/... -run TestOpenSearchSpecificCreate

# 运行特定用例
go test -v ./tests/at/catalog/physical/opensearch/... -run OS101
```
