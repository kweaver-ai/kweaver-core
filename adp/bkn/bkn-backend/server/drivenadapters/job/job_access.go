// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package job

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

const (
	OT_TABLE_NAME   = "t_object_type"
	RT_TABLE_NAME   = "t_relation_type"
	JOB_TABLE_NAME  = "t_kn_job"
	TASK_TABLE_NAME = "t_kn_task"
)

var (
	jAccessOnce sync.Once
	jAccess     interfaces.JobAccess
)

type jobAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewJobAccess(appSetting *common.AppSetting) interfaces.JobAccess {
	jAccessOnce.Do(func() {
		jAccess = &jobAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return jAccess
}

// 创建job
func (ja *jobAccess) CreateJob(ctx context.Context, tx *sql.Tx, jobInfo *interfaces.JobInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateJob", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	jobConceptConfigStr, err := sonic.MarshalString(jobInfo.JobConceptConfig)
	if err != nil {
		logger.Errorf("Failed to marshal job concept config, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to marshal job concept config, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Marshal job concept config failed ")
		return err
	}

	sqlStr, vals, err := sq.Insert(JOB_TABLE_NAME).
		Columns(
			"f_id",
			"f_name",
			"f_kn_id",
			"f_branch",
			"f_job_type",
			"f_job_concept_config",
			"f_state",
			"f_state_detail",
			"f_creator",
			"f_creator_type",
			"f_create_time",
		).
		Values(
			jobInfo.ID,
			jobInfo.Name,
			jobInfo.KNID,
			jobInfo.Branch,
			jobInfo.JobType,
			jobConceptConfigStr,
			jobInfo.State,
			jobInfo.StateDetail,
			jobInfo.Creator.ID,
			jobInfo.Creator.Type,
			jobInfo.CreateTime,
		).ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert job, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert job, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建job的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除jobs
func (ja *jobAccess) DeleteJobsByIDs(ctx context.Context, tx *sql.Tx, jobIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteJobsByIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(jobIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(JOB_TABLE_NAME).
		Where(sq.Eq{"f_id": jobIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete jobs, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete jobs, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除job的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (ja *jobAccess) DeleteTasksByJobIDs(ctx context.Context, tx *sql.Tx, jobIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteTasksByJobIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(jobIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(TASK_TABLE_NAME).
		Where(sq.Eq{"f_job_id": jobIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete tasks, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete tasks, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除task的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

// 更新job状态
func (ja *jobAccess) UpdateJobState(ctx context.Context, tx *sql.Tx, jobID string, stateInfo interfaces.JobStateInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateJobState", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(stateInfo.StateDetail) > interfaces.MAX_STATE_DETAIL_SIZE {
		stateInfo.StateDetail = stateInfo.StateDetail[:interfaces.MAX_STATE_DETAIL_SIZE]
	}
	builder := sq.Update(JOB_TABLE_NAME).
		Set("f_state", stateInfo.State).
		Set("f_state_detail", stateInfo.StateDetail).
		Where(sq.Eq{"f_id": jobID})

	if stateInfo.FinishTime != 0 {
		builder = builder.Set("f_finish_time", stateInfo.FinishTime)
	}
	if stateInfo.TimeCost != 0 {
		builder = builder.Set("f_time_cost", stateInfo.TimeCost)
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update job status, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update job status, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("更新job状态的 sql 语句: %s", sqlStr))

	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlStr, vals...)
	} else {
		_, err = ja.db.ExecContext(ctx, sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("update data error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 根据ID查询job
func (ja *jobAccess) GetJobByID(ctx context.Context, jobID string) (*interfaces.JobInfo, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetJobByID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if jobID == "" {
		return nil, nil
	}

	query := sq.Select(
		"f_id",
		"f_name",
		"f_kn_id",
		"f_branch",
		"f_job_type",
		"f_job_concept_config",
		"f_state",
		"f_state_detail",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_finish_time",
		"f_time_cost",
	).From(JOB_TABLE_NAME).
		Where(sq.Eq{"f_id": jobID})

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get job, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get jobs, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("根据ID查询job的 sql 语句: %s", sqlStr))

	row := ja.db.QueryRowContext(ctx, sqlStr, vals...)

	var jobConceptConfigStr string
	jobInfo := interfaces.JobInfo{}
	err = row.Scan(
		&jobInfo.ID,
		&jobInfo.Name,
		&jobInfo.KNID,
		&jobInfo.Branch,
		&jobInfo.JobType,
		&jobConceptConfigStr,
		&jobInfo.State,
		&jobInfo.StateDetail,
		&jobInfo.Creator.ID,
		&jobInfo.Creator.Type,
		&jobInfo.CreateTime,
		&jobInfo.FinishTime,
		&jobInfo.TimeCost,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		logger.Errorf("scan data error: %v\n", err)
		span.SetStatus(codes.Error, "Scan data error")
		o11y.Error(ctx, fmt.Sprintf("Scan data error: %v ", err))
		return nil, err
	}

	err = sonic.UnmarshalString(jobConceptConfigStr, &jobInfo.JobConceptConfig)
	if err != nil {
		logger.Errorf("unmarshal job concept config error: %v\n", err)
		span.SetStatus(codes.Error, "Unmarshal job concept config error")
		o11y.Error(ctx, fmt.Sprintf("Unmarshal job concept config error: %v ", err))
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return &jobInfo, nil
}

// 根据ID查询job
func (ja *jobAccess) GetJobsByIDs(ctx context.Context, jobIDs []string) (map[string]*interfaces.JobInfo, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetJobsByIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(jobIDs) == 0 {
		return map[string]*interfaces.JobInfo{}, nil
	}

	query := sq.Select(
		"f_id",
		"f_name",
		"f_kn_id",
		"f_branch",
		"f_job_type",
		"f_job_concept_config",
		"f_state",
		"f_state_detail",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_finish_time",
		"f_time_cost",
	).From(JOB_TABLE_NAME).
		Where(sq.Eq{"f_id": jobIDs})

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get jobs, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get jobs, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("根据ID查询job的 sql 语句: %s", sqlStr))

	rows, err := ja.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("query data error: %v\n", err)
		span.SetStatus(codes.Error, "Query data error")
		o11y.Error(ctx, fmt.Sprintf("Query data error: %v ", err))
		return nil, err
	}
	defer rows.Close()

	jobInfos := map[string]*interfaces.JobInfo{}
	for rows.Next() {
		var jobConceptConfigStr string
		var jobInfo interfaces.JobInfo
		err := rows.Scan(
			&jobInfo.ID,
			&jobInfo.Name,
			&jobInfo.KNID,
			&jobInfo.Branch,
			&jobInfo.JobType,
			&jobConceptConfigStr,
			&jobInfo.State,
			&jobInfo.StateDetail,
			&jobInfo.Creator.ID,
			&jobInfo.Creator.Type,
			&jobInfo.CreateTime,
			&jobInfo.FinishTime,
			&jobInfo.TimeCost,
		)
		if err != nil {
			logger.Errorf("scan data error: %v\n", err)
			span.SetStatus(codes.Error, "Scan data error")
			o11y.Error(ctx, fmt.Sprintf("Scan data error: %v ", err))
			return nil, err
		}

		err = sonic.UnmarshalString(jobConceptConfigStr, &jobInfo.JobConceptConfig)
		if err != nil {
			logger.Errorf("unmarshal job concept config error: %v\n", err)
			span.SetStatus(codes.Error, "Unmarshal job concept config error")
			o11y.Error(ctx, fmt.Sprintf("Unmarshal job concept config error: %v ", err))
			return nil, err
		}
		jobInfos[jobInfo.ID] = &jobInfo
	}

	span.SetStatus(codes.Ok, "")
	return jobInfos, nil
}

// 根据ID查询job
func (ja *jobAccess) GetJobIDsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetJobIDsByKnID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	query := sq.Select(
		"f_id",
	).From(JOB_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch})

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get jobs, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get jobs, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("根据ID查询job的 sql 语句: %s", sqlStr))

	rows, err := ja.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("query data error: %v\n", err)
		span.SetStatus(codes.Error, "Query data error")
		o11y.Error(ctx, fmt.Sprintf("Query data error: %v ", err))
		return nil, err
	}
	defer rows.Close()

	jobIDs := []string{}
	for rows.Next() {
		var jobID string
		err := rows.Scan(
			&jobID,
		)
		if err != nil {
			logger.Errorf("scan data error: %v\n", err)
			span.SetStatus(codes.Error, "Scan data error")
			o11y.Error(ctx, fmt.Sprintf("Scan data error: %v ", err))
			return nil, err
		}
		jobIDs = append(jobIDs, jobID)
	}

	span.SetStatus(codes.Ok, "")
	return jobIDs, nil
}

// 更新task状态
func (ja *jobAccess) UpdateTaskState(ctx context.Context, taskID string, stateInfo interfaces.TaskStateInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateTaskState", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(stateInfo.StateDetail) > interfaces.MAX_STATE_DETAIL_SIZE {
		stateInfo.StateDetail = stateInfo.StateDetail[:interfaces.MAX_STATE_DETAIL_SIZE]
	}
	builder := sq.Update(TASK_TABLE_NAME).
		Where(sq.Eq{"f_id": taskID})
	if stateInfo.State != "" {
		builder = builder.Set("f_state", stateInfo.State)
	}
	if stateInfo.StateDetail != "" {
		builder = builder.Set("f_state_detail", stateInfo.StateDetail)
	}
	if stateInfo.Index != "" {
		builder = builder.Set("f_index", stateInfo.Index)
	}
	if stateInfo.DocCount != 0 {
		builder = builder.Set("f_doc_count", stateInfo.DocCount)
	}
	if stateInfo.StartTime != 0 {
		builder = builder.Set("f_start_time", stateInfo.StartTime)
	}
	if stateInfo.FinishTime != 0 {
		builder = builder.Set("f_finish_time", stateInfo.FinishTime)
	}
	if stateInfo.TimeCost != 0 {
		builder = builder.Set("f_time_cost", stateInfo.TimeCost)
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update task status, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update task status, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("更新task状态的 sql 语句: %s", sqlStr))

	_, err = ja.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("update data error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询job列表
func (ja *jobAccess) ListJobs(ctx context.Context, query interfaces.JobsQueryParams) ([]*interfaces.JobInfo, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListJobs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Select(
		"f_id",
		"f_name",
		"f_kn_id",
		"f_branch",
		"f_job_type",
		"f_job_concept_config",
		"f_state",
		"f_state_detail",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_finish_time",
		"f_time_cost",
	).From(JOB_TABLE_NAME)

	if query.KNID != "" {
		builder = builder.Where(sq.Eq{"f_kn_id": query.KNID})
	}

	// 过滤job名称
	if query.NamePattern != "" {
		builder = builder.Where(sq.Like{"f_name": fmt.Sprintf("%%%s%%", query.NamePattern)})
	}
	if query.JobType != "" {
		builder = builder.Where(sq.Eq{"f_job_type": query.JobType})
	}
	if len(query.State) > 0 {
		builder = builder.Where(sq.Eq{"f_state": query.State})
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
		logger.Errorf("Failed to build the sql of list jobs, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of list jobs, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询job列表的 sql 语句: %s", sqlStr))
	logger.Debugf("查询job列表的 sql 语句: %s", sqlStr)

	rows, err := ja.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("query data error: %v\n", err)
		span.SetStatus(codes.Error, "Query data error")
		o11y.Error(ctx, fmt.Sprintf("Query data error: %v ", err))
		return nil, err
	}
	defer rows.Close()

	jobInfos := []*interfaces.JobInfo{}
	for rows.Next() {
		var jobConceptConfigStr string
		jobInfo := interfaces.JobInfo{}
		err := rows.Scan(
			&jobInfo.ID,
			&jobInfo.Name,
			&jobInfo.KNID,
			&jobInfo.Branch,
			&jobInfo.JobType,
			&jobConceptConfigStr,
			&jobInfo.State,
			&jobInfo.StateDetail,
			&jobInfo.Creator.ID,
			&jobInfo.Creator.Type,
			&jobInfo.CreateTime,
			&jobInfo.FinishTime,
			&jobInfo.TimeCost,
		)
		if err != nil {
			logger.Errorf("scan data error: %v\n", err)
			span.SetStatus(codes.Error, "Scan data error")
			o11y.Error(ctx, fmt.Sprintf("Scan data error: %v ", err))
			return nil, err
		}

		err = sonic.UnmarshalString(jobConceptConfigStr, &jobInfo.JobConceptConfig)
		if err != nil {
			logger.Errorf("unmarshal job concept config error: %v\n", err)
			span.SetStatus(codes.Error, "Unmarshal job concept config error")
			o11y.Error(ctx, fmt.Sprintf("Unmarshal job concept config error: %v ", err))
			return nil, err
		}
		jobInfos = append(jobInfos, &jobInfo)
	}

	span.SetStatus(codes.Ok, "")
	return jobInfos, nil
}

// 查询job总数
func (ja *jobAccess) GetJobsTotal(ctx context.Context, queryParams interfaces.JobsQueryParams) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetJobsTotal", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	query := sq.Select("COUNT(*)").From(JOB_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": queryParams.KNID})

	// 过滤job名称
	if queryParams.NamePattern != "" {
		query = query.Where(sq.Like{"f_name": fmt.Sprintf("%%%s%%", queryParams.NamePattern)})
	}
	if len(queryParams.State) > 0 {
		query = query.Where(sq.Eq{"f_state": queryParams.State})
	}

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get jobs total, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get jobs total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询job总数的 sql 语句: %s", sqlStr))

	var total int64
	err = ja.db.QueryRowContext(ctx, sqlStr, vals...).Scan(&total)
	if err != nil {
		logger.Errorf("query data error: %v\n", err)
		span.SetStatus(codes.Error, "Query data error")
		o11y.Error(ctx, fmt.Sprintf("Query data error: %v ", err))
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 批量创建tasks
func (ja *jobAccess) CreateTasks(ctx context.Context, tx *sql.Tx, taskInfos map[string]*interfaces.TaskInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateTasks", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(taskInfos) == 0 {
		return nil
	}

	builder := sq.Insert(TASK_TABLE_NAME).
		Columns(
			"f_id",
			"f_name",
			"f_job_id",
			"f_concept_type",
			"f_concept_id",
			"f_state",
			"f_state_detail",
		)

	for _, task := range taskInfos {
		builder = builder.Values(
			task.ID,
			task.Name,
			task.JobID,
			task.ConceptType,
			task.ConceptID,
			task.State,
			task.StateDetail,
		)
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert tasks, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert tasks, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("批量创建tasks的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询task列表
func (ja *jobAccess) ListTasks(ctx context.Context, query interfaces.TasksQueryParams) ([]*interfaces.TaskInfo, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListTasks", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Select(
		"f_id",
		"f_name",
		"f_job_id",
		"f_concept_type",
		"f_concept_id",
		"f_index",
		"f_doc_count",
		"f_state",
		"f_state_detail",
		"f_start_time",
		"f_finish_time",
		"f_time_cost",
	).From(TASK_TABLE_NAME)

	if query.JobID != "" {
		builder = builder.Where(sq.Eq{"f_job_id": query.JobID})
	}

	if query.NamePattern != "" {
		builder = builder.Where(sq.Like{"f_name": fmt.Sprintf("%%%s%%", query.NamePattern)})
	}
	if query.ConceptType != "" {
		builder = builder.Where(sq.Eq{"f_concept_type": query.ConceptType})
	}
	if len(query.State) != 0 {
		builder = builder.Where(sq.Eq{"f_state": query.State})
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
		logger.Errorf("Failed to build the sql of list tasks, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of list tasks, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询task列表的 sql 语句: %s", sqlStr))
	logger.Debugf("查询task列表的 sql 语句: %s", sqlStr)

	rows, err := ja.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("query data error: %v\n", err)
		span.SetStatus(codes.Error, "Query data error")
		o11y.Error(ctx, fmt.Sprintf("Query data error: %v ", err))
		return nil, err
	}
	defer rows.Close()

	taskInfos := []*interfaces.TaskInfo{}
	for rows.Next() {
		taskInfo := interfaces.TaskInfo{}
		err := rows.Scan(
			&taskInfo.ID,
			&taskInfo.Name,
			&taskInfo.JobID,
			&taskInfo.ConceptType,
			&taskInfo.ConceptID,
			&taskInfo.Index,
			&taskInfo.DocCount,
			&taskInfo.State,
			&taskInfo.StateDetail,
			&taskInfo.StartTime,
			&taskInfo.FinishTime,
			&taskInfo.TimeCost,
		)
		if err != nil {
			logger.Errorf("scan data error: %v\n", err)
			span.SetStatus(codes.Error, "Scan data error")
			o11y.Error(ctx, fmt.Sprintf("Scan data error: %v ", err))
			return nil, err
		}
		taskInfos = append(taskInfos, &taskInfo)
	}

	span.SetStatus(codes.Ok, "")
	return taskInfos, nil
}

// 查询task总数
func (ja *jobAccess) GetTasksTotal(ctx context.Context, queryParams interfaces.TasksQueryParams) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetTasksTotal", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	query := sq.Select(
		"count(*)",
	).From(TASK_TABLE_NAME).
		Where(sq.Eq{"f_job_id": queryParams.JobID})

	if queryParams.ConceptType != "" {
		query = query.Where(sq.Eq{"f_concept_type": queryParams.ConceptType})
	}
	if queryParams.NamePattern != "" {
		query = query.Where(sq.Like{"f_name": fmt.Sprintf("%%%s%%", queryParams.NamePattern)})
	}
	if len(queryParams.State) != 0 {
		query = query.Where(sq.Eq{"f_state": queryParams.State})
	}

	sqlStr, vals, err := query.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get tasks total, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get tasks total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询task总数的 sql 语句: %s", sqlStr))

	var total int64
	err = ja.db.QueryRowContext(ctx, sqlStr, vals...).Scan(&total)
	if err != nil {
		logger.Errorf("query data error: %v\n", err)
		span.SetStatus(codes.Error, "Query data error")
		o11y.Error(ctx, fmt.Sprintf("Query data error: %v ", err))
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}
