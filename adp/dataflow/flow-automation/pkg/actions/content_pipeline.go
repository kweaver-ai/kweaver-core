package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	lock "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/lock"
	libstore "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
)

type SourceType string

const (
	SourceTypeDocID SourceType = "docid"
	SourceTypeUrl   SourceType = "url"
)

const (
	taskCacheLockPrefix  = "automation:content_pipeline"
	taskCacheLockTTL     = 30 * time.Second
	taskCacheLockWaitTTL = 60 * time.Second
)

type ContentPipelineFullText struct {
	DocID   string `json:"docid"`
	Version string `json:"version"`
}

func (a *ContentPipelineFullText) Name() string {
	return common.OpContentPipelineFullText
}

func (a *ContentPipelineFullText) ParameterNew() interface{} {
	return &ContentPipelineFullText{}
}

func (a *ContentPipelineFullText) Run(ctx entity.ExecuteContext, params interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)
	input := params.(*ContentPipelineFullText)

	if strings.HasPrefix(input.DocID, "dfs://") {
		return a.runDFSFullText(ctx, input)
	}

	return a.runPipelineFullText(ctx, input)
}

func (a *ContentPipelineFullText) runDFSFullText(ctx entity.ExecuteContext, input *ContentPipelineFullText) (interface{}, error) {
	log := traceLog.WithContext(ctx.Context())

	result, err := ctx.NewRepo().ExtractFullText(ctx.Context(), input.DocID)
	if err != nil {
		log.Warnf("[ContentPipelineFullText] ExtractFullText err: %s, docid: %s", err.Error(), input.DocID)
		return nil, err
	}

	ctx.ShareData().Set(ctx.GetTaskID(), result)
	return result, nil
}

func (a *ContentPipelineFullText) runPipelineFullText(ctx entity.ExecuteContext, input *ContentPipelineFullText) (interface{}, error) {
	log := traceLog.WithContext(ctx.Context())
	config := common.NewConfig()
	efast := drivenadapters.NewEfast()
	og := drivenadapters.NewOssGateWay()
	pipeline := drivenadapters.NewContentPipeline()

	taskCache := rds.GetTaskCache()

	ossInfo, err := efast.GetOssInfo(ctx.Context(), input.DocID, input.Version)
	if err != nil {
		log.Warnf("[ContentPipelineFullText] GetOssInfo err: %s, docid: %s, version: %s", err.Error(), input.DocID, input.Version)
		return nil, err
	}

	attr, err := efast.GetDocMsg(ctx.Context(), input.DocID)

	if err != nil {
		log.Warnf("[ContentPipelineFullText] GetDocMsg err: %s, docid: %s", err.Error(), input.DocID)
		return nil, err
	}

	ossPath := fmt.Sprintf("%s/%s", ossInfo.OssID, ossInfo.ObjectName)

	taskIns := ctx.GetTaskInstance()
	taskIns.Hash = hash(fmt.Sprintf("%s:%s", a.Name(), ossPath))

	err = taskIns.Patch(ctx.Context(), &entity.TaskInstance{
		BaseInfo: taskIns.BaseInfo,
		Hash:     taskIns.Hash,
	})

	if err != nil {
		log.Warnf("[ContentPipelineFullText] Patch err: %s, taskInsID: %s, hash: %s", err.Error(), taskIns.ID, taskIns.Hash)
		return nil, err
	}

	task, err := taskCache.GetByHash(ctx.Context(), taskIns.Hash)

	if err != nil {
		log.Warnf("[ContentPipelineFullText] GetByHash err: %s, hash: %s", err.Error(), taskIns.Hash)
		return nil, err
	}

	if task == nil {
		err = withTaskCacheLock(ctx.Context(), taskIns.Hash, taskIns.ID, taskCache, func(lockCtx context.Context) error {
			task, err = taskCache.GetByHash(lockCtx, taskIns.Hash)
			if err != nil {
				log.Warnf("[ContentPipelineFullText] GetByHash with lock err: %s, hash: %s", err.Error(), taskIns.Hash)
				return err
			}

			if task != nil {
				return nil
			}

			ossID, err := og.GetAvaildOSS(lockCtx)
			if err != nil {
				log.Warnf("[ContentPipelineFullText] GetAvaildOSS err: %s", err.Error())
				return err
			}

			ossKey := fmt.Sprintf(`%s/task_results/%s`, config.Server.StoragePrefix, taskIns.Hash)
			passback := fmt.Sprintf("automation:%s", taskIns.Hash)

			_, err = pipeline.NewJob(lockCtx, &drivenadapters.NewJobReq{
				Passback: passback,
				Source: &drivenadapters.SourceData{
					Type:  "oss",
					Value: ossPath,
				},
				Tasks: []*drivenadapters.TaskReq{
					{
						Key: "full_text",
						Steps: []*drivenadapters.StepReq{
							{
								Key: "full_text",
								Parameters: map[string]any{
									"filename": attr.Name,
									"upload_oss_resource": map[string]any{
										"oss_id":     ossID,
										"object_key": ossKey,
									},
								},
								Priority: 0,
							},
						},
					},
				},
			})

			if err != nil {
				log.Warnf("[ContentPipelineFullText] NewJob err: %s, docid: %s, hash: %s", err.Error(), input.DocID, taskIns.Hash)
				return err
			}

			now := time.Now().Unix()
			task = &rds.TaskCacheItem{
				ID:         store.NextID(),
				Hash:       taskIns.Hash,
				Status:     rds.TaskStatusPending,
				OssID:      ossID,
				OssKey:     ossKey,
				Ext:        ".txt",
				Size:       0,
				ErrMsg:     "",
				CreateTime: now,
				ModifyTime: now,
				ExpireTime: now + config.ActionConfig.FullText.ExpireSec,
			}

			err = taskCache.Insert(lockCtx, task)

			if err != nil {
				log.Warnf("[ContentPipelineFullText] Insert err: %s, hash: %s", err.Error(), taskIns.Hash)
				return err
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		if task == nil {
			task, err = taskCache.GetByHash(ctx.Context(), taskIns.Hash)
			if err != nil {
				log.Warnf("[ContentPipelineFullText] GetByHash after lock err: %s, hash: %s", err.Error(), taskIns.Hash)
				return nil, err
			}
		}
	}

	result := map[string]any{}
	switch task.Status {
	case rds.TaskStatusFailed:
		ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusFailed)
		return nil, fmt.Errorf("%s", task.ErrMsg)
	case rds.TaskStatusPending:
		ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusBlocked)
		return result, nil
	default:
		servicePrefix := false
		reader := og.NewReader(task.OssID, task.OssKey, drivenadapters.OssOpt{StoragePrifix: &servicePrefix})

		result["url"], _ = reader.Url(ctx.Context())
		result["text"], _ = reader.Text(ctx.Context())

		ctx.ShareData().Set(ctx.GetTaskID(), result)
	}

	return result, nil
}

func (a *ContentPipelineFullText) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {

	taskIns := ctx.GetTaskInstance()
	status, ok := ctx.ShareData().Get("__status_" + taskIns.ID)

	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

type ContentPipelineDocFormatConvert struct {
	DocID   string `json:"docid"`
	Version string `json:"version"`
}

func (a *ContentPipelineDocFormatConvert) Name() string {
	return common.OpContentPipelineDocFormatConvert
}

func (a *ContentPipelineDocFormatConvert) ParameterNew() interface{} {
	return &ContentPipelineDocFormatConvert{}
}

func (a *ContentPipelineDocFormatConvert) Run(ctx entity.ExecuteContext, params interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)
	input := params.(*ContentPipelineDocFormatConvert)

	if strings.HasPrefix(input.DocID, "dfs://") {
		return a.runDFSDocFormatConvert(ctx, input)
	}

	return a.runPipelineDocFormatConvert(ctx, input)
}

func (a *ContentPipelineDocFormatConvert) runDFSDocFormatConvert(ctx entity.ExecuteContext, input *ContentPipelineDocFormatConvert) (interface{}, error) {
	log := traceLog.WithContext(ctx.Context())
	taskIns := ctx.GetTaskInstance()

	err := ctx.NewRepo().ConvertToPDF(ctx.Context(), taskIns.ID, input.DocID)
	if err != nil {
		log.Warnf("[ContentPipelineDocFormatConvert] ConvertToPDF err: %s, docid: %s", err.Error(), input.DocID)
		return nil, err
	}

	ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusBlocked)
	return map[string]any{}, nil
}

func (a *ContentPipelineDocFormatConvert) runPipelineDocFormatConvert(ctx entity.ExecuteContext, input *ContentPipelineDocFormatConvert) (interface{}, error) {
	log := traceLog.WithContext(ctx.Context())
	config := common.NewConfig()
	efast := drivenadapters.NewEfast()
	og := drivenadapters.NewOssGateWay()
	pipeline := drivenadapters.NewContentPipeline()
	taskIns := ctx.GetTaskInstance()

	taskCache := rds.GetTaskCache()

	ossInfo, err := efast.GetOssInfo(ctx.Context(), input.DocID, input.Version)
	if err != nil {
		log.Warnf("[ContentPipelineDocFormatConvert] GetOssInfo err: %s, docid: %s, version: %s", err.Error(), input.DocID, input.Version)
		return nil, err
	}

	attr, err := efast.GetDocMsg(ctx.Context(), input.DocID)

	if err != nil {
		log.Warnf("[ContentPipelineDocFormatConvert] GetDocMsg err: %s, docid: %s", err.Error(), input.DocID)
		return nil, err
	}

	ossPath := fmt.Sprintf("%s/%s", ossInfo.OssID, ossInfo.ObjectName)

	taskIns.Hash = hash(fmt.Sprintf("%s:%s", a.Name(), ossPath))

	err = taskIns.Patch(ctx.Context(), &entity.TaskInstance{
		BaseInfo: taskIns.BaseInfo,
		Hash:     taskIns.Hash,
	})

	if err != nil {
		log.Warnf("[ContentPipelineDocFormatConvert] Patch err: %s, taskInsID: %s, hash: %s", err.Error(), taskIns.ID, taskIns.Hash)
		return nil, err
	}

	task, err := taskCache.GetByHash(ctx.Context(), taskIns.Hash)

	if err != nil {
		log.Warnf("[ContentPipelineDocFormatConvert] GetByHash err: %s, hash: %s", err.Error(), taskIns.Hash)
		return nil, err
	}

	if task == nil {
		err = withTaskCacheLock(ctx.Context(), taskIns.Hash, taskIns.ID, taskCache, func(lockCtx context.Context) error {
			task, err = taskCache.GetByHash(lockCtx, taskIns.Hash)
			if err != nil {
				log.Warnf("[ContentPipelineDocFormatConvert] GetByHash with lock err: %s, hash: %s", err.Error(), taskIns.Hash)
				return err
			}

			if task != nil {
				return nil
			}

			ossID, err := og.GetAvaildOSS(lockCtx)

			if err != nil {
				log.Warnf("[ContentPipelineDocFormatConvert] GetAvaildOSS err: %s", err.Error())
				return err
			}

			ossKey := fmt.Sprintf(`%s/task_results/%s`, config.Server.StoragePrefix, taskIns.Hash)
			passback := fmt.Sprintf("automation:%s", taskIns.Hash)

			_, err = pipeline.NewJob(lockCtx, &drivenadapters.NewJobReq{
				Passback: passback,
				Source: &drivenadapters.SourceData{
					Type:  "oss",
					Value: ossPath,
				},
				Tasks: []*drivenadapters.TaskReq{
					{
						Key: "doc_format_convert",
						Steps: []*drivenadapters.StepReq{
							{
								Key: "doc_format_convert",
								Parameters: map[string]any{
									"request_from":     "pipeline",
									"file_name":        attr.Name,
									"target_format":    "pdf",
									"update_data_type": "oss",
									"upload_location":  fmt.Sprintf("%s/%s", ossID, ossKey),
								},
								Priority: 0,
							},
						},
					},
				},
			})

			if err != nil {
				log.Warnf("[ContentPipelineDocFormatConvert] NewJob err: %s, docid: %s, hash: %s", err.Error(), input.DocID, taskIns.Hash)
				return err
			}

			now := time.Now().Unix()
			task = &rds.TaskCacheItem{
				ID:         store.NextID(),
				Hash:       taskIns.Hash,
				Status:     rds.TaskStatusPending,
				OssID:      ossID,
				OssKey:     ossKey,
				Ext:        ".pdf",
				Size:       0,
				ErrMsg:     "",
				CreateTime: now,
				ModifyTime: now,
				ExpireTime: now + config.ActionConfig.DocFormatConvert.ExpireSec,
			}

			err = taskCache.Insert(lockCtx, task)

			if err != nil {
				log.Warnf("[ContentPipelineDocFormatConvert] Insert err: %s, hash: %s", err.Error(), taskIns.Hash)
				return err
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		if task == nil {
			task, err = taskCache.GetByHash(ctx.Context(), taskIns.Hash)
			if err != nil {
				log.Warnf("[ContentPipelineDocFormatConvert] GetByHash after lock err: %s, hash: %s", err.Error(), taskIns.Hash)
				return nil, err
			}
		}
	}

	result := map[string]any{}
	switch task.Status {
	case rds.TaskStatusFailed:
		ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusFailed)
		return nil, fmt.Errorf("%s", task.ErrMsg)
	case rds.TaskStatusPending:
		ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusBlocked)
		return result, nil
	default:
		servicePrefix := false
		reader := og.NewReader(task.OssID, task.OssKey, drivenadapters.OssOpt{StoragePrifix: &servicePrefix})
		result["url"], _ = reader.Url(ctx.Context())

		ctx.ShareData().Set(ctx.GetTaskID(), result)
	}

	return result, nil

}

func (a *ContentPipelineDocFormatConvert) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {

	taskIns := ctx.GetTaskInstance()
	status, ok := ctx.ShareData().Get("__status_" + taskIns.ID)

	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

func withTaskCacheLock(ctx context.Context, hash, owner string, taskCache rds.TaskCache, fn func(context.Context) error) error {
	lockClient := lock.NewDistributeLock(libstore.NewRedis(), fmt.Sprintf("%s:%s", taskCacheLockPrefix, hash), owner)
	lockCtx, cancel := context.WithTimeout(ctx, taskCacheLockWaitTTL)
	defer cancel()

	err := lockClient.TryLock(lockCtx, taskCacheLockTTL, false)
	if err != nil {
		cachedTask, err := taskCache.GetByHash(ctx, hash)
		if err == nil && cachedTask != nil {
			return nil
		}
		return fmt.Errorf("acquire task cache lock timeout, hash: %s", hash)
	}
	defer lockClient.Release()

	return fn(lockCtx)
}
