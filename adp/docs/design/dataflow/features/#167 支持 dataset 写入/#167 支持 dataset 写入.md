# Dataset 写入节点设计

## 1. 背景与目标

### 1.1 问题背景

当前非结构化数据处理链路中，Dataflow 将解析结果写入 OpenSearch，再经过数据连接扫描、原子视图生成、BKN 建模等步骤完成知识网络构建。这条链路较长，OpenSearch 主要承担"中间存储层"角色。

为降低链路复杂度，引入 dataset 作为 Dataflow 的直接写入目标，使链路收敛为：

```
Dataflow → Dataset → BKN
```

### 1.2 设计目标

新增 Dataset 写入节点，支持：

- 接收上游任意结构化输出
- 自动补充运行时元字段
- 批量调用 dataset docs 写入接口
- 返回执行结果统计

### 1.3 非目标

节点**不负责**：

- 创建 dataset / 更新 schema
- 创建 BKN / 建立映射关系
- 触发 BKN 构建任务

这些能力由平台资源管理侧承担，kweaver cli 提供快捷创建命令。

---

## 2. 职责边界

### 2.1 节点职责

Dataset 写入节点职责：

| 职责 | 说明 |
|------|------|
| 接收输入数据 | 兼容多种输入格式 |
| 数据规范化 | 统一为 dataset docs 写入格式，补充元字段 |
| 批量写入 | 调用 dataset docs 接口，汇总结果 |

### 2.2 平台职责

| 职责 | 说明 |
|------|------|
| 创建 dataset | 平台预先创建目标 dataset |
| 创建 BKN | 平台预先创建知识网络 |
| 维护映射关系 | dataset 与 BKN 的映射由平台维护 |

### 2.3 整体数据流

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              整体数据流                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────┐                                                       │
│   │ 平台资源管理侧   │                                                       │
│   │                 │                                                       │
│   │  ① 创建 Dataset │────────────────────────┐                              │
│   │  ② 创建 BKN     │                        │                              │
│   │  ③ 建立映射     │                        ▼                              │
│   └─────────────────┘              ┌──────────────────┐                     │
│                                    │    Dataset       │                     │
│                                    │  (预先创建)       │                     │
│                                    └────────▲─────────┘                     │
│                                             │                               │
│   ┌─────────────────┐                       │                               │
│   │   Dataflow      │                       │                               │
│   │                 │                       │                               │
│   │  ┌───────────┐  │    ┌──────────────────┴──────────────────┐           │
│   │  │ 文档解析   │  │    │         DatasetWriteDocs           │           │
│   │  │ 节点      │──┼───►│              (本节点)                │           │
│   │  └───────────┘  │    │                                    │           │
│   │  ┌───────────┐  │    │  • 归一化输入数据                    │           │
│   │  │ 文本分割   │  │    │  • 补充运行时元字段                  │           │
│   │  │ 节点      │──┼───►│  • 批量写入                         │           │
│   │  └───────────┘  │    │  • 返回执行结果                      │           │
│   │                 │    └──────────────────┬──────────────────┘           │
│   └─────────────────┘                       │                               │
│                                             ▼                               │
│                                    ┌──────────────────┐                     │
│                                    │    执行结果      │                     │
│                                    │  total/success   │                     │
│                                    │  /failed/reasons │                     │
│                                    └──────────────────┘                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. 节点设计

### 3.1 节点名称

**`DatasetWriteDocs`**

语义上直接对应 dataset docs 写入接口。

### 3.2 参数定义

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `dataset_id` | string | 是 | - | 目标 dataset 资源 ID |
| `documents` | any | 是 | - | 待写入数据，支持多种格式 |
| `record_type` | string | 否 | - | 记录业务类型，如 `chunk`、`element` |
| `auto_timestamp` | bool | 否 | true | 自动补充 `@timestamp` 字段 |
| `batch_size` | int | 否 | 1000 | 分批写入大小 |

### 3.3 输入格式支持

`documents` 字段支持以下格式：

| 格式 | 示例 |
|------|------|
| JSON object | `{"id": "1", "text": "..."}` |
| JSON array | `[{"id": "1"}, {"id": "2"}]` |
| JSON string | `"{\"id\": \"1\"}"` |
| JSON string array | `["{\"id\": \"1\"}", "{\"id\": \"2\"}"]` |

### 3.4 输出格式

```json
{
  "total": 100,
  "success": 90,
  "failed": 10,
  "reasons": ["[0-50] dataset write timeout", "[50-100] schema mismatch"]
}
```

### 3.5 配置示例

**写入 chunk 数据：**

```json
{
  "dataset_id": "ds_resume_parse_001",
  "record_type": "chunk",
  "auto_timestamp": true,
  "batch_size": 500,
  "documents": "{{__3.documents}}"
}
```

**写入 element 数据：**

```json
{
  "dataset_id": "ds_resume_parse_001",
  "record_type": "element",
  "documents": "{{__4.elements}}"
}
```

---

## 4. 服务依赖

### 4.1 依赖接口

节点依赖 VegaBackend 的 dataset 文档写入接口：

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 批量写文档 | POST | `/api/vega-backend/in/v1/resources/dataset/{id}/docs` | 核心依赖 |

### 4.2 客户端复用

复用 `execution-factory` 中已有的 `VegaBackendClient`：

```go
// 已有方法，可直接调用
WriteDatasetDocuments(ctx context.Context, datasetID string, documents []map[string]any) error
```

### 4.3 配置依赖

节点通过环境变量获取 VegaBackend 服务地址：

| 配置项 | 说明 |
|--------|------|
| `VEGA_BACKEND_PRIVATE_HOST` | 内网地址 |
| `VEGA_BACKEND_PRIVATE_PORT` | 内网端口 |
| `VEGA_BACKEND_PRIVATE_PROTOCOL` | 协议 (http/https) |

---

## 5. 前置准备

### 5.1 Dataset 创建流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Dataset 创建流程                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   步骤1: 创建 Catalog (若不存在)                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ POST /api/vega-backend/v1/catalogs                                  │   │
│   │ {                                                                    │   │
│   │   "name": "dataset-catalog",                                        │   │
│   │   "description": "Dataset存储目录",                                  │   │
│   │   "connector_type": "mariadb",                                      │   │
│   │   "connector_config": { ... }                                       │   │
│   │ }                                                                    │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│   步骤2: 创建 Dataset Resource                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ POST /api/vega-backend/v1/resources                                 │   │
│   │ {                                                                    │   │
│   │   "catalog_id": "{catalog_id}",                                     │   │
│   │   "name": "ds_resume_parse",                                        │   │
│   │   "category": "dataset",                                            │   │
│   │   "status": "active",                                               │   │
│   │   "schema_definition": [ ... ]                                      │   │
│   │ }                                                                    │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│   步骤3: 创建 BKN (平台管理)                                                 │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 通过 BKN 管理接口创建知识网络                                          │   │
│   │ 建立 Dataset 与 BKN 的映射关系                                        │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│   步骤4: Dataflow 配置 DatasetWriteDocs 节点                                │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ 在 Dataflow 流程中配置节点，使用步骤2创建的 dataset_id                 │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Dataset 创建参数示例

#### 5.2.1 Chunk 数据 Dataset

用于存储文档分块数据：

```json
{
  "catalog_id": "default",
  "name": "ds_doc_chunks",
  "tags": ["dataflow", "chunk"],
  "description": "文档分块数据集",
  "category": "dataset",
  "status": "active",
  "source_identifier": "dataflow_chunks",
  "schema_definition": [
    {"name": "id", "type": "keyword", "display_name": "ID", "description": "分块唯一标识"},
    {"name": "document_id", "type": "keyword", "display_name": "文档ID", "description": "来源文档标识"},
    {"name": "@timestamp", "type": "long", "display_name": "时间戳", "description": "事件时间"},
    {"name": "__write_time", "type": "long", "display_name": "写入时间", "description": "写入时间戳"},
    {"name": "__record_type", "type": "keyword", "display_name": "记录类型", "description": "chunk/element"},
    {"name": "__task_id", "type": "keyword", "display_name": "任务ID", "description": "Dataflow任务标识"},
    {"name": "chunk_text", "type": "text", "display_name": "分块文本", "description": "分块内容"},
    {"name": "chunk_index", "type": "integer", "display_name": "分块索引", "description": "分块序号"},
    {"name": "doc_name", "type": "text", "display_name": "文档名称", "description": "来源文档名称"},
    {"name": "content_vector", "type": "vector", "display_name": "内容向量", "description": "文本向量", "features": [
      {
        "name": "content_vector",
        "feature_type": "vector",
        "ref_property": "content_vector",
        "is_default": true,
        "is_native": true,
        "config": {
          "dimension": 768,
          "method": {"name": "hnsw", "engine": "lucene", "parameters": {"ef_construction": 256}}
        }
      }
    ]}
  ]
}
```

#### 5.2.2 Element 数据 Dataset

用于存储文档解析后的结构化元素：

```json
{
  "catalog_id": "default",
  "name": "ds_doc_elements",
  "tags": ["dataflow", "element"],
  "description": "文档解析元素数据集",
  "category": "dataset",
  "status": "active",
  "source_identifier": "dataflow_elements",
  "schema_definition": [
    {"name": "id", "type": "keyword", "display_name": "ID", "description": "元素唯一标识"},
    {"name": "document_id", "type": "keyword", "display_name": "文档ID", "description": "来源文档标识"},
    {"name": "@timestamp", "type": "long", "display_name": "时间戳", "description": "事件时间"},
    {"name": "__write_time", "type": "long", "display_name": "写入时间", "description": "写入时间戳"},
    {"name": "__record_type", "type": "keyword", "display_name": "记录类型", "description": "chunk/element"},
    {"name": "__task_id", "type": "keyword", "display_name": "任务ID", "description": "Dataflow任务标识"},
    {"name": "element_type", "type": "keyword", "display_name": "元素类型", "description": "table/image/text等"},
    {"name": "element_content", "type": "text", "display_name": "元素内容", "description": "元素文本内容"},
    {"name": "page_number", "type": "integer", "display_name": "页码", "description": "所在页码"},
    {"name": "bbox", "type": "json", "display_name": "边界框", "description": "元素位置信息"}
  ]
}
```

#### 5.2.3 完整文档解析 Dataset

用于存储完整的文档解析结果：

```json
{
  "catalog_id": "default",
  "name": "ds_resume_parse_001",
  "tags": ["dataflow", "resume"],
  "description": "简历解析结果数据集",
  "category": "dataset",
  "status": "active",
  "source_identifier": "dataflow_resume",
  "schema_definition": [
    {"name": "id", "type": "keyword", "display_name": "ID", "description": "记录唯一标识"},
    {"name": "document_id", "type": "keyword", "display_name": "文档ID", "description": "来源文档标识"},
    {"name": "@timestamp", "type": "long", "display_name": "时间戳", "description": "事件时间"},
    {"name": "__write_time", "type": "long", "display_name": "写入时间", "description": "写入时间戳"},
    {"name": "__record_type", "type": "keyword", "display_name": "记录类型", "description": "记录业务类型"},
    {"name": "__task_id", "type": "keyword", "display_name": "任务ID", "description": "Dataflow任务标识"},
    {"name": "doc_name", "type": "text", "display_name": "文档名称", "description": "来源文档名称", "features": [
      {"name": "keyword_name", "feature_type": "keyword", "ref_property": "doc_name", "is_default": true, "is_native": true}
    ]},
    {"name": "chunk_text", "type": "text", "display_name": "分块文本", "description": "分块内容"},
    {"name": "element_type", "type": "keyword", "display_name": "元素类型", "description": "元素类型"},
    {"name": "content_vector", "type": "vector", "display_name": "内容向量", "description": "文本向量", "features": [
      {
        "name": "content_vector",
        "feature_type": "vector",
        "ref_property": "content_vector",
        "is_default": true,
        "is_native": true,
        "config": {
          "dimension": 768,
          "method": {"name": "hnsw", "engine": "lucene", "parameters": {"ef_construction": 256}}
        }
      }
    ]}
  ]
}
```

### 5.3 数据契约

Dataset 数据字段定义：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | keyword | 文档唯一标识 |
| `document_id` | keyword | 来源文档 ID |
| `@timestamp` | long | 时间戳 |
| `__write_time` | long | 写入时间 |
| `__record_type` | keyword | 记录类型 (运行时补充) |
| `__task_id` | keyword | 任务 ID (运行时补充) |

---

## 6. 执行流程

### 6.1 核心执行流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           核心执行流程                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐                                                        │
│  │ 1. 参数校验      │                                                        │
│  └────────┬────────┘                                                        │
│           │ dataset_id 必填                                                 │
│           ▼                                                                 │
│  ┌─────────────────┐                                                        │
│  │ 2. 设置默认值    │                                                        │
│  └────────┬────────┘                                                        │
│           │ batch_size=1000, auto_timestamp=true                            │
│           ▼                                                                 │
│  ┌─────────────────┐                                                        │
│  │ 3. 输入归一化    │                                                        │
│  └────────┬────────┘                                                        │
│           │ • 将 documents 转换为 []map[string]any                           │
│           │ • 补充运行时元字段                                                │
│           ▼                                                                 │
│  ┌─────────────────┐     ┌─────────────────┐                                │
│  │ 4. 空输入检查    │────►│ 返回空结果       │                                │
│  └────────┬────────┘     └─────────────────┘                                │
│           │ len(documents) > 0                                              │
│           ▼                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ 5. 分批写入循环                                                      │    │
│  │ ┌─────────────────────────────────────────────────────────────────┐ │    │
│  │ │ for i := 0; i < len(documents); i += batchSize {                │ │    │
│  │ │   batch := documents[i:min(i+batchSize, len(documents))]        │ │    │
│  │ │                                                                 │ │    │
│  │ │   err := WriteDatasetDocuments(ctx, datasetID, batch)           │ │    │
│  │ │                                                                 │ │    │
│  │ │   if err != nil {                                               │ │    │
│  │ │     failed += len(batch)                                        │ │    │
│  │ │     reasons = append(reasons, error)                            │ │    │
│  │ │   } else {                                                      │ │    │
│  │ │     success += len(batch)                                       │ │    │
│  │ │   }                                                             │ │    │
│  │ │ }                                                               │ │    │
│  │ └─────────────────────────────────────────────────────────────────┘ │    │
│  └────────────────────────────────────────┬────────────────────────────┘    │
│                                           │                                 │
│           ┌───────────────────────────────┘                                 │
│           ▼                                                                 │
│  ┌─────────────────┐                                                        │
│  │ 6. 结果汇总      │                                                        │
│  └────────┬────────┘                                                        │
│           │ total = len(documents)                                          │
│           ▼                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ 返回结果                                                              │    │
│  │ {                                                                     │    │
│  │   "total": N,                                                         │    │
│  │   "success": success,                                                 │    │
│  │   "failed": failed,                                                   │    │
│  │   "reasons": reasons (若有失败)                                        │    │
│  │ }                                                                     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.2 归一化流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           输入归一化流程                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  输入: documents (any)                                                      │
│                                                                             │
│           ┌─────────────────────────────────────────────────────────┐       │
│           │                     类型判断                              │       │
│           └─────────────────────────────────────────────────────────┘       │
│                          │                                                  │
│          ┌───────────────┼───────────────┬───────────────┐                  │
│          ▼               ▼               ▼               ▼                  │
│   ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐           │
│   │ string     │  │ map[string]│  │ []any      │  │ 其他       │           │
│   │            │  │ any        │  │            │  │            │           │
│   └─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘           │
│         │               │               │               │                  │
│         ▼               │               │               ▼                  │
│   ┌────────────┐        │               │         ┌────────────┐           │
│   │ JSON解析   │        │               │         │ 返回 error │           │
│   │ 递归处理   │        │               │         └────────────┘           │
│   └─────┬──────┘        │               │                                  │
│         │               │               │                                  │
│         └───────────────┴───────────────┘                                  │
│                         │                                                  │
│                         ▼                                                  │
│           ┌─────────────────────────────────────────────────────────┐       │
│           │ 对每条 document 执行：                                    │       │
│           │                                                         │       │
│           │ 1. 若缺失 @timestamp 且 auto_timestamp=true:             │       │
│           │    doc["@timestamp"] = time.Now().UnixMilli()           │       │
│           │                                                         │       │
│           │ 2. 补充 __write_time:                                   │       │
│           │    doc["__write_time"] = time.Now().UnixMilli()         │       │
│           │                                                         │       │
│           │ 3. 若配置了 record_type:                                 │       │
│           │    doc["__record_type"] = recordType                    │       │
│           │                                                         │       │
│           │ 4. 补充 __task_id:                                      │       │
│           │    doc["__task_id"] = ctx.GetTaskID()                   │       │
│           └─────────────────────────────────────────────────────────┘       │
│                         │                                                  │
│                         ▼                                                  │
│           输出: []map[string]any                                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.3 自动补充字段

| 字段 | 条件 | 说明 |
|------|------|------|
| `@timestamp` | 缺失且 `auto_timestamp=true` | 当前毫秒时间戳 |
| `__write_time` | 始终 | 写入时间戳 |
| `__record_type` | 配置了 `record_type` 参数 | 记录类型 |
| `__task_id` | 始终 | 当前任务 ID |

### 6.4 与 OpenSearch 节点对比

| 对比项 | OpenSearchBulkUpsert | DatasetWriteDocs |
|--------|---------------------|------------------|
| 目标资源 | index | dataset resource |
| 创建能力 | 内置 settings/mappings | 不创建 schema |
| 写入接口 | bulk upsert | docs create |
| 批量大小 | 1000 | 1000 (配置) |
| 输出格式 | total/success/failed/reasons | 相同 |
| 元字段 | `__base_type`, `__data_type`, `__catetory` | `__record_type`, `__task_id` |

---

## 7. 异常处理

### 7.1 参数错误

| 场景 | 处理方式 |
|------|----------|
| `dataset_id` 为空 | 直接返回 error |
| `documents` 类型不支持 | 直接返回 error |
| JSON 字符串无法解析 | 直接返回 error |

### 7.2 写入错误

| 场景 | 处理方式 |
|------|----------|
| 单批写入失败 | 当前批记为失败，记录 reason，继续后续批次 |
| dataset 服务不可用 | 批次失败，继续尝试后续批次 |
| dataset 不存在 | 批次失败，继续尝试后续批次 |
| schema 不匹配 | 批次失败，继续尝试后续批次 |

### 7.3 空输入

上游无输出或输出为空数组时，直接返回：

```json
{
  "total": 0,
  "success": 0,
  "failed": 0
}
```

---


## 附录 A：代码参考

### A.1 节点结构体

```go
type DatasetWriteDocs struct {
    DatasetID     string `json:"dataset_id"`
    Documents     any    `json:"documents"`
    RecordType    string `json:"record_type,omitempty"`
    AutoTimestamp bool   `json:"auto_timestamp,omitempty"`
    BatchSize     int    `json:"batch_size,omitempty"`
}
```

### A.2 归一化函数签名

```go
func normalizeDatasetDocuments(
    documents any,
    recordType string,
    autoTimestamp bool,
    ctx entity.ExecuteContext,
) []map[string]any
```

### A.3 执行伪代码

```go
func (b *DatasetWriteDocs) Run(
    ctx entity.ExecuteContext,
    params interface{},
    token *entity.Token,
) (interface{}, error) {
    // 1. 参数校验
    input := params.(*DatasetWriteDocs)
    if input.DatasetID == "" {
        return nil, fmt.Errorf("dataset_id is required")
    }

    // 2. 设置默认值
    batchSize := input.BatchSize
    if batchSize <= 0 {
        batchSize = 1000
    }

    // 3. 归一化文档
    documents := normalizeDatasetDocuments(
        input.Documents,
        input.RecordType,
        input.AutoTimestamp,
        ctx,
    )

    // 4. 空输入处理
    if len(documents) == 0 {
        return map[string]any{"total": 0, "success": 0, "failed": 0}, nil
    }

    // 5. 批量写入
    result := map[string]any{}
    success, failed := 0, 0
    reasons := []string{}

    for i := 0; i < len(documents); i += batchSize {
        end := min(i+batchSize, len(documents))
        batch := documents[i:end]

        err := datasetClient.WriteDatasetDocuments(ctx.Context(), input.DatasetID, batch)
        if err != nil {
            reasons = append(reasons, fmt.Sprintf("[%d-%d] %s", i, end, err.Error()))
            failed += len(batch)
        } else {
            success += len(batch)
        }
    }

    // 6. 返回结果
    result["total"] = len(documents)
    result["success"] = success
    result["failed"] = failed
    if len(reasons) > 0 {
        result["reasons"] = reasons
    }

    return result, nil
}
```

### A.4 归一化实现参考

```go
func normalizeDocuments(documents any, recordType string, autoTimestamp bool, ctx entity.ExecuteContext) []map[string]any {
    switch v := documents.(type) {
    case string:
        var parsed any
        if err := json.Unmarshal([]byte(v), &parsed); err != nil {
            return nil
        }
        return normalizeDocuments(parsed, recordType, autoTimestamp, ctx)
    case map[string]any:
        writeTime := time.Now().UnixMilli()
        v["__write_time"] = writeTime
        v["__task_id"] = ctx.GetTaskID()
        if recordType != "" {
            v["__record_type"] = recordType
        }
        if _, ok := v["@timestamp"]; !ok && autoTimestamp {
            v["@timestamp"] = writeTime
        }
        return []map[string]any{v}
    case []any:
        results := []map[string]any{}
        for _, item := range v {
            if nested := normalizeDocuments(item, recordType, autoTimestamp, ctx); nested != nil {
                results = append(results, nested...)
            }
        }
        return results
    default:
        return nil
    }
}
```

---

## 附录 B：接口详情

### B.1 批量写文档接口

**请求：**

```http
POST /api/vega-backend/in/v1/resources/dataset/{datasetID}/docs
Content-Type: application/json

[
  {"id": "doc1", "text": "content1", "@timestamp": 1234567890000},
  {"id": "doc2", "text": "content2", "@timestamp": 1234567890000}
]
```

**响应：**

```json
{
  "ids": ["generated_id_1", "generated_id_2"]
}
```

### B.2 创建资源接口

**请求：**

```http
POST /api/vega-backend/v1/resources
Content-Type: application/json

{
  "catalog_id": "default",
  "name": "ds_example",
  "category": "dataset",
  "status": "active",
  "schema_definition": [...]
}
```

**响应：**

```json
{
  "id": "generated_dataset_id",
  "name": "ds_example",
  "category": "dataset"
}
```

---

## 附录 C：节点注册

### C.1 操作符常量

在 `common/actionMap.go` 中添加：

```go
const (
    OpDatasetWriteDocs = "@dataset/write-docs"
)
```

### C.2 Action 映射

在 `ActionMap` 中添加：

```go
var ActionMap = map[string]string{
    OpDatasetWriteDocs: "dataset/writedocs.json",
}
```

### C.3 Schema 文件

创建 `schemas/dataset/writedocs.json` 定义节点配置表单。
