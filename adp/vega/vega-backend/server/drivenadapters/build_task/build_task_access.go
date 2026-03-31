// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package build_task provides BuildTask data access layer.
package build_task

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	btaOnce sync.Once
	bta     interfaces.BuildTaskAccess
)

const (
	BUILD_TASK_TABLE_NAME = "t_build_task"
)

type buildTaskAccess struct {
	db *sql.DB
}

// NewBuildTaskAccess creates a new BuildTaskAccess.
func NewBuildTaskAccess(appSetting *common.AppSetting) interfaces.BuildTaskAccess {
	btaOnce.Do(func() {
		bta = &buildTaskAccess{
			db: db.NewDB(&appSetting.DBSetting),
		}
	})
	return bta
}

// Create creates a new build task.
func (bta *buildTaskAccess) Create(ctx context.Context, buildTask *interfaces.BuildTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create build task", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		INSERT INTO ` + BUILD_TASK_TABLE_NAME + ` (
			f_id, f_resource_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := bta.db.ExecContext(
		ctx,
		query,
		buildTask.ID,
		buildTask.ResourceID,
		buildTask.Status,
		buildTask.Mode,
		buildTask.TotalCount,
		buildTask.SyncedCount,
		buildTask.VectorizedCount,
		buildTask.SyncedMark,
		buildTask.ErrorMsg,
		buildTask.Creator.ID,
		buildTask.Creator.Type,
		buildTask.CreateTime,
		buildTask.Updater.ID,
		buildTask.Updater.Type,
		buildTask.UpdateTime,
	)

	if err != nil {
		logger.Errorf("Create build task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create build task failed: %v", err))
		span.SetStatus(codes.Error, "Create build task failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByID retrieves a build task by ID.
func (bta *buildTaskAccess) GetByID(ctx context.Context, id string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build task by ID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT 
			f_id, f_resource_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_id = ?
	`

	row := bta.db.QueryRowContext(ctx, query, id)

	buildTask := &interfaces.BuildTask{}
	var creatorID, creatorType, updaterID, updaterType string

	err := row.Scan(
		&buildTask.ID,
		&buildTask.ResourceID,
		&buildTask.Status,
		&buildTask.Mode,
		&buildTask.TotalCount,
		&buildTask.SyncedCount,
		&buildTask.VectorizedCount,
		&buildTask.SyncedMark,
		&buildTask.ErrorMsg,
		&creatorID,
		&creatorType,
		&buildTask.CreateTime,
		&updaterID,
		&updaterType,
		&buildTask.UpdateTime,
	)

	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "Build task not found")
		return nil, nil
	}

	if err != nil {
		logger.Errorf("Get build task by ID failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get build task by ID failed: %v", err))
		span.SetStatus(codes.Error, "Get build task by ID failed")
		return nil, err
	}

	buildTask.Creator = interfaces.AccountInfo{
		ID:   creatorID,
		Type: creatorType,
	}

	buildTask.Updater = interfaces.AccountInfo{
		ID:   updaterID,
		Type: updaterType,
	}

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}

// GetByResourceID retrieves build tasks by resource ID.
func (bta *buildTaskAccess) GetByResourceID(ctx context.Context, resourceID string) ([]*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build tasks by resource ID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT 
			f_id, f_resource_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_resource_id = ?
		ORDER BY f_create_time DESC
	`

	rows, err := bta.db.QueryContext(ctx, query, resourceID)
	if err != nil {
		logger.Errorf("Get build tasks by resource ID failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get build tasks by resource ID failed: %v", err))
		span.SetStatus(codes.Error, "Get build tasks by resource ID failed")
		return nil, err
	}
	defer rows.Close()

	buildTasks := []*interfaces.BuildTask{}
	for rows.Next() {
		buildTask := &interfaces.BuildTask{}
		var creatorID, creatorType, updaterID, updaterType string

		err := rows.Scan(
			&buildTask.ID,
			&buildTask.ResourceID,
			&buildTask.Status,
			&buildTask.Mode,
			&buildTask.TotalCount,
			&buildTask.SyncedCount,
			&buildTask.VectorizedCount,
			&buildTask.SyncedMark,
			&buildTask.ErrorMsg,
			&creatorID,
			&creatorType,
			&buildTask.CreateTime,
			&updaterID,
			&updaterType,
			&buildTask.UpdateTime,
		)

		if err != nil {
			logger.Errorf("Scan build task row failed: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Scan build task row failed: %v", err))
			span.SetStatus(codes.Error, "Scan build task row failed")
			return nil, err
		}

		buildTask.Creator = interfaces.AccountInfo{
			ID:   creatorID,
			Type: creatorType,
		}

		buildTask.Updater = interfaces.AccountInfo{
			ID:   updaterID,
			Type: updaterType,
		}

		buildTasks = append(buildTasks, buildTask)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("Rows iteration failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Rows iteration failed: %v", err))
		span.SetStatus(codes.Error, "Rows iteration failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return buildTasks, nil
}

// CheckResourceHasUncompletedTasks checks if resource has uncompleted build tasks by resource ID.
func (bta *buildTaskAccess) CheckResourceHasUncompletedTasks(ctx context.Context, resourceID string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check resource has uncompleted tasks", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT COUNT(*)
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_resource_id = ? AND f_status IN (?, ?)
	`

	var count int
	err := bta.db.QueryRowContext(ctx, query, resourceID, interfaces.BuildTaskStatusPending, interfaces.BuildTaskStatusRunning).Scan(&count)
	if err != nil {
		logger.Errorf("Check resource has uncompleted tasks failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Check resource has uncompleted tasks failed: %v", err))
		span.SetStatus(codes.Error, "Check resource has uncompleted tasks failed")
		return false, err
	}

	span.SetStatus(codes.Ok, "")
	return count > 0, nil
}

// UpdateStatus updates a build task's status and other fields.
func (bta *buildTaskAccess) UpdateStatus(ctx context.Context, id string, updates map[string]interface{}) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update build task status", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// Build update query dynamically
	setClauses := []string{}
	args := []interface{}{}

	// Map field names to column names
	fieldMap := map[string]string{
		"status":          "f_status",
		"totalCount":      "f_total_count",
		"syncedCount":     "f_synced_count",
		"vectorizedCount": "f_vectorized_count",
		"syncedMark":      "f_synced_mark",
		"errorMsg":        "f_error_msg",
	}

	// Add fields to update
	for field, value := range updates {
		if column, ok := fieldMap[field]; ok {
			setClauses = append(setClauses, column+" = ?")
			args = append(args, value)
		}
	}

	// Always update update time
	setClauses = append(setClauses, "f_update_time = ?")
	updateTime := time.Now().UnixMilli()
	args = append(args, updateTime)

	// Add id to args
	args = append(args, id)

	// Build final query
	query := `
		UPDATE ` + BUILD_TASK_TABLE_NAME + `
		SET ` + strings.Join(setClauses, ", ") + `
		WHERE f_id = ?
	`

	_, err := bta.db.ExecContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		logger.Errorf("Update build task status failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Update build task status failed: %v", err))
		span.SetStatus(codes.Error, "Update build task status failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetUncompletedTasks retrieves uncompleted build tasks (pending and running) with limit.
func (bta *buildTaskAccess) GetUncompletedTasks(ctx context.Context, limit int) ([]*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get uncompleted build tasks", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT 
			f_id, f_resource_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_status IN (?, ?)
		ORDER BY f_create_time ASC
		LIMIT ?
	`

	rows, err := bta.db.QueryContext(ctx, query, interfaces.BuildTaskStatusPending, interfaces.BuildTaskStatusRunning, limit)
	if err != nil {
		logger.Errorf("Get uncompleted build tasks failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get uncompleted build tasks failed: %v", err))
		span.SetStatus(codes.Error, "Get uncompleted build tasks failed")
		return nil, err
	}
	defer rows.Close()

	buildTasks := []*interfaces.BuildTask{}
	for rows.Next() {
		buildTask := &interfaces.BuildTask{}
		var creatorID, creatorType, updaterID, updaterType string

		err := rows.Scan(
			&buildTask.ID,
			&buildTask.ResourceID,
			&buildTask.Status,
			&buildTask.Mode,
			&buildTask.TotalCount,
			&buildTask.SyncedCount,
			&buildTask.VectorizedCount,
			&buildTask.SyncedMark,
			&buildTask.ErrorMsg,
			&creatorID,
			&creatorType,
			&buildTask.CreateTime,
			&updaterID,
			&updaterType,
			&buildTask.UpdateTime,
		)

		if err != nil {
			logger.Errorf("Scan build task row failed: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Scan build task row failed: %v", err))
			span.SetStatus(codes.Error, "Scan build task row failed")
			return nil, err
		}

		buildTask.Creator = interfaces.AccountInfo{
			ID:   creatorID,
			Type: creatorType,
		}

		buildTask.Updater = interfaces.AccountInfo{
			ID:   updaterID,
			Type: updaterType,
		}

		buildTasks = append(buildTasks, buildTask)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("Rows iteration failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Rows iteration failed: %v", err))
		span.SetStatus(codes.Error, "Rows iteration failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return buildTasks, nil
}

// GetLastBuildTask retrieves the last completed build task with non-empty synced mark by resource ID.
func (bta *buildTaskAccess) GetLastBuildTask(ctx context.Context, resourceID string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get last build task", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT 
			f_id, f_resource_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_resource_id = ? AND f_synced_mark != ''
		ORDER BY f_create_time DESC
		LIMIT 1
	`

	row := bta.db.QueryRowContext(ctx, query, resourceID)

	buildTask := &interfaces.BuildTask{}
	var creatorID, creatorType, updaterID, updaterType string

	err := row.Scan(
		&buildTask.ID,
		&buildTask.ResourceID,
		&buildTask.Status,
		&buildTask.Mode,
		&buildTask.TotalCount,
		&buildTask.SyncedCount,
		&buildTask.VectorizedCount,
		&buildTask.SyncedMark,
		&buildTask.ErrorMsg,
		&creatorID,
		&creatorType,
		&buildTask.CreateTime,
		&updaterID,
		&updaterType,
		&buildTask.UpdateTime,
	)

	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "No last build task found")
		return nil, nil
	}

	if err != nil {
		logger.Errorf("Get last build task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get last build task failed: %v", err))
		span.SetStatus(codes.Error, "Get last build task failed")
		return nil, err
	}

	buildTask.Creator = interfaces.AccountInfo{
		ID:   creatorID,
		Type: creatorType,
	}

	buildTask.Updater = interfaces.AccountInfo{
		ID:   updaterID,
		Type: updaterType,
	}

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}
