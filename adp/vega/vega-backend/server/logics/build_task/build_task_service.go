// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package build_task provides BuildTask management business logic.
package build_task

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	asynqAccess "vega-backend/drivenadapters/asynq"
	taskAccess "vega-backend/drivenadapters/build_task"
	"vega-backend/drivenadapters/model_factory"
	resourceAccess "vega-backend/drivenadapters/resource"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
)

var (
	btsOnce sync.Once
	btsInst interfaces.BuildTaskService
)

type buildTaskService struct {
	appSetting *common.AppSetting
	cs         interfaces.CatalogService
	ra         interfaces.ResourceAccess
	bta        interfaces.BuildTaskAccess
	mfa        interfaces.ModelFactoryAccess
}

// NewBuildTaskService creates a new BuildTaskService.
func NewBuildTaskService(appSetting *common.AppSetting) interfaces.BuildTaskService {
	btsOnce.Do(func() {
		btsInst = &buildTaskService{
			appSetting: appSetting,
			cs:         catalog.NewCatalogService(appSetting),
			ra:         resourceAccess.NewResourceAccess(appSetting),
			bta:        taskAccess.NewBuildTaskAccess(appSetting),
			mfa:        model_factory.NewModelFactoryAccess(appSetting),
		}
	})
	return btsInst
}

// CreateBuildTask creates a new build task for a resource.
func (s *buildTaskService) CreateBuildTask(ctx context.Context, resourceID string, req *interfaces.BuildTaskRequest) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create build task")
	defer span.End()

	resource, err := s.ra.GetByID(ctx, resourceID)
	if err != nil {
		span.SetStatus(codes.Error, "Get resource failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
	}

	if resource.Category != interfaces.ResourceCategoryTable {
		span.SetStatus(codes.Error, "Resource category is not table")
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails("Resource category must be table")
	}

	cat, err := s.cs.GetByID(ctx, resource.CatalogID, false)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if cat == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
	}

	existing, err := s.bta.GetByResourceID(ctx, resourceID)
	if err != nil {
		logger.Errorf("Check existing build task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Check existing build task failed: %v", err))
		span.SetStatus(codes.Error, "Check existing build task failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if existing != nil {
		span.SetStatus(codes.Error, "Resource already has a build task")
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_Exist).
			WithErrorDetails("Resource already has a build task")
	}

	if req.Mode == interfaces.BuildTaskModeStreaming {
		primaryKeys := []any{}
		if resource.SourceMetadata != nil {
			if v, ok := resource.SourceMetadata["primary_keys"]; ok {
				primaryKeys = v.([]any)
			}
		}
		if len(primaryKeys) == 0 {
			span.SetStatus(codes.Error, "Resource has no primary key for build task")
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_InternalError_CreateFailed).
				WithErrorDetails("Resource has no primary key")
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	if req.EmbeddingModel == "" && req.EmbeddingFields != "" {
		req.EmbeddingModel = interfaces.DEFAULT_EMBEDDING_MODEL
	}
	if req.EmbeddingModel != "" && req.ModelDimensions == 0 {
		embeddingModel, err := s.mfa.GetModelByName(ctx, req.EmbeddingModel)
		if err != nil {
			span.SetStatus(codes.Error, "Get model by name failed")
			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_CreateFailed).
				WithErrorDetails(err.Error())
		}
		req.ModelDimensions = embeddingModel.EmbeddingDim
	}

	now := time.Now().UnixMilli()
	buildTask := &interfaces.BuildTask{
		ID:              xid.New().String(),
		ResourceID:      resourceID,
		CatalogID:       resource.CatalogID,
		Status:          interfaces.BuildTaskStatusInit,
		Mode:            req.Mode,
		Creator:         accountInfo,
		CreateTime:      now,
		Updater:         accountInfo,
		UpdateTime:      now,
		EmbeddingFields: req.EmbeddingFields,
		BuildKeyFields:  req.BuildKeyFields,
		EmbeddingModel:  req.EmbeddingModel,
		ModelDimensions: req.ModelDimensions,
	}

	if err := s.bta.Create(ctx, buildTask); err != nil {
		logger.Errorf("Create build task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create build task failed: %v", err))
		span.SetStatus(codes.Error, "Create build task failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return buildTask.ID, nil
}

// GetBuildTaskByID retrieves a build task by ID.
func (s *buildTaskService) GetBuildTaskByID(ctx context.Context, id string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build task")
	defer span.End()

	buildTask, err := s.bta.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get build task failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if buildTask == nil {
		span.SetStatus(codes.Error, "Build task not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_BuildTask_NotFound)
	}

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}

// GetBuildTaskByResourceID retrieves a build task by resource ID.
func (s *buildTaskService) GetBuildTaskByResourceID(ctx context.Context, resourceID string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build task by resource ID")
	defer span.End()

	buildTask, err := s.bta.GetByResourceID(ctx, resourceID)
	if err != nil {
		span.SetStatus(codes.Error, "Get build task failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}

// ListBuildTasks retrieves build tasks with filters and pagination.
func (s *buildTaskService) ListBuildTasks(ctx context.Context, params interfaces.BuildTasksQueryParams) ([]*interfaces.BuildTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List build tasks")
	defer span.End()

	buildTasks, total, err := s.bta.GetAllWithFilters(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List build tasks failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return buildTasks, total, nil
}

// UpdateBuildTaskStatus updates a build task's status (called by worker; HTTP path uses Start/Stop).
func (s *buildTaskService) UpdateBuildTaskStatus(ctx context.Context, taskID string, req *interfaces.UpdateBuildTaskStatusRequest) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update build task status")
	defer span.End()

	if req.ExecuteType == "" {
		req.ExecuteType = interfaces.BuildTaskExecuteTypeIncremental
	} else if req.ExecuteType != interfaces.BuildTaskExecuteTypeIncremental && req.ExecuteType != interfaces.BuildTaskExecuteTypeFull {
		span.SetStatus(codes.Error, "Invalid execute type")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_InvalidExecuteType).
			WithErrorDetails("Invalid execute type")
	}

	buildTask, err := s.bta.GetByID(ctx, taskID)
	if err != nil {
		span.SetStatus(codes.Error, "Get build task failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if buildTask == nil {
		span.SetStatus(codes.Error, "Build task not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_BuildTask_NotFound)
	}

	updates := map[string]interface{}{}
	switch req.Status {
	case interfaces.BuildTaskStatusRunning:
		switch buildTask.Status {
		case interfaces.BuildTaskStatusRunning:
			span.SetStatus(codes.Error, "Task is already running")
			return rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_BuildTask_InvalidStateTransition).
				WithErrorDetails("Task is already running")
		case interfaces.BuildTaskStatusStopping:
			span.SetStatus(codes.Error, "Task is stopping")
			return rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_BuildTask_InvalidStateTransition).
				WithErrorDetails("Task is stopping")
		}
		// status transition to running is performed by worker on actual execution
	case interfaces.BuildTaskStatusStopped:
		if buildTask.Status == interfaces.BuildTaskStatusStopped || buildTask.Status == interfaces.BuildTaskStatusStopping || buildTask.Status == interfaces.BuildTaskStatusFailed {
			span.SetStatus(codes.Ok, "Task is already stopped")
			return nil
		}
		if buildTask.Status == interfaces.BuildTaskStatusInit {
			span.SetStatus(codes.Error, "Task is init, cannot stop")
			return rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_BuildTask_InvalidStateTransition).
				WithErrorDetails("Task is init, cannot stop")
		}
		updates["status"] = interfaces.BuildTaskStatusStopping
	default:
		span.SetStatus(codes.Error, "Invalid status")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_InvalidStatus).
			WithErrorDetails("Invalid status")
	}

	if req.Status == interfaces.BuildTaskStatusRunning {
		payload, err := sonic.Marshal(&interfaces.BatchBuildTaskMessage{
			TaskID:      taskID,
			ExecuteType: req.ExecuteType,
		})
		if err != nil {
			logger.Errorf("Marshal build task message failed: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Marshal build task message failed: %v", err))
		} else {
			typename := interfaces.BuildTaskTypeBatch
			if buildTask.Mode == interfaces.BuildTaskModeStreaming {
				typename = interfaces.BuildTaskTypeStreaming
			}
			asynqTask := asynq.NewTask(typename, payload)
			client := asynqAccess.NewAsynqAccess(s.appSetting).CreateClient(context.Background())
			if _, err := client.Enqueue(asynqTask,
				asynq.Queue(interfaces.DefaultQueue),
				asynq.MaxRetry(interfaces.BUILD_TASK_MAX_RETRY_COUNT),
				asynq.Timeout(math.MaxInt64),
				asynq.Deadline(time.Unix(math.MaxInt64/1000000000, math.MaxInt64%1000000000)),
			); err != nil {
				logger.Errorf("Enqueue build task failed: %v", err)
				o11y.Error(ctx, fmt.Sprintf("Enqueue build task failed: %v", err))
			} else {
				logger.Infof("Build task %s enqueued for execution", taskID)
			}
		}
	} else {
		if err := s.bta.UpdateStatus(ctx, taskID, updates); err != nil {
			logger.Errorf("Update build task status failed: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Update build task status failed: %v", err))
			span.SetStatus(codes.Error, "Update build task status failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_UpdateFailed).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// StartBuildTask transitions a task from {init, stopped} to running.
// Note: persisted status remains init/stopped until the worker picks it up — clients should poll.
func (s *buildTaskService) StartBuildTask(ctx context.Context, taskID string, executeType string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Start build task")
	defer span.End()

	if executeType == "" {
		executeType = interfaces.BuildTaskExecuteTypeIncremental
	}
	if executeType != interfaces.BuildTaskExecuteTypeIncremental && executeType != interfaces.BuildTaskExecuteTypeFull {
		span.SetStatus(codes.Error, "Invalid execute type")
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_InvalidExecuteType).
			WithErrorDetails("Invalid execute type")
	}

	buildTask, err := s.bta.GetByID(ctx, taskID)
	if err != nil {
		span.SetStatus(codes.Error, "Get build task failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if buildTask == nil {
		span.SetStatus(codes.Error, "Build task not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_BuildTask_NotFound)
	}
	if buildTask.Status != interfaces.BuildTaskStatusInit && buildTask.Status != interfaces.BuildTaskStatusStopped {
		span.SetStatus(codes.Error, "Invalid state transition for start")
		return nil, rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_BuildTask_InvalidStateTransition).
			WithErrorDetails(fmt.Sprintf("cannot start task in status: %s", buildTask.Status))
	}

	if err := s.UpdateBuildTaskStatus(ctx, taskID, &interfaces.UpdateBuildTaskStatusRequest{
		Status:      interfaces.BuildTaskStatusRunning,
		ExecuteType: executeType,
	}); err != nil {
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}

// StopBuildTask transitions a task from running to stopping.
// Note: persisted status remains running until the worker advances it — clients should poll.
func (s *buildTaskService) StopBuildTask(ctx context.Context, taskID string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Stop build task")
	defer span.End()

	buildTask, err := s.bta.GetByID(ctx, taskID)
	if err != nil {
		span.SetStatus(codes.Error, "Get build task failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if buildTask == nil {
		span.SetStatus(codes.Error, "Build task not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_BuildTask_NotFound)
	}
	if buildTask.Status != interfaces.BuildTaskStatusRunning {
		span.SetStatus(codes.Error, "Invalid state transition for stop")
		return nil, rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_BuildTask_InvalidStateTransition).
			WithErrorDetails(fmt.Sprintf("cannot stop task in status: %s", buildTask.Status))
	}

	if err := s.UpdateBuildTaskStatus(ctx, taskID, &interfaces.UpdateBuildTaskStatusRequest{
		Status: interfaces.BuildTaskStatusStopped,
	}); err != nil {
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}

// DeleteBuildTask deletes a build task by ID.
func (s *buildTaskService) DeleteBuildTask(ctx context.Context, taskID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete build task")
	defer span.End()

	buildTask, err := s.bta.GetByID(ctx, taskID)
	if err != nil {
		span.SetStatus(codes.Error, "Get build task failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if buildTask == nil {
		span.SetStatus(codes.Error, "Build task not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_BuildTask_NotFound)
	}
	if buildTask.Status == interfaces.BuildTaskStatusRunning || buildTask.Status == interfaces.BuildTaskStatusStopping {
		span.SetStatus(codes.Error, "Cannot delete running or stopping task")
		return rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_BuildTask_HasRunningExecution).
			WithErrorDetails("Cannot delete running or stopping task; stop it first")
	}
	if err := s.bta.Delete(ctx, taskID); err != nil {
		logger.Errorf("Delete build task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Delete build task failed: %v", err))
		span.SetStatus(codes.Error, "Delete build task failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
