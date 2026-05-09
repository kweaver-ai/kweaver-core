# 808924 文档解析节点适配 MinerU 官方 API —— 逻辑设计

---

## 一、需求分析

### 1.1 需求背景

#### 需求信息

| 字段     | 内容                                    |
| -------- | --------------------------------------- |
| 需求号   | 808924                                  |
| 类型     | Feature                                 |
| 标题     | 文档解析节点适配 MinerU 官方 API        |
| 状态     | In Realizing                            |
| 需求来源 | 产品规划                                |

#### 需求场景

文档解析节点依赖部署的 MinerU 文档解析服务（file_parse 接口），需要自行部署和维护 MinerU 服务。为了简化开发和部署，提供接入 MinerU 官方 API 的模式，使系统可以：

- 直接调用 MinerU 官方云端 API 进行文档解析
- 无需自行部署 MinerU 服务
- 支持非公网可访问的文件 URL（通过流式上传解决）
- 保持与内部服务相同的接口，对调用方透明

#### 用户期望

本次需求目标包括：

1. **支持官方 API 模式**
   - 通过配置切换使用内部服务或官方 API
   - 无需修改业务代码
   - 支持标准 MinerU API 协议

2. **保持接口兼容性**
   - 保持与内部服务相同的接口签名
   - 业务代码无需修改
   - 无缝切换解析后端

3. **灵活的配置管理**
   - 支持 UseMineru 配置开关
   - 支持配置 API 地址和 Token
   - 通过环境变量管理配置

4. **完整的文档解析支持**
   - 支持 PDF、Word、Excel 等文档格式
   - 返回 Markdown 内容
   - 返回结构化内容列表（content_list）

5. **解决内网文件问题**
   - 支持非公网可访问的文件 URL
   - 通过流式上传解决数据源访问问题

---

### 1.2 用户故事

| 角色           | 痛点（Why）                          | 活动（What）                 | 价值（Value）                   |
| -------------- | ------------------------------------ | ---------------------------- | ------------------------------ |
| 社区版部署者   | 需要自行部署 MinerU 服务             | 配置官方 API 模式            | 无需部署 MinerU 服务            |
| 系统运维人员   | 需要维护 MinerU 服务集群             | 切换到官方 API               | 降低运维成本                   |
| 应用开发者     | 希望代码在不同部署模式下保持一致     | 使用统一的解析接口           | 降低开发和维护成本             |
| 数据流设计者   | 需要解析内网存储的文档               | 使用流式上传方案             | 支持任意文件 URL               |

---

## 二、业务功能设计

### 2.1 概念与术语

| 中文         | 英文                    | 定义                                           |
| ------------ | ----------------------- | ---------------------------------------------- |
| MinerU API   | MinerU API              | MinerU 官方提供的云端文档解析 API              |
| 预签名 URL   | Presigned URL           | 临时授权的上传地址，具有过期时间               |
| 任务 ID      | Task ID                 | 文档解析任务的唯一标识                         |
| 批量上传     | Batch Upload            | 通过获取预签名 URL 上传文件的流程              |
| 内容列表     | Content List            | 文档解析输出的结构化内容 JSON                  |

---

### 2.2 业务用例

#### 用例名称

**使用 MinerU 官方 API 解析文档**

#### 用例说明

| 项目     | 描述                                           |
| -------- | ---------------------------------------------- |
| 参与者   | 数据流设计者、应用开发者                       |
| 前置条件 | 已配置 MinerU API Token；文件 URL 可访问       |
| 后置条件 | 文档解析完成，返回 Markdown 和结构化内容       |

---

### 2.3 业务功能定义

#### 配置开关

##### UseMineru 字段

在 `StructureExtractor` 配置中通过 `UseMineru` 字段控制使用哪种解析模式：

| 字段名     | 类型   | 可选值        | 描述                   |
| ---------- | ------ | ------------- | ---------------------- |
| UseMineru  | bool   | true / false  | 是否使用 MinerU 官方 API |

##### 解析模式选择逻辑

- **UseMineru = true**：使用 `FileParseMineru` 实现，调用 MinerU 官方 API
- **UseMineru = false**：使用 `FileParseInternal` 实现，调用内部部署服务

---

#### MinerU API 调用流程

##### 流程步骤

```
1. 获取上传地址 (POST /api/v4/file-urls/batch)
   - 请求体: 文件名、解析参数等
   - 响应: task_id、预签名上传 URL、上传 Headers

2. 上传文件到预签名地址 (PUT 预签名URL)
   - 使用响应中的 Headers（如 Content-Type）
   - 流式上传文件二进制内容

3. 轮询任务状态 (GET /api/v4/extract/task/{task_id})
   - 状态: pending -> processing -> done / failed
   - 轮询间隔: 2秒
   - 最大等待时间: 30分钟

4. 获取解析结果
   - 下载 full_zip_url 指向的 zip 文件
   - 解压获取 content_list.json 和 full.md
```

##### 解析参数

| 参数名         | 类型     | 默认值  | 描述                           |
| -------------- | -------- | ------- | ------------------------------ |
| is_ocr         | bool     | false   | 是否启用 OCR                   |
| enable_formula | bool     | true    | 是否启用公式识别               |
| enable_table   | bool     | true    | 是否启用表格识别               |
| is_chem        | bool     | false   | 是否启用化学结构识别           |
| model_version  | string   | "vlm"   | 解析模型版本                   |
| language       | *string  | null    | OCR 语言（null 为自动检测）    |

---

#### 配置说明

##### 配置文件

MinerU 相关配置在环境变量或配置文件中：

```yaml
structureextractor:
  use_mineru: true
  mineru_base_url: https://mineru.net
  mineru_token: <your-token>

  # 内部服务配置（UseMineru=false 时使用）
  private_host: <internal-host>
  private_port: <internal-port>
```

##### 环境变量配置

| 环境变量                              | 描述                       |
| ------------------------------------- | -------------------------- |
| STRUCTUREEXTRACTOR_USE_MINERU         | 是否使用 MinerU 官方 API   |
| STRUCTUREEXTRACTOR_BASE_URL           | MinerU API 地址            |
| STRUCTUREEXTRACTOR_MINERU_TOKEN       | MinerU API Token           |

---

### 2.4 业务流程

#### 2.4.1 文档解析流程选择

```
开始
  ↓
调用 FileParse()
  ↓
读取 UseMineru 配置
  ↓
UseMineru == true?
  ├─ 是 → 调用 FileParseMineru()
  │        ↓
  │      官方 API 解析流程
  │
  └─ 否 → 调用 FileParseInternal()
           ↓
         内部服务解析流程
  ↓
返回解析结果
  ↓
结束
```

#### 2.4.2 MinerU 官方 API 解析流程

```
开始
  ↓
从 fileUrl 下载文件（流式）
  ↓
调用 getMineruUploadInfo() 获取上传信息
  ↓
POST /api/v4/file-urls/batch
  - 请求: 文件名、解析参数
  - 响应: task_id, file_url, headers
  ↓
调用 uploadFileToMineru() 上传文件
  ↓
PUT 预签名 URL
  - 使用响应中的 headers
  - 流式上传文件内容
  ↓
调用 pollMineruTaskResult() 轮询结果
  ↓
GET /api/v4/extract/task/{task_id}
  - 轮询直到 state = "done"
  ↓
调用 downloadAndExtractMineruResult() 获取结果
  ↓
下载 zip 文件并解压
  - 提取 full.md
  - 提取 content_list.json
  ↓
返回解析结果
  ↓
结束
```

---

## 三、技术设计

### 3.1 架构设计

#### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     业务层                                   │
│  (数据流、工作流、文档解析节点等)                             │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      │ 调用 FileParse 接口
                      ↓
┌─────────────────────────────────────────────────────────────┐
│              StructureExtractor 接口层                      │
│  - FileParse(ctx, fileUrl, fileName)                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
          ┌───────────┴───────────┐
          │                       │
          ↓                       ↓
┌─────────────────────┐  ┌─────────────────────┐
│   官方 API 实现     │  │   内部服务实现      │
│   FileParseMineru   │  │   FileParseInternal │
└──────────┬──────────┘  └──────────┬──────────┘
           │                        │
           │ HTTP API               │ HTTP POST
           ↓                        ↓
┌─────────────────────┐  ┌─────────────────────┐
│   MinerU 官方 API   │  │   内部 MinerU 服务  │
│   (mineru.net)      │  │   (file_parse)      │
└─────────────────────┘  └─────────────────────┘
```

#### 模块划分

| 模块名称           | 文件路径                                      | 职责                       |
| ------------------ | --------------------------------------------- | -------------------------- |
| 结构解析接口       | `drivenadapters/structure_extractor.go`      | 文档解析接口定义           |
| 官方 API 实现      | `drivenadapters/structure_extractor.go`      | MinerU 官方 API 调用实现   |
| 内部服务实现       | `drivenadapters/structure_extractor.go`      | 内部 MinerU 服务调用实现   |
| 配置定义           | `common/config.go`                           | 配置结构定义               |
| HTTP 客户端        | `drivenadapters/http_client.go`              | HTTP 请求封装              |

---

### 3.2 数据结构设计

#### MineruFile

```go
type MineruFile struct {
    Name       string `json:"name"`                  // 文件名
    DataID     string `json:"data_id"`               // 数据 ID
    PageRanges string `json:"page_ranges,omitempty"` // 页码范围（可选）
}
```

#### MineruFileUrlsBatchReq

```go
type MineruFileUrlsBatchReq struct {
    IsOCR         bool         `json:"is_ocr"`
    EnableFormula bool         `json:"enable_formula"`
    EnableTable   bool         `json:"enable_table"`
    ModelVersion  string       `json:"model_version"`
    Language      *string      `json:"language,omitempty"`
    IsChem        bool         `json:"is_chem"`
    Files         []MineruFile `json:"files"`
}
```

#### MineruFileUrlsBatchResp

```go
type MineruFileUrlsBatchResp struct {
    Code    int                         `json:"code"`
    Message string                      `json:"msg"`
    TraceID string                      `json:"trace_id"`
    Data    MineruFileUrlsBatchRespData `json:"data"`
}

type MineruFileUrlsBatchRespData struct {
    BatchID  string              `json:"batch_id"`
    FileURLs []string            `json:"file_urls"`
    TaskIDs  []string            `json:"task_ids"`
    Headers  []map[string]string `json:"headers"`
}
```

#### MineruTaskResultResp

```go
type MineruTaskResultResp struct {
    Code    int                  `json:"code"`
    Msg     string               `json:"msg"`
    TraceID string               `json:"trace_id"`
    Data    MineruTaskResultData `json:"data"`
}

type MineruTaskResultData struct {
    TaskID       string             `json:"task_id"`
    State        string             `json:"state"`
    ErrMsg       string             `json:"err_msg"`
    FullZipURL   string             `json:"full_zip_url"`
    LayoutURL    string             `json:"layout_url"`
    FullMDLink   string             `json:"full_md_link"`
    FileName     string             `json:"file_name"`
    URL          string             `json:"url"`
    Type         string             `json:"type"`
    FileInfo     MineruTaskFileInfo `json:"file_info"`
    ModelVersion string             `json:"model_version"`
    IsChem       bool               `json:"is_chem"`
    ImageReady   bool               `json:"image_ready"`
}
```

#### FileParseResultItem

```go
type FileParseResultItem struct {
    Images      map[string]string `json:"images"`
    MdContent   string            `json:"md_content"`
    ContentList string            `json:"content_list"`
}
```

---

### 3.3 接口设计

#### StructureExtractor 接口

```go
type StructureExtractor interface {
    FileParse(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error)
}
```

#### 核心方法

```go
// 根据配置选择解析模式
func (s *structureExtractor) FileParse(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error)

// 使用 MinerU 官方 API 解析
func (s *structureExtractor) FileParseMineru(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error)

// 使用内部服务解析
func (s *structureExtractor) FileParseInternal(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error)

// 获取 MinerU 上传信息
func (s *structureExtractor) getMineruUploadInfo(ctx context.Context, reqBody *MineruFileUrlsBatchReq) (*MineruFileUrlsBatchResp, error)

// 上传文件到 MinerU
func (s *structureExtractor) uploadFileToMineru(ctx context.Context, uploadURL string, headers map[string]string, fileContent io.Reader) error

// 轮询任务结果
func (s *structureExtractor) pollMineruTaskResult(ctx context.Context, taskID string) (*MineruTaskResultData, error)

// 获取任务结果
func (s *structureExtractor) getMineruTaskResult(ctx context.Context, taskID string) (*MineruTaskResultData, error)

// 下载并解压结果
func (s *structureExtractor) downloadAndExtractMineruResult(ctx context.Context, zipURL string) (*FileParseResultItem, []*ContentItem, error)
```

---

### 3.4 API 规范

#### 3.4.1 获取上传地址

**请求**

```
POST /api/v4/file-urls/batch
Authorization: Bearer <token>
Content-Type: application/json

{
    "is_ocr": false,
    "enable_formula": true,
    "enable_table": true,
    "model_version": "vlm",
    "language": null,
    "is_chem": false,
    "files": [
        {
            "name": "document.pdf",
            "data_id": "unique-id",
            "page_ranges": ""
        }
    ]
}
```

**响应**

```json
{
    "code": 0,
    "msg": "ok",
    "trace_id": "xxx",
    "data": {
        "batch_id": "batch-uuid",
        "file_urls": ["https://oss-endpoint/...?signature=xxx"],
        "task_ids": ["task-uuid"],
        "headers": [{"Content-Type": "application/pdf"}]
    }
}
```

#### 3.4.2 上传文件

**请求**

```
PUT <预签名URL>
Content-Type: application/pdf

<文件二进制内容>
```

**响应**

```
200 OK
```

#### 3.4.3 查询任务结果

**请求**

```
GET /api/v4/extract/task/{task_id}
Authorization: Bearer <token>
```

**响应**

```json
{
    "code": 0,
    "msg": "ok",
    "trace_id": "xxx",
    "data": {
        "task_id": "task-uuid",
        "state": "done",
        "err_msg": "",
        "full_zip_url": "https://oss-endpoint/result.zip",
        "full_md_link": "https://oss-endpoint/full.md",
        "file_name": "document.pdf",
        "file_info": {
            "pages": 10,
            "file_size": 102400
        }
    }
}
```

#### 3.4.4 任务状态说明

| 状态       | 说明         |
| ---------- | ------------ |
| pending    | 任务排队中   |
| processing | 任务处理中   |
| done       | 任务完成     |
| failed     | 任务失败     |

---

## 四、错误处理

### 4.1 错误类型

| 错误类型         | 描述                                   |
| ---------------- | -------------------------------------- |
| 文件下载失败     | 从 fileUrl 下载文件时网络错误          |
| 获取上传地址失败 | API 返回 code != 0 或网络错误          |
| 文件上传失败     | PUT 上传失败                           |
| 任务执行失败     | state == "failed"，检查 err_msg        |
| 超时错误         | 超过最大等待时间（30分钟）             |
| 结果获取失败     | zip 下载失败、解压失败、文件缺失       |

### 4.2 错误日志

```go
// 文件下载失败
log.Warnf("FileParseMineru download source file err: %s, url: %s", err.Error(), fileUrl)

// 获取上传地址失败
log.Warnf("FileParseMineru get upload url err: %s", err.Error())

// 文件上传失败
log.Warnf("FileParseMineru upload file err: %s, task_id: %s", err.Error(), taskID)

// 任务执行失败
log.Warnf("FileParseMineru task failed: %s, task_id: %s", errMsg, taskID)

// 超时
log.Warnf("FileParseMineru task timeout: task_id: %s", taskID)

// 结果获取失败
log.Warnf("FileParseMineru download/extract zip err: %s", err.Error())
```

---

## 五、部署说明

### 5.1 前置条件

- 已获取 MinerU API Token
- MinerU 官方 API 服务可用（https://mineru.net）

### 5.2 配置步骤

1. **配置环境变量**

   ```bash
   export STRUCTUREEXTRACTOR_USE_MINERU=true
   export STRUCTUREEXTRACTOR_BASE_URL=https://mineru.net
   export STRUCTUREEXTRACTOR_MINERU_TOKEN=<your-token>
   ```

2. **或在配置文件中配置**

   ```yaml
   structureextractor:
     use_mineru: true
     mineru_base_url: https://mineru.net
     mineru_token: <your-token>
   ```

3. **验证配置**

   - 检查日志确认 MinerU 模式已启用
   - 测试文档解析功能

### 5.3 配置示例

```yaml
# 使用 MinerU 官方 API
structureextractor:
  use_mineru: true
  mineru_base_url: https://mineru.net
  mineru_token: eyJ0eXBlIjoiSldUIiwiYWxnIjoiSFM1MTIifQ...

# 使用内部服务
structureextractor:
  use_mineru: false
  private_host: mineru-service
  private_port: 8000
  output_dir: /tmp/output
  backend: vlm
```

---

## 六、附录

### 6.1 相关文件清单

| 文件路径                                                    | 说明                           |
| ----------------------------------------------------------- | ------------------------------ |
| `drivenadapters/structure_extractor.go`                    | 文档解析接口和实现             |
| `drivenadapters/http_client.go`                            | HTTP 客户端封装                |
| `common/config.go`                                         | 配置结构定义                   |
| `docs/features/808924文档解析节点适配minerU官方API/`       | 需求设计文档                   |

### 6.2 返回数据格式

#### ZIP 文件内容结构

```
{task_id}/
├── {uuid}_content_list.json    # 内容列表 (JSON)
├── full.md                      # Markdown 内容
├── layout.json                  # 布局信息
├── {uuid}_model.json           # 模型输出
├── {uuid}_origin.pdf           # 原始文件
└── images/                      # 图片目录
    └── *.jpg
```

#### content_list.json 格式

```json
[
    {
        "type": "text",
        "text": "标题内容",
        "text_level": 1,
        "bbox": [71, 59, 552, 85],
        "page_idx": 0
    },
    {
        "type": "image",
        "img_path": "images/xxx.jpg",
        "image_caption": [],
        "image_footnote": [],
        "bbox": [861, 157, 880, 174],
        "page_idx": 1
    },
    {
        "type": "table",
        "table_body": "<html>...</html>",
        "table_caption": ["表格标题"],
        "bbox": [100, 200, 500, 400],
        "page_idx": 0
    }
]
```

### 6.3 参考资料

- [MinerU 官网](https://mineru.net)
- [MinerU API 文档](https://mineru.net/doc/docs/index.html)
- [MinerU GitHub](https://github.com/opendatalab/MinerU)