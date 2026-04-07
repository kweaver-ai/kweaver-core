package rds

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type dagInstanceEventRepository struct {
	db *gorm.DB
}

var (
	dagInstanceEventRepositoryOnce sync.Once
	dagInstanceEventRepositoryIns  rds.DagInstanceEventRepository
)

func NewDagInstanceEventRepository() rds.DagInstanceEventRepository {
	dagInstanceEventRepositoryOnce.Do(func() {
		dagInstanceEventRepositoryIns = &dagInstanceEventRepository{
			db: db.NewDB(),
		}
	})
	return dagInstanceEventRepositoryIns
}

func (d *dagInstanceEventRepository) InsertMany(ctx context.Context, events []*rds.DagInstanceEvent) (err error) {
	if len(events) == 0 {
		return nil
	}
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EVENT_TABLE))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "INSERT INTO t_dag_instance_event (f_id, f_type, f_instance_id, f_operator, f_task_id, f_status, f_name, f_data, f_size, f_inline, f_visibility, f_timestamp) VALUES "
	values := make([]interface{}, 0, len(events)*12)
	for _, e := range events {
		sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
		values = append(values, e.ID, e.Type, e.InstanceID, e.Operator, e.TaskID, e.Status, e.Name, e.Data, e.Size, e.Inline, e.Visibility, e.Timestamp)
	}
	sqlStr = strings.TrimRight(sqlStr, ",")
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, fmt.Sprintf("%v", values)))

	result := d.db.Exec(sqlStr, values...)
	if result.Error != nil {
		log.Warnf("[DagInstanceEventRepository.InsertMany] insert err: %s", result.Error.Error())
		err = result.Error
	}

	return err
}

func (d *dagInstanceEventRepository) ListCount(ctx context.Context, opts *rds.DagInstanceEventListOptions) (int, error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EVENT_TABLE))
	defer func() { trace.TelemetrySpanEnd(span, nil) }()
	log := traceLog.WithContext(newCtx)

	var args []interface{}
	where := " WHERE f_instance_id = ?"
	args = append(args, opts.DagInstanceID)

	if len(opts.Names) > 0 {
		where += " AND f_name IN ?"
		args = append(args, opts.Names)
	}
	if len(opts.Types) > 0 {
		where += " AND f_type IN ?"
		args = append(args, opts.Types)
	}
	if len(opts.Visibilities) > 0 {
		where += " AND f_visibility IN ?"
		args = append(args, opts.Visibilities)
	}
	if opts.Inline != nil {
		where += " AND f_inline = ?"
		args = append(args, *opts.Inline)
	}

	sql := fmt.Sprintf("SELECT COUNT(*) FROM t_dag_instance_event%s", where)
	if opts.LatestOnly {
		sql = fmt.Sprintf(`
			SELECT COUNT(*) FROM (
				SELECT 1
				FROM (
					SELECT f_type, f_name,
						ROW_NUMBER() OVER (PARTITION BY f_type, f_name ORDER BY f_id DESC) AS rn
					FROM t_dag_instance_event%s
				) t
				WHERE rn = 1
			) cnt
		`, where)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", args)))

	var count int
	err := d.db.Raw(sql, args...).Scan(&count).Error
	if err != nil {
		log.Warnf("[DagInstanceEventRepository.ListCount] query failed: %s", err.Error())
		return 0, err
	}

	return count, nil
}

func (d *dagInstanceEventRepository) List(ctx context.Context, opts *rds.DagInstanceEventListOptions) (result []*rds.DagInstanceEvent, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EVENT_TABLE))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	fields := opts.Fields
	if len(fields) == 0 {
		fields = rds.DagInstanceEventFieldAll
	}

	fieldStrs := make([]string, 0, len(fields))
	for _, f := range fields {
		fieldStrs = append(fieldStrs, string(f))
	}
	var args []interface{}
	where := " WHERE f_instance_id = ?"
	args = append(args, opts.DagInstanceID)

	if len(opts.Names) > 0 {
		where += " AND f_name IN ?"
		args = append(args, opts.Names)
	}
	if len(opts.Types) > 0 {
		where += " AND f_type IN ?"
		args = append(args, opts.Types)
	}
	if len(opts.Visibilities) > 0 {
		where += " AND f_visibility IN ?"
		args = append(args, opts.Visibilities)
	}
	if opts.Inline != nil {
		where += " AND f_inline = ?"
		args = append(args, *opts.Inline)
	}

	sql := fmt.Sprintf("SELECT %s FROM t_dag_instance_event%s", strings.Join(fieldStrs, ", "), where)
	if opts.LatestOnly {
		sql = fmt.Sprintf(`
			SELECT %s FROM (
				SELECT %s,
					ROW_NUMBER() OVER (PARTITION BY f_type, f_name ORDER BY f_id DESC) AS rn
				FROM t_dag_instance_event%s
			) t
			WHERE rn = 1
		`, strings.Join(fieldStrs, ", "), strings.Join(fieldStrs, ", "), where)
	}

	sql += " ORDER BY f_id"

	if opts.Limit > 0 {
		sql = fmt.Sprintf("%s LIMIT ? OFFSET ?", sql)
		args = append(args, opts.Limit, opts.Offset)
	}
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", args)))

	var rows []*struct {
		ID         uint64                         `gorm:"column:f_id"`
		Type       rds.DagInstanceEventType       `gorm:"column:f_type"`
		InstanceID string                         `gorm:"column:f_instance_id"`
		Operator   string                         `gorm:"column:f_operator"`
		TaskID     string                         `gorm:"column:f_task_id"`
		Status     string                         `gorm:"column:f_status"`
		Name       string                         `gorm:"column:f_name"`
		Data       string                         `gorm:"column:f_data"`
		Size       int                            `gorm:"column:f_size"`
		Inline     bool                           `gorm:"column:f_inline"`
		Timestamp  int64                          `gorm:"column:f_timestamp"`
		Visibility rds.DagInstanceEventVisibility `gorm:"column:f_visibility"`
	}
	err = d.db.Raw(sql, args...).Scan(&rows).Error
	if err != nil {
		log.Warnf("[DagInstanceEventRepository.List] query failed: %s", err.Error())
		return
	}
	result = make([]*rds.DagInstanceEvent, 0, len(rows))
	for _, row := range rows {
		result = append(result, &rds.DagInstanceEvent{
			ID:         row.ID,
			Type:       row.Type,
			InstanceID: row.InstanceID,
			Operator:   row.Operator,
			TaskID:     row.TaskID,
			Status:     row.Status,
			Name:       row.Name,
			Data:       row.Data,
			Size:       row.Size,
			Inline:     row.Inline,
			Timestamp:  row.Timestamp,
			Visibility: row.Visibility,
		})
	}
	return
}

func (d *dagInstanceEventRepository) DeleteByInstanceIDs(ctx context.Context, instanceIDs []string) (err error) {
	if len(instanceIDs) == 0 {
		return nil
	}
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.DAG_INSTANCE_EVENT_TABLE))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_dag_instance_event WHERE f_instance_id IN ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", instanceIDs)))
	result := d.db.Exec(sqlStr, instanceIDs)
	if result.Error != nil {
		log.Warnf("[DagInstanceEventRepository.DeleteByInstanceIDs] delete err: %s", result.Error.Error())
		err = result.Error
		return err
	}
	return err
}
