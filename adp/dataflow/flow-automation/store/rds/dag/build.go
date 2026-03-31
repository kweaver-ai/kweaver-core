package dagmodel

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Converter MongoDB BSON查询条件转SQL转换器
type Converter struct {
	tableName   string
	fieldMap    map[string]string // 自定义字段映射（优先级最高）
	autoConvert bool              // 是否自动转换字段名
}

type Option func(*Converter)

// WithFieldMap 自定义字段映射，优先级最高
func WithFieldMap(m map[string]string) Option {
	return func(c *Converter) {
		c.fieldMap = m
	}
}

// WithAutoConvert 开启自动字段名转换（默认开启）
func WithAutoConvert(on bool) Option {
	return func(c *Converter) {
		c.autoConvert = on
	}
}

func NewConverter(tableName string, opts ...Option) *Converter {
	c := &Converter{
		tableName:   tableName,
		fieldMap:    make(map[string]string),
		autoConvert: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// convertField 将 MongoDB 字段名转为 SQL 字段名
//
// 规则：
//  1. 优先使用自定义映射 fieldMap
//  2. 自动转换：驼峰 → f_ + 下划线分隔
//
// 示例：
//
//	_id        → f_id
//	name       → f_name
//	userId     → f_user_id
//	createdAt  → f_created_at
//	HTMLParser → f_html_parser
func (c *Converter) convertField(mongoField string) string {
	// 逻辑操作符不转换
	if strings.HasPrefix(mongoField, "$") {
		return mongoField
	}

	// 1. 优先自定义映射
	if mapped, ok := c.fieldMap[mongoField]; ok {
		return mapped
	}

	// 2. 不开启自动转换，直接返回
	if !c.autoConvert {
		return mongoField
	}

	// 3. 自动转换：camelCase / PascalCase → f_snake_case
	return camelToFSnake(mongoField)
}

// camelToFSnake 驼峰/任意命名 → f_snake_case
//
//	_id          → f_id
//	name         → f_name
//	userId       → f_user_id
//	createdAt    → f_created_at
//	myHTTPServer → f_my_http_server
//	HTMLParser   → f_html_parser
//	user_name    → f_user_name  (已经是下划线的保持不变)
//	already_f_xx → f_already_f_xx (不会重复加 f_)
func camelToFSnake(s string) string {
	if strings.HasPrefix(s, "f_") {
		return s
	}

	if s == "userid" {
		return "f_user_id"
	}

	// 去除开头的 _ (如 _id → id，后面统一加 f_)
	s = strings.TrimLeft(s, "_")
	if s == "" {
		return "f_"
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// 需要插入下划线的情况：
			// 1. 不是第一个字符
			// 2. 前一个字符不是大写（普通驼峰：userId → user_id）
			//    或者后一个字符是小写（连续大写末尾：HTMLParser → html_parser）
			if i > 0 {
				prevIsUpper := unicode.IsUpper(runes[i-1])
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

				if !prevIsUpper || nextIsLower {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	snaked := result.String()

	// 如果已经有下划线连接的，去除多余的
	snaked = strings.ReplaceAll(snaked, "__", "_")

	return "f_" + snaked
}

type Result struct {
	SQL    string
	Conds  string
	Params []interface{}
}

func (c *Converter) Convert(query interface{}) (*Result, error) {
	m, err := toMap(query)
	if err != nil {
		return nil, fmt.Errorf("invalid query type: %w", err)
	}

	where, params, err := c.parseCondition(m)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("SELECT * FROM %s", c.tableName)
	if where != "" {
		sql += " WHERE " + where
	}

	return &Result{SQL: sql, Params: params}, nil
}

func (c *Converter) ConvertConds(query interface{}) (*Result, error) {
	m, err := toMap(query)
	if err != nil {
		return nil, fmt.Errorf("invalid query type: %w", err)
	}

	where, params, err := c.parseCondition(m)
	if err != nil {
		return nil, err
	}

	return &Result{Conds: where, Params: params}, nil
}

func toMap(v interface{}) (map[string]interface{}, error) {
	switch val := v.(type) {
	case bson.M:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[k] = v
		}
		return result, nil
	case bson.D:
		result := make(map[string]interface{}, len(val))
		for _, elem := range val {
			result[elem.Key] = elem.Value
		}
		return result, nil
	case map[string]interface{}:
		return val, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func toSlice(v interface{}) ([]interface{}, error) {
	switch val := v.(type) {
	case bson.A:
		return []interface{}(val), nil
	case []interface{}:
		return val, nil
	case []bson.M:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	case []bson.D:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	default:
		if v == nil {
			return nil, fmt.Errorf("unsupported slice type: %T", v)
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			result := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = rv.Index(i).Interface()
			}
			return result, nil
		}
		return nil, fmt.Errorf("unsupported slice type: %T", v)
	}
}

func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case primitive.ObjectID:
		return val.Hex()
	case primitive.DateTime:
		return val.Time().Format("2006-01-02 15:04:05")
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case primitive.Timestamp:
		return time.Unix(int64(val.T), 0).Format("2006-01-02 15:04:05")
	case primitive.Decimal128:
		return val.String()
	case primitive.Regex:
		return val.Pattern
	case int32:
		return int64(val)
	case float32:
		return float64(val)
	default:
		return v
	}
}

func (c *Converter) parseCondition(query map[string]interface{}) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := query[key]

		switch key {
		case "$and":
			sql, p, err := c.parseLogicalOp(value, "AND")
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, sql)
			params = append(params, p...)

		case "$or":
			sql, p, err := c.parseLogicalOp(value, "OR")
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, "("+sql+")")
			params = append(params, p...)

		case "$nor":
			sql, p, err := c.parseLogicalOp(value, "OR")
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, "NOT ("+sql+")")
			params = append(params, p...)

		case "$not":
			subMap, err := toMap(value)
			if err != nil {
				return "", nil, fmt.Errorf("$not must be object: %w", err)
			}
			sql, p, err := c.parseCondition(subMap)
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, "NOT ("+sql+")")
			params = append(params, p...)

		default:
			// 这里对字段名做转换
			sqlField := c.convertField(key)
			sqls, p, err := c.parseFieldCondition(sqlField, value)
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, sqls...)
			params = append(params, p...)
		}
	}

	return strings.Join(conditions, " AND "), params, nil
}

func (c *Converter) parseFieldCondition(field string, value interface{}) ([]string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	if subMap, err := toMap(value); err == nil {
		hasOperator := false
		for k := range subMap {
			if strings.HasPrefix(k, "$") {
				hasOperator = true
				break
			}
		}
		if hasOperator {
			ops := make([]string, 0, len(subMap))
			for op := range subMap {
				ops = append(ops, op)
			}
			sort.Strings(ops)
			for _, op := range ops {
				sql, p, err := c.parseOperator(field, op, subMap[op])
				if err != nil {
					return nil, nil, err
				}
				conditions = append(conditions, sql)
				params = append(params, p...)
			}
			return conditions, params, nil
		}
	}

	if regex, ok := value.(primitive.Regex); ok {
		likePattern := regexToLike(regex.Pattern)
		if strings.Contains(regex.Options, "i") {
			conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
		}
		params = append(params, likePattern)
		return conditions, params, nil
	}

	if value == nil {
		conditions = append(conditions, fmt.Sprintf("%s IS NULL", field))
	} else {
		conditions = append(conditions, fmt.Sprintf("%s = ?", field))
		params = append(params, normalizeValue(value))
	}

	return conditions, params, nil
}

func (c *Converter) parseOperator(field, op string, value interface{}) (string, []interface{}, error) {
	switch op {
	case "$eq":
		if value == nil {
			return fmt.Sprintf("%s IS NULL", field), nil, nil
		}
		return fmt.Sprintf("%s = ?", field), []interface{}{normalizeValue(value)}, nil
	case "$ne":
		if value == nil {
			return fmt.Sprintf("%s IS NOT NULL", field), nil, nil
		}
		return fmt.Sprintf("%s != ?", field), []interface{}{normalizeValue(value)}, nil
	case "$gt":
		return fmt.Sprintf("%s > ?", field), []interface{}{normalizeValue(value)}, nil
	case "$gte":
		return fmt.Sprintf("%s >= ?", field), []interface{}{normalizeValue(value)}, nil
	case "$lt":
		return fmt.Sprintf("%s < ?", field), []interface{}{normalizeValue(value)}, nil
	case "$lte":
		return fmt.Sprintf("%s <= ?", field), []interface{}{normalizeValue(value)}, nil
	case "$in":
		return c.parseInOp(field, value, false)
	case "$nin":
		return c.parseInOp(field, value, true)
	case "$exists":
		exists, ok := value.(bool)
		if !ok {
			return "", nil, fmt.Errorf("$exists must be bool")
		}
		if exists {
			return fmt.Sprintf("%s IS NOT NULL", field), nil, nil
		}
		return fmt.Sprintf("%s IS NULL", field), nil, nil
	case "$regex":
		pattern := ""
		switch v := value.(type) {
		case string:
			pattern = v
		case primitive.Regex:
			pattern = v.Pattern
		default:
			return "", nil, fmt.Errorf("$regex must be string or Regex")
		}
		return fmt.Sprintf("%s LIKE ?", field), []interface{}{regexToLike(pattern)}, nil
	case "$mod":
		arr, err := toSlice(value)
		if err != nil || len(arr) != 2 {
			return "", nil, fmt.Errorf("$mod must be [divisor, remainder]")
		}
		return fmt.Sprintf("%s %% ? = ?", field),
			[]interface{}{normalizeValue(arr[0]), normalizeValue(arr[1])}, nil
	case "$not":
		subMap, err := toMap(value)
		if err != nil {
			if regex, ok := value.(primitive.Regex); ok {
				return fmt.Sprintf("%s NOT LIKE ?", field), []interface{}{regexToLike(regex.Pattern)}, nil
			}
			return "", nil, fmt.Errorf("$not must be object or regex")
		}
		var subConds []string
		var subParams []interface{}
		for subOp, subVal := range subMap {
			sql, p, err := c.parseOperator(field, subOp, subVal)
			if err != nil {
				return "", nil, err
			}
			subConds = append(subConds, sql)
			subParams = append(subParams, p...)
		}
		return "NOT (" + strings.Join(subConds, " AND ") + ")", subParams, nil
	case "$size":
		return fmt.Sprintf("JSON_LENGTH(%s) = ?", field), []interface{}{normalizeValue(value)}, nil
	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", op)
	}
}

func (c *Converter) parseLogicalOp(value interface{}, logicOp string) (string, []interface{}, error) {
	arr, err := toSlice(value)
	if err != nil {
		return "", nil, fmt.Errorf("$%s must be array: %w", strings.ToLower(logicOp), err)
	}
	var conditions []string
	var params []interface{}
	for _, item := range arr {
		subMap, err := toMap(item)
		if err != nil {
			return "", nil, fmt.Errorf("$%s elements must be objects: %w", strings.ToLower(logicOp), err)
		}
		sql, p, err := c.parseCondition(subMap)
		if err != nil {
			return "", nil, err
		}
		if sql != "" {
			conditions = append(conditions, sql)
			params = append(params, p...)
		}
	}
	return strings.Join(conditions, fmt.Sprintf(" %s ", logicOp)), params, nil
}

func (c *Converter) parseInOp(field string, value interface{}, negate bool) (string, []interface{}, error) {
	arr, err := toSlice(value)
	if err != nil {
		return "", nil, fmt.Errorf("$in/$nin must be array: %w", err)
	}
	if len(arr) == 0 {
		if negate {
			return "1=1", nil, nil
		}
		return "1=0", nil, nil
	}
	var hasNull bool
	var vals []interface{}
	for _, v := range arr {
		if v == nil {
			hasNull = true
		} else {
			vals = append(vals, normalizeValue(v))
		}
	}
	ph := make([]string, len(vals))
	for i := range vals {
		ph[i] = "?"
	}
	var parts []string
	var params []interface{}
	if len(vals) > 0 {
		op := "IN"
		if negate {
			op = "NOT IN"
		}
		parts = append(parts, fmt.Sprintf("%s %s (%s)", field, op, strings.Join(ph, ", ")))
		params = append(params, vals...)
	}
	if hasNull {
		if negate {
			parts = append(parts, fmt.Sprintf("%s IS NOT NULL", field))
		} else {
			parts = append(parts, fmt.Sprintf("%s IS NULL", field))
		}
	}
	joiner := " OR "
	if negate {
		joiner = " AND "
	}
	result := strings.Join(parts, joiner)
	if len(parts) > 1 {
		result = "(" + result + ")"
	}
	return result, params, nil
}

func regexToLike(pattern string) string {
	pattern = strings.TrimPrefix(pattern, "/")
	if idx := strings.LastIndex(pattern, "/"); idx > 0 {
		pattern = pattern[:idx]
	}
	result := pattern
	hasPrefix := strings.HasPrefix(result, "^")
	hasSuffix := strings.HasSuffix(result, "$")
	result = strings.TrimPrefix(result, "^")
	result = strings.TrimSuffix(result, "$")
	result = strings.ReplaceAll(result, "\\.", "{{DOT}}")
	result = strings.ReplaceAll(result, "%", "\\%")
	result = strings.ReplaceAll(result, "_", "\\_")
	result = strings.ReplaceAll(result, ".*", "%")
	result = strings.ReplaceAll(result, ".+", "_%")
	result = strings.ReplaceAll(result, "{{DOT}}", ".")
	if !hasPrefix {
		result = "%" + result
	}
	if !hasSuffix {
		result = result + "%"
	}
	return result
}

func BuildGroupDagInstanceQuery(input *mod.GroupInput) (string, []interface{}, error) {
	if input == nil {
		return "", nil, nil
	}

	conv := NewConverter(DAGINSTANCE_TABLENAME, WithAutoConvert(true))
	baseConds := make([]string, 0)
	baseArgs := make([]interface{}, 0)

	for _, opt := range input.SearchOptions {
		m := map[string]interface{}{
			opt.Field: map[string]interface{}{
				opt.Condition: opt.Value,
			},
		}
		res, err := conv.ConvertConds(m)
		if err != nil {
			return "", nil, err
		}
		if res.Conds != "" {
			baseConds = append(baseConds, res.Conds)
			baseArgs = append(baseArgs, res.Params...)
		}
	}

	if input.TimeRange != nil {
		m := map[string]interface{}{
			input.TimeRange.Field: map[string]interface{}{
				"$gte": input.TimeRange.Begin,
				"$lte": input.TimeRange.End,
			},
		}
		res, err := conv.ConvertConds(m)
		if err != nil {
			return "", nil, err
		}
		if res.Conds != "" {
			baseConds = append(baseConds, res.Conds)
			baseArgs = append(baseArgs, res.Params...)
		}
	}

	groupCols := make([]string, 0)
	if input.GroupBy != "" {
		groupCols = append(groupCols, camelToFSnake(input.GroupBy))
	}
	for _, g := range input.GroupBys {
		groupCols = append(groupCols, camelToFSnake(g))
	}
	if len(groupCols) == 0 {
		return "", nil, nil
	}

	where := ""
	if len(baseConds) > 0 {
		where = " WHERE " + strings.Join(baseConds, " AND ")
	}

	if !input.IsFirst {
		sql := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s GROUP BY %s", DAGINSTANCE_TABLENAME, where, strings.Join(groupCols, ", "))
		args := append([]interface{}{}, baseArgs...)
		if input.Limit > 0 {
			sql += " LIMIT ?"
			args = append(args, input.Limit)
		}
		return sql, args, nil
	}

	sortCol := "f_id"
	if input.SortBy != "" {
		sortCol = camelToFSnake(input.SortBy)
	}

	minmax := "MAX"
	if input.Order > 0 {
		minmax = "MIN"
	}

	selectCols := "di.*"
	if len(input.ProjectFields) > 0 {
		cols := make([]string, 0, len(input.ProjectFields))
		for _, f := range input.ProjectFields {
			cols = append(cols, "di."+camelToFSnake(f))
		}
		selectCols = strings.Join(cols, ", ")
	}

	selectTotal := "0 AS total"
	gSelectCols := strings.Join(groupCols, ", ")
	if input.IsSum {
		gSelectCols = gSelectCols + ", COUNT(*) AS total"
		selectTotal = "g.total"
	}

	// 子查询 g：算"每组的总数 + 每组的极值 sort_col"
	gSQL := fmt.Sprintf("SELECT %s, %s(%s) AS sort_val FROM %s%s GROUP BY %s", gSelectCols, minmax, sortCol, DAGINSTANCE_TABLENAME, where, strings.Join(groupCols, ", "))
	gArgs := append([]interface{}{}, baseArgs...)
	if input.Limit > 0 {
		gSQL += " LIMIT ?"
		gArgs = append(gArgs, input.Limit)
	}

	// 子查询 p：在每组 + 每个 sort_col 下，取 f_id 最大的那条
	pSQL := fmt.Sprintf("SELECT %s, %s AS sort_val, MAX(f_id) AS max_id FROM %s%s GROUP BY %s, %s", strings.Join(groupCols, ", "), sortCol, DAGINSTANCE_TABLENAME, where, strings.Join(groupCols, ", "), sortCol)
	pArgs := append([]interface{}{}, baseArgs...)

	// 把 g 和 p 连起来：只保留每组的"极值 sort_col"对应的那条 max_id
	joinConds := make([]string, 0, len(groupCols)+1)
	for _, col := range groupCols {
		joinConds = append(joinConds, fmt.Sprintf("g.%s = p.%s", col, col))
	}
	joinConds = append(joinConds, "p.sort_val = g.sort_val")

	// JOIN 回原表 di：拿到那条代表记录的明细
	finalSQL := fmt.Sprintf("SELECT %s, %s FROM (%s) g JOIN (%s) p ON %s JOIN %s di ON di.f_id = p.max_id",
		selectTotal, selectCols, gSQL, pSQL, strings.Join(joinConds, " AND "), DAGINSTANCE_TABLENAME)

	finalArgs := append([]interface{}{}, gArgs...)
	finalArgs = append(finalArgs, pArgs...)

	return finalSQL, finalArgs, nil
}

func buildKeywordLike(val interface{}) (string, bool) {
	if val == nil {
		return "", false
	}
	switch v := val.(type) {
	case bson.M:
		if pattern, ok := v["$regex"]; ok {
			return regexToLike(fmt.Sprintf("%v", pattern)), true
		}
	case map[string]interface{}:
		if pattern, ok := v["$regex"]; ok {
			return regexToLike(fmt.Sprintf("%v", pattern)), true
		}
	case primitive.Regex:
		return regexToLike(v.Pattern), true
	}
	return fmt.Sprintf("%v", val) + "%", true
}

func BuildDagVars(dag *entity.Dag) []*DagVarModel {
	var vars []*DagVarModel
	for k, v := range dag.Vars {
		id, _ := utils.GetUniqueID()
		dagID, _ := strconv.ParseUint(dag.ID, 10, 64)
		vars = append(vars, &DagVarModel{
			ID:           id,
			DagID:        dagID,
			VarName:      k,
			DefaultValue: v.DefaultValue,
			VarType:      "string",
			Description:  v.Desc,
		})
	}

	return vars
}

func BuildDagStepIndex(dag *entity.Dag) []*DagStepModel {
	if dag == nil {
		return nil
	}
	dagID, _ := strconv.ParseUint(dag.ID, 10, 64)
	rows := []*DagStepModel{}

	addRow := func(operator, sourceID string, hasDatasource bool) {
		id, _ := utils.GetUniqueID()
		rows = append(rows, &DagStepModel{
			ID:            id,
			DagID:         dagID,
			Operator:      operator,
			SourceID:      sourceID,
			HasDatasource: hasDatasource,
		})
	}

	var walkSteps func(steps []entity.Step)
	walkSteps = func(steps []entity.Step) {
		for _, step := range steps {
			hasDatasource := step.DataSource != nil
			if step.Operator != "" {
				addRow(step.Operator, "", hasDatasource)
			} else if hasDatasource {
				addRow("", "", true)
			}
			if len(step.Steps) > 0 {
				walkSteps(step.Steps)
			}
			if len(step.Branches) > 0 {
				for _, br := range step.Branches {
					if len(br.Steps) > 0 {
						walkSteps(br.Steps)
					}
				}
			}
		}
	}

	walkSteps(dag.Steps)
	return rows
}

func BuildDagAccessorIndex(dag *entity.Dag) []*DagAccessorModel {
	if dag == nil {
		return nil
	}
	dagID, _ := strconv.ParseUint(dag.ID, 10, 64)
	rows := []*DagAccessorModel{}
	for _, accessor := range dag.Accessors {
		if accessor.ID == "" {
			continue
		}
		id, _ := utils.GetUniqueID()
		rows = append(rows, &DagAccessorModel{
			ID:         id,
			DagID:      dagID,
			AccessorID: accessor.ID,
		})
	}
	return rows
}

func BuildDagIndexSubquery(input *mod.ListDagInput) (string, []interface{}) {
	if input == nil {
		return "", nil
	}

	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if input.Scope == "all" {
		unionParts := make([]string, 0, 2)
		if len(input.Accessors) > 0 {
			sub := "SELECT f_dag_id FROM t_flow_dag_accessor WHERE f_accessor_id IN ?"
			args = append(args, input.Accessors)
			if len(input.TriggerExclude) > 0 {
				sub += " AND NOT EXISTS (SELECT 1 FROM t_flow_dag_step ds WHERE ds.f_dag_id = t_flow_dag.f_id AND ds.f_operator IN ?)"
				args = append(args, input.TriggerExclude)
			}
			unionParts = append(unionParts, sub)
		}
		if input.UserID != "" {
			unionParts = append(unionParts, "SELECT f_id FROM t_flow_dag WHERE f_user_id = ?")
			args = append(args, input.UserID)
		}
		if len(unionParts) > 0 {
			conds = append(conds, fmt.Sprintf("f_id IN (%s)", strings.Join(unionParts, " UNION ALL ")))
		}
		conds = append(conds, "NOT EXISTS  (SELECT 1 FROM t_flow_dag_step ds WHERE ds.f_dag_id = t_flow_dag.f_id AND ds.f_has_datasource = 1)")
		return strings.Join(conds, " AND "), args
	}

	if input.Accessors != nil && input.UserID == "" {
		conds = append(conds, "f_id IN (SELECT f_dag_id FROM t_flow_dag_accessor WHERE f_accessor_id IN ?)")
		args = append(args, input.Accessors)
	}

	return strings.Join(conds, " AND "), args
}

func buildDagInstanceKeywordRows(dagInsID uint64, keywords []string) []DagInstanceKeywordModel {
	if len(keywords) == 0 {
		return nil
	}
	rows := make([]DagInstanceKeywordModel, 0, len(keywords))
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		id, _ := utils.GetUniqueID()
		rows = append(rows, DagInstanceKeywordModel{ID: id, DagInsID: dagInsID, Keyword: keyword})
	}
	return rows
}
