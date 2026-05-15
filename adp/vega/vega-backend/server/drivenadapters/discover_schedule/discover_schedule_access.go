// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package discover_schedule

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/robfig/cron/v3"
	_ "github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	"vega-backend/interfaces"
)

// Helper functions for strategies array handling
func strategiesToString(strategies []string) string {
	if len(strategies) == 0 {
		return ""
	}
	data, err := sonic.Marshal(strategies)
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
	if err := sonic.Unmarshal([]byte(s), &strategies); err != nil {
		// Fallback: try comma-separated format
		return strings.Split(s, ",")
	}
	return strategies
}

const (
	DISCOVER_SCHEDULE_TABLE_NAME = "t_discover_schedule"
)

var (
	dsAccessOnce sync.Once
	dsAccess     interfaces.DiscoverScheduleAccess
)

type discoverScheduleAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// NewDiscoverScheduleAccess creates a new DiscoverScheduleAccess.
func NewDiscoverScheduleAccess(appSetting *common.AppSetting) interfaces.DiscoverScheduleAccess {
	dsAccessOnce.Do(func() {
		dsAccess = &discoverScheduleAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return dsAccess
}

func (dsa *discoverScheduleAccess) Enable(ctx context.Context, id string) error {
	_, span := oteltrace.StartNamedClientSpan(ctx, "Enable discover_schedule")
	defer span.End()

	span.SetAttributes(attr.Key("schedule_id").String(id))

	// Get schedule to calculate next run time
	schedule, err := dsa.GetByID(ctx, id)
	if err != nil {
		otellog.LogError(ctx, "Failed to get discover schedule", err)
		return err
	}

	// Calculate next run time from now
	nextRun, err := calculateNextRun(schedule.CronExpr, time.Now())
	if err != nil {
		otellog.LogError(ctx, "Failed to calculate next run time", err)
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	// Build update SQL
	sqlStr, vals, err := sq.Update(DISCOVER_SCHEDULE_TABLE_NAME).
		Set("f_enabled", 1).
		Set("f_next_run", nextRun.UnixMilli()).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build enable discover_schedule sql", err)
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Enable discover_schedule SQL: %s", sqlStr))

	// Execute update
	_, err = dsa.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Enable discover_schedule failed", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Enabled discover schedule: id=%s, next_run=%d", id, nextRun.UnixMilli())
	return nil
}

func (dsa *discoverScheduleAccess) Disable(ctx context.Context, id string) error {
	_, span := oteltrace.StartNamedClientSpan(ctx, "Disable discover_schedule")
	defer span.End()

	span.SetAttributes(attr.Key("schedule_id").String(id))

	// Build update SQL
	sqlStr, vals, err := sq.Update(DISCOVER_SCHEDULE_TABLE_NAME).
		Set("f_enabled", 0).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build disable discover_schedule sql", err)
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Disable discover_schedule SQL: %s", sqlStr))

	// Execute update
	_, err = dsa.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Disable discover_schedule failed", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Disabled discover schedule: id=%s", id)
	return nil
}

/**
 * 创建定时发现任务
 * @param ctx 上下文信息，用于追踪和传递请求范围的数据
 * @param schedule 定时发现任务结构体指针，包含任务的所有信息
 * @return error 执行结果，成功为nil，失败为错误信息
 */
func (dsa *discoverScheduleAccess) Create(ctx context.Context, schedule *interfaces.DiscoverSchedule) error {
	// 使用OpenTelemetry追踪函数执行过程，创建一个客户端类型的span
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Insert into t_discover_schedule")
	defer span.End() // 确保span在函数结束时结束
	// 设置span的属性，包含数据库URL和类型信息
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	// Calculate next run time
	nextRun, err := calculateNextRun(schedule.CronExpr, time.Now())
	if err != nil {
		otellog.LogError(ctx, "Failed to calculate next run time", err)
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	schedule.NextRun = nextRun.UnixMilli()

	// Build insert SQL
	strategiesStr := strategiesToString(schedule.Strategies)
	sqlStr, vals, err := sq.Insert(DISCOVER_SCHEDULE_TABLE_NAME).
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
			"f_creator",
			"f_creator_type",
			"f_create_time",
		).
		Values(
			schedule.ID,
			schedule.CatalogID,
			schedule.CronExpr,
			schedule.StartTime,
			schedule.EndTime,
			schedule.Enabled,
			strategiesStr,
			schedule.LastRun,
			schedule.NextRun,
			schedule.Creator.ID,
			schedule.Creator.Type,
			schedule.CreateTime,
		).ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build insert discover_schedule sql", err)
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Insert discover_schedule SQL: %s", sqlStr))

	// Execute insert
	_, err = dsa.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Insert discover_schedule failed", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Created discover schedule: id=%s, catalog_id=%s, cron=%s", schedule.ID, schedule.CatalogID, schedule.CronExpr)
	return nil
}

// GetByID retrieves a discover schedule by ID.
func (dsa *discoverScheduleAccess) GetByID(ctx context.Context, id string) (*interfaces.DiscoverSchedule, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Query discover_schedule by ID")
	defer span.End()

	span.SetAttributes(attr.Key("schedule_id").String(id))

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
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(DISCOVER_SCHEDULE_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select discover_schedule sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	schedule := &interfaces.DiscoverSchedule{}
	var strategiesStr string

	// Execute query
	row := dsa.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&schedule.ID,
		&schedule.CatalogID,
		&schedule.CronExpr,
		&schedule.StartTime,
		&schedule.EndTime,
		&schedule.Enabled,
		&strategiesStr,
		&schedule.LastRun,
		&schedule.NextRun,
		&schedule.Creator.ID,
		&schedule.Creator.Type,
		&schedule.CreateTime,
		&schedule.Updater.ID,
		&schedule.Updater.Type,
		&schedule.UpdateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan discover_schedule failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// Parse strategies string to array
	schedule.Strategies = stringToStrategies(strategiesStr)

	span.SetStatus(codes.Ok, "")
	return schedule, nil
}

// List lists discover schedules with filters.
func (dsa *discoverScheduleAccess) List(ctx context.Context, params interfaces.DiscoverScheduleQueryParams) ([]*interfaces.DiscoverSchedule, int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "List discover_schedules")
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
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(DISCOVER_SCHEDULE_TABLE_NAME)

	// Apply filters
	if params.CatalogID != "" {
		builder = builder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.Enabled != nil {
		builder = builder.Where(sq.Eq{"f_enabled": *params.Enabled})
	}

	// Get total count
	countBuilder := sq.Select("COUNT(*)").From(DISCOVER_SCHEDULE_TABLE_NAME)
	if params.CatalogID != "" {
		countBuilder = countBuilder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.Enabled != nil {
		countBuilder = countBuilder.Where(sq.Eq{"f_enabled": *params.Enabled})
	}

	countSql, countVals, err := countBuilder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build count discover_schedule sql: %v", err)
		span.SetStatus(codes.Error, "Build count sql failed")
		return nil, 0, err
	}

	var total int64
	err = dsa.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Count discover_schedule failed: %v", err)
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
		logger.Errorf("Failed to build select discover_schedule sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, 0, err
	}

	// Execute query
	rows, err := dsa.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query discover_schedule failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	schedules := []*interfaces.DiscoverSchedule{}
	for rows.Next() {
		schedule := &interfaces.DiscoverSchedule{}
		var strategiesStr string
		err := rows.Scan(
			&schedule.ID,
			&schedule.CatalogID,
			&schedule.CronExpr,
			&schedule.StartTime,
			&schedule.EndTime,
			&schedule.Enabled,
			&strategiesStr,
			&schedule.LastRun,
			&schedule.NextRun,
			&schedule.Creator.ID,
			&schedule.Creator.Type,
			&schedule.CreateTime,
			&schedule.Updater.ID,
			&schedule.Updater.Type,
			&schedule.UpdateTime,
		)
		if err != nil {
			logger.Errorf("Scan discover_schedule failed: %v", err)
			span.SetStatus(codes.Error, "Scan failed")
			return nil, 0, err
		}
		// Parse strategies string to array
		schedule.Strategies = stringToStrategies(strategiesStr)
		schedules = append(schedules, schedule)
	}

	span.SetStatus(codes.Ok, "")
	return schedules, total, nil
}

// Update updates a discover schedule.
func (dsa *discoverScheduleAccess) Update(ctx context.Context, schedule *interfaces.DiscoverSchedule) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Update discover_schedule")
	defer span.End()

	span.SetAttributes(attr.Key("schedule_id").String(schedule.ID))

	strategiesStr := strategiesToString(schedule.Strategies)

	// Recalculate next run time if cron expression changed
	nextRun, err := calculateNextRun(schedule.CronExpr, time.Now())
	if err != nil {
		otellog.LogError(ctx, "Failed to calculate next run time", err)
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	schedule.NextRun = nextRun.UnixMilli()

	// Build update SQL - only update non-zero value fields
	updateBuilder := sq.Update(DISCOVER_SCHEDULE_TABLE_NAME).
		Set("f_catalog_id", schedule.CatalogID).
		Set("f_cron_expr", schedule.CronExpr).
		Set("f_start_time", schedule.StartTime).
		Set("f_end_time", schedule.EndTime).
		Set("f_strategies", strategiesStr).
		Set("f_next_run", schedule.NextRun).
		Set("f_enabled", schedule.Enabled).
		Set("f_updater", schedule.Updater.ID).
		Set("f_updater_type", schedule.Updater.Type).
		Set("f_update_time", schedule.UpdateTime).
		Where(sq.Eq{"f_id": schedule.ID})

	sqlStr, vals, err := updateBuilder.ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build update discover_schedule sql", err)
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Update discover_schedule SQL: %s", sqlStr))

	// Execute update
	result, err := dsa.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Update discover_schedule failed", err)
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
		logger.Warnf("No rows affected when updating discover_schedule: id=%s", schedule.ID)
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Updated discover_schedule: id=%s", schedule.ID)
	return nil
}

// Delete deletes a discover schedule by ID.
func (dsa *discoverScheduleAccess) Delete(ctx context.Context, id string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Delete discover_schedule")
	defer span.End()

	span.SetAttributes(attr.Key("schedule_id").String(id))

	// Build delete SQL
	sqlStr, vals, err := sq.Delete(DISCOVER_SCHEDULE_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build delete discover_schedule sql", err)
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Delete discover_schedule SQL: %s", sqlStr))

	// Execute delete
	result, err := dsa.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Delete discover_schedule failed", err)
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
		logger.Warnf("No rows affected when deleting discover_schedule: id=%s", id)
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Deleted discover_schedule: id=%s", id)
	return nil
}

// GetEnabledSchedules retrieves all enabled discover schedules.
func (dsa *discoverScheduleAccess) GetEnabledSchedules(ctx context.Context) ([]*interfaces.DiscoverSchedule, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Query enabled discover_schedules")
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
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(DISCOVER_SCHEDULE_TABLE_NAME).
		Where(sq.Eq{"f_enabled": true}).
		Where(sq.Or{
			sq.Eq{"f_end_time": 0},
			sq.Gt{"f_end_time": now},
		}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select enabled discover_schedule sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	// Execute query
	rows, err := dsa.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query enabled discover_schedule failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	schedules := []*interfaces.DiscoverSchedule{}
	for rows.Next() {
		schedule := &interfaces.DiscoverSchedule{}
		err := rows.Scan(
			&schedule.ID,
			&schedule.CatalogID,
			&schedule.CronExpr,
			&schedule.StartTime,
			&schedule.EndTime,
			&schedule.Enabled,
			&schedule.LastRun,
			&schedule.NextRun,
			&schedule.Creator.ID,
			&schedule.Creator.Type,
			&schedule.CreateTime,
			&schedule.Updater.ID,
			&schedule.Updater.Type,
			&schedule.UpdateTime,
		)
		if err != nil {
			logger.Errorf("Scan discover_schedule failed: %v", err)
			span.SetStatus(codes.Error, "Scan failed")
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	span.SetStatus(codes.Ok, "")
	return schedules, nil
}

// UpdateLastRun updates the last run time and calculates next run time.
func (dsa *discoverScheduleAccess) UpdateLastRun(ctx context.Context, id string, lastRun int64) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Update last run for discover_schedule")
	defer span.End()

	// Get schedule to calculate next run
	schedule, err := dsa.GetByID(ctx, id)
	if err != nil {
		otellog.LogError(ctx, "Failed to get discover schedule", err)
		return err
	}

	// Calculate next run time
	nextRun, err := calculateNextRun(schedule.CronExpr, time.UnixMilli(lastRun))
	if err != nil {
		otellog.LogError(ctx, "Failed to calculate next run time", err)
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	span.SetAttributes(
		attr.Key("schedule_id").String(id),
		attr.Key("last_run").Int64(lastRun),
		attr.Key("next_run").Int64(nextRun.UnixMilli()),
	)

	// Build update SQL
	sqlStr, vals, err := sq.Update(DISCOVER_SCHEDULE_TABLE_NAME).
		Set("f_last_run", lastRun).
		Set("f_next_run", nextRun.UnixMilli()).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build update last run discover_schedule sql", err)
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Update last run discover_schedule SQL: %s", sqlStr))

	// Execute update
	result, err := dsa.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Update last run discover_schedule failed", err)
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
		logger.Warnf("No rows affected when updating last run for discover_schedule: id=%s", id)
	}

	span.SetStatus(codes.Ok, "")
	logger.Infof("Updated last run time for discover_schedule: id=%s, last_run=%d, next_run=%d", id, lastRun, nextRun.UnixMilli())
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
