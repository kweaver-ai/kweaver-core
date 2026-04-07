// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mariadb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/interfaces"
)

// qualTable 返回带反引号限定的表名
func qualTable(res *interfaces.Resource) string {
	ident := res.SourceIdentifier
	if res.Database != "" && !strings.Contains(ident, ".") {
		ident = res.Database + "." + ident
	}
	parts := strings.SplitN(ident, ".", 2)
	for i, p := range parts {
		parts[i] = "`" + strings.ReplaceAll(p, "`", "``") + "`"
	}
	return strings.Join(parts, ".")
}

// ExecuteJoinQuery 执行多表 JOIN 查询；支持 keyset 游标或 OFFSET/LIMIT
func (c *MariaDBConnector) ExecuteJoinQuery(ctx context.Context, catalog *interfaces.Catalog, params *interfaces.JoinQueryParams) (*interfaces.QueryResult, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	if len(params.Resources) == 0 {
		return nil, fmt.Errorf("resources cannot be empty")
	}

	// 构建 alias -> 限定表名 映射
	aliasToTable := make(map[string]string)
	aliasToResource := make(map[string]*interfaces.Resource)
	for i, res := range params.Resources {
		alias := params.ResourceIDToAlias[res.ID]
		if alias == "" {
			alias = fmt.Sprintf("t%d", i+1)
		}
		aliasToTable[alias] = qualTable(res)
		aliasToResource[alias] = res
	}

	// 合并所有资源的 fieldMap（带别名前缀）
	fieldMap := make(map[string]*interfaces.Property)
	for alias, res := range aliasToResource {
		for _, prop := range res.SchemaDefinition {
			key := alias + "." + prop.Name
			fieldMap[key] = prop
			fieldMap[prop.Name] = prop // 无前缀也支持
		}
	}

	var filterCond sq.Sqlizer
	if params.ActualFilterCond != nil {
		var err error
		filterCond, err = c.ConvertFilterConditionWithAlias(ctx, params.ActualFilterCond, fieldMap, aliasToResource)
		if err != nil {
			return nil, err
		}
	}

	result := &interfaces.QueryResult{Rows: make([]map[string]any, 0)}

	// 构建 SELECT 字段
	selectFields := params.OutputFields
	if len(selectFields) == 0 {
		selectFields = []string{"*"}
	}
	// 将 output_fields 转为 SQL 片段
	selectExprs := make([]string, 0, len(selectFields))
	for _, f := range selectFields {
		if f == "*" {
			selectExprs = append(selectExprs, "*")
		} else {
			qualified := qualifyFieldForSQL(f)
			selectExprs = append(selectExprs, qualified+" AS `"+strings.ReplaceAll(f, "`", "``")+"`")
		}
	}

	// 构建 FROM 和 JOIN
	firstAlias := ""
	firstTable := ""
	for a, t := range aliasToTable {
		firstAlias = a
		firstTable = t
		break
	}

	builder := sq.Select(selectExprs...).From(firstTable + " AS " + firstAlias)

	if len(params.Joins) > 0 {
		// 按 JOIN 顺序添加右表
		joined := map[string]bool{firstAlias: true}
		for _, j := range params.Joins {
			if _, rok := aliasToTable[j.RightTableAlias]; !rok {
				return nil, fmt.Errorf("join right alias not in tables: %s", j.RightTableAlias)
			}
			if joined[j.RightTableAlias] {
				continue
			}
			rt := aliasToTable[j.RightTableAlias]
			onParts := make([]string, 0, len(j.On))
			for _, on := range j.On {
				onParts = append(onParts, qualifyFieldForSQL(on.LeftField)+" = "+qualifyFieldForSQL(on.RightField))
			}
			joinType := strings.ToUpper(j.Type)
			if joinType == "" {
				joinType = "INNER"
			}
			joinClause := rt + " AS " + j.RightTableAlias + " ON " + strings.Join(onParts, " AND ")
			switch joinType {
			case "LEFT":
				builder = builder.LeftJoin(joinClause)
			case "RIGHT":
				builder = builder.RightJoin(joinClause)
			default:
				builder = builder.Join(joinClause)
			}
			joined[j.RightTableAlias] = true
		}
	} else if len(aliasToTable) > 1 {
		// 多表无 JOIN：FROM t1, t2
		fromParts := make([]string, 0, len(aliasToTable))
		for a, t := range aliasToTable {
			fromParts = append(fromParts, t+" AS "+a)
		}
		builder = sq.Select(selectExprs...).From(strings.Join(fromParts, ", "))
	}

	// WHERE：filter + keyset
	var whereConds []sq.Sqlizer
	if filterCond != nil {
		whereConds = append(whereConds, filterCond)
	}
	if params.CursorEncoded != "" {
		keysetCond, err := buildKeysetCondition(params.Sort, params.CursorEncoded, aliasToTable)
		if err != nil {
			return nil, err
		}
		if keysetCond != nil {
			whereConds = append(whereConds, keysetCond)
		}
	}
	if len(whereConds) > 0 {
		builder = builder.Where(sq.And(whereConds))
	}

	// ORDER BY
	sortFields := params.Sort
	if len(sortFields) == 0 && len(params.Resources) > 0 {
		// 默认补主键
		if pk := getFirstPK(params.Resources[0]); pk != "" {
			sortFields = []*interfaces.SortField{{Field: pk, Direction: interfaces.ASC_DIRECTION}}
		}
	}
	for _, sf := range sortFields {
		dir := "ASC"
		if sf.Direction == interfaces.DESC_DIRECTION {
			dir = "DESC"
		}
		builder = builder.OrderBy(qualifyField(sf.Field, aliasToTable) + " " + dir)
	}

	// NeedTotal: COUNT
	if params.NeedTotal {
		countBuilder := sq.Select("COUNT(1)").From(firstTable + " AS " + firstAlias)
		if len(params.Joins) > 0 {
			for _, j := range params.Joins {
				rt := aliasToTable[j.RightTableAlias]
				onParts := make([]string, 0, len(j.On))
				for _, on := range j.On {
					onParts = append(onParts, qualifyFieldForSQL(on.LeftField)+" = "+qualifyFieldForSQL(on.RightField))
				}
				countBuilder = countBuilder.Join(rt + " AS " + j.RightTableAlias + " ON " + strings.Join(onParts, " AND "))
			}
		} else if len(aliasToTable) > 1 {
			fromParts := make([]string, 0, len(aliasToTable))
			for a, t := range aliasToTable {
				fromParts = append(fromParts, t+" AS "+a)
			}
			countBuilder = sq.Select("COUNT(1)").From(strings.Join(fromParts, ", "))
		}
		if len(whereConds) > 0 {
			countBuilder = countBuilder.Where(sq.And(whereConds))
		}
		cq, ca, err := countBuilder.ToSql()
		if err != nil {
			return nil, fmt.Errorf("build count query: %w", err)
		}
		var total int64
		if err := c.db.QueryRowContext(ctx, cq, ca...).Scan(&total); err != nil {
			return nil, fmt.Errorf("count query: %w", err)
		}
		result.Total = total
	}

	// LIMIT / OFFSET
	if params.CursorEncoded == "" {
		builder = builder.Offset(uint64(params.Offset))
	}
	builder = builder.Limit(uint64(params.Limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	logger.Debugf("join query: %s, args: %v", query, args)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result.Columns = cols

	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any)
		for i, col := range cols {
			row[col] = convertValue(values[i])
		}
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

// qualifyField 将 "alias.col" 或 "col" 转为 SQL 片段（用于 SELECT/ORDER BY）
func qualifyField(field string, aliasToTable map[string]string) string {
	field = strings.TrimSpace(field)
	return qualifyFieldForSQL(field)
}

// qualifyFieldForSQL 将 "alias.col" 转为 "`alias`.`col`"，"col" 转为 "`col`"
func qualifyFieldForSQL(field string) string {
	field = strings.TrimSpace(field)
	if idx := strings.Index(field, "."); idx >= 0 {
		alias := field[:idx]
		col := field[idx+1:]
		return "`" + strings.ReplaceAll(alias, "`", "``") + "`." + "`" + strings.ReplaceAll(col, "`", "``") + "`"
	}
	return "`" + strings.ReplaceAll(field, "`", "``") + "`"
}

func getFirstPK(res *interfaces.Resource) string {
	if res.SourceMetadata == nil {
		return ""
	}
	if v, ok := res.SourceMetadata["primary_keys"]; ok {
		if arr, ok := v.([]any); ok && len(arr) > 0 {
			if s, ok := arr[0].(string); ok {
				return s
			}
		}
		if arr, ok := v.([]string); ok && len(arr) > 0 {
			return arr[0]
		}
	}
	return ""
}

// buildKeysetCondition 从编码游标构建 (sort_cols) > (cursor_vals)
func buildKeysetCondition(sort []*interfaces.SortField, cursorEncoded string, _ map[string]string) (sq.Sqlizer, error) {
	if len(sort) == 0 {
		return nil, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(cursorEncoded)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}
	var vals []any
	if err := json.Unmarshal(decoded, &vals); err != nil {
		return nil, fmt.Errorf("unmarshal cursor: %w", err)
	}
	if len(vals) != len(sort) {
		return nil, fmt.Errorf("cursor length mismatch")
	}
	// 构建 (col1, col2, ...) > (v1, v2, ...)
	cols := make([]string, len(sort))
	for i, sf := range sort {
		cols[i] = qualifyFieldForSQL(sf.Field)
	}
	placeholders := make([]string, len(vals))
	for i := range vals {
		placeholders[i] = "?"
	}
	return sq.Expr("("+strings.Join(cols, ",")+") > ("+strings.Join(placeholders, ",")+")", vals...), nil
}

// ConvertFilterConditionWithAlias 支持带别名的字段名
func (c *MariaDBConnector) ConvertFilterConditionWithAlias(ctx context.Context, cond interfaces.FilterCondition,
	fieldMap map[string]*interfaces.Property, aliasToResource map[string]*interfaces.Resource) (sq.Sqlizer, error) {
	// 复用现有 ConvertFilterCondition，fieldMap 已包含 alias.field
	return c.ConvertFilterCondition(ctx, cond, fieldMap)
}
