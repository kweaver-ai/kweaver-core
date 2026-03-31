// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package postgresql

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/interfaces"
)

func convertRawValue(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

// convertValue 将带时区的时间值转换为当前时区,并处理其他类型
func convertValue(v any, colName string, origTypeMap map[string]string) any {
	if v == nil {
		return nil
	}

	// 从 origTypeMap 中获取原始类型信息
	origType, ok := origTypeMap[colName]
	if !ok {
		return convertRawValue(v)
	}

	// 只有带时区的时间类型需要转换
	// PostgreSQL 原始类型: timestamptz, timetz, timestamp with time zone, time with time zone
	needsConversion := false
	switch origType {
	case "timestamptz", "timetz", "timestamp with time zone", "time with time zone":
		needsConversion = true
	}

	if !needsConversion {
		return convertRawValue(v)
	}

	// 处理时间类型
	switch t := v.(type) {
	case time.Time:
		// 转换为本地时区
		return t.Local()
	default:
		return convertRawValue(v)
	}
}

// ExecuteQuery 执行单表查询。
func (c *PostgresqlConnector) ExecuteQuery(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {

	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}

	// 提前构建 origTypeMap，只存储列名和原始类型的对应关系
	origTypeMap := map[string]string{}
	if resource.SourceMetadata != nil {
		if columnsAny, ok := resource.SourceMetadata["columns"].([]any); ok {
			for _, colAny := range columnsAny {
				if col, ok := colAny.(map[string]any); ok {
					if name, ok := col["name"].(string); ok {
						if origType, ok := col["orig_type"].(string); ok {
							origTypeMap[name] = origType
						}
					}
				}
			}
		}
	}

	var condition sq.Sqlizer
	var err error
	if params.ActualFilterCond != nil {
		condition, err = c.ConvertFilterCondition(ctx, params.ActualFilterCond, fieldMap)
		if err != nil {
			return nil, err
		}
	}

	result := &interfaces.QueryResult{
		Rows: make([]map[string]any, 0),
	}

	tableRef := qualTable(resource)

	if params.NeedTotal {
		countBuilder := pgSq.Select("COUNT(1)").From(tableRef)
		if condition != nil {
			countBuilder = countBuilder.Where(condition)
		}
		query, args, err := countBuilder.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build query: %w", err)
		}
		logger.Debugf("postgresql count query: %s, args: %v", query, args)
		var total int64
		row := c.db.QueryRowContext(ctx, query, args...)
		if err := row.Scan(&total); err != nil {
			return nil, fmt.Errorf("failed to scan total: %w", err)
		}
		result.Total = total
	}

	fields := []string{"*"}
	if len(params.OutputFields) > 0 {
		fields = params.OutputFields
	}

	builder := pgSq.Select(fields...).From(tableRef)
	if condition != nil {
		builder = builder.Where(condition)
	}

	// ORDER BY
	for _, sf := range params.Sort {
		dir := "ASC"
		if sf.Direction == interfaces.DESC_DIRECTION {
			dir = "DESC"
		}
		builder = builder.OrderBy(sf.Field + " " + dir)
	}

	// LIMIT / OFFSET
	if params.CursorEncoded == "" {
		builder = builder.Offset(uint64(params.Offset))
	}
	builder = builder.Limit(uint64(params.Limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	logger.Debugf("postgresql query: %s, args: %v", query, args)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result.Columns = columns

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any)
		for i, col := range columns {
			row[col] = convertValue(values[i], col, origTypeMap)
		}
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}
