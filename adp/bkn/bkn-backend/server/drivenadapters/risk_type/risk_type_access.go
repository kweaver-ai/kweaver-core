// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package risk_type

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

const (
	RT_TABLE_NAME = "t_risk_type"
)

var (
	rtAccessOnce sync.Once
	rtAccess     interfaces.RiskTypeAccess
)

type riskTypeAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewRiskTypeAccess(appSetting *common.AppSetting) interfaces.RiskTypeAccess {
	rtAccessOnce.Do(func() {
		rtAccess = &riskTypeAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return rtAccess
}

func (rta *riskTypeAccess) CheckRiskTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error) {
	_, span := ar_trace.Tracer.Start(ctx, "CheckRiskTypeExistByID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	sqlStr, vals, err := sq.Select("f_name").
		From(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": rtID}).
		ToSql()
	if err != nil {
		return "", false, err
	}

	var name string
	err = rta.db.QueryRow(sqlStr, vals...).Scan(&name)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return name, true, nil
}

func (rta *riskTypeAccess) CheckRiskTypeExistByName(ctx context.Context, knID string, branch string, rtName string) (string, bool, error) {
	_, span := ar_trace.Tracer.Start(ctx, "CheckRiskTypeExistByName", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	sqlStr, vals, err := sq.Select("f_id").
		From(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_name": rtName}).
		ToSql()
	if err != nil {
		return "", false, err
	}

	var rtID string
	err = rta.db.QueryRow(sqlStr, vals...).Scan(&rtID)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return rtID, true, nil
}

func (rta *riskTypeAccess) CreateRiskType(ctx context.Context, tx *sql.Tx, riskType *interfaces.RiskType) error {
	_, span := ar_trace.Tracer.Start(ctx, "CreateRiskType", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	tagsStr := libCommon.TagSlice2TagString(riskType.Tags)

	sqlStr, vals, err := sq.Insert(RT_TABLE_NAME).
		Columns(
			"f_id",
			"f_name",
			"f_comment",
			"f_tags",
			"f_icon",
			"f_color",
			"f_bkn_raw_content",
			"f_kn_id",
			"f_branch",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			riskType.RTID,
			riskType.RTName,
			riskType.Comment,
			tagsStr,
			riskType.Icon,
			riskType.Color,
			riskType.BKNRawContent,
			riskType.KNID,
			riskType.Branch,
			riskType.Creator.ID,
			riskType.Creator.Type,
			riskType.CreateTime,
			riskType.Updater.ID,
			riskType.Updater.Type,
			riskType.UpdateTime,
		).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(sqlStr, vals...)
	return err
}

func (rta *riskTypeAccess) ListRiskTypes(ctx context.Context, query interfaces.RiskTypesQueryParams) ([]*interfaces.RiskType, error) {
	_, span := ar_trace.Tracer.Start(ctx, "ListRiskTypes", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	subBuilder := sq.Select(
		"f_id",
		"f_name",
		"f_comment",
		"f_tags",
		"f_icon",
		"f_color",
		"f_bkn_raw_content",
		"f_kn_id",
		"f_branch",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RT_TABLE_NAME)

	builder := processRiskTypeQueryCondition(query, subBuilder)
	if query.Sort != "" {
		sortCol := query.Sort // 来自 validatePaginationQueryParameters，已是 DB 列名
		dir := query.Direction
		if dir == "" {
			dir = interfaces.DESC_DIRECTION
		}
		builder = builder.OrderBy(fmt.Sprintf("%s %s", sortCol, dir))
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := rta.db.Query(sqlStr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var riskTypes []*interfaces.RiskType
	for rows.Next() {
		rt := &interfaces.RiskType{ModuleType: interfaces.MODULE_TYPE_RISK_TYPE}
		var tagsStr string

		err := rows.Scan(
			&rt.RTID,
			&rt.RTName,
			&rt.Comment,
			&tagsStr,
			&rt.Icon,
			&rt.Color,
			&rt.BKNRawContent,
			&rt.KNID,
			&rt.Branch,
			&rt.Creator.ID,
			&rt.Creator.Type,
			&rt.CreateTime,
			&rt.Updater.ID,
			&rt.Updater.Type,
			&rt.UpdateTime,
		)
		if err != nil {
			return nil, err
		}

		rt.Tags = libCommon.TagString2TagSlice(tagsStr)

		riskTypes = append(riskTypes, rt)
	}

	span.SetStatus(codes.Ok, "")
	return riskTypes, nil
}

func (rta *riskTypeAccess) GetRiskTypesTotal(ctx context.Context, query interfaces.RiskTypesQueryParams) (int, error) {
	_, span := ar_trace.Tracer.Start(ctx, "GetRiskTypesTotal", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	subBuilder := sq.Select("COUNT(f_id)").From(RT_TABLE_NAME)
	builder := processRiskTypeQueryCondition(query, subBuilder)
	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var total int
	err = rta.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		return 0, err
	}
	span.SetStatus(codes.Ok, "")
	return total, nil
}

func (rta *riskTypeAccess) GetRiskTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*interfaces.RiskType, error) {
	_, span := ar_trace.Tracer.Start(ctx, "GetRiskTypesByIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if len(rtIDs) == 0 {
		return []*interfaces.RiskType{}, nil
	}

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_name",
		"f_comment",
		"f_tags",
		"f_icon",
		"f_color",
		"f_bkn_raw_content",
		"f_kn_id",
		"f_branch",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": rtIDs}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := rta.db.Query(sqlStr, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var riskTypes []*interfaces.RiskType
	for rows.Next() {
		rt := &interfaces.RiskType{ModuleType: interfaces.MODULE_TYPE_RISK_TYPE}
		var tagsStr string

		err := rows.Scan(
			&rt.RTID,
			&rt.RTName,
			&rt.Comment,
			&tagsStr,
			&rt.Icon,
			&rt.Color,
			&rt.BKNRawContent,
			&rt.KNID,
			&rt.Branch,
			&rt.Creator.ID,
			&rt.Creator.Type,
			&rt.CreateTime,
			&rt.Updater.ID,
			&rt.Updater.Type,
			&rt.UpdateTime,
		)
		if err != nil {
			return nil, err
		}

		rt.Tags = libCommon.TagString2TagSlice(tagsStr)

		riskTypes = append(riskTypes, rt)
	}

	span.SetStatus(codes.Ok, "")
	return riskTypes, nil
}

func (rta *riskTypeAccess) UpdateRiskType(ctx context.Context, tx *sql.Tx, riskType *interfaces.RiskType) error {
	_, span := ar_trace.Tracer.Start(ctx, "UpdateRiskType", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	tagsStr := libCommon.TagSlice2TagString(riskType.Tags)

	data := map[string]any{
		"f_name":            riskType.RTName,
		"f_comment":         riskType.Comment,
		"f_tags":            tagsStr,
		"f_icon":            riskType.Icon,
		"f_color":           riskType.Color,
		"f_bkn_raw_content": riskType.BKNRawContent,
		"f_updater":         riskType.Updater.ID,
		"f_updater_type":    riskType.Updater.Type,
		"f_update_time":     riskType.UpdateTime,
	}

	sqlStr, vals, err := sq.Update(RT_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": riskType.RTID}).
		Where(sq.Eq{"f_kn_id": riskType.KNID}).
		Where(sq.Eq{"f_branch": riskType.Branch}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(sqlStr, vals...)
	return err
}

func (rta *riskTypeAccess) DeleteRiskTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) (int64, error) {
	_, span := ar_trace.Tracer.Start(ctx, "DeleteRiskTypesByIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if len(rtIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": rtIDs}).
		ToSql()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (rta *riskTypeAccess) GetAllRiskTypesByKnID(ctx context.Context, knID string, branch string) ([]*interfaces.RiskType, error) {
	return rta.ListRiskTypes(ctx, interfaces.RiskTypesQueryParams{
		KNID:   knID,
		Branch: branch,
	})
}

func (rta *riskTypeAccess) DeleteRiskTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	_, span := ar_trace.Tracer.Start(ctx, "DeleteRiskTypesByKnID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	sqlStr, vals, err := sq.Delete(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func processRiskTypeQueryCondition(query interfaces.RiskTypesQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
	if query.NamePattern != "" {
		subBuilder = subBuilder.Where(sq.Expr("(instr(f_name, ?) > 0 OR instr(f_id, ?) > 0)", query.NamePattern, query.NamePattern))
	}
	if query.Tag != "" {
		subBuilder = subBuilder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+query.Tag+`"`))
	}
	if query.KNID != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_kn_id": query.KNID})
	}
	if query.Branch != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_branch": query.Branch})
	} else {
		subBuilder = subBuilder.Where(sq.Eq{"f_branch": interfaces.MAIN_BRANCH})
	}
	return subBuilder
}
