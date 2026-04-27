package drivenadapters

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libcomm "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

const (
	PIPELINE_TABLE_NAME = "t_stream_data_pipeline"

	DRIVER_NAME = "proton-rds"
)

var (
	pmAccessOnce sync.Once
	pmAccess     interfaces.PipelineMgmtAccess
	dbUrl        string
)

type pipelineMgmtAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewPipelineMgmtAccess(appSetting *common.AppSetting) interfaces.PipelineMgmtAccess {
	pmAccessOnce.Do(func() {
		pmAccess = &pipelineMgmtAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return pmAccess
}

func (pmAccess *pipelineMgmtAccess) CreatePipeline(ctx context.Context, pipeline *interfaces.Pipeline) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert pipeline into database", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
	)
	defer span.End()

	tagsStr := libcomm.TagSlice2TagString(pipeline.Tags)

	deployCfg, err := sonic.Marshal(pipeline.DeploymentConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal deployment config failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal deployment config failed")

		return err
	}

	insertMap := map[string]interface{}{
		"f_pipeline_id":             pipeline.PipelineID,
		"f_pipeline_name":           pipeline.PipelineName,
		"f_tags":                    tagsStr,
		"f_comment":                 pipeline.Comment,
		"f_builtin":                 pipeline.Builtin,
		"f_output_type":             pipeline.OutputType,
		"f_index_base":              pipeline.IndexBase,
		"f_use_index_base_in_data":  pipeline.UseIndexBaseInData,
		"f_pipeline_status":         pipeline.PipelineStatus,
		"f_pipeline_status_details": pipeline.PipelineStatusDetails,
		"f_deployment_config":       deployCfg,
		"f_create_time":             pipeline.CreateTime,
		"f_update_time":             pipeline.UpdateTime,
		"f_creator":                 pipeline.Creator.ID,
		"f_creator_type":            pipeline.Creator.Type,
		"f_updater":                 pipeline.Updater.ID,
		"f_updater_type":            pipeline.Updater.Type,
	}
	sqlStr, args, err := sq.Insert(PIPELINE_TABLE_NAME).SetMap(insertMap).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("failed to build the sql of creating pipeline, error: %s", err)
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "build the create pipeline sql")

		return err
	}

	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("sql stmt for creating pipeline is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = pmAccess.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("failed to insert pipeline info, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "execute create pipeline sql failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (pmAccess *pipelineMgmtAccess) DeletePipeline(ctx context.Context, pipelineID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete pipeline from database", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("pipeline_id").String(pipelineID),
	)
	defer span.End()

	sqlStr, args, err := sq.Delete(PIPELINE_TABLE_NAME).
		Where(sq.Eq{"f_pipeline_id": pipelineID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("failed to build the sql of deleting pipeline, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "build delete sql failed")

		return err
	}

	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("sql stmt for deleting pipeline is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = pmAccess.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("failed to delete pipeline, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "exec delete sql failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 更新管道配置，不更新管道状态
func (pmAccess *pipelineMgmtAccess) UpdatePipeline(ctx context.Context, pipeline *interfaces.Pipeline) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "update pipeline from database", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("pipeline_id").String(pipeline.PipelineID),
	)
	defer span.End()

	tagsStr := libcomm.TagSlice2TagString(pipeline.Tags)
	deployCfg, err := sonic.Marshal(pipeline.DeploymentConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal deployment config failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal deployment config failed")

		return err
	}

	updateMap := map[string]interface{}{
		"f_pipeline_name":          pipeline.PipelineName,
		"f_tags":                   tagsStr,
		"f_comment":                pipeline.Comment,
		"f_index_base":             pipeline.IndexBase,
		"f_use_index_base_in_data": pipeline.UseIndexBaseInData,
		"f_deployment_config":      deployCfg,
		"f_update_time":            pipeline.UpdateTime,
		"f_updater":                pipeline.Updater.ID,
		"f_updater_type":           pipeline.Updater.Type,
	}
	sqlStr, args, err := sq.Update(PIPELINE_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_pipeline_id": pipeline.PipelineID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("failed to build the sql of updating pipeline, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "build sql failed")

		return err
	}

	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("sql stmt for updating pipeline is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = pmAccess.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("failed to update pipeline, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "execute sql failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 获取单条管道配置
func (pmAccess *pipelineMgmtAccess) GetPipeline(ctx context.Context, pipelineID string) (*interfaces.Pipeline, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "get a pipeline from database", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("pipeline_id").String(pipelineID),
	)
	defer span.End()

	sqlStr, args, err := sq.Select(
		"f_pipeline_id",
		"f_pipeline_name",
		"f_tags",
		"f_comment",
		"f_builtin",
		"f_output_type",
		"f_index_base",
		"f_use_index_base_in_data",
		"f_pipeline_status",
		"f_pipeline_status_details",
		"f_deployment_config",
		"f_create_time",
		"f_update_time",
		"f_creator",
		"f_creator_type",
		"f_updater",
		"f_updater_type",
	).
		From(PIPELINE_TABLE_NAME).
		Where(sq.Eq{"f_pipeline_id": pipelineID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("failed to build the sql of retrieving pipeline, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "build sql failed")

		return nil, false, err
	}
	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("the sql of retrieving pipeline: %s", sqlStr)
	o11y.Info(ctx, sqlStmt)
	logger.Debug(sqlStmt)

	row := pmAccess.db.QueryRow(sqlStr, args...)

	var tagsStr string
	var deployCfg []byte
	pipeline := &interfaces.Pipeline{}
	err = row.Scan(
		&pipeline.PipelineID,
		&pipeline.PipelineName,
		&tagsStr,
		&pipeline.Comment,
		&pipeline.Builtin,
		&pipeline.OutputType,
		&pipeline.IndexBase,
		&pipeline.UseIndexBaseInData,
		&pipeline.PipelineStatus,
		&pipeline.PipelineStatusDetails,
		&deployCfg,
		&pipeline.CreateTime,
		&pipeline.UpdateTime,
		&pipeline.Creator.ID,
		&pipeline.Creator.Type,
		&pipeline.Updater.ID,
		&pipeline.Updater.Type,
	)

	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")

		return &interfaces.Pipeline{}, false, nil
	}

	if err != nil {
		logger.Errorf("row scan failed, error: %s", err.Error())

		span.SetStatus(codes.Error, "scan pipeline failed")
		o11y.Error(ctx, fmt.Sprintf("failed to scan pipeline, error: %v", err))

		return nil, false, err
	}

	pipeline.Tags = libcomm.TagString2TagSlice(tagsStr)

	// 反序列化
	err = sonic.Unmarshal([]byte(deployCfg), &pipeline.DeploymentConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Unmarshal deployment config failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Unmarshal deployment config failed")

		return nil, false, err
	}

	span.SetStatus(codes.Ok, "")
	return pipeline, true, nil
}

func (pmAccess *pipelineMgmtAccess) ListPipelines(ctx context.Context, query *interfaces.ListPipelinesQuery) ([]*interfaces.Pipeline, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "list pipelines from database", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("name_pattern").String(query.NamePattern),
		attr.Key("offset").String(fmt.Sprintf("%d", query.Offset)),
		attr.Key("limit").String(fmt.Sprintf("%d", query.Limit)),
		attr.Key("sort").String(query.Sort),
		attr.Key("direction").String(query.Direction),
		attr.Key("tag").String(query.Tag),
		attr.Key("builtin").String(fmt.Sprintf("%v", query.Builtin)),
		attr.Key("status").StringSlice(query.PipelineStatus),
	)
	defer span.End()

	pipelineList := make([]*interfaces.Pipeline, 0)

	builder := sq.Select(
		"f_pipeline_id",
		"f_pipeline_name",
		"f_tags",
		"f_comment",
		"f_builtin",
		"f_output_type",
		"f_index_base",
		"f_use_index_base_in_data",
		"f_pipeline_status",
		"f_pipeline_status_details",
		"f_deployment_config",
		"f_create_time",
		"f_update_time",
		"f_creator",
		"f_creator_type",
		"f_updater",
		"f_updater_type",
	).
		From(PIPELINE_TABLE_NAME)

	builder = jointPipelineListQuerySQL(query, builder)
	// 排序
	builder = builder.OrderBy(fmt.Sprintf("%s %s", query.Sort, query.Direction))

	// 添加分页参数 limit=-1时 不分页
	// if query.Limit != interfaces.NO_LIMIT {
	// 	builder = builder.Limit(uint64(query.Limit)).Offset(uint64(query.Offset))
	// }

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("failed to build the sql of list pipeline, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "build sql failed")

		return nil, err
	}

	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("the sql of listing pipelines: %s", sqlStr)
	o11y.Info(ctx, sqlStmt)
	logger.Debug(sqlStmt)

	rows, err := pmAccess.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("failed to list pipeline, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "query pipelines failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		pipeline := &interfaces.Pipeline{}
		var tagsStr string
		var deployCfg []byte
		err := rows.Scan(
			&pipeline.PipelineID,
			&pipeline.PipelineName,
			&tagsStr,
			&pipeline.Comment,
			&pipeline.Builtin,
			&pipeline.OutputType,
			&pipeline.IndexBase,
			&pipeline.UseIndexBaseInData,
			&pipeline.PipelineStatus,
			&pipeline.PipelineStatusDetails,
			&deployCfg,
			&pipeline.CreateTime,
			&pipeline.UpdateTime,
			&pipeline.Creator.ID,
			&pipeline.Creator.Type,
			&pipeline.Updater.ID,
			&pipeline.Updater.Type,
		)

		if err != nil {
			errDetails := fmt.Sprintf("failed to scan row, error: %s", err.Error())
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "scan row failed")

			return nil, err
		}

		pipeline.Tags = libcomm.TagString2TagSlice(tagsStr)

		err = sonic.Unmarshal([]byte(deployCfg), &pipeline.DeploymentConfig)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal deployment config failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Unmarshal deployment config failed")

			return nil, err
		}

		pipelineList = append(pipelineList, pipeline)
	}

	span.SetStatus(codes.Ok, "")
	return pipelineList, nil
}

func (pmAccess *pipelineMgmtAccess) GetPipelinesTotal(ctx context.Context, query *interfaces.ListPipelinesQuery) (int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "get the total of pipelines from database", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("name_pattern").String(query.NamePattern),
		attr.Key("tag").String(query.Tag),
		attr.Key("builtin").String(fmt.Sprintf("%v", query.Builtin)),
		attr.Key("status").StringSlice(query.PipelineStatus),
	)
	defer span.End()

	builder := sq.Select("COUNT(f_pipeline_id)").From(PIPELINE_TABLE_NAME)

	builder = jointPipelineListQuerySQL(query, builder)

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("failed to builder select pipeline totals sql, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "build sql failed")

		return 0, err
	}
	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("the sql of getting the total of pipelines: %s", sqlStr)
	o11y.Info(ctx, sqlStmt)
	logger.Debug(sqlStmt)

	var totals int
	err = pmAccess.db.QueryRow(sqlStr, vals...).Scan(&totals)
	if err != nil {
		errDetails := fmt.Sprintf("failed to scan totals, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "scan totals failed")

		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return totals, nil
}

// 拼接 sql 过滤条件
func jointPipelineListQuerySQL(query *interfaces.ListPipelinesQuery, builder sq.SelectBuilder) sq.SelectBuilder {
	if query.NamePattern != "" {
		builder = builder.Where(sq.Expr("instr(f_pipeline_name, ?) > 0", query.NamePattern))
	}

	if len(query.Builtin) > 0 {
		builder = builder.Where(sq.Eq{"f_builtin": query.Builtin})
	}

	// 拼接按标签过滤
	if query.Tag != "" {
		// 格式为: %"tagname"%
		builder = builder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+query.Tag+`"`))

	}

	// 管道状态过滤
	if len(query.PipelineStatus) != 0 {
		builder = builder.Where(sq.Eq{"f_pipeline_status": query.PipelineStatus})
	}

	return builder
}

func (pmAccess *pipelineMgmtAccess) UpdatePipelineStatus(ctx context.Context, pipeline *interfaces.Pipeline, isInnerRequest bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "update pipeline status from database", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("pipeline_id").String(pipeline.PipelineID))
	defer span.End()

	updateMap := map[string]interface{}{
		"f_pipeline_status":         pipeline.PipelineStatus,
		"f_pipeline_status_details": pipeline.PipelineStatusDetails,
	}
	if !isInnerRequest {
		updateMap["f_update_time"] = pipeline.UpdateTime
		updateMap["f_updater"] = pipeline.Updater.ID
		updateMap["f_updater_type"] = pipeline.Updater.Type
	}
	sqlStr, vals, err := sq.Update(PIPELINE_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_pipeline_id": pipeline.PipelineID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("failed to build the sql of updating pipeline status, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "build sql failed")

		return err
	}
	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("the sql of updating pipeline status: %s", sqlStr)
	o11y.Info(ctx, sqlStmt)
	logger.Debug(sqlStmt)

	_, err = pmAccess.db.Exec(sqlStr, vals...)
	if err != nil {
		errDetails := fmt.Sprintf("failed to update pipeline status, error: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "exec sql failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 根据ID获取管道
func (pmAccess *pipelineMgmtAccess) CheckPipelineExistByID(ctx context.Context, pipelineID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check pipeline exist by ID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("view_id").String(pipelineID),
	)

	sqlStr, args, err := sq.Select("f_pipeline_name").
		From(PIPELINE_TABLE_NAME).
		Where(sq.Eq{"f_pipeline_id": pipelineID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'check pipeline exist by ID' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return "", false, err
	}

	sqlStmt := fmt.Sprintf("sql stmt for checking view exists by ID is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var name string
	err = pmAccess.db.QueryRow(sqlStr, args...).Scan(&name)

	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return "", false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return name, true, nil
}

// 根据管道名称获取管道
func (pmAccess *pipelineMgmtAccess) CheckPipelineExistByName(ctx context.Context, pipelineName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check pipeline exist by name", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_type").String(os.Getenv("DB_TYPE")),
		attr.Key("db_url").String(dbUrl),
		attr.Key("pipeline_name").String(pipelineName),
	)

	sqlStr, args, err := sq.Select("f_pipeline_id").
		From(PIPELINE_TABLE_NAME).
		Where(sq.Eq{"f_pipeline_name": pipelineName}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'check pipeline exist by name' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return "", false, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for checking pipeline exists by name is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var pipelineID string
	err = pmAccess.db.QueryRow(sqlStr, args...).Scan(&pipelineID)

	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return "", false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return pipelineID, true, nil
}
