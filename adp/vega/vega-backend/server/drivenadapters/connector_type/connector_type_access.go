// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package connectortype provides ConnectorType data access operations.
package connectortype

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
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
	CONNECTOR_TYPE_TABLE_NAME = "t_connector_type"
)

var (
	ctAccessOnce sync.Once
	ctAccess     interfaces.ConnectorTypeAccess
)

type connectorTypeAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// NewConnectorTypeAccess creates a new ConnectorTypeAccess.
func NewConnectorTypeAccess(appSetting *common.AppSetting) interfaces.ConnectorTypeAccess {
	ctAccessOnce.Do(func() {
		ctAccess = &connectorTypeAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return ctAccess
}

// Create creates a new ConnectorType.
func (cta *connectorTypeAccess) Create(ctx context.Context, ct *interfaces.ConnectorType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into connector_type",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(ct.Tags)

	// Serialize FieldConfig to JSON
	fieldConfigStr, err := sonic.MarshalString(ct.FieldConfig)
	if err != nil {
		logger.Errorf("Failed to marshal field config: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to marshal field config: %v", err))
		span.SetStatus(codes.Error, "Marshal field failed")
		return err
	}

	sqlStr, vals, err := sq.Insert(CONNECTOR_TYPE_TABLE_NAME).
		Columns(
			"f_type",
			"f_name",
			"f_tags",
			"f_description",
			"f_mode",
			"f_category",
			"f_endpoint",
			"f_field_config",
			"f_enabled",
		).
		Values(
			ct.Type,
			ct.Name,
			tagsStr,
			ct.Description,
			string(ct.Mode),
			string(ct.Category),
			ct.Endpoint,
			fieldConfigStr,
			ct.Enabled,
		).ToSql()
	if err != nil {
		logger.Errorf("Failed to build insert connector_type sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build insert connector_type sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Insert connector_type SQL: %s", sqlStr))

	_, err = cta.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Insert connector_type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Insert connector_type failed: %v", err))
		span.SetStatus(codes.Error, "Insert failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByType retrieves a ConnectorType by Type.
func (cta *connectorTypeAccess) GetByType(ctx context.Context, tp string) (*interfaces.ConnectorType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query connector_type by Type",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("connector_type").String(tp))

	sqlStr, vals, err := sq.Select(
		"f_type",
		"f_name",
		"f_tags",
		"f_description",
		"f_mode",
		"f_category",
		"f_endpoint",
		"f_field_config",
		"f_enabled",
	).From(CONNECTOR_TYPE_TABLE_NAME).
		Where(sq.Eq{"f_type": tp}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select connector_type sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	ct := &interfaces.ConnectorType{}
	var tagsStr string
	var fieldConfigStr string

	row := cta.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&ct.Type,
		&ct.Name,
		&tagsStr,
		&ct.Description,
		&ct.Mode,
		&ct.Category,
		&ct.Endpoint,
		&fieldConfigStr,
		&ct.Enabled,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan connector_type failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// tags string 转成数组的格式
	ct.Tags = libCommon.TagString2TagSlice(tagsStr)

	// Deserialize FieldConfig
	if fieldConfigStr != "" {
		err = sonic.UnmarshalString(fieldConfigStr, &ct.FieldConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal field config: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal field config: %v", err))
			span.SetStatus(codes.Error, "Unmarshal field failed")
			return nil, err
		}
	}

	span.SetStatus(codes.Ok, "")
	return ct, nil
}

// GetByName retrieves a ConnectorType by Name.
func (cta *connectorTypeAccess) GetByName(ctx context.Context, name string) (*interfaces.ConnectorType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query connector_type by Name",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("name").String(name))

	sqlStr, vals, err := sq.Select(
		"f_type",
		"f_name",
		"f_tags",
		"f_description",
		"f_mode",
		"f_category",
		"f_endpoint",
		"f_field_config",
		"f_enabled",
	).From(CONNECTOR_TYPE_TABLE_NAME).
		Where(sq.Eq{"f_name": name}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select connector_type sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	ct := &interfaces.ConnectorType{}
	var tagsStr string
	var fieldConfigStr string

	row := cta.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&ct.Type,
		&ct.Name,
		&tagsStr,
		&ct.Description,
		&ct.Mode,
		&ct.Category,
		&ct.Endpoint,
		&fieldConfigStr,
		&ct.Enabled,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan connector_type failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// tags string 转成数组的格式
	ct.Tags = libCommon.TagString2TagSlice(tagsStr)

	// Deserialize FieldConfig
	if fieldConfigStr != "" {
		err = sonic.UnmarshalString(fieldConfigStr, &ct.FieldConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal field config: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal field config: %v", err))
			span.SetStatus(codes.Error, "Unmarshal field failed")
			return nil, err
		}
	}

	span.SetStatus(codes.Ok, "")
	return ct, nil
}

// List lists ConnectorTypes with filters.
func (cta *connectorTypeAccess) List(ctx context.Context, params interfaces.ConnectorTypesQueryParams) ([]*interfaces.ConnectorType, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List connector_types",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	builder := sq.Select(
		"f_type",
		"f_name",
		"f_tags",
		"f_description",
		"f_mode",
		"f_category",
		"f_endpoint",
		"f_field_config",
		"f_enabled",
	).From(CONNECTOR_TYPE_TABLE_NAME)

	countBuilder := sq.Select("COUNT(*)").From(CONNECTOR_TYPE_TABLE_NAME)

	// Apply filters
	if params.Mode != "" {
		builder = builder.Where(sq.Eq{"f_mode": params.Mode})
		countBuilder = countBuilder.Where(sq.Eq{"f_mode": params.Mode})
	}
	if params.Category != "" {
		builder = builder.Where(sq.Eq{"f_category": params.Category})
		countBuilder = countBuilder.Where(sq.Eq{"f_category": params.Category})
	}
	if params.Enabled != nil {
		builder = builder.Where(sq.Eq{"f_enabled": *params.Enabled})
		countBuilder = countBuilder.Where(sq.Eq{"f_enabled": *params.Enabled})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := cta.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Count connector_type failed: %v", err)
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
		builder = builder.OrderBy("f_name ASC")
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, 0, err
	}

	rows, err := cta.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	connectorTypes := make([]*interfaces.ConnectorType, 0)
	for rows.Next() {
		ct := &interfaces.ConnectorType{}
		var tagsStr string
		var fieldConfigStr string
		err := rows.Scan(
			&ct.Type,
			&ct.Name,
			&tagsStr,
			&ct.Description,
			&ct.Mode,
			&ct.Category,
			&ct.Endpoint,
			&fieldConfigStr,
			&ct.Enabled,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, 0, err
		}

		// tags string 转成数组的格式
		ct.Tags = libCommon.TagString2TagSlice(tagsStr)

		// Deserialize FieldConfig
		if fieldConfigStr != "" {
			err = sonic.UnmarshalString(fieldConfigStr, &ct.FieldConfig)
			if err != nil {
				logger.Errorf("Failed to unmarshal field config: %v", err)
				o11y.Error(ctx, fmt.Sprintf("Failed to unmarshal field config: %v", err))
				span.SetStatus(codes.Error, "Unmarshal field failed")
				return nil, 0, err
			}
		}

		connectorTypes = append(connectorTypes, ct)
	}

	span.SetStatus(codes.Ok, "")
	return connectorTypes, total, nil
}

// Update updates a ConnectorType.
func (cta *connectorTypeAccess) Update(ctx context.Context, ct *interfaces.ConnectorType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update connector_type",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("connector_type").String(ct.Type))

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(ct.Tags)

	// Serialize FieldConfig to JSON
	fieldConfigStr, err := sonic.MarshalString(ct.FieldConfig)
	if err != nil {
		span.SetStatus(codes.Error, "Marshal field config failed")
		return err
	}

	sqlStr, vals, err := sq.Update(CONNECTOR_TYPE_TABLE_NAME).
		Set("f_name", ct.Name).
		Set("f_tags", tagsStr).
		Set("f_description", ct.Description).
		Set("f_mode", string(ct.Mode)).
		Set("f_category", string(ct.Category)).
		Set("f_endpoint", ct.Endpoint).
		Set("f_field_config", fieldConfigStr).
		Set("f_enabled", ct.Enabled).
		Where(sq.Eq{"f_type": ct.Type}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = cta.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Delete deletes a ConnectorType by Type.
func (cta *connectorTypeAccess) DeleteByType(ctx context.Context, tp string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete connector_type",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("connector_type").String(tp))

	sqlStr, vals, _ := sq.Delete(CONNECTOR_TYPE_TABLE_NAME).
		Where(sq.Eq{"f_type": tp}).
		ToSql()

	_, err := cta.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// SetEnabled enables/disables a ConnectorType.
func (cta *connectorTypeAccess) SetEnabled(ctx context.Context, tp string, enabled bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Set connector_type enabled",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("connector_type").String(tp))

	sqlStr, vals, _ := sq.Update(CONNECTOR_TYPE_TABLE_NAME).
		Set("f_enabled", enabled).
		Where(sq.Eq{"f_type": tp}).
		ToSql()

	_, err := cta.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update enabled failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
