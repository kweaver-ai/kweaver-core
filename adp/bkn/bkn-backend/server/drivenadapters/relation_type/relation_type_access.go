// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package relation_type

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"

	"bkn-backend/interfaces"
)

const (
	RT_TABLE_NAME = "t_relation_type"
)

var (
	rtAccessOnce sync.Once
	rtAccess     interfaces.RelationTypeAccess
)

type relationTypeAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewRelationTypeAccess(appSetting *common.AppSetting) interfaces.RelationTypeAccess {
	rtAccessOnce.Do(func() {
		rtAccess = &relationTypeAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return rtAccess
}

// 根据ID获取关系类存在性
func (rta *relationTypeAccess) CheckRelationTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CheckRelationTypeExistByID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_name").
		From(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": rtID}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of get relation type id by f_id, error", err)
		return "", false, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("获取关系类信息的 sql 语句: %s", sqlStr))

	var name string
	err = rta.db.QueryRow(sqlStr, vals...).Scan(&name)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return "", false, nil
	} else if err != nil {
		otellog.LogError(ctx, "Row scan failed, err", err)
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return name, true, nil
}

// 创建关系类
func (rta *relationTypeAccess) CreateRelationType(ctx context.Context, tx *sql.Tx, relationType *interfaces.RelationType) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CreateRelationType")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(relationType.Tags)

	// 2.0 序列化数据来源
	mappingRulesBytes, err := sonic.Marshal(relationType.MappingRules)
	if err != nil {
		logger.Errorf("Failed to marshal MappingRules, err: %v", err.Error())
		return err
	}

	sqlStr, vals, err := sq.Insert(RT_TABLE_NAME).
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
			"f_source_object_type_id",
			"f_target_object_type_id",
			"f_type",
			"f_mapping_rules",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			relationType.RTID,
			relationType.RTName,
			tagsStr,
			relationType.Comment,
			relationType.Icon,
			relationType.Color,
			relationType.BKNRawContent,
			relationType.KNID,
			relationType.Branch,
			relationType.SourceObjectTypeID,
			relationType.TargetObjectTypeID,
			relationType.Type,
			mappingRulesBytes,
			relationType.Creator.ID,
			relationType.Creator.Type,
			relationType.CreateTime,
			relationType.Updater.ID,
			relationType.Updater.Type,
			relationType.UpdateTime).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of insert relation type, error", err)
		return err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("创建关系类的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Insert data error", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询关系类列表。查主线的当前版本为true的关系类
func (rta *relationTypeAccess) ListRelationTypes(ctx context.Context, query interfaces.RelationTypesQueryParams) ([]*interfaces.RelationType, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "ListRelationTypes")
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
		"f_source_object_type_id",
		"f_target_object_type_id",
		"f_type",
		"f_mapping_rules",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time").
		From(RT_TABLE_NAME)

	builder := processQueryCondition(query, subBuilder)

	//排序
	if query.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", query.Sort, query.Direction))
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select relation types, error", err)
		return []*interfaces.RelationType{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询关系类列表的 sql 语句: %s; queryParams: %v", sqlStr, query))

	rows, err := rta.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return []*interfaces.RelationType{}, err
	}
	defer func() { _ = rows.Close() }()

	relationTypes := make([]*interfaces.RelationType, 0)
	for rows.Next() {
		relationType := interfaces.RelationType{
			ModuleType: interfaces.MODULE_TYPE_RELATION_TYPE,
		}
		tagsStr := ""
		var mappingRulesBytes []byte
		err := rows.Scan(
			&relationType.RTID,
			&relationType.RTName,
			&tagsStr,
			&relationType.Comment,
			&relationType.Icon,
			&relationType.Color,
			&relationType.BKNRawContent,
			&relationType.KNID,
			&relationType.Branch,
			&relationType.SourceObjectTypeID,
			&relationType.TargetObjectTypeID,
			&relationType.Type,
			&mappingRulesBytes,
			&relationType.Creator.ID,
			&relationType.Creator.Type,
			&relationType.CreateTime,
			&relationType.Updater.ID,
			&relationType.Updater.Type,
			&relationType.UpdateTime,
		)
		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return []*interfaces.RelationType{}, err
		}

		// tags string 转成数组的格式
		relationType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化dMappingRules
		if relationType.Type == interfaces.RELATION_TYPE_DIRECT {
			var mappings []interfaces.Mapping
			err = sonic.Unmarshal(mappingRulesBytes, &mappings)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return []*interfaces.RelationType{}, err
			}
			relationType.MappingRules = mappings
		}
		if relationType.Type == interfaces.RELATION_TYPE_DATA_VIEW {
			var mappings interfaces.InDirectMapping
			err = sonic.Unmarshal(mappingRulesBytes, &mappings)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return []*interfaces.RelationType{}, err
			}
			relationType.MappingRules = &mappings
		}
		if relationType.Type == interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN {
			var fcj interfaces.FilteredCrossJoinMapping
			err = sonic.Unmarshal(mappingRulesBytes, &fcj)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return []*interfaces.RelationType{}, err
			}
			relationType.MappingRules = &fcj
		}

		relationTypes = append(relationTypes, &relationType)
	}

	span.SetStatus(codes.Ok, "")
	return relationTypes, nil
}

func (rta *relationTypeAccess) GetRelationTypesTotal(ctx context.Context, query interfaces.RelationTypesQueryParams) (int, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetRelationTypesTotal")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	subBuilder := sq.Select("COUNT(f_id)").From(RT_TABLE_NAME)
	builder := processQueryCondition(query, subBuilder)

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select relation types total, error", err)
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询关系类总数的 sql 语句: %s; queryParams: %v", sqlStr, query))

	total := 0
	err = rta.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		otellog.LogError(ctx, "Get relation type total error", err)
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

func (rta *relationTypeAccess) GetRelationTypeByID(ctx context.Context, knID string, branch string, rtID string) (*interfaces.RelationType, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetRelationTypeByID")
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
		"f_source_object_type_id",
		"f_target_object_type_id",
		"f_type",
		"f_mapping_rules",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": rtID}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select relation type by id, error", err)
		return nil, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询关系类列表的 sql 语句: %s.", sqlStr))

	relationType := interfaces.RelationType{
		ModuleType: interfaces.MODULE_TYPE_RELATION_TYPE,
	}
	tagsStr := ""
	var mappingRulesBytes []byte

	row := rta.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&relationType.RTID,
		&relationType.RTName,
		&tagsStr,
		&relationType.Comment,
		&relationType.Icon,
		&relationType.Color,
		&relationType.BKNRawContent,
		&relationType.KNID,
		&relationType.Branch,
		&relationType.SourceObjectTypeID,
		&relationType.TargetObjectTypeID,
		&relationType.Type,
		&mappingRulesBytes,
		&relationType.Creator.ID,
		&relationType.Creator.Type,
		&relationType.CreateTime,
		&relationType.Updater.ID,
		&relationType.Updater.Type,
		&relationType.UpdateTime,
	)
	if err != nil {
		otellog.LogError(ctx, "Row scan error", err)
		return nil, err
	}

	// tags string 转成数组的格式
	relationType.Tags = libCommon.TagString2TagSlice(tagsStr)

	// 2.0 反序列化dMappingRules
	if relationType.Type == interfaces.RELATION_TYPE_DIRECT {
		var mappings []interfaces.Mapping
		err = sonic.Unmarshal(mappingRulesBytes, &mappings)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
			return nil, err
		}
		relationType.MappingRules = mappings
	}
	if relationType.Type == interfaces.RELATION_TYPE_DATA_VIEW {
		var mappings interfaces.InDirectMapping
		err = sonic.Unmarshal(mappingRulesBytes, &mappings)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
			return nil, err
		}
		relationType.MappingRules = &mappings
	}
	if relationType.Type == interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN {
		var fcj interfaces.FilteredCrossJoinMapping
		err = sonic.Unmarshal(mappingRulesBytes, &fcj)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
			return nil, err
		}
		relationType.MappingRules = &fcj
	}

	span.SetStatus(codes.Ok, "")
	return &relationType, nil
}

func (rta *relationTypeAccess) GetRelationTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*interfaces.RelationType, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetRelationTypesByIDs")
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
		"f_source_object_type_id",
		"f_target_object_type_id",
		"f_type",
		"f_mapping_rules",
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
		otellog.LogError(ctx, "Failed to build the sql of select relation type by id, error", err)
		return []*interfaces.RelationType{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询关系类列表的 sql 语句: %s.", sqlStr))

	rows, err := rta.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return []*interfaces.RelationType{}, err
	}
	defer func() { _ = rows.Close() }()

	relationTypes := make([]*interfaces.RelationType, 0)
	for rows.Next() {
		relationType := interfaces.RelationType{
			ModuleType: interfaces.MODULE_TYPE_RELATION_TYPE,
		}
		tagsStr := ""
		var mappingRulesBytes []byte

		err := rows.Scan(
			&relationType.RTID,
			&relationType.RTName,
			&tagsStr,
			&relationType.Comment,
			&relationType.Icon,
			&relationType.Color,
			&relationType.BKNRawContent,
			&relationType.KNID,
			&relationType.Branch,
			&relationType.SourceObjectTypeID,
			&relationType.TargetObjectTypeID,
			&relationType.Type,
			&mappingRulesBytes,
			&relationType.Creator.ID,
			&relationType.Creator.Type,
			&relationType.CreateTime,
			&relationType.Updater.ID,
			&relationType.Updater.Type,
			&relationType.UpdateTime,
		)

		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return []*interfaces.RelationType{}, err
		}

		// tags string 转成数组的格式
		relationType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化dMappingRules
		if relationType.Type == interfaces.RELATION_TYPE_DIRECT {
			var mappings []interfaces.Mapping
			err = sonic.Unmarshal(mappingRulesBytes, &mappings)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return []*interfaces.RelationType{}, err
			}
			relationType.MappingRules = mappings
		}
		if relationType.Type == interfaces.RELATION_TYPE_DATA_VIEW {
			var mappings interfaces.InDirectMapping
			err = sonic.Unmarshal(mappingRulesBytes, &mappings)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return []*interfaces.RelationType{}, err
			}
			relationType.MappingRules = &mappings
		}
		if relationType.Type == interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN {
			var fcj interfaces.FilteredCrossJoinMapping
			err = sonic.Unmarshal(mappingRulesBytes, &fcj)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return []*interfaces.RelationType{}, err
			}
			relationType.MappingRules = &fcj
		}

		relationTypes = append(relationTypes, &relationType)
	}

	span.SetStatus(codes.Ok, "")
	return relationTypes, nil
}

func (rta *relationTypeAccess) UpdateRelationType(ctx context.Context, tx *sql.Tx, relationType *interfaces.RelationType) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "UpdateRelationType")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(relationType.Tags)
	// 2.0 序列化数据来源
	mappingRulesBytes, err := sonic.Marshal(relationType.MappingRules)
	if err != nil {
		logger.Errorf("Failed to marshal MappingRules, err: %v", err.Error())
		return err
	}

	data := map[string]any{
		"f_name":                  relationType.RTName,
		"f_tags":                  tagsStr,
		"f_comment":               relationType.Comment,
		"f_icon":                  relationType.Icon,
		"f_color":                 relationType.Color,
		"f_bkn_raw_content":       relationType.BKNRawContent,
		"f_source_object_type_id": relationType.SourceObjectTypeID,
		"f_target_object_type_id": relationType.TargetObjectTypeID,
		"f_type":                  relationType.Type,
		"f_mapping_rules":         mappingRulesBytes,
		"f_updater":               relationType.Updater.ID,
		"f_updater_type":          relationType.Updater.Type,
		"f_update_time":           relationType.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(RT_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": relationType.RTID}).
		Where(sq.Eq{"f_kn_id": relationType.KNID}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of update relation type by relation type id, error", err)
		return err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("修改关系类的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "update relation type error", err)
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		otellog.LogError(ctx, "Get RowsAffected error", err)
		return err
	}

	if RowsAffected != 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		otellog.LogWarn(ctx, fmt.Sprintf("Update %s RowsAffected not equal 1, RowsAffected is %d, RelationType is %v",
			relationType.RTID, RowsAffected, relationType))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (rta *relationTypeAccess) DeleteRelationTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteRelationTypesByIDs")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(rtIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": rtIDs}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of delete relation type by relation type id, error", err)
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除关系类的 sql 语句: %s; 删除的关系类ids: %v", sqlStr, rtIDs))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Delete data error", err)
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		otellog.LogError(ctx, "Get RowsAffected error", err)
		return 0, err
	}

	if RowsAffected != int64(len(rtIDs)) {
		otellog.LogWarn(ctx, fmt.Sprintf("Delete %d RowsAffected not equal %d, rtIDs is %v",
			len(rtIDs), RowsAffected, rtIDs))
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (rta *relationTypeAccess) DeleteRelationTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteRelationTypesByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Delete(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of delete relation type by relation type id, error", err)
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除关系类的 sql 语句: %s; 删除的关系类kn_id: %s, branch: %s", sqlStr, knID, branch))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Delete data error", err)
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		otellog.LogError(ctx, "Get RowsAffected error", err)
		return 0, err
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (rta *relationTypeAccess) GetRelationTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetRelationTypeIDsByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Select(
		"f_id",
	).From(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select relation type by id, error", err)
		return nil, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询关系类列表的 sql 语句: %s.", sqlStr))

	rows, err := rta.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	rtIDs := []string{}
	for rows.Next() {

		var rtID string

		err := rows.Scan(
			&rtID,
		)

		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return nil, err
		}

		rtIDs = append(rtIDs, rtID)
	}

	span.SetStatus(codes.Ok, "")
	return rtIDs, nil
}

// 拼接 sql 过滤条件
func processQueryCondition(query interfaces.RelationTypesQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
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

	if len(query.SourceObjectTypeIDs) > 0 {
		subBuilder = subBuilder.Where(sq.Eq{"f_source_object_type_id": query.SourceObjectTypeIDs})
	}

	if len(query.TargetObjectTypeIDs) > 0 {
		subBuilder = subBuilder.Where(sq.Eq{"f_target_object_type_id": query.TargetObjectTypeIDs})
	}

	if len(query.BoundObjectTypeIDs) > 0 {
		subBuilder = subBuilder.Where(sq.Or{
			sq.Eq{"f_source_object_type_id": query.BoundObjectTypeIDs},
			sq.Eq{"f_target_object_type_id": query.BoundObjectTypeIDs},
		})
	}

	return subBuilder
}

func (rta *relationTypeAccess) GetAllRelationTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*interfaces.RelationType, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetAllRelationTypesByKnID")
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
		"f_source_object_type_id",
		"f_target_object_type_id",
		"f_type",
		"f_mapping_rules",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time").
		From(RT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()

	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select relation types, error", err)
		return map[string]*interfaces.RelationType{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询关系类列表的 sql 语句: %s; knID: %s", sqlStr, knID))

	rows, err := rta.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return map[string]*interfaces.RelationType{}, err
	}
	defer func() { _ = rows.Close() }()

	relationTypes := make(map[string]*interfaces.RelationType)
	for rows.Next() {
		relationType := interfaces.RelationType{
			ModuleType: interfaces.MODULE_TYPE_RELATION_TYPE,
		}
		tagsStr := ""
		var mappingRulesBytes []byte
		err := rows.Scan(
			&relationType.RTID,
			&relationType.RTName,
			&tagsStr,
			&relationType.Comment,
			&relationType.Icon,
			&relationType.Color,
			&relationType.BKNRawContent,
			&relationType.KNID,
			&relationType.Branch,
			&relationType.SourceObjectTypeID,
			&relationType.TargetObjectTypeID,
			&relationType.Type,
			&mappingRulesBytes,
			&relationType.Creator.ID,
			&relationType.Creator.Type,
			&relationType.CreateTime,
			&relationType.Updater.ID,
			&relationType.Updater.Type,
			&relationType.UpdateTime,
		)
		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return map[string]*interfaces.RelationType{}, err
		}

		// tags string 转成数组的格式
		relationType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化dMappingRules
		if relationType.Type == interfaces.RELATION_TYPE_DIRECT {
			var mappings []interfaces.Mapping
			err = sonic.Unmarshal(mappingRulesBytes, &mappings)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return map[string]*interfaces.RelationType{}, err
			}
			relationType.MappingRules = mappings
		}
		if relationType.Type == interfaces.RELATION_TYPE_DATA_VIEW {
			var mappings interfaces.InDirectMapping
			err = sonic.Unmarshal(mappingRulesBytes, &mappings)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return map[string]*interfaces.RelationType{}, err
			}
			relationType.MappingRules = &mappings
		}
		if relationType.Type == interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN {
			var fcj interfaces.FilteredCrossJoinMapping
			err = sonic.Unmarshal(mappingRulesBytes, &fcj)
			if err != nil {
				otellog.LogError(ctx, "Failed to unmarshal mappingRules after getting relation type, err", err)
				return map[string]*interfaces.RelationType{}, err
			}
			relationType.MappingRules = &fcj
		}

		relationTypes[relationType.RTID] = &relationType
	}

	span.SetStatus(codes.Ok, "")
	return relationTypes, nil
}
