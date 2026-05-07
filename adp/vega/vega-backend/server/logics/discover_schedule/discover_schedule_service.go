// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package discover_schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"vega-backend/drivenadapters/discover_task"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	dsServiceOnce sync.Once
	dsService     interfaces.DiscoverScheduleService
)

type discoverScheduleService struct {
	appSetting *common.AppSetting
	dsa        interfaces.DiscoverScheduleAccess
	dts        interfaces.DiscoverTaskService
}

func (dss *discoverScheduleService) UpdateLastRun(ctx context.Context, id string, lastRun int64) error {
	return dss.dsa.UpdateLastRun(ctx, id, lastRun)
}

// NewDiscoverScheduleService creates a new DiscoverScheduleService.
func NewDiscoverScheduleService(appSetting *common.AppSetting, dts interfaces.DiscoverTaskService) interfaces.DiscoverScheduleService {
	dsServiceOnce.Do(func() {
		dsService = &discoverScheduleService{
			appSetting: appSetting,
			dsa:        discover_task.NewDiscoverScheduleAccess(appSetting),
			dts:        dts,
		}
	})
	return dsService
}

/**
 * 创建定时发现任务服务
 * @param ctx context.Context 上下文信息，用于传递请求范围的数据和取消信号
 * @param req *interfaces.DiscoverSchedule 定时发现任务请求结构体
 * @return string 返回创建的任务ID
 * @return error 返回操作过程中可能发生的错误
 */
func (dss *discoverScheduleService) Create(ctx context.Context, req *interfaces.DiscoverScheduleRequest) (string, error) {
	// 使用OpenTelemetry追踪请求执行过程
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.Create",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if ai, ok := ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo); ok {
		accountInfo = ai
	}

	// Validate cron expression
	if req.CronExpr == "" {
		logger.Error("Cron expression is required")
		o11y.Error(ctx, "Cron expression is required")
		return "", fmt.Errorf("cron_expr is required")
	}

	currentTime := time.Now().UnixMilli()
	schedule := &interfaces.DiscoverSchedule{
		ID:         req.CatalogID,
		CatalogID:  req.CatalogID,
		CronExpr:   req.CronExpr,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Enabled:    req.Enabled,
		Strategies: req.Strategies,

		Creator:    accountInfo,
		CreateTime: currentTime,
		Updater:    accountInfo,
		UpdateTime: currentTime,
	}

	// Create task
	if err := dss.dsa.Create(ctx, schedule); err != nil {
		logger.Errorf("Failed to create scheduled discover task: %v", err)
		o11y.Error(ctx, "Failed to create scheduled discover task")
		return "", err
	}

	logger.Infof("Created scheduled discover task: id=%s, catalog_id=%s, cron=%s",
		schedule.ID, schedule.CatalogID, schedule.CronExpr)
	return schedule.ID, nil
}

// GetByID retrieves a scheduled discover task by ID.
func (dss *discoverScheduleService) GetByID(ctx context.Context, id string) (*interfaces.DiscoverSchedule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.GetByID",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dss.dsa.GetByID(ctx, id)
}

// List lists scheduled discover tasks.
func (dss *discoverScheduleService) List(ctx context.Context, params interfaces.DiscoverScheduleQueryParams) ([]*interfaces.DiscoverSchedule, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.List",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dss.dsa.List(ctx, params)
}

// Update updates a scheduled discover task.
func (dss *discoverScheduleService) Update(ctx context.Context, id string, req *interfaces.DiscoverSchedule) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.Update",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if ai, ok := ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo); ok {
		accountInfo = ai
	}

	// Set updater
	req.Updater = accountInfo
	req.UpdateTime = time.Now().Unix()
	// Update task
	if err := dss.dsa.Update(ctx, req); err != nil {
		logger.Errorf("Failed to update scheduled discover task: %v", err)
		o11y.Error(ctx, "Failed to update scheduled discover task")
		return err
	}
	logger.Infof("Updated scheduled discover task: id=%s", id)
	return nil
}

// Delete deletes a scheduled discover task by ID.
func (dss *discoverScheduleService) Delete(ctx context.Context, id string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.Delete",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	if err := dss.dsa.Delete(ctx, id); err != nil {
		logger.Errorf("Failed to delete scheduled discover task: %v", err)
		o11y.Error(ctx, "Failed to delete scheduled discover task")
		return err
	}

	logger.Infof("Deleted scheduled discover task: id=%s", id)
	return nil
}

// Enable enables a scheduled discover task.
func (dss *discoverScheduleService) Enable(ctx context.Context, id string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.Enable",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	task, err := dss.dsa.GetByID(ctx, id)
	if err != nil {
		logger.Errorf("Failed to get scheduled discover task: %v", err)
		o11y.Error(ctx, "Failed to get scheduled discover task")
		return err
	}

	task.Enabled = true
	return dss.Update(ctx, id, task)
}

// Disable disables a scheduled discover task.
func (dss *discoverScheduleService) Disable(ctx context.Context, id string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.Disable",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	task, err := dss.dsa.GetByID(ctx, id)
	if err != nil {
		logger.Errorf("Failed to get scheduled discover task: %v", err)
		o11y.Error(ctx, "Failed to get scheduled discover task")
		return err
	}

	task.Enabled = false
	return dss.Update(ctx, id, task)
}

// GetEnabledTasks retrieves all enabled scheduled discover tasks.
func (dss *discoverScheduleService) GetEnabledTasks(ctx context.Context) ([]*interfaces.DiscoverSchedule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.GetEnabledTasks",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	return dss.dsa.GetEnabledTasks(ctx)
}

// ExecuteTask 是一个执行计划发现任务的方法
// 它接收一个上下文和一个计划发现任务作为参数，返回一个错误
func (dss *discoverScheduleService) ExecuteTask(ctx context.Context, task *interfaces.DiscoverSchedule) error {
	// 使用追踪器创建一个新的span，用于监控和追踪请求的执行过程
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoverScheduleService.ExecuteTask",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End() // 确保在函数返回时结束span

	// 检查DiscoverTaskService是否已设置
	if dss.dts == nil {
		logger.Error("DiscoverTaskService not set")
		o11y.Error(ctx, "DiscoverTaskService not set")
		return fmt.Errorf("DiscoverTaskService not set")
	}

	// 检查是否有正在执行的相同任务
	_, tasks, err := dss.dts.List(ctx, interfaces.DiscoverTaskQueryParams{
		CatalogID:   task.CatalogID,
		Status:      interfaces.DiscoverTaskStatusRunning,
		TriggerType: interfaces.DiscoverTaskTriggerScheduled,
	})
	if err != nil {
		logger.Errorf("Failed to check running tasks for catalog %s: %v", task.CatalogID, err)
		o11y.Error(ctx, "Failed to check running tasks")
		return err
	}
	if tasks > 0 {
		logger.Warnf("There is already a running discover task for catalog %s, skipping execution", task.CatalogID)
		return nil
	}
	// Create discover task：这里会创建一个task然后发送到redis mq里面去
	strategiesStr := strategiesToString(task.Strategies)
	_, err = dss.dts.Create(ctx, task.CatalogID, interfaces.DiscoverTaskTriggerScheduled, task.ID, strategiesStr)
	if err != nil {
		logger.Errorf("Failed to create discover task for catalog %s: %v", task.CatalogID, err)
		o11y.Error(ctx, "Failed to create discover task")
		return err
	}
	// Update last run time
	now := time.Now().UnixMilli()
	if err := dss.UpdateLastRun(ctx, task.ID, now); err != nil {
		logger.Errorf("Failed to update last run time: %v", err)
		// Don't return error here as the task was already executed
	}
	logger.Infof("Executed scheduled discover task: id=%s, catalog_id=%s", task.ID, task.CatalogID)
	return nil
}

// Helper functions for strategies array handling
func strategiesToString(strategies []string) string {
	if len(strategies) == 0 {
		return ""
	}
	data, err := json.Marshal(strategies)
	if err != nil {
		return ""
	}
	return string(data)
}
