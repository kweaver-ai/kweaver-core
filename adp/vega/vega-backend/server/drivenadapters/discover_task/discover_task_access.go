// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package discover_task provides DiscoverTask data access operations.
package discover_task

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	"vega-backend/interfaces"
)

const (
	DISCOVER_TASK_TABLE_NAME = "t_discover_task"
)

var (
	dtAccessOnce sync.Once
	dtAccess     interfaces.DiscoverTaskAccess
)

type discoverTaskAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// GetScheduledTaskStrategies retrieves strategies from t_scheduled_discover_task table by ID.
func (da *discoverTaskAccess) GetScheduledTaskStrategies(ctx context.Context, scheduledTaskID string) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query scheduled_discover_task by ID",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	sqlStr, vals, err := sq.Select("f_strategies").
		From("t_scheduled_discover_task").
		Where(sq.Eq{"f_id": scheduledTaskID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select scheduled_discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build select scheduled_discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	var strategiesStr string
	err = da.db.QueryRowContext(ctx, sqlStr, vals...).Scan(&strategiesStr)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return []string{}, nil
	}
	if err != nil {
		logger.Errorf("Scan scheduled_discover_task failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// Parse strategies string to array
	if strategiesStr == "" {
		return []string{}, nil
	}
	var strategies []string
	if err := sonic.Unmarshal([]byte(strategiesStr), &strategies); err != nil {
		logger.Errorf("Failed to unmarshal strategies: %v", err)
		return []string{}, nil
	}

	span.SetStatus(codes.Ok, "")
	return strategies, nil
}

// NewDiscoverTaskAccess creates a new DiscoverTaskAccess.
func NewDiscoverTaskAccess(appSetting *common.AppSetting) interfaces.DiscoverTaskAccess {
	dtAccessOnce.Do(func() {
		dtAccess = &discoverTaskAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return dtAccess
}

// Create creates a new DiscoverTask.
func (da *discoverTaskAccess) Create(ctx context.Context, task *interfaces.DiscoverTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	// Serialize strategies to JSON string
	strategiesStr := ""
	if len(task.Strategies) > 0 {
		strategiesBytes, err := sonic.Marshal(task.Strategies)
		if err != nil {
			logger.Errorf("Failed to marshal strategies: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to marshal strategies: %v", err))
			span.SetStatus(codes.Error, "Marshal strategies failed")
			return err
		}
		strategiesStr = string(strategiesBytes)
	}

	sqlStr, vals, err := sq.Insert(DISCOVER_TASK_TABLE_NAME).
		Columns(
			"f_id",
			"f_catalog_id",
			"f_scheduled_id",
			"f_strategies",
			"f_trigger_type",
			"f_status",
			"f_progress",
			"f_message",
			"f_start_time",
			"f_finish_time",
			"f_result",
			"f_creator",
			"f_creator_type",
			"f_create_time",
		).
		Values(
			task.ID,
			task.CatalogID,
			task.ScheduleID,
			strategiesStr,
			task.TriggerType,
			task.Status,
			task.Progress,
			task.Message,
			task.StartTime,
			task.FinishTime,
			"", // result initially empty
			task.Creator.ID,
			task.Creator.Type,
			task.CreateTime,
		).ToSql()
	if err != nil {
		logger.Errorf("Failed to build insert discover_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build insert discover_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Insert discover_task SQL: %s", sqlStr))

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Insert discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Insert discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Insert failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByID retrieves a DiscoverTask by ID.
func (da *discoverTaskAccess) GetByID(ctx context.Context, id string) (*interfaces.DiscoverTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query discover_task by ID",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("task_id").String(id))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_scheduled_id",
		"f_strategies",
		"f_trigger_type",
		"f_status",
		"f_progress",
		"f_message",
		"f_start_time",
		"f_finish_time",
		"f_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
	).From(DISCOVER_TASK_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select discover_task sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	task := &interfaces.DiscoverTask{}
	var resultStr sql.NullString
	var strategiesStr sql.NullString

	row := da.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&task.ID,
		&task.CatalogID,
		&task.ScheduleID,
		&strategiesStr,
		&task.TriggerType,
		&task.Status,
		&task.Progress,
		&task.Message,
		&task.StartTime,
		&task.FinishTime,
		&resultStr,
		&task.Creator.ID,
		&task.Creator.Type,
		&task.CreateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan discover_task failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// Parse strategies string to array
	if strategiesStr.Valid && strategiesStr.String != "" {
		if err := sonic.Unmarshal([]byte(strategiesStr.String), &task.Strategies); err != nil {
			logger.Errorf("Failed to unmarshal strategies: %v", err)
			task.Strategies = []string{} // Set to empty array on error
		}
	} else {
		task.Strategies = []string{}
	}

	// Deserialize result
	if resultStr.Valid && resultStr.String != "" {
		task.Result = &interfaces.DiscoverResult{}
		_ = sonic.UnmarshalString(resultStr.String, task.Result)
	}

	span.SetStatus(codes.Ok, "")
	return task, nil
}

// List lists DiscoverTasks with filters.
func (da *discoverTaskAccess) List(ctx context.Context, params interfaces.DiscoverTaskQueryParams) ([]*interfaces.DiscoverTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List discover_tasks",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	builder := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_scheduled_id",
		"f_strategies",
		"f_trigger_type",
		"f_status",
		"f_progress",
		"f_message",
		"f_start_time",
		"f_finish_time",
		"f_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
	).From(DISCOVER_TASK_TABLE_NAME)

	countBuilder := sq.Select("COUNT(*)").From(DISCOVER_TASK_TABLE_NAME)

	if params.CatalogID != "" {
		builder = builder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
		countBuilder = countBuilder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.ScheduleID != "" {
		builder = builder.Where(sq.Eq{"f_scheduled_id": params.ScheduleID})
		countBuilder = countBuilder.Where(sq.Eq{"f_scheduled_id": params.ScheduleID})
	}
	if params.Status != "" {
		builder = builder.Where(sq.Eq{"f_status": params.Status})
		countBuilder = countBuilder.Where(sq.Eq{"f_status": params.Status})
	}
	if params.TriggerType != "" {
		builder = builder.Where(sq.Eq{"f_trigger_type": params.TriggerType})
		countBuilder = countBuilder.Where(sq.Eq{"f_trigger_type": params.TriggerType})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := da.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count discover_tasks: %v", err)
		span.SetStatus(codes.Error, "Count failed")
		return nil, 0, err
	}

	// Pagination
	if params.Limit > 0 {
		builder = builder.Limit(uint64(params.Limit)).Offset(uint64(params.Offset))
	}
	builder = builder.OrderBy("f_create_time DESC")

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, 0, err
	}

	rows, err := da.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	tasks := make([]*interfaces.DiscoverTask, 0)
	for rows.Next() {
		task := &interfaces.DiscoverTask{}
		var resultStr sql.NullString
		var strategiesStr sql.NullString

		err := rows.Scan(
			&task.ID,
			&task.CatalogID,
			&task.ScheduleID,
			&strategiesStr,
			&task.TriggerType,
			&task.Status,
			&task.Progress,
			&task.Message,
			&task.StartTime,
			&task.FinishTime,
			&resultStr,
			&task.Creator.ID,
			&task.Creator.Type,
			&task.CreateTime,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, 0, err
		}

		// Parse strategies string to array
		if strategiesStr.Valid && strategiesStr.String != "" {
			if err := sonic.Unmarshal([]byte(strategiesStr.String), &task.Strategies); err != nil {
				logger.Errorf("Failed to unmarshal strategies: %v", err)
				task.Strategies = []string{} // Set to empty array on error
			}
		} else {
			task.Strategies = []string{}
		}

		// Deserialize result
		if resultStr.Valid && resultStr.String != "" {
			task.Result = &interfaces.DiscoverResult{}
			_ = sonic.UnmarshalString(resultStr.String, task.Result)
		}

		tasks = append(tasks, task)
	}

	span.SetStatus(codes.Ok, "")
	return tasks, total, nil
}

// UpdateStatus updates a DiscoverTask's status and message.
func (da *discoverTaskAccess) UpdateStatus(ctx context.Context, id, status, message string, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update discover_task status",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("task_id").String(id),
		attr.Key("status").String(status),
	)

	data := map[string]any{
		"f_status":  status,
		"f_message": message,
	}
	if status == interfaces.DiscoverTaskStatusRunning {
		data["f_start_time"] = stime
	}
	sqlStr, vals, err := sq.Update(DISCOVER_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateProgress updates a DiscoverTask's progress.
func (da *discoverTaskAccess) UpdateProgress(ctx context.Context, id string, progress int) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update discover_task progress",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	sqlStr, vals, err := sq.Update(DISCOVER_TASK_TABLE_NAME).
		Set("f_progress", progress).
		Set("f_update_time", time.Now().UnixMilli()).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateResult updates a DiscoverTask's result and sets status to completed.
func (da *discoverTaskAccess) UpdateResult(ctx context.Context, id string, result *interfaces.DiscoverResult, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update discover_task result",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	resultBytes, _ := sonic.MarshalString(result)

	sqlStr, vals, err := sq.Update(DISCOVER_TASK_TABLE_NAME).
		Set("f_status", interfaces.DiscoverTaskStatusCompleted).
		Set("f_result", resultBytes).
		Set("f_progress", 100).
		Set("f_finish_time", stime).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExistByStatuses checks if DiscoverTasks exist by catalog ID and statuses.
func (da *discoverTaskAccess) CheckExistByStatuses(ctx context.Context, catalogID string, statuses []string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check discover_tasks exist",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	countBuilder := sq.Select("COUNT(*)").From(DISCOVER_TASK_TABLE_NAME)

	if catalogID != "" {
		countBuilder = countBuilder.Where(sq.Eq{"f_catalog_id": catalogID})
	}
	if len(statuses) > 0 {
		countBuilder = countBuilder.Where(sq.Eq{"f_status": statuses})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := da.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count discover_tasks: %v", err)
		span.SetStatus(codes.Error, "Count failed")
		return false, err
	}

	span.SetStatus(codes.Ok, "")
	return total > 0, nil
}

// Delete deletes a DiscoverTask by ID. Returns sql.ErrNoRows if no row was affected.
func (da *discoverTaskAccess) Delete(ctx context.Context, id string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete discover_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("id").String(id))

	sqlStr, vals, err := sq.Delete(DISCOVER_TASK_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build delete discover_task sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	res, err := da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Delete discover_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Delete discover_task failed: %v", err))
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		span.SetStatus(codes.Error, "RowsAffected failed")
		return err
	}
	if affected == 0 {
		span.SetStatus(codes.Ok, "discover_task not found")
		return sql.ErrNoRows
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
