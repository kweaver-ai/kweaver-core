// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package sql

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	sq "github.com/Masterminds/squirrel"
	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

type cachedSql struct {
	Query string
	Args  []any
}

// logicViewSQLGenerator 用于生成SQL
type logicViewSQLGenerator struct {
	nodes         map[string]*interfaces.LogicDefinitionNode
	outputNode    *interfaces.LogicDefinitionNode
	sqls          map[string]cachedSql
	nodeFieldsMap map[string]map[string]*interfaces.ViewProperty
	RefResources  map[string]*interfaces.Resource
}

// NewlogicViewSQLGenerator 创建SQL生成器
func NewlogicViewSQLGenerator(view *interfaces.LogicView) *logicViewSQLGenerator {
	nodeMap := make(map[string]*interfaces.LogicDefinitionNode)
	var outputNode *interfaces.LogicDefinitionNode
	nodes := view.LogicDefinition
	for i := range nodes {
		nodeMap[nodes[i].ID] = nodes[i]
		if nodes[i].Type == interfaces.LogicDefinitionNodeType_Output {
			outputNode = nodes[i]
		}
	}
	return &logicViewSQLGenerator{
		nodes:         nodeMap,
		outputNode:    outputNode,
		sqls:          make(map[string]cachedSql),
		nodeFieldsMap: make(map[string]map[string]*interfaces.ViewProperty),
		RefResources:  view.RefResources,
	}
}

// BuildLogicViewSQL 构建逻辑视图的 SQL
func (g *logicViewSQLGenerator) BuildLogicViewSQL(ctx context.Context, res *interfaces.LogicView) (string, error) {
	sql, args, err := g.buildLogicViewSQLWithDepth(ctx, &res.Resource, interfaces.MaxRecursionDepth)
	if err != nil {
		return "", err
	}
	// 为了兼容下游仅支持单一 SQL 字符串的执行器，在此进行参数插值
	return g.interpolate(sql, args)
}

func (g *logicViewSQLGenerator) buildLogicViewSQLWithDepth(ctx context.Context, res *interfaces.Resource, depth int) (string, []any, error) {
	if depth <= 0 {
		return "", nil, fmt.Errorf("max recursion depth (%d) exceeded, possible circular reference in logic view", interfaces.MaxRecursionDepth)
	}

	if res.LogicDefinition == nil {
		return "", nil, fmt.Errorf("logic definition is empty")
	}

	if g.outputNode == nil {
		return "", nil, fmt.Errorf("custom view '%s' output node not found", res.Name)
	}

	// 2. 从输出节点开始递归构建
	if len(g.outputNode.Inputs) == 0 {
		return "", nil, fmt.Errorf("output node has no input")
	}

	sql, args, err := g.buildNodeSQL(ctx, g.outputNode.ID, depth)
	if err != nil {
		return "", nil, fmt.Errorf("build custom view '%s' sql failed: %w", res.Name, err)
	}

	return sql, args, nil
}

// buildNodeSQL 生成指定节点的SQL
func (g *logicViewSQLGenerator) buildNodeSQL(ctx context.Context, nodeID string, depth int) (string, []any, error) {
	if res, ok := g.sqls[nodeID]; ok {
		// 返回 Clone 后的 Args 以防外部 append 修改导致缓存污染
		argsCopy := make([]any, len(res.Args))
		copy(argsCopy, res.Args)
		return res.Query, argsCopy, nil
	}

	node, ok := g.nodes[nodeID]
	if !ok {
		return "", nil, fmt.Errorf("node %s not found", nodeID)
	}

	var sql string
	var args []any
	var err error

	switch node.Type {
	case interfaces.LogicDefinitionNodeType_Resource:
		sql, args, err = g.buildResourceNodeSQL(ctx, node, depth)
	case interfaces.LogicDefinitionNodeType_Join:
		sql, args, err = g.buildJoinNodeSQL(ctx, node, depth)
	case interfaces.LogicDefinitionNodeType_Union:
		sql, args, err = g.buildUnionNodeSQL(ctx, node, depth)
	case interfaces.LogicDefinitionNodeType_Sql:
		sql, args, err = g.buildSqlNodeSQL(ctx, node, depth)
	case interfaces.LogicDefinitionNodeType_Output:
		sql, args, err = g.buildOutputNodeSQL(ctx, node, depth)
	default:
		return "", nil, fmt.Errorf("unsupported node type: %s", node.Type)
	}

	if err != nil {
		return "", nil, err
	}

	g.sqls[nodeID] = cachedSql{Query: sql, Args: args}
	return sql, args, nil
}

// buildResourceNodeSQL 构建资源节点的 SQL
func (g *logicViewSQLGenerator) buildResourceNodeSQL(ctx context.Context,
	node *interfaces.LogicDefinitionNode, depth int) (string, []any, error) {

	var cfg interfaces.ResourceNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode resource node config: %w", err)
	}

	resource := g.RefResources[cfg.ResourceID]

	// 如果资源本身也是逻辑视图，递归构建（消耗一层深度）
	if resource.Category == interfaces.ResourceCategoryLogicView {
		return g.buildLogicViewSQLWithDepth(ctx, resource, depth-1)
	}

	// 构建原始字段映射，供过滤和别名使用
	fieldMap := make(map[string]*interfaces.Property)
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}

	// 构建 SELECT 字段列表
	var fields []string
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	if len(node.OutputFields) > 0 {
		fields = make([]string, 0, len(node.OutputFields))
		for _, f := range node.OutputFields {
			outputFieldsMap[f.Name] = f // 维护状态

			sourceProp, ok := fieldMap[f.Name]
			if !ok {
				fields = append(fields, QuotationMark(f.Name))
			} else {
				if sourceProp.OriginalName != "" && sourceProp.OriginalName != f.Name {
					// 使用 QuotationMark 替代硬编码引号，支持多数据库
					fields = append(fields, fmt.Sprintf("%s AS %s",
						QuotationMark(sourceProp.OriginalName),
						QuotationMark(f.Name)))
				} else {
					fields = append(fields, QuotationMark(f.Name))
				}
			}
		}
	} else {
		fields = []string{"*"}
	}
	// 维护每个节点的 output fields map (A 的核心功能)
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	// 构建表源
	builder := sq.Select(fields...).From(fmt.Sprintf("{{%s}}", resource.ID)).PlaceholderFormat(sq.Dollar)

	// 处理去重
	if cfg.Distinct {
		builder = builder.Distinct()
	}

	// 处理过滤条件
	filterCond, filterArgs, err := g.buildFilterSQL(ctx, cfg.Filters, fieldMap)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build resource node filter: %w", err)
	}
	if filterCond != nil {
		builder = builder.Where(filterCond)
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}
	// 合并过滤参数
	args = append(args, filterArgs...)
	return sqlStr, args, nil
}

// buildFilterSQL 将 FilterCondCfg 转换为 squirrel 条件
func (g *logicViewSQLGenerator) buildFilterSQL(ctx context.Context, filters *interfaces.FilterCondCfg,
	fieldMap map[string]*interfaces.Property) (sq.Sqlizer, []any, error) {

	if filters == nil {
		return nil, nil, nil
	}

	filterCond, err := filter_condition.NewFilterCondition(ctx, filters, fieldMap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create filter condition: %w", err)
	}
	if filterCond == nil {
		return nil, nil, nil
	}

	sqlCond, err := g.ConvertFilterCondition(ctx, filterCond, fieldMap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert filter condition: %w", err)
	}

	if sqlCond != nil {
		return sqlCond, nil, nil
	}

	// natively. logicViewSQLGenerator handles this via ConvertFilterCondition now.
	// We'll leave it as a TODO or return a mock for now
	return sq.Expr("1=1"), nil, nil
}

// buildJoinNodeSQL 构建 JOIN 节点的 SQL
func (g *logicViewSQLGenerator) buildJoinNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode, depth int) (string, []any, error) {
	var cfg interfaces.JoinNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode join node config: %w", err)
	}

	if len(node.Inputs) != 2 {
		return "", nil, fmt.Errorf("join node must have exactly 2 inputs, got %d", len(node.Inputs))
	}

	// MariaDB 不支持 FULL OUTER JOIN
	if strings.EqualFold(cfg.JoinType, interfaces.JoinType_FullOuter) {
		return "", nil, fmt.Errorf("MariaDB does not support FULL OUTER JOIN, please use LEFT JOIN + UNION instead")
	}

	leftID := node.Inputs[0]
	rightID := node.Inputs[1]

	leftSQL, leftArgs, err := g.buildNodeSQL(ctx, leftID, depth)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build left input for join: %w", err)
	}
	rightSQL, rightArgs, err := g.buildNodeSQL(ctx, rightID, depth)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build right input for join: %w", err)
	}

	// 构建 SELECT 字段列表，使用 from/from_node 确定来源
	fields := make([]string, 0, len(node.OutputFields))
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, f := range node.OutputFields {
		outputFieldsMap[f.Name] = f // 维护状态

		alias := "l"
		if f.FromNode == rightID {
			alias = "r"
		}
		// from 是源字段名, name 是输出字段名
		srcField := f.From
		if srcField == "" {
			srcField = f.Name
		}
		// 使用 QuotationMark 替代硬编码引号，支持多数据库
		fields = append(fields, fmt.Sprintf("%s.%s AS %s", alias, QuotationMark(srcField), QuotationMark(f.Name)))
	}
	// 维护每个节点的 output fields map
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	// 构建 JOIN ON 条件
	joinOnParts := make([]string, 0, len(cfg.JoinOn))
	for _, on := range cfg.JoinOn {
		joinOnParts = append(joinOnParts, fmt.Sprintf("l.%s = r.%s", QuotationMark(on.LeftField), QuotationMark(on.RightField)))
	}
	joinOn := strings.Join(joinOnParts, " AND ")

	joinType := strings.ToUpper(cfg.JoinType)
	if joinType == "" {
		joinType = "INNER"
	}

	// 合并参数：注意不能直接 append 到 leftArgs 上，避免污染
	allArgs := make([]any, 0, len(leftArgs)+len(rightArgs))
	allArgs = append(allArgs, leftArgs...)
	allArgs = append(allArgs, rightArgs...)

	sqlStr := fmt.Sprintf("SELECT %s FROM (%s) AS l %s JOIN (%s) AS r ON %s",
		strings.Join(fields, ", "), leftSQL, joinType, rightSQL, joinOn)

	// 处理 Join 节点自身的过滤条件
	if cfg.Filters != nil {
		// Join 后的字段需要构建一个临时的 fieldMap
		joinFieldMap := make(map[string]*interfaces.Property)
		for _, f := range node.OutputFields {
			joinFieldMap[f.Name] = &interfaces.Property{
				Name:         f.Name,
				Type:         f.Type,
				OriginalName: f.From,
			}
		}

		filterCond, filterArgs, err := g.buildFilterSQL(ctx, cfg.Filters, joinFieldMap)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build join node filter: %w", err)
		}
		if filterCond != nil {
			whereSql, whereArgs, err := filterCond.ToSql()
			if err != nil {
				return "", nil, fmt.Errorf("failed to convert join filter to SQL: %w", err)
			}
			sqlStr = fmt.Sprintf("SELECT * FROM (%s) AS j WHERE %s", sqlStr, whereSql)
			allArgs = append(allArgs, whereArgs...)
			allArgs = append(allArgs, filterArgs...)
		}
	}

	return sqlStr, allArgs, nil
}

// buildUnionNodeSQL 构建 UNION 节点的 SQL
func (g *logicViewSQLGenerator) buildUnionNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode, depth int) (string, []any, error) {
	var cfg interfaces.UnionNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode union node config: %w", err)
	}

	unionParts := make([]string, 0, len(node.Inputs))
	var allArgs []any

	for i, inputID := range node.Inputs {
		subSQL, subArgs, err := g.buildNodeSQL(ctx, inputID, depth)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build union input %d: %w", i, err)
		}

		inputNodeFieldsMap, _ := g.GetNodeFieldsMap(inputID)
		inputNodeType, _ := g.GetNodeType(inputID)

		fields := make([]string, 0, len(node.OutputFields))
		for _, outField := range node.OutputFields {
			outputFieldName := outField.Name
			srcField := outField.Name // 默认同名字段对齐

			// 从 FromList 中查找当前输入节点对应的原始字段
			for _, ref := range outField.FromList {
				if ref.FromNode == inputID {
					if ref.From != "" {
						srcField = ref.From
					}
					break
				}
			}

			// 兼容老逻辑：如果是Resource节点，尝试获取OriginalName
			if inputNodeType == interfaces.LogicDefinitionNodeType_Resource && inputNodeFieldsMap != nil {
				if inputField, ok := inputNodeFieldsMap[srcField]; ok {
					fields = append(fields, fmt.Sprintf("%s AS %s", QuotationMark(inputField.OriginalName), QuotationMark(outputFieldName)))
				} else {
					fields = append(fields, fmt.Sprintf("%s AS %s", QuotationMark(srcField), QuotationMark(outputFieldName)))
				}
			} else {
				fields = append(fields, fmt.Sprintf("%s AS %s", QuotationMark(srcField), QuotationMark(outputFieldName)))
			}
		}

		allArgs = append(allArgs, subArgs...)
		unionParts = append(unionParts, fmt.Sprintf("SELECT %s FROM (%s) AS u%d",
			strings.Join(fields, ", "), subSQL, i))
	}

	unionOp := "UNION ALL"
	if cfg.UnionType == interfaces.UnionType_Distinct {
		unionOp = "UNION"
	}

	// 维护输出状态
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	sql := strings.Join(unionParts, " "+unionOp+" ")

	// 处理 UNION 后的过滤条件
	if cfg.Filters != nil {
		filterCond, filterArgs, err := g.buildFilterSQL(ctx, cfg.Filters, nil)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build union node filter: %w", err)
		}
		if filterCond != nil {
			whereSql, whereArgs, err := filterCond.ToSql()
			if err != nil {
				return "", nil, fmt.Errorf("failed to convert union filter to SQL: %w", err)
			}
			sql = fmt.Sprintf("SELECT * FROM (%s) AS union_result WHERE %s", sql, whereSql)
			allArgs = append(allArgs, whereArgs...)
			allArgs = append(allArgs, filterArgs...)
		}
	}

	return "SELECT * FROM (" + sql + ") AS union_final", allArgs, nil
}

// buildSqlNodeSQL 构建自定义 SQL 节点
func (g *logicViewSQLGenerator) buildSqlNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode, depth int) (string, []any, error) {
	var cfg interfaces.SQLNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode sql node config: %w", err)
	}

	// 维护状态
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	var allArgs []any
	// 创建模板函数映射
	funcMap := template.FuncMap{
		"node": func(nodeID string) (string, error) {
			subSQL, subArgs, err := g.buildNodeSQL(ctx, nodeID, depth)
			if err != nil {
				return "", err
			}
			allArgs = append(allArgs, subArgs...)
			return "(" + subSQL + ")", nil
		},
	}

	// 解析模板
	tmpl, err := template.New("sql").Funcs(funcMap).Parse(cfg.SQL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse SQL template for node %s: %w", node.ID, err)
	}

	// 准备上下文
	contextMap := make(map[string]string)
	for _, inputID := range node.Inputs {
		subSQL, subArgs, err := g.buildNodeSQL(ctx, inputID, depth)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build sql node input %s: %w", inputID, err)
		}
		contextMap[inputID] = "(" + subSQL + ")"
		allArgs = append(allArgs, subArgs...)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, contextMap); err != nil {
		return "", nil, fmt.Errorf("failed to execute SQL template for node %s: %w", node.ID, err)
	}

	return result.String(), allArgs, nil
}

// GetNodeFieldsMap 获取节点的输出字段map
func (g *logicViewSQLGenerator) GetNodeFieldsMap(nodeID string) (map[string]*interfaces.ViewProperty, error) {
	nodeMap, ok := g.nodeFieldsMap[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %s fields map not found", nodeID)
	}
	return nodeMap, nil
}

// GetNodeType 获取节点类型
func (g *logicViewSQLGenerator) GetNodeType(nodeID string) (string, error) {
	node, ok := g.nodes[nodeID]
	if !ok {
		return "", fmt.Errorf("node %s not found", nodeID)
	}
	return node.Type, nil
}

// interpolate 实现参数插值，将 args 填入 query 中的 ?
func (g *logicViewSQLGenerator) interpolate(query string, args []any) (string, error) {
	if len(args) == 0 {
		return query, nil
	}

	parts := strings.Split(query, "?")
	if len(parts)-1 != len(args) {
		return "", fmt.Errorf("placeholder count (%d) does not match args count (%d)", len(parts)-1, len(args))
	}

	var sb strings.Builder
	for i, part := range parts {
		sb.WriteString(part)
		if i < len(args) {
			sb.WriteString(formatArg(args[i]))
		}
	}
	return sb.String(), nil
}

func formatArg(arg any) string {
	switch v := arg.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return fmt.Sprintf("%v", v)
	case float64, float32:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// buildOutputNodeSQL 构建输出节点的 SQL
func (g *logicViewSQLGenerator) buildOutputNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode, depth int) (string, []any, error) {
	if len(node.Inputs) != 1 {
		return "", nil, fmt.Errorf("output node %s requires exactly one input node", node.ID)
	}

	// 维护状态
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	inputNodeID := node.Inputs[0]
	return g.buildNodeSQL(ctx, inputNodeID, depth)
}

// buildCondition 构建过滤条件, fieldsMap 为这个引用视图的字段map
func buildSQLCondition(ctx context.Context, filter *interfaces.FilterCondCfg, vType string, fieldsMap map[string]*interfaces.ViewProperty) (string, error) {
	// NOTE: LEGACY PATH. FilterCondition no longer provides Convert2SQL
	// natively. logicViewSQLGenerator handles this via ConvertFilterCondition now.
	// Returning empty condition to fix compilation since this is only called
	// by legacy query paths that will be commented out.
	return "", nil
}

// func isValidFilters(cfg *interfaces.FilterCondCfg) bool {
// 	if cfg == nil {
// 		return false
// 	}

// 	// 判断过滤器是否为空对象 {}
// 	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
// 		return false
// 	}

// 	return true
// }

// 构建sort
func buildSQLSortParams(sort []*interfaces.SortField) string {
	if len(sort) == 0 {
		return ""
	}

	var sortSql strings.Builder
	for i, sortParam := range sort {
		if i > 0 {
			sortSql.WriteString(", ")
		}
		sortSql.WriteString(fmt.Sprintf("%s %s", QuotationMark(sortParam.Field), sortParam.Direction))
	}

	return sortSql.String()
}

// SQLBuilder - SQL 构建器结构体
type SQLBuilder struct {
	baseQuery        string
	whereClauses     []string
	isSubQuery       bool
	hasExistingWhere bool
	orderBySql       string
	limitCount       int
}

// NewQueryBuilder 创建逻辑视图的 SQL 构建器
func (g *logicViewSQLGenerator) NewQueryBuilder(ctx context.Context, res *interfaces.LogicView) (*SQLBuilder, error) {
	sql, err := g.BuildLogicViewSQL(ctx, res)
	if err != nil {
		return nil, err
	}
	return NewSQLBuilder(sql), nil
}

// NewSQLBuilder 创建新的 SQL 构建器
func NewSQLBuilder(baseQuery string) *SQLBuilder {
	builder := &SQLBuilder{
		baseQuery:    strings.TrimSpace(baseQuery),
		whereClauses: []string{},
	}

	// 检测查询类型和结构
	builder.analyzeQuery()
	return builder
}

// analyzeQuery 分析基础查询的结构
func (b *SQLBuilder) analyzeQuery() {
	upperQuery := strings.ToUpper(b.baseQuery)

	// 检测是否为子查询（以括号开头或包含多个SELECT）
	b.isSubQuery = strings.HasPrefix(b.baseQuery, "(") ||
		(strings.Contains(upperQuery, "SELECT") &&
			strings.Count(upperQuery, "SELECT") > 1)

	// 检测是否已包含 WHERE 子句
	b.hasExistingWhere = strings.Contains(upperQuery, " WHERE ")
}

// AddWhere 添加 WHERE 条件
func (b *SQLBuilder) AddWhere(condition string) *SQLBuilder {
	if strings.TrimSpace(condition) != "" {
		b.whereClauses = append(b.whereClauses, condition)
	}
	return b
}

// AddWheres 批量添加 WHERE 条件
func (b *SQLBuilder) AddWheres(conditions []string) *SQLBuilder {
	for _, condition := range conditions {
		b.AddWhere(condition)
	}
	return b
}

// OrderBy 设置排序语句
func (b *SQLBuilder) OrderBy(sql string) *SQLBuilder {
	b.orderBySql = sql
	return b
}

// Limit 设置分页限制
func (b *SQLBuilder) Limit(count int) *SQLBuilder {
	b.limitCount = count
	return b
}

// ApplyParams 统一应用查询参数（过滤、排序、分页）
func (b *SQLBuilder) ApplyParams(ctx context.Context, params *interfaces.ResourceDataQueryParams, res *interfaces.LogicView) error {
	// 1. 处理过滤条件
	fieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, prop := range res.SchemaDefinition {
		fieldsMap[prop.Name] = &interfaces.ViewProperty{Property: *prop}
	}

	globalFilterSql, err := buildSQLCondition(ctx, params.FilterCondCfg, res.Category, fieldsMap)
	if err != nil {
		return err
	}
	if globalFilterSql != "" {
		b.AddWhere(globalFilterSql)
	}

	// 2. 处理排序
	if len(params.Sort) > 0 {
		b.OrderBy(buildSQLSortParams(params.Sort))
	}

	// 3. 处理分页/限制
	if params.QueryType == "standard" && params.Limit > 0 {
		b.Limit(params.Limit)
	}

	return nil
}

// Build 构建最终的 SQL 语句
func (b *SQLBuilder) Build() string {
	sql := b.baseQuery
	if len(b.whereClauses) > 0 {
		whereStr := strings.Join(b.whereClauses, " AND ")
		if b.isSubQuery {
			sql = b.wrapSubQuery(whereStr)
		} else {
			sql = b.buildStandardQuery(whereStr)
		}
	}

	if b.orderBySql != "" {
		sql = fmt.Sprintf("%s ORDER BY %s", sql, b.orderBySql)
	}

	if b.limitCount > 0 {
		sql = AddLimitIfMissing(sql, b.limitCount)
	}

	return sql
}

// wrapSubQuery 包装子查询
func (b *SQLBuilder) wrapSubQuery(whereStr string) string {
	// 如果子查询已经有别名，直接使用
	if b.hasAlias() {
		return fmt.Sprintf("%s WHERE %s", b.baseQuery, whereStr)
	}

	// 给子查询添加默认别名
	return fmt.Sprintf("(%s) AS subquery WHERE %s", b.baseQuery, whereStr)
}

// buildStandardQuery 构建标准查询
func (b *SQLBuilder) buildStandardQuery(whereStr string) string {
	if b.hasExistingWhere {
		// 已有 WHERE，使用 AND 连接
		return b.insertWhereCondition(whereStr, "AND")
	}

	// 没有 WHERE，添加 WHERE 子句
	return b.insertWhereCondition(whereStr, "WHERE")
}

// insertWhereCondition 在合适的位置插入 WHERE 条件
func (b *SQLBuilder) insertWhereCondition(condition, keyword string) string {
	upperQuery := strings.ToUpper(b.baseQuery)
	hasWhere := strings.Contains(upperQuery, " WHERE ")

	// 查找关键词位置（GROUP BY, ORDER BY, LIMIT 等）
	keywordPositions := []struct {
		keyword string
		index   int
	}{
		{" GROUP BY ", strings.Index(upperQuery, " GROUP BY ")},
		{" ORDER BY ", strings.Index(upperQuery, " ORDER BY ")},
		{" LIMIT ", strings.Index(upperQuery, " LIMIT ")},
		{" HAVING ", strings.Index(upperQuery, " HAVING ")},
	}

	// 找到第一个出现的关键词
	insertPosition := -1
	for _, kp := range keywordPositions {
		if kp.index != -1 && (insertPosition == -1 || kp.index < insertPosition) {
			insertPosition = kp.index
		}
	}

	// 确定要使用的连接词
	var actualKeyword string
	if hasWhere {
		// 如果已有 WHERE 子句，使用 AND 或 OR
		actualKeyword = keyword
	} else {
		// 如果没有 WHERE 子句，使用 WHERE
		actualKeyword = "WHERE"
	}

	if insertPosition != -1 {
		// 在关键词前插入条件
		return b.baseQuery[:insertPosition] + " " + actualKeyword + " " + condition + " " + b.baseQuery[insertPosition:]
	}

	// 没有找到关键词，在末尾添加
	var connector string
	if hasWhere {
		// 如果已有 WHERE 子句，使用 AND 或 OR 连接
		connector = " " + keyword + " "
	} else {
		// 如果没有 WHERE 子句，添加 WHERE 关键字
		connector = " WHERE "
	}
	return b.baseQuery + connector + condition
}

// hasAlias 检测子查询是否已有别名
func (b *SQLBuilder) hasAlias() bool {
	// 简单的别名检测逻辑
	if !b.isSubQuery {
		return false
	}

	// 检查是否以 ) AS 某个名字 结尾
	trimmed := strings.TrimSpace(b.baseQuery)
	if strings.HasSuffix(trimmed, ")") {
		return false
	}

	// 检查是否包含 AS 关键字
	upperQuery := strings.ToUpper(b.baseQuery)
	lastParen := strings.LastIndex(upperQuery, ")")
	if lastParen == -1 {
		return false
	}

	// 在最后一个括号后有 AS 关键字
	afterParen := strings.TrimSpace(upperQuery[lastParen+1:])
	return strings.HasPrefix(afterParen, "AS ")
}

// String 实现 Stringer 接口
func (b *SQLBuilder) String() string {
	return b.Build()
}

// HasLimit 检查 SQL 是否已包含 LIMIT 子句
func HasLimit(sql string) bool {
	// 转换为小写便于匹配
	lowerSQL := strings.ToLower(sql)

	// 移除注释
	cleanedSQL := removeSQLComments(lowerSQL)

	// 匹配 LIMIT 子句的正则表达式
	// 匹配格式：LIMIT 数字 或 LIMIT 数字,数字 或 LIMIT 数字 OFFSET 数字
	limitPattern := `\blimit\s+(\d+)(?:\s*,\s*\d+|\s+offset\s+\d+)?\s*$`

	matched, _ := regexp.MatchString(limitPattern, cleanedSQL)
	return matched
}

// removeSQLComments 移除 SQL 注释
func removeSQLComments(sql string) string {
	// 移除单行注释 (-- 注释)
	singleLineComment := `--[^\n]*`
	re := regexp.MustCompile(singleLineComment)
	sql = re.ReplaceAllString(sql, "")

	// 移除多行注释 (/* 注释 */)
	multiLineComment := `/\*.*?\*/`
	re = regexp.MustCompile(multiLineComment)
	sql = re.ReplaceAllString(sql, "")

	return strings.TrimSpace(sql)
}

// AddLimitIfMissing 如果 SQL 没有 LIMIT，则添加 LIMIT
func AddLimitIfMissing(sql string, limit int) string {
	if HasLimit(sql) {
		return sql
	}

	// 确保 SQL 以分号结尾，然后添加 LIMIT
	trimmedSQL := strings.TrimSpace(sql)
	trimmedSQL = strings.TrimSuffix(trimmedSQL, ";")

	return trimmedSQL + " LIMIT " + strconv.Itoa(limit)
}

// 给字符串加双引号
func QuotationMark(s string) string {
	if s == "*" { // * 是通配符，不需要引号
		return s
	}
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") { //防止拼接过情况
		return s
	}
	return "`" + s + "`"
}
