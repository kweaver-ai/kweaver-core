// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package job

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/infra/otel/otellog"
	"bkn-backend/infra/otel/oteltrace"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
	"bkn-backend/logics/object_type"
	"bkn-backend/logics/permission"
	"bkn-backend/logics/user_mgmt"
	"bkn-backend/worker"
)

var (
	jServiceOnce sync.Once
	jService     interfaces.JobService
)

type jobService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	ja         interfaces.JobAccess
	je         interfaces.JobExecutor
	ps         interfaces.PermissionService
	ots        interfaces.ObjectTypeService
	ums        interfaces.UserMgmtService
}

func NewJobService(appSetting *common.AppSetting) interfaces.JobService {
	jServiceOnce.Do(func() {
		jService = &jobService{
			appSetting: appSetting,
			db:         logics.DB,
			ja:         logics.JA,
			je:         worker.NewJobExecutor(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			ots:        object_type.NewObjectTypeService(appSetting),
			ums:        user_mgmt.NewUserMgmtService(appSetting),
		}
	})
	return jService
}

func (js *jobService) CreateJob(ctx context.Context, jobInfo *interfaces.JobInfo) (jobID string, err error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Create job")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err = js.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   jobInfo.KNID,
	}, []string{interfaces.OPERATION_TYPE_TASK_MANAGE})
	if err != nil {
		return "", err
	}

	jobList, err := js.ja.ListJobs(ctx, interfaces.JobsQueryParams{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Sort:      "f_id",
			Direction: interfaces.DESC_DIRECTION,
			Offset:    0,
			Limit:     -1,
		},
		KNID:   jobInfo.KNID,
		Branch: jobInfo.Branch,
		State: []interfaces.JobState{
			interfaces.JobStateRunning,
			interfaces.JobStatePending,
		},
	})
	if err != nil {
		return "", err
	}
	if len(jobList) > 0 {
		return "", rest.NewHTTPError(ctx, http.StatusForbidden,
			berrors.BknBackend_Job_CreateConflict).
			WithErrorDetails("there is an other Job already running or pending in system")
	}

	// 若提交的模型id为空，生成分布式ID
	if jobInfo.ID == "" {
		jobInfo.ID = xid.New().String()
	}
	jobID = jobInfo.ID

	jobInfo.State = interfaces.JobStatePending
	jobInfo.StateDetail = ""

	currentTime := time.Now().UnixMilli()
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	jobInfo.Creator = accountInfo
	jobInfo.CreateTime = currentTime
	jobInfo.FinishTime = 0
	jobInfo.TimeCost = 0

	if jobInfo.JobConceptConfig == nil {
		jobInfo.JobConceptConfig = []interfaces.ConceptConfig{}
	}

	taskInfos := map[string]*interfaces.TaskInfo{}
	if len(jobInfo.JobConceptConfig) == 0 {
		objectTypes, err := js.ots.GetAllObjectTypesByKnID(ctx, jobInfo.KNID, jobInfo.Branch)
		if err != nil {
			return "", err
		}
		for _, objectType := range objectTypes {
			if objectType.DataSource == nil {
				continue
			}
			if len(objectType.PrimaryKeys) == 0 {
				return "", rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_Job_InvalidObjectType).
					WithErrorDetails(fmt.Sprintf("ObjectType %s has no primary key", objectType.OTName))
			}

			dsType := objectType.DataSource.Type
			if dsType == "" {
				dsType = interfaces.DATA_SOURCE_TYPE_DATA_VIEW
			}
			if dsType == interfaces.DATA_SOURCE_TYPE_RESOURCE {
				continue
			}

			task_id := xid.New().String()

			taskInfos[task_id] = &interfaces.TaskInfo{
				ID:          task_id,
				Name:        objectType.OTName,
				JobID:       jobInfo.ID,
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				ConceptID:   objectType.OTID,
				TaskStateInfo: interfaces.TaskStateInfo{
					State:       interfaces.TaskStatePending,
					StateDetail: "",
					StartTime:   0,
					FinishTime:  0,
					TimeCost:    0,
				},
			}
		}
	} else {
		objectTypeIDs := map[string]bool{}
		for _, conceptConfig := range jobInfo.JobConceptConfig {
			if conceptConfig.ConceptType == interfaces.MODULE_TYPE_OBJECT_TYPE {
				objectTypeIDs[conceptConfig.ConceptID] = true
			}
		}

		objectTypes, err := js.ots.GetAllObjectTypesByKnID(ctx, jobInfo.KNID, jobInfo.Branch)
		if err != nil {
			return "", err
		}
		for _, objectType := range objectTypes {
			if objectType.DataSource == nil {
				continue
			}
			if len(objectType.PrimaryKeys) == 0 {
				return "", rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_Job_InvalidObjectType).
					WithErrorDetails(fmt.Sprintf("ObjectType %s has no primary key", objectType.OTName))
			}
			if !objectTypeIDs[objectType.OTID] {
				continue
			}

			dsType := objectType.DataSource.Type
			if dsType == "" {
				dsType = interfaces.DATA_SOURCE_TYPE_DATA_VIEW
			}
			if dsType == interfaces.DATA_SOURCE_TYPE_RESOURCE {
				continue
			}

			task_id := xid.New().String()

			taskInfos[task_id] = &interfaces.TaskInfo{
				ID:          task_id,
				Name:        objectType.OTName,
				JobID:       jobInfo.ID,
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				ConceptID:   objectType.OTID,
				TaskStateInfo: interfaces.TaskStateInfo{
					State:       interfaces.TaskStatePending,
					StateDetail: "",
					StartTime:   0,
					FinishTime:  0,
					TimeCost:    0,
				},
			}
		}
	}
	jobInfo.TaskInfos = taskInfos
	if len(taskInfos) == 0 {
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_Job_NoneConceptType).
			WithErrorDetails("JobConceptConfig is empty, or no indexable object types: resource-backed object types do not generate index build tasks")
	}

	tx, err := js.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		otellog.LogError(ctx, "Begin transaction error", err)

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Job_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("CreateJob Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				otellog.LogError(ctx, "CreateJob Transaction Rollback Error", err)
			}
		}
	}()

	// 创建
	err = js.ja.CreateJob(ctx, tx, jobInfo)
	if err != nil {
		logger.Errorf("CreateJob error: %s", err.Error())
		span.SetStatus(codes.Error, "创建任务失败")

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	err = js.ja.CreateTasks(ctx, tx, taskInfos)
	if err != nil {
		logger.Errorf("CreateTasks error: %s", err.Error())
		span.SetStatus(codes.Error, "创建子任务失败")

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	err = tx.Commit()
	if err != nil {
		logger.Errorf("CreateJob Transaction Commit Failed:%v", err)
		span.SetStatus(codes.Error, "提交事务失败")
		otellog.LogError(ctx, "CreateJob Transaction Commit Failed", err)
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Job_InternalError_CommitTransactionFailed).
			WithErrorDetails(err.Error())
	}

	_ = js.je.AddJob(ctx, jobInfo)

	span.SetStatus(codes.Ok, "")
	return jobInfo.ID, nil
}

func (js *jobService) DeleteJobsByIDs(ctx context.Context, knID string, branch string, jobIDs []string) (err error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete jobs by ids")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err = js.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_TASK_MANAGE})
	if err != nil {
		return err
	}

	tx, err := js.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		otellog.LogError(ctx, "Begin transaction error", err)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Job_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("DeleteJobs Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				otellog.LogError(ctx, "DeleteJobs Transaction Commit Failed", err)
				return
			}
			logger.Infof("DeleteJobs Transaction Commit Success")
			otellog.LogDebug(ctx, "DeleteJobs Transaction Commit Success")
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("DeleteJobs Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				otellog.LogError(ctx, "DeleteJobs Transaction Rollback Error", err)
			}
		}
	}()

	// 删除
	_, err = js.ja.DeleteJobsByIDs(ctx, tx, jobIDs)
	if err != nil {
		logger.Errorf("DeleteJobsByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除任务失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	_, err = js.ja.DeleteTasksByJobIDs(ctx, tx, jobIDs)
	if err != nil {
		logger.Errorf("DeleteTasksByJobIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除子任务失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return err
}

func (js *jobService) DeleteJobsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (err error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete jobs by kn id")
	defer span.End()

	if tx == nil {
		otellog.LogError(ctx, "missing transaction", nil)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Job_InternalError_MissingTransaction).
			WithErrorDetails("missing transaction")
	}

	jobIDs, err := js.ja.GetJobIDsByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("GetJobIDsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "查询任务失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}
	// 删除
	_, err = js.ja.DeleteJobsByIDs(ctx, tx, jobIDs)
	if err != nil {
		logger.Errorf("DeleteJobsByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除任务失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	_, err = js.ja.DeleteTasksByJobIDs(ctx, tx, jobIDs)
	if err != nil {
		logger.Errorf("DeleteTasksByJobIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除子任务失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return err
}

func (js *jobService) ListJobs(ctx context.Context, queryParams interfaces.JobsQueryParams) ([]*interfaces.JobInfo, int64, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "List jobs")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := js.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   queryParams.KNID,
	}, []string{interfaces.OPERATION_TYPE_TASK_MANAGE})
	if err != nil {
		return nil, 0, err
	}

	// 列表
	jobs, err := js.ja.ListJobs(ctx, queryParams)
	if err != nil {
		logger.Errorf("ListJobs error: %s", err.Error())
		span.SetStatus(codes.Error, "查询任务失败")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	total, err := js.ja.GetJobsTotal(ctx, queryParams)
	if err != nil {
		logger.Errorf("GetJobsTotal error: %s", err.Error())
		span.SetStatus(codes.Error, "查询任务总数失败")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(jobs))
	for _, job := range jobs {
		accountInfos = append(accountInfos, &job.Creator)
	}

	err = js.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.JobInfo{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Job_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return jobs, total, nil
}

func (js *jobService) ListTasks(ctx context.Context, queryParams interfaces.TasksQueryParams) ([]*interfaces.TaskInfo, int64, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "List tasks")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := js.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   queryParams.KNID,
	}, []string{interfaces.OPERATION_TYPE_TASK_MANAGE})
	if err != nil {
		return nil, 0, err
	}

	// 列表
	tasks, err := js.ja.ListTasks(ctx, queryParams)
	if err != nil {
		logger.Errorf("ListTasks error: %s", err.Error())
		span.SetStatus(codes.Error, "查询任务失败")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	// 总数
	total, err := js.ja.GetTasksTotal(ctx, queryParams)
	if err != nil {
		logger.Errorf("GetTasksTotal error: %s", err.Error())
		span.SetStatus(codes.Error, "查询任务总数失败")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return tasks, total, nil
}

func (js *jobService) GetJobsByIDs(ctx context.Context, jobIDs []string) (map[string]*interfaces.JobInfo, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Get jobs by ids")
	defer span.End()

	// 查询
	jobs, err := js.ja.GetJobsByIDs(ctx, jobIDs)
	if err != nil {
		logger.Errorf("GetJobsByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "查询任务失败")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return jobs, nil
}

func (js *jobService) GetJobByID(ctx context.Context, jobID string) (*interfaces.JobInfo, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Get job by id")
	defer span.End()

	// 查询
	job, err := js.ja.GetJobByID(ctx, jobID)
	if err != nil {
		logger.Errorf("GetJobByID error: %s", err.Error())
		span.SetStatus(codes.Error, "查询任务失败")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Job_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return job, nil
}
