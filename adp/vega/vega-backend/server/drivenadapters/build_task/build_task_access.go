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
			f_id, f_resource_id, f_catalog_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time, f_embedding_fields, f_build_key_fields, f_embedding_model, f_model_dimensions
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := bta.db.ExecContext(
		ctx,
		query,
		buildTask.ID,
		buildTask.ResourceID,
		buildTask.CatalogID,
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
		buildTask.EmbeddingFields,
		buildTask.BuildKeyFields,
		buildTask.EmbeddingModel,
		buildTask.ModelDimensions,
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
			f_id, f_resource_id, f_catalog_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time, f_embedding_fields, f_build_key_fields, f_embedding_model, f_model_dimensions
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_id = ?
	`

	row := bta.db.QueryRowContext(ctx, query, id)

	buildTask := &interfaces.BuildTask{}
	var creatorID, creatorType, updaterID, updaterType string

	err := row.Scan(
		&buildTask.ID,
		&buildTask.ResourceID,
		&buildTask.CatalogID,
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
		&buildTask.EmbeddingFields,
		&buildTask.BuildKeyFields,
		&buildTask.EmbeddingModel,
		&buildTask.ModelDimensions,
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

// GetByResourceID retrieves a build task by resource ID.
func (bta *buildTaskAccess) GetByResourceID(ctx context.Context, resourceID string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build task by resource ID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT 
			f_id, f_resource_id, f_catalog_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time, f_embedding_fields, f_build_key_fields, f_embedding_model, f_model_dimensions
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_resource_id = ?
		LIMIT 1
	`

	row := bta.db.QueryRowContext(ctx, query, resourceID)

	buildTask := &interfaces.BuildTask{}
	var creatorID, creatorType, updaterID, updaterType string

	err := row.Scan(
		&buildTask.ID,
		&buildTask.ResourceID,
		&buildTask.CatalogID,
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
		&buildTask.EmbeddingFields,
		&buildTask.BuildKeyFields,
		&buildTask.EmbeddingModel,
		&buildTask.ModelDimensions,
	)

	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "Build task not found")
		return nil, nil
	}

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

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}

// GetByCatalogID retrieves build tasks by catalog ID.
func (bta *buildTaskAccess) GetByCatalogID(ctx context.Context, catalogID string) ([]*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build tasks by catalog ID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT 
			f_id, f_resource_id, f_catalog_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time, f_embedding_fields, f_build_key_fields, f_embedding_model, f_model_dimensions
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_catalog_id = ?
	`

	rows, err := bta.db.QueryContext(ctx, query, catalogID)
	if err != nil {
		logger.Errorf("Get build tasks by catalog ID failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get build tasks by catalog ID failed: %v", err))
		span.SetStatus(codes.Error, "Get build tasks by catalog ID failed")
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	buildTasks := []*interfaces.BuildTask{}
	for rows.Next() {
		buildTask := &interfaces.BuildTask{}
		var creatorID, creatorType, updaterID, updaterType string

		err := rows.Scan(
			&buildTask.ID,
			&buildTask.ResourceID,
			&buildTask.CatalogID,
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
			&buildTask.EmbeddingFields,
			&buildTask.BuildKeyFields,
			&buildTask.EmbeddingModel,
			&buildTask.ModelDimensions,
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

// GetStatus retrieves the status of a build task by ID.
func (bta *buildTaskAccess) GetStatus(ctx context.Context, id string) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build task status", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		SELECT f_status
		FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_id = ?
	`

	var status string
	err := bta.db.QueryRowContext(ctx, query, id).Scan(&status)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "Build task not found")
		return "", fmt.Errorf("build task not found")
	}

	if err != nil {
		logger.Errorf("Get build task status failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get build task status failed: %v", err))
		span.SetStatus(codes.Error, "Get build task status failed")
		return "", err
	}

	span.SetStatus(codes.Ok, "")
	return status, nil
}

// List retrieves build tasks with optional filters and pagination.
func (bta *buildTaskAccess) List(ctx context.Context, params interfaces.BuildTasksQueryParams) ([]*interfaces.BuildTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build tasks with filters", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	whereClauses := []string{}
	args := []interface{}{}
	if params.ResourceID != "" {
		whereClauses = append(whereClauses, "f_resource_id = ?")
		args = append(args, params.ResourceID)
	}
	if params.CatalogID != "" {
		whereClauses = append(whereClauses, "f_catalog_id = ?")
		args = append(args, params.CatalogID)
	}
	if params.Status != "" {
		whereClauses = append(whereClauses, "f_status = ?")
		args = append(args, params.Status)
	}
	if params.Mode != "" {
		whereClauses = append(whereClauses, "f_mode = ?")
		args = append(args, params.Mode)
	}
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM ` + BUILD_TASK_TABLE_NAME + ` ` + whereClause
	if err := bta.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		logger.Errorf("Count build tasks failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Count build tasks failed: %v", err))
		span.SetStatus(codes.Error, "Count build tasks failed")
		return nil, 0, err
	}

	orderBy := "f_update_time"
	switch params.Sort {
	case "create_time":
		orderBy = "f_create_time"
	case "update_time":
		orderBy = "f_update_time"
	case "status":
		orderBy = "f_status"
	case "mode":
		orderBy = "f_mode"
	}
	direction := "DESC"
	if params.Direction == interfaces.ASC_DIRECTION {
		direction = "ASC"
	}

	query := `
		SELECT
			f_id, f_resource_id, f_catalog_id, f_status, f_mode, f_total_count, f_synced_count, f_vectorized_count, f_synced_mark, f_error_msg,
			f_creator_id, f_creator_type, f_create_time, f_updater_id, f_updater_type, f_update_time, f_embedding_fields, f_build_key_fields, f_embedding_model, f_model_dimensions
		FROM ` + BUILD_TASK_TABLE_NAME + ` ` + whereClause + `
		ORDER BY ` + orderBy + ` ` + direction + `
		LIMIT ? OFFSET ?
	`
	queryArgs := append(args, params.Limit, params.Offset)
	rows, err := bta.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		logger.Errorf("Get build tasks with filters failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get build tasks with filters failed: %v", err))
		span.SetStatus(codes.Error, "Get build tasks with filters failed")
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	buildTasks := []*interfaces.BuildTask{}
	for rows.Next() {
		buildTask := &interfaces.BuildTask{}
		var creatorID, creatorType, updaterID, updaterType string
		if err := rows.Scan(
			&buildTask.ID, &buildTask.ResourceID, &buildTask.CatalogID, &buildTask.Status, &buildTask.Mode,
			&buildTask.TotalCount, &buildTask.SyncedCount, &buildTask.VectorizedCount, &buildTask.SyncedMark, &buildTask.ErrorMsg,
			&creatorID, &creatorType, &buildTask.CreateTime, &updaterID, &updaterType, &buildTask.UpdateTime,
			&buildTask.EmbeddingFields, &buildTask.BuildKeyFields, &buildTask.EmbeddingModel, &buildTask.ModelDimensions,
		); err != nil {
			logger.Errorf("Scan build task row failed: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Scan build task row failed: %v", err))
			span.SetStatus(codes.Error, "Scan build task row failed")
			return nil, 0, err
		}
		buildTask.Creator = interfaces.AccountInfo{ID: creatorID, Type: creatorType}
		buildTask.Updater = interfaces.AccountInfo{ID: updaterID, Type: updaterType}
		buildTasks = append(buildTasks, buildTask)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Rows iteration failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Rows iteration failed: %v", err))
		span.SetStatus(codes.Error, "Rows iteration failed")
		return nil, 0, err
	}

	span.SetStatus(codes.Ok, "")
	return buildTasks, totalCount, nil
}

// Delete deletes a build task by ID.
func (bta *buildTaskAccess) Delete(ctx context.Context, id string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete build task", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	query := `
		DELETE FROM ` + BUILD_TASK_TABLE_NAME + `
		WHERE f_id = ?
	`

	result, err := bta.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Errorf("Delete build task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Delete build task failed: %v", err))
		span.SetStatus(codes.Error, "Delete build task failed")
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("Get rows affected failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Get rows affected failed: %v", err))
		span.SetStatus(codes.Error, "Get rows affected failed")
		return err
	}

	if affected == 0 {
		span.SetStatus(codes.Ok, "Build task not found")
		return fmt.Errorf("build task not found")
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
