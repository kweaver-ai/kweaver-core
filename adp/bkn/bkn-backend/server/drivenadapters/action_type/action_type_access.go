// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_type

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
	AT_TABLE_NAME = "t_action_type"
)

var (
	atAccessOnce sync.Once
	atAccess     interfaces.ActionTypeAccess
)

type actionTypeAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewActionTypeAccess(appSetting *common.AppSetting) interfaces.ActionTypeAccess {
	atAccessOnce.Do(func() {
		atAccess = &actionTypeAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return atAccess
}

// 根据ID获取行动类存在性
func (ata *actionTypeAccess) CheckActionTypeExistByID(ctx context.Context, knID string, branch string, atID string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CheckActionTypeExistByID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_name").
		From(AT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": atID}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of get action type id by f_id, error", err)
		return "", false, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("获取行动类信息的 sql 语句: %s", sqlStr))

	var name string
	err = ata.db.QueryRow(sqlStr, vals...).Scan(&name)
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

// 根据名称获取行动类存在性
func (ata *actionTypeAccess) CheckActionTypeExistByName(ctx context.Context, knID string, branch string, atName string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CheckActionTypeExistByName")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_id").
		From(AT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_name": atName}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of get id by name, error", err)
		return "", false, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("获取行动类信息的 sql 语句: %s", sqlStr))

	var atID string
	err = ata.db.QueryRow(sqlStr, vals...).Scan(
		&atID,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return "", false, nil
	} else if err != nil {
		otellog.LogError(ctx, "Row scan failed, err", err)
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return atID, true, nil
}

// 创建行动类
func (ata *actionTypeAccess) CreateActionType(ctx context.Context, tx *sql.Tx, actionType *interfaces.ActionType) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "CreateActionType")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(actionType.Tags)

	// 2.0 序列化 condition
	conditionBytes, err := sonic.Marshal(actionType.Condition)
	if err != nil {
		logger.Errorf("Failed to marshal Condition, err: %v", err.Error())
		return err
	}

	// 2.1 序列化 affect
	affectBytes, err := sonic.Marshal(actionType.Affect)
	if err != nil {
		logger.Errorf("Failed to marshal Affect, err: %v", err.Error())
		return err
	}

	// 2.2 序列化 action_source
	actionSourceBytes, err := sonic.Marshal(actionType.ActionSource)
	if err != nil {
		logger.Errorf("Failed to marshal ActionSource, err: %v", err.Error())
		return err
	}

	// 2.3 序列化 parameters
	parameterBytes, err := sonic.Marshal(actionType.Parameters)
	if err != nil {
		logger.Errorf("Failed to marshal Parameters, err: %v", err.Error())
		return err
	}

	// 2.4 序列化 schedule
	scheduleBytes, err := sonic.Marshal(actionType.Schedule)
	if err != nil {
		logger.Errorf("Failed to marshal Schedule, err: %v", err.Error())
		return err
	}

	sqlStr, vals, err := sq.Insert(AT_TABLE_NAME).
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
			"f_action_type",
			"f_object_type_id",
			"f_condition",
			"f_affect",
			"f_action_source",
			"f_parameters",
			"f_schedule",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			actionType.ATID,
			actionType.ATName,
			tagsStr,
			actionType.Comment,
			actionType.Icon,
			actionType.Color,
			actionType.BKNRawContent,
			actionType.KNID,
			actionType.Branch,
			actionType.ActionType,
			actionType.ObjectTypeID,
			conditionBytes,
			affectBytes,
			actionSourceBytes,
			parameterBytes,
			scheduleBytes,
			actionType.Creator.ID,
			actionType.Creator.Type,
			actionType.CreateTime,
			actionType.Updater.ID,
			actionType.Updater.Type,
			actionType.UpdateTime).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of insert action type, error", err)
		return err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("创建行动类的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Insert data error", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询行动类列表。查主线的当前版本为true的行动类
func (ata *actionTypeAccess) ListActionTypes(ctx context.Context, query interfaces.ActionTypesQueryParams) ([]*interfaces.ActionType, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "ListActionTypes")
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
		"f_action_type",
		"f_object_type_id",
		"f_condition",
		"f_affect",
		"f_action_source",
		"f_parameters",
		"f_schedule",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time").
		From(AT_TABLE_NAME)

	builder := processQueryCondition(query, subBuilder)

	//排序
	if query.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", query.Sort, query.Direction))
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select action types, error", err)
		return []*interfaces.ActionType{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询行动类列表的 sql 语句: %s; queryParams: %v", sqlStr, query))

	rows, err := ata.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return []*interfaces.ActionType{}, err
	}
	defer func() { _ = rows.Close() }()

	actionTypes := make([]*interfaces.ActionType, 0)
	for rows.Next() {
		actionType := interfaces.ActionType{
			ModuleType: interfaces.MODULE_TYPE_ACTION_TYPE,
		}
		tagsStr := ""
		var (
			conditionBytes    []byte
			affectBytes       []byte
			actionSourceBytes []byte
			parametersBytes   []byte
			scheduleBytes     []byte
		)
		err := rows.Scan(
			&actionType.ATID,
			&actionType.ATName,
			&tagsStr,
			&actionType.Comment,
			&actionType.Icon,
			&actionType.Color,
			&actionType.BKNRawContent,
			&actionType.KNID,
			&actionType.Branch,
			&actionType.ActionType,
			&actionType.ObjectTypeID,
			&conditionBytes,
			&affectBytes,
			&actionSourceBytes,
			&parametersBytes,
			&scheduleBytes,
			&actionType.Creator.ID,
			&actionType.Creator.Type,
			&actionType.CreateTime,
			&actionType.Updater.ID,
			&actionType.Updater.Type,
			&actionType.UpdateTime,
		)
		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return []*interfaces.ActionType{}, err
		}

		// tags string 转成数组的格式
		actionType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化 condition
		err = sonic.Unmarshal(conditionBytes, &actionType.Condition)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Condition after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.1 反序列化 affect
		err = sonic.Unmarshal(affectBytes, &actionType.Affect)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Affect after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.2 反序列化  action_source
		err = sonic.Unmarshal(actionSourceBytes, &actionType.ActionSource)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal ActionSource after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.3 反序列化  parameters
		err = sonic.Unmarshal(parametersBytes, &actionType.Parameters)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Parameters after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.4 反序列化 schedule
		err = sonic.Unmarshal(scheduleBytes, &actionType.Schedule)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Schedule after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}

		actionTypes = append(actionTypes, &actionType)
	}

	span.SetStatus(codes.Ok, "")
	return actionTypes, nil
}

func (ata *actionTypeAccess) GetActionTypesTotal(ctx context.Context, query interfaces.ActionTypesQueryParams) (int, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetActionTypesTotal")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	subBuilder := sq.Select("COUNT(f_id)").From(AT_TABLE_NAME)
	builder := processQueryCondition(query, subBuilder)
	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select action types total, error", err)
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询行动类总数的 sql 语句: %s; queryParams: %v", sqlStr, query))

	total := 0
	err = ata.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		otellog.LogError(ctx, "Get action type total error", err)
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

func (ata *actionTypeAccess) GetActionTypesByIDs(ctx context.Context, knID string, branch string, atIDs []string) ([]*interfaces.ActionType, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetActionTypesByIDs")
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
		"f_action_type",
		"f_object_type_id",
		"f_condition",
		"f_affect",
		"f_action_source",
		"f_parameters",
		"f_schedule",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(AT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": atIDs}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select action type by id, error", err)
		return []*interfaces.ActionType{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询行动类列表的 sql 语句: %s.", sqlStr))

	rows, err := ata.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return []*interfaces.ActionType{}, err
	}
	defer func() { _ = rows.Close() }()

	actionTypes := make([]*interfaces.ActionType, 0)
	for rows.Next() {
		actionType := interfaces.ActionType{
			ModuleType: interfaces.MODULE_TYPE_ACTION_TYPE,
		}
		tagsStr := ""
		var (
			conditionBytes    []byte
			affectBytes       []byte
			actionSourceBytes []byte
			parametersBytes   []byte
			scheduleBytes     []byte
		)

		err := rows.Scan(
			&actionType.ATID,
			&actionType.ATName,
			&tagsStr,
			&actionType.Comment,
			&actionType.Icon,
			&actionType.Color,
			&actionType.BKNRawContent,
			&actionType.KNID,
			&actionType.Branch,
			&actionType.ActionType,
			&actionType.ObjectTypeID,
			&conditionBytes,
			&affectBytes,
			&actionSourceBytes,
			&parametersBytes,
			&scheduleBytes,
			&actionType.Creator.ID,
			&actionType.Creator.Type,
			&actionType.CreateTime,
			&actionType.Updater.ID,
			&actionType.Updater.Type,
			&actionType.UpdateTime,
		)

		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return []*interfaces.ActionType{}, err
		}

		// tags string 转成数组的格式
		actionType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化 condition
		err = sonic.Unmarshal(conditionBytes, &actionType.Condition)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Condition after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.1 反序列化 affect
		err = sonic.Unmarshal(affectBytes, &actionType.Affect)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Affect after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.2 反序列化  action_source
		err = sonic.Unmarshal(actionSourceBytes, &actionType.ActionSource)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal ActionSource after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.3 反序列化  parameters
		err = sonic.Unmarshal(parametersBytes, &actionType.Parameters)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Parameters after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}
		// 2.4 反序列化 schedule
		err = sonic.Unmarshal(scheduleBytes, &actionType.Schedule)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Schedule after getting action type, err", err)
			return []*interfaces.ActionType{}, err
		}

		actionTypes = append(actionTypes, &actionType)
	}

	span.SetStatus(codes.Ok, "")
	return actionTypes, nil
}

func (ata *actionTypeAccess) UpdateActionType(ctx context.Context, tx *sql.Tx, actionType *interfaces.ActionType) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "UpdateActionType")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(actionType.Tags)
	// 2.0 序列化 condition
	conditionBytes, err := sonic.Marshal(actionType.Condition)
	if err != nil {
		otellog.LogError(ctx, "Failed to marshal Condition, err", err)
		return err
	}

	// 2.1 序列化 affect
	affectBytes, err := sonic.Marshal(actionType.Affect)
	if err != nil {
		otellog.LogError(ctx, "Failed to marshal Affect, err", err)
		return err
	}

	// 2.2 序列化 action_source
	actionSourceBytes, err := sonic.Marshal(actionType.ActionSource)
	if err != nil {
		otellog.LogError(ctx, "Failed to marshal ActionSource, err", err)
		return err
	}

	// 2.3 序列化 parameters
	parameterBytes, err := sonic.Marshal(actionType.Parameters)
	if err != nil {
		otellog.LogError(ctx, "Failed to marshal Parameters, err", err)
		return err
	}

	// 2.4 序列化 schedule
	scheduleBytes, err := sonic.Marshal(actionType.Schedule)
	if err != nil {
		otellog.LogError(ctx, "Failed to marshal Schedule, err", err)
		return err
	}

	data := map[string]any{
		"f_name":            actionType.ATName,
		"f_tags":            tagsStr,
		"f_comment":         actionType.Comment,
		"f_icon":            actionType.Icon,
		"f_color":           actionType.Color,
		"f_bkn_raw_content": actionType.BKNRawContent,
		"f_action_type":     actionType.ActionType,
		"f_object_type_id":  actionType.ObjectTypeID,
		"f_condition":       conditionBytes,
		"f_affect":          affectBytes,
		"f_action_source":   actionSourceBytes,
		"f_parameters":      parameterBytes,
		"f_schedule":        scheduleBytes,
		"f_updater":         actionType.Updater.ID,
		"f_updater_type":    actionType.Updater.Type,
		"f_update_time":     actionType.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(AT_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": actionType.ATID}).
		Where(sq.Eq{"f_kn_id": actionType.KNID}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of update action type by action type id, error", err)
		return err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("修改行动类的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "update action type error", err)
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
		otellog.LogWarn(ctx, fmt.Sprintf("Update %s RowsAffected not equal 1, RowsAffected is %d, ActionType is %v",
			actionType.ATID, RowsAffected, actionType))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ata *actionTypeAccess) DeleteActionTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, atIDs []string) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteActionTypesByIDs")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(atIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(AT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": atIDs}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of delete action type by action type id, error", err)
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除行动类的 sql 语句: %s; 删除的行动类ids: %v", sqlStr, atIDs))

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

func (ata *actionTypeAccess) DeleteActionTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "DeleteActionTypesByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Delete(AT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of delete action type by kn_id, error", err)
		return 0, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("删除行动类的 sql 语句: %s; 删除的行动类kn_id: %s, branch: %s", sqlStr, knID, branch))

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

func (ata *actionTypeAccess) GetActionTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetActionTypeIDsByKnID")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_id",
	).From(AT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select action type ids by kn_id, error", err)
		return nil, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询行动类的 sql 语句: %s.", sqlStr))

	rows, err := ata.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	atIDs := []string{}
	for rows.Next() {

		var atID string
		err := rows.Scan(
			&atID,
		)
		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return nil, err
		}

		atIDs = append(atIDs, atID)
	}

	span.SetStatus(codes.Ok, "")
	return atIDs, nil
}

// 拼接 sql 过滤条件
func processQueryCondition(query interfaces.ActionTypesQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
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

	if query.ActionType != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_action_type": query.ActionType})
	}

	if len(query.ObjectTypeIDs) > 0 {
		subBuilder = subBuilder.Where(sq.Eq{"f_object_type_id": query.ObjectTypeIDs})
	}

	return subBuilder
}

// 查询行动类列表。查主线的当前版本为true的行动类
func (ata *actionTypeAccess) GetAllActionTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*interfaces.ActionType, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetAllActionTypesByKnID")
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
		"f_action_type",
		"f_object_type_id",
		"f_condition",
		"f_affect",
		"f_action_source",
		"f_parameters",
		"f_schedule",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time").
		From(AT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()

	if err != nil {
		otellog.LogError(ctx, "Failed to build the sql of select action types, error", err)
		return map[string]*interfaces.ActionType{}, err
	}

	// 记录处理的 sql 字符串
	otellog.LogInfo(ctx, fmt.Sprintf("查询行动类列表的 sql 语句: %s.", sqlStr))

	rows, err := ata.db.Query(sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "List data error", err)
		return map[string]*interfaces.ActionType{}, err
	}
	defer func() { _ = rows.Close() }()

	actionTypes := make(map[string]*interfaces.ActionType)
	for rows.Next() {
		actionType := interfaces.ActionType{
			ModuleType: interfaces.MODULE_TYPE_ACTION_TYPE,
		}
		tagsStr := ""
		var (
			conditionBytes    []byte
			affectBytes       []byte
			actionSourceBytes []byte
			parametersBytes   []byte
			scheduleBytes     []byte
		)
		err := rows.Scan(
			&actionType.ATID,
			&actionType.ATName,
			&tagsStr,
			&actionType.Comment,
			&actionType.Icon,
			&actionType.Color,
			&actionType.BKNRawContent,
			&actionType.KNID,
			&actionType.Branch,
			&actionType.ActionType,
			&actionType.ObjectTypeID,
			&conditionBytes,
			&affectBytes,
			&actionSourceBytes,
			&parametersBytes,
			&scheduleBytes,
			&actionType.Creator.ID,
			&actionType.Creator.Type,
			&actionType.CreateTime,
			&actionType.Updater.ID,
			&actionType.Updater.Type,
			&actionType.UpdateTime,
		)
		if err != nil {
			otellog.LogError(ctx, "Row scan error", err)
			return map[string]*interfaces.ActionType{}, err
		}

		// tags string 转成数组的格式
		actionType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化 condition
		err = sonic.Unmarshal(conditionBytes, &actionType.Condition)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Condition after getting action type, err", err)
			return map[string]*interfaces.ActionType{}, err
		}
		// 2.1 反序列化 affect
		err = sonic.Unmarshal(affectBytes, &actionType.Affect)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Affect after getting action type, err", err)
			return map[string]*interfaces.ActionType{}, err
		}
		// 2.2 反序列化  action_source
		err = sonic.Unmarshal(actionSourceBytes, &actionType.ActionSource)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal ActionSource after getting action type, err", err)
			return map[string]*interfaces.ActionType{}, err
		}
		// 2.3 反序列化  parameters
		err = sonic.Unmarshal(parametersBytes, &actionType.Parameters)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Parameters after getting action type, err", err)
			return map[string]*interfaces.ActionType{}, err
		}
		// 2.4 反序列化 schedule
		err = sonic.Unmarshal(scheduleBytes, &actionType.Schedule)
		if err != nil {
			otellog.LogError(ctx, "Failed to unmarshal Schedule after getting action type, err", err)
			return map[string]*interfaces.ActionType{}, err
		}

		actionTypes[actionType.ATID] = &actionType
	}

	span.SetStatus(codes.Ok, "")
	return actionTypes, nil
}
