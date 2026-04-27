package rds

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type taskCache struct {
	db *gorm.DB
}

var (
	taskCacheIns  rds.TaskCache
	taskCacheOnce sync.Once
)

func NewTaskCache() rds.TaskCache {
	taskCacheOnce.Do(func() {
		taskCacheIns = &taskCache{
			db: db.NewDB(),
		}
	})

	return taskCacheIns
}

func (d *taskCache) Insert(ctx context.Context, task *rds.TaskCacheItem) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)
	tableName := fmt.Sprintf(rds.TaskCacheTableFormat, task.Hash[0:1])
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, tableName))
	sql := fmt.Sprintf(`
		INSERT INTO %s (
			f_id, f_hash, f_type, f_status, f_oss_id, f_oss_key, f_ext, f_size, 
			f_err_msg, f_create_time, f_modify_time, f_expire_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, tableName)

	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))
	result := d.db.Exec(sql,
		task.ID,
		task.Hash,
		task.Type,
		task.Status,
		task.OssID,
		task.OssKey,
		task.Ext,
		task.Size,
		task.ErrMsg,
		task.CreateTime,
		task.ModifyTime,
		task.ExpireTime,
	)

	if result.Error != nil {
		log.Warnf("[TaskCache.Insert] insert failed: %s", result.Error.Error())
		return result.Error
	}

	return nil
}

func (d *taskCache) Update(ctx context.Context, task *rds.TaskCacheItem) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)
	if task.Hash == "" {
		log.Warnf("[TaskCache.Update] hash is empty")
		return fmt.Errorf("hash is empty")
	}

	tableName := fmt.Sprintf(rds.TaskCacheTableFormat, task.Hash[0:1])
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, tableName))

	var (
		fields []string
		args   []any
	)

	if task.Status != 0 {
		fields = append(fields, "f_status = ?")
		args = append(args, task.Status)
	}

	if task.Ext != "" {
		fields = append(fields, "f_ext = ?")
		args = append(args, task.Ext)
	}

	if task.Size != 0 {
		fields = append(fields, "f_size = ?")
		args = append(args, task.Size)
	}

	if task.ErrMsg != "" {
		fields = append(fields, "f_err_msg = ?")
		args = append(args, task.ErrMsg)
	}

	if len(args) == 0 {
		return nil
	}

	args = append(args, task.Hash)

	sql := fmt.Sprintf(`
	    UPDATE %s SET %s WHERE f_hash = ?
	`, tableName, strings.Join(fields, ","))

	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

	if err = d.db.Exec(sql, args...).Error; err != nil {
		return err
	}

	return nil
}

func (d *taskCache) GetByHash(ctx context.Context, hash string) (*rds.TaskCacheItem, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	if len(hash) == 0 {
		return nil, nil
	}

	tableName := fmt.Sprintf(rds.TaskCacheTableFormat, hash[0:1])
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, tableName))
	sql := fmt.Sprintf(`
		SELECT 
			f_id, f_hash, f_type, f_status, f_oss_id, f_oss_key, f_ext, f_size, 
			f_err_msg, f_create_time, f_modify_time, f_expire_time
		FROM %s
		WHERE f_hash = ? 
		LIMIT 1`, tableName)

	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))
	var task rds.TaskCacheItem
	result := d.db.Raw(sql, hash).Scan(&task)
	if result.Error != nil {
		log.Errorf("[TaskCacheItem.GetTaskByHash] query failed: %v", result.Error)
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &task, nil
}

func (d *taskCache) DeleteByHash(ctx context.Context, hash string) error {
	var err error
	newCtx, span := trace.StartConsumerSpan(ctx)
	defer func() {
		trace.TelemetrySpanEnd(span, err)
	}()
	log := traceLog.WithContext(newCtx)

	tableName := fmt.Sprintf(rds.TaskCacheTableFormat, hash[0:1])
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, tableName))

	sql := fmt.Sprintf(`DELETE FROM %s WHERE f_hash = ?`, tableName)

	result := d.db.Exec(sql, hash)

	if result.Error != nil {
		log.Errorf("[TaskCacheItem.DeleteTaskByHash] failed: %v", result.Error.Error())
		return result.Error
	}

	return nil
}

func (d *taskCache) ListTaskCache(ctx context.Context, opts rds.ListTaskCacheOptions) ([]*rds.TaskCacheItem, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	tableName := fmt.Sprintf(rds.TaskCacheTableFormat, opts.TableSuffix)

	var (
		whereConds []string
		args       []interface{}
	)

	if opts.Expired != nil {
		if *opts.Expired {
			whereConds = append(whereConds, "f_expire_time > 0 AND f_expire_time < ?")
			args = append(args, time.Now().Unix())
		} else {
			whereConds = append(whereConds, "(f_expire_time = 0 OR f_expire_time >= ?)")
			args = append(args, time.Now().Unix())
		}
	}

	orderClause := ""
	if opts.MinID != 0 {
		whereConds = append(whereConds, "f_id > ?")
		args = append(args, opts.MinID)
		orderClause = "ORDER BY f_id ASC"
	}

	whereClause := ""
	if len(whereConds) > 0 {
		whereClause = "WHERE " + strings.Join(whereConds, " AND ")
	}

	limitClause := ""
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}

	sql := fmt.Sprintf(`
		SELECT 
			f_id, f_hash, f_type, f_status, f_oss_id, f_oss_key, f_ext, f_size, 
			f_err_msg, f_create_time, f_modify_time, f_expire_time
		FROM %s
		%s
		%s
		%s
	`, tableName, whereClause, orderClause, limitClause)

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, tableName))
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

	var tasks []*rds.TaskCacheItem
	result := d.db.Raw(sql, args...).Scan(&tasks)
	if result.Error != nil {
		log.Errorf("[TaskCache.ListTaskCache] query failed: %v", result.Error)
		return nil, result.Error
	}

	return tasks, nil
}

func (d taskCache) BatchDeleteByHash(ctx context.Context, hashes []any) error {
	if len(hashes) == 0 {
		return nil
	}

	tblHashMap := make(map[string][]any)
	for _, h := range hashes {
		hashStr, ok := h.(string)
		if !ok || len(hashStr) == 0 {
			continue
		}
		tableSuffix := hashStr[0:1]
		table := fmt.Sprintf(rds.TaskCacheTableFormat, tableSuffix)
		tblHashMap[table] = append(tblHashMap[table], hashStr)
	}

	tx := d.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for table, hashList := range tblHashMap {
		if len(hashList) == 0 {
			continue
		}
		placeholders := strings.Repeat("?,", len(hashList))
		placeholders = placeholders[:len(placeholders)-1]
		sql := fmt.Sprintf("DELETE FROM %s WHERE f_hash IN (%s)", table, placeholders)
		result := tx.Exec(sql, hashList...)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
