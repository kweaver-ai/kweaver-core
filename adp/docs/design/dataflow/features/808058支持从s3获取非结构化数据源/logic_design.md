# 808058 【Dataflow】支持从S3获取非结构化数据源 —— 逻辑设计

---

## 一、需求分析

### 1.1 需求背景

#### 需求信息

| 字段 | 内容 |
|---|---|
| 需求号 | 808058 |
| 类型 | Feature |
| 标题 | 【Dataflow】支持从S3获取非结构化数据源 |
| 状态 | In Planning |
| 负责人 | 燕楠 |
| 需求来源 | 产品规划 / 产品反馈 |

#### 需求场景

当前系统的数据流触发方式主要依赖于**添加文档库**作为数据源入口，用户需要将文件上传到文档库后才能触发数据流处理。这种方式存在以下局限性:

- **数据源单一**：仅支持文档库作为数据源，无法直接对接外部存储系统
- **扩展性受限**：难以支持其他云存储服务（如OSS、COS等）的接入

在企业数据处理场景中，大量非结构化数据（如文档、图片、视频等）通常存储在S3或其他对象存储服务中。用户期望能够直接从这些存储服务中读取数据并触发数据流处理，而无需经过文档库的中转。

#### 用户期望

本次需求目标包括：

1. **扩展数据源类型**
   - 支持S3作为数据流的触发源
   - 保持与现有文档库触发方式的一致性

2. **简化操作流程**
   - 用户仅可通过**上传文件**方式准备数据
   - 系统自动管理文件存储路径（`dataflow-doc/{dag_id}`）
   - 不支持配置S3存储桶或已有路径

3. **提升数据处理实时性**
   - 支持直接读取S3中的文件
   - 为后续支持S3事件触发预留扩展空间

4. **降低存储成本**
   - 避免数据在多个存储系统间的冗余
   - 直接从源存储读取数据

5. **提升系统扩展性**
   - 建立统一的外部存储接入框架
   - 为后续支持其他云存储服务（OSS、COS等）奠定基础

---

### 1.2 用户故事

| 角色 | 痛点（Why） | 活动（What） | 价值（Value） |
|---|---|---|---|
| 数据分析师 | S3中的数据需要下载后上传到文档库才能处理，操作繁琐 | 直接配置S3数据源触发数据流 | 简化操作流程，提升工作效率 |
| 系统管理员 | 数据在S3和文档库中重复存储，造成存储成本增加 | 直接从S3读取数据进行处理 | 降低存储成本，优化资源利用 |
| 开发者 | 需要对接多种云存储服务，但缺乏统一的接入方式 | 使用统一的外部存储接入框架 | 提升系统扩展性，降低开发成本 |

---

### 1.3 交互设计示意图

![S3数据读取配置界面](4a8a6ad004f59c46fb89478f4d9f69c6.jpg)

界面说明：
- **触发方式**：提供两种触发方式的单选项
  - 添加文档库（现有方式）
  - S3 数据读取配置（新增方式）
   - **S3功能项**：
     - **上传文件**：点击上传按钮上传文件（自动存储至 `dataflow-doc/{dag_id}`）
     - **文件列表**：展示已上传的文件列表，支持删除
- **仅支持上传模式**：不支持手动输入S3路径

---

## 二、业务功能设计

### 2.1 概念与术语

| 中文 | 英文 | 定义 |
|---|---|---|
| 数据源 | Data Source | 数据流的输入来源，可以是文档库或外部存储系统 |
| 文档库 | Document Library | 系统内置的文档存储服务 |
| S3存储桶 | S3 Bucket | Amazon S3或兼容S3协议的对象存储服务中的存储容器 |
| 对象路径 | Object Path | S3存储桶中对象的完整路径，支持前缀匹配 |
| 触发方式 | Trigger Method | 启动数据流执行的方式，包括文档库触发和S3触发 |

---

### 2.2 业务用例

#### 用例名称

**配置S3数据源触发数据流**

#### 用例说明

| 项目 | 描述 |
|---|---|
| 参与者 | 数据流设计者 |
| 前置条件 | 1. 用户具有数据流编辑权限<br>2. 已配置S3访问凭证（AccessKey、SecretKey）<br>3. S3存储桶已存在且可访问 |
| 后置条件 | S3数据源配置完成，数据流可以从S3读取文件并执行 |
| 主要流程 | 1. 用户选择"S3 数据读取配置"触发方式<br>2. 界面显示上传组件<br>3. 用户上传文件（自动保存到S3）<br>4. 列表显示已上传文件<br>5. 保存配置 |
| 异常流程 | 1. 上传失败：提示网络或权限错误<br>2. S3不可用：提示联系管理员 |

---

### 2.3 业务功能定义

#### 2.3.1 触发方式选择

用户在配置数据流时，可以选择以下两种触发方式之一：

1. **添加文档库**（现有功能）
   - 从系统内置的文档库中选择文档
   - 支持选择多个文档库
   - 文档库类型包括：个人文档库、部门文档库、自定义文档库

2. **S3 数据读取配置**（新增功能）
   - 支持**上传文件**：直接上传文件触发流程（路径固定为 `dataflow-doc/{dag_id}`）
   - Bucket通过环境变量统一配置，用户不可选

> [!IMPORTANT]
> 两种触发方式为互斥关系，用户在一个数据流中只能选择其中一种触发方式。

---

#### 2.3.2 S3数据源配置

##### 配置项说明

| 配置项 | 是否必填 | 说明 | 验证规则 |
|---|---|---|---|
| 上传文件 | 是 | 上传需要处理的文件 | 1. 文件名合法性检查<br>2. 大小限制（可选，默认100MB） |

> [!NOTE]
> 1. Bucket名称由环境变量 `S3_BUCKET_NAME` 统一指定。
> 2. 文件**强制**存储在 `dataflow-doc/{dag_id}/` 目录下。
> 3. **新建流程场景**：若流程尚未创建（无dag_id），前端可生成临时Session ID（如UUID）作为路径的一部分（`dataflow-doc/temp/{session_id}/`）上传文件。流程创建时，后端会自动将文件移动至正式目录。

---

#### 2.3.3 S3凭证管理

##### 凭证配置方式

系统支持以下配置方式：

1. **全局配置**（推荐）
   - 通过Kubernetes ConfigMap和Secret进行配置
   - 服务启动时从环境变量加载
   - 所有数据流共享同一套凭证
   - 适用于企业内部统一的S3服务

> [!NOTE]
> 本期实现采用**全局配置**方式。

##### 凭证存储安全

- 敏感信息（AccessKey、SecretKey）通过Kubernetes Secret配置
- 服务端仅从环境变量中读取
- 日志中不得记录明文凭证信息

---

#### 2.3.4 文件读取策略

##### 读取模式

1. **全量读取**
   - 读取指定路径下的所有文件
   - 适用于批量处理场景

2. **增量读取**（后续扩展）
   - 仅读取新增或修改的文件
   - 需要记录上次读取的时间戳
   - 适用于定时任务场景

> [!NOTE]
> 本期实现**全量读取**模式，增量读取作为后续扩展功能。

---

## 三、技术设计

### 3.1 系统架构

#### 3.1.1 整体架构

```mermaid
graph TB
    subgraph "前端层"
        UI[数据流配置界面]
    end
    
    subgraph "服务层"
        API[DataFlow API]
        Validator[配置验证器]
        Executor[数据流执行器]
    end
    
    subgraph "适配器层"
        DocAdapter[文档库适配器]
        S3Adapter[S3适配器]
    end
    
    subgraph "存储层"
        MongoDB[(MongoDB)]
        DocLib[文档库服务]
        S3[S3存储]
    end
    
    UI -->|上传文件/配置| API
    API -->|验证系统配置| Validator
    Validator -->|检查Bucket| S3Adapter
    API -->|上传文件| S3Adapter
    API -->|保存配置| MongoDB
    Executor -->|读取配置| MongoDB
    Executor -->|读取S3文件| S3Adapter
    S3Adapter -->|调用SDK| S3
```

#### 3.1.2 数据流程

##### 场景一：已有流程上传文件

```mermaid
sequenceDiagram
    participant User as 用户
    participant UI as 前端界面
    participant API as DataFlow API
    participant S3Adapter as S3适配器
    participant MongoDB as MongoDB
    participant S3 as S3存储
    
    User->>UI: 选择S3数据源
    UI->>User: 显示上传组件 (已关联DAG ID)
    User->>UI: 点击上传文件
    UI->>API: 上传文件流 (POST /files/upload)
    API->>S3Adapter: 上传文件
    S3Adapter->>S3: PutObject (dataflow-doc/{dag_id})
    S3-->>S3Adapter: 上传成功
    S3Adapter-->>API: 返回文件信息
    API-->>UI: 显示已上传文件
    User->>UI: 保存配置
    UI->>API: 提交配置 (包含文件列表)
    API->>MongoDB: 保存数据流配置
    MongoDB-->>API: 保存成功
    API-->>UI: 返回成功
```

##### 场景二：新建流程上传文件 (Temp Upload + Move)

```mermaid
sequenceDiagram
    participant User as 用户
    participant UI as 前端界面
    participant API as DataFlow API
    participant S3Adapter as S3适配器
    participant MongoDB as MongoDB
    participant S3 as S3存储
    
    User->>UI: 创建新流程
    UI->>UI: 生成临时Session ID (UUID)
    User->>UI: 点击上传文件
    UI->>API: 上传文件 (POST /{session_id}/files/upload)
    API->>S3Adapter: 识别为临时ID
    S3Adapter->>S3: PutObject (dataflow-doc/temp/{session_id})
    API-->>UI: 返回临时文件息
    
    User->>UI: 保存/创建流程
    UI->>API: 提交创建请求 (Sources包含临时路径Key)
    API->>MongoDB: 生成正式DAG ID
    API->>S3Adapter: MoveS3Files (Temp -> Formal)
    S3Adapter->>S3: CopyObject (Temp -> Formal)
    S3Adapter->>S3: DeleteObject (Temp)
    API->>MongoDB: 保存配置 (使用正式路径Key)
    API-->>UI: 返回创建成功 (包含正式DAG ID)
```

---

### 3.2 数据模型设计

#### 3.2.1 数据流配置模型

采用通用的 `trigger_config` 结构，通过 `operator` 区分数据源类型。

```go
// TriggerConfig 触发器配置
type TriggerConfig struct {
    Operator   string      `json:"operator" bson:"operator"`     // 例如 "@trigger/manual"
    DataSource *DataSource `json:"dataSource" bson:"dataSource"` // 数据源配置
}

// DataSource 数据源
type DataSource struct {
    Operator   string                 `json:"operator" bson:"operator"`     // 例如 "@s3/list-objects"
    Parameters map[string]interface{} `json:"parameters" bson:"parameters"` // 具体的参数
}
```

**S3数据源参数 (`@s3/list-objects`)：**

| 参数名 | 类型 | 说明 |
|---|---|---|
| mode | string | 数据源模式，取值：`upload` (上传文件)、`path` (指定路径)。<br>**注**：本期仅支持 `upload` 模式。 |
| sources | array | S3数据源列表 |
| sources[].key | string | S3对象Key（`upload` 模式下必填） |
| sources[].name | string | 文件名 |
| sources[].size | int64 | 文件大小 |
| sources[].path | string | 路径前缀（仅用于 `path` 模式，本期禁用） |


#### 3.2.2 S3全局配置（ConfigMap）

全局配置不存储在数据库中，而是通过配置文件或环境变量加载。

| 配置项 | 环境变量/配置Key | 说明 |
|---|---|---|
| Endpoint | S3_ENDPOINT | S3服务端点 |
| Region | S3_REGION | S3区域 |
| AccessKeyID | S3_ACCESS_KEY_ID | 访问密钥ID |
| SecretAccessKey | S3_SECRET_ACCESS_KEY | 访问密钥 |
| BucketName | S3_BUCKET_NAME | 默认Bucket名称 |
| UseSSL | S3_USE_SSL | 是否使用SSL |


---

### 3.3 接口设计

#### 3.3.1 数据流配置接口

##### 创建/更新数据流配置

**请求**

```http
POST /api/v1/dataflows
Content-Type: application/json
Authorization: Bearer {token}

{
  "name": "S3数据处理流程",
  "trigger_config": {
    "operator": "@trigger/manual",
    "dataSource": {
      "operator": "@s3/list-objects",
      "parameters": {
        "mode": "upload",
        "sources": [
          {
            "key": "dataflow-doc/df_123/test.pdf",
            "name": "test.pdf",
            "size": 1024
          }
        ]
      }
    }
  }
}
```

**响应**

```json
{
  "id": "df_123456",
  "name": "S3数据处理流程",
  "trigger_config": {
     "operator": "@trigger/manual",
     "dataSource": {
       "operator": "@s3/list-objects",
       "parameters": {
         "sources": [...]
       }
     }
  }
}
```


##### 验证S3配置

**请求** (验证环境是否可用)

```http
POST /api/v1/dataflows/validate-s3
Content-Type: application/json
Authorization: Bearer {token}

{}
```

**响应**

```json
{
  "valid": true,
  "bucket_exists": true
}
```

---

#### 3.3.2 文件管理接口

##### 上传文件

**请求**

- **URL参数**: `id` 为DAG ID 或 临时Session ID (UUID)

```http
POST /api/v1/dataflows/{id}/files/upload
Content-Type: multipart/form-data
Authorization: Bearer {token}

file: (binary)
```

**响应 (已有DAG ID)**

```json
{
  "id": "bucket-name/dataflow-doc/df_123456/example.pdf",
  "bucket": "bucket-name",
  "key": "dataflow-doc/df_123456/example.pdf",
  "name": "example.pdf",
  "size": 1024,
  "last_modified": "2026-02-02T10:00:00Z",
  "download_url": "https://s3.example.com/..."
}
```

**响应 (临时Session ID)**

```json
{
  "id": "bucket-name/dataflow-doc/temp/uuid-1234/example.pdf",
  "bucket": "bucket-name",
  "key": "dataflow-doc/temp/uuid-1234/example.pdf",
  "name": "example.pdf",
  "size": 1024,
  "last_modified": "2026-02-02T10:00:00Z",
  "download_url": "https://s3.example.com/..."
}
```

##### 列举文件

**请求**

```http
GET /api/v1/dataflows/{id}/files
Authorization: Bearer {token}
```

**响应**

```json
{
  "files": [
    {
      "id": "bucket-name/dataflow-doc/df_123456/example.pdf",
      "bucket": "bucket-name",
      "key": "dataflow-doc/df_123456/example.pdf",
      "name": "example.pdf",
      "size": 1024,
      "last_modified": "2026-02-02T10:00:00Z",
      "download_url": "https://s3.example.com/..."
    }
  ]
}
```

##### 删除文件

**请求**

```http
DELETE /api/v1/dataflows/{id}/files?key=dataflow-doc/df_123456/example.pdf
Authorization: Bearer {token}
```

**响应**

```http
HTTP/1.1 204 No Content
```

HTTP/1.1 204 No Content
```

##### 下载文件

**请求**

- **URL参数**: `key` 文件的完整路径

```http
GET /api/v1/dataflows/{id}/files/download?key=dataflow-doc/df_123456/example.pdf
Authorization: Bearer {token}
```

**响应**

- **状态码**: `302 Found`
- **Location**: S3预签名下载URL (有效期1小时)

```http
HTTP/1.1 302 Found
Location: https://s3.example.com/bucket-name/dataflow-doc/df_123456/example.pdf?X-Amz-Algorithm=...
```

---



### 3.4 核心模块设计

#### 3.4.1 S3适配器模块

**主要职责：**
- 封装S3 SDK，提供统一的文件读取接口
- 验证S3 Bucket和路径的有效性
- 列出指定路径下的所有对象
- 读取S3对象内容和元数据

**核心功能：**

| 功能 | 说明 |
|---|---|
| ValidateBucket | 验证Bucket是否存在且可访问 |
| ValidatePath | 验证路径是否可访问 |
| ListObjects | 列出指定路径下的所有对象 |
| GetObject | 获取指定对象的内容 |
| GetObjectMetadata | 获取对象元数据（大小、类型、修改时间等） |
| **CopyObject** | **复制S3对象（用于文件移动）** |
| **DeleteObject** | **删除S3对象** |

**设计要点：**
- 使用AWS SDK for Go或MinIO SDK
- 支持连接池复用，提升性能
- 实现重试机制，处理网络异常
- 统一错误处理和日志记录

---

#### 3.4.2 配置验证器模块

**主要职责：**
- 验证S3配置的有效性
- 提供详细的错误信息
- 支持异步验证

**验证流程：**

```mermaid
graph TD
    A[开始验证] --> B{Bucket名称格式}
    B -->|格式错误| C[返回错误]
    B -->|格式正确| D{Bucket是否存在}
    D -->|不存在| C
    D -->|存在| E{路径是否指定}
    E -->|未指定| F[验证通过]
    E -->|已指定| G{路径是否可访问}
    G -->|不可访问| C
    G -->|可访问| F
```

**验证规则：**

| 验证项 | 规则 |
|---|---|
| Bucket名称格式 | 验证环境变量配置的Bucket名称格式 |
| Bucket存在性 | 调用S3 API验证Bucket是否存在 |
| 访问权限 | 验证凭证是否具有Bucket的读取权限 |
| 路径有效性 | 如果指定路径，验证路径是否可访问 |

---

#### 3.4.3 数据流执行器扩展

**主要职责：**
- 根据触发类型选择数据源（文档库或S3）
- 从S3读取文件列表并逐个处理
- 实现错误处理和重试机制
- 支持并发处理提升性能

**执行流程：**

```mermaid
graph TD
    A[开始执行] --> B{触发类型}
    B -->|文档库| C[执行文档库触发逻辑]
    B -->|S3| D[执行S3触发逻辑]
    D --> E[遍历S3数据源]
    E --> F[列出对象列表]
    F --> G[并发处理对象]
    G --> H[读取对象内容]
    H --> I[执行数据流处理]
    I --> J{是否有更多数据源}
    J -->|是| E
    J -->|否| K[执行完成]
```

**设计要点：**
- 支持并发处理多个文件（工作池模式）
- 单个文件失败不影响整体流程
- 记录详细的执行日志
- 支持大文件的流式处理

---

#### 3.4.4 文件移动逻辑 (CreateDataFlow)

**主要职责：**
- 在流程创建时，识别配置中的临时文件路径
- 将文件从临时目录移动到正式DAG目录
- 更新流程配置中的文件路径

**执行流程：**
1. 解析 `TriggerConfig.DataSource.Parameters`
2. 检查 `sources` 列表中是否包含 `dataflow-doc/temp/` 前缀的Key
3. 如果存在：
   - 生成正式DAG ID (如果尚未生成)
   - 遍历临时Key列表
   - 调用 `S3Adapter.CopyObject` 将文件复制到 `dataflow-doc/{dag_id}/`
   - 调用 `S3Adapter.DeleteObject` 删除临时文件
   - 更新 `sources` 列表中的Key为新路径
   - 保存更新后的配置到数据库

---

#### 3.4.4 S3数据源输出结构

**主要职责:**
- 定义S3触发器的标准输出格式
- 提供下游操作符所需的完整S3对象信息
- 生成安全的预签名下载链接

**输出数据结构:**

```json
{
  "_type": "s3",
  "id": "bucket-name/path/to/object.txt",
  "bucket": "bucket-name",
  "key": "path/to/object.txt",
  "name": "object.txt",
  "size": "1024",
  "last_modified": "2026-02-02T09:30:31Z",
  "etag": "\"5d41402abc4b2a76b9719d911017c592\"",
  "md5": "\"5d41402abc4b2a76b9719d911017c592\"",
  "download_url": "https://s3.example.com/bucket-name/path/to/object.txt?X-Amz-Algorithm=..."
}
```

**字段说明:**

| 字段名 | 类型 | 说明 |
|---|---|---|
| _type | string | 固定为 "s3",标识数据源类型 |
| id | string | S3对象唯一标识符,格式为 `bucket/key` |
| bucket | string | S3存储桶名称 |
| key | string | S3对象键(完整路径) |
| name | string | 文件名(从key中提取的最后一段) |
| size | string | 对象大小(字节),字符串格式 |
| last_modified | string | 最后修改时间,ISO 8601格式 |
| etag | string | S3 ETag值(通常是MD5哈希) |
| md5 | string | 与etag相同,便于使用 |
| download_url | string | 预签名下载URL,有效期7天 |

**预签名URL特性:**

- **有效期**: 7天(168小时)
- **授权方式**: URL中包含临时访问凭证
- **安全性**: 无需暴露S3访问密钥给最终用户
- **容错机制**: 如果生成失败,降级为简单的endpoint URL


---

### 3.5 安全设计

#### 3.5.1 凭证安全配置

**设计原则：**
- 敏感信息（AccessKey、SecretKey）不得硬编码在代码或镜像中
- 使用Kubernetes Secret存储敏感凭证
- 通过环境变量注入到容器中

**配置示例：**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dataflow-s3-secret
type: Opaque
data:
  access_key_id: QUtJ... # Base64 encoded
  secret_access_key: d0ph... # Base64 encoded
```

**安全要求：**
- 生产环境必须使用Secret
- 访问凭证应遵循最小权限原则
- 定期轮换访问凭证


---

#### 3.5.2 权限控制

| 接口 | 所需权限 | 说明 |
|---|---|---|

| 创建/编辑数据流 | 数据流编辑权限 | 可配置S3数据源 |
| 验证S3配置 | 数据流编辑权限 | 验证配置有效性 |
| 执行数据流 | 数据流执行权限 | 读取S3文件并处理 |

---


## 四、后续扩展

### 4.1 增量读取支持

- 记录上次读取的时间戳
- 仅处理新增或修改的文件
- 支持定时任务触发

### 4.2 S3事件触发

- 监听S3 Bucket事件（通过SQS或SNS）
- 文件上传时自动触发数据流
- 实时处理能力

### 4.3 其他云存储支持

- 阿里云OSS
- 腾讯云COS
- 华为云OBS
- 统一的云存储适配器接口

### 4.4 文件类型过滤

- 支持按文件扩展名过滤
- 支持按文件大小过滤
- 支持按MIME类型过滤

### 4.5 数据流级凭证配置

- 支持在数据流中单独配置S3凭证
- 支持访问多个S3账户
- 凭证隔离和安全管理

---

## 五、风险与应对

### 5.1 技术风险

| 风险 | 影响 | 概率 | 应对措施 |
|---|---|---|---|
| S3 SDK兼容性问题 | 高 | 低 | 使用官方SDK，充分测试 |
| 大文件处理性能问题 | 中 | 中 | 实现流式处理，限制文件大小 |
| 网络不稳定导致读取失败 | 中 | 中 | 实现重试机制，记录失败日志 |
| 凭证泄露风险 | 高 | 低 | 加密存储，权限控制，日志脱敏 |

### 5.2 业务风险

| 风险 | 影响 | 概率 | 应对措施 |
|---|---|---|---|
| 用户配置错误导致无法访问 | 中 | 中 | 提供详细的错误提示和文档 |
| S3费用超预期 | 中 | 低 | 提供使用量监控和告警 |
| 数据安全合规问题 | 高 | 低 | 遵循数据安全规范，加密传输 |

---

## 六、参考资料

### 6.1 相关文档

- [AWS S3 SDK for Go](https://docs.aws.amazon.com/sdk-for-go/api/service/s3/)
- [S3 Bucket命名规则](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html)
- [MinIO Go Client SDK](https://min.io/docs/minio/linux/developers/go/API.html)

### 6.2 相关需求

- 792990 【Dataflow】数据流支持子流程调用
- 801755 【Dataflow】数据流版本管理

---

## 附录

### A. 错误码定义

| 错误码 | 错误信息 | 说明 |
|---|---|---|
| S3_INVALID_CREDENTIALS | S3访问凭证无效 | AccessKey或SecretKey错误 |
| S3_BUCKET_NOT_FOUND | S3存储桶不存在 | Bucket名称错误或不存在 |
| S3_ACCESS_DENIED | 无权访问S3存储桶 | 凭证权限不足 |
| S3_INVALID_PATH | S3路径格式错误 | 路径格式不符合规范 |
| S3_NETWORK_ERROR | 无法连接到S3服务 | 网络连接问题 |
| S3_FILE_NOT_FOUND | S3文件不存在 | 文件已被删除或路径错误 |
| S3_READ_ERROR | 读取S3文件失败 | 文件读取过程中出错 |
| S3_FILE_TOO_LARGE | S3文件超过大小限制 | 文件大小超过系统限制 |

### B. 配置示例

#### B.1 S3全局配置（环境变量）

```bash
# S3服务端点
S3_ENDPOINT=https://s3.amazonaws.com

# S3区域
S3_REGION=us-east-1

# 访问凭证
S3_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
S3_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
S3_BUCKET_NAME=my-app-data

# SSL开关
S3_USE_SSL=true
```


#### B.2 数据流S3配置示例 (完整JSON)

```json
{
    "id": "df_s3_example",
    "title": "S3数据流示例",
    "description": "从S3读取数据并处理",
    "status": "normal",
    "type": "data-flow",
    "steps": [
        {
            "id": "0",
            "title": "触发器",
            "operator": "@trigger/dataflow-s3"
        },
        {
            "id": "1",
            "title": "自定义处理",
            "operator": "@custom/process",
            "parameters": {}
        }
    ],
    "trigger_config": {
        "operator": "@trigger/manual",
        "dataSource": {
            "operator": "@s3/list-objects",
            "parameters": {
                "mode": "upload",
                "source": [{
                  "path": "dataflow-doc/temp/temp-d290f1ee-6c54-4b01-90e6-d701748f0851/doc1.txt"
                }]
            }
        }
    }
}
```

