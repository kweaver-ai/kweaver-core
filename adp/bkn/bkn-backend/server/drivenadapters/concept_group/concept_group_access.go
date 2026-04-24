// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package concept_group

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	"bkn-backend/drivenadapters/action_type"
	"bkn-backend/drivenadapters/object_type"
	"bkn-backend/drivenadapters/relation_type"
	"bkn-backend/infra/otel/otellog"
	"bkn-backend/infra/otel/oteltrace"
	"bkn-backend/interfaces"
)

const (
	CONCEPT_GROUP_TABLE_NAME          = "t_concept_group"
	CONCEPT_GROUP_RELATION_TABLE_NAME = "t_concept_group_relation"
)

var (
	cgAccessOnce sync.Once
	cgAccess     interfaces.ConceptGroupAccess
)

type conceptGroupAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewConceptGroupAccess(appSetting *common.AppSetting) interfaces.ConceptGroupAccess {
	cgAccessOnce.Do(func() {
		cgAccess = &conceptGroupAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return cgAccess
}

// 根据ID获取概念分组存在性
func (cga *conceptGroupAccess) CheckConceptGroupExistByID(ctx context.Context, knID string, branch string, cgID string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CheckConceptGroupExistByID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_name").
		From(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_id": cgID}).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get concept group id by f_id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of get concept group id by f_id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return "", false, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("获取概念分组信息的 sql 语句: %s", sqlStr))

	var name string
	err = cga.db.QueryRow(sqlStr, vals...).Scan(&name)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return "", false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)
		otellog.LogError(ctx, "Row scan failed, err", err)
		span.SetStatus(codes.Error, "Row scan failed ")
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return name, true, nil
}

// 根据名称获取概念分组存在性
func (cga *conceptGroupAccess) CheckConceptGroupExistByName(ctx context.Context, knID string, branch string, name string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CheckConceptGroupExistByName")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	//查询
	sqlStr, vals, err := sq.Select(
		"f_id").
		From(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_name": name}).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get id by name, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of get id by name, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return "", false, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("获取概念分组信息的 sql 语句: %s", sqlStr))

	var cgID string
	err = cga.db.QueryRow(sqlStr, vals...).Scan(
		&cgID,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return "", false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)
		otellog.LogError(ctx, "Row scan failed, err", err)
		span.SetStatus(codes.Error, "Row scan failed ")
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return cgID, true, nil
}

// 创建概念分组
func (cga *conceptGroupAccess) CreateConceptGroup(ctx context.Context, tx *sql.Tx, conceptGroup *interfaces.ConceptGroup) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CreateConceptGroup")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(conceptGroup.Tags)

	sqlStr, vals, err := sq.Insert(CONCEPT_GROUP_TABLE_NAME).
		Columns(
			"f_id",
			"f_name",
			"f_tags",
			"f_comment",
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
			conceptGroup.CGID,
			conceptGroup.CGName,
			tagsStr,
			conceptGroup.Comment,
			conceptGroup.Icon,
			conceptGroup.Color,
			conceptGroup.BKNRawContent,
			conceptGroup.KNID,
			conceptGroup.Branch,
			conceptGroup.Creator.ID,
			conceptGroup.Creator.Type,
			conceptGroup.CreateTime,
			conceptGroup.Updater.ID,
			conceptGroup.Updater.Type,
			conceptGroup.UpdateTime).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert concept group, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of insert concept group, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("创建概念分组的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		otellog.LogError(ctx, "Insert data error", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询概念分组列表。查主线的当前版本为true的概念分组
func (cga *conceptGroupAccess) ListConceptGroups(ctx context.Context, query interfaces.ConceptGroupsQueryParams) ([]*interfaces.ConceptGroup, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "ListConceptGroups")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	subBuilder := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_comment",
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
		"f_update_time").
		From(CONCEPT_GROUP_TABLE_NAME)

	builder := processQueryCondition(query, subBuilder)

	//排序
	if query.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", query.Sort, query.Direction))
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept groups, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept groups, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return []*interfaces.ConceptGroup{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组列表的 sql 语句: %s; queryParams: %v", sqlStr, query))

	rows, err := cga.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return []*interfaces.ConceptGroup{}, err
	}
	defer rows.Close()

	conceptGroups := make([]*interfaces.ConceptGroup, 0)

	for rows.Next() {
		conceptGroup := interfaces.ConceptGroup{
			ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP,
		}
		tagsStr := ""
		err := rows.Scan(
			&conceptGroup.CGID,
			&conceptGroup.CGName,
			&tagsStr,
			&conceptGroup.Comment,
			&conceptGroup.Icon,
			&conceptGroup.Color,
			&conceptGroup.BKNRawContent,
			&conceptGroup.KNID,
			&conceptGroup.Branch,
			&conceptGroup.Creator.ID,
			&conceptGroup.Creator.Type,
			&conceptGroup.CreateTime,
			&conceptGroup.Updater.ID,
			&conceptGroup.Updater.Type,
			&conceptGroup.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return []*interfaces.ConceptGroup{}, err
		}

		// tags string 转成数组的格式
		conceptGroup.Tags = libCommon.TagString2TagSlice(tagsStr)

		conceptGroups = append(conceptGroups, &conceptGroup)
	}

	span.SetStatus(codes.Ok, "")
	return conceptGroups, nil
}

// 批量获取概念分组
func (cga *conceptGroupAccess) GetConceptGroupsByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, cgIDs []string) ([]*interfaces.ConceptGroup, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetConceptGroupsByIDs")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_comment",
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
	).From(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_id": cgIDs}).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept group by id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept group by id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return []*interfaces.ConceptGroup{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("批量查询概念分组信息的 sql 语句: %s.", sqlStr))
	rows, err := tx.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return []*interfaces.ConceptGroup{}, err
	}
	defer rows.Close()

	conceptGroups := make([]*interfaces.ConceptGroup, 0)
	for rows.Next() {
		conceptGroup := interfaces.ConceptGroup{
			ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP,
		}
		tagsStr := ""
		err := rows.Scan(
			&conceptGroup.CGID,
			&conceptGroup.CGName,
			&tagsStr,
			&conceptGroup.Comment,
			&conceptGroup.Icon,
			&conceptGroup.Color,
			&conceptGroup.BKNRawContent,
			&conceptGroup.KNID,
			&conceptGroup.Branch,
			&conceptGroup.Creator.ID,
			&conceptGroup.Creator.Type,
			&conceptGroup.CreateTime,
			&conceptGroup.Updater.ID,
			&conceptGroup.Updater.Type,
			&conceptGroup.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return []*interfaces.ConceptGroup{}, err
		}

		// tags string 转成数组的格式
		conceptGroup.Tags = libCommon.TagString2TagSlice(tagsStr)

		conceptGroups = append(conceptGroups, &conceptGroup)
	}

	span.SetStatus(codes.Ok, "")
	return conceptGroups, nil
}

// 获取概念分组总数
func (cga *conceptGroupAccess) GetConceptGroupsTotal(ctx context.Context, query interfaces.ConceptGroupsQueryParams) (int, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetConceptGroupsTotal")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	subBuilder := sq.Select("COUNT(f_id)").From(CONCEPT_GROUP_TABLE_NAME)
	builder := processQueryCondition(query, subBuilder)
	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept groups total, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept groups total, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组总数的 sql 语句: %s; queryParams: %v", sqlStr, query))

	total := 0
	err = cga.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		logger.Errorf("get concept group total error: %v\n", err)
		otellog.LogError(ctx, "Get concept group total error", err)
		span.SetStatus(codes.Error, "Get concept group total error")
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

func (cga *conceptGroupAccess) GetConceptGroupByID(ctx context.Context, knID string, branch string, cgID string) (*interfaces.ConceptGroup, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetConceptGroupByID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_comment",
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
	).From(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_id": cgID}).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept group by id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept group by id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return &interfaces.ConceptGroup{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组信息的 sql 语句: %s.", sqlStr))
	tagsStr := ""
	conceptGroup := &interfaces.ConceptGroup{
		ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP,
	}
	err = cga.db.QueryRow(sqlStr, vals...).Scan(
		&conceptGroup.CGID,
		&conceptGroup.CGName,
		&tagsStr,
		&conceptGroup.Comment,
		&conceptGroup.Icon,
		&conceptGroup.Color,
		&conceptGroup.BKNRawContent,
		&conceptGroup.KNID,
		&conceptGroup.Branch,
		&conceptGroup.Creator.ID,
		&conceptGroup.Creator.Type,
		&conceptGroup.CreateTime,
		&conceptGroup.Updater.ID,
		&conceptGroup.Updater.Type,
		&conceptGroup.UpdateTime,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return nil, nil
	} else if err != nil {
		logger.Errorf("Get concept group by id error: %v\n", err)
		span.SetStatus(codes.Error, "Get concept group by id error")
		otellog.LogError(ctx, "Get concept group by id error", err)
		return nil, err
	}

	// tags string 转成数组的格式
	conceptGroup.Tags = libCommon.TagString2TagSlice(tagsStr)

	span.SetStatus(codes.Ok, "")
	return conceptGroup, nil
}

func (cga *conceptGroupAccess) UpdateConceptGroup(ctx context.Context, tx *sql.Tx, conceptGroup *interfaces.ConceptGroup) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "UpdateConceptGroup")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(conceptGroup.Tags)

	data := map[string]any{
		"f_name":            conceptGroup.CGName,
		"f_tags":            tagsStr,
		"f_comment":         conceptGroup.Comment,
		"f_icon":            conceptGroup.Icon,
		"f_color":           conceptGroup.Color,
		"f_bkn_raw_content": conceptGroup.BKNRawContent,
		"f_updater":         conceptGroup.Updater.ID,
		"f_updater_type":    conceptGroup.Updater.Type,
		"f_update_time":     conceptGroup.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(CONCEPT_GROUP_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": conceptGroup.CGID}).
		Where(sq.Eq{"f_kn_id": conceptGroup.KNID}).
		Where(sq.Eq{"f_branch": conceptGroup.Branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update concept group by concept group id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of update concept group by concept group id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("修改概念分组的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update concept group error: %v\n", err)
		otellog.LogError(ctx, "Update data error", err)
		span.SetStatus(codes.Error, "Update data error")
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		otellog.LogWarn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
		return err
	}

	if RowsAffected != 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected not equal 1, RowsAffected is %d, ActionType is %v",
			conceptGroup.CGID, RowsAffected, conceptGroup)
		otellog.LogWarn(ctx, fmt.Sprintf("Update %s RowsAffected not equal 1, RowsAffected is %d, ActionType is %v",
			conceptGroup.CGID, RowsAffected, conceptGroup))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (cga *conceptGroupAccess) UpdateConceptGroupDetail(ctx context.Context, knID string, branch string, cgID string, detail string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "UpdateConceptGroupDetail")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	data := map[string]any{
		"f_bkn_raw_content": detail,
	}
	sqlStr, vals, err := sq.Update(CONCEPT_GROUP_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": cgID}).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update concept group detail by concept group id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of update concept group detail by concept group id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("修改概念分组详情的 sql 语句: %s", sqlStr))

	ret, err := cga.db.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update concept groupk detail error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		otellog.LogError(ctx, "Update data error", err)
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		otellog.LogWarn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	if RowsAffected != 1 {
		logger.Errorf("UPDATE concept group detail %d RowsAffected not equal 1, RowsAffected is %d, KNID is %s",
			knID, RowsAffected, knID)
		otellog.LogWarn(ctx, fmt.Sprintf("Update concept group detail %s RowsAffected not equal 1, RowsAffected is %d",
			knID, RowsAffected))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (cga *conceptGroupAccess) DeleteConceptGroupByID(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteConceptGroupByID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if cgID == "" {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_id": cgID}).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete concept group by concept group id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of delete concept group by concept group id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除概念分组的 sql 语句: %s; 删除的概念分组id: %s in kn_id [%s] branch [%s]", sqlStr, cgID, knID, branch))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		otellog.LogError(ctx, "Delete data error", err)
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		otellog.LogWarn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (cga *conceptGroupAccess) DeleteConceptGroupsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteConceptGroupByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Delete(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete concept group by concept group id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of delete concept group by concept group id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除概念分组的 sql 语句: %s; 删除的概念分组: kn_id [%s] branch [%s]", sqlStr, knID, branch))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		otellog.LogError(ctx, "Delete data error", err)
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		otellog.LogWarn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (cga *conceptGroupAccess) DeleteConceptGroupRelationsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteConceptGroupRelationsByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Delete(CONCEPT_GROUP_RELATION_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete concept group relation by kn_id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of delete concept group relation by kn_id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除概念分组关联的 sql 语句: %s; 删除的概念分组关联: kn_id [%s] branch [%s]", sqlStr, knID, branch))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		otellog.LogError(ctx, "Delete data error", err)
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		otellog.LogWarn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (cga *conceptGroupAccess) GetConceptGroupIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetConceptGroupIDsByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_id",
	).From(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept group ids by kn_id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept group ids by kn_id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组的 sql 语句: %s.", sqlStr))
	rows, err := cga.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return nil, err
	}
	defer rows.Close()

	cgIDs := []string{}
	for rows.Next() {

		var atID string
		err := rows.Scan(
			&atID,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return nil, err
		}

		cgIDs = append(cgIDs, atID)
	}

	span.SetStatus(codes.Ok, "")
	return cgIDs, nil
}

// 拼接 sql 过滤条件
func processQueryCondition(query interfaces.ConceptGroupsQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
	if query.NamePattern != "" {
		// 模糊查询，名称或id进行模糊查询，匹配任一即可
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
		// 查主线分支的业务知识网络
		subBuilder = subBuilder.Where(sq.Eq{"f_branch": interfaces.MAIN_BRANCH})
	}

	if len(query.CGIDs) > 0 {
		subBuilder = subBuilder.Where(sq.Eq{"f_id": query.CGIDs})
	}

	return subBuilder
}

// 查询概念分组列表。查主线的当前版本为true的概念分组
func (cga *conceptGroupAccess) GetAllConceptGroupsByKnID(ctx context.Context, knID string, branch string) (map[string]*interfaces.ConceptGroup, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetAllConceptGroupsByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_comment",
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
		"f_update_time").
		From(CONCEPT_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()

	if err != nil {
		logger.Errorf("Failed to build the sql of select concept groups, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept groups, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return map[string]*interfaces.ConceptGroup{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组列表的 sql 语句: %s.", sqlStr))

	rows, err := cga.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return map[string]*interfaces.ConceptGroup{}, err
	}
	defer rows.Close()

	conceptGroups := make(map[string]*interfaces.ConceptGroup)
	for rows.Next() {
		conceptGroup := interfaces.ConceptGroup{
			ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP,
		}
		tagsStr := ""

		err := rows.Scan(
			&conceptGroup.CGID,
			&conceptGroup.CGName,
			&tagsStr,
			&conceptGroup.Comment,
			&conceptGroup.Icon,
			&conceptGroup.Color,
			&conceptGroup.BKNRawContent,
			&conceptGroup.KNID,
			&conceptGroup.Branch,
			&conceptGroup.Creator.ID,
			&conceptGroup.Creator.Type,
			&conceptGroup.CreateTime,
			&conceptGroup.Updater.ID,
			&conceptGroup.Updater.Type,
			&conceptGroup.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return map[string]*interfaces.ConceptGroup{}, err
		}

		// tags string 转成数组的格式
		conceptGroup.Tags = libCommon.TagString2TagSlice(tagsStr)

		conceptGroups[conceptGroup.CGID] = &conceptGroup
	}

	span.SetStatus(codes.Ok, "")
	return conceptGroups, nil
}

// 获取指定分组下的对象类的绑定关系
func (cga *conceptGroupAccess) ListConceptGroupRelations(ctx context.Context, tx *sql.Tx,
	query interfaces.ConceptGroupRelationsQueryParams) ([]interfaces.ConceptGroupRelation, error) {

	ctx, span := oteltrace.StartNamedClientSpan(ctx, "ListConceptGroupRelations")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	subBuilder := sq.Select(
		"f_id",
		"f_kn_id",
		"f_branch",
		"f_group_id",
		"f_concept_type",
		"f_concept_id",
		"f_create_time",
	).From(CONCEPT_GROUP_RELATION_TABLE_NAME)

	builder := processConceptGroupRelationsQueryCondition(query, subBuilder, "")

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept group by id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept group by id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return []interfaces.ConceptGroupRelation{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组信息的 sql 语句: %s.", sqlStr))

	rows, err := tx.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return []interfaces.ConceptGroupRelation{}, err
	}
	defer rows.Close()

	conceptGroupRelations := make([]interfaces.ConceptGroupRelation, 0)
	for rows.Next() {
		conceptGroupRelation := interfaces.ConceptGroupRelation{
			ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP_RELATION,
		}

		err := rows.Scan(
			&conceptGroupRelation.ID,
			&conceptGroupRelation.KNID,
			&conceptGroupRelation.Branch,
			&conceptGroupRelation.CGID,
			&conceptGroupRelation.ConceptType,
			&conceptGroupRelation.ConceptID,
			&conceptGroupRelation.CreateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return []interfaces.ConceptGroupRelation{}, err
		}

		conceptGroupRelations = append(conceptGroupRelations, conceptGroupRelation)
	}

	return conceptGroupRelations, nil
}

func (cga *conceptGroupAccess) CreateConceptGroupRelation(ctx context.Context, tx *sql.Tx, conceptGroupRelation *interfaces.ConceptGroupRelation) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CreateConceptGroupRelation")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Insert(CONCEPT_GROUP_RELATION_TABLE_NAME).
		Columns(
			"f_id",
			"f_kn_id",
			"f_branch",
			"f_group_id",
			"f_concept_type",
			"f_concept_id",
			"f_create_time",
		).
		Values(
			conceptGroupRelation.ID,
			conceptGroupRelation.KNID,
			conceptGroupRelation.Branch,
			conceptGroupRelation.CGID,
			conceptGroupRelation.ConceptType,
			conceptGroupRelation.ConceptID,
			conceptGroupRelation.CreateTime).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert concept group relation, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of insert concept group relation, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("创建概念与分组关系的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		otellog.LogError(ctx, "Insert data error", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 拼接 sql 过滤条件
func processConceptGroupRelationsQueryCondition(query interfaces.ConceptGroupRelationsQueryParams, subBuilder sq.SelectBuilder, fieldPrefix string) sq.SelectBuilder {

	if query.KNID != "" {
		subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_kn_id"): query.KNID})
	}

	if query.Branch != "" {
		subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_branch"): query.Branch})
	} else {
		// 查主线分支的业务知识网络
		subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_branch"): interfaces.MAIN_BRANCH})
	}

	if len(query.CGIDs) > 0 {
		subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_group_id"): query.CGIDs})
	}

	if query.ConceptType != "" {
		subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_concept_type"): query.ConceptType})
	}

	if len(query.OTIDs) != 0 {
		subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_concept_id"): query.OTIDs})
	}

	return subBuilder
}

// 从分组中删除对象类，即删除概念与分组的绑定关系
func (cga *conceptGroupAccess) DeleteObjectTypesFromGroup(ctx context.Context, tx *sql.Tx, query interfaces.ConceptGroupRelationsQueryParams) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteObjectTypesFromGroup")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Delete(CONCEPT_GROUP_RELATION_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": query.KNID}).
		Where(sq.Eq{"f_branch": query.Branch}).
		Where(sq.Eq{"f_concept_type": query.ConceptType})

	if len(query.CGIDs) > 0 {
		builder = builder.Where(sq.Eq{"f_group_id": query.CGIDs})
	}

	if len(query.OTIDs) > 0 {
		builder = builder.Where(sq.Eq{"f_concept_id": query.OTIDs})
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete concept group by concept group id, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of delete concept group by concept group id, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除概念与分组关系的 sql 语句: %s; 删除的概念分组id: %s in kn_id [%s] branch [%s] concept_ids [%v]",
		sqlStr, query.CGIDs, query.KNID, query.Branch, query.OTIDs))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		otellog.LogError(ctx, "Delete data error", err)
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		otellog.LogWarn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (cga *conceptGroupAccess) GetConceptIDsByConceptGroupIDs(ctx context.Context, knID string, branch string, cgIDs []string, conceptType string) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetConceptIDsByConceptGroupIDs")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_concept_id",
	).From(CONCEPT_GROUP_RELATION_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_concept_type": conceptType}).
		Where(sq.Eq{"f_group_id": cgIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept ids by concept group, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept ids by concept group, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return []string{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组下的概念ID的 sql 语句: %s. 分组ids: %s", sqlStr, cgIDs))

	rows, err := cga.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return []string{}, err
	}
	defer rows.Close()

	conceptIDs := make([]string, 0)
	for rows.Next() {
		var conceptID string
		err := rows.Scan(
			&conceptID,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return []string{}, err
		}

		conceptIDs = append(conceptIDs, conceptID)
	}

	span.SetStatus(codes.Ok, "")
	return conceptIDs, nil
}

// 获取概念分组下的关系类ID
func (cga *conceptGroupAccess) GetRelationTypeIDsFromConceptGroupRelation(ctx context.Context, query interfaces.ConceptGroupRelationsQueryParams) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetRelationTypeIDsFromConceptGroupRelation")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// 子查询：获取指定概念组中的概念ID（object_type类型）
	subQueryBuilder := sq.Select("cgr.f_concept_id").
		From(CONCEPT_GROUP_RELATION_TABLE_NAME + " AS cgr").
		Join(object_type.OT_TABLE_NAME + " AS ot ON cgr.f_concept_id = ot.f_id AND cgr.f_branch = ot.f_branch AND cgr.f_kn_id = ot.f_kn_id").
		Join(CONCEPT_GROUP_TABLE_NAME + " AS cg on cgr.f_group_id = cg.f_id and cgr.f_branch = cg.f_branch and cgr.f_kn_id = cg.f_kn_id")

	subQueryBuilder = processConceptGroupRelationsQueryCondition(query, subQueryBuilder, "cgr.")

	// 主查询
	builder := sq.Select(
		"f_id",
	).From(relation_type.RT_TABLE_NAME).
		Where(sq.Expr("f_source_object_type_id IN (?)", subQueryBuilder)).
		Where(sq.Expr("f_target_object_type_id IN (?)", subQueryBuilder))

	if query.KNID != "" {
		builder = builder.Where(sq.Eq{"f_kn_id": query.KNID})
	}

	if query.Branch != "" {
		builder = builder.Where(sq.Eq{"f_branch": query.Branch})
	} else {
		// 查主线分支的业务知识网络
		builder = builder.Where(sq.Eq{"f_branch": interfaces.MAIN_BRANCH})
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select relation type ids by concept group, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select relation type ids by concept group, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return []string{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组下的关系类ID的 sql 语句: %s.", sqlStr))

	rows, err := cga.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return []string{}, err
	}
	defer rows.Close()

	rtIDs := make([]string, 0)
	for rows.Next() {
		var rtID string

		err := rows.Scan(
			&rtID,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return []string{}, err
		}

		rtIDs = append(rtIDs, rtID)
	}

	span.SetStatus(codes.Ok, "")
	return rtIDs, nil
}

// 获取概念分组下的行动类ID
func (cga *conceptGroupAccess) GetActionTypeIDsFromConceptGroupRelation(ctx context.Context, query interfaces.ConceptGroupRelationsQueryParams) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetActionTypeIDsFromConceptGroupRelation")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// 子查询：获取指定概念组中的概念ID（object_type类型）
	subQueryBuilder := sq.Select("cgr.f_concept_id").
		From(CONCEPT_GROUP_RELATION_TABLE_NAME + " AS cgr").
		Join(object_type.OT_TABLE_NAME + " AS ot ON cgr.f_concept_id = ot.f_id AND cgr.f_branch = ot.f_branch AND cgr.f_kn_id = ot.f_kn_id").
		Join(CONCEPT_GROUP_TABLE_NAME + " AS cg on cgr.f_group_id = cg.f_id and cgr.f_branch = cg.f_branch and cgr.f_kn_id = cg.f_kn_id")

	subQueryBuilder = processConceptGroupRelationsQueryCondition(query, subQueryBuilder, "cgr.")

	// 主查询
	builder := sq.Select(
		"f_id",
	).From(action_type.AT_TABLE_NAME).
		Where(sq.Expr("f_object_type_id IN (?)", subQueryBuilder))

	if query.KNID != "" {
		builder = builder.Where(sq.Eq{"f_kn_id": query.KNID})
	}

	if query.Branch != "" {
		builder = builder.Where(sq.Eq{"f_branch": query.Branch})
	} else {
		// 查主线分支的业务知识网络
		builder = builder.Where(sq.Eq{"f_branch": interfaces.MAIN_BRANCH})
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select action type ids by concept group, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select action type ids by concept group, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return []string{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询概念分组下的行动类ID的 sql 语句: %s.", sqlStr))

	rows, err := cga.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return []string{}, err
	}
	defer rows.Close()

	atIDs := make([]string, 0)
	for rows.Next() {
		var atID string

		err := rows.Scan(
			&atID,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return []string{}, err
		}

		atIDs = append(atIDs, atID)
	}

	span.SetStatus(codes.Ok, "")
	return atIDs, nil
}

// 获取概念所属的分组信息
func (cga *conceptGroupAccess) GetConceptGroupsByOTIDs(ctx context.Context, tx *sql.Tx,
	query interfaces.ConceptGroupRelationsQueryParams) (map[string][]*interfaces.ConceptGroup, error) {

	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetConceptGroupsByOTIDs")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	subBuilder := sq.Select(
		"cgr.f_concept_id",
		"cg.f_id",
		"cg.f_name",
		"cg.f_tags",
		"cg.f_comment",
		"cg.f_icon",
		"cg.f_color",
		// "cg.f_bkn_raw_content",
		"cg.f_kn_id",
		"cg.f_branch",
	).From(CONCEPT_GROUP_TABLE_NAME + " AS cg").
		Join(CONCEPT_GROUP_RELATION_TABLE_NAME + " AS cgr on cgr.f_group_id = cg.f_id and cgr.f_kn_id  = cg.f_kn_id and cgr.f_branch =cg.f_branch")

	builder := processConceptGroupRelationsQueryCondition(query, subBuilder, "cgr.")
	// Where(sq.Eq{"cgr.f_kn_id": knID}).
	// Where(sq.Eq{"cgr.f_branch": branch}).
	// Where(sq.Eq{"cgr.f_concept_type": interfaces.MODULE_TYPE_OBJECT_TYPE}).
	// Where(sq.Eq{"cgr.f_concept_id": otIDs})

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select concept group by object type ids, error: %s", err.Error())
		otellog.LogError(ctx, "Failed to build the sql of select concept group by object type ids, error", err)
		span.SetStatus(codes.Error, "Build sql failed ")
		return map[string][]*interfaces.ConceptGroup{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询对象类ID所属的概念分组的 sql 语句: %s.", sqlStr))

	rows, err := tx.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		otellog.LogError(ctx, "List data error", err)
		span.SetStatus(codes.Error, "List data error")
		return map[string][]*interfaces.ConceptGroup{}, err
	}
	defer rows.Close()

	results := map[string][]*interfaces.ConceptGroup{}
	for rows.Next() {
		var otID string
		conceptGroup := &interfaces.ConceptGroup{
			ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP,
		}
		tagsStr := ""
		err := rows.Scan(
			&otID,
			&conceptGroup.CGID,
			&conceptGroup.CGName,
			&tagsStr,
			&conceptGroup.Comment,
			&conceptGroup.Icon,
			&conceptGroup.Color,
			// &conceptGroup.Detail,
			&conceptGroup.KNID,
			&conceptGroup.Branch,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			otellog.LogError(ctx, "Row scan error", err)
			span.SetStatus(codes.Error, "Row scan error")
			return map[string][]*interfaces.ConceptGroup{}, err
		}

		// tags string 转成数组的格式
		conceptGroup.Tags = libCommon.TagString2TagSlice(tagsStr)

		results[otID] = append(results[otID], conceptGroup)
	}

	span.SetStatus(codes.Ok, "")
	return results, nil
}
