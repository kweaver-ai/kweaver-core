# Dataflow 文件下载接口实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 Dataflow DFS 文件协议的前端访问接口，支持按流程实例查询文件列表、获取单个文件信息、获取文件下载链接。

**Architecture:** 在现有 `dataflow_doc` 模块扩展三个 REST 接口，服务层新增 `FlowFileService` 处理文件查询与下载逻辑，复用现有权限校验 `CheckDagAndPerm` 和 `OssGateway.GetDownloadURL`。

**Tech Stack:** Go, Gin, MongoDB (dag_instance), MySQL (flow_file, flow_storage), OssGateway

---

## 文件结构

```
driveradapters/dataflow_doc/
├── rest_handler.go              # 修改：注册新路由
└── file_handler.go              # 新增：文件接口 REST Handler

logics/mgnt/
├── mgnt.go                      # 修改：接口定义新增方法
└── flow_file.go                 # 新增：文件服务逻辑实现

errors/
└── code.go                      # 修改：新增错误码

pkg/rds/
└── types.go                     # 修改：FlowFileQueryOptions 添加排序字段

store/rds/
└── flow_file.go                 # 修改：List 方法支持排序
```

---

## Task 1: 新增错误码

**Files:**
- Modify: `errors/code.go`

- [ ] **Step 1: 新增文件相关错误码**

在 `errors/code.go` 的主错误码区域添加：

```go
var (
    // ... 现有错误码 ...

    // 文件相关错误码
    FileNotReady    = "FileNotReady"
    StorageNotReady = "StorageNotReady"
    // 注：DagInsNotFound 已存在，复用该错误码
)
```

- [ ] **Step 2: 新增错误码消息映射**

在 `ErrorsMsg` 映射中添加：

```go
FileNotReady: {
    Languages[0]: {"文件未就绪", "文件正在上传或下载中，暂时无法下载"},
    Languages[1]: {"文件未就緒", "文件正在上傳或下載中，暫時無法下載"},
    Languages[2]: {"File not ready", "File is uploading or downloading, cannot download now"},
},
StorageNotReady: {
    Languages[0]: {"存储对象不可用", "文件存储异常，请联系管理员"]},
    Languages[1]: {"存儲對象不可用", "文件存儲異常，請聯繫管理員"],
    Languages[2]: {"Storage not ready", "File storage is abnormal, please contact administrator"],
},
// 注：DagInsNotFound 已存在，无需新增
```

- [ ] **Step 3: 新增 HTTP 状态码映射**

在 `ErrorsHttpCode` 映射中添加：

```go
var ErrorsHttpCode = map[string]int{
    // ... 现有映射 ...
    FileNotReady:    http.StatusConflict,           // 409
    StorageNotReady: http.StatusPreconditionFailed, // 412
}
```

**注**：`DagInsNotFound` 已在现有代码中映射为 404。

- [ ] **Step 4: 提交**

```bash
git add errors/code.go
git commit -m "feat(#437): add error codes for file download API

Add FileNotReady, StorageNotReady, DagInstanceNotFound error codes
with i18n messages and HTTP status code mappings.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 扩展 DAO 层排序支持

**Files:**
- Modify: `pkg/rds/types.go`
- Modify: `store/rds/flow_file.go`

- [ ] **Step 1: 扩展 FlowFileQueryOptions**

在 `pkg/rds/types.go` 的 `FlowFileQueryOptions` 结构体中添加：

```go
// FlowFileQueryOptions FlowFile 查询选项
type FlowFileQueryOptions struct {
    ID            *uint64
    IDs           []uint64
    DagID         string
    DagInstanceID string
    StorageID     *uint64
    Status        *FlowFileStatus
    Statuses      []FlowFileStatus
    ExpiresBefore int64 // 过期时间早于此值
    Limit         int
    Offset        int
    OrderBy       string // 排序字段，如 "created_at"
    Order         string // 排序方向，"asc" 或 "desc"
}
```

- [ ] **Step 2: 修改 FlowFileDao.List 方法**

**首先，在文件顶部添加 `"strings"` 导入。**

然后在 `store/rds/flow_file.go` 的 `List` 方法中添加排序支持：

找到以下代码段（约第 128 行）：

```go
sqlStr := fmt.Sprintf("SELECT f_id, f_dag_id, f_dag_instance_id, f_storage_id, f_status, f_name, f_expires_at, f_created_at, f_updated_at FROM t_flow_file%s ORDER BY f_id", where)
```

替换为：

```go
// 构建排序子句
orderBy := "f_id"
orderDir := "ASC"
if opts.OrderBy != "" {
    // 映射字段名，防止 SQL 注入
    fieldMap := map[string]string{
        "id":         "f_id",
        "created_at": "f_created_at",
        "updated_at": "f_updated_at",
        "status":     "f_status",
    }
    if dbField, ok := fieldMap[opts.OrderBy]; ok {
        orderBy = dbField
    }
}
if opts.Order != "" && (strings.ToUpper(opts.Order) == "DESC" || strings.ToUpper(opts.Order) == "ASC") {
    orderDir = strings.ToUpper(opts.Order)
}

sqlStr := fmt.Sprintf("SELECT f_id, f_dag_id, f_dag_instance_id, f_storage_id, f_status, f_name, f_expires_at, f_created_at, f_updated_at FROM t_flow_file%s ORDER BY %s %s", where, orderBy, orderDir)
```

- [ ] **Step 3: 提交**

```bash
git add pkg/rds/types.go store/rds/flow_file.go
git commit -m "feat(#437): add OrderBy and Order to FlowFileQueryOptions

Support sorting by created_at, updated_at, status fields in file list query.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 新增服务层接口定义

**Files:**
- Modify: `logics/mgnt/mgnt.go`

- [ ] **Step 1: 在 MgntHandler 接口新增方法**

在 `logics/mgnt/mgnt.go` 的 `MgntHandler` 接口末尾（约第 367 行后）添加：

```go
// Dataflow 文件访问接口
ListFlowFiles(ctx context.Context, dagInstanceID string, userInfo *drivenadapters.UserInfo) ([]*FlowFileInfo, error)
GetFlowFile(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*FlowFileInfo, error)
GetFlowFileDownloadURL(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*FlowFileDownloadInfo, error)
```

- [ ] **Step 2: 新增响应结构体**

在 `logics/mgnt/mgnt.go` 的结构体定义区域（约第 290 行后）添加：

```go
// FlowFileInfo 文件信息
type FlowFileInfo struct {
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

// FlowFileDownloadInfo 文件下载信息
type FlowFileDownloadInfo struct {
    FileID string `json:"file_id"`
    DocID  string `json:"docid"`
    Name   string `json:"name"`
    URL    string `json:"url"`
    Size   int64  `json:"size"`
}
```

- [ ] **Step 3: 提交**

```bash
git add logics/mgnt/mgnt.go
git commit -m "feat(#437): add FlowFile service interface definitions

Add ListFlowFiles, GetFlowFile, GetFlowFileDownloadURL methods to MgntHandler.
Add FlowFileInfo and FlowFileDownloadInfo response structs.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 实现服务层逻辑

**Files:**
- Create: `logics/mgnt/flow_file.go`

- [ ] **Step 1: 创建 flow_file.go 文件**

创建 `logics/mgnt/flow_file.go`，实现文件服务逻辑：

```go
package mgnt

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// DefaultDownloadURLExpires 预签名 URL 默认有效期（秒）
	DefaultDownloadURLExpires = 900 // 15 分钟
)

// flowFileStatusToString 状态转换为字符串
func flowFileStatusToString(status rds.FlowFileStatus) string {
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

// ListFlowFiles 按流程实例查询文件列表
func (m *mgnt) ListFlowFiles(ctx context.Context, dagInstanceID string, userInfo *drivenadapters.UserInfo) ([]*FlowFileInfo, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 1. 校验 dag_instance 存在性并获取 dag_id
	dagIns, err := m.mongo.GetDagInstanceByFields(ctx, map[string]interface{}{"_id": dagInstanceID}, "dagId")
	if err != nil {
		log.Warnf("[ListFlowFiles] GetDagInstanceByFields err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if dagIns == nil {
		return nil, ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]interface{}{"dag_instance_id": dagInstanceID})
	}

	dagID := dagIns.DagID

	// 2. 权限校验
	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.RunStatisticsOperation},
			common.DagTypeComboOperator: {perm.ViewOperation},
			common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
		},
	}
	if userInfo.AccountType == common.APP.ToString() {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldAppTokenOperation}
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return nil, err
	}

	// 3. 查询文件列表
	files, err := m.flowFileDao.List(ctx, &rds.FlowFileQueryOptions{
		DagInstanceID: dagInstanceID,
		OrderBy:       "created_at",
		Order:         "desc",
	})
	if err != nil {
		log.Warnf("[ListFlowFiles] List flow files err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 4. 批量查询存储信息
	result := make([]*FlowFileInfo, 0, len(files))
	storageIDs := make([]uint64, 0)
	for _, f := range files {
		if f.StorageID > 0 {
			storageIDs = append(storageIDs, f.StorageID)
		}
	}

	storageMap := make(map[uint64]*rds.FlowStorage)
	if len(storageIDs) > 0 {
		storages, err := m.flowStorageDao.List(ctx, &rds.FlowStorageQueryOptions{IDs: storageIDs})
		if err != nil {
			log.Warnf("[ListFlowFiles] List storages err: %s", err.Error())
		}
		for _, s := range storages {
			storageMap[s.ID] = s
		}
	}

	// 5. 组装响应
	for _, f := range files {
		info := &FlowFileInfo{
			FileID:        string(f.ID),
			DocID:         common.BuildDFSURI(f.ID),
			DagInstanceID: f.DagInstanceID,
			Name:          f.Name,
			Status:        flowFileStatusToString(f.Status),
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
		}

		if s, ok := storageMap[f.StorageID]; ok {
			info.Size = int64(s.Size)
			info.ContentType = s.ContentType
		}

		result = append(result, info)
	}

	return result, nil
}

// GetFlowFile 获取单个文件信息
func (m *mgnt) GetFlowFile(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*FlowFileInfo, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 1. 解析 file_id
	id, err := common.NormalizeFileID(fileID)
	if err != nil {
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"file_id": fileID})
	}

	// 2. 查询文件
	file, err := m.flowFileDao.GetByID(ctx, id)
	if err != nil {
		log.Warnf("[GetFlowFile] GetByID err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if file == nil {
		return nil, ierrors.NewIError(ierrors.FileNotFound, "", map[string]interface{}{"file_id": fileID})
	}

	// 3. 校验 dag_instance 存在性
	err = m.isDagInstanceExist(ctx, map[string]interface{}{"_id": file.DagInstanceID})
	if err != nil {
		return nil, err
	}

	// 4. 获取 dag_id 并校验权限
	dagIns, err := m.mongo.GetDagInstanceByFields(ctx, map[string]interface{}{"_id": file.DagInstanceID}, "dagId")
	if err != nil {
		log.Warnf("[GetFlowFile] GetDagInstanceByFields err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.RunStatisticsOperation},
			common.DagTypeComboOperator: {perm.ViewOperation},
			common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
		},
	}
	if userInfo.AccountType == common.APP.ToString() {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldAppTokenOperation}
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagIns.DagID, userInfo, opMap)
	if err != nil {
		return nil, err
	}

	// 5. 查询存储信息
	var size int64 = 0
	var contentType string
	if file.StorageID > 0 {
		storage, err := m.flowStorageDao.GetByID(ctx, file.StorageID)
		if err != nil {
			log.Warnf("[GetFlowFile] Get storage err: %s", err.Error())
		}
		if storage != nil {
			size = int64(storage.Size)
			contentType = storage.ContentType
		}
	}

	// 6. 组装响应
	return &FlowFileInfo{
		FileID:        string(file.ID),
		DocID:         common.BuildDFSURI(file.ID),
		DagID:         file.DagID,
		DagInstanceID: file.DagInstanceID,
		Name:          file.Name,
		Status:        flowFileStatusToString(file.Status),
		Size:          size,
		ContentType:   contentType,
		CreatedAt:     file.CreatedAt,
		UpdatedAt:     file.UpdatedAt,
	}, nil
}

// GetFlowFileDownloadURL 获取文件下载链接
func (m *mgnt) GetFlowFileDownloadURL(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*FlowFileDownloadInfo, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(ctx, attribute.String("file_id", fileID))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 1. 获取文件信息（包含权限校验）
	fileInfo, err := m.GetFlowFile(ctx, fileID, userInfo)
	if err != nil {
		return nil, err
	}

	// 2. 解析 file_id 获取原始 ID
	id, _ := common.NormalizeFileID(fileID)

	// 3. 查询文件记录（需要 storage_id）
	file, err := m.flowFileDao.GetByID(ctx, id)
	if err != nil {
		log.Warnf("[GetFlowFileDownloadURL] GetByID err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 4. 校验文件状态
	if file.Status != rds.FlowFileStatusReady {
		return nil, ierrors.NewIError(ierrors.FileNotReady, "", map[string]interface{}{
			"file_id": fileID,
			"status":  flowFileStatusToString(file.Status),
		})
	}

	// 5. 校验 storage_id
	if file.StorageID == 0 {
		return nil, ierrors.NewIError(ierrors.StorageNotReady, "", map[string]interface{}{"file_id": fileID})
	}

	// 6. 查询存储信息
	storage, err := m.flowStorageDao.GetByID(ctx, file.StorageID)
	if err != nil {
		log.Warnf("[GetFlowFileDownloadURL] Get storage err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if storage == nil || storage.Status != rds.FlowStorageStatusNormal {
		return nil, ierrors.NewIError(ierrors.StorageNotReady, "", map[string]interface{}{"file_id": fileID})
	}

	// 7. 获取预签名下载 URL
	url, err := m.ossGateway.GetDownloadURL(ctx, storage.OssID, storage.ObjectKey, DefaultDownloadURLExpires, false)
	if err != nil {
		log.Warnf("[GetFlowFileDownloadURL] GetDownloadURL err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 8. 组装响应
	return &FlowFileDownloadInfo{
		FileID: fileInfo.FileID,
		DocID:  fileInfo.DocID,
		Name:   fileInfo.Name,
		URL:    url,
		Size:   fileInfo.Size,
	}, nil
}
```

- [ ] **Step 2: 在 mgnt 结构体中添加依赖**

在 `logics/mgnt/mgnt.go` 的 `mgnt` 结构体定义中添加字段（约第 377 行后）：

找到 `type mgnt struct` 结构体，添加：

```go
type mgnt struct {
    // ... 现有字段 ...
    flowFileDao     rds.FlowFileDao
    flowStorageDao  rds.FlowStorageDao
    // 注：ossGateway 已存在，无需添加
}
```

在 `NewMgnt()` 函数中初始化（约第 395 行后）：

```go
func NewMgnt() MgntHandler {
    mOnce.Do(func() {
        m = &mgnt{
            // ... 现有字段 ...
            flowFileDao:    rds.GetFlowFileDao(),    // 使用单例模式
            flowStorageDao: rds.GetFlowStorageDao(), // 使用单例模式
            // 注：ossGateway 已在现有初始化中
        }
    })
    return m
}
```

**重要**：使用 `rds.GetFlowFileDao()` 和 `rds.GetFlowStorageDao()` 单例模式，与现有代码风格保持一致。

- [ ] **Step 3: 提交**

```bash
git add logics/mgnt/flow_file.go logics/mgnt/mgnt.go
git commit -m "feat(#437): implement FlowFile service logic

Add ListFlowFiles, GetFlowFile, GetFlowFileDownloadURL methods.
- Permission check based on dag_instance dimension
- File status and storage validation for download
- Pre-signed URL generation via OssGateway

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 实现 REST Handler

**Files:**
- Create: `driveradapters/dataflow_doc/file_handler.go`

- [ ] **Step 1: 创建 file_handler.go 文件**

创建 `driveradapters/dataflow_doc/file_handler.go`：

```go
package dataflow_doc

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/mgnt"
)

// listFiles handles GET /api/automation/v2/dataflow-doc/files
func (h *restHandler) listFiles(c *gin.Context) {
	dagInstanceID := c.Query("dag_instance_id")
	if dagInstanceID == "" {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"dag_instance_id": "required"}))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	files, err := h.mgnt.ListFlowFiles(c.Request.Context(), dagInstanceID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// getFile handles GET /api/automation/v2/dataflow-doc/files/:file_id
func (h *restHandler) getFile(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"file_id": "required"}))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	file, err := h.mgnt.GetFlowFile(c.Request.Context(), fileID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, file)
}

// downloadFile handles GET /api/automation/v2/dataflow-doc/files/:file_id/download
func (h *restHandler) downloadFile(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"file_id": "required"}))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	downloadInfo, err := h.mgnt.GetFlowFileDownloadURL(c.Request.Context(), fileID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, downloadInfo)
}
```

- [ ] **Step 2: 注册路由**

在 `driveradapters/dataflow_doc/rest_handler.go` 的 `RegisterAPIv2` 方法中添加路由：

```go
// RegisterAPIv2 registers v2 version APIs
func (h *restHandler) RegisterAPIv2(engine *gin.RouterGroup) {
	engine.POST("/dataflow-doc/trigger/:dagId", middleware.TokenAuth(), h.trigger)
	engine.POST("/dataflow-doc/complete", middleware.TokenAuth(), h.complete)

	// 文件访问接口
	engine.GET("/dataflow-doc/files", middleware.TokenAuth(), h.listFiles)
	engine.GET("/dataflow-doc/files/:file_id", middleware.TokenAuth(), h.getFile)
	engine.GET("/dataflow-doc/files/:file_id/download", middleware.TokenAuth(), h.downloadFile)
}
```

- [ ] **Step 3: 提交**

```bash
git add driveradapters/dataflow_doc/file_handler.go driveradapters/dataflow_doc/rest_handler.go
git commit -m "feat(#437): add REST handlers for file access API

Add three endpoints:
- GET /dataflow-doc/files?dag_instance_id=xxx
- GET /dataflow-doc/files/:file_id
- GET /dataflow-doc/files/:file_id/download

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 编译验证

**Files:**
- None

- [ ] **Step 1: 编译项目**

```bash
cd /home/zhang/kweaver/adp/dataflow/flow-automation
go build ./...
```

预期输出：编译成功，无错误。

- [ ] **Step 2: 检查未使用导入**

```bash
go vet ./...
```

预期输出：无警告。

- [ ] **Step 3: 提交（如有修改）**

```bash
git add -A
git commit -m "fix(#437): fix compilation issues

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: 更新 Mock 文件

**Files:**
- Modify: `tests/mock_logics/mgnt_mock.go`

- [ ] **Step 1: 重新生成 mock 文件**

```bash
cd /home/zhang/kweaver/adp/dataflow/flow-automation
go generate ./logics/mgnt/...
```

或者手动运行 mockgen：

```bash
mockgen -package mock_logics -source logics/mgnt/mgnt.go -destination tests/mock_logics/mgnt_mock.go
```

- [ ] **Step 2: 提交**

```bash
git add tests/mock_logics/mgnt_mock.go
git commit -m "chore(#437): update mock for MgntHandler interface

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: 集成测试

**Files:**
- Test manually via API calls

- [ ] **Step 1: 启动服务**

启动本地服务（根据项目启动方式）。

- [ ] **Step 2: 测试文件列表接口**

```bash
curl -X GET "http://localhost:8080/api/automation/v2/dataflow-doc/files?dag_instance_id=<dag_instance_id>" \
  -H "Authorization: Bearer <token>"
```

预期：返回文件列表或权限错误。

- [ ] **Step 3: 测试单个文件接口**

```bash
curl -X GET "http://localhost:8080/api/automation/v2/dataflow-doc/files/<file_id>" \
  -H "Authorization: Bearer <token>"
```

预期：返回文件信息或 404 错误。

- [ ] **Step 4: 测试下载接口**

```bash
curl -X GET "http://localhost:8080/api/automation/v2/dataflow-doc/files/<file_id>/download" \
  -H "Authorization: Bearer <token>"
```

预期：返回预签名 URL 或错误（文件未就绪/存储不可用）。

---

## 验收清单

- [ ] `GET /dataflow-doc/files?dag_instance_id=xxx` 返回文件列表
- [ ] 文件列表按 `created_at` 倒序排列
- [ ] `GET /dataflow-doc/files/:file_id` 返回单个文件信息
- [ ] `file_id` 支持纯数字和 `dfs://xxx` 两种格式
- [ ] `GET /dataflow-doc/files/:file_id/download` 返回预签名 URL
- [ ] 无权限访问时返回 `403`
- [ ] 文件不存在时返回 `404`
- [ ] 文件状态非 `ready` 时请求下载返回 `409`
- [ ] `storage_id` 为空时请求下载返回 `412`
- [ ] 响应字段 `docid` 格式为 `dfs://<file_id>`
- [ ] Mock 文件已更新并编译通过
- [ ] 错误码复用 `DagInsNotFound` 而非新增
- [ ] DAO 初始化使用单例模式 `rds.GetFlowFileDao()`

---

## 实现依赖

| 依赖项 | 位置 | 说明 |
|--------|------|------|
| `FlowFileDao` | `store/rds/flow_file.go` | 文件 DAO |
| `FlowStorageDao` | `store/rds/flow_storage.go` | 存储 DAO |
| `OssGateWay.GetDownloadURL` | `drivenadapters/ossgateway.go` | 获取预签名 URL |
| `CheckDagAndPerm` | `logics/perm/perm.go` | 权限校验 |
| `NormalizeFileID` | `common/flow_file.go` | 解析 `dfs://xxx` 格式 |
| `BuildDFSURI` | `common/flow_file.go` | 生成 `dfs://xxx` 格式 |
| `isDagInstanceExist` | `logics/mgnt/mgnt.go` | 校验实例存在性 |
