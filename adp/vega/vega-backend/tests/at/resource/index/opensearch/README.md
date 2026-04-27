# OpenSearch Resource AT 测试

## 概述

本目录包含 OpenSearch Resource 的验收测试（AT 测试），运行通用 Resource 测试用例。

## 测试文件

| 文件 | 描述 |
|------|------|
| `resource_test.go` | 通用测试入口，运行所有 RMxxx 系列测试 |

## 测试用例清单

### 通用测试（RMxxx）

来自 `resource/internal/test_cases.go`，与 MariaDB 共用，详见 [../mariadb/README.md](../mariadb/README.md)。

## 运行测试

```bash
# 运行所有 OpenSearch Resource 测试
go test -v ./tests/at/resource/opensearch/...

# 运行特定测试套件
go test -v ./tests/at/resource/opensearch/... -run TestOpenSearchResourceCommon

# 运行单个用例
go test -v ./tests/at/resource/opensearch/... -run RM101
```
