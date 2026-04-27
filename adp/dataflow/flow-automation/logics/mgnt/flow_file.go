package mgnt

import (
	"context"
	"fmt"

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
	dagIns, err := m.mongo.GetDagInstanceByFields(ctx, map[string]interface{}{"_id": dagInstanceID})
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
			FileID:        fmt.Sprintf("%d", f.ID),
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
	dagIns, err := m.mongo.GetDagInstanceByFields(ctx, map[string]interface{}{"_id": file.DagInstanceID})
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
		FileID:        fmt.Sprintf("%d", file.ID),
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
