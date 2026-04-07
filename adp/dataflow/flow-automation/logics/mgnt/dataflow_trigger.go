// Package mgnt Dataflow 文件子系统触发器逻辑
package mgnt

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
	normalizeutil "github.com/kweaver-ai/adp/autoflow/flow-automation/utils/normalize"
)

// TriggerDataflowDocParams 触发Dataflow文档处理的参数
type TriggerDataflowDocParams struct {
	DagID       string                 `json:"dag_id"`       // 流程定义ID
	SourceFrom  string                 `json:"source_from"`  // 文件来源: form, local, remote
	Name        string                 `json:"name"`         // 文件名
	Size        int64                  `json:"size"`         // 文件大小
	ContentType string                 `json:"content_type"` // MIME类型
	URL         string                 `json:"url"`          // 源文件URL(仅remote模式)
	Data        map[string]interface{} `json:"data"`         // 触发器扩展字段
	File        io.ReadCloser          `json:"-"`            // 上传的文件(仅form模式)，调用者负责关闭
}

// TriggerDataflowDocResult 触发Dataflow文档处理的结果
type TriggerDataflowDocResult struct {
	DagID         string                        `json:"dag_id"`
	DagInstanceID string                        `json:"dag_instance_id"`
	FileID        string                        `json:"file_id"`
	DocID         string                        `json:"docid"`  // dfs://<file_id>
	Status        string                        `json:"status"` // ready, pending, processing
	Name          string                        `json:"name"`
	Size          int64                         `json:"size"`
	UploadReq     *drivenadapters.UploadRequest `json:"upload_req,omitempty"`
}

// CompleteDataflowDocUploadParams 完成上传的参数
type CompleteDataflowDocUploadParams struct {
	FileID string `json:"file_id"` // 支持纯ID或dfs://格式
	Etag   string `json:"etag"`
	Size   int64  `json:"size"`
}

// CompleteDataflowDocUploadResult 完成上传的结果
type CompleteDataflowDocUploadResult struct {
	FileID    string `json:"file_id"`
	DocID     string `json:"docid"`
	Status    string `json:"status"`
	Continued bool   `json:"continued"`
}

// TriggerDataflowDoc 触发Dataflow文档处理
// 支持三种来源：form(表单直接上传), local(先触发后上传), remote(URL下载触发)
func (m *mgnt) TriggerDataflowDoc(ctx context.Context, params *TriggerDataflowDocParams, userInfo *drivenadapters.UserInfo) (*TriggerDataflowDocResult, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 1. 验证DAG存在且可访问
	dag, err := m.mongo.GetDag(ctx, params.DagID)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return nil, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": params.DagID})
		}
		log.Warnf("[logic.TriggerDataflowDoc] GetDag err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	if dag.Type != common.DagTypeDataFlow {
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]string{"message": "dag type is not dataflow"})
	}

	// 2. 验证触发器类型
	if dag.Steps[0].Operator != common.DataflowDocTrigger {
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]string{"message": "must be dataflow_doc_trigger"})
	}

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow: {perm.ManualExecOperation},
			common.DagTypeDefault:  {perm.OldAdminOperation, perm.OldShareOperation},
		},
	}

	if userInfo.AccountType == common.APP.ToString() {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldAppTokenOperation}
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dag.ID, userInfo, opMap)
	if err != nil {
		return nil, err
	}

	// 3. 获取用户详情
	userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
	if err != nil {
		log.Warnf("[logic.TriggerDataflowDoc] GetUserInfoByType err, detail: %s", err.Error())
		return nil, err
	}

	// 4. 根据来源处理
	var result *TriggerDataflowDocResult
	switch params.SourceFrom {
	case "form":
		result, err = m.triggerFormUpload(ctx, params, dag, userInfo, userDetail)
	case "local":
		result, err = m.triggerLocalUpload(ctx, params, dag, userInfo, userDetail)
	case "remote":
		result, err = m.triggerRemoteDownload(ctx, params, dag, userInfo, userDetail)
	default:
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", []interface{}{fmt.Sprintf("invalid source_from: %s", params.SourceFrom)})
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// triggerFormUpload 处理表单直接上传来源
func (m *mgnt) triggerFormUpload(ctx context.Context, params *TriggerDataflowDocParams, dag *entity.Dag, userInfo *drivenadapters.UserInfo, userDetail drivenadapters.UserInfo) (*TriggerDataflowDocResult, error) {
	log := traceLog.WithContext(ctx)

	// 1. 上传文件到OssGateway
	ossID, err := m.ossGateway.GetAvaildOSS(ctx)
	if err != nil {
		log.Warnf("[triggerFormUpload] GetAvaildOSS err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", map[string]string{"message": "no available OSS"})
	}

	// 2. 创建flow_file记录(先创建以获取file_id)
	now := time.Now().Unix()
	flowFile := &rds.FlowFile{
		ID:        store.NextID(),
		DagID:     params.DagID,
		Status:    rds.FlowFileStatusPending, // 先设为pending，上传成功后改为ready
		Name:      params.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := rds.GetFlowFileDao().Insert(ctx, flowFile); err != nil {
		log.Warnf("[triggerFormUpload] Insert flow_file err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 3. 生成对象存储路径
	objectKey, err := common.BuildFlowFileObjectKey(flowFile.ID, params.Name)
	if err != nil {
		// 清理flow_file
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", []interface{}{"invalid filename"})
	}

	// 4. 上传文件
	if err := m.ossGateway.UploadFile(ctx, ossID, objectKey, false, params.File, params.Size); err != nil {
		log.Warnf("[triggerFormUpload] UploadFile err: %s", err.Error())
		// 标记flow_file为invalid
		rds.GetFlowFileDao().UpdateStatus(ctx, flowFile.ID, rds.FlowFileStatusInvalid)
		return nil, ierrors.NewIError(ierrors.InternalError, "", map[string]string{"message": "failed to upload file"})
	}

	// 5. 创建flow_storage记录
	flowStorage := &rds.FlowStorage{
		ID:          store.NextID(),
		OssID:       ossID,
		ObjectKey:   objectKey,
		Name:        params.Name,
		ContentType: params.ContentType,
		Size:        uint64(params.Size),
		Status:      rds.FlowStorageStatusNormal,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := rds.GetFlowStorageDao().Insert(ctx, flowStorage); err != nil {
		log.Warnf("[triggerFormUpload] Insert flow_storage err: %s", err.Error())
		// 尝试删除已上传的文件
		m.ossGateway.DeleteFile(ctx, ossID, objectKey, false)
		rds.GetFlowFileDao().UpdateStatus(ctx, flowFile.ID, rds.FlowFileStatusInvalid)
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 6. 更新flow_file状态为ready
	readyStatus := rds.FlowFileStatusReady
	if err := rds.GetFlowFileDao().Update(ctx, flowFile.ID, &rds.FlowFileUpdateParams{
		StorageID: &flowStorage.ID,
		Status:    &readyStatus,
	}); err != nil {
		log.Warnf("[triggerFormUpload] Update flow_file err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	flowFile.StorageID = flowStorage.ID
	flowFile.Status = readyStatus

	// 7. 构建DagInstanceVars
	dfsURI := common.BuildDFSURI(flowFile.ID)
	runVar := m.buildDataflowDocRunVar(dfsURI, "", "form", userDetail, userInfo)

	// 解析扩展字段
	if params.Data != nil {
		if fields, ok := normalizeutil.AsSlice(dag.Steps[0].Parameters["fields"]); ok {
			if err := ParseFields(ctx, fields, params.Data, runVar, ErrTypeV1).BuildError(); err != nil {
				// 标记flow_file为invalid
				rds.GetFlowFileDao().UpdateStatus(ctx, flowFile.ID, rds.FlowFileStatusInvalid)
				return nil, err
			}
		}
	}

	// 8. 创建并运行dag_instance
	dag.SetPushMessage(m.executeMethods.Publish)
	dagIns, dagErr := dag.Run(ctx, entity.TriggerDocument, runVar, entity.WithKeyWords(params.Name, userDetail.UserName))
	if dagErr != nil {
		rds.GetFlowFileDao().UpdateStatus(ctx, flowFile.ID, rds.FlowFileStatusInvalid)
		return nil, ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id": dag.ID, "status": dag.Status})
	}

	if _, err := m.mongo.CreateDagIns(ctx, dagIns); err != nil {
		log.Warnf("[triggerFormUpload] CreateDagIns err: %s", err.Error())
		rds.GetFlowFileDao().UpdateStatus(ctx, flowFile.ID, rds.FlowFileStatusInvalid)
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 更新flow_file的dag_instance_id
	if err := rds.GetFlowFileDao().Update(ctx, flowFile.ID, &rds.FlowFileUpdateParams{
		DagInstanceID: &dagIns.ID,
	}); err != nil {
		log.Warnf("[triggerFormUpload] Update dag_instance_id err: %s", err.Error())
	}

	return &TriggerDataflowDocResult{
		DagID:         params.DagID,
		DagInstanceID: dagIns.ID,
		FileID:        fmt.Sprintf("%d", flowFile.ID),
		DocID:         dfsURI,
		Status:        "ready",
		Name:          params.Name,
		Size:          params.Size,
	}, nil
}

// triggerLocalUpload 处理先触发后上传来源
func (m *mgnt) triggerLocalUpload(ctx context.Context, params *TriggerDataflowDocParams, dag *entity.Dag, userInfo *drivenadapters.UserInfo, userDetail drivenadapters.UserInfo) (*TriggerDataflowDocResult, error) {
	log := traceLog.WithContext(ctx)

	// 1. 获取可用OSS
	ossID, err := m.ossGateway.GetAvaildOSS(ctx)
	if err != nil {
		log.Warnf("[triggerLocalUpload] GetAvaildOSS err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", map[string]string{"message": "no available OSS"})
	}

	// 2. 创建flow_file记录
	now := time.Now().Unix()
	flowFile := &rds.FlowFile{
		ID:        store.NextID(),
		DagID:     params.DagID,
		Status:    rds.FlowFileStatusPending, // 等待上传
		Name:      params.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := rds.GetFlowFileDao().Insert(ctx, flowFile); err != nil {
		log.Warnf("[triggerLocalUpload] Insert flow_file err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 3. 生成对象存储路径和预签名上传URL
	objectKey, err := common.BuildFlowFileObjectKey(flowFile.ID, params.Name)
	if err != nil {
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", []interface{}{"invalid filename"})
	}

	// 获取预签名上传URL (有效期1小时)
	uploadReq, err := m.ossGateway.GetUploadReq(ctx, ossID, objectKey, 3600, false)
	if err != nil {
		log.Warnf("[triggerLocalUpload] GetUploadReq err: %s", err.Error())
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.InternalError, "", map[string]string{"message": "failed to get upload URL"})
	}

	// 4. 构建DagInstanceVars
	dfsURI := common.BuildDFSURI(flowFile.ID)
	runVar := m.buildDataflowDocRunVar(dfsURI, "", "local", userDetail, userInfo)

	// 解析扩展字段
	if params.Data != nil {
		if fields, ok := normalizeutil.AsSlice(dag.Steps[0].Parameters["fields"]); ok {
			if err := ParseFields(ctx, fields, params.Data, runVar, ErrTypeV1).BuildError(); err != nil {
				rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
				return nil, err
			}
		}
	}

	// 5. 创建并运行dag_instance
	dag.SetPushMessage(m.executeMethods.Publish)
	dagIns, dagErr := dag.Run(ctx, entity.TriggerDocument, runVar, entity.WithKeyWords(params.Name, userDetail.UserName))
	if dagErr != nil {
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id": dag.ID, "status": dag.Status})
	}

	if _, err := m.mongo.CreateDagIns(ctx, dagIns); err != nil {
		log.Warnf("[triggerLocalUpload] CreateDagIns err: %s", err.Error())
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 更新flow_file的dag_instance_id
	if err := rds.GetFlowFileDao().Update(ctx, flowFile.ID, &rds.FlowFileUpdateParams{
		DagInstanceID: &dagIns.ID,
	}); err != nil {
		log.Warnf("[triggerLocalUpload] Update dag_instance_id err: %s", err.Error())
	}

	return &TriggerDataflowDocResult{
		DagID:         params.DagID,
		DagInstanceID: dagIns.ID,
		FileID:        fmt.Sprintf("%d", flowFile.ID),
		DocID:         dfsURI,
		Status:        "pending",
		Name:          params.Name,
		Size:          params.Size,
		UploadReq:     uploadReq,
	}, nil
}

// triggerRemoteDownload 处理URL下载触发来源
func (m *mgnt) triggerRemoteDownload(ctx context.Context, params *TriggerDataflowDocParams, dag *entity.Dag, userInfo *drivenadapters.UserInfo, userDetail drivenadapters.UserInfo) (*TriggerDataflowDocResult, error) {
	log := traceLog.WithContext(ctx)

	// 1. 创建flow_file记录
	now := time.Now().Unix()
	flowFile := &rds.FlowFile{
		ID:        store.NextID(),
		DagID:     params.DagID,
		Status:    rds.FlowFileStatusPending, // 等待下载
		Name:      params.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := rds.GetFlowFileDao().Insert(ctx, flowFile); err != nil {
		log.Warnf("[triggerRemoteDownload] Insert flow_file err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 2. 创建下载任务
	downloadJob := &rds.FlowFileDownloadJob{
		ID:          store.NextID(),
		FileID:      flowFile.ID,
		Status:      rds.FlowFileDownloadJobStatusPending,
		MaxRetry:    3,
		DownloadURL: params.URL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := rds.GetFlowFileDownloadJobDao().Insert(ctx, downloadJob); err != nil {
		log.Warnf("[triggerRemoteDownload] Insert download_job err: %s", err.Error())
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 3. 构建DagInstanceVars
	dfsURI := common.BuildDFSURI(flowFile.ID)
	runVar := m.buildDataflowDocRunVar(dfsURI, params.URL, "remote", userDetail, userInfo)

	// 解析扩展字段
	if params.Data != nil {
		if fields, ok := normalizeutil.AsSlice(dag.Steps[0].Parameters["fields"]); ok {
			if err := ParseFields(ctx, fields, params.Data, runVar, ErrTypeV1).BuildError(); err != nil {
				rds.GetFlowFileDownloadJobDao().Delete(ctx, downloadJob.ID)
				rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
				return nil, err
			}
		}
	}

	// 4. 创建并运行dag_instance
	dag.SetPushMessage(m.executeMethods.Publish)
	dagIns, dagErr := dag.Run(ctx, entity.TriggerDocument, runVar, entity.WithKeyWords(params.Name, userDetail.UserName))
	if dagErr != nil {
		rds.GetFlowFileDownloadJobDao().Delete(ctx, downloadJob.ID)
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id": dag.ID, "status": dag.Status})
	}

	if _, err := m.mongo.CreateDagIns(ctx, dagIns); err != nil {
		log.Warnf("[triggerRemoteDownload] CreateDagIns err: %s", err.Error())
		rds.GetFlowFileDownloadJobDao().Delete(ctx, downloadJob.ID)
		rds.GetFlowFileDao().Delete(ctx, flowFile.ID)
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 更新flow_file的dag_instance_id
	if err := rds.GetFlowFileDao().Update(ctx, flowFile.ID, &rds.FlowFileUpdateParams{
		DagInstanceID: &dagIns.ID,
	}); err != nil {
		log.Warnf("[triggerRemoteDownload] Update dag_instance_id err: %s", err.Error())
	}

	return &TriggerDataflowDocResult{
		DagID:         params.DagID,
		DagInstanceID: dagIns.ID,
		FileID:        fmt.Sprintf("%d", flowFile.ID),
		DocID:         dfsURI,
		Status:        "processing",
		Name:          params.Name,
		Size:          params.Size,
	}, nil
}

// CompleteDataflowDocUpload 完成上传（local来源）
func (m *mgnt) CompleteDataflowDocUpload(ctx context.Context, params *CompleteDataflowDocUploadParams, userInfo *drivenadapters.UserInfo) (*CompleteDataflowDocUploadResult, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 1. 归一化file_id
	fileID, err := common.NormalizeFileID(params.FileID)
	if err != nil {
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", []interface{}{"invalid file_id"})
	}

	// 2. 查询flow_file
	flowFile, err := rds.GetFlowFileDao().GetByID(ctx, fileID)
	if err != nil {
		return nil, ierrors.NewIError(ierrors.FileNotFound, "", map[string]string{"file_id": params.FileID})
	}

	// 3. 验证状态
	if flowFile.Status != rds.FlowFileStatusPending {
		// 如果已经是ready，返回幂等成功
		if flowFile.Status == rds.FlowFileStatusReady {
			return &CompleteDataflowDocUploadResult{
				FileID:    fmt.Sprintf("%d", flowFile.ID),
				DocID:     common.BuildDFSURI(flowFile.ID),
				Status:    "ready",
				Continued: false,
			}, nil
		}
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]string{"message": "file status is not pending"})
	}

	// 4. 获取可用OSS并验证对象存在
	ossID, err := m.ossGateway.GetAvaildOSS(ctx)
	if err != nil {
		return nil, ierrors.NewIError(ierrors.InternalError, "", map[string]string{"message": "no available OSS"})
	}

	objectKey, _ := common.BuildFlowFileObjectKey(flowFile.ID, flowFile.Name)

	// 验证对象是否存在
	size, err := m.ossGateway.GetObjectMeta(ctx, ossID, objectKey, false)
	if err != nil {
		return nil, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]string{"message": "object not found in OSS"})
	}

	// 5. 创建flow_storage记录
	now := time.Now().Unix()
	flowStorage := &rds.FlowStorage{
		ID:        store.NextID(),
		OssID:     ossID,
		ObjectKey: objectKey,
		Name:      flowFile.Name,
		Size:      uint64(size),
		Etag:      params.Etag,
		Status:    rds.FlowStorageStatusNormal,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := rds.GetFlowStorageDao().Insert(ctx, flowStorage); err != nil {
		log.Warnf("[CompleteDataflowDocUpload] Insert flow_storage err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 6. 更新flow_file状态
	readyStatus := rds.FlowFileStatusReady
	if err := rds.GetFlowFileDao().Update(ctx, flowFile.ID, &rds.FlowFileUpdateParams{
		StorageID: &flowStorage.ID,
		Status:    &readyStatus,
	}); err != nil {
		log.Warnf("[CompleteDataflowDocUpload] Update flow_file err: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 7. 查询task_resume记录并恢复阻塞任务
	resumeList, _ := rds.GetFlowTaskResumeDao().List(ctx, &rds.FlowTaskResumeQueryOptions{
		ResourceType: "file",
		ResourceID:   &flowFile.ID,
		Limit:        1,
	})
	if len(resumeList) > 0 {
		taskResume := resumeList[0]
		// 调用ContinueBlockInstances恢复执行
		if err := m.ContinueBlockInstances(ctx, []string{taskResume.TaskInstanceID}, map[string]interface{}{
			"id":     common.BuildDFSURI(flowFile.ID),
			"docid":  common.BuildDFSURI(flowFile.ID),
			"status": "ready",
		}, entity.TaskInstanceStatusSuccess); err != nil {
			log.Warnf("[CompleteDataflowDocUpload] ContinueBlockInstances err: %s", err.Error())
		} else {
			// 恢复成功后删除resume记录
			rds.GetFlowTaskResumeDao().Delete(ctx, taskResume.ID)
		}
	}

	return &CompleteDataflowDocUploadResult{
		FileID:    fmt.Sprintf("%d", flowFile.ID),
		DocID:     common.BuildDFSURI(flowFile.ID),
		Status:    "ready",
		Continued: true,
	}, nil
}

// buildDataflowDocRunVar 构建Dataflow文档触发的运行变量
func (m *mgnt) buildDataflowDocRunVar(dfsURI, downloadURL, sourceFrom string, userDetail drivenadapters.UserInfo, userInfo *drivenadapters.UserInfo) map[string]string {
	runVar := map[string]string{
		"id":            dfsURI,
		"_type":         "file",
		"source_type":   "doc",
		"source_from":   sourceFrom,
		"userid":        userInfo.UserID,
		"operator_id":   userInfo.UserID,
		"operator_name": userDetail.UserName,
		"operator_type": userInfo.AccountType,
	}

	if downloadURL != "" {
		runVar["download_url"] = downloadURL
	}

	return runVar
}
