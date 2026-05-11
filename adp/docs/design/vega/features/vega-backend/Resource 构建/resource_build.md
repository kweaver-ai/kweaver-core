# Resource 构建

## 1. 构建概述

Resource 的"构建"是指将 VEGA 中定义的 table 类型资源转换为文档格式，写入 OpenSearch 索引中的过程。构建过程还包括自动对需建向量的字段创建向量。由于计算向量非常耗时，构建时先写入非向量字段，后异步写入向量字段。

Resource 的构建有两种模式：批次构建和流式构建。

## 2. 批次构建（Batch Build：全量/增量）

### 2.1 原理

批次构建的原理是根据创建 Resource 构建任务时指定的 `build_key_fields`（如有自增主键或时间戳等有时序性的字段），分批次读取数据，写入索引中。每次写入完成，记录 `build_key_fields` 对应的值到 `synced_mark` 中，用于保存构建进度，下次根据 `synced_mark` 继续读取数据，以此支持全量和增量两种方式：
- **全量构建**：从头开始构建，等价于把 `synced_mark` 清空后执行，会从源数据中读取所有数据
- **增量构建**：从上次的 `synced_mark` 继续向后读取，只处理新增或变更的数据

### 2.2 具体步骤

1. 从 Resource 所属的 Catalog 获取数据源连接信息
2. 建立数据库连接
3. 根据 `execute_type` 确定构建类型，如果是全量，清空 `synced_mark`，并创建一个新索引，从头开始构建；如果是增量，继续读取上次构建的 `synced_mark`
4. 基于 `build_key_fields` 排序字段批量读取数据
5. 处理数据（跳过向量字段）并写入索引中
6. 更新 `synced_mark` 为本次处理的最后一条记录的排序字段值
7. 更新构建任务状态
8. 构建完成后，触发异步向量化任务

### 2.3 关键配置

- **构建键字段**：`build_key_fields`，用于批读游标和排序，`synced_mark` 会根据此自动生成。批量构建依赖的有时序性的字段
- **需向量化字段**：`embedding_fields`，需要生成向量的字段，多个字段用逗号分隔，无则留空字符串或不配置，索引中根据这些字段生成后缀为 `_vector` 的向量字段
- **嵌入模型**：`embedding_model`，用于生成向量的模型，默认值为 `embedding`。
- **执行类型**：`execute_type`，控制全量（`full`）或增量（`incremental`，默认）

## 3. 流式构建（Streaming Build：Kafka CDC）

### 3.1 原理

流式构建依赖 Kafka connector 和 Kafka CDC 消息（Debezium 的 `payload.before/after` 结构）。首先通过 Kafka connector 将数据源的全量和增量数据写入 Kafka 主题，然后 vega-backend 订阅此 Kafka 主题，接收数据变更事件并实时同步到索引中。支持 insert、update、delete 操作的实时处理。

Kafka Topic 命名格式：
- 数据主题：`vega-build-{resourceID}.{sourceIdentifier}` , sourceIdentifier 为创建 Resource 时指定的 sourceIdentifier，主要，MySQL 需指定为 dbname.table_name，PostgreSQL 则为 schema_name.table_name
- 向量化主题：`vega-build-{resourceID}-embedding`
- Schema 变更主题：`vega-build-schema-changes`
- Connector Topic Prefix：`vega-build-{resourceID}`

### 3.2 具体步骤

1. 从 Catalog 获取数据源连接信息
2. 配置并启动 Kafka connector：
   - 配置数据源连接参数
   - 设置初始全量同步模式
   - 配置增量变更捕获（CDC）模式
   - 指定目标 Kafka 主题：`vega-build-{resourceID}.{sourceIdentifier}`
3. Kafka connector 执行数据同步：
   - 首先执行全量同步，将数据源中的所有数据写入 Kafka
   - 然后进入增量模式，实时捕获数据源的变更（insert/update/delete）并写入 Kafka
4. vega-backend 订阅 Kafka 主题：`vega-build-{resourceID}.{sourceIdentifier}`
5. 接收并处理数据变更消息：
   - **Insert**：从 `payload.after` 构造文档并写入索引
   - **Update**：用主键查询目标文档 `_id`，再用 `payload.after` 更新
   - **Delete**：用主键构造过滤条件删除匹配文档
6. 实时同步数据到索引（向量字段写为占位值）
7. 监控任务状态，支持停止操作
8. 启动时触发异步向量化任务，保证持续补向量

### 3.3 关键配置

- **唯一键配置**：`build_key_fields`，用于流式构建中唯一标识某行的字段，确保修改和删除操作的准确执行
- **需向量化字段**：`embedding_fields`，需要生成向量的字段，多个字段用逗号分隔，无则留空字符串或不配置，索引中根据这些字段生成后缀为 `_vector` 的向量字段
- **嵌入模型**：`embedding_model`，用于生成向量的模型，默认值为 `embedding`。

### 3.4 数据库配置要求

- **MySQL 配置要求**：开启 binlog ，确保 binlog_format = ROW，可登录数据库使用 `SHOW VARIABLES LIKE 'binlog_format';` 查看当前配置。
- **PostgreSql 配置要求**：开启 wal 日志，确保 wal_level = logical，可登录数据库使用 `SHOW wal_level;` 查看当前配置。

## 4. 两段式构建

Resource 采用两段式构建模式：先同步结构化字段，再异步补齐向量（embedding）。这是批次构建和流式构建都包含的过程：

### 4.1 第1段：结构化字段同步

- **批次构建**：基于 `build_key_fields` 分批次读取数据，写入非向量字段
- **流式构建**：通过 Kafka CDC 接收变更事件，实时写入非向量字段（向量字段写为占位值）

### 4.2 第2段：异步向量化

#### 4.2.1 原理

异步向量化任务负责为索引中未生成向量的文档补齐向量字段。通过订阅 `vega-build-{resourceID}-embedding` Kafka 主题，读取源文本字段，生成向量后回写到索引中。

#### 4.2.2 具体步骤

1. 从 `resource.schema_definition` 找到所有 `vector` 字段，并记录每个 vector 对应的源文本字段
2. 以批次方式查询"未向量化"的文档（使用 `operation: "null"` 条件）
3. 对每条文档生成向量（基于源文本字段）
4. 构造更新请求，将生成的向量回写到索引
5. 更新任务状态中的 `vectorized_count`
6. 重复执行直到达到停止条件

#### 4.2.3 触发时机

- **批次构建完成后**：入队 `embedding:execute` 任务
- **流式构建启动时**：入队 `embedding:execute` 任务，保证持续补向量

## 5. 构建方式对比

| 构建方式 | 适用场景 | 优点 | 缺点 |
|---------|---------|------|------|
| **批次构建** | 静态或近静态数据；初始数据导入；周期性同步 | 实现简单；吞吐高 ；资源消耗低 | 实时性差；依赖排序字段；无法处理 update/delete 操作 |
| **流式构建** | 实时数据同步；频繁变更的数据源 | 实时性强；天然支持 update/delete | 依赖 Kafka CDC；运维复杂；需要稳定主键/唯一键 |

### 5.1 批次构建：全量 vs 增量

| 执行类型 | 适用场景 |
|---------|---------|
| **全量构建** | 适合历史数据有变更场景 |
| **增量构建** | 适合历史数据无变更，只有新增数据场景 |

## 6. 构建管理 API

| API 路径 | 方法 | 描述 |
|---------|------|------|
| `/api/vega-backend/v1/resources/dataset/build` | GET | 获取构建任务列表 |
| `/api/vega-backend/v1/resources/buildtask/{resourceId}` | POST | 创建构建任务（仅创建，状态为 pending） |
| `/api/vega-backend/v1/resources/buildtask/{resourceId}/{taskId}` | GET | 获取构建任务详情 |
| `/api/vega-backend/v1/resources/buildtask/{resourceId}/{taskId}/status` | PUT | 启动/停止构建任务（running/stopped），并可指定 batch 的执行类型（full/incremental） |
| `/api/vega-backend/v1/resources/buildtask/{taskId}` | DELETE | 删除构建任务 |

## 7. 代码示例

### 7.1 创建批次/流式构建任务（仅创建）

```json
POST /api/vega-backend/v1/resources/buildtask/{resourceId}
{
  "mode": "batch",
  "embedding_fields": "description,content",
  "embedding_model": "embedding",
  "build_key_fields": "id,name"
}
```

```json
POST /api/vega-backend/v1/resources/buildtask/{resourceId}
{
  "mode": "streaming",
  "embedding_fields": "description,content",
  "embedding_model": "embedding",
  "build_key_fields": "id"
}
```

### 7.2 启动/停止构建任务（真正执行）

创建任务后，需要显式启动（默认增量）：

```json
PUT /api/vega-backend/v1/resources/buildtask/{resourceId}/{taskId}/status
{
  "status": "running",
  "execute_type": "incremental"
}
```

批次全量执行：

```json
PUT /api/vega-backend/v1/resources/buildtask/{resourceId}/{taskId}/status
{
  "status": "running",
  "execute_type": "full"
}
```

停止执行：

```json
PUT /api/vega-backend/v1/resources/buildtask/{resourceId}/{taskId}/status
{
  "status": "stopped"
}
```

## 8. 最佳实践

1. **初始数据导入**：使用批次构建进行初始数据导入
2. **增量水位选择**：为批次构建设置稳定的 `build_key_fields`，避免"排序字段重复/不可比较"导致漏数或重复
3. **实时数据同步**：需要准实时更新时使用流式构建，并设置正确的唯一键字段
4. **向量字段策略**：
   - 批次构建不要写入占位向量（如全 0），交给 embedding 任务补齐
   - 明确"未向量化"的判定方式（字段缺失 vs 显式 null），保证能被 embedding 条件筛到
5. **Schema 设计**：全文字段使用 `text`，精确过滤字段使用 `keyword`/数值类型，向量字段使用 `vector` 并配置维度等参数
6. **可观测性**：关注 `synced_count / total_count / vectorized_count` 三个计数，便于判断"同步完成但向量未完成"等状态

## 9. 常见问题

### 9.1 构建失败

- **原因**：数据源连接失败、权限不足、网络问题等
- **解决方法**：检查 Catalog 配置、数据源状态和网络连接

### 9.2 数据同步延迟

- **原因**：Kafka 消息积压、处理速度慢等
- **解决方法**：优化处理逻辑、增加消费者数量、检查系统资源

### 9.3 主键冲突 / 更新找不到文档

- **原因**：主键设置不当或数据重复
- **解决方法**：确保主键的唯一性，检查数据源中的重复数据

### 9.4 topic 冲突

- **原因**：kafka connect 会创建 `vega-build-{resourceID}.{sourceIdentifier}` 的 topic，如果存在不同数据库实例下有同名的库和表，会导致 topic 冲突
- **解决方法**：每个构建任务的 resourceID 都是唯一的，topic 不会冲突