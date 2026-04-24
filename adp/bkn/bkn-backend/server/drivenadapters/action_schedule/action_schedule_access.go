// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_schedule

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	"bkn-backend/infra/otel/otellog"
	"bkn-backend/infra/otel/oteltrace"
	"bkn-backend/interfaces"
)

const (
	SCHEDULE_TABLE_NAME = "t_action_schedule"
)

var (
	asAccessOnce sync.Once
	asAccess     interfaces.ActionScheduleAccess
)

type actionScheduleAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewActionScheduleAccess(appSetting *common.AppSetting) interfaces.ActionScheduleAccess {
	asAccessOnce.Do(func() {
		asAccess = &actionScheduleAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return asAccess
}

// CreateSchedule creates a new action schedule
func (a *actionScheduleAccess) CreateSchedule(ctx context.Context, tx *sql.Tx, schedule *interfaces.ActionSchedule) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("Create schedule[%s]", schedule.Name))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	instanceIdentitiesStr, err := sonic.MarshalString(schedule.InstanceIdentities)
	if err != nil {
		logger.Errorf("Failed to marshal _instance_identities: %s", err.Error())
		span.SetStatus(codes.Error, "Marshal _instance_identities failed")
		return err
	}

	dynamicParamsStr, err := sonic.MarshalString(schedule.DynamicParams)
	if err != nil {
		logger.Errorf("Failed to marshal dynamic_params: %s", err.Error())
		span.SetStatus(codes.Error, "Marshal dynamic_params failed")
		return err
	}

	sqlStr, vals, err := sq.Insert(SCHEDULE_TABLE_NAME).
		Columns(
			"f_id",
			"f_name",
			"f_kn_id",
			"f_branch",
			"f_action_type_id",
			"f_cron_expression",
			"f_instance_identities",
			"f_dynamic_params",
			"f_status",
			"f_next_run_time",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			schedule.ID,
			schedule.Name,
			schedule.KNID,
			schedule.Branch,
			schedule.ActionTypeID,
			schedule.CronExpression,
			instanceIdentitiesStr,
			dynamicParamsStr,
			schedule.Status,
			schedule.NextRunTime,
			schedule.Creator.ID,
			schedule.Creator.Type,
			schedule.CreateTime,
			schedule.Updater.ID,
			schedule.Updater.Type,
			schedule.UpdateTime,
		).ToSql()
	if err != nil {
		logger.Errorf("Failed to build insert sql: %s", err.Error())
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Create schedule sql: %s", sqlStr))

	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlStr, vals...)
	} else {
		_, err = a.db.ExecContext(ctx, sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("Insert schedule error: %v", err)
		span.SetStatus(codes.Error, "Insert data error")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateSchedule updates an existing action schedule
func (a *actionScheduleAccess) UpdateSchedule(ctx context.Context, tx *sql.Tx, schedule *interfaces.ActionSchedule) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("Update schedule[%s]", schedule.ID))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	builder := sq.Update(SCHEDULE_TABLE_NAME).Where(sq.Eq{"f_id": schedule.ID})

	if schedule.Name != "" {
		builder = builder.Set("f_name", schedule.Name)
	}
	if schedule.CronExpression != "" {
		builder = builder.Set("f_cron_expression", schedule.CronExpression)
	}
	if schedule.InstanceIdentities != nil {
		instanceIdentitiesStr, err := sonic.MarshalString(schedule.InstanceIdentities)
		if err != nil {
			logger.Errorf("Failed to marshal _instance_identities: %s", err.Error())
			span.SetStatus(codes.Error, "Marshal _instance_identities failed")
			return err
		}
		builder = builder.Set("f_instance_identities", instanceIdentitiesStr)
	}
	if schedule.DynamicParams != nil {
		dynamicParamsStr, err := sonic.MarshalString(schedule.DynamicParams)
		if err != nil {
			logger.Errorf("Failed to marshal dynamic_params: %s", err.Error())
			span.SetStatus(codes.Error, "Marshal dynamic_params failed")
			return err
		}
		builder = builder.Set("f_dynamic_params", dynamicParamsStr)
	}
	if schedule.NextRunTime != 0 {
		builder = builder.Set("f_next_run_time", schedule.NextRunTime)
	}

	builder = builder.Set("f_updater", schedule.Updater.ID)
	builder = builder.Set("f_updater_type", schedule.Updater.Type)
	builder = builder.Set("f_update_time", schedule.UpdateTime)

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build update sql: %s", err.Error())
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Update schedule sql: %s", sqlStr))

	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlStr, vals...)
	} else {
		_, err = a.db.ExecContext(ctx, sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("Update schedule error: %v", err)
		span.SetStatus(codes.Error, "Update data error")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateScheduleStatus updates the status and next run time
func (a *actionScheduleAccess) UpdateScheduleStatus(ctx context.Context, scheduleID, status string, nextRunTime int64) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("Update schedule status[%s] to %s", scheduleID, status))
	defer span.End()

	builder := sq.Update(SCHEDULE_TABLE_NAME).
		Set("f_status", status).
		Set("f_next_run_time", nextRunTime).
		Where(sq.Eq{"f_id": scheduleID})

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = a.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update status error")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteSchedules deletes schedules by IDs
func (a *actionScheduleAccess) DeleteSchedules(ctx context.Context, tx *sql.Tx, scheduleIDs []string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("Delete schedules[%v]", scheduleIDs))
	defer span.End()

	if len(scheduleIDs) == 0 {
		return nil
	}

	sqlStr, vals, err := sq.Delete(SCHEDULE_TABLE_NAME).
		Where(sq.Eq{"f_id": scheduleIDs}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Delete schedules sql: %s", sqlStr))

	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlStr, vals...)
	} else {
		_, err = a.db.ExecContext(ctx, sqlStr, vals...)
	}
	if err != nil {
		span.SetStatus(codes.Error, "Delete data error")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetSchedule gets a single schedule by ID
func (a *actionScheduleAccess) GetSchedule(ctx context.Context, scheduleID string) (*interfaces.ActionSchedule, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("Get schedule[%s]", scheduleID))
	defer span.End()

	if scheduleID == "" {
		return nil, nil
	}

	query := a.buildSelectQuery().Where(sq.Eq{"f_id": scheduleID})

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	row := a.db.QueryRowContext(ctx, sqlStr, vals...)
	schedule, err := a.scanSchedule(row)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		span.SetStatus(codes.Error, "Scan data error")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return schedule, nil
}

// GetSchedules gets schedules by IDs
func (a *actionScheduleAccess) GetSchedules(ctx context.Context, scheduleIDs []string) (map[string]*interfaces.ActionSchedule, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("Get schedules[%v]", scheduleIDs))
	defer span.End()

	if len(scheduleIDs) == 0 {
		return map[string]*interfaces.ActionSchedule{}, nil
	}

	query := a.buildSelectQuery().Where(sq.Eq{"f_id": scheduleIDs})

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	rows, err := a.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query data error")
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*interfaces.ActionSchedule)
	for rows.Next() {
		schedule, err := a.scanScheduleFromRows(rows)
		if err != nil {
			span.SetStatus(codes.Error, "Scan data error")
			return nil, err
		}
		result[schedule.ID] = schedule
	}

	span.SetStatus(codes.Ok, "")
	return result, nil
}

// ListSchedules lists schedules with pagination
func (a *actionScheduleAccess) ListSchedules(ctx context.Context, query interfaces.ActionScheduleQueryParams) ([]*interfaces.ActionSchedule, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "List schedules")
	defer span.End()

	builder := a.buildSelectQuery()

	if query.KNID != "" {
		builder = builder.Where(sq.Eq{"f_kn_id": query.KNID})
	}
	if query.Branch != "" {
		builder = builder.Where(sq.Eq{"f_branch": query.Branch})
	}
	if query.NamePattern != "" {
		builder = builder.Where(sq.Like{"f_name": fmt.Sprintf("%%%s%%", query.NamePattern)})
	}
	if query.ActionTypeID != "" {
		builder = builder.Where(sq.Eq{"f_action_type_id": query.ActionTypeID})
	}
	if query.Status != "" {
		builder = builder.Where(sq.Eq{"f_status": query.Status})
	}

	if query.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", query.Sort, query.Direction))
	}

	if query.Offset > 0 {
		builder = builder.Offset(uint64(query.Offset))
	}
	if query.Limit > 0 {
		builder = builder.Limit(uint64(query.Limit))
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("List schedules sql: %s", sqlStr))

	rows, err := a.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query data error")
		return nil, err
	}
	defer rows.Close()

	var schedules []*interfaces.ActionSchedule
	for rows.Next() {
		schedule, err := a.scanScheduleFromRows(rows)
		if err != nil {
			span.SetStatus(codes.Error, "Scan data error")
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	span.SetStatus(codes.Ok, "")
	return schedules, nil
}

// GetSchedulesTotal gets total count of schedules
func (a *actionScheduleAccess) GetSchedulesTotal(ctx context.Context, queryParams interfaces.ActionScheduleQueryParams) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Get schedules total")
	defer span.End()

	query := sq.Select("COUNT(*)").From(SCHEDULE_TABLE_NAME)

	if queryParams.KNID != "" {
		query = query.Where(sq.Eq{"f_kn_id": queryParams.KNID})
	}
	if queryParams.Branch != "" {
		query = query.Where(sq.Eq{"f_branch": queryParams.Branch})
	}
	if queryParams.NamePattern != "" {
		query = query.Where(sq.Like{"f_name": fmt.Sprintf("%%%s%%", queryParams.NamePattern)})
	}
	if queryParams.ActionTypeID != "" {
		query = query.Where(sq.Eq{"f_action_type_id": queryParams.ActionTypeID})
	}
	if queryParams.Status != "" {
		query = query.Where(sq.Eq{"f_status": queryParams.Status})
	}

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return 0, err
	}

	var total int64
	err = a.db.QueryRowContext(ctx, sqlStr, vals...).Scan(&total)
	if err != nil {
		span.SetStatus(codes.Error, "Query data error")
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// TryAcquireLock attempts to acquire execution lock for a schedule
func (a *actionScheduleAccess) TryAcquireLock(ctx context.Context, scheduleID, podID string, now, lockTimeout int64) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("TryAcquireLock[%s]", scheduleID))
	defer span.End()

	// Atomic lock acquisition with timeout handling
	sqlStr := `UPDATE t_action_schedule 
		SET f_lock_holder = ?, f_lock_time = ?
		WHERE f_id = ? 
		  AND f_status = 'active'
		  AND f_next_run_time <= ?
		  AND (f_lock_holder IS NULL OR f_lock_time < ?)`

	staleTime := now - lockTimeout

	result, err := a.db.ExecContext(ctx, sqlStr, podID, now, scheduleID, now, staleTime)
	if err != nil {
		logger.Errorf("TryAcquireLock error: %v", err)
		span.SetStatus(codes.Error, "Lock acquisition failed")
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.SetStatus(codes.Error, "Get rows affected failed")
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return rowsAffected, nil
}

// ReleaseLock releases the execution lock and updates run times
func (a *actionScheduleAccess) ReleaseLock(ctx context.Context, scheduleID, podID string, lastRunTime, nextRunTime int64) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("ReleaseLock[%s]", scheduleID))
	defer span.End()

	sqlStr := `UPDATE t_action_schedule 
		SET f_lock_holder = NULL, f_lock_time = 0,
		    f_last_run_time = ?, f_next_run_time = ?
		WHERE f_id = ? AND f_lock_holder = ?`

	_, err := a.db.ExecContext(ctx, sqlStr, lastRunTime, nextRunTime, scheduleID, podID)
	if err != nil {
		logger.Errorf("ReleaseLock error: %v", err)
		span.SetStatus(codes.Error, "Lock release failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetDueSchedules returns schedules that are due for execution
func (a *actionScheduleAccess) GetDueSchedules(ctx context.Context, now int64) ([]*interfaces.ActionSchedule, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetDueSchedules")
	defer span.End()

	query := a.buildSelectQuery().
		Where(sq.Eq{"f_status": interfaces.ScheduleStatusActive}).
		Where(sq.LtOrEq{"f_next_run_time": now}).
		Where(sq.Gt{"f_next_run_time": 0})

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	rows, err := a.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query data error")
		return nil, err
	}
	defer rows.Close()

	var schedules []*interfaces.ActionSchedule
	for rows.Next() {
		schedule, err := a.scanScheduleFromRows(rows)
		if err != nil {
			span.SetStatus(codes.Error, "Scan data error")
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	span.SetStatus(codes.Ok, "")
	return schedules, nil
}

// UpdateNextRunTime updates the next run time for a schedule
func (a *actionScheduleAccess) UpdateNextRunTime(ctx context.Context, scheduleID string, nextRunTime int64) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, fmt.Sprintf("UpdateNextRunTime[%s]", scheduleID))
	defer span.End()

	sqlStr, vals, err := sq.Update(SCHEDULE_TABLE_NAME).
		Set("f_next_run_time", nextRunTime).
		Where(sq.Eq{"f_id": scheduleID}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = a.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update data error")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Helper methods

func (a *actionScheduleAccess) buildSelectQuery() sq.SelectBuilder {
	return sq.Select(
		"f_id",
		"f_name",
		"f_kn_id",
		"f_branch",
		"f_action_type_id",
		"f_cron_expression",
		"f_instance_identities",
		"f_dynamic_params",
		"f_status",
		"f_last_run_time",
		"f_next_run_time",
		"f_lock_holder",
		"f_lock_time",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(SCHEDULE_TABLE_NAME)
}

func (a *actionScheduleAccess) scanSchedule(row *sql.Row) (*interfaces.ActionSchedule, error) {
	var schedule interfaces.ActionSchedule
	var instanceIdentitiesStr, dynamicParamsStr string
	var lockHolder sql.NullString

	err := row.Scan(
		&schedule.ID,
		&schedule.Name,
		&schedule.KNID,
		&schedule.Branch,
		&schedule.ActionTypeID,
		&schedule.CronExpression,
		&instanceIdentitiesStr,
		&dynamicParamsStr,
		&schedule.Status,
		&schedule.LastRunTime,
		&schedule.NextRunTime,
		&lockHolder,
		&schedule.LockTime,
		&schedule.Creator.ID,
		&schedule.Creator.Type,
		&schedule.CreateTime,
		&schedule.Updater.ID,
		&schedule.Updater.Type,
		&schedule.UpdateTime,
	)
	if err != nil {
		return nil, err
	}

	if lockHolder.Valid {
		schedule.LockHolder = lockHolder.String
	}

	if instanceIdentitiesStr != "" {
		if err := sonic.UnmarshalString(instanceIdentitiesStr, &schedule.InstanceIdentities); err != nil {
			logger.Warnf("Failed to unmarshal _instance_identities for schedule %s: %v", schedule.ID, err)
			// Initialize to empty slice to avoid nil pointer issues
			schedule.InstanceIdentities = []map[string]any{}
		}
	}
	if dynamicParamsStr != "" {
		if err := sonic.UnmarshalString(dynamicParamsStr, &schedule.DynamicParams); err != nil {
			logger.Warnf("Failed to unmarshal dynamic_params for schedule %s: %v", schedule.ID, err)
			// Initialize to empty map to avoid nil pointer issues
			schedule.DynamicParams = map[string]any{}
		}
	}

	return &schedule, nil
}

func (a *actionScheduleAccess) scanScheduleFromRows(rows *sql.Rows) (*interfaces.ActionSchedule, error) {
	var schedule interfaces.ActionSchedule
	var instanceIdentitiesStr, dynamicParamsStr string
	var lockHolder sql.NullString

	err := rows.Scan(
		&schedule.ID,
		&schedule.Name,
		&schedule.KNID,
		&schedule.Branch,
		&schedule.ActionTypeID,
		&schedule.CronExpression,
		&instanceIdentitiesStr,
		&dynamicParamsStr,
		&schedule.Status,
		&schedule.LastRunTime,
		&schedule.NextRunTime,
		&lockHolder,
		&schedule.LockTime,
		&schedule.Creator.ID,
		&schedule.Creator.Type,
		&schedule.CreateTime,
		&schedule.Updater.ID,
		&schedule.Updater.Type,
		&schedule.UpdateTime,
	)
	if err != nil {
		return nil, err
	}

	if lockHolder.Valid {
		schedule.LockHolder = lockHolder.String
	}

	if instanceIdentitiesStr != "" {
		if err := sonic.UnmarshalString(instanceIdentitiesStr, &schedule.InstanceIdentities); err != nil {
			logger.Warnf("Failed to unmarshal _instance_identities for schedule %s: %v", schedule.ID, err)
			// Initialize to empty slice to avoid nil pointer issues
			schedule.InstanceIdentities = []map[string]any{}
		}
	}
	if dynamicParamsStr != "" {
		if err := sonic.UnmarshalString(dynamicParamsStr, &schedule.DynamicParams); err != nil {
			logger.Warnf("Failed to unmarshal dynamic_params for schedule %s: %v", schedule.ID, err)
			// Initialize to empty map to avoid nil pointer issues
			schedule.DynamicParams = map[string]any{}
		}
	}

	return &schedule, nil
}
