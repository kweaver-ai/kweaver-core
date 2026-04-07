// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package object_type

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

const (
	OT_TABLE_NAME        = "t_object_type"
	OT_STATUS_TABLE_NAME = "t_object_type_status"
)

var (
	otAccessOnce sync.Once
	otAccess     interfaces.ObjectTypeAccess
)

type objectTypeAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewObjectTypeAccess(appSetting *common.AppSetting) interfaces.ObjectTypeAccess {
	otAccessOnce.Do(func() {
		otAccess = &objectTypeAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return otAccess
}

// 根据ID获取对象类存在性
func (ota *objectTypeAccess) CheckObjectTypeExistByID(ctx context.Context, knID string, branch string, otID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CheckObjectTypeExistByID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_name").
		From(OT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": otID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get object type id by f_id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get object type id by f_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return "", false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取对象类信息的 sql 语句: %s", sqlStr))

	var name string
	err = ota.db.QueryRow(sqlStr, vals...).Scan(&name)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return "", false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
		span.SetStatus(codes.Error, "Row scan failed ")
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return name, true, nil
}

// 根据名称获取对象类存在性
func (ota *objectTypeAccess) CheckObjectTypeExistByName(ctx context.Context, knID string, branch string, name string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CheckObjectTypeExistByName", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select("f_id").
		From(OT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_name": name}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get id by name, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get id by name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return "", false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取对象类信息的 sql 语句: %s", sqlStr))

	var otID string
	err = ota.db.QueryRow(sqlStr, vals...).Scan(
		&otID,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return "", false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
		span.SetStatus(codes.Error, "Row scan failed ")
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return otID, true, nil
}

// 创建对象类
func (ota *objectTypeAccess) CreateObjectType(ctx context.Context, tx *sql.Tx, objectType *interfaces.ObjectType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateObjectType", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(objectType.Tags)

	// 2.0 序列化数据来源
	dataSourceBytes, err := sonic.Marshal(objectType.DataSource)
	if err != nil {
		logger.Errorf("Failed to marshal DataSource, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to marshal DataSource, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Marshal DataSource failed ")
		return err
	}
	// 2.1 序列化数据属性
	dataPropertiesBytes, err := sonic.Marshal(objectType.DataProperties)
	if err != nil {
		logger.Errorf("Failed to marshal DataProperties, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to marshal DataProperties, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Marshal DataProperties failed ")
		return err
	}
	// 2.2 序列化逻辑属性
	logicPropertiesBytes, err := sonic.Marshal(objectType.LogicProperties)
	if err != nil {
		logger.Errorf("Failed to marshal LogicProperties, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to marshal LogicProperties, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Marshal LogicProperties failed ")
		return err
	}
	// 2.3 序列化主键数组
	primaryKeysBytes, err := sonic.Marshal(objectType.PrimaryKeys)
	if err != nil {
		logger.Errorf("Failed to marshal PrimaryKeys, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to marshal PrimaryKeys, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Marshal PrimaryKeys failed ")
		return err
	}

	sqlStr, vals, err := sq.Insert(OT_TABLE_NAME).
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
			"f_data_source",
			"f_data_properties",
			"f_logic_properties",
			"f_primary_keys",
			"f_display_key",
			"f_incremental_key",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			objectType.OTID,
			objectType.OTName,
			tagsStr,
			objectType.Comment,
			objectType.Icon,
			objectType.Color,
			objectType.BKNRawContent,
			objectType.KNID,
			objectType.Branch,
			dataSourceBytes,
			dataPropertiesBytes,
			logicPropertiesBytes,
			primaryKeysBytes,
			objectType.DisplayKey,
			objectType.IncrementalKey,
			objectType.Creator.ID,
			objectType.Creator.Type,
			objectType.CreateTime,
			objectType.Updater.ID,
			objectType.Updater.Type,
			objectType.UpdateTime).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert object type, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert object type, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建对象类的 sql 语句: %s", sqlStr))

	if tx != nil {
		_, err = tx.Exec(sqlStr, vals...)
	} else {
		_, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		span.SetStatus(codes.Error, "Insert data error")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 创建对象类状态
func (ota *objectTypeAccess) CreateObjectTypeStatus(ctx context.Context, tx *sql.Tx, objectType *interfaces.ObjectType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateObjectTypeStatus", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Insert(OT_STATUS_TABLE_NAME).
		Columns(
			"f_id",
			"f_kn_id",
			"f_branch",
			"f_incremental_key",
			"f_update_time",
		).
		Values(
			objectType.OTID,
			objectType.KNID,
			objectType.Branch,
			objectType.IncrementalKey,
			objectType.UpdateTime).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert object type status, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert object type status, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建对象类状态的 sql 语句: %s", sqlStr))

	if tx != nil {
		_, err = tx.Exec(sqlStr, vals...)
	} else {
		_, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		span.SetStatus(codes.Error, "Insert data error")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询对象类列表。查主线的当前版本为true的对象类
func (ota *objectTypeAccess) ListObjectTypes(ctx context.Context, tx *sql.Tx, query interfaces.ObjectTypesQueryParams) ([]*interfaces.ObjectType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListObjectTypes", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	subBuilder := sq.Select(
		"ot.f_id",
		"ot.f_name",
		"ot.f_tags",
		"ot.f_comment",
		"ot.f_icon",
		"ot.f_color",
		"ot.f_bkn_raw_content",
		"ot.f_kn_id",
		"ot.f_branch",
		"ot.f_data_source",
		"ot.f_data_properties",
		"ot.f_logic_properties",
		"ot.f_primary_keys",
		"ot.f_display_key",
		"ot.f_incremental_key",
		"ot.f_creator",
		"ot.f_creator_type",
		"ot.f_create_time",
		"ot.f_updater",
		"ot.f_updater_type",
		"ot.f_update_time",

		"ots.f_incremental_key",
		"ots.f_incremental_value",
		"ots.f_index",
		"ots.f_index_available",
		"ots.f_doc_count",
		"ots.f_storage_size",
		"ots.f_update_time",
	).From(OT_TABLE_NAME + " AS ot").
		Join(OT_STATUS_TABLE_NAME + " AS ots ON ot.f_id = ots.f_id AND ot.f_kn_id = ots.f_kn_id AND ot.f_branch = ots.f_branch")

	builder := processQueryCondition(query, subBuilder)

	//排序
	if query.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("ot.%s %s", query.Sort, query.Direction))
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select object types, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select object types, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return []*interfaces.ObjectType{}, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询对象类列表的 sql 语句: %s; queryParams: %v", sqlStr, query))

	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.Query(sqlStr, vals...)
	} else {
		rows, err = ota.db.Query(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))
		span.SetStatus(codes.Error, "List data error")
		return []*interfaces.ObjectType{}, err
	}
	defer rows.Close()

	objectTypes := make([]*interfaces.ObjectType, 0)
	for rows.Next() {
		objectType := interfaces.ObjectType{
			ModuleType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			Status:     &interfaces.ObjectTypeStatus{},
		}
		tagsStr := ""
		var (
			dataSourceBytes      []byte
			dataPropertiesBytes  []byte
			logicPropertiesBytes []byte
			primaryKeysBytes     []byte
		)
		err := rows.Scan(
			&objectType.OTID,
			&objectType.OTName,
			&tagsStr,
			&objectType.Comment,
			&objectType.Icon,
			&objectType.Color,
			&objectType.BKNRawContent,
			&objectType.KNID,
			&objectType.Branch,
			&dataSourceBytes,
			&dataPropertiesBytes,
			&logicPropertiesBytes,
			&primaryKeysBytes,
			&objectType.DisplayKey,
			&objectType.IncrementalKey,
			&objectType.Creator.ID,
			&objectType.Creator.Type,
			&objectType.CreateTime,
			&objectType.Updater.ID,
			&objectType.Updater.Type,
			&objectType.UpdateTime,

			&objectType.Status.IncrementalKey,
			&objectType.Status.IncrementalValue,
			&objectType.Status.Index,
			&objectType.Status.IndexAvailable,
			&objectType.Status.DocCount,
			&objectType.Status.StorageSize,
			&objectType.Status.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			span.SetStatus(codes.Error, "Row scan error")
			return []*interfaces.ObjectType{}, err
		}

		// tags string 转成数组的格式
		objectType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化datasource
		err = sonic.Unmarshal(dataSourceBytes, &objectType.DataSource)
		if err != nil {
			logger.Errorf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal dataSource error")
			return []*interfaces.ObjectType{}, err
		}

		// 2.1 反序列化DataProperties
		err = sonic.Unmarshal(dataPropertiesBytes, &objectType.DataProperties)
		if err != nil {
			logger.Errorf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal dataProperties error")
			return []*interfaces.ObjectType{}, err
		}

		// 2.2 反序列化LogicProperties
		err = sonic.Unmarshal(logicPropertiesBytes, &objectType.LogicProperties)
		if err != nil {
			logger.Errorf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal logicProperties error")
			return []*interfaces.ObjectType{}, err
		}

		// 2.3 反序列化主键
		err = sonic.Unmarshal(primaryKeysBytes, &objectType.PrimaryKeys)
		if err != nil {
			logger.Errorf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal primaryKeys error")
			return []*interfaces.ObjectType{}, err
		}

		objectTypes = append(objectTypes, &objectType)
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, nil
}

func (ota *objectTypeAccess) GetObjectTypesTotal(ctx context.Context, query interfaces.ObjectTypesQueryParams) (int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetObjectTypesTotal", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	subBuilder := sq.Select("COUNT(ot.f_id)").
		From(OT_TABLE_NAME + " AS ot")
	builder := processQueryCondition(query, subBuilder)

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select object types total, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select object types total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询对象类总数的 sql 语句: %s; queryParams: %v", sqlStr, query))

	total := 0
	err = ota.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		logger.Errorf("get object type totals error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Get object type totals error: %v", err))
		span.SetStatus(codes.Error, "Get object type totals error")
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

func (ota *objectTypeAccess) GetObjectTypeByID(ctx context.Context, tx *sql.Tx, knID string, branch string, otID string) (*interfaces.ObjectType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetObjectTypeByID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"ot.f_id",
		"ot.f_name",
		"ot.f_tags",
		"ot.f_comment",
		"ot.f_icon",
		"ot.f_color",
		"ot.f_bkn_raw_content",
		"ot.f_kn_id",
		"ot.f_branch",
		"ot.f_data_source",
		"ot.f_data_properties",
		"ot.f_logic_properties",
		"ot.f_primary_keys",
		"ot.f_display_key",
		"ot.f_incremental_key",
		"ot.f_creator",
		"ot.f_creator_type",
		"ot.f_create_time",
		"ot.f_updater",
		"ot.f_updater_type",
		"ot.f_update_time",

		"ots.f_incremental_key",
		"ots.f_incremental_value",
		"ots.f_index",
		"ots.f_index_available",
		"ots.f_doc_count",
		"ots.f_storage_size",
		"ots.f_update_time",
	).From(OT_TABLE_NAME + " AS ot").
		Join(OT_STATUS_TABLE_NAME + " AS ots ON ot.f_id = ots.f_id AND ot.f_kn_id = ots.f_kn_id AND ot.f_branch = ots.f_branch").
		Where(sq.Eq{"ot.f_kn_id": knID}).
		Where(sq.Eq{"ot.f_branch": branch}).
		Where(sq.Eq{"ot.f_id": otID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select object type by id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select object type by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询对象类列表的 sql 语句: %s.", sqlStr))

	objectType := interfaces.ObjectType{
		ModuleType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		Status:     &interfaces.ObjectTypeStatus{},
	}
	tagsStr := ""
	var (
		dataSourceBytes      []byte
		dataPropertiesBytes  []byte
		logicPropertiesBytes []byte
		primaryKeysBytes     []byte
	)

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, sqlStr, vals...)
	} else {
		row = ota.db.QueryRowContext(ctx, sqlStr, vals...)
	}
	err = row.Scan(
		&objectType.OTID,
		&objectType.OTName,
		&tagsStr,
		&objectType.Comment,
		&objectType.Icon,
		&objectType.Color,
		&objectType.BKNRawContent,
		&objectType.KNID,
		&objectType.Branch,
		&dataSourceBytes,
		&dataPropertiesBytes,
		&logicPropertiesBytes,
		&primaryKeysBytes,
		&objectType.DisplayKey,
		&objectType.IncrementalKey,
		&objectType.Creator.ID,
		&objectType.Creator.Type,
		&objectType.CreateTime,
		&objectType.Updater.ID,
		&objectType.Updater.Type,
		&objectType.UpdateTime,

		&objectType.Status.IncrementalKey,
		&objectType.Status.IncrementalValue,
		&objectType.Status.Index,
		&objectType.Status.IndexAvailable,
		&objectType.Status.DocCount,
		&objectType.Status.StorageSize,
		&objectType.Status.UpdateTime,
	)
	if err == sql.ErrNoRows {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ObjectType_ObjectTypeNotFound).
			WithErrorDetails(fmt.Sprintf("对象类[%s]不存在: %v", otID, err))
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
		span.SetStatus(codes.Error, "Row scan error")
		return nil, err
	}

	// tags string 转成数组的格式
	objectType.Tags = libCommon.TagString2TagSlice(tagsStr)

	// 2.0 反序列化datasource
	err = sonic.Unmarshal(dataSourceBytes, &objectType.DataSource)
	if err != nil {
		logger.Errorf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Failed to unmarshal dataSource after getting object type")
		return nil, err
	}

	// 2.1 反序列化DataProperties
	err = sonic.Unmarshal(dataPropertiesBytes, &objectType.DataProperties)
	if err != nil {
		logger.Errorf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Failed to unmarshal dataProperties after getting object type")
		return nil, err
	}

	// 2.2 反序列化LogicProperties
	err = sonic.Unmarshal(logicPropertiesBytes, &objectType.LogicProperties)
	if err != nil {
		logger.Errorf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Failed to unmarshal logicProperties after getting object type")
		return nil, err
	}

	// 2.3 反序列化主键
	err = sonic.Unmarshal(primaryKeysBytes, &objectType.PrimaryKeys)
	if err != nil {
		logger.Errorf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error()))
		span.SetStatus(codes.Error, "Failed to unmarshal primaryKeys after getting object type")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return &objectType, nil
}

func (ota *objectTypeAccess) GetObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) ([]*interfaces.ObjectType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetObjectTypesByIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"ot.f_id",
		"ot.f_name",
		"ot.f_tags",
		"ot.f_comment",
		"ot.f_icon",
		"ot.f_color",
		"ot.f_bkn_raw_content",
		"ot.f_kn_id",
		"ot.f_branch",
		"ot.f_data_source",
		"ot.f_data_properties",
		"ot.f_logic_properties",
		"ot.f_primary_keys",
		"ot.f_display_key",
		"ot.f_incremental_key",
		"ot.f_creator",
		"ot.f_creator_type",
		"ot.f_create_time",
		"ot.f_updater",
		"ot.f_updater_type",
		"ot.f_update_time",

		"ots.f_incremental_key",
		"ots.f_incremental_value",
		"ots.f_index",
		"ots.f_index_available",
		"ots.f_doc_count",
		"ots.f_storage_size",
		"ots.f_update_time",
	).From(OT_TABLE_NAME + " AS ot").
		Join(OT_STATUS_TABLE_NAME + " AS ots ON ot.f_id = ots.f_id AND ot.f_kn_id = ots.f_kn_id AND ot.f_branch = ots.f_branch").
		Where(sq.Eq{"ot.f_kn_id": knID}).
		Where(sq.Eq{"ot.f_branch": branch}).
		Where(sq.Eq{"ot.f_id": otIDs}).
		ToSql()

		// if len(cgIds) > 0 {
		// 	// 子查询：获取指定概念组中的概念ID（object_type类型）
		// 	subQueryBuilder := sq.Select("cgr.f_concept_id").
		// 		From("t_concept_group_relation AS cgr").
		// 		Join(OT_TABLE_NAME + " AS ot ON cgr.f_concept_id = ot.f_id AND cgr.f_branch = ot.f_branch AND cgr.f_kn_id = ot.f_kn_id").
		// 		Join("t_concept_group AS cg ON cgr.f_group_id = cg.f_id AND cgr.f_branch = cg.f_branch AND cgr.f_kn_id = cg.f_kn_id").
		// 		Where(sq.Eq{"cgr.f_kn_id": knID}).
		// 		Where(sq.Eq{"cgr.f_branch": "main"}).
		// 		Where(sq.Eq{"cgr.f_group_id": cgIds}).
		// 		Where(sq.Eq{"cgr.f_concept_type": interfaces.MODULE_TYPE_OBJECT_TYPE})

		// 	builder = builder.Where(sq.Expr("f_id IN (?)", subQueryBuilder))
		// 	// if query.Branch != "" {
		// 	// 	subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_branch"): query.Branch})
		// 	// } else {
		// 	// 	// 查主线分支的业务知识网络
		// 	// 	subBuilder = subBuilder.Where(sq.Eq{fmt.Sprintf("%s%s", fieldPrefix, "f_branch"): interfaces.MAIN_BRANCH})
		// 	// }
		// }

	if err != nil {
		logger.Errorf("Failed to build the sql of select object type by id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select object type by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return []*interfaces.ObjectType{}, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询对象类列表的 sql 语句: %s.", sqlStr))

	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.Query(sqlStr, vals...)
	} else {
		rows, err = ota.db.Query(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))
		span.SetStatus(codes.Error, "List data error")
		return []*interfaces.ObjectType{}, err
	}
	defer rows.Close()

	objectTypes := make([]*interfaces.ObjectType, 0)
	for rows.Next() {
		objectType := interfaces.ObjectType{
			ModuleType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			Status:     &interfaces.ObjectTypeStatus{},
		}
		tagsStr := ""
		var (
			dataSourceBytes      []byte
			dataPropertiesBytes  []byte
			logicPropertiesBytes []byte
			primaryKeysBytes     []byte
		)

		err := rows.Scan(
			&objectType.OTID,
			&objectType.OTName,
			&tagsStr,
			&objectType.Comment,
			&objectType.Icon,
			&objectType.Color,
			&objectType.BKNRawContent,
			&objectType.KNID,
			&objectType.Branch,
			&dataSourceBytes,
			&dataPropertiesBytes,
			&logicPropertiesBytes,
			&primaryKeysBytes,
			&objectType.DisplayKey,
			&objectType.IncrementalKey,
			&objectType.Creator.ID,
			&objectType.Creator.Type,
			&objectType.CreateTime,
			&objectType.Updater.ID,
			&objectType.Updater.Type,
			&objectType.UpdateTime,

			&objectType.Status.IncrementalKey,
			&objectType.Status.IncrementalValue,
			&objectType.Status.Index,
			&objectType.Status.IndexAvailable,
			&objectType.Status.DocCount,
			&objectType.Status.StorageSize,
			&objectType.Status.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			span.SetStatus(codes.Error, "Row scan error")
			return []*interfaces.ObjectType{}, err
		}

		// tags string 转成数组的格式
		objectType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化datasource
		err = sonic.Unmarshal(dataSourceBytes, &objectType.DataSource)
		if err != nil {
			logger.Errorf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Failed to unmarshal dataSource after getting object type")
			return []*interfaces.ObjectType{}, err
		}

		// 2.1 反序列化DataProperties
		err = sonic.Unmarshal(dataPropertiesBytes, &objectType.DataProperties)
		if err != nil {
			logger.Errorf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Failed to unmarshal dataProperties after getting object type")
			return []*interfaces.ObjectType{}, err
		}

		// 2.2 反序列化LogicProperties
		err = sonic.Unmarshal(logicPropertiesBytes, &objectType.LogicProperties)
		if err != nil {
			logger.Errorf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Failed to unmarshal logicProperties after getting object type")
			return []*interfaces.ObjectType{}, err
		}

		// 2.3 反序列化主键
		err = sonic.Unmarshal(primaryKeysBytes, &objectType.PrimaryKeys)
		if err != nil {
			logger.Errorf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Failed to unmarshal primaryKeys after getting object type")
			return []*interfaces.ObjectType{}, err
		}

		objectTypes = append(objectTypes, &objectType)
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, nil
}

func (ota *objectTypeAccess) UpdateObjectType(ctx context.Context, tx *sql.Tx, objectType *interfaces.ObjectType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateObjectType", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(objectType.Tags)
	// 2.0 序列化数据来源
	dataSourceBytes, err := sonic.Marshal(objectType.DataSource)
	if err != nil {
		logger.Errorf("Failed to marshal DataSource, err: %v", err.Error())
		return err
	}
	// 2.1 序列化数据属性
	dataPropertiesBytes, err := sonic.Marshal(objectType.DataProperties)
	if err != nil {
		logger.Errorf("Failed to marshal DataProperties, err: %v", err.Error())
		return err
	}
	// 2.2 序列化逻辑属性
	logicPropertiesBytes, err := sonic.Marshal(objectType.LogicProperties)
	if err != nil {
		logger.Errorf("Failed to marshal LogicProperties, err: %v", err.Error())
		return err
	}
	// 2.3 序列化主键数组
	primaryKeysBytes, err := sonic.Marshal(objectType.PrimaryKeys)
	if err != nil {
		logger.Errorf("Failed to marshal PrimaryKeys, err: %v", err.Error())
		return err
	}

	data := map[string]any{
		"f_name":             objectType.OTName,
		"f_tags":             tagsStr,
		"f_comment":          objectType.Comment,
		"f_icon":             objectType.Icon,
		"f_color":            objectType.Color,
		"f_bkn_raw_content":  objectType.BKNRawContent,
		"f_data_source":      dataSourceBytes,
		"f_data_properties":  dataPropertiesBytes,
		"f_logic_properties": logicPropertiesBytes,
		"f_primary_keys":     primaryKeysBytes,
		"f_display_key":      objectType.DisplayKey,
		"f_incremental_key":  objectType.IncrementalKey,
		"f_updater":          objectType.Updater.ID,
		"f_updater_type":     objectType.Updater.Type,
		"f_update_time":      objectType.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(OT_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": objectType.OTID}).
		Where(sq.Eq{"f_kn_id": objectType.KNID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update object type by object type id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update object type by object type id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改对象类的 sql 语句: %s", sqlStr))

	var ret sql.Result
	if tx != nil {
		ret, err = tx.Exec(sqlStr, vals...)
	} else {
		ret, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("update object type error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		span.SetStatus(codes.Error, "Update data error")
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
		return err
	}

	if RowsAffected != 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected not equal 1, RowsAffected is %d, ObjectType is %v",
			objectType.OTID, RowsAffected, objectType)
		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected not equal 1, RowsAffected is %d, ObjectType is %v",
			objectType.OTID, RowsAffected, objectType))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ota *objectTypeAccess) UpdateDataProperties(ctx context.Context, tx *sql.Tx, objectType *interfaces.ObjectType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateDataProperties", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	// 2.1 序列化数据属性
	dataPropertiesBytes, err := sonic.Marshal(objectType.DataProperties)
	if err != nil {
		logger.Errorf("Failed to marshal DataProperties, err: %v", err.Error())
		return err
	}

	data := map[string]any{
		"f_data_properties": dataPropertiesBytes,
		"f_bkn_raw_content": objectType.BKNRawContent,
		"f_updater":         objectType.Updater.ID,
		"f_updater_type":    objectType.Updater.Type,
		"f_update_time":     objectType.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(OT_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": objectType.OTID}).
		Where(sq.Eq{"f_kn_id": objectType.KNID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update object type by object type id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update object type by object type id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改对象类的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update object type error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		span.SetStatus(codes.Error, "Update data error")
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		span.SetStatus(codes.Error, "Get RowsAffected error")
		return err
	}

	if RowsAffected != 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected not equal 1, RowsAffected is %d, ObjectType is %v",
			objectType.OTID, RowsAffected, objectType)
		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected not equal 1, RowsAffected is %d, ObjectType is %v",
			objectType.OTID, RowsAffected, objectType))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ota *objectTypeAccess) DeleteObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteObjectTypesByIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(otIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(OT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": otIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete object type by object type id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete object type by object type id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除对象类的 sql 语句: %s; 删除的对象类ids: %v", sqlStr, otIDs))

	var ret sql.Result
	if tx != nil {
		ret, err = tx.Exec(sqlStr, vals...)
	} else {
		ret, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, "Get RowsAffected error")
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	if RowsAffected != int64(len(otIDs)) {
		// 影响行数不等于删除的对象类数量不报错，删除操作已经发生
		logger.Errorf("DELETE %d RowsAffected not equal %d, ObjectType ids is %v",
			len(otIDs), RowsAffected, otIDs)
		o11y.Warn(ctx, fmt.Sprintf("Delete %d RowsAffected not equal %d, ObjectType ids is %v",
			len(otIDs), RowsAffected, otIDs))
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (ota *objectTypeAccess) DeleteObjectTypeStatusByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteObjectTypeStatusByIDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	if len(otIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(OT_STATUS_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": otIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete object type status by object type id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete object type status by object type id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除对象类状态的 sql 语句: %s; 删除的对象类ids: %v", sqlStr, otIDs))

	var ret sql.Result
	if tx != nil {
		ret, err = tx.Exec(sqlStr, vals...)
	} else {
		ret, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, "Get RowsAffected error")
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	if RowsAffected != int64(len(otIDs)) {
		// 影响行数不等于删除的对象类数量不报错，删除操作已经发生
		logger.Errorf("DELETE %d RowsAffected not equal %d, ObjectType ids is %v",
			len(otIDs), RowsAffected, otIDs)
		o11y.Warn(ctx, fmt.Sprintf("Delete %d RowsAffected not equal %d, ObjectType ids is %v",
			len(otIDs), RowsAffected, otIDs))
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (ota *objectTypeAccess) DeleteObjectTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteObjectTypesByKnID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Delete(OT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete object type by object type id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete object type by object type id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除对象类的 sql 语句: %s; 删除的对象类kn_id: %s, branch: %s", sqlStr, knID, branch))

	var ret sql.Result
	if tx != nil {
		ret, err = tx.Exec(sqlStr, vals...)
	} else {
		ret, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, "Get RowsAffected error")
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (ota *objectTypeAccess) DeleteObjectTypeStatusByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteObjectTypeStatusByKnID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, vals, err := sq.Delete(OT_STATUS_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete object type status by object type id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete object type status by object type id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除对象类状态的 sql 语句: %s; 删除的对象类kn_id: %s, branch: %s", sqlStr, knID, branch))

	var ret sql.Result
	if tx != nil {
		ret, err = tx.Exec(sqlStr, vals...)
	} else {
		ret, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		span.SetStatus(codes.Error, "Delete data error")
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, "Get RowsAffected error")
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

func (ota *objectTypeAccess) GetObjectTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetObjectTypeIDsByKnID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//查询
	sqlStr, vals, err := sq.Select(
		"f_id",
	).From(OT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select object type ids by kn_id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select object type ids by kn_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询对象类的 sql 语句: %s.", sqlStr))

	rows, err := ota.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))
		span.SetStatus(codes.Error, "List data error")
		return nil, err
	}
	defer rows.Close()

	otIDs := []string{}
	for rows.Next() {
		var otID string
		err := rows.Scan(
			&otID,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			span.SetStatus(codes.Error, "Row scan error")
			return nil, err
		}

		otIDs = append(otIDs, otID)
	}

	span.SetStatus(codes.Ok, "")
	return otIDs, nil
}

func (ota *objectTypeAccess) UpdateObjectTypeStatus(ctx context.Context, tx *sql.Tx, knID string, branch string, otID string, otStatus interfaces.ObjectTypeStatus) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateObjectTypeStatus", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	//更新
	sqlStr, vals, err := sq.Update(OT_STATUS_TABLE_NAME).
		Set("f_incremental_key", otStatus.IncrementalKey).
		Set("f_incremental_value", otStatus.IncrementalValue).
		Set("f_index", otStatus.Index).
		Set("f_index_available", otStatus.IndexAvailable).
		Set("f_doc_count", otStatus.DocCount).
		Set("f_storage_size", otStatus.StorageSize).
		Set("f_update_time", otStatus.UpdateTime).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		Where(sq.Eq{"f_id": otID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update object type index, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update object type index, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 执行更新
	if tx != nil {
		_, err = tx.Exec(sqlStr, vals...)
	} else {
		_, err = ota.db.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("Failed to exec the sql of update object type index, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to exec the sql of update object type index, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Exec sql failed ")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 拼接 sql 过滤条件
func processQueryCondition(query interfaces.ObjectTypesQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
	if query.NamePattern != "" {
		// 模糊查询，名称或id进行模糊查询，匹配任一即可
		subBuilder = subBuilder.Where(sq.Expr("(instr(ot.f_name, ?) > 0 OR instr(ot.f_id, ?) > 0)", query.NamePattern, query.NamePattern))
	}

	if query.Tag != "" {
		subBuilder = subBuilder.Where(sq.Expr("instr(ot.f_tags, ?) > 0", `"`+query.Tag+`"`))
	}

	if query.KNID != "" {
		subBuilder = subBuilder.Where(sq.Eq{"ot.f_kn_id": query.KNID})
	}

	if query.Branch != "" {
		subBuilder = subBuilder.Where(sq.Eq{"ot.f_branch": query.Branch})
	} else {
		// 查主线分支的业务知识网络
		subBuilder = subBuilder.Where(sq.Eq{"ot.f_branch": interfaces.MAIN_BRANCH})
	}

	if query.OTIDS != nil {
		subBuilder = subBuilder.Where(sq.Eq{"ot.f_id": query.OTIDS})
	}

	return subBuilder
}

func (ota *objectTypeAccess) GetAllObjectTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*interfaces.ObjectType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetAllObjectTypesByKnID", trace.WithSpanKind(trace.SpanKindClient))
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
		"f_data_source",
		"f_data_properties",
		"f_logic_properties",
		"f_primary_keys",
		"f_display_key",
		"f_incremental_key",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(OT_TABLE_NAME).
		Where(sq.Eq{"f_kn_id": knID}).
		Where(sq.Eq{"f_branch": branch}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select object types by kn_id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select object types by kn_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询对象类列表的 sql 语句: %s; knID: %s", sqlStr, knID))

	rows, err := ota.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))
		span.SetStatus(codes.Error, "List data error")
		return map[string]*interfaces.ObjectType{}, err
	}
	defer rows.Close()

	objectTypes := make(map[string]*interfaces.ObjectType)
	for rows.Next() {
		objectType := interfaces.ObjectType{
			ModuleType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		}
		tagsStr := ""
		var (
			dataSourceBytes      []byte
			dataPropertiesBytes  []byte
			logicPropertiesBytes []byte
			primaryKeysBytes     []byte
		)
		err := rows.Scan(
			&objectType.OTID,
			&objectType.OTName,
			&tagsStr,
			&objectType.Comment,
			&objectType.Icon,
			&objectType.Color,
			&objectType.BKNRawContent,
			&objectType.KNID,
			&objectType.Branch,
			&dataSourceBytes,
			&dataPropertiesBytes,
			&logicPropertiesBytes,
			&primaryKeysBytes,
			&objectType.DisplayKey,
			&objectType.IncrementalKey,
			&objectType.Creator.ID,
			&objectType.Creator.Type,
			&objectType.CreateTime,
			&objectType.Updater.ID,
			&objectType.Updater.Type,
			&objectType.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			span.SetStatus(codes.Error, "Row scan error")
			return map[string]*interfaces.ObjectType{}, err
		}

		// tags string 转成数组的格式
		objectType.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 2.0 反序列化datasource
		err = sonic.Unmarshal(dataSourceBytes, &objectType.DataSource)
		if err != nil {
			logger.Errorf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataSource after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal dataSource error")
			return map[string]*interfaces.ObjectType{}, err
		}

		// 2.1 反序列化DataProperties
		err = sonic.Unmarshal(dataPropertiesBytes, &objectType.DataProperties)
		if err != nil {
			logger.Errorf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal dataProperties after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal dataProperties error")
			return map[string]*interfaces.ObjectType{}, err
		}

		// 2.2 反序列化LogicProperties
		err = sonic.Unmarshal(logicPropertiesBytes, &objectType.LogicProperties)
		if err != nil {
			logger.Errorf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal logicProperties after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal logicProperties error")
			return map[string]*interfaces.ObjectType{}, err
		}

		// 2.3 反序列化主键
		err = sonic.Unmarshal(primaryKeysBytes, &objectType.PrimaryKeys)
		if err != nil {
			logger.Errorf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error())
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal primaryKeys after getting object type, err: %v", err.Error()))
			span.SetStatus(codes.Error, "Unmarshal primaryKeys error")
			return map[string]*interfaces.ObjectType{}, err
		}

		objectTypes[objectType.OTID] = &objectType
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, nil
}
