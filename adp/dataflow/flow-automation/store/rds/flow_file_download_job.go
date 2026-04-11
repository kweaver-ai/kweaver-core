package rds

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type flowFileDownloadJobDao struct {
	db *gorm.DB
}

func NewFlowFileDownloadJobDao() rds.FlowFileDownloadJobDao {
	return &flowFileDownloadJobDao{
		db: db.NewDB(),
	}
}

func (d *flowFileDownloadJobDao) Insert(ctx context.Context, job *rds.FlowFileDownloadJob) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "INSERT INTO t_flow_file_download_job (f_id, f_file_id, f_status, f_retry_count, f_max_retry, f_next_retry_at, f_error_code, f_error_message, f_download_url, f_started_at, f_finished_at, f_created_at, f_updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, job.ID, job.FileID, job.Status, job.RetryCount, job.MaxRetry, job.NextRetryAt, job.ErrorCode, job.ErrorMessage, job.DownloadURL, job.StartedAt, job.FinishedAt, job.CreatedAt, job.UpdatedAt)
	if result.Error != nil {
		log.Warnf("[FlowFileDownloadJobDao.Insert] insert err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowFileDownloadJobDao) GetByID(ctx context.Context, id uint64) (result *rds.FlowFileDownloadJob, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT f_id, f_file_id, f_status, f_retry_count, f_max_retry, f_next_retry_at, f_error_code, f_error_message, f_download_url, f_started_at, f_finished_at, f_created_at, f_updated_at FROM t_flow_file_download_job WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var row struct {
		ID           uint64                         `gorm:"column:f_id"`
		FileID       uint64                         `gorm:"column:f_file_id"`
		Status       rds.FlowFileDownloadJobStatus  `gorm:"column:f_status"`
		RetryCount   int                            `gorm:"column:f_retry_count"`
		MaxRetry     int                            `gorm:"column:f_max_retry"`
		NextRetryAt  int64                          `gorm:"column:f_next_retry_at"`
		ErrorCode    string                         `gorm:"column:f_error_code"`
		ErrorMessage string                         `gorm:"column:f_error_message"`
		DownloadURL  string                         `gorm:"column:f_download_url"`
		StartedAt    int64                          `gorm:"column:f_started_at"`
		FinishedAt   int64                          `gorm:"column:f_finished_at"`
		CreatedAt    int64                          `gorm:"column:f_created_at"`
		UpdatedAt    int64                          `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, id).Scan(&row).Error
	if err != nil {
		log.Warnf("[FlowFileDownloadJobDao.GetByID] query err: %s", err.Error())
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	result = &rds.FlowFileDownloadJob{
		ID:           row.ID,
		FileID:       row.FileID,
		Status:       row.Status,
		RetryCount:   row.RetryCount,
		MaxRetry:     row.MaxRetry,
		NextRetryAt:  row.NextRetryAt,
		ErrorCode:    row.ErrorCode,
		ErrorMessage: row.ErrorMessage,
		DownloadURL:  row.DownloadURL,
		StartedAt:    row.StartedAt,
		FinishedAt:   row.FinishedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	return result, nil
}

func (d *flowFileDownloadJobDao) GetByFileID(ctx context.Context, fileID uint64) (result *rds.FlowFileDownloadJob, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "SELECT f_id, f_file_id, f_status, f_retry_count, f_max_retry, f_next_retry_at, f_error_code, f_error_message, f_download_url, f_started_at, f_finished_at, f_created_at, f_updated_at FROM t_flow_file_download_job WHERE f_file_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var row struct {
		ID           uint64                         `gorm:"column:f_id"`
		FileID       uint64                         `gorm:"column:f_file_id"`
		Status       rds.FlowFileDownloadJobStatus  `gorm:"column:f_status"`
		RetryCount   int                            `gorm:"column:f_retry_count"`
		MaxRetry     int                            `gorm:"column:f_max_retry"`
		NextRetryAt  int64                          `gorm:"column:f_next_retry_at"`
		ErrorCode    string                         `gorm:"column:f_error_code"`
		ErrorMessage string                         `gorm:"column:f_error_message"`
		DownloadURL  string                         `gorm:"column:f_download_url"`
		StartedAt    int64                          `gorm:"column:f_started_at"`
		FinishedAt   int64                          `gorm:"column:f_finished_at"`
		CreatedAt    int64                          `gorm:"column:f_created_at"`
		UpdatedAt    int64                          `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, fileID).Scan(&row).Error
	if err != nil {
		log.Warnf("[FlowFileDownloadJobDao.GetByFileID] query err: %s", err.Error())
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	result = &rds.FlowFileDownloadJob{
		ID:           row.ID,
		FileID:       row.FileID,
		Status:       row.Status,
		RetryCount:   row.RetryCount,
		MaxRetry:     row.MaxRetry,
		NextRetryAt:  row.NextRetryAt,
		ErrorCode:    row.ErrorCode,
		ErrorMessage: row.ErrorMessage,
		DownloadURL:  row.DownloadURL,
		StartedAt:    row.StartedAt,
		FinishedAt:   row.FinishedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	return result, nil
}

func (d *flowFileDownloadJobDao) List(ctx context.Context, opts *rds.FlowFileDownloadJobQueryOptions) (result []*rds.FlowFileDownloadJob, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	var args []interface{}
	where := " WHERE 1=1"

	if opts.ID != nil {
		where += " AND f_id = ?"
		args = append(args, *opts.ID)
	}
	if opts.FileID != nil {
		where += " AND f_file_id = ?"
		args = append(args, *opts.FileID)
	}
	if opts.Status != nil {
		where += " AND f_status = ?"
		args = append(args, *opts.Status)
	}
	if len(opts.Statuses) > 0 {
		where += " AND f_status IN ?"
		args = append(args, opts.Statuses)
	}
	if opts.RetryBefore > 0 {
		where += " AND (f_next_retry_at = 0 OR f_next_retry_at < ?)"
		args = append(args, opts.RetryBefore)
	}

	sqlStr := fmt.Sprintf("SELECT f_id, f_file_id, f_status, f_retry_count, f_max_retry, f_next_retry_at, f_error_code, f_error_message, f_download_url, f_started_at, f_finished_at, f_created_at, f_updated_at FROM t_flow_file_download_job%s ORDER BY f_id", where)
	if opts.Limit > 0 {
		sqlStr += " LIMIT ? OFFSET ?"
		args = append(args, opts.Limit, opts.Offset)
	}
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var rows []struct {
		ID           uint64                         `gorm:"column:f_id"`
		FileID       uint64                         `gorm:"column:f_file_id"`
		Status       rds.FlowFileDownloadJobStatus  `gorm:"column:f_status"`
		RetryCount   int                            `gorm:"column:f_retry_count"`
		MaxRetry     int                            `gorm:"column:f_max_retry"`
		NextRetryAt  int64                          `gorm:"column:f_next_retry_at"`
		ErrorCode    string                         `gorm:"column:f_error_code"`
		ErrorMessage string                         `gorm:"column:f_error_message"`
		DownloadURL  string                         `gorm:"column:f_download_url"`
		StartedAt    int64                          `gorm:"column:f_started_at"`
		FinishedAt   int64                          `gorm:"column:f_finished_at"`
		CreatedAt    int64                          `gorm:"column:f_created_at"`
		UpdatedAt    int64                          `gorm:"column:f_updated_at"`
	}
	err = d.db.Raw(sqlStr, args...).Scan(&rows).Error
	if err != nil {
		log.Warnf("[FlowFileDownloadJobDao.List] query err: %s", err.Error())
		return nil, err
	}
	result = make([]*rds.FlowFileDownloadJob, 0, len(rows))
	for _, row := range rows {
		result = append(result, &rds.FlowFileDownloadJob{
			ID:           row.ID,
			FileID:       row.FileID,
			Status:       row.Status,
			RetryCount:   row.RetryCount,
			MaxRetry:     row.MaxRetry,
			NextRetryAt:  row.NextRetryAt,
			ErrorCode:    row.ErrorCode,
			ErrorMessage: row.ErrorMessage,
			DownloadURL:  row.DownloadURL,
			StartedAt:    row.StartedAt,
			FinishedAt:   row.FinishedAt,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return result, nil
}

func (d *flowFileDownloadJobDao) Update(ctx context.Context, id uint64, params *rds.FlowFileDownloadJobUpdateParams) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	var setClauses []string
	var args []interface{}

	if params.Status != nil {
		setClauses = append(setClauses, "f_status = ?")
		args = append(args, *params.Status)
	}
	if params.RetryCount != nil {
		setClauses = append(setClauses, "f_retry_count = ?")
		args = append(args, *params.RetryCount)
	}
	if params.NextRetryAt != nil {
		setClauses = append(setClauses, "f_next_retry_at = ?")
		args = append(args, *params.NextRetryAt)
	}
	if params.ErrorCode != nil {
		setClauses = append(setClauses, "f_error_code = ?")
		args = append(args, *params.ErrorCode)
	}
	if params.ErrorMessage != nil {
		setClauses = append(setClauses, "f_error_message = ?")
		args = append(args, *params.ErrorMessage)
	}
	if params.StartedAt != nil {
		setClauses = append(setClauses, "f_started_at = ?")
		args = append(args, *params.StartedAt)
	}
	if params.FinishedAt != nil {
		setClauses = append(setClauses, "f_finished_at = ?")
		args = append(args, *params.FinishedAt)
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "f_updated_at = ?")
	args = append(args, time.Now().Unix())
	args = append(args, id)

	setClause := strings.Join(setClauses, ", ")
	sqlStr := fmt.Sprintf("UPDATE t_flow_file_download_job SET %s WHERE f_id = ?", setClause)
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, args...)
	if result.Error != nil {
		log.Warnf("[FlowFileDownloadJobDao.Update] update err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

// ClaimJob 乐观锁抢占任务
func (d *flowFileDownloadJobDao) ClaimJob(ctx context.Context, id uint64, startedAt int64) (claimed bool, err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	runningStatus := rds.FlowFileDownloadJobStatusRunning
	pendingStatus := rds.FlowFileDownloadJobStatusPending
	now := time.Now().Unix()

	sqlStr := "UPDATE t_flow_file_download_job SET f_status = ?, f_started_at = ?, f_updated_at = ? WHERE f_id = ? AND f_status = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, runningStatus, startedAt, now, id, pendingStatus)
	if result.Error != nil {
		log.Warnf("[FlowFileDownloadJobDao.ClaimJob] claim err: %s", result.Error.Error())
		return false, result.Error
	}

	if result.RowsAffected == 0 {
		log.Infof("[FlowFileDownloadJobDao.ClaimJob] job %d already claimed by another worker", id)
		return false, nil
	}

	log.Infof("[FlowFileDownloadJobDao.ClaimJob] job %d claimed successfully", id)
	return true, nil
}

func (d *flowFileDownloadJobDao) Delete(ctx context.Context, id uint64) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_flow_file_download_job WHERE f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, id)
	if result.Error != nil {
		log.Warnf("[FlowFileDownloadJobDao.Delete] delete err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}

func (d *flowFileDownloadJobDao) DeleteByFileID(ctx context.Context, fileID uint64) (err error) {
	newCtx, span := trace.StartInternalSpan(ctx)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.FLOW_FILE_DOWNLOAD_JOB_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	sqlStr := "DELETE FROM t_flow_file_download_job WHERE f_file_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	result := d.db.Exec(sqlStr, fileID)
	if result.Error != nil {
		log.Warnf("[FlowFileDownloadJobDao.DeleteByFileID] delete err: %s", result.Error.Error())
		err = result.Error
	}
	return err
}