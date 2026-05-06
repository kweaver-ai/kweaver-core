// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mariadb provides MariaDB database connector implementation.
package mariadb

import (
	"context"
	"fmt"
	"strings"

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

	// 构建SELECT子句
	selectFields := []string{}

	// 添加GROUP BY字段（聚合查询时）
	for _, groupByItem := range params.GroupBy {
		if field, ok := fieldMap[groupByItem.Property]; ok {
			// 检查是否需要使用 calendar_interval
			if groupByItem.CalendarInterval != "" {
				dateFmt := c.buildDateFormat(groupByItem.Property, field.OriginalName, groupByItem.CalendarInterval)
				selectFields = append(selectFields, dateFmt+" AS "+groupByItem.Property)
			} else {
				selectFields = append(selectFields, field.OriginalName)
			}
		} else {
			// 检查是否需要使用 calendar_interval
			if groupByItem.CalendarInterval != "" {
				dateFmt := c.buildDateFormat(groupByItem.Property, groupByItem.Property, groupByItem.CalendarInterval)
				selectFields = append(selectFields, dateFmt+" AS "+groupByItem.Property)
			} else {
				selectFields = append(selectFields, groupByItem.Property)
			}
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
	} else if params.Having != nil && params.Having.Field == "count(*)" {
		// 当HAVING使用count(*)时，自动添加COUNT(*)聚合
		aggAlias = "__value"
		selectFields = append(selectFields, "COUNT(*) AS "+aggAlias)
	}

	// 如果不是聚合查询且没有指定GROUP BY，则添加所有字段
	if len(params.GroupBy) == 0 && params.Aggregation == nil {
		if len(params.OutputFields) > 0 {
			for _, outName := range params.OutputFields {
				if field, ok := fieldMap[outName]; ok {
					selectFields = append(selectFields, field.OriginalName)
				} else {
					// 对于未在Schema中定义的字段，直接使用字段名
					selectFields = append(selectFields, outName)
				}
			}
		} else if len(selectFields) == 0 {
			// 没有指定输出字段，则查询所有字段
			for _, prop := range resource.SchemaDefinition {
				selectFields = append(selectFields, prop.OriginalName)
			}
		}
	} else if len(params.OutputFields) > 0 {
		// 对于聚合查询或GROUP BY查询，确保output_fields中的字段在selectFields中
		for _, outName := range params.OutputFields {
			found := false
			for _, field := range selectFields {
				// 检查字段是否已存在（包括别名）
				if field == outName || strings.HasSuffix(field, " AS "+outName) {
					found = true
					break
				}
			}
			if !found {
				if field, ok := fieldMap[outName]; ok {
					selectFields = append(selectFields, field.OriginalName)
				} else {
					// 对于未在Schema中定义的字段，直接使用字段名
					selectFields = append(selectFields, outName)
				}
			}
		}
	}

	// 构建查询
	builder := sq.Select(selectFields...).From(resource.SourceIdentifier)

	// 添加WHERE条件
	if condition != nil {
		builder = builder.Where(condition)
	}

	// 添加GROUP BY（聚合查询时）
	if len(params.GroupBy) > 0 {
		groupByFields := []string{}
		for _, groupByItem := range params.GroupBy {
			if field, ok := fieldMap[groupByItem.Property]; ok {
				// 检查是否需要使用 calendar_interval
				if groupByItem.CalendarInterval != "" {
					dateFmt := c.buildDateFormat(groupByItem.Property, field.OriginalName, groupByItem.CalendarInterval)
					groupByFields = append(groupByFields, dateFmt)
				} else {
					groupByFields = append(groupByFields, field.OriginalName)
				}
			} else {
				// 检查是否需要使用 calendar_interval
				if groupByItem.CalendarInterval != "" {
					dateFmt := c.buildDateFormat(groupByItem.Property, groupByItem.Property, groupByItem.CalendarInterval)
					groupByFields = append(groupByFields, dateFmt)
				} else {
					groupByFields = append(groupByFields, groupByItem.Property)
				}
			}
		}
		builder = builder.GroupBy(groupByFields...)
	}

	// 添加HAVING条件（聚合查询时）
	if params.Having != nil && (params.Aggregation != nil || (params.Having.Field == "count(*)")) {
		havingCond, err := c.buildHavingCondition(params.Having, aggAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to build HAVING condition: %w", err)
		}
		if havingCond != "" {
			builder = builder.Having(havingCond)
		}
	}

	// 添加ORDER BY
	if len(params.Sort) > 0 {
		for _, sortItem := range params.Sort {
			dir := "ASC"
			if sortItem.Direction == interfaces.DESC_DIRECTION {
				dir = "DESC"
			}

			// 检查是否是 GROUP BY 字段且使用了 calendar_interval
			sortField := sortItem.Field
			for _, groupByItem := range params.GroupBy {
				if groupByItem.Property == sortItem.Field && groupByItem.CalendarInterval != "" {
					// 使用完整的 date_format 表达式
					if field, ok := fieldMap[groupByItem.Property]; ok {
						sortField = c.buildDateFormat(groupByItem.Property, field.OriginalName, groupByItem.CalendarInterval)
					} else {
						sortField = c.buildDateFormat(groupByItem.Property, groupByItem.Property, groupByItem.CalendarInterval)
					}
					break
				}
			}

			builder = builder.OrderBy(sortField + " " + dir)
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
		logger.Debugf("aggregate query: %s, args: %v", query, args)
	} else {
		logger.Debugf("query: %s, args: %v", query, args)
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
			row[col] = convertValue(values[i])
		}
		result.Rows = append(result.Rows, row)
	}

	// 处理总数：设置为实际返回的记录数
	if params.NeedTotal {
		result.Total = int64(len(result.Rows))
	}

	return result, nil
}

// buildHavingCondition 构建HAVING条件
func (c *MariaDBConnector) buildHavingCondition(having *interfaces.HavingClause, aggAlias string) (string, error) {
	// 支持 __value 和 count(*) 字段
	if having.Field != "__value" && having.Field != "count(*)" {
		return "", fmt.Errorf("HAVING field must be '__value' or 'count(*)'")
	}

	// 确定HAVING子句中使用的字段表达式
	var fieldExpr string
	if having.Field == "count(*)" {
		fieldExpr = "COUNT(*)"
	} else {
		fieldExpr = aggAlias
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
		return fmt.Sprintf("%s IN (%s)", fieldExpr, formatInValues(having.Value)), nil
	case "not_in":
		return fmt.Sprintf("%s NOT IN (%s)", fieldExpr, formatInValues(having.Value)), nil
	case "range":
		if values, ok := having.Value.([]any); ok && len(values) == 2 {
			return fmt.Sprintf("%s BETWEEN ? AND ?", fieldExpr), nil
		}
		return "", fmt.Errorf("range operation requires an array with 2 values")
	case "out_range":
		if values, ok := having.Value.([]any); ok && len(values) == 2 {
			return fmt.Sprintf("%s NOT BETWEEN ? AND ?", fieldExpr), nil
		}
		return "", fmt.Errorf("out_range operation requires an array with 2 values")
	default:
		return "", fmt.Errorf("unsupported HAVING operation: %s", having.Operation)
	}

	// 格式化HAVING条件的值
	var valueStr string
	switch v := having.Value.(type) {
	case string:
		valueStr = fmt.Sprintf("'%s'", v)
	case int, int64, float64:
		valueStr = fmt.Sprintf("%v", v)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%s %s %s", fieldExpr, op, valueStr), nil
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

// buildDateFormat 根据 calendar_interval 构建 date_format 表达式
// 支持 OpenSearch 的 calendar_interval 枚举值：minute, hour, day, week, month, quarter, year
// 注意：calendar_interval 的有效性已经在 validate_resource_data.go 中的 validateCalendarInterval 方法中验证过
func (c *MariaDBConnector) buildDateFormat(alias, dateField, calendarInterval string) string {
	var dateFmt string
	switch calendarInterval {
	case interfaces.CALENDAR_UNIT_MINUTE:
		dateFmt = fmt.Sprintf(`date_format(%s,'%s')`, dateField, `%Y-%m-%d %H:%i`)
	case interfaces.CALENDAR_UNIT_HOUR:
		dateFmt = fmt.Sprintf(`date_format(%s,'%s')`, dateField, `%Y-%m-%d %H`)
	case interfaces.CALENDAR_UNIT_DAY:
		dateFmt = fmt.Sprintf(`date_format(%s,'%s')`, dateField, `%Y-%m-%d`)
	case interfaces.CALENDAR_UNIT_WEEK:
		dateFmt = fmt.Sprintf(`date_format(%s,'%s')`, dateField, `%x-%v`)
	case interfaces.CALENDAR_UNIT_MONTH:
		dateFmt = fmt.Sprintf(`date_format(%s,'%s')`, dateField, `%Y-%m`)
	case interfaces.CALENDAR_UNIT_QUARTER:
		dateFmt = fmt.Sprintf(`format('%%d-Q%%d',year(%s),quarter(%s))`, dateField, dateField)
	case interfaces.CALENDAR_UNIT_YEAR:
		dateFmt = fmt.Sprintf(`date_format(%s,'%s')`, dateField, `%Y`)
	}
	return dateFmt
}
