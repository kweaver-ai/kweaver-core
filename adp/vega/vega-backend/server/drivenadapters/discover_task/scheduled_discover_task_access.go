// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package discover_task

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/xid"

	sq "github.com/Masterminds/squirrel"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/robfig/cron/v3"
	_ "github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	"vega-backend/interfaces"
)

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

func stringToStrategies(s string) []string {
	if s == "" {
		return []string{}
	}
	var strategies []string
	if err := json.Unmarshal([]byte(s), &strategies); err != nil {
		// Fallback: try comma-separated format
		return strings.Split(s, ",")
	}
	return strategies
}

const (
	SCHEDULED_DISCOVER_TASK_TABLE_NAME = "t_scheduled_discover_task"
)

var (
	sdtAccessOnce sync.Once
	sdtAccess     interfaces.ScheduledDiscoverTaskAccess
)

type scheduledDiscoverTaskAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// NewScheduledDiscoverTaskAccess creates a new ScheduledDiscoverTaskAccess.
func NewScheduledDiscoverTaskAccess(appSetting *common.AppSetting) interfaces.ScheduledDiscoverTaskAccess {
	sdtAccessOnce.Do(func() {
		sdtAccess = &scheduledDiscoverTaskAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return sdtAccess
}

func (sda *scheduledDiscoverTaskAccess) Enable(ctx context.Context, id string) error {
	_, span := ar_trace.Tracer.Start(ctx, "Enable scheduled_discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("task_id").String(id))

	// Get task to calculate next run time
	task, err := sda.GetByID(ctx, id)
	if err != nil {
		logger.Errorf("Failed to get scheduled discover task: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to get scheduled discover task: %v", err))
		span.SetStatus(codes.Error, "Get task failed")
		return err
	}

	// Calculate next run time from now
	nextRun, err := calculateNextRun(task.CronExpr, time.Now())
	if err != nil {
		logger.Errorf("Failed to calculate next run time: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to calculate next run time: %v", err))
		span.SetStatus(codes.Error, "Calculate next run failed")
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Build update SQL
	sqlStr, vals, err := sq.Update(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Set("f_enabled", 1).
		Set("f_next_run", nextRun.UnixMilli()).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build enable scheduled_discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build enable scheduled_discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Enable scheduled_discover_task SQL: %s", sqlStr))

	// Execute update
	_, err = sda.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Enable scheduled_discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Enable scheduled_discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Enable failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Enabled scheduled discover task: id=%s, next_run=%d", id, nextRun.UnixMilli())
	return nil
}

func (sda *scheduledDiscoverTaskAccess) Disable(ctx context.Context, id string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Disable scheduled_discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("task_id").String(id))

	// Build update SQL
	sqlStr, vals, err := sq.Update(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Set("f_enabled", 0).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build disable scheduled_discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build disable scheduled_discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Disable scheduled_discover_task SQL: %s", sqlStr))

	// Execute update
	_, err = sda.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Disable scheduled_discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Disable scheduled_discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Disable failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Disabled scheduled discover task: id=%s", id)
	return nil
}

func (sda *scheduledDiscoverTaskAccess) ExecuteTask(ctx context.Context, task *interfaces.ScheduledDiscoverTask) error {
	_, span := ar_trace.Tracer.Start(ctx, "Execute scheduled_discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("task_id").String(task.ID),
		attr.Key("catalog_id").String(task.CatalogID),
	)

	// ExecuteTask is implemented in the service layer, this is just a placeholder
	// The actual execution is handled by the DiscoverTaskService
	logger.Infof("Executing scheduled discover task: id=%s, catalog_id=%s", task.ID, task.CatalogID)

	span.SetStatus(codes.Ok, "")
	return nil
}

/**
 * 创建定时发现任务
 * @param ctx 上下文信息，用于追踪和传递请求范围的数据
 * @param task 定时发现任务结构体指针，包含任务的所有信息
 * @return error 执行结果，成功为nil，失败为错误信息
 */
func (sda *scheduledDiscoverTaskAccess) Create(ctx context.Context, task *interfaces.ScheduledDiscoverTask) error {
	// 使用OpenTelemetry追踪函数执行过程，创建一个客户端类型的span
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into t_scheduled_discover_task", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End() // 确保span在函数结束时结束
	// 设置span的属性，包含数据库URL和类型信息
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	// Generate ID if not provided
	if task.ID == "" {
		task.ID = xid.New().String()
	}

	// Set create time if not provided
	if task.CreateTime == 0 {
		task.CreateTime = time.Now().UnixMilli()
	}

	// Calculate next run time
	nextRun, err := calculateNextRun(task.CronExpr, time.Now())
	if err != nil {
		logger.Errorf("Failed to calculate next run time: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to calculate next run time: %v", err))
		span.SetStatus(codes.Error, "Calculate next run failed")
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	task.NextRun = nextRun.UnixMilli()

	// Build insert SQL
	strategiesStr := strategiesToString(task.Strategies)
	sqlStr, vals, err := sq.Insert(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Columns(
			"f_id",
			"f_catalog_id",
			"f_cron_expr",
			"f_start_time",
			"f_end_time",
			"f_enabled",
			"f_strategies",
			"f_last_run",
			"f_next_run",
			"f_creator_id",
			"f_creator_type",
			"f_create_time",
		).
		Values(
			task.ID,
			task.CatalogID,
			task.CronExpr,
			task.StartTime,
			task.EndTime,
			task.Enabled,
			strategiesStr,
			task.LastRun,
			task.NextRun,
			task.Creator.ID,
			task.Creator.Type,
			task.CreateTime,
		).ToSql()
	if err != nil {
		logger.Errorf("Failed to build insert scheduled_discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build insert scheduled_discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Insert scheduled_discover_task SQL: %s", sqlStr))

	// Execute insert
	_, err = sda.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Insert scheduled_discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Insert scheduled_discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Insert failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Created scheduled discover task: id=%s, catalog_id=%s, cron=%s", task.ID, task.CatalogID, task.CronExpr)
	return nil
}

// GetByID retrieves a scheduled discover task by ID.
func (sda *scheduledDiscoverTaskAccess) GetByID(ctx context.Context, id string) (*interfaces.ScheduledDiscoverTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query scheduled_discover_task by ID",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("task_id").String(id))

	// Build select SQL
	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_cron_expr",
		"f_start_time",
		"f_end_time",
		"f_enabled",
		"f_strategies",
		"f_last_run",
		"f_next_run",
		"f_creator_id",
		"f_creator_type",
		"f_create_time",
		"f_updater_id",
		"f_updater_type",
		"f_update_time",
	).From(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select scheduled_discover_task sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	task := &interfaces.ScheduledDiscoverTask{}
	var strategiesStr string

	// Execute query
	row := sda.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&task.ID,
		&task.CatalogID,
		&task.CronExpr,
		&task.StartTime,
		&task.EndTime,
		&task.Enabled,
		&strategiesStr,
		&task.LastRun,
		&task.NextRun,
		&task.Creator.ID,
		&task.Creator.Type,
		&task.CreateTime,
		&task.Updater.ID,
		&task.Updater.Type,
		&task.UpdateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan scheduled_discover_task failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// Parse strategies string to array
	task.Strategies = stringToStrategies(strategiesStr)

	span.SetStatus(codes.Ok, "")
	return task, nil
}

// List lists scheduled discover tasks with filters.
func (sda *scheduledDiscoverTaskAccess) List(ctx context.Context, params interfaces.ScheduledDiscoverTaskQueryParams) ([]*interfaces.ScheduledDiscoverTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List scheduled_discover_tasks",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// Build select query
	builder := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_cron_expr",
		"f_start_time",
		"f_end_time",
		"f_enabled",
		"f_strategies",
		"f_last_run",
		"f_next_run",
		"f_creator_id",
		"f_creator_type",
		"f_create_time",
		"f_updater_id",
		"f_updater_type",
		"f_update_time",
	).From(SCHEDULED_DISCOVER_TASK_TABLE_NAME)

	// Apply filters
	if params.CatalogID != "" {
		builder = builder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.Enabled != nil {
		builder = builder.Where(sq.Eq{"f_enabled": *params.Enabled})
	}

	// Get total count
	countBuilder := sq.Select("COUNT(*)").From(SCHEDULED_DISCOVER_TASK_TABLE_NAME)
	if params.CatalogID != "" {
		countBuilder = countBuilder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.Enabled != nil {
		countBuilder = countBuilder.Where(sq.Eq{"f_enabled": *params.Enabled})
	}

	countSql, countVals, err := countBuilder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build count scheduled_discover_task sql: %v", err)
		span.SetStatus(codes.Error, "Build count sql failed")
		return nil, 0, err
	}

	var total int64
	err = sda.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Count scheduled_discover_task failed: %v", err)
		span.SetStatus(codes.Error, "Count failed")
		return nil, 0, err
	}

	// Apply ordering and pagination
	builder = builder.OrderBy("f_create_time DESC")
	// Pagination
	if params.Limit > 0 {
		builder = builder.Limit(uint64(params.Limit)).Offset(uint64(params.Offset))
	}
	// Build query
	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build select scheduled_discover_task sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, 0, err
	}

	// Execute query
	rows, err := sda.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query scheduled_discover_task failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	tasks := []*interfaces.ScheduledDiscoverTask{}
	for rows.Next() {
		task := &interfaces.ScheduledDiscoverTask{}
		var strategiesStr string
		err := rows.Scan(
			&task.ID,
			&task.CatalogID,
			&task.CronExpr,
			&task.StartTime,
			&task.EndTime,
			&task.Enabled,
			&strategiesStr,
			&task.LastRun,
			&task.NextRun,
			&task.Creator.ID,
			&task.Creator.Type,
			&task.CreateTime,
			&task.Updater.ID,
			&task.Updater.Type,
			&task.UpdateTime,
		)
		if err != nil {
			logger.Errorf("Scan scheduled_discover_task failed: %v", err)
			span.SetStatus(codes.Error, "Scan failed")
			return nil, 0, err
		}
		// Parse strategies string to array
		task.Strategies = stringToStrategies(strategiesStr)
		tasks = append(tasks, task)
	}

	span.SetStatus(codes.Ok, "")
	return tasks, total, nil
}

// Update updates a scheduled discover task.
func (sda *scheduledDiscoverTaskAccess) Update(ctx context.Context, task *interfaces.ScheduledDiscoverTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update scheduled_discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("task_id").String(task.ID))

	// Set update time
	task.UpdateTime = time.Now().UnixMilli()

	// Recalculate next run time if cron expression changed
	nextRun, err := calculateNextRun(task.CronExpr, time.Now())
	if err != nil {
		logger.Errorf("Failed to calculate next run time: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to calculate next run time: %v", err))
		span.SetStatus(codes.Error, "Calculate next run failed")
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	task.NextRun = nextRun.UnixMilli()

	// Build update SQL - only update non-zero value fields
	updateBuilder := sq.Update(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Where(sq.Eq{"f_id": task.ID})

	// Only update fields that are explicitly set
	if task.CatalogID != "" {
		updateBuilder = updateBuilder.Set("f_catalog_id", task.CatalogID)
	}
	if task.CronExpr != "" {
		updateBuilder = updateBuilder.Set("f_cron_expr", task.CronExpr)
	}
	if task.StartTime != 0 {
		updateBuilder = updateBuilder.Set("f_start_time", task.StartTime)
	}
	if task.EndTime != 0 {
		updateBuilder = updateBuilder.Set("f_end_time", task.EndTime)
	}
	if len(task.Strategies) > 0 {
		strategiesStr := strategiesToString(task.Strategies)
		updateBuilder = updateBuilder.Set("f_strategies", strategiesStr)
	}
	// Always update NextRun when CronExpr changes
	updateBuilder = updateBuilder.Set("f_next_run", task.NextRun)
	// update f_enabled
	var enableFlag int
	if !task.Enabled {
		enableFlag = 0
	} else {
		enableFlag = 1
	}
	updateBuilder = updateBuilder.
		Set("f_updater_id", task.Updater.ID).
		Set("f_updater_type", task.Updater.Type).
		Set("f_update_time", task.UpdateTime).
		Set("f_enabled", enableFlag)

	sqlStr, vals, err := updateBuilder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build update scheduled_discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build update scheduled_discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Update scheduled_discover_task SQL: %s", sqlStr))

	// Execute update
	result, err := sda.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Update scheduled_discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Update scheduled_discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("Failed to get rows affected: %v", err)
		span.SetStatus(codes.Error, "Get rows affected failed")
		return err
	}
	if rowsAffected == 0 {
		logger.Warnf("No rows affected when updating scheduled_discover_task: id=%s", task.ID)
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Updated scheduled discover task: id=%s", task.ID)
	return nil
}

// Delete deletes a scheduled discover task by ID.
func (sda *scheduledDiscoverTaskAccess) Delete(ctx context.Context, id string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete scheduled_discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("task_id").String(id))

	// Build delete SQL
	sqlStr, vals, err := sq.Delete(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build delete scheduled_discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build delete scheduled_discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Delete scheduled_discover_task SQL: %s", sqlStr))

	// Execute delete
	result, err := sda.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Delete scheduled_discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Delete scheduled_discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("Failed to get rows affected: %v", err)
		span.SetStatus(codes.Error, "Get rows affected failed")
		return err
	}
	if rowsAffected == 0 {
		logger.Warnf("No rows affected when deleting scheduled_discover_task: id=%s", id)
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Deleted scheduled discover task: id=%s", id)
	return nil
}

// GetEnabledTasks retrieves all enabled scheduled discover tasks.
func (sda *scheduledDiscoverTaskAccess) GetEnabledTasks(ctx context.Context) ([]*interfaces.ScheduledDiscoverTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query enabled scheduled_discover_tasks",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	now := time.Now().UnixMilli()

	// Build select SQL
	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_cron_expr",
		"f_start_time",
		"f_end_time",
		"f_enabled",
		"f_last_run",
		"f_next_run",
		"f_creator_id",
		"f_creator_type",
		"f_create_time",
		"f_updater_id",
		"f_updater_type",
		"f_update_time",
	).From(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Where(sq.Eq{"f_enabled": true}).
		Where(sq.Or{
			sq.Eq{"f_end_time": 0},
			sq.Gt{"f_end_time": now},
		}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select enabled scheduled_discover_task sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	// Execute query
	rows, err := sda.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query enabled scheduled_discover_task failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	tasks := []*interfaces.ScheduledDiscoverTask{}
	for rows.Next() {
		task := &interfaces.ScheduledDiscoverTask{}
		err := rows.Scan(
			&task.ID,
			&task.CatalogID,
			&task.CronExpr,
			&task.StartTime,
			&task.EndTime,
			&task.Enabled,
			&task.LastRun,
			&task.NextRun,
			&task.Creator.ID,
			&task.Creator.Type,
			&task.CreateTime,
			&task.Updater.ID,
			&task.Updater.Type,
			&task.UpdateTime,
		)
		if err != nil {
			logger.Errorf("Scan scheduled_discover_task failed: %v", err)
			span.SetStatus(codes.Error, "Scan failed")
			return nil, err
		}
		tasks = append(tasks, task)
	}

	span.SetStatus(codes.Ok, "")
	return tasks, nil
}

// UpdateLastRun updates the last run time and calculates next run time.
func (sda *scheduledDiscoverTaskAccess) UpdateLastRun(ctx context.Context, id string, lastRun int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update last run for scheduled_discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// Get task to calculate next run
	task, err := sda.GetByID(ctx, id)
	if err != nil {
		logger.Errorf("Failed to get scheduled discover task: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to get scheduled discover task: %v", err))
		span.SetStatus(codes.Error, "Get task failed")
		return err
	}

	// Calculate next run time
	nextRun, err := calculateNextRun(task.CronExpr, time.UnixMilli(lastRun))
	if err != nil {
		logger.Errorf("Failed to calculate next run time: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to calculate next run time: %v", err))
		span.SetStatus(codes.Error, "Calculate next run failed")
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	span.SetAttributes(
		attr.Key("task_id").String(id),
		attr.Key("last_run").Int64(lastRun),
		attr.Key("next_run").Int64(nextRun.UnixMilli()),
	)

	// Build update SQL
	sqlStr, vals, err := sq.Update(SCHEDULED_DISCOVER_TASK_TABLE_NAME).
		Set("f_last_run", lastRun).
		Set("f_next_run", nextRun.UnixMilli()).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build update last run scheduled_discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build update last run scheduled_discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Update last run scheduled_discover_task SQL: %s", sqlStr))

	// Execute update
	result, err := sda.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Update last run scheduled_discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Update last run scheduled_discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("Failed to get rows affected: %v", err)
		span.SetStatus(codes.Error, "Get rows affected failed")
		return err
	}
	if rowsAffected == 0 {
		logger.Warnf("No rows affected when updating last run for scheduled_discover_task: id=%s", id)
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Updated last run time for scheduled discover task: id=%s, last_run=%d, next_run=%d", id, lastRun, nextRun.UnixMilli())
	return nil
}

// calculateNextRun calculates the next run time based on cron expression.
// calculateNextRun 计算给定的cron表达式从指定时间开始的下一次运行时间
// 参数:
//
//	cronExpr: cron表达式字符串，用于定义定时任务的执行规则
//	from: 起始时间，从此时间点开始计算下一次执行时间
//
// 返回值:
//
//	time.Time: 计算得到的下一次运行时间
//	error: 如果cron表达式无效，则返回错误信息
func calculateNextRun(cronExpr string, from time.Time) (time.Time, error) {
	// Parse cron expression
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	// Get next run time
	nextRun := schedule.Next(from)
	return nextRun, nil
}
