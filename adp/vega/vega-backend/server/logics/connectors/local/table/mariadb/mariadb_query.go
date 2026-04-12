// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mariadb provides MariaDB database connector implementation.
package mariadb

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/interfaces"
)

// convertValue converts []byte to string for MariaDB driver compatibility
func convertValue(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func (c *MariaDBConnector) ExecuteQuery(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {

	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
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

	if params.NeedTotal {
		countBuilder := sq.Select("COUNT(1)").
			From(resource.SourceIdentifier)

		if condition != nil {
			countBuilder = countBuilder.Where(condition)
		}

		query, args, err := countBuilder.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build query: %w", err)
		}

		logger.Debugf("count query: %s, args: %v", query, args)

		var total int64
		row := c.db.QueryRowContext(ctx, query, args...)
		if err := row.Scan(&total); err != nil {
			return nil, fmt.Errorf("failed to scan total: %w", err)
		}

		result.Total = total
	}

	fields := []string{}
	if len(params.OutputFields) > 0 {
		for _, outName := range params.OutputFields {
			if field, ok := fieldMap[outName]; ok {
				fields = append(fields, field.OriginalName)
			}
		}
	}
	if len(fields) == 0 {
		fields = append(fields, "*")
	}

	builder := sq.Select(fields...).
		From(resource.SourceIdentifier)

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

	logger.Debugf("query: %s, args: %v", query, args)

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
			row[col] = convertValue(values[i])
		}
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}
