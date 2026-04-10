package rds

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type flowStorageDao struct {
	db *gorm.DB
}

func NewFlowStorageDao() rds.FlowStorageDao {
	return &flowStorageDao{
		db: db.NewDB(),
	}
}

func (d *flowStorageDao) Insert(ctx context.Context, storage *rds.FlowStorage) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_STORAGE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "INSERT INTO t_flow_storage (f_id, f_oss_id, f_object_key, f_name, f_content_type, f_size, f_etag, f_status, f_created_at, f_updated_at, f_deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, storage.ID, storage.OssID, storage.ObjectKey, storage.Name, storage.ContentType, storage.Size, storage.Etag, storage.Status, storage.CreatedAt, storage.UpdatedAt, storage.DeletedAt)
	if result.Error != nil {
		log.Warnf("[FlowStorageDao.Insert] insert err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowStorageDao) GetByID(ctx context.Context, id uint64) (result *rds.FlowStorage, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_STORAGE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT f_id, f_oss_id, f_object_key, f_name, f_content_type, f_size, f_etag, f_status, f_created_at, f_updated_at, f_deleted_at FROM t_flow_storage WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var row struct {
		ID          uint64              `gorm:"column:f_id"`
		OssID       string              `gorm:"column:f_oss_id"`
		ObjectKey   string              `gorm:"column:f_object_key"`
		Name        string              `gorm:"column:f_name"`
		ContentType string              `gorm:"column:f_content_type"`
		Size        uint64              `gorm:"column:f_size"`
		Etag        string              `gorm:"column:f_etag"`
		Status      rds.FlowStorageStatus `gorm:"column:f_status"`
		CreatedAt   int64               `gorm:"column:f_created_at"`
		UpdatedAt   int64               `gorm:"column:f_updated_at"`
		DeletedAt   int64               `gorm:"column:f_deleted_at"`
	}
	err = d.db.Raw(sqlStr, id).Scan(&row).Error
	if err != nil {
		log.Warnf("[FlowStorageDao.GetByID] query err: %s", err.Error())
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	result = &rds.FlowStorage{
		ID:          row.ID,
		OssID:       row.OssID,
		ObjectKey:   row.ObjectKey,
		Name:        row.Name,
		ContentType: row.ContentType,
		Size:        row.Size,
		Etag:        row.Etag,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		DeletedAt:   row.DeletedAt,
	}
	return result, nil
}

func (d *flowStorageDao) GetByOssIDAndObjectKey(ctx context.Context, ossID, objectKey string) (result *rds.FlowStorage, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_STORAGE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT f_id, f_oss_id, f_object_key, f_name, f_content_type, f_size, f_etag, f_status, f_created_at, f_updated_at, f_deleted_at FROM t_flow_storage WHERE f_oss_id = ? AND f_object_key = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var row struct {
		ID          uint64              `gorm:"column:f_id"`
		OssID       string              `gorm:"column:f_oss_id"`
		ObjectKey   string              `gorm:"column:f_object_key"`
		Name        string              `gorm:"column:f_name"`
		ContentType string              `gorm:"column:f_content_type"`
		Size        uint64              `gorm:"column:f_size"`
		Etag        string              `gorm:"column:f_etag"`
		Status      rds.FlowStorageStatus `gorm:"column:f_status"`
		CreatedAt   int64               `gorm:"column:f_created_at"`
		UpdatedAt   int64               `gorm:"column:f_updated_at"`
		DeletedAt   int64               `gorm:"column:f_deleted_at"`
	}
	err = d.db.Raw(sqlStr, ossID, objectKey).Scan(&row).Error
	if err != nil {
		log.Warnf("[FlowStorageDao.GetByOssIDAndObjectKey] query err: %s", err.Error())
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	result = &rds.FlowStorage{
		ID:          row.ID,
		OssID:       row.OssID,
		ObjectKey:   row.ObjectKey,
		Name:        row.Name,
		ContentType: row.ContentType,
		Size:        row.Size,
		Etag:        row.Etag,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		DeletedAt:   row.DeletedAt,
	}
	return result, nil
}

func (d *flowStorageDao) List(ctx context.Context, opts *rds.FlowStorageQueryOptions) (result []*rds.FlowStorage, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_STORAGE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	var args []interface{}
	where := " WHERE 1=1"

	if len(opts.IDs) > 0 {
		where += " AND f_id IN ?"
		args = append(args, opts.IDs)
	}
	if opts.OssID != "" {
		where += " AND f_oss_id = ?"
		args = append(args, opts.OssID)
	}
	if opts.ObjectKey != "" {
		where += " AND f_object_key = ?"
		args = append(args, opts.ObjectKey)
	}
	if opts.Status != nil {
		where += " AND f_status = ?"
		args = append(args, *opts.Status)
	}

	sqlStr := fmt.Sprintf("SELECT f_id, f_oss_id, f_object_key, f_name, f_content_type, f_size, f_etag, f_status, f_created_at, f_updated_at, f_deleted_at FROM t_flow_storage%s ORDER BY f_id", where)
	if opts.Limit > 0 {
		sqlStr += " LIMIT ? OFFSET ?"
		args = append(args, opts.Limit, opts.Offset)
	}
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var rows []struct {
		ID          uint64              `gorm:"column:f_id"`
		OssID       string              `gorm:"column:f_oss_id"`
		ObjectKey   string              `gorm:"column:f_object_key"`
		Name        string              `gorm:"column:f_name"`
		ContentType string              `gorm:"column:f_content_type"`
		Size        uint64              `gorm:"column:f_size"`
		Etag        string              `gorm:"column:f_etag"`
		Status      rds.FlowStorageStatus `gorm:"column:f_status"`
		CreatedAt   int64               `gorm:"column:f_created_at"`
		UpdatedAt   int64               `gorm:"column:f_updated_at"`
		DeletedAt   int64               `gorm:"column:f_deleted_at"`
	}
	err = d.db.Raw(sqlStr, args...).Scan(&rows).Error
	if err != nil {
		log.Warnf("[FlowStorageDao.List] query err: %s", err.Error())
		return nil, err
	}
	result = make([]*rds.FlowStorage, 0, len(rows))
	for _, row := range rows {
		result = append(result, &rds.FlowStorage{
			ID:          row.ID,
			OssID:       row.OssID,
			ObjectKey:   row.ObjectKey,
			Name:        row.Name,
			ContentType: row.ContentType,
			Size:        row.Size,
			Etag:        row.Etag,
			Status:      row.Status,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			DeletedAt:   row.DeletedAt,
		})
	}
	return result, nil
}

func (d *flowStorageDao) Update(ctx context.Context, storage *rds.FlowStorage) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_STORAGE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "UPDATE t_flow_storage SET f_oss_id = ?, f_object_key = ?, f_name = ?, f_content_type = ?, f_size = ?, f_etag = ?, f_status = ?, f_updated_at = ?, f_deleted_at = ? WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, storage.OssID, storage.ObjectKey, storage.Name, storage.ContentType, storage.Size, storage.Etag, storage.Status, storage.UpdatedAt, storage.DeletedAt, storage.ID)
	if result.Error != nil {
		log.Warnf("[FlowStorageDao.Update] update err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowStorageDao) UpdateStatus(ctx context.Context, id uint64, status rds.FlowStorageStatus) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_STORAGE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "UPDATE t_flow_storage SET f_status = ?, f_updated_at = ? WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, status, 0, id) // updated_at will be set by caller
	if result.Error != nil {
		log.Warnf("[FlowStorageDao.UpdateStatus] update err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowStorageDao) Delete(ctx context.Context, id uint64) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_STORAGE_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_flow_storage WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, id)
	if result.Error != nil {
		log.Warnf("[FlowStorageDao.Delete] delete err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}