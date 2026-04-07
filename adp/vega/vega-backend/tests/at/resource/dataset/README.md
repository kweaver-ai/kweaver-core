# 数据集资源 AT 测试

## 概述

本目录包含数据集资源（Dataset Resource）的验收测试（AT 测试）。数据集资源是系统中用于管理和操作数据集的核心组件，支持数据集的创建、查询、更新、删除等操作。

## 测试文件

| 文件 | 描述 |
|------|------|
| `dataset_test.py` | 数据集资源 CRUD 测试入口 |

## 测试用例清单

### 数据集资源测试（DT1xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| DT101 | 创建数据集资源 - 基本场景 | 201 Created |
| DT102 | 获取存在的数据集资源 | 200 OK |
| DT103 | 更新数据集资源 | 204 No Content |
| DT104 | 删除数据集资源 | 204 No Content |
| DT105 | 列出所有数据集资源 | 200 OK |

### 数据集文档测试（DT2xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| DT201 | 创建数据集文档 | 201 Created |
| DT202 | 列出数据集文档 | 200 OK |
| DT203 | 获取数据集文档 | 200 OK |
| DT204 | 更新数据集文档 | 204 No Content |
| DT205 | 删除数据集文档 | 204 No Content |

## 运行测试

```bash
cd tests
# 运行所有 dataset 测试
go test -v ./at/resource/dataset/...
# 运行指定用例
go test -v ./at/resource/dataset/... -run TestDatasetDocumentsList
```