package rds

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type flowFileDao struct {
	db *gorm.DB
}

func NewFlowFileDao() rds.FlowFileDao {
	return &flowFileDao{
		db: db.NewDB(),
	}
}

func (d *flowFileDao) Insert(ctx context.Context, file *rds.FlowFile) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "INSERT INTO t_flow_file (f_id, f_dag_id, f_dag_instance_id, f_storage_id, f_status, f_name, f_expires_at, f_created_at, f_updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, file.ID, file.DagID, file.DagInstanceID, file.StorageID, file.Status, file.Name, file.ExpiresAt, file.CreatedAt, file.UpdatedAt)
	if result.Error != nil {
		log.Warnf("[FlowFileDao.Insert] insert err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowFileDao) GetByID(ctx context.Context, id uint64) (result *rds.FlowFile, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT f_id, f_dag_id, f_dag_instance_id, f_storage_id, f_status, f_name, f_expires_at, f_created_at, f_updated_at FROM t_flow_file WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var row struct {
		ID            uint64            `gorm:"column:f_id"`
		DagID         string            `gorm:"column:f_dag_id"`
		DagInstanceID string            `gorm:"column:f_dag_instance_id"`
		StorageID     uint64            `gorm:"column:f_storage_id"`
		Status        rds.FlowFileStatus `gorm:"column:f_status"`
		Name          string            `gorm:"column:f_name"`
		ExpiresAt     int64             `gorm:"column:f_expires_at"`
		CreatedAt     int64             `gorm:"column:f_created_at"`
		UpdatedAt     int64             `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, id).Scan(&row).Error
	if err != nil {
		log.Warnf("[FlowFileDao.GetByID] query err: %s", err.Error())
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	result = &rds.FlowFile{
		ID:            row.ID,
		DagID:         row.DagID,
		DagInstanceID: row.DagInstanceID,
		StorageID:     row.StorageID,
		Status:        row.Status,
		Name:          row.Name,
		ExpiresAt:     row.ExpiresAt,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	return result, nil
}

func (d *flowFileDao) List(ctx context.Context, opts *rds.FlowFileQueryOptions) (result []*rds.FlowFile, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	var args []interface{}
	where := " WHERE 1=1"

	if opts.ID != nil {
		where += " AND f_id = ?"
		args = append(args, *opts.ID)
	}
	if len(opts.IDs) > 0 {
		where += " AND f_id IN ?"
		args = append(args, opts.IDs)
	}
	if opts.DagID != "" {
		where += " AND f_dag_id = ?"
		args = append(args, opts.DagID)
	}
	if opts.DagInstanceID != "" {
		where += " AND f_dag_instance_id = ?"
		args = append(args, opts.DagInstanceID)
	}
	if opts.StorageID != nil {
		where += " AND f_storage_id = ?"
		args = append(args, *opts.StorageID)
	}
	if opts.Status != nil {
		where += " AND f_status = ?"
		args = append(args, *opts.Status)
	}
	if len(opts.Statuses) > 0 {
		where += " AND f_status IN ?"
		args = append(args, opts.Statuses)
	}
	if opts.ExpiresBefore > 0 {
		where += " AND f_expires_at > 0 AND f_expires_at < ?"
		args = append(args, opts.ExpiresBefore)
	}

	sqlStr := fmt.Sprintf("SELECT f_id, f_dag_id, f_dag_instance_id, f_storage_id, f_status, f_name, f_expires_at, f_created_at, f_updated_at FROM t_flow_file%s ORDER BY f_id", where)
	if opts.Limit > 0 {
		sqlStr += " LIMIT ? OFFSET ?"
		args = append(args, opts.Limit, opts.Offset)
	}
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var rows []struct {
		ID            uint64            `gorm:"column:f_id"`
		DagID         string            `gorm:"column:f_dag_id"`
		DagInstanceID string            `gorm:"column:f_dag_instance_id"`
		StorageID     uint64            `gorm:"column:f_storage_id"`
		Status        rds.FlowFileStatus `gorm:"column:f_status"`
		Name          string            `gorm:"column:f_name"`
		ExpiresAt     int64             `gorm:"column:f_expires_at"`
		CreatedAt     int64             `gorm:"column:f_created_at"`
		UpdatedAt     int64             `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, args...).Scan(&rows).Error
	if err != nil {
		log.Warnf("[FlowFileDao.List] query err: %s", err.Error())
		return nil, err
	}
	result = make([]*rds.FlowFile, 0, len(rows))
	for _, row := range rows {
		result = append(result, &rds.FlowFile{
			ID:            row.ID,
			DagID:         row.DagID,
			DagInstanceID: row.DagInstanceID,
			StorageID:     row.StorageID,
			Status:        row.Status,
			Name:          row.Name,
			ExpiresAt:     row.ExpiresAt,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return result, nil
}

func (d *flowFileDao) Update(ctx context.Context, id uint64, params *rds.FlowFileUpdateParams) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	var setClauses []string
	var args []interface{}

	if params.StorageID != nil {
		setClauses = append(setClauses, "f_storage_id = ?")
		args = append(args, *params.StorageID)
	}
	if params.Status != nil {
		setClauses = append(setClauses, "f_status = ?")
		args = append(args, *params.Status)
	}
	if params.Name != nil {
		setClauses = append(setClauses, "f_name = ?")
		args = append(args, *params.Name)
	}
	if params.ExpiresAt != nil {
		setClauses = append(setClauses, "f_expires_at = ?")
		args = append(args, *params.ExpiresAt)
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "f_updated_at = ?")
	args = append(args, time.Now().Unix())
	args = append(args, id)

	setClause := strings.Join(setClauses, ", ")
	sqlStr := fmt.Sprintf("UPDATE t_flow_file SET %s WHERE f_id = ?", setClause)
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, args...)
	if result.Error != nil {
		log.Warnf("[FlowFileDao.Update] update err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowFileDao) UpdateStatus(ctx context.Context, id uint64, status rds.FlowFileStatus) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "UPDATE t_flow_file SET f_status = ?, f_updated_at = ? WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, status, time.Now().Unix(), id)
	if result.Error != nil {
		log.Warnf("[FlowFileDao.UpdateStatus] update err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowFileDao) Delete(ctx context.Context, id uint64) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_flow_file WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, id)
	if result.Error != nil {
		log.Warnf("[FlowFileDao.Delete] delete err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowFileDao) CountByStorageID(ctx context.Context, storageID uint64) (count int64, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT COUNT(*) FROM t_flow_file WHERE f_storage_id = ? AND f_status IN (1, 2)" // pending or ready
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	err = d.db.Raw(sqlStr, storageID).Scan(&count).Error
	if err != nil {
		log.Warnf("[FlowFileDao.CountByStorageID] query err: %s", err.Error())
		return 0, err
	}
	return count, nil
}