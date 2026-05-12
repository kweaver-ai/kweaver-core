// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

const (
	METRIC_TABLE_NAME = "t_metric_definition"
)

var (
	metricAccessOnce sync.Once
	metricAccessInst interfaces.MetricAccess
)

type metricAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewMetricAccess(appSetting *common.AppSetting) interfaces.MetricAccess {
	metricAccessOnce.Do(func() {
		metricAccessInst = &metricAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return metricAccessInst
}

func jsonOrNull(v any) interface{} {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		logger.Errorf("metric jsonOrNull marshal error: %v", err)
		return nil
	}
	return string(b)
}

func scanMetricFromRow(scanner interface {
	Scan(dest ...any) error
}) (*interfaces.MetricDefinition, error) {

	metric := &interfaces.MetricDefinition{}
	var (
		timeDim, calcFormula, analysisDim sql.NullString
		tags                              sql.NullString
	)
	err := scanner.Scan(
		&metric.ID,
		&metric.KnID,
		&metric.Branch,
		&metric.Name,
		&metric.Comment,
		&tags,
		&metric.Icon,
		&metric.Color,
		&metric.BKNRawContent,
		&metric.UnitType,
		&metric.Unit,
		&metric.MetricType,
		&metric.ScopeType,
		&metric.ScopeRef,
		&timeDim,
		&calcFormula,
		&analysisDim,
		&metric.Creator.ID,
		&metric.Creator.Type,
		&metric.CreateTime,
		&metric.Updater.ID,
		&metric.Updater.Type,
		&metric.UpdateTime,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Errorf("metric row scan error: %v", err)
		}
		return nil, err
	}

	if tags.Valid {
		metric.Tags = libCommon.TagString2TagSlice(tags.String)
	}

	if timeDim.Valid && timeDim.String != "" {
		var td interfaces.MetricTimeDimension
		if err := json.Unmarshal([]byte(timeDim.String), &td); err != nil {
			logger.Errorf("unmarshal metric f_time_dimension: %v", err)
			return nil, fmt.Errorf("unmarshal metric time_dimension: %w", err)
		}
		metric.TimeDimension = &td
	}
	if calcFormula.Valid && calcFormula.String != "" {
		if err := json.Unmarshal([]byte(calcFormula.String), &metric.CalculationFormula); err != nil {
			logger.Errorf("unmarshal metric f_calculation_formula: %v", err)
			return nil, fmt.Errorf("unmarshal metric calculation_formula: %w", err)
		}
	}
	if analysisDim.Valid && analysisDim.String != "" {
		if err := json.Unmarshal([]byte(analysisDim.String), &metric.AnalysisDimensions); err != nil {
			logger.Errorf("unmarshal metric f_analysis_dimensions: %v", err)
			return nil, fmt.Errorf("unmarshal metric analysis_dimensions: %w", err)
		}
	}
	return metric, nil
}

func (ma *metricAccess) CreateMetric(ctx context.Context, tx *sql.Tx, def *interfaces.MetricDefinition) error {
	_, span := oteltrace.StartNamedClientSpan(ctx, "CreateMetric")
	defer span.End()

	td := jsonOrNull(def.TimeDimension)
	cf := jsonOrNull(def.CalculationFormula)
	ad := jsonOrNull(def.AnalysisDimensions)
	tagsStr := libCommon.TagSlice2TagString(def.Tags)

	sqlStr, vals, err := sq.Insert(METRIC_TABLE_NAME).
		Columns(
			"f_id",
			"f_kn_id",
			"f_branch",
			"f_name",
			"f_comment",
			"f_tags",
			"f_icon",
			"f_color",
			"f_bkn_raw_content",
			"f_unit_type",
			"f_unit",
			"f_metric_type",
			"f_scope_type",
			"f_scope_ref",
			"f_time_dimension",
			"f_calculation_formula",
			"f_analysis_dimensions",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			def.ID,
			def.KnID,
			def.Branch,
			def.Name,
			def.Comment,
			tagsStr,
			def.Icon,
			def.Color,
			def.BKNRawContent,
			def.UnitType,
			def.Unit,
			def.MetricType,
			def.ScopeType,
			def.ScopeRef,
			td,
			cf,
			ad,
			def.Creator.ID,
			def.Creator.Type,
			def.CreateTime,
			def.Updater.ID,
			def.Updater.Type,
			def.UpdateTime,
		).
		ToSql()
	if err != nil {
		logger.Errorf("CreateMetric build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if tx != nil {
		_, err = tx.Exec(sqlStr, vals...)
	} else {
		_, err = ma.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("CreateMetric insert metric definition error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ma *metricAccess) CheckMetricExistByID(ctx context.Context, knID string, branch string, metricID string) (string, bool, error) {

	_, span := oteltrace.StartNamedClientSpan(ctx, "CheckMetricExistByID")
	defer span.End()

	sqlStr, vals, err := sq.Select("f_name").
		From(METRIC_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": metricID}).
		ToSql()
	if err != nil {
		logger.Errorf("CheckMetricExistByID build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, err
	}
	var name string
	err = ma.db.QueryRow(sqlStr, vals...).Scan(&name)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		logger.Errorf("CheckMetricExistByID scan error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, err
	}
	return name, true, nil
}

func (ma *metricAccess) CheckMetricExistByName(ctx context.Context, knID string, branch string, name string) (string, bool, error) {

	_, span := oteltrace.StartNamedClientSpan(ctx, "CheckMetricExistByName")
	defer span.End()

	sqlStr, vals, err := sq.Select("f_id").
		From(METRIC_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_name": name}).
		ToSql()
	if err != nil {
		logger.Errorf("CheckMetricExistByName build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, err
	}
	var id string
	err = ma.db.QueryRow(sqlStr, vals...).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		logger.Errorf("CheckMetricExistByName scan error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, err
	}
	return id, true, nil
}

func metricSelectColumns() []string {
	return []string{
		"f_id",
		"f_kn_id",
		"f_branch",
		"f_name",
		"f_comment",
		"f_tags",
		"f_icon",
		"f_color",
		"f_bkn_raw_content",
		"f_unit_type",
		"f_unit",
		"f_metric_type",
		"f_scope_type",
		"f_scope_ref",
		"f_time_dimension",
		"f_calculation_formula",
		"f_analysis_dimensions",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	}
}

func (ma *metricAccess) GetMetricByID(ctx context.Context, knID string, branch string, metricID string) (*interfaces.MetricDefinition, error) {

	_, span := oteltrace.StartNamedClientSpan(ctx, "GetMetricByID")
	defer span.End()

	sqlStr, vals, err := sq.Select(metricSelectColumns()...).
		From(METRIC_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": metricID}).
		ToSql()
	if err != nil {
		logger.Errorf("GetMetricByID build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	row := ma.db.QueryRow(sqlStr, vals...)
	metric, err := scanMetricFromRow(row)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	return metric, nil
}

func (ma *metricAccess) GetMetricsByIDs(ctx context.Context, knID string, branch string, metricIDs []string) ([]*interfaces.MetricDefinition, error) {

	_, span := oteltrace.StartNamedClientSpan(ctx, "GetMetricsByIDs")
	defer span.End()

	if len(metricIDs) == 0 {
		return []*interfaces.MetricDefinition{}, nil
	}

	sqlStr, vals, err := sq.Select(metricSelectColumns()...).
		From(METRIC_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": metricIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("GetMetricsByIDs build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	rows, err := ma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("GetMetricsByIDs query error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var metrics []*interfaces.MetricDefinition
	for rows.Next() {
		metric, err := scanMetricFromRow(rows)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("GetMetricsByIDs rows error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	return metrics, nil
}

func processMetricQueryCondition(query interfaces.MetricsListQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
	if query.NamePattern != "" {
		subBuilder = subBuilder.Where(sq.Expr("(instr(f_name, ?) > 0 OR instr(f_id, ?) > 0)",
			query.NamePattern, query.NamePattern))
	}
	if query.KNID != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_kn_id": query.KNID})
	}
	if query.Branch != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_branch": query.Branch})
	} else {
		subBuilder = subBuilder.Where(sq.Eq{"f_branch": interfaces.MAIN_BRANCH})
	}
	// 统计主体类型
	if query.ScopeType != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_scope_type": query.ScopeType})
	}
	// 统计主体id
	if query.ScopeRef != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_scope_ref": query.ScopeRef})
	}
	if query.Tag != "" {
		subBuilder = subBuilder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+query.Tag+`"`))
	}
	return subBuilder
}

func (ma *metricAccess) ListMetrics(ctx context.Context, query interfaces.MetricsListQueryParams) ([]*interfaces.MetricDefinition, error) {
	_, span := oteltrace.StartNamedClientSpan(ctx, "ListMetrics")
	defer span.End()

	subBuilder := sq.Select(metricSelectColumns()...).From(METRIC_TABLE_NAME)
	builder := processMetricQueryCondition(query, subBuilder)
	if query.Sort != "" {
		sortCol := query.Sort
		dir := query.Direction
		if dir == "" {
			dir = interfaces.DESC_DIRECTION
		}
		builder = builder.OrderBy(fmt.Sprintf("%s %s", sortCol, dir))
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("ListMetrics build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	rows, err := ma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("ListMetrics query error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var metrics []*interfaces.MetricDefinition
	for rows.Next() {
		metric, err := scanMetricFromRow(rows)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("ListMetrics rows error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	return metrics, nil
}

func (ma *metricAccess) GetMetricsTotal(ctx context.Context, query interfaces.MetricsListQueryParams) (int, error) {
	_, span := oteltrace.StartNamedClientSpan(ctx, "GetMetricsTotal")
	defer span.End()

	subBuilder := sq.Select("COUNT(f_id)").From(METRIC_TABLE_NAME)
	builder := processMetricQueryCondition(query, subBuilder)
	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("GetMetricsTotal build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	var total int
	err = ma.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		logger.Errorf("GetMetricsTotal scan error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	span.SetStatus(codes.Ok, "")
	return total, nil
}

func (ma *metricAccess) UpdateMetric(ctx context.Context, tx *sql.Tx, metric *interfaces.MetricDefinition) error {
	_, span := oteltrace.StartNamedClientSpan(ctx, "UpdateMetric")
	defer span.End()

	data := map[string]any{
		"f_comment":             metric.Comment,
		"f_tags":                libCommon.TagSlice2TagString(metric.Tags),
		"f_icon":                metric.Icon,
		"f_color":               metric.Color,
		"f_bkn_raw_content":     metric.BKNRawContent,
		"f_unit_type":           metric.UnitType,
		"f_unit":                metric.Unit,
		"f_metric_type":         metric.MetricType,
		"f_time_dimension":      jsonOrNull(metric.TimeDimension),
		"f_calculation_formula": jsonOrNull(metric.CalculationFormula),
		"f_analysis_dimensions": jsonOrNull(metric.AnalysisDimensions),
		"f_updater":             metric.Updater.ID,
		"f_updater_type":        metric.Updater.Type,
		"f_update_time":         metric.UpdateTime,
	}

	sqlStr, vals, err := sq.Update(METRIC_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": metric.ID}).
		Where(sq.Eq{"f_kn_id": metric.KnID}).
		Where(sq.Eq{"f_branch": metric.Branch}).
		ToSql()
	if err != nil {
		logger.Errorf("UpdateMetric build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update metric definition error: %v\n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Update metric definition error: %v", err.Error()))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Get RowsAffected error: %v", err.Error()))
		return err
	}

	if RowsAffected != 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected not equal 1, RowsAffected is %d, Metric is %v",
			metric.ID, RowsAffected, metric)
		span.SetStatus(codes.Error, fmt.Sprintf("Update %s RowsAffected not equal 1, RowsAffected is %d, Metric is %v",
			metric.ID, RowsAffected, metric))
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ma *metricAccess) DeleteMetricsByIDs(ctx context.Context, tx *sql.Tx, knID, branch string, metricIDs []string) error {
	_, span := oteltrace.StartNamedClientSpan(ctx, "DeleteMetricsByIDs")
	defer span.End()

	if len(metricIDs) == 0 {
		return nil
	}
	sqlStr, vals, err := sq.Delete(METRIC_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": metricIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("DeleteMetricsByIDs build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete metric definition error: %v\n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Delete metric definition error: %v", err.Error()))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Delete Metrics By IDs[%v] RowsAffected error: %v\n", metricIDs, err)
		span.SetStatus(codes.Error, fmt.Sprintf("Delete Metrics By IDs[%v] RowsAffected error: %v", metricIDs, err.Error()))
		return err
	}

	if RowsAffected != int64(len(metricIDs)) {
		// 影响行数不等于删除的指标数量不报错，删除操作已经发生
		logger.Warnf("DELETE metrics by ids[%v] RowsAffected not equal %d",
			metricIDs, RowsAffected)
		span.SetStatus(codes.Error, fmt.Sprintf("Delete metrics by ids[%v] RowsAffected not equal %d",
			metricIDs, RowsAffected))
	}
	logger.Infof("Delete Metrics By IDs[%v] RowsAffected: %d", metricIDs, RowsAffected)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ma *metricAccess) GetMetricIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	_, span := oteltrace.StartNamedClientSpan(ctx, "GetMetricIDsByKnID")
	defer span.End()

	sqlStr, vals, err := sq.Select("f_id").
		From(METRIC_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("GetMetricIDsByKnID build sql error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	rows, err := ma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("GetMetricIDsByKnID query error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			logger.Errorf("GetMetricIDsByKnID scan error: %v", err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("GetMetricIDsByKnID rows error: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	return ids, nil
}

func (ma *metricAccess) DeleteMetricsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	_, span := oteltrace.StartNamedClientSpan(ctx, "DeleteMetricsByKnID")
	defer span.End()

	sqlStr, vals, err := sq.Delete(METRIC_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("DeleteMetricsByKnID by kn_id=%s branch=%s build sql error: %v", knID, branch, err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("DeleteMetricsByKnID by kn_id=%s branch=%s exec error: %v", knID, branch, err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("DeleteMetricsByKnID by kn_id=%s branch=%s RowsAffected error: %v", knID, branch, err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	logger.Infof("DeleteMetricsByKnID by kn_id=%s branch=%s RowsAffected=%d", knID, branch, n)
	span.SetStatus(codes.Ok, "")
	return n, nil
}
