package rds

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type flowTaskResumeDao struct {
	db *gorm.DB
}

func NewFlowTaskResumeDao() rds.FlowTaskResumeDao {
	return &flowTaskResumeDao{
		db: db.NewDB(),
	}
}

func (d *flowTaskResumeDao) Insert(ctx context.Context, resume *rds.FlowTaskResume) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_TASK_RESUME_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "INSERT INTO t_flow_task_resume (f_id, f_task_instance_id, f_dag_instance_id, f_resource_type, f_resource_id, f_created_at, f_updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, resume.ID, resume.TaskInstanceID, resume.DagInstanceID, resume.ResourceType, resume.ResourceID, resume.CreatedAt, resume.UpdatedAt)
	if result.Error != nil {
		log.Warnf("[FlowTaskResumeDao.Insert] insert err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowTaskResumeDao) GetByID(ctx context.Context, id uint64) (result *rds.FlowTaskResume, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_TASK_RESUME_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT f_id, f_task_instance_id, f_dag_instance_id, f_resource_type, f_resource_id, f_created_at, f_updated_at FROM t_flow_task_resume WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var row struct {
		ID             uint64 `gorm:"column:f_id"`
		TaskInstanceID string `gorm:"column:f_task_instance_id"`
		DagInstanceID  string `gorm:"column:f_dag_instance_id"`
		ResourceType   string `gorm:"column:f_resource_type"`
		ResourceID     uint64 `gorm:"column:f_resource_id"`
		CreatedAt      int64  `gorm:"column:f_created_at"`
		UpdatedAt      int64  `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, id).Scan(&row).Error
	if err != nil {
		log.Warnf("[FlowTaskResumeDao.GetByID] query err: %s", err.Error())
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	result = &rds.FlowTaskResume{
		ID:             row.ID,
		TaskInstanceID: row.TaskInstanceID,
		DagInstanceID:  row.DagInstanceID,
		ResourceType:   row.ResourceType,
		ResourceID:     row.ResourceID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	return result, nil
}

func (d *flowTaskResumeDao) GetByTaskInstanceID(ctx context.Context, taskInstanceID string) (result *rds.FlowTaskResume, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_TASK_RESUME_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT f_id, f_task_instance_id, f_dag_instance_id, f_resource_type, f_resource_id, f_created_at, f_updated_at FROM t_flow_task_resume WHERE f_task_instance_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var row struct {
		ID             uint64 `gorm:"column:f_id"`
		TaskInstanceID string `gorm:"column:f_task_instance_id"`
		DagInstanceID  string `gorm:"column:f_dag_instance_id"`
		ResourceType   string `gorm:"column:f_resource_type"`
		ResourceID     uint64 `gorm:"column:f_resource_id"`
		CreatedAt      int64  `gorm:"column:f_created_at"`
		UpdatedAt      int64  `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, taskInstanceID).Scan(&row).Error
	if err != nil {
		log.Warnf("[FlowTaskResumeDao.GetByTaskInstanceID] query err: %s", err.Error())
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	result = &rds.FlowTaskResume{
		ID:             row.ID,
		TaskInstanceID: row.TaskInstanceID,
		DagInstanceID:  row.DagInstanceID,
		ResourceType:   row.ResourceType,
		ResourceID:     row.ResourceID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	return result, nil
}

func (d *flowTaskResumeDao) List(ctx context.Context, opts *rds.FlowTaskResumeQueryOptions) (result []*rds.FlowTaskResume, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_TASK_RESUME_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	var args []interface{}
	where := " WHERE 1=1"

	if opts.ID != nil {
		where += " AND f_id = ?"
		args = append(args, *opts.ID)
	}
	if opts.TaskInstanceID != "" {
		where += " AND f_task_instance_id = ?"
		args = append(args, opts.TaskInstanceID)
	}
	if opts.DagInstanceID != "" {
		where += " AND f_dag_instance_id = ?"
		args = append(args, opts.DagInstanceID)
	}
	if opts.ResourceType != "" {
		where += " AND f_resource_type = ?"
		args = append(args, opts.ResourceType)
	}
	if opts.ResourceID != nil {
		where += " AND f_resource_id = ?"
		args = append(args, *opts.ResourceID)
	}

	sqlStr := fmt.Sprintf("SELECT f_id, f_task_instance_id, f_dag_instance_id, f_resource_type, f_resource_id, f_created_at, f_updated_at FROM t_flow_task_resume%s ORDER BY f_id", where)
	if opts.Limit > 0 {
		sqlStr += " LIMIT ? OFFSET ?"
		args = append(args, opts.Limit, opts.Offset)
	}
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var rows []struct {
		ID             uint64 `gorm:"column:f_id"`
		TaskInstanceID string `gorm:"column:f_task_instance_id"`
		DagInstanceID  string `gorm:"column:f_dag_instance_id"`
		ResourceType   string `gorm:"column:f_resource_type"`
		ResourceID     uint64 `gorm:"column:f_resource_id"`
		CreatedAt      int64  `gorm:"column:f_created_at"`
		UpdatedAt      int64  `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, args...).Scan(&rows).Error
	if err != nil {
		log.Warnf("[FlowTaskResumeDao.List] query err: %s", err.Error())
		return nil, err
	}
	result = make([]*rds.FlowTaskResume, 0, len(rows))
	for _, row := range rows {
		result = append(result, &rds.FlowTaskResume{
			ID:             row.ID,
			TaskInstanceID: row.TaskInstanceID,
			DagInstanceID:  row.DagInstanceID,
			ResourceType:   row.ResourceType,
			ResourceID:     row.ResourceID,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return result, nil
}

func (d *flowTaskResumeDao) Delete(ctx context.Context, id uint64) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_TASK_RESUME_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_flow_task_resume WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, id)
	if result.Error != nil {
		log.Warnf("[FlowTaskResumeDao.Delete] delete err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowTaskResumeDao) DeleteByTaskInstanceID(ctx context.Context, taskInstanceID string) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_TASK_RESUME_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_flow_task_resume WHERE f_task_instance_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, taskInstanceID)
	if result.Error != nil {
		log.Warnf("[FlowTaskResumeDao.DeleteByTaskInstanceID] delete err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowTaskResumeDao) DeleteByResource(ctx context.Context, resourceType string, resourceID uint64) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_TASK_RESUME_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_flow_task_resume WHERE f_resource_type = ? AND f_resource_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, resourceType, resourceID)
	if result.Error != nil {
		log.Warnf("[FlowTaskResumeDao.DeleteByResource] delete err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}