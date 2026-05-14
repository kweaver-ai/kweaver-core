// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package catalog provides Catalog data access operations.
package catalog

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

	"vega-backend/common"
	"vega-backend/interfaces"
)

const (
	CATALOG_TABLE_NAME = "t_catalog"
)

var (
	cAccessOnce sync.Once
	cAccess     interfaces.CatalogAccess
)

type catalogAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// NewCatalogAccess creates ca new CatalogAccess.
func NewCatalogAccess(appSetting *common.AppSetting) interfaces.CatalogAccess {
	cAccessOnce.Do(func() {
		cAccess = &catalogAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return cAccess
}

// Create creates ca new Catalog.
func (ca *catalogAccess) Create(ctx context.Context, catalog *interfaces.Catalog) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Insert into catalog")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(catalog.Tags)

	// Serialize connector config
	connectorConfigStr, err := sonic.MarshalString(catalog.ConnectorCfg)
	if err != nil {
		otellog.LogError(ctx, "Failed to marshal connector config", err)
		return err
	}

	metadataStr, err := sonic.MarshalString(catalog.Metadata)
	if err != nil {
		otellog.LogError(ctx, "Failed to marshal metadata", err)
		return err
	}

	sqlStr, vals, err := sq.Insert(CATALOG_TABLE_NAME).
		Columns(
			"f_id",
			"f_name",
			"f_tags",
			"f_description",
			"f_type",
			"f_connector_type",
			"f_connector_config",
			"f_metadata",
			"f_health_check_enabled",
			"f_health_check_status",
			"f_last_check_time",
			"f_health_check_result",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			catalog.ID,
			catalog.Name,
			tagsStr,
			catalog.Description,
			catalog.Type,
			catalog.ConnectorType,
			connectorConfigStr,
			metadataStr,
			catalog.HealthCheckEnabled,
			catalog.HealthCheckStatus,
			catalog.LastCheckTime,
			catalog.HealthCheckResult,
			catalog.Creator.ID,
			catalog.Creator.Type,
			catalog.CreateTime,
			catalog.Updater.ID,
			catalog.Updater.Type,
			catalog.UpdateTime,
		).ToSql()
	if err != nil {
		otellog.LogError(ctx, "Failed to build insert catalog sql", err)
		return err
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Insert catalog SQL: %s", sqlStr))

	_, err = ca.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		otellog.LogError(ctx, "Insert catalog failed", err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByID retrieves ca Catalog by ID.
func (ca *catalogAccess) GetByID(ctx context.Context, id string) (*interfaces.Catalog, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Query catalog by ID")
	defer span.End()

	span.SetAttributes(attr.Key("catalog_id").String(id))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_type",
		"f_connector_type",
		"f_connector_config",
		"f_metadata",
		"f_health_check_enabled",
		"f_health_check_status",
		"f_last_check_time",
		"f_health_check_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(CATALOG_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select catalog sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	catalog := &interfaces.Catalog{}
	var tagsStr string
	var connectorConfigStr string
	var metadataStr string

	row := ca.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&catalog.ID,
		&catalog.Name,
		&tagsStr,
		&catalog.Description,
		&catalog.Type,
		&catalog.ConnectorType,
		&connectorConfigStr,
		&metadataStr,
		&catalog.HealthCheckEnabled,
		&catalog.HealthCheckStatus,
		&catalog.LastCheckTime,
		&catalog.HealthCheckResult,
		&catalog.Creator.ID,
		&catalog.Creator.Type,
		&catalog.CreateTime,
		&catalog.Updater.ID,
		&catalog.Updater.Type,
		&catalog.UpdateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan catalog failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// tags string 转成数组的格式
	catalog.Tags = libCommon.TagString2TagSlice(tagsStr)

	// Deserialize connector config
	if connectorConfigStr != "" {
		err = sonic.UnmarshalString(connectorConfigStr, &catalog.ConnectorCfg)
		if err != nil {
			logger.Errorf("Failed to unmarshal connector config: %v", err)
			span.SetStatus(codes.Error, "Unmarshal connector failed")
			return nil, err
		}
	}

	if metadataStr != "" {
		err = sonic.UnmarshalString(metadataStr, &catalog.Metadata)
		if err != nil {
			logger.Errorf("Failed to unmarshal metadata: %v", err)
			span.SetStatus(codes.Error, "Unmarshal metadata failed")
			return nil, err
		}
	}

	span.SetStatus(codes.Ok, "")
	return catalog, nil
}

// GetByIDs retrieves ca Catalog by IDs.
func (ca *catalogAccess) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Catalog, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Query catalog by IDs")
	defer span.End()

	span.SetAttributes(attr.Key("catalog_ids").StringSlice(ids))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_type",
		"f_connector_type",
		"f_connector_config",
		"f_metadata",
		"f_health_check_enabled",
		"f_health_check_status",
		"f_last_check_time",
		"f_health_check_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(CATALOG_TABLE_NAME).
		Where(sq.Eq{"f_id": ids}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select catalog sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return []*interfaces.Catalog{}, err
	}

	rows, err := ca.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query catalog failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return []*interfaces.Catalog{}, err
	}
	defer func() { _ = rows.Close() }()

	catalogs := make([]*interfaces.Catalog, 0)
	for rows.Next() {
		catalog := &interfaces.Catalog{}
		var tagsStr string
		var connectorConfigStr string
		var metadataStr string

		err := rows.Scan(
			&catalog.ID,
			&catalog.Name,
			&tagsStr,
			&catalog.Description,
			&catalog.Type,
			&catalog.ConnectorType,
			&connectorConfigStr,
			&metadataStr,
			&catalog.HealthCheckEnabled,
			&catalog.HealthCheckStatus,
			&catalog.LastCheckTime,
			&catalog.HealthCheckResult,
			&catalog.Creator.ID,
			&catalog.Creator.Type,
			&catalog.CreateTime,
			&catalog.Updater.ID,
			&catalog.Updater.Type,
			&catalog.UpdateTime,
		)
		if err != nil {
			logger.Errorf("Scan catalog row failed: %v", err)
			span.SetStatus(codes.Error, "Scan row failed")
			return []*interfaces.Catalog{}, err
		}

		// tags string 转成数组的格式
		catalog.Tags = libCommon.TagString2TagSlice(tagsStr)

		if connectorConfigStr != "" {
			err = sonic.UnmarshalString(connectorConfigStr, &catalog.ConnectorCfg)
			if err != nil {
				logger.Errorf("Failed to unmarshal connector config: %v", err)
				span.SetStatus(codes.Error, "Unmarshal connector config failed")
				return []*interfaces.Catalog{}, err
			}
		}

		if metadataStr != "" {
			err = sonic.UnmarshalString(metadataStr, &catalog.Metadata)
			if err != nil {
				logger.Errorf("Failed to unmarshal metadata: %v", err)
				span.SetStatus(codes.Error, "Unmarshal metadata failed")
				return []*interfaces.Catalog{}, err
			}
		}

		catalogs = append(catalogs, catalog)
	}

	span.SetStatus(codes.Ok, "")
	return catalogs, nil
}

// GetByName retrieves ca Catalog by name.
func (ca *catalogAccess) GetByName(ctx context.Context, name string) (*interfaces.Catalog, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Query catalog by Name")
	defer span.End()

	span.SetAttributes(attr.Key("catalog_name").String(name))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_type",
		"f_connector_type",
		"f_connector_config",
		"f_metadata",
		"f_health_check_enabled",
		"f_health_check_status",
		"f_last_check_time",
		"f_health_check_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(CATALOG_TABLE_NAME).
		Where(sq.Eq{"f_name": name}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select catalog sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	catalog := &interfaces.Catalog{}
	var tagsStr string
	var connectorConfigStr string
	var metadataStr string

	row := ca.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&catalog.ID,
		&catalog.Name,
		&tagsStr,
		&catalog.Description,
		&catalog.Type,
		&catalog.ConnectorType,
		&connectorConfigStr,
		&metadataStr,
		&catalog.HealthCheckEnabled,
		&catalog.HealthCheckStatus,
		&catalog.LastCheckTime,
		&catalog.HealthCheckResult,
		&catalog.Creator.ID,
		&catalog.Creator.Type,
		&catalog.CreateTime,
		&catalog.Updater.ID,
		&catalog.Updater.Type,
		&catalog.UpdateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan catalog failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// tags string 转成数组的格式
	catalog.Tags = libCommon.TagString2TagSlice(tagsStr)

	// Deserialize connector config
	if connectorConfigStr != "" {
		err = sonic.UnmarshalString(connectorConfigStr, &catalog.ConnectorCfg)
		if err != nil {
			logger.Errorf("Failed to unmarshal connector config: %v", err)
			span.SetStatus(codes.Error, "Unmarshal connector failed")
			return nil, err
		}
	}

	if metadataStr != "" {
		err = sonic.UnmarshalString(metadataStr, &catalog.Metadata)
		if err != nil {
			logger.Errorf("Failed to unmarshal metadata: %v", err)
			span.SetStatus(codes.Error, "Unmarshal metadata failed")
			return nil, err
		}
	}

	span.SetStatus(codes.Ok, "")
	return catalog, nil
}

// ListIDs lists Catalog IDs with filters.
func (ca *catalogAccess) ListIDs(ctx context.Context, params interfaces.CatalogsQueryParams) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "List catalog IDs")
	defer span.End()

	builder := sq.Select("f_id").From(CATALOG_TABLE_NAME)

	if params.Tag != "" {
		tag := "%" + params.Tag + "%"
		builder = builder.Where(sq.Like{"f_tags": tag})
	}

	if params.Type != "" {
		builder = builder.Where(sq.Eq{"f_type": params.Type})
	}
	if params.HealthCheckStatus != "" {
		builder = builder.Where(sq.Eq{"f_health_check_status": params.HealthCheckStatus})
	}

	// 排序
	if params.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", params.Sort, params.Direction))
	} else {
		builder = builder.OrderBy("f_update_time DESC")
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	rows, err := ca.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, err
		}
		ids = append(ids, id)
	}

	span.SetStatus(codes.Ok, "")
	return ids, nil
}

// List lists Catalogs with filters.
func (ca *catalogAccess) List(ctx context.Context, params interfaces.CatalogsQueryParams) ([]*interfaces.Catalog, int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "List catalogs")
	defer span.End()

	builder := sq.Select(
		"f_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_type",
		"f_connector_type",
		"f_connector_config",
		"f_metadata",
		"f_health_check_enabled",
		"f_health_check_status",
		"f_last_check_time",
		"f_health_check_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(CATALOG_TABLE_NAME)

	countBuilder := sq.Select("COUNT(*)").From(CATALOG_TABLE_NAME)

	if params.Tag != "" {
		tag := "%" + params.Tag + "%"
		builder = builder.Where(sq.Like{"f_tags": tag})
		countBuilder = countBuilder.Where(sq.Like{"f_tags": tag})
	}

	if params.Type != "" {
		builder = builder.Where(sq.Eq{"f_type": params.Type})
		countBuilder = countBuilder.Where(sq.Eq{"f_type": params.Type})
	}
	if params.HealthCheckStatus != "" {
		builder = builder.Where(sq.Eq{"f_health_check_status": params.HealthCheckStatus})
		countBuilder = countBuilder.Where(sq.Eq{"f_health_check_status": params.HealthCheckStatus})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := ca.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count catalogs: %v", err)
		span.SetStatus(codes.Error, "Count failed")
		return nil, 0, err
	}

	// 分页
	// if params.Limit > 0 {
	// 	builder = builder.Limit(uint64(params.Limit)).Offset(uint64(params.Offset))
	// }
	// 排序
	if params.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", params.Sort, params.Direction))
	} else {
		builder = builder.OrderBy("f_update_time DESC")
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, 0, err
	}

	rows, err := ca.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	catalogs := make([]*interfaces.Catalog, 0)
	for rows.Next() {
		catalog := &interfaces.Catalog{}
		var tagsStr string
		var connectorConfigStr string
		var metadataStr string

		err := rows.Scan(
			&catalog.ID,
			&catalog.Name,
			&tagsStr,
			&catalog.Description,
			&catalog.Type,
			&catalog.ConnectorType,
			&connectorConfigStr,
			&metadataStr,
			&catalog.HealthCheckEnabled,
			&catalog.HealthCheckStatus,
			&catalog.LastCheckTime,
			&catalog.HealthCheckResult,
			&catalog.Creator.ID,
			&catalog.Creator.Type,
			&catalog.CreateTime,
			&catalog.Updater.ID,
			&catalog.Updater.Type,
			&catalog.UpdateTime,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, 0, err
		}

		// tags string 转成数组的格式
		catalog.Tags = libCommon.TagString2TagSlice(tagsStr)

		if connectorConfigStr != "" {
			err = sonic.UnmarshalString(connectorConfigStr, &catalog.ConnectorCfg)
			if err != nil {
				span.SetStatus(codes.Error, "Unmarshal connector config failed")
				return nil, 0, err
			}
		}

		if metadataStr != "" {
			err = sonic.UnmarshalString(metadataStr, &catalog.Metadata)
			if err != nil {
				span.SetStatus(codes.Error, "Unmarshal metadata failed")
				return nil, 0, err
			}
		}

		catalogs = append(catalogs, catalog)
	}

	span.SetStatus(codes.Ok, "")
	return catalogs, total, nil
}

// ListCatalogSrcsIDs lists Catalog Source IDs with filters.
func (ca *catalogAccess) ListCatalogSrcsIDs(ctx context.Context, params interfaces.ListCatalogsQueryParams) ([]string, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "ListCatalogSrcsIDs")
	defer span.End()

	builder := sq.Select("f_id").From(CATALOG_TABLE_NAME)

	if params.ID != "" {
		builder = builder.Where(sq.Eq{"f_id": params.ID})
	}

	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		builder = builder.Where(sq.Like{"f_name": keyword})
	}

	// 排序
	if params.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", params.Sort, params.Direction))
	} else {
		builder = builder.OrderBy("f_update_time DESC")
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	rows, err := ca.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, err
		}
		ids = append(ids, id)
	}

	span.SetStatus(codes.Ok, "")
	return ids, nil
}

// ListCatalogSrcsByIDs lists Catalog Sources by IDs.
func (ca *catalogAccess) ListCatalogSrcsByIDs(ctx context.Context, ids []string) ([]*interfaces.ListCatalogEntry, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "ListCatalogSrcsByIDs")
	defer span.End()

	if len(ids) == 0 {
		return []*interfaces.ListCatalogEntry{}, nil
	}

	builder := sq.Select(
		"f_id",
		"f_name",
	).From(CATALOG_TABLE_NAME).Where(sq.Eq{"f_id": ids})

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	rows, err := ca.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	entries := make([]*interfaces.ListCatalogEntry, 0)
	for rows.Next() {
		entry := &interfaces.ListCatalogEntry{}

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, err
		}

		entry.Type = interfaces.RESOURCE_TYPE_CATALOG
		entries = append(entries, entry)
	}

	span.SetStatus(codes.Ok, "")
	return entries, nil
}

// ListCatalogSrcs lists Catalog Sources with filters.
func (ca *catalogAccess) ListCatalogSrcs(ctx context.Context, params interfaces.ListCatalogsQueryParams) ([]*interfaces.ListCatalogEntry, int64, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "ListCatalogSrcs catalogs")
	defer span.End()

	builder := sq.Select(
		"f_id",
		"f_name",
	).From(CATALOG_TABLE_NAME)

	countBuilder := sq.Select("COUNT(*)").From(CATALOG_TABLE_NAME)

	if params.ID != "" {
		builder = builder.Where(sq.Eq{"f_id": params.ID})
		countBuilder = countBuilder.Where(sq.Eq{"f_id": params.ID})
	}

	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		builder = builder.Where(sq.Like{"f_name": keyword})
		countBuilder = countBuilder.Where(sq.Like{"f_name": keyword})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := ca.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count catalog sources: %v", err)
		span.SetStatus(codes.Error, "Count failed")
		return nil, 0, err
	}

	// 排序
	if params.Sort != "" {
		builder = builder.OrderBy(fmt.Sprintf("%s %s", params.Sort, params.Direction))
	} else {
		builder = builder.OrderBy("f_update_time DESC")
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, 0, err
	}

	rows, err := ca.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	entries := make([]*interfaces.ListCatalogEntry, 0)
	for rows.Next() {
		entry := &interfaces.ListCatalogEntry{}

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, 0, err
		}

		entry.Type = interfaces.RESOURCE_TYPE_CATALOG
		entries = append(entries, entry)
	}

	span.SetStatus(codes.Ok, "")
	return entries, total, nil
}

// Update updates ca Catalog.
func (ca *catalogAccess) Update(ctx context.Context, catalog *interfaces.Catalog) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Update catalog")
	defer span.End()

	span.SetAttributes(attr.Key("catalog_id").String(catalog.ID))

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(catalog.Tags)

	connectorConfigBytes, err := sonic.Marshal(catalog.ConnectorCfg)
	if err != nil {
		span.SetStatus(codes.Error, "Marshal connector config failed")
		return err
	}
	metadataBytes, err := sonic.Marshal(catalog.Metadata)
	if err != nil {
		span.SetStatus(codes.Error, "Marshal metadata failed")
		return err
	}

	sqlStr, vals, err := sq.Update(CATALOG_TABLE_NAME).
		Set("f_name", catalog.Name).
		Set("f_tags", tagsStr).
		Set("f_description", catalog.Description).
		Set("f_connector_type", catalog.ConnectorType).
		Set("f_connector_config", string(connectorConfigBytes)).
		Set("f_metadata", string(metadataBytes)).
		Set("f_health_check_enabled", catalog.HealthCheckEnabled).
		Set("f_health_check_status", catalog.HealthCheckStatus).
		Set("f_last_check_time", catalog.LastCheckTime).
		Set("f_health_check_result", catalog.HealthCheckResult).
		Set("f_updater", catalog.Updater.ID).
		Set("f_updater_type", catalog.Updater.Type).
		Set("f_update_time", catalog.UpdateTime).
		Where(sq.Eq{"f_id": catalog.ID}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = ca.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteByIDs deletes Catalogs by IDs.
func (ca *catalogAccess) DeleteByIDs(ctx context.Context, ids []string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Delete catalogs")
	defer span.End()

	span.SetAttributes(attr.Key("catalog_ids").StringSlice(ids))

	if len(ids) == 0 {
		return nil
	}

	sqlStr, vals, _ := sq.Delete(CATALOG_TABLE_NAME).
		Where(sq.Eq{"f_id": ids}).
		ToSql()

	_, err := ca.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateStatus updates Catalog status.
func (ca *catalogAccess) UpdateHealthCheckStatus(ctx context.Context, id string, status interfaces.CatalogHealthCheckStatus) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Update catalog status")
	defer span.End()

	sqlStr, vals, _ := sq.Update(CATALOG_TABLE_NAME).
		Set("f_health_check_status", status.HealthCheckStatus).
		Set("f_last_check_time", status.LastCheckTime).
		Set("f_health_check_result", status.HealthCheckResult).
		Where(sq.Eq{"f_id": id}).
		ToSql()

	_, err := ca.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update status failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ca *catalogAccess) UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "Update catalog metadata")
	defer span.End()

	metadataBytes, _ := sonic.Marshal(metadata)

	sqlStr, vals, err := sq.Update(CATALOG_TABLE_NAME).
		Set("f_metadata", string(metadataBytes)).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = ca.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
