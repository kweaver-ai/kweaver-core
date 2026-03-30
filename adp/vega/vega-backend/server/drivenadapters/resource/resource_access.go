// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package resource provides Resource data access operations.
package resource

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	"vega-backend/interfaces"
)

const (
	RESOURCE_TABLE_NAME = "t_resource"
)

var (
	rAccessOnce sync.Once
	rAccess     interfaces.ResourceAccess
)

type resourceAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// NewResourceAccess creates ra new ResourceAccess.
func NewResourceAccess(appSetting *common.AppSetting) interfaces.ResourceAccess {
	rAccessOnce.Do(func() {
		rAccess = &resourceAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return rAccess
}

// Create creates ra new Resource.
func (ra *resourceAccess) Create(ctx context.Context, resource *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into resource",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(resource.Tags)

	// 序列化 SourceMetadata, SchemaDefinition, LogicDefinition
	sourceMetadataBytes, _ := json.Marshal(resource.SourceMetadata)
	if resource.SourceMetadata == nil {
		sourceMetadataBytes = []byte("{}")
	}
	schemaDefinitionBytes, _ := json.Marshal(resource.SchemaDefinition)
	if resource.SchemaDefinition == nil {
		schemaDefinitionBytes = []byte("[]")
	}
	logicDefinitionBytes, _ := json.Marshal(resource.LogicDefinition)
	if resource.LogicDefinition == nil {
		logicDefinitionBytes = []byte("[]")
	}

	sqlStr, vals, err := sq.Insert(RESOURCE_TABLE_NAME).
		Columns(
			"f_id",
			"f_catalog_id",
			"f_name",
			"f_tags",
			"f_description",
			"f_category",
			"f_status",
			"f_status_message",
			"f_database",
			"f_source_identifier",
			"f_source_metadata",
			"f_schema_definition",

			"f_logic_type",
			"f_logic_definition",

			"f_local_enabled",
			"f_local_storage_engine",
			"f_local_storage_config",
			"f_local_index_name",

			"f_sync_strategy",
			"f_sync_config",
			"f_sync_status",
			"f_last_sync_time",
			"f_sync_error_message",

			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		).
		Values(
			resource.ID,
			resource.CatalogID,
			resource.Name,
			tagsStr,
			resource.Description,
			resource.Category,
			resource.Status,
			resource.StatusMessage,
			resource.Database,
			resource.SourceIdentifier,
			string(sourceMetadataBytes),
			string(schemaDefinitionBytes),

			resource.LogicType,
			string(logicDefinitionBytes),

			false,
			"",
			"",
			"",

			"",
			"",
			"",
			0,
			"",

			resource.Creator.ID,
			resource.Creator.Type,
			resource.CreateTime,
			resource.Updater.ID,
			resource.Updater.Type,
			resource.UpdateTime,
		).ToSql()
	if err != nil {
		logger.Errorf("Failed to build insert resource sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build insert resource sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Insert resource SQL: %s", sqlStr))

	_, err = ra.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Insert resource failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Insert resource failed: %v", err))
		span.SetStatus(codes.Error, "Insert failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByID retrieves ra Resource by ID.
func (ra *resourceAccess) GetByID(ctx context.Context, id string) (*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query resource by ID",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("resource_id").String(id))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_category",
		"f_status",
		"f_status_message",
		"f_database",
		"f_source_identifier",
		"f_source_metadata",
		"f_schema_definition",
		"f_logic_type",
		"f_logic_definition",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RESOURCE_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build query resource sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	resource := &interfaces.Resource{}
	var tagsStr string
	var database, sourceIdentifier, sourceMetadata, schemaDefinition, logicDefinition sql.NullString

	row := ra.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&resource.ID,
		&resource.CatalogID,
		&resource.Name,
		&tagsStr,
		&resource.Description,
		&resource.Category,
		&resource.Status,
		&resource.StatusMessage,
		&database,
		&sourceIdentifier,
		&sourceMetadata,
		&schemaDefinition,
		&resource.LogicType,
		&logicDefinition,
		&resource.Creator.ID,
		&resource.Creator.Type,
		&resource.CreateTime,
		&resource.Updater.ID,
		&resource.Updater.Type,
		&resource.UpdateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan resource failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// tags string 转成数组的格式
	resource.Tags = libCommon.TagString2TagSlice(tagsStr)
	resource.Database = database.String
	resource.SourceIdentifier = sourceIdentifier.String
	if sourceMetadata.Valid && sourceMetadata.String != "" {
		_ = json.Unmarshal([]byte(sourceMetadata.String), &resource.SourceMetadata)
	}
	if schemaDefinition.Valid && schemaDefinition.String != "" {
		_ = json.Unmarshal([]byte(schemaDefinition.String), &resource.SchemaDefinition)
	}
	if logicDefinition.Valid && logicDefinition.String != "" {
		_ = json.Unmarshal([]byte(logicDefinition.String), &resource.LogicDefinition)
	}

	span.SetStatus(codes.Ok, "")
	return resource, nil
}

// GetByIDs retrieves ra Resource by IDs.
func (ra *resourceAccess) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query resources by IDs",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("resource_ids").StringSlice(ids))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_category",
		"f_status",
		"f_status_message",
		"f_database",
		"f_source_identifier",
		"f_source_metadata",
		"f_schema_definition",
		"f_logic_type",
		"f_logic_definition",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RESOURCE_TABLE_NAME).
		Where(sq.Eq{"f_id": ids}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build query resource sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return []*interfaces.Resource{}, err
	}

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query resources failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return []*interfaces.Resource{}, err
	}
	defer rows.Close()

	resources := make([]*interfaces.Resource, 0)
	for rows.Next() {
		resource := &interfaces.Resource{}
		var tagsStr string
		var database, sourceIdentifier, sourceMetadata, schemaDefinition, logicDefinition sql.NullString

		err := rows.Scan(
			&resource.ID,
			&resource.CatalogID,
			&resource.Name,
			&tagsStr,
			&resource.Description,
			&resource.Category,
			&resource.Status,
			&resource.StatusMessage,
			&database,
			&sourceIdentifier,
			&sourceMetadata,
			&schemaDefinition,
			&resource.LogicType,
			&logicDefinition,
			&resource.Creator.ID,
			&resource.Creator.Type,
			&resource.CreateTime,
			&resource.Updater.ID,
			&resource.Updater.Type,
			&resource.UpdateTime,
		)

		if err != nil {
			logger.Errorf("Scan resource row failed: %v", err)
			span.SetStatus(codes.Error, "Scan row failed")
			return []*interfaces.Resource{}, err
		}

		// tags string 转成数组的格式
		resource.Tags = libCommon.TagString2TagSlice(tagsStr)
		resource.Database = database.String
		resource.SourceIdentifier = sourceIdentifier.String
		if sourceMetadata.Valid && sourceMetadata.String != "" {
			_ = json.Unmarshal([]byte(sourceMetadata.String), &resource.SourceMetadata)
		}
		if schemaDefinition.Valid && schemaDefinition.String != "" {
			_ = json.Unmarshal([]byte(schemaDefinition.String), &resource.SchemaDefinition)
		}
		if logicDefinition.Valid && logicDefinition.String != "" {
			_ = json.Unmarshal([]byte(logicDefinition.String), &resource.LogicDefinition)
		}

		resources = append(resources, resource)
	}

	span.SetStatus(codes.Ok, "")
	return resources, nil
}

// GetByIDsBasic retrieves Resources by IDs without parsing sourceMetadata, schemaDefinition and logicDefinition.
// This method is optimized for memory usage when these large fields are not needed.
func (ra *resourceAccess) GetByIDsBasic(ctx context.Context, ids []string) ([]*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query resources by IDs (basic)",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("resource_ids").StringSlice(ids))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_category",
		"f_status",
		"f_status_message",
		"f_database",
		"f_source_identifier",
		"f_logic_type",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RESOURCE_TABLE_NAME).
		Where(sq.Eq{"f_id": ids}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build query resource sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return []*interfaces.Resource{}, err
	}

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query resources failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return []*interfaces.Resource{}, err
	}
	defer rows.Close()

	resources := make([]*interfaces.Resource, 0)
	for rows.Next() {
		resource := &interfaces.Resource{}
		var tagsStr string
		var database, sourceIdentifier sql.NullString

		err := rows.Scan(
			&resource.ID,
			&resource.CatalogID,
			&resource.Name,
			&tagsStr,
			&resource.Description,
			&resource.Category,
			&resource.Status,
			&resource.StatusMessage,
			&database,
			&sourceIdentifier,
			&resource.LogicType,
			&resource.Creator.ID,
			&resource.Creator.Type,
			&resource.CreateTime,
			&resource.Updater.ID,
			&resource.Updater.Type,
			&resource.UpdateTime,
		)

		if err != nil {
			logger.Errorf("Scan resource row failed: %v", err)
			span.SetStatus(codes.Error, "Scan row failed")
			return []*interfaces.Resource{}, err
		}

		// tags string 转成数组的格式
		resource.Tags = libCommon.TagString2TagSlice(tagsStr)
		resource.Database = database.String
		resource.SourceIdentifier = sourceIdentifier.String
		// 不解析sourceMetadata、schemaDefinition和logicDefinition，以减少内存占用
		// 这些字段在需要时可以单独获取

		resources = append(resources, resource)
	}

	span.SetStatus(codes.Ok, "")
	return resources, nil
}

// GetByName retrieves ra Resource by catalog and name.
func (ra *resourceAccess) GetByName(ctx context.Context, catalogID string, name string) (*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query resource by name",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("resource_name").String(name))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_category",
		"f_status",
		"f_status_message",
		"f_database",
		"f_source_identifier",
		"f_source_metadata",
		"f_schema_definition",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RESOURCE_TABLE_NAME).
		Where(sq.Eq{"f_catalog_id": catalogID}).
		Where(sq.Eq{"f_name": name}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select resource sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	resource := &interfaces.Resource{}
	var tagsStr string
	var database, sourceIdentifier, sourceMetadata, schemaDefinition sql.NullString

	row := ra.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&resource.ID,
		&resource.CatalogID,
		&resource.Name,
		&tagsStr,
		&resource.Description,
		&resource.Category,
		&resource.Status,
		&resource.StatusMessage,
		&database,
		&sourceIdentifier,
		&sourceMetadata,
		&schemaDefinition,
		&resource.Creator.ID,
		&resource.Creator.Type,
		&resource.CreateTime,
		&resource.Updater.ID,
		&resource.Updater.Type,
		&resource.UpdateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan resource failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// tags string 转成数组的格式
	resource.Tags = libCommon.TagString2TagSlice(tagsStr)
	resource.Database = database.String
	resource.SourceIdentifier = sourceIdentifier.String
	if sourceMetadata.Valid && sourceMetadata.String != "" {
		_ = json.Unmarshal([]byte(sourceMetadata.String), &resource.SourceMetadata)
	}
	if schemaDefinition.Valid && schemaDefinition.String != "" {
		_ = json.Unmarshal([]byte(schemaDefinition.String), &resource.SchemaDefinition)
	}

	span.SetStatus(codes.Ok, "")
	return resource, nil
}

// ListIDs lists Resource IDs with filters.
func (ra *resourceAccess) ListIDs(ctx context.Context, params interfaces.ResourcesQueryParams) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List resource IDs",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	builder := sq.Select("f_id").From(RESOURCE_TABLE_NAME)

	if params.CatalogID != "" {
		builder = builder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.Category != "" {
		builder = builder.Where(sq.Eq{"f_category": params.Category})
	}
	if params.Status != "" {
		builder = builder.Where(sq.Eq{"f_status": params.Status})
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

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer rows.Close()

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

// List lists Resources with filters.
func (ra *resourceAccess) List(ctx context.Context, params interfaces.ResourcesQueryParams) ([]*interfaces.Resource, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List resources",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	builder := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_category",
		"f_status",
		"f_status_message",
		"f_database",
		"f_source_identifier",
		"f_source_metadata",
		"f_schema_definition",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RESOURCE_TABLE_NAME)

	countBuilder := sq.Select("COUNT(*)").From(RESOURCE_TABLE_NAME)

	if params.CatalogID != "" {
		builder = builder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
		countBuilder = countBuilder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.Category != "" {
		builder = builder.Where(sq.Eq{"f_category": params.Category})
		countBuilder = countBuilder.Where(sq.Eq{"f_category": params.Category})
	}
	if params.Status != "" {
		builder = builder.Where(sq.Eq{"f_status": params.Status})
		countBuilder = countBuilder.Where(sq.Eq{"f_status": params.Status})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := ra.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count resources: %v", err)
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

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer rows.Close()

	resources := make([]*interfaces.Resource, 0)
	for rows.Next() {
		resource := &interfaces.Resource{}
		var tagsStr string
		var database, sourceIdentifier, sourceMetadata, schemaDefinition sql.NullString

		err := rows.Scan(
			&resource.ID,
			&resource.CatalogID,
			&resource.Name,
			&tagsStr,
			&resource.Description,
			&resource.Category,
			&resource.Status,
			&resource.StatusMessage,
			&database,
			&sourceIdentifier,
			&sourceMetadata,
			&schemaDefinition,
			&resource.Creator.ID,
			&resource.Creator.Type,
			&resource.CreateTime,
			&resource.Updater.ID,
			&resource.Updater.Type,
			&resource.UpdateTime,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, 0, err
		}

		// tags string 转成数组的格式
		resource.Tags = libCommon.TagString2TagSlice(tagsStr)
		resource.Database = database.String
		resource.SourceIdentifier = sourceIdentifier.String
		if sourceMetadata.Valid && sourceMetadata.String != "" {
			_ = json.Unmarshal([]byte(sourceMetadata.String), &resource.SourceMetadata)
		}
		if schemaDefinition.Valid && schemaDefinition.String != "" {
			_ = json.Unmarshal([]byte(schemaDefinition.String), &resource.SchemaDefinition)
		}

		resources = append(resources, resource)
	}

	span.SetStatus(codes.Ok, "")
	return resources, total, nil
}

// Update updates ra Resource.
func (ra *resourceAccess) Update(ctx context.Context, resource *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("resource_id").String(resource.ID))

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(resource.Tags)

	// 序列化 SourceMetadata, SchemaDefinition, logicDefinition
	sourceMetadataBytes, _ := json.Marshal(resource.SourceMetadata)
	if resource.SourceMetadata == nil {
		sourceMetadataBytes = []byte("{}")
	}
	schemaDefinitionBytes, _ := json.Marshal(resource.SchemaDefinition)
	if resource.SchemaDefinition == nil {
		schemaDefinitionBytes = []byte("[]")
	}
	logicDefinitionBytes, _ := json.Marshal(resource.LogicDefinition)
	if resource.LogicDefinition == nil {
		logicDefinitionBytes = []byte("[]")
	}

	sqlStr, vals, err := sq.Update(RESOURCE_TABLE_NAME).
		Set("f_catalog_id", resource.CatalogID).
		Set("f_name", resource.Name).
		Set("f_tags", tagsStr).
		Set("f_description", resource.Description).
		Set("f_source_metadata", string(sourceMetadataBytes)).
		Set("f_schema_definition", string(schemaDefinitionBytes)).
		Set("f_logic_type", resource.LogicType).
		Set("f_logic_definition", string(logicDefinitionBytes)).
		Set("f_updater", resource.Updater.ID).
		Set("f_updater_type", resource.Updater.Type).
		Set("f_update_time", resource.UpdateTime).
		Where(sq.Eq{"f_id": resource.ID}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = ra.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByCatalogID retrieves all Resources under a Catalog.
func (ra *resourceAccess) GetByCatalogID(ctx context.Context, catalogID string) ([]*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query resources by catalog ID",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("catalog_id").String(catalogID))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_name",
		"f_tags",
		"f_description",
		"f_category",
		"f_status",
		"f_status_message",
		"f_database",
		"f_source_identifier",
		"f_source_metadata",
		"f_schema_definition",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_updater",
		"f_updater_type",
		"f_update_time",
	).From(RESOURCE_TABLE_NAME).
		Where(sq.Eq{"f_catalog_id": catalogID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build query resources sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Query resources failed: %v", err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer rows.Close()

	resources := make([]*interfaces.Resource, 0)
	for rows.Next() {
		resource := &interfaces.Resource{}
		var tagsStr string
		var database, sourceIdentifier, sourceMetadata, schemaDefinition sql.NullString

		err := rows.Scan(
			&resource.ID,
			&resource.CatalogID,
			&resource.Name,
			&tagsStr,
			&resource.Description,
			&resource.Category,
			&resource.Status,
			&resource.StatusMessage,
			&database,
			&sourceIdentifier,
			&sourceMetadata,
			&schemaDefinition,
			&resource.Creator.ID,
			&resource.Creator.Type,
			&resource.CreateTime,
			&resource.Updater.ID,
			&resource.Updater.Type,
			&resource.UpdateTime,
		)
		if err != nil {
			logger.Errorf("Scan resource row failed: %v", err)
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, err
		}

		resource.Tags = libCommon.TagString2TagSlice(tagsStr)
		resource.Database = database.String
		resource.SourceIdentifier = sourceIdentifier.String
		if sourceMetadata.Valid && sourceMetadata.String != "" {
			_ = json.Unmarshal([]byte(sourceMetadata.String), &resource.SourceMetadata)
		}
		if schemaDefinition.Valid && schemaDefinition.String != "" {
			_ = json.Unmarshal([]byte(schemaDefinition.String), &resource.SchemaDefinition)
		}

		resources = append(resources, resource)
	}

	span.SetStatus(codes.Ok, "")
	return resources, nil
}

// UpdateStatus updates a Resource's status.
func (ra *resourceAccess) UpdateStatus(ctx context.Context, id string, status string, statusMessage string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource status",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("resource_id").String(id),
		attr.Key("status").String(status),
	)

	sqlStr, vals, err := sq.Update(RESOURCE_TABLE_NAME).
		Set("f_status", status).
		Set("f_status_message", statusMessage).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = ra.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ra *resourceAccess) DeleteByIDs(ctx context.Context, ids []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete resources",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("resource_ids").StringSlice(ids))

	if len(ids) == 0 {
		return nil
	}

	sqlStr, vals, _ := sq.Delete(RESOURCE_TABLE_NAME).
		Where(sq.Eq{"f_id": ids}).
		ToSql()

	_, err := ra.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// ListResourceSrcsIDs lists Resource Source IDs with filters.
func (ra *resourceAccess) ListResourceSrcsIDs(ctx context.Context, params interfaces.ListResourcesQueryParams) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListResourceSrcsIDs",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	builder := sq.Select("f_id").From(RESOURCE_TABLE_NAME)

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

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer rows.Close()

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

// ListResourceSrcsByIDs lists Resource Sources by IDs.
func (ra *resourceAccess) ListResourceSrcsByIDs(ctx context.Context, ids []string) ([]*interfaces.ListResourceEntry, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListResourceSrcsByIDs",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if len(ids) == 0 {
		return []*interfaces.ListResourceEntry{}, nil
	}

	builder := sq.Select(
		"f_id",
		"f_name",
	).From(RESOURCE_TABLE_NAME).Where(sq.Eq{"f_id": ids})

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, err
	}
	defer rows.Close()

	entries := make([]*interfaces.ListResourceEntry, 0)
	for rows.Next() {
		entry := &interfaces.ListResourceEntry{}

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, err
		}
		entry.Type = interfaces.RESOURCE_TYPE_RESOURCE
		entries = append(entries, entry)
	}

	span.SetStatus(codes.Ok, "")
	return entries, nil
}

// ListResourceSrcs lists Resource Sources with filters.
func (ra *resourceAccess) ListResourceSrcs(ctx context.Context, params interfaces.ListResourcesQueryParams) ([]*interfaces.ListResourceEntry, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListResourceSrcs",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	builder := sq.Select(
		"f_id",
		"f_name",
	).From(RESOURCE_TABLE_NAME)

	countBuilder := sq.Select("COUNT(*)").From(RESOURCE_TABLE_NAME)

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
	err := ra.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count resource sources: %v", err)
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

	rows, err := ra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer rows.Close()

	entries := make([]*interfaces.ListResourceEntry, 0)
	for rows.Next() {
		entry := &interfaces.ListResourceEntry{}

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, 0, err
		}
		entry.Type = interfaces.RESOURCE_TYPE_RESOURCE
		entries = append(entries, entry)
	}

	span.SetStatus(codes.Ok, "")
	return entries, total, nil
}

func (ra *resourceAccess) CheckExistByCategories(ctx context.Context, catalogID string, categories []string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check resources exist",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("catalog_id").String(catalogID))

	countBuilder := sq.Select("COUNT(*)").From(RESOURCE_TABLE_NAME)

	if catalogID != "" {
		countBuilder = countBuilder.Where(sq.Eq{"f_catalog_id": catalogID})
	}
	if len(categories) > 0 {
		countBuilder = countBuilder.Where(sq.Eq{"f_category": categories})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := ra.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count resources: %v", err)
		span.SetStatus(codes.Error, "Count failed")
		return false, err
	}

	span.SetStatus(codes.Ok, "")
	return total > 0, nil
}

func (ra *resourceAccess) DeleteByCatalogIDs(ctx context.Context, catalogIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete resources by catalog IDs",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("catalog_ids").StringSlice(catalogIDs))

	if len(catalogIDs) == 0 {
		return nil
	}

	sqlStr, vals, _ := sq.Delete(RESOURCE_TABLE_NAME).
		Where(sq.Eq{"f_catalog_id": catalogIDs}).
		ToSql()

	_, err := ra.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
