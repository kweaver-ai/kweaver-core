// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package postgresql

import (
	"context"
	"fmt"
	"strings"
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

	// 构建SELECT子句
	selectFields := []string{}

	// 添加GROUP BY字段（聚合查询时）
	for _, groupByItem := range params.GroupBy {
		if field, ok := fieldMap[groupByItem.Property]; ok {
			selectFields = append(selectFields, field.OriginalName)
		} else {
			selectFields = append(selectFields, groupByItem.Property)
		}
	}

	// 添加聚合字段（聚合查询时）
	var aggAlias string
	if params.Aggregation != nil {
		aggField := params.Aggregation.Property
		if field, ok := fieldMap[aggField]; ok {
			aggField = field.OriginalName
		}

		// 确定聚合函数
		aggFunc := params.Aggregation.Aggr
		switch aggFunc {
		case "count_distinct":
			aggFunc = "COUNT(DISTINCT " + aggField + ")"
		default:
			aggFunc = strings.ToUpper(aggFunc) + "(" + aggField + ")"
		}

		// 确定别名
		if params.Aggregation.Alias != "" {
			aggAlias = params.Aggregation.Alias
		} else {
			aggAlias = "__value"
		}

		selectFields = append(selectFields, aggFunc+" AS "+aggAlias)
	}

	// 如果不是聚合查询且没有指定GROUP BY，则添加所有字段
	if len(params.GroupBy) == 0 && params.Aggregation == nil {
		if len(params.OutputFields) > 0 {
			for _, field := range params.OutputFields {
				if prop, ok := fieldMap[field]; ok {
					selectFields = append(selectFields, prop.OriginalName)
				} else {
					selectFields = append(selectFields, field)
				}
			}
		} else if len(selectFields) == 0 {
			// 没有指定输出字段，则查询所有字段
			for _, prop := range resource.SchemaDefinition {
				selectFields = append(selectFields, prop.OriginalName)
			}
		}
	}

	// 构建查询
	builder := pgSq.Select(selectFields...).From(tableRef)

	// 添加WHERE条件
	if condition != nil {
		builder = builder.Where(condition)
	}

	// 添加GROUP BY（聚合查询时）
	if len(params.GroupBy) > 0 {
		groupByFields := []string{}
		for _, groupByItem := range params.GroupBy {
			if field, ok := fieldMap[groupByItem.Property]; ok {
				groupByFields = append(groupByFields, field.OriginalName)
			} else {
				groupByFields = append(groupByFields, groupByItem.Property)
			}
		}
		builder = builder.GroupBy(groupByFields...)
	}

	// 添加HAVING条件（聚合查询时）
	if params.Having != nil && params.Aggregation != nil {
		havingCond, havingErr := c.buildHavingCondition(params.Having, aggAlias)
		if havingErr != nil {
			return nil, fmt.Errorf("failed to build HAVING condition: %w", havingErr)
		}
		if havingCond != "" {
			builder = builder.Where(havingCond)
		}
	}

	// 添加ORDER BY
	if len(params.Sort) > 0 {
		for _, sortItem := range params.Sort {
			dir := "ASC"
			if sortItem.Direction == interfaces.DESC_DIRECTION {
				dir = "DESC"
			}
			builder = builder.OrderBy(sortItem.Field + " " + dir)
		}
	}

	// 添加LIMIT和OFFSET
	if params.CursorEncoded == "" {
		builder = builder.Offset(uint64(params.Offset))
	}
	builder = builder.Limit(uint64(params.Limit))

	// 构建SQL并执行
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	isAggregate := params.Aggregation != nil || len(params.GroupBy) > 0 || params.Having != nil
	if isAggregate {
		logger.Debugf("postgresql aggregate query: %s, args: %v", query, args)
	} else {
		logger.Debugf("postgresql query: %s, args: %v", query, args)
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

	// 处理总数（仅明细查询）
	if params.NeedTotal && !isAggregate {
		countBuilder := pgSq.Select("COUNT(1)").From(tableRef)
		if condition != nil {
			countBuilder = countBuilder.Where(condition)
		}
		countQuery, countArgs, countErr := countBuilder.ToSql()
		if countErr != nil {
			return nil, fmt.Errorf("failed to build count query: %w", countErr)
		}
		logger.Debugf("postgresql count query: %s, args: %v", countQuery, countArgs)
		var total int64
		row := c.db.QueryRowContext(ctx, countQuery, countArgs...)
		if err := row.Scan(&total); err != nil {
			return nil, fmt.Errorf("failed to scan total: %w", err)
		}
		result.Total = total
	}

	return result, nil
}

// buildHavingCondition 构建HAVING条件
func (c *PostgresqlConnector) buildHavingCondition(having *interfaces.HavingClause, aggAlias string) (string, error) {
	if having.Field != "__value" {
		return "", fmt.Errorf("HAVING field must be '__value'")
	}

	var op string
	switch having.Operation {
	case "==":
		op = "="
	case "!=":
		op = "<>"
	case ">":
		op = ">"
	case ">=":
		op = ">="
	case "<":
		op = "<"
	case "<=":
		op = "<="
	case "in":
		return fmt.Sprintf("%s IN (%s)", aggAlias, formatInValues(having.Value)), nil
	case "not_in":
		return fmt.Sprintf("%s NOT IN (%s)", aggAlias, formatInValues(having.Value)), nil
	case "range":
		if values, ok := having.Value.([]any); ok && len(values) == 2 {
			return fmt.Sprintf("%s BETWEEN ? AND ?", aggAlias), nil
		}
		return "", fmt.Errorf("range operation requires an array with 2 values")
	case "out_range":
		if values, ok := having.Value.([]any); ok && len(values) == 2 {
			return fmt.Sprintf("%s NOT BETWEEN ? AND ?", aggAlias), nil
		}
		return "", fmt.Errorf("out_range operation requires an array with 2 values")
	default:
		return "", fmt.Errorf("unsupported HAVING operation: %s", having.Operation)
	}

	return fmt.Sprintf("%s %s ?", aggAlias, op), nil
}

// formatInValues 格式化IN操作的值列表
func formatInValues(value any) string {
	switch v := value.(type) {
	case []any:
		var values []string
		for _, item := range v {
			values = append(values, fmt.Sprintf("%v", item))
		}
		return strings.Join(values, ", ")
	case []string:
		var values []string
		for _, item := range v {
			values = append(values, fmt.Sprintf("'%s'", item))
		}
		return strings.Join(values, ", ")
	default:
		return fmt.Sprintf("%v", value)
	}
}
