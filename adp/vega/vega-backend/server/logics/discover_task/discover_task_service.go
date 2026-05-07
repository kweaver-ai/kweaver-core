// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package discover_task provides DiscoverTask business logic.
package discover_task

import (
	"context"
	"fmt"
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
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	asynq_access "vega-backend/drivenadapters/asynq"
	discovertaskaccess "vega-backend/drivenadapters/discover_task"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

var (
	dtsOnce    sync.Once
	dtsService interfaces.DiscoverTaskService
)

type discoverTaskService struct {
	appSetting *common.AppSetting
	client     *asynq.Client
	dta        interfaces.DiscoverTaskAccess
}

// NewDiscoverTaskService creates or returns the singleton DiscoverTaskService.
func NewDiscoverTaskService(appSetting *common.AppSetting) interfaces.DiscoverTaskService {
	dtsOnce.Do(func() {
		asynqAccess := asynq_access.NewAsynqAccess(appSetting)
		dtsService = &discoverTaskService{
			appSetting: appSetting,
			client:     asynqAccess.CreateClient(context.Background()),
			dta:        discovertaskaccess.NewDiscoverTaskAccess(appSetting),
		}
	})
	return dtsService
}

// Create creates a new DiscoverTask and enqueues it to the task queue.
// Create 创建一个新的发现任务
// 参数:
//   - ctx: 上下文，用于传递请求范围的数据和取消信号
//   - catalogID: 目录ID，用于标识要执行发现任务的目录
//
// 返回值:
//   - string: 创建的任务ID
//   - error: 错误信息，如果创建失败则返回错误
func (dts *discoverTaskService) Create(ctx context.Context, catalogID string, scheduledTask ...string) (string, error) {
	// 使用分布式追踪系统创建一个span，用于追踪服务调用
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverTaskService.Create",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End() // 确保span在函数结束时结束

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if ai, ok := ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo); ok {
		accountInfo = ai
	}

	// 处理可选的taskType参数
	// 默认根据上下文判断：如果是从定时任务服务调用，则默认为scheduled，否则为manual
	triggerType := interfaces.DiscoverTaskTriggerManual
	scheduledId := ""
	strategies := []string{}
	if len(scheduledTask) > 0 && scheduledTask[0] != "" {
		triggerType = scheduledTask[0]
		if len(scheduledTask) > 1 {
			scheduledId = scheduledTask[1]
		}
		if len(scheduledTask) > 2 && scheduledTask[2] != "" {
			// 反序列化 strategies 字符串为 []string
			if err := sonic.Unmarshal([]byte(scheduledTask[2]), &strategies); err != nil {
				logger.Warnf("Failed to unmarshal strategies: %v", err)
				strategies = []string{}
			}
		}
	}
	now := time.Now().UnixMilli()
	task := &interfaces.DiscoverTask{
		ID:          xid.New().String(),
		CatalogID:   catalogID,
		ScheduleID: scheduledId,
		Strategies:  strategies,
		TriggerType: triggerType,
		Status:      interfaces.DiscoverTaskStatusPending,
		Progress:    0,
		Message:     "",
		Creator:     accountInfo,
		CreateTime:  now,
	}

	// 1. Write to database
	if err := dts.dta.Create(ctx, task); err != nil {
		logger.Errorf("Failed to create discover task: %v", err)
		o11y.Error(ctx, "Failed to create discover task")
		return "", err
	}

	// 2. Enqueue task to task queue
	payload, err := sonic.Marshal(&interfaces.DiscoverTaskMessage{
		TaskID: task.ID,
	})
	if err != nil {
		logger.Errorf("Failed to marshal discover task: %v", err)
		o11y.Error(ctx, "Failed to marshal discover task")
		return "", err
	}
	// 设置任务执行超时时间为 30 分钟
	asynqTask := asynq.NewTask(interfaces.DiscoverTaskType, payload)
	info, err := dts.client.Enqueue(asynqTask,
		asynq.Queue(interfaces.HighQueue),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Minute),
		asynq.Deadline(time.Now().Add(12*time.Hour)),
	)
	if err != nil {
		logger.Errorf("Failed to enqueue discover task: %v", err)
		o11y.Error(ctx, "Failed to enqueue discover task")
		return "", err
	}

	logger.Infof("Enqueued task: id=%s, type=%s, queue=%s", info.ID, info.Type, info.Queue)
	return task.ID, nil
}

// CreateScheduled method removed - scheduled tasks are now managed by ScheduledDiscoverTaskService

// GetByID retrieves a DiscoverTask by ID.
func (dts *discoverTaskService) GetByID(ctx context.Context, id string) (*interfaces.DiscoverTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverTaskService.GetByID",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.GetByID(ctx, id)
}

// List lists DiscoverTasks for a catalog.
func (dts *discoverTaskService) List(ctx context.Context, params interfaces.DiscoverTaskQueryParams) ([]*interfaces.DiscoverTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverTaskService.List",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.List(ctx, params)
}

// UpdateStatus updates a DiscoverTask's status.
func (dts *discoverTaskService) UpdateStatus(ctx context.Context, id, status, message string, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverTaskService.UpdateStatus",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.UpdateStatus(ctx, id, status, message, stime)
}

// UpdateResult updates a DiscoverTask's result.
func (dts *discoverTaskService) UpdateResult(ctx context.Context, id string, result *interfaces.DiscoverResult, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverTaskService.UpdateResult",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.UpdateResult(ctx, id, result, stime)
}

// CheckExistByStatuses checks if DiscoverTasks exists by catalog ID and statuses.
func (dts *discoverTaskService) CheckExistByStatuses(ctx context.Context, catalogID string, statuses []string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverTaskService.CheckExistByStatuses",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.CheckExistByStatuses(ctx, catalogID, statuses)
}

// Delete atomically deletes discover tasks by IDs after pre-validating existence and status.
//
// Behavior:
//   - Input ids are de-duplicated.
//   - Loads each id; if any task is in pending/running, returns 409 HasRunningExecution
//     with {running_ids: [...]}. This check cannot be bypassed.
//   - If any id is missing, returns 404 NotFound with {missing_ids: [...]} unless
//     ignoreMissing=true (then missing ids are silently dropped from the delete set).
//   - Deletes pass-through tasks one-by-one. Mid-loop errors return 500 (rare, bounded
//     by pre-validation).
func (dts *discoverTaskService) Delete(ctx context.Context, ids []string, ignoreMissing bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverTaskService.Delete",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Dedupe ids while preserving order.
	seen := make(map[string]struct{}, len(ids))
	uniqueIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	toDelete := make([]string, 0, len(uniqueIDs))
	missingIDs := make([]string, 0)
	runningIDs := make([]string, 0)

	for _, id := range uniqueIDs {
		task, err := dts.dta.GetByID(ctx, id)
		if err != nil {
			logger.Errorf("Get discover_task %s failed: %v", id, err)
			o11y.Error(ctx, fmt.Sprintf("Get discover_task %s failed: %v", id, err))
			span.SetStatus(codes.Error, "Get discover_task failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverTask_InternalError_GetFailed).
				WithErrorDetails(err.Error())
		}
		if task == nil {
			missingIDs = append(missingIDs, id)
			continue
		}
		if task.Status == interfaces.DiscoverTaskStatusPending || task.Status == interfaces.DiscoverTaskStatusRunning {
			runningIDs = append(runningIDs, id)
			continue
		}
		toDelete = append(toDelete, id)
	}

	if len(runningIDs) > 0 {
		span.SetStatus(codes.Error, "Some tasks are pending or running")
		return rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_DiscoverTask_HasRunningExecution).
			WithErrorDetails(map[string]any{"running_ids": runningIDs})
	}
	if len(missingIDs) > 0 && !ignoreMissing {
		span.SetStatus(codes.Error, "Some discover tasks not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_DiscoverTask_NotFound).
			WithErrorDetails(map[string]any{"missing_ids": missingIDs})
	}

	for _, id := range toDelete {
		if err := dts.dta.Delete(ctx, id); err != nil {
			logger.Errorf("Delete discover_task %s failed: %v", id, err)
			o11y.Error(ctx, fmt.Sprintf("Delete discover_task %s failed: %v", id, err))
			span.SetStatus(codes.Error, "Delete discover_task failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverTask_InternalError_DeleteFailed).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
