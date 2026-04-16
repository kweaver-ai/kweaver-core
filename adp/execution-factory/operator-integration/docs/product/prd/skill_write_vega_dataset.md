#  Skill数据写入vega dataset设计文档

## 1. 背景

- 上游需求参考：
  - 线上文档：`https://github.com/criller/kweaver/blob/feature/109-contextloader-skill-recall/adp/context-loader/agent-retrieval/docs/prd/issue-109-contextloader-skill-recall-prd.md`
- 端到端链路目标：Skill 不应只是脱离业务对象的通用能力包，而应能够与 BKN 中的知识网络、对象类、对象实例建立语义关联，并在 Agent 处理具体业务对象时被自动召回。
- 模块分工：`context-loader` 负责运行时召回，`bkn` 负责 Skill 与业务对象的绑定语义，`execution-factory` 负责 Skill 的权威元数据、渐进式读取与执行接入。
- 当前需求是上述 Skill 能力的上游支撑项，职责不在于消费召回结果，而在于为召回链路生产稳定、可索引、可语义检索的 Skill 数据。
- 执行工厂当前已经具备 Skill 的注册、查询、详情、内容读取、文件读取、下载和删除能力。
- 现阶段 Skill 主要依赖关系型存储承载主数据，缺少面向语义检索和统一资源检索的专用索引。
- 为支撑上游 Skill 召回场景、Skill 市场检索和后续统一检索能力，需要将 Skill 主数据同步到 Vega Dataset，形成执行工厂侧的索引生产能力。

## 2. 目标

- 建立执行工厂 Skill 到 Vega Dataset 的独立同步链路。
- 以 `t_skill_repository` 中的 Skill 主数据为数据源，构建可被 Vega 检索的数据文档。
- 使用 `f_name` 与 `f_description` 生成 embedding 向量，支持 Skill 语义检索。
- 保证 Skill 发布、更新、下架时，Vega Dataset 中的数据与主表状态一致。
- 为上游 `context-loader` 的 Skill召回提供稳定的索引数据基础，而不是在执行工厂内实现召回编排逻辑。

## 3. 非目标

- 不在本需求中定义 Skill 市场前端检索页面改造。
- 不在本需求中引入版本级 Skill 历史检索能力。
- 不在本需求中修改既有 Skill 管理接口语义。
- 不在本需求中删除 Vega catalog 或 dataset resource 本身。

## 4. 业务对象与范围

数据源表：
- `t_skill_repository`

本次需要同步的字段：
- `f_skill_id`
- `f_name`
- `f_description`
- `f_version`
- `f_category`
- `f_create_user`
- `f_create_time`
- `f_update_user`
- `f_update_time`

向量源字段：
- `f_name`
- `f_description`

## 5. 总体流程

Skill 对接 Vega Dataset 的流程分为两层：

1. 初始化层
- 负责校验或创建 Vega `catalog` 和 `dataset resource`
- 只处理资源存在性与 schema 初始化

2. 文档同步层
- 负责将单个 Skill 映射为一条 dataset document
- 处理发布、更新、下架时的增删改

职责边界：
- 初始化层不负责写入单个 Skill 文档
- 文档同步层不负责反复创建 catalog/resource

## 6. Embedding 获取需求

### 6.1 获取 embedding 向量维度

接口要求：
- Method：`GET`
- URL：`http://mf-model-manage:9898/api/private/mf-model-manage/v1/small-model/list?model_name=embedding&model_type=embedding`
- Header：
  - `x-account-id`
  - `x-account-type`

需求说明：
- 需要从模型管理服务中获取当前使用的 embedding 模型信息。
- 需要读取 embedding 模型的向量维度，用于生成 Vega dataset schema 中 vector 字段的 `dimension`。
- 若接口返回多个模型，需要明确选择当前生效模型；本需求默认使用查询结果中的目标 embedding 模型。

### 6.2 获取 embedding 向量

接口要求：
- Method：`POST`
- URL：`http://mf-model-api:9898/api/private/mf-model-api/v1/small-model/embeddings`
- Header：
  - `x-account-id`
  - `x-account-type`
- Body：

```json
{
  "model": "embedding",
  "input": []
}
```

需求说明：
- `input` 为待向量化的文本数组。
- 单个 Skill 的 embedding 输入由 `f_name` 和 `f_description` 拼接生成。
- 推荐输入格式：

```text
技能名称：{f_name}
技能描述：{f_description}
```

约束：
- 不将 `category`、`version`、`create_user`、`update_user` 等结构化过滤字段混入 embedding 文本。
- 向量维度必须与 dataset schema 中 vector 字段的 `dimension` 保持一致。

## 7. Vega 初始化需求

对接接口文档：
- [common.yaml（OpenAPI 根）](/Users/chenshu/go/src/github.com/kweaver-ai/adp/docs/api/vega/vega-backend/common.yaml)

初始化流程：

1. 根据 `catalog_id` 查询 catalog 是否存在：`/catalogs/{ids}`
- 存在：继续下一步
- 不存在：创建 catalog

2. 根据 `resource_id` 查询 resource 是否存在
- 存在：初始化完成
- 不存在：创建 dataset resource

3. 初始化完成后，Skill 文档同步链路才允许执行

关键要求：
- catalog 与 resource 初始化必须具备幂等性
- 重复初始化不能重复创建 catalog
- 重复初始化不能重复创建 resource

职责边界：
- 本流程创建的是共享的 Skill dataset resource
- 不是为每个 Skill 创建独立 resource

## 8. Dataset 数据模型需求

### 8.1 文档模型

每个 Skill 在 dataset 中对应一条 document。

建议字段：

| 文档字段 | 来源 | 说明 |
|------|------|------|
| `_id` | `f_skill_id` | 文档唯一标识，用于幂等写入 |
| `skill_id` | `f_skill_id` | Skill ID |
| `name` | `f_name` | Skill 名称 |
| `description` | `f_description` | Skill 描述 |
| `version` | `f_version` | 当前版本 |
| `category` | `f_category` | 分类 |
| `create_user` | `f_create_user` | 创建人 |
| `create_time` | `f_create_time` | 创建时间 |
| `update_user` | `f_update_user` | 更新人 |
| `update_time` | `f_update_time` | 更新时间 |
| `_vector` | `f_name + f_description` | 向量字段 |

示例：

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

### 8.2 Schema 能力要求

结构化字段要求：
- `skill_id` 支持精确匹配
- `version` 支持精确过滤
- `category` 支持精确过滤
- `create_user` 支持精确过滤
- `update_user` 支持精确过滤
- `create_time` 支持排序
- `update_time` 支持排序

全文字段要求：
- `name` 支持精确过滤与全文检索
- `description` 支持精确过滤与全文检索

向量字段要求：
- `_vector` 支持向量检索
- `_vector.dimension` 必须与 embedding 模型维度一致

## 9. 生命周期同步需求

### 9.1 Skill 发布

当 Skill 发布时，需要：

1. 从 `t_skill_repository` 读取当前 Skill 主数据
2. 使用 `f_name` 和 `f_description` 生成 embedding 输入
3. 调用 embedding 接口获取向量
4. 构建 dataset document
5. 将 document 写入 Vega dataset

结果要求：
- 发布后，Skill 可在 Vega dataset 中被查询和召回
- 同一 Skill 在 dataset 中只保留一条当前有效文档

### 9.2 Skill 更新

当 Skill 更新时，需要：

1. 重新读取最新主数据
2. 若 `name` 或 `description` 发生变化，则重新生成 embedding
3. 更新 dataset 中对应的 document

结果要求：
- 更新后的名称、描述、版本、分类、更新时间等字段在 dataset 中同步生效
- 更新后的语义向量与最新 `name + description` 保持一致

### 9.3 Skill 下架

当 Skill 下架时，需要：

1. 删除 dataset 中对应 Skill 的 document

结果要求：
- 下架后，Skill 不再从 dataset 中可查

明确约束：
- 下架删除的是单条 Skill 文档
- 不是删除整个 dataset resource

## 10. 重复写入与幂等需求

### 10.1 问题定义

需要明确同一个 Skill 重复写入 Vega dataset 时的表现：
- 是新增一条
- 是覆盖原数据
- 还是直接报错

### 10.2 已确认行为

基于 Vega 当前实现，dataset 文档写入底层使用 OpenSearch bulk `index`：

1. 未传 `_id`
- Vega/OpenSearch 会为文档自动生成 `_id`
- 重复写入同一 Skill 会新增一条新文档
- 不会覆盖旧文档
- 通常不会因为“业务重复”直接报错

2. 传入相同 `_id`
- 使用相同 `_id` 再次写入时，会覆盖已有文档
- 不会新增第二条同 `_id` 文档
- 不会因为重复 `_id` 直接报错

### 10.3 本需求要求

- Skill 写入 dataset 时必须显式传 `_id`
- `_id` 固定使用 `skill_id`
- Skill 发布和更新均按相同 `_id` 覆盖写
- Skill 下架按相同 `_id` 删除

这样可以保证：
- 同一 Skill 不会因重复同步产生多条重复文档
- 发布和更新链路天然具备幂等性

## 11. 错误处理与补偿需求

失败场景包括：
- embedding 模型列表查询失败
- embedding 向量生成失败
- Vega catalog 查询或创建失败
- Vega resource 查询或创建失败
- dataset 文档写入失败
- dataset 文档删除失败

处理要求：
- 初始化失败时，阻断文档同步
- 主业务成功但索引同步失败时，需要保留补偿能力
- 下架删除失败时，需要保留重试或补偿机制，避免脏数据长期残留

## 12. 验收标准

1. 初始化链路
- 首次执行时可成功创建 catalog 和 dataset resource
- 重复执行初始化不会重复创建 catalog 和 resource

2. 发布链路
- Skill 发布后可在 Vega dataset 中查到对应文档
- 文档包含结构化字段和 `_vector`

3. 更新链路
- Skill 更新后，dataset 中对应文档被覆盖更新
- `name`、`description` 变化后，向量同步更新

4. 下架链路
- Skill 下架后，dataset 中对应文档被删除
- 其他 Skill 文档不受影响

5. 幂等链路
- 同一 Skill 重复发布或重复同步不会出现多条重复文档

## 13. 失败条件

- 未显式设置 `_id`，导致同一 Skill 在 dataset 中出现多条重复文档
- 将 Skill 下架实现成删除整个 dataset resource
- embedding 维度与 dataset schema 中 vector 维度不一致
- Skill 主表已更新，但 dataset 文档未更新且无补偿机制
- catalog 或 resource 初始化没有幂等控制，导致重复创建
