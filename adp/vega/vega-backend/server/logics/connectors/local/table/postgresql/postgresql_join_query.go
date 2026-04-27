// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package postgresql

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

// ExecuteJoinQuery 执行多表 JOIN 查询；支持 keyset 游标或 OFFSET/LIMIT。
func (c *PostgresqlConnector) ExecuteJoinQuery(ctx context.Context, catalog *interfaces.Catalog, params *interfaces.JoinQueryParams) (*interfaces.QueryResult, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	if len(params.Resources) == 0 {
		return nil, fmt.Errorf("resources cannot be empty")
	}

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

	fieldMap := make(map[string]*interfaces.Property)
	for alias, res := range aliasToResource {
		for _, prop := range res.SchemaDefinition {
			key := alias + "." + prop.Name
			fieldMap[key] = prop
			fieldMap[prop.Name] = prop
		}
	}

	// 提前构建 aliasToOrigTypeMap，只存储列名和原始类型的对应关系
	aliasToOrigTypeMap := make(map[string]map[string]string)
	for alias, res := range aliasToResource {
		origTypeMap := make(map[string]string)
		if res.SourceMetadata != nil {
			if columnsAny, ok := res.SourceMetadata["columns"].([]any); ok {
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
		aliasToOrigTypeMap[alias] = origTypeMap
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

	selectFields := params.OutputFields
	if len(selectFields) == 0 {
		selectFields = []string{"*"}
	}
	selectExprs := make([]string, 0, len(selectFields))
	for _, f := range selectFields {
		if f == "*" {
			selectExprs = append(selectExprs, "*")
		} else {
			qualified := qualifyFieldForSQL(f)
			selectExprs = append(selectExprs, qualified+" AS "+pgQuoteIdent(f))
		}
	}

	firstAlias := ""
	firstTable := ""
	for a, t := range aliasToTable {
		firstAlias = a
		firstTable = t
		break
	}

	builder := pgSq.Select(selectExprs...).From(firstTable + " AS " + pgQuoteIdent(firstAlias))

	if len(params.Joins) > 0 {
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
			joinClause := rt + " AS " + pgQuoteIdent(j.RightTableAlias) + " ON " + strings.Join(onParts, " AND ")
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
		fromParts := make([]string, 0, len(aliasToTable))
		for a, t := range aliasToTable {
			fromParts = append(fromParts, t+" AS "+pgQuoteIdent(a))
		}
		builder = pgSq.Select(selectExprs...).From(strings.Join(fromParts, ", "))
	}

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

	sortFields := params.Sort
	if len(sortFields) == 0 && len(params.Resources) > 0 {
		if pk := getFirstPK(params.Resources[0]); pk != "" {
			sortFields = []*interfaces.SortField{{Field: pk, Direction: interfaces.ASC_DIRECTION}}
		}
	}
	for _, sf := range sortFields {
		dir := "ASC"
		if sf.Direction == interfaces.DESC_DIRECTION {
			dir = "DESC"
		}
		builder = builder.OrderBy(qualifyFieldForSQL(sf.Field) + " " + dir)
	}

	if params.NeedTotal {
		countBuilder := pgSq.Select("COUNT(1)").From(firstTable + " AS " + pgQuoteIdent(firstAlias))
		if len(params.Joins) > 0 {
			for _, j := range params.Joins {
				rt := aliasToTable[j.RightTableAlias]
				onParts := make([]string, 0, len(j.On))
				for _, on := range j.On {
					onParts = append(onParts, qualifyFieldForSQL(on.LeftField)+" = "+qualifyFieldForSQL(on.RightField))
				}
				countBuilder = countBuilder.Join(rt + " AS " + pgQuoteIdent(j.RightTableAlias) + " ON " + strings.Join(onParts, " AND "))
			}
		} else if len(aliasToTable) > 1 {
			fromParts := make([]string, 0, len(aliasToTable))
			for a, t := range aliasToTable {
				fromParts = append(fromParts, t+" AS "+pgQuoteIdent(a))
			}
			countBuilder = pgSq.Select("COUNT(1)").From(strings.Join(fromParts, ", "))
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

	if params.CursorEncoded == "" {
		builder = builder.Offset(uint64(params.Offset))
	}
	builder = builder.Limit(uint64(params.Limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	logger.Debugf("postgresql join query: %s, args: %v", query, args)

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
			// 检查是否是带别名的字段（如 t1.column_name）
			if dotIdx := strings.LastIndex(col, "."); dotIdx > 0 {
				alias := col[:dotIdx]
				colName := col[dotIdx+1:]
				if origTypeMap, ok := aliasToOrigTypeMap[alias]; ok {
					// 使用字段名而不是带别名的完整名称
					row[col] = convertValue(values[i], colName, origTypeMap)
					continue
				}
			}
			// 如果不是带别名的字段，或者找不到对应的 origTypeMap，使用原始字段名
			row[col] = convertValue(values[i], col, nil)
		}
		result.Rows = append(result.Rows, row)
	}

	return result, nil
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

// ConvertFilterConditionWithAlias 支持带别名的字段名。
func (c *PostgresqlConnector) ConvertFilterConditionWithAlias(ctx context.Context, cond interfaces.FilterCondition,
	fieldMap map[string]*interfaces.Property, _ map[string]*interfaces.Resource) (sq.Sqlizer, error) {
	return c.ConvertFilterCondition(ctx, cond, fieldMap)
}
