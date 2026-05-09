# 【Dataflow】支持 dataflow 文件下载接口设计

## 1. 背景与范围

Dataflow 已实现 DFS 文件协议用于内部数据交换，支持上传文件触发流程。但当前缺少前端访问能力，无法在运行记录中查看或下载文件。

本期目标：提供接口用于根据 DFS 协议获取文件信息（基于数据库）、获取下载链接（基于 OssGateway）。

### 1.1 已有能力

- `t_flow_file` - 业务文件表，承载 `dfs://<file_id>` 协议
- `t_flow_storage` - 存储对象表，关联 OssGateway
- DAO 层完整实现（`FlowFileDao`, `FlowStorageDao`）
- `OssGateway.GetDownloadURL` - 获取预签名下载地址
- 触发接口 `POST /api/automation/v2/dataflow-doc/trigger/:dagId`
- 完成接口 `POST /api/automation/v2/dataflow-doc/complete`

### 1.2 本期范围

- 按流程实例维度查询文件列表
- 按文件 ID 获取单个文件信息
- 获取文件下载链接

### 1.3 本期不覆盖

- 文件删除接口
- 批量下载
- 文件内容预览（非下载）

---

## 2. 接口设计

### 2.1 获取文件列表

**请求：**

```http
GET /api/automation/v2/dataflow-doc/files?dag_instance_id={dag_instance_id}
Authorization: Bearer {token}
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| dag_instance_id | string | 是 | 流程实例 ID |

**响应：**

```json
{
  "files": [
    {
      "file_id": "604851178156619196",
      "docid": "dfs://604851178156619196",
      "name": "document.pdf",
      "status": "ready",
      "size": 123456,
      "content_type": "application/pdf",
      "created_at": 1712000000
    }
  ]
}
```

**状态码：**

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 缺少 `dag_instance_id` 参数 |
| 403 | 无权访问该流程实例 |
| 404 | 流程实例不存在 |

### 2.2 获取单个文件信息

**请求：**

```http
GET /api/automation/v2/dataflow-doc/files/:file_id
Authorization: Bearer {token}
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| file_id | string | 文件 ID，支持纯数字或 `dfs://xxx` 格式 |

**响应：**

```json
{
  "file_id": "604851178156619196",
  "docid": "dfs://604851178156619196",
  "dag_id": "187654321098765432",
  "dag_instance_id": "187654321198765432",
  "name": "document.pdf",
  "status": "ready",
  "size": 123456,
  "content_type": "application/pdf",
  "created_at": 1712000000,
  "updated_at": 1712000100
}
```

**状态码：**

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 404 | 文件不存在 |
| 403 | 无权访问该文件关联的流程实例 |

### 2.3 获取文件下载链接

**请求：**

```http
GET /api/automation/v2/dataflow-doc/files/:file_id/download
Authorization: Bearer {token}
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| file_id | string | 文件 ID，支持纯数字或 `dfs://xxx` 格式 |

**响应：**

```json
{
  "file_id": "604851178156619196",
  "docid": "dfs://604851178156619196",
  "name": "document.pdf",
  "url": "https://oss.example.com/bucket/key?signature=xxx&expires=xxx",
  "size": 123456
}
```

**状态码：**

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 404 | 文件不存在 |
| 403 | 无权访问 |
| 409 | 文件状态非 `ready`，无法下载 |
| 412 | 存储对象不可用 |

---

## 3. 权限校验

### 3.1 校验流程

文件访问权限基于 `dag_instance` 维度，校验流程如下：

```
file_id / dag_instance_id → dag_id → CheckDagAndPerm(dag_id, userInfo, opMap)
```

参考 `ListTaskInstance` 的权限校验实现（`logics/mgnt/mgnt.go:2937-3094`）。

### 3.2 操作权限映射

```go
opMap := &perm.MapOperationProvider{
    OpMap: map[string][]string{
        common.DagTypeDataFlow:      {perm.RunStatisticsOperation},
        common.DagTypeComboOperator: {perm.ViewOperation},
        common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
    },
}
```

### 3.3 各接口权限校验步骤

#### 获取文件列表

1. 校验 `dag_instance_id` 参数存在
2. 调用 `isDagInstanceExist(ctx, map[string]interface{}{"_id": dagInstanceID})` 校验实例存在性
3. 查询 `dag_instance` 获取 `dag_id`
4. 调用 `CheckDagAndPerm(dag_id, userInfo, opMap)` 校验权限
5. 校验 `dag_instance.dagId` 与 `dag_id` 一致

#### 获取单个文件信息

1. 解析 `file_id`（支持纯数字或 `dfs://xxx` 格式）
2. 查询 `t_flow_file` 获取 `dag_instance_id`
3. 通过 `dag_instance_id` 获取 `dag_id`
4. 调用 `CheckDagAndPerm(dag_id, userInfo, opMap)`

#### 获取文件下载链接

1. 执行与"获取单个文件信息"相同的权限校验
2. 额外校验文件状态为 `ready`（`flow_file.status = 2`）
3. 额外校验 `storage_id` 有效，且 `t_flow_storage.status = 1`（normal）

---

## 4. 状态定义

### 4.1 文件状态（`t_flow_file.f_status`）

| 状态值 | 名称 | 说明 | 是否可下载 |
|--------|------|------|------------|
| 1 | pending | 待就绪（上传中/下载中） | 否 |
| 2 | ready | 就绪 | 是 |
| 3 | invalid | 失效（上传失败/下载失败） | 否 |
| 4 | expired | 已过期 | 否 |

### 4.2 存储状态（`t_flow_storage.f_status`）

| 状态值 | 名称 | 说明 |
|--------|------|------|
| 1 | normal | 正常 |
| 2 | pending_delete | 待删除 |
| 3 | deleted | 已删除 |

---

## 5. 核心流程

### 5.1 获取文件列表

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   请求      │────▶│  参数校验   │────▶│  权限校验   │────▶│  查询文件   │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                   │
                                                                   ▼
                                                            ┌─────────────┐
                                                            │  组装响应   │
                                                            └─────────────┘
```

**详细步骤：**

1. 校验 `dag_instance_id` 参数存在
2. 查询 `dag_instance` 获取 `dag_id`，校验存在性
3. 调用 `CheckDagAndPerm(dag_id, userInfo, opMap)` 校验权限
4. 查询 `t_flow_file` 表：`WHERE f_dag_instance_id = ? AND f_status IN (1, 2, 3, 4)`
5. 组装响应（对于每个文件，查询 `t_flow_storage` 获取 `size`、`content_type`）

### 5.2 获取单个文件信息

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   请求      │────▶│  解析file_id│────▶│  查询文件   │────▶│  权限校验   │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                   │
                                                                   ▼
                                                            ┌─────────────┐
                                                            │  组装响应   │
                                                            └─────────────┘
```

**详细步骤：**

1. 解析 `file_id`（调用 `common.NormalizeFileID`）
2. 查询 `t_flow_file` 获取文件信息
3. 通过 `dag_instance_id` 获取 `dag_id`
4. 调用 `CheckDagAndPerm(dag_id, userInfo, opMap)` 校验权限
5. 若 `storage_id > 0`，查询 `t_flow_storage` 获取 `size`、`content_type`
6. 若 `storage_id = 0`，`size` 返回 `0`，`content_type` 返回空字符串
7. 组装响应

### 5.3 获取文件下载链接

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   请求      │────▶│  解析file_id│────▶│  查询文件   │────▶│  权限校验   │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                   │
                                                                   ▼
                        ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
                        │  返回URL    │◀────│ OssGateway  │◀────│  查询存储   │
                        └─────────────┘     └─────────────┘     └─────────────┘
```

**详细步骤：**

1. 执行与"获取单个文件信息"相同的流程
2. 校验文件状态为 `ready`，否则返回 `409`
3. 校验 `storage_id > 0`，否则返回 `412`
4. 查询 `t_flow_storage`，校验 `status = 1`（normal），否则返回 `412`
5. 调用 `OssGateway.GetDownloadURL(oss_id, object_key, expires, internalRequest)`
6. 组装响应

**预签名 URL 有效期：**

```go
// 默认 URL 有效期：15 分钟
const DefaultDownloadURLExpires = 900 // 秒
```

可通过配置项调整有效期。

---

## 6. 代码结构

### 6.1 新增文件

```
driveradapters/dataflow_doc/
├── rest_handler.go          # 现有 - 触发/完成接口
└── file_handler.go          # 新增 - 文件查询/下载接口

logics/mgnt/
├── mgnt.go                  # 现有
└── flow_file.go             # 新增 - 文件服务逻辑
```

### 6.2 接口注册

在 `rest_handler.go` 的 `RegisterAPIv2` 方法中新增：

```go
func (h *restHandler) RegisterAPIv2(engine *gin.RouterGroup) {
    // 现有接口
    engine.POST("/dataflow-doc/trigger/:dagId", middleware.TokenAuth(), h.trigger)
    engine.POST("/dataflow-doc/complete", middleware.TokenAuth(), h.complete)

    // 新增接口
    engine.GET("/dataflow-doc/files", middleware.TokenAuth(), h.listFiles)
    engine.GET("/dataflow-doc/files/:file_id", middleware.TokenAuth(), h.getFile)
    engine.GET("/dataflow-doc/files/:file_id/download", middleware.TokenAuth(), h.downloadFile)
}
```

### 6.3 服务层接口

```go
// FlowFileService 文件服务接口
type FlowFileService interface {
    // ListFiles 按流程实例查询文件列表
    ListFiles(ctx context.Context, dagInstanceID string, userInfo *drivenadapters.UserInfo) ([]*FileInfo, error)

    // GetFile 获取单个文件信息
    GetFile(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*FileInfo, error)

    // GetDownloadURL 获取文件下载链接
    GetDownloadURL(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*DownloadInfo, error)
}

// FileInfo 文件信息
type FileInfo struct {
    FileID        string `json:"file_id"`
    DocID         string `json:"docid"`
    DagID         string `json:"dag_id,omitempty"`
    DagInstanceID string `json:"dag_instance_id,omitempty"`
    Name          string `json:"name"`
    Status        string `json:"status"`
    Size          int64  `json:"size"`
    ContentType   string `json:"content_type"`
    CreatedAt     int64  `json:"created_at"`
    UpdatedAt     int64  `json:"updated_at,omitempty"`
}

// DownloadInfo 下载信息
type DownloadInfo struct {
    FileID string `json:"file_id"`
    DocID  string `json:"docid"`
    Name   string `json:"name"`
    URL    string `json:"url"`
    Size   int64  `json:"size"`
}
```

### 6.4 状态转换函数

数据库中 `f_status` 为整型，需转换为字符串返回：

```go
// FlowFileStatusToString 状态转换
func FlowFileStatusToString(status rds.FlowFileStatus) string {
    switch status {
    case rds.FlowFileStatusPending:
        return "pending"
    case rds.FlowFileStatusReady:
        return "ready"
    case rds.FlowFileStatusInvalid:
        return "invalid"
    case rds.FlowFileStatusExpired:
        return "expired"
    default:
        return "unknown"
    }
}
```

### 6.5 DAO 层扩展

在 `FlowFileQueryOptions` 中添加排序支持：

```go
// FlowFileQueryOptions FlowFile 查询选项
type FlowFileQueryOptions struct {
    // ... 现有字段 ...
    OrderBy     string // 排序字段，如 "created_at"
    Order       string // 排序方向，"asc" 或 "desc"
}
```

修改 `FlowFileDao.List` 方法支持按 `created_at DESC` 排序。

---

## 7. 错误码定义

### 7.1 新增错误码

在 `errors/code.go` 中添加以下错误码：

```go
const (
    // ... 现有错误码 ...

    // 文件相关错误码
    FileNotFound       = "FileNotFound"       // 文件不存在
    FileNotReady       = "FileNotReady"       // 文件状态非 ready，无法下载
    StorageNotReady    = "StorageNotReady"    // 存储对象不可用
    DagInstanceNotFound = "DagInstanceNotFound" // 流程实例不存在
)
```

### 7.2 错误码映射

| 错误码 | HTTP 状态 | 说明 |
|--------|-----------|------|
| InvalidParameter | 400 | 参数错误（缺少 dag_instance_id、file_id 格式错误） |
| DagInstanceNotFound | 404 | 流程实例不存在 |
| FileNotFound | 404 | 文件不存在 |
| Forbidden | 403 | 无权访问 |
| FileNotReady | 409 | 文件状态非 ready，无法下载 |
| StorageNotReady | 412 | 存储对象不可用（storage_id 为空或存储状态异常） |

---

## 8. 验收清单

- [ ] `GET /dataflow-doc/files?dag_instance_id=xxx` 返回该实例关联的所有文件
- [ ] 文件列表按 `created_at` 倒序排列
- [ ] `GET /dataflow-doc/files/:file_id` 返回单个文件信息
- [ ] `file_id` 支持纯数字和 `dfs://xxx` 两种格式
- [ ] `GET /dataflow-doc/files/:file_id/download` 返回有效的预签名下载 URL
- [ ] 无权限访问时返回 `403`
- [ ] 文件不存在时返回 `404`
- [ ] 文件状态为 `pending`/`invalid`/`expired` 时请求下载返回 `409`
- [ ] `storage_id` 为空或存储状态异常时请求下载返回 `412`
- [ ] 响应字段 `docid` 格式为 `dfs://<file_id>`

---

## 9. 失败条件

- 权限校验未基于 `dag_instance` 维度
- 文件信息接口暴露 `oss_id`、`object_key` 等存储细节
- 下载接口直接返回 302 重定向而非预签名 URL
- 未校验文件状态直接返回下载链接
- `file_id` 不支持 `dfs://xxx` 格式解析
- 新增接口未复用现有权限校验逻辑（`CheckDagAndPerm`）

---

## 10. 补充说明

### 10.1 `storage_id = 0` 时的响应处理

当文件尚未落库到 OSS（`storage_id = 0`）时：

| 字段 | 返回值 |
|------|--------|
| `size` | `0` |
| `content_type` | `""`（空字符串） |

文件信息接口仍可正常返回，下载接口会返回 `StorageNotReady` (412) 错误。

### 10.2 批量查询存储信息

文件列表接口需要批量查询 `t_flow_storage`：

```go
// 伪代码
storageIDs := make([]uint64, 0)
for _, file := range files {
    if file.StorageID > 0 {
        storageIDs = append(storageIDs, file.StorageID)
    }
}
storages, _ := flowStorageDao.List(ctx, &rds.FlowStorageQueryOptions{IDs: storageIDs})
// 构建 storageID -> storage 的 map
storageMap := make(map[uint64]*rds.FlowStorage)
for _, s := range storages {
    storageMap[s.ID] = s
}
// 组装响应
for _, file := range files {
    if s, ok := storageMap[file.StorageID]; ok {
        fileInfo.Size = int64(s.Size)
        fileInfo.ContentType = s.ContentType
    }
}
```

---

## 11. 分阶段实现建议

### 阶段一：基础接口

- 实现 `file_handler.go` REST 层
- 实现 `flow_file.go` 服务层
- 注册 API 路由
- 集成现有权限校验逻辑

### 阶段二：测试覆盖

- 单元测试：服务层逻辑
- 集成测试：API 端到端测试
- 边缘案例测试：权限校验、状态校验、存储校验

---

## 12. 实现依赖

本设计依赖以下已有实现：

| 依赖项 | 位置 | 说明 |
|--------|------|------|
| `FlowFileDao` | `store/rds/flow_file.go` | 文件 DAO |
| `FlowStorageDao` | `store/rds/flow_storage.go` | 存储 DAO |
| `OssGateway.GetDownloadURL` | `drivenadapters/ossgateway.go` | 获取预签名 URL |
| `CheckDagAndPerm` | `logics/perm/perm.go` | 权限校验 |
| `NormalizeFileID` | `common/flow_file.go` | 解析 `dfs://xxx` 格式 |
| `isDagInstanceExist` | `logics/mgnt/mgnt.go` | 校验实例存在性 |

---

## 13. 版本历史

| 版本 | 日期 | 作者 | 说明 |
|------|------|------|------|
| v1.0 | 2026-04-01 | Claude | 初版设计 |
