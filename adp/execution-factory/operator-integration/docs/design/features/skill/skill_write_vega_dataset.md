# Skill数据写入vega dataset设计文档

> **文档定位**：执行工厂侧实现设计文档  
> **关联需求文档**: [prd/skill_write_vega_dataset.md](../../product/prd/skill_write_vega_dataset.md)

## 1. 设计目标

- 为上游 `context-loader` 的 Skill召回提供稳定、可检索、可语义召回的 Skill 索引数据。
- 将执行工厂中的 Skill 主数据从关系型主表同步到 Vega Dataset。
- 在不改变现有 Skill 管理、读取、下载、删除主流程的前提下，补充独立的索引初始化与文档同步能力。

## 2. 设计边界

本设计只覆盖执行工厂侧的“索引生产能力”，不覆盖以下内容：
- 不实现 `context-loader` 的召回编排、排序和召回消费逻辑
- 不实现 `bkn` 中 Skill 与业务对象的绑定解释逻辑
- 不修改既有 Skill 管理接口语义
- 不为每个 Skill 创建独立 Vega resource

模块边界：
- `execution-factory`：Skill 权威元数据、向量生成、Vega Dataset 同步
- `bkn`：Skill 与知识网络/对象类/对象实例的业务绑定语义
- `context-loader`：运行时召回消费与使用

## 3. 总体架构

本设计拆分为两层：

### 3.1 初始化层

职责：
- 确认 Vega catalog 是否存在
- 确认共享 dataset resource 是否存在
- 在 resource 不存在时创建 resource 并写入 schema_definition

约束：
- 初始化流程必须幂等
- catalog/resource 为共享逻辑资源，不按 Skill 维度创建

### 3.2 文档同步层

职责：
- 将单个 Skill 映射为一条 dataset document
- 处理发布、更新、下架时的 document 增删改

约束：
- 下架删除的是 document，不是整个 dataset resource
- 同一个 Skill 在 dataset 中只保留一条当前有效文档

## 4. 数据源模型

主数据表：
- `t_skill_repository`

本次参与索引构建的源字段：

| 源字段 | 用途 |
|------|------|
| `f_skill_id` | 文档主键与业务标识 |
| `f_name` | 名称检索 + embedding 输入 |
| `f_description` | 描述检索 + embedding 输入 |
| `f_version` | 精确过滤 |
| `f_category` | 精确过滤 |
| `f_create_user` | 精确过滤 |
| `f_create_time` | 排序/回显 |
| `f_update_user` | 精确过滤 |
| `f_update_time` | 排序/回显 |

## 5. Embedding 设计

### 5.1 模型维度获取

接口：
- `GET http://mf-model-manage:9898/api/private/mf-model-manage/v1/small-model/list?model_name=embedding&model_type=embedding`
- Header: `x-account-id`, `x-account-type`

用途：
- 读取当前使用的 embedding 模型维度
- 初始化 Vega vector 字段 `dimension`

### 5.2 向量生成

接口：
- `POST http://mf-model-api:9898/api/private/mf-model-api/v1/small-model/embeddings`
- Header: `x-account-id`, `x-account-type`

请求体：

```json
{
  "model": "embedding",
  "input": [
    "xxx\nyyy"
  ]
}
```

设计约束：
- 输入文本统一使用 `f_name + f_description`
- 不把 `category/version/create_user/update_user` 混入 embedding 文本
- 生成结果维度必须与 schema 中 vector 字段维度一致


## 6. Vega 初始化设计

对接文档：
- [common.yaml（OpenAPI 根）](/Users/chenshu/go/src/github.com/kweaver-ai/adp/docs/api/vega/vega-backend/common.yaml)

### 6.1 Create Catalog 接口

当前需求对接方式：
- Path：`POST /api/vega-backend/v1/catalogs`
- 当前实现参考 BKN dataset 初始化写入 Vega 的流程
- `id` 必须显式传
- `connector_type` 当前不传

固定请求体：

```json
{
  "id": "kweaver_execution_factory_catalog",
  "name": "kweaver_execution_factory_catalog",
  "tags": ["execution-factory", "索引"],
  "description": "执行工厂的逻辑命名空间"
}
```

当前需求要求：
- `id`：固定为 `kweaver_execution_factory_catalog`
- `name`：固定为 `kweaver_execution_factory_catalog`
- `tags`：固定为 `["execution-factory", "索引"]`
- `description`：固定为 `执行工厂的逻辑命名空间`
- `connector_type`：当前阶段不传
- `connector_config`：当前阶段不传

### 6.2 Create Resource 接口

当前需求对接方式：
- Path：`POST /api/vega-backend/v1/resources`
- 当前实现参考 BKN dataset 初始化写入 Vega 的流程
- `id` 必须显式传
- `category` 固定为 `dataset`
- `catalog_id` 使用上一步创建或确认存在的固定 catalog ID

固定请求体：

```json
{
  "id": "kweaver_execution_factory_skill_dataset",
  "catalog_id": "kweaver_execution_factory_catalog",
  "name": "kweaver_execution_factory_skill_dataset",
  "tags": ["execution-factory", "skill", "索引"],
  "description": "执行工厂的Skill索引数据集",
  "category": "dataset",
  "status": "active",
  "source_identifier": "kweaver_execution_factory_skill_dataset",
  "schema_definition": [
    {
      "name": "skill_id",
      "type": "string",
      "display_name": "skill_id",
      "original_name": "skill_id",
      "description": "Skill 业务主键",
      "features": [
        {
          "name": "keyword_skill_id",
          "display_name": "keyword_skill_id",
          "feature_type": "keyword",
          "description": "Skill ID 精确过滤",
          "ref_property": "skill_id",
          "is_default": true,
          "is_native": false,
          "config": {
            "ignore_above": 1024
          }
        }
      ]
    },
    {
      "name": "name",
      "type": "text",
      "display_name": "name",
      "original_name": "name",
      "description": "Skill 名称",
      "features": [
        {
          "name": "keyword_name",
          "display_name": "keyword_name",
          "feature_type": "keyword",
          "description": "Skill 名称精确过滤",
          "ref_property": "name",
          "is_default": true,
          "is_native": false,
          "config": {
            "ignore_above": 1024
          }
        },
        {
          "name": "fulltext_name",
          "display_name": "fulltext_name",
          "feature_type": "fulltext",
          "description": "Skill 名称全文检索",
          "ref_property": "name",
          "is_default": true,
          "is_native": false,
          "config": {
            "analyzer": "standard"
          }
        }
      ]
    },
    {
      "name": "description",
      "type": "text",
      "display_name": "description",
      "original_name": "description",
      "description": "Skill 描述",
      "features": [
        {
          "name": "keyword_description",
          "display_name": "keyword_description",
          "feature_type": "keyword",
          "description": "Skill 描述精确过滤",
          "ref_property": "description",
          "is_default": true,
          "is_native": false,
          "config": {
            "ignore_above": 1024
          }
        },
        {
          "name": "fulltext_description",
          "display_name": "fulltext_description",
          "feature_type": "fulltext",
          "description": "Skill 描述全文检索",
          "ref_property": "description",
          "is_default": true,
          "is_native": false,
          "config": {
            "analyzer": "standard"
          }
        }
      ]
    },
    {
      "name": "version",
      "type": "string",
      "display_name": "version",
      "original_name": "version",
      "description": "Skill 当前版本",
      "features": [
        {
          "name": "keyword_version",
          "display_name": "keyword_version",
          "feature_type": "keyword",
          "description": "Skill 版本精确过滤",
          "ref_property": "version",
          "is_default": true,
          "is_native": false,
          "config": {
            "ignore_above": 1024
          }
        }
      ]
    },
    {
      "name": "category",
      "type": "string",
      "display_name": "category",
      "original_name": "category",
      "description": "Skill 分类",
      "features": [
        {
          "name": "keyword_category",
          "display_name": "keyword_category",
          "feature_type": "keyword",
          "description": "Skill 分类精确过滤",
          "ref_property": "category",
          "is_default": true,
          "is_native": false,
          "config": {
            "ignore_above": 1024
          }
        }
      ]
    },
    {
      "name": "create_user",
      "type": "string",
      "display_name": "create_user",
      "original_name": "create_user",
      "description": "Skill 创建人",
      "features": [
        {
          "name": "keyword_create_user",
          "display_name": "keyword_create_user",
          "feature_type": "keyword",
          "description": "Skill 创建人精确过滤",
          "ref_property": "create_user",
          "is_default": true,
          "is_native": false,
          "config": {
            "ignore_above": 1024
          }
        }
      ]
    },
    {
      "name": "create_time",
      "type": "datetime",
      "display_name": "create_time",
      "original_name": "create_time",
      "description": "Skill 创建时间",
      "features": []
    },
    {
      "name": "update_user",
      "type": "string",
      "display_name": "update_user",
      "original_name": "update_user",
      "description": "Skill 更新人",
      "features": [
        {
          "name": "keyword_update_user",
          "display_name": "keyword_update_user",
          "feature_type": "keyword",
          "description": "Skill 更新人精确过滤",
          "ref_property": "update_user",
          "is_default": true,
          "is_native": false,
          "config": {
            "ignore_above": 1024
          }
        }
      ]
    },
    {
      "name": "update_time",
      "type": "datetime",
      "display_name": "update_time",
      "original_name": "update_time",
      "description": "Skill 更新时间",
      "features": []
    },
    {
      "name": "_vector",
      "type": "vector",
      "display_name": "_vector",
      "original_name": "_vector",
      "description": "Skill 语义向量",
      "features": [
        {
          "name": "vector_skill",
          "display_name": "vector_skill",
          "feature_type": "vector",
          "description": "Skill 语义检索向量",
          "ref_property": "_vector",
          "is_default": true,
          "is_native": false,
          "config": {
            "dimension": 768
          }
        }
      ]
    }
  ]
}
```

当前需求要求：
- `id`：固定为 `kweaver_execution_factory_skill_dataset`
- `catalog_id`：固定为 `kweaver_execution_factory_catalog`
- `name`：固定为 `kweaver_execution_factory_skill_dataset`
- `tags`：固定为 `["execution-factory", "skill", "索引"]`
- `description`：固定为 `执行工厂的Skill索引数据集`
- `category`：固定为 `dataset`
- `status`：固定为 `active`
- `source_identifier`：固定为 `kweaver_execution_factory_skill_dataset`
- `schema_definition`：固定为完整正式字段集合，创建 resource 时一次性传入

说明：
- `schema_definition` 使用本设计文档第 7 节定义的固定字段集合
- `_vector.config.dimension` 必须替换为模型管理接口返回的真实维度

初始化步骤：

1. 查询 catalog：
- `GET /catalogs/{ids}`
- 不存在则创建 catalog

2. 查询 resource：
- `GET /resources/{ids}`
- 不存在则创建 dataset resource，并写入 `schema_definition`

设计要求：
- 初始化必须幂等
- 重复执行不能产生多个 catalog
- 重复执行不能产生多个 dataset resource

## 7. Dataset Schema 设计

### 7.1 字段映射

| Dataset 字段 | 来源字段 | 类型 | 说明 |
|------|------|------|------|
| `skill_id` | `f_skill_id` | `string` | Skill 业务主键 |
| `name` | `f_name` | `text` | Skill 名称 |
| `description` | `f_description` | `text` | Skill 描述 |
| `version` | `f_version` | `string` | 当前版本 |
| `category` | `f_category` | `string` | 分类 |
| `create_user` | `f_create_user` | `string` | 创建人 |
| `create_time` | `f_create_time` | `datetime` | 创建时间 |
| `update_user` | `f_update_user` | `string` | 更新人 |
| `update_time` | `f_update_time` | `datetime` | 更新时间 |
| `_vector` | `f_name + f_description` | `vector` | 语义向量 |

### 7.2 字段能力

- `skill_id/version/category/create_user/update_user`：`keyword`
- `name/description`：`keyword + fulltext`
- `_vector`：`vector`

### 7.3 固定文档模型

```json
{
  "_id": "skill-123",
  "skill_id": "skill-123",
  "name": "mysql_query_skill",
  "description": "执行 MySQL 查询并返回结果",
  "version": "1.0.0",
  "category": "data_query",
  "create_user": "u001",
  "create_time": 1710000000000000000,
  "update_user": "u002",
  "update_time": 1710000100000000000,
  "_vector": [0.12, -0.33, 0.08]
}
```

说明：
- `_id` 显式设置为 `skill_id`
- 时间字段复用执行工厂主表中的 `UnixNano` 语义

## 8. 生命周期同步设计

### 8.1 Skill 发布

1. 读取 `t_skill_repository` 当前 Skill 主数据
2. 基于 `name + description` 生成 embedding 输入
3. 调用 embedding 接口获取向量
4. 构建 dataset document
5. 写入 Vega dataset

### 8.2 Skill 更新

1. 重新读取最新主数据
2. 若 `name` 或 `description` 变化，则重新生成 embedding
3. 使用相同 `_id=skill_id` 覆盖写入 dataset document

### 8.3 Skill 下架

1. 根据 `skill_id` 删除 dataset 中对应 document
2. 保留 Vega catalog 与 dataset resource

## 9. 幂等与重复写入设计

Vega dataset 当前底层通过 OpenSearch bulk `index` 写入。

已确认语义：
- 不传 `_id`：底层自动生成 `_id`，重复写入会新增文档
- 传相同 `_id`：覆盖已有文档，不新增第二条，不因重复报错

设计结论：
- Skill 索引写入必须显式传 `_id=skill_id`
- 发布和更新采用同 `_id` 覆盖写
- 下架按同 `_id` 删除

扩展策略：
- 如未来要保留历史版本，可改为 `_id=skill_id:version`
- 消费侧再实现按 `skill_id` 聚合最新版本

## 10. 失败处理与补偿

失败场景：
- embedding 模型查询失败
- embedding 向量生成失败
- Vega catalog/resource 初始化失败
- dataset document 写入失败
- dataset document 删除失败

处理策略：
- 初始化失败时阻断后续文档同步
- Skill 主流程与索引同步建议解耦
- 索引同步失败进入补偿或重试队列
- 下架删除失败保留补偿机制，避免 Recall 命中脏数据

## 11. 验收清单

- 初始化层与文档同步层职责分离
- catalog/resource 初始化具备幂等性
- 同一 Skill 多次发布后，dataset 中始终只有一条 `_id=skill_id` 的文档
- Skill 更新后，结构化字段与向量字段同步更新
- Skill 下架只删除单条文档，不影响其他 Skill 数据

## 12. 失败条件

- 把 Recall 消费侧逻辑写入执行工厂
- 把 dataset resource 生命周期与 document 生命周期混淆
- 未显式设置 `_id=skill_id`
- 将 Skill 下架实现成删除整个 dataset resource
- embedding 维度与 vector schema 不一致
