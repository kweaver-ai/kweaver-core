package rds

import (
	"context"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type dagInstanceExtDataDao struct {
	db *gorm.DB
}

var (
	dagInstanceExtDataDaoIns  rds.DagInstanceExtDataDao
	dagInstanceExtDataDaoOnce sync.Once
)

func NewDagInstanceExtDataDao() rds.DagInstanceExtDataDao {
	dagInstanceExtDataDaoOnce.Do(func() {
		dagInstanceExtDataDaoIns = &dagInstanceExtDataDao{
			db.NewDB(),
		}
	})
	return dagInstanceExtDataDaoIns
}

func (d *dagInstanceExtDataDao) InsertMany(ctx context.Context, items []*rds.DagInstanceExtData) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	err = d.db.Transaction(func(tx *gorm.DB) error {
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EXT_DATA_TABLE))
		sql := "insert into t_automation_dag_instance_ext_data (f_id, f_created_at, f_updated_at, f_dag_id, f_dag_ins_id, f_field, f_oss_id, f_oss_key, f_size, f_removed) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

		for _, item := range items {
			result := tx.Exec(sql,
				item.ID,
				item.CreatedAt,
				item.UpdatedAt,
				item.DagID,
				item.DagInsID,
				item.Field,
				item.OssID,
				item.OssKey,
				item.Size,
				item.Removed)

			if result.Error != nil {
				log.Warnf("[dagInstanceExtDataDao.Create] create failed: %s", result.Error.Error())
				return result.Error
			}
		}

		return nil
	})

	return err
}

func (d *dagInstanceExtDataDao) List(ctx context.Context, opts *rds.ExtDataQueryOptions) ([]*rds.DagInstanceExtData, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EXT_DATA_TABLE))

	var results []*rds.DagInstanceExtData
	selectFields := "f_id, f_created_at, f_updated_at, f_dag_id, f_dag_ins_id, f_field, f_oss_id, f_oss_key, f_size, f_removed"
	if len(opts.SelectField) > 0 {
		selectFields = ""
		for i, field := range opts.SelectField {
			if i > 0 {
				selectFields += ", "
			}
			selectFields += field
		}
	}
	baseSQL := "SELECT " + selectFields + " FROM t_automation_dag_instance_ext_data WHERE f_removed = ?"
	args := []interface{}{opts.Removed}

	if len(opts.IDs) > 0 {
		baseSQL += " AND f_id IN ?"
		args = append(args, opts.IDs)
	}
	if opts.DagID != "" {
		baseSQL += " AND f_dag_id = ?"
		args = append(args, opts.DagID)
	}
	if opts.DagInsID != "" {
		baseSQL += " AND f_dag_ins_id = ?"
		args = append(args, opts.DagInsID)
	}

	if opts.MinID != "" {
		baseSQL += " AND f_id > ?"
		args = append(args, opts.MinID)
	}

	baseSQL += " ORDER BY f_id ASC"

	if opts.Limit > 0 {
		baseSQL += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, baseSQL))
	err = d.db.Raw(baseSQL, args...).Scan(&results).Error
	if err != nil {
		log.Warnf("[dagInstanceExtDataDao.List] query failed: %s", err.Error())
		return nil, err
	}

	return results, nil
}

func (d *dagInstanceExtDataDao) Remove(ctx context.Context, opts *rds.ExtDataQueryOptions) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EXT_DATA_TABLE))

	err = d.db.Transaction(func(tx *gorm.DB) error {
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EXT_DATA_TABLE))
		sql := "UPDATE t_automation_dag_instance_ext_data SET f_removed = ?, f_updated_at = ? WHERE f_removed = ?"
		trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

		batchSize := 1000
		updatedAt := time.Now().Unix()

		if len(opts.IDs) > 0 {
			for i := 0; i < len(opts.IDs); i += batchSize {
				end := i + batchSize
				if end > len(opts.IDs) {
					end = len(opts.IDs)
				}
				batch := opts.IDs[i:end]

				result := tx.Exec(sql+" AND f_id IN ?",
					true, updatedAt, false, batch)
				if result.Error != nil {
					log.Warnf("[dagInstanceExtDataDao.Remove] batch update failed: %s", result.Error.Error())
					return result.Error
				}
			}
			return nil
		}

		where := ""
		args := []interface{}{true, updatedAt, false}
		if opts.DagID != "" {
			where += " AND f_dag_id = ?"
			args = append(args, opts.DagID)
		}
		if opts.DagInsID != "" {
			where += " AND f_dag_ins_id = ?"
			args = append(args, opts.DagInsID)
		}

		for {
			result := tx.Exec(sql+where+" LIMIT ?",
				append(args, batchSize)...)
			if result.Error != nil {
				log.Warnf("[dagInstanceExtDataDao.Remove] batch update failed: %s", result.Error.Error())
				return result.Error
			}
			if result.RowsAffected < int64(batchSize) {
				break
			}
		}

		return nil
	})

	return err
}

func (d *dagInstanceExtDataDao) Delete(ctx context.Context, opts *rds.ExtDataQueryOptions) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EXT_DATA_TABLE))

	err = d.db.Transaction(func(tx *gorm.DB) error {
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EXT_DATA_TABLE))
		sql := "DELETE FROM t_automation_dag_instance_ext_data WHERE 1=1"
		trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql))

		batchSize := 1000

		if len(opts.IDs) > 0 {
			for i := 0; i < len(opts.IDs); i += batchSize {
				end := i + batchSize
				if end > len(opts.IDs) {
					end = len(opts.IDs)
				}
				batch := opts.IDs[i:end]

				result := tx.Exec(sql+" AND f_id IN ?", batch)
				if result.Error != nil {
					log.Warnf("[dagInstanceExtDataDao.Delete] batch delete failed: %s", result.Error.Error())
					return result.Error
				}
			}
			return nil
		}

		where := ""
		args := []interface{}{}
		if opts.DagID != "" {
			where += " AND f_dag_id = ?"
			args = append(args, opts.DagID)
		}
		if opts.DagInsID != "" {
			where += " AND f_dag_ins_id = ?"
			args = append(args, opts.DagInsID)
		}

		for {
			result := tx.Exec(sql+where+" LIMIT ?",
				append(args, batchSize)...)
			if result.Error != nil {
				log.Warnf("[dagInstanceExtDataDao.Delete] batch delete failed: %s", result.Error.Error())
				return result.Error
			}
			if result.RowsAffected < int64(batchSize) {
				break
			}
		}

		return nil
	})

	return err
}

var _ rds.DagInstanceExtDataDao = (*dagInstanceExtDataDao)(nil)
