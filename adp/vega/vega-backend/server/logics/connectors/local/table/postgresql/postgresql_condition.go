// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package postgresql

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

var Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)

func (c *PostgresqlConnector) ConvertFilterCondition(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationAnd:
		return c.ConvertFilterConditionAnd(ctx, condition, fieldsMap)

	case filter_condition.OperationOr:
		return c.ConvertFilterConditionOr(ctx, condition, fieldsMap)

	default:
		return c.ConvertFilterConditionWithOpr(ctx, condition, fieldsMap)
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionAnd(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	condAnd, ok := condition.(*filter_condition.AndCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.AndCond")
	}

	convertedConds := sq.And{}
	for _, subCond := range condAnd.SubConds {
		convertedCond, err := c.ConvertFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		convertedConds = append(convertedConds, convertedCond)
	}

	return convertedConds, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionOr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	condOr, ok := condition.(*filter_condition.OrCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OrCond")
	}

	convertedConds := sq.Or{}
	for _, subCond := range condOr.SubConds {
		convertedCond, err := c.ConvertFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		convertedConds = append(convertedConds, convertedCond)
	}

	return convertedConds, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionWithOpr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationEqual, filter_condition.OperationEqual2:
		return c.ConvertFilterConditionEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEqual, filter_condition.OperationNotEqual2:
		return c.ConvertFilterConditionNotEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationGt, filter_condition.OperationGt2:
		return c.ConvertFilterConditionGt(ctx, condition, fieldsMap)
	case filter_condition.OperationGte, filter_condition.OperationGte2:
		return c.ConvertFilterConditionGte(ctx, condition, fieldsMap)
	case filter_condition.OperationLt, filter_condition.OperationLt2:
		return c.ConvertFilterConditionLt(ctx, condition, fieldsMap)
	case filter_condition.OperationLte, filter_condition.OperationLte2:
		return c.ConvertFilterConditionLte(ctx, condition, fieldsMap)
	case filter_condition.OperationIn:
		return c.ConvertFilterConditionIn(ctx, condition, fieldsMap)
	case filter_condition.OperationNotIn:
		return c.ConvertFilterConditionNotIn(ctx, condition, fieldsMap)
	case filter_condition.OperationLike:
		return c.ConvertFilterConditionLike(ctx, condition, fieldsMap)
	case filter_condition.OperationNotLike:
		return c.ConvertFilterConditionNotLike(ctx, condition, fieldsMap)
	case filter_condition.OperationContain:
		return c.ConvertFilterConditionContain(ctx, condition, fieldsMap)
	case filter_condition.OperationNotContain:
		return c.ConvertFilterConditionNotContain(ctx, condition, fieldsMap)
	case filter_condition.OperationRange:
		return c.ConvertFilterConditionRange(ctx, condition, fieldsMap)
	case filter_condition.OperationOutRange:
		return c.ConvertFilterConditionOutRange(ctx, condition, fieldsMap)
	case filter_condition.OperationNull:
		return c.ConvertFilterConditionNull(ctx, condition, fieldsMap)
	case filter_condition.OperationNotNull:
		return c.ConvertFilterConditionNotNull(ctx, condition, fieldsMap)
	case filter_condition.OperationEmpty:
		return c.ConvertFilterConditionEmpty(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEmpty:
		return c.ConvertFilterConditionNotEmpty(ctx, condition, fieldsMap)
	case filter_condition.OperationPrefix:
		return c.ConvertFilterConditionPrefix(ctx, condition, fieldsMap)
	case filter_condition.OperationNotPrefix:
		return c.ConvertFilterConditionNotPrefix(ctx, condition, fieldsMap)
	case filter_condition.OperationBetween:
		return c.ConvertFilterConditionBetween(ctx, condition, fieldsMap)
	case filter_condition.OperationExist:
		return c.ConvertFilterConditionExist(ctx, condition, fieldsMap)
	case filter_condition.OperationNotExist:
		return c.ConvertFilterConditionNotExist(ctx, condition, fieldsMap)
	case filter_condition.OperationRegex:
		return c.ConvertFilterConditionRegex(ctx, condition, fieldsMap)
	case filter_condition.OperationTrue:
		return c.ConvertFilterConditionTrue(ctx, condition, fieldsMap)
	case filter_condition.OperationFalse:
		return c.ConvertFilterConditionFalse(ctx, condition, fieldsMap)
	case filter_condition.OperationBefore:
		return c.ConvertFilterConditionBefore(ctx, condition, fieldsMap)
	case filter_condition.OperationCurrent:
		return c.ConvertFilterConditionCurrent(ctx, condition, fieldsMap)
	default:
		return nil, fmt.Errorf("operation %s is not supported", condition.GetOperation())
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionEqual(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.EqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EqualCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(quoteColumnName(cond.Lfield.OriginalName) + " = " + quoteColumnName(cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionNotEqual(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotEqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEqualCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(quoteColumnName(cond.Lfield.OriginalName) + " <> " + quoteColumnName(cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionGt(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.GtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.Gt{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(quoteColumnName(cond.Lfield.OriginalName) + " > " + quoteColumnName(cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionGte(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.GteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.GtOrEq{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(quoteColumnName(cond.Lfield.OriginalName) + " >= " + quoteColumnName(cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionLt(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.LtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.Lt{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(quoteColumnName(cond.Lfield.OriginalName) + " < " + quoteColumnName(cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionLte(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.LteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.LtOrEq{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(quoteColumnName(cond.Lfield.OriginalName) + " <= " + quoteColumnName(cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func (c *PostgresqlConnector) ConvertFilterConditionIn(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.InCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.InCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNotIn(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotInCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotInCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): cond.Value}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionLike(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.LikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := "%" + Special.Replace(cond.Value) + "%"
	return sq.Like{quoteColumnName(cond.Lfield.OriginalName): vStr}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNotLike(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotLikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotLikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := "%" + Special.Replace(cond.Value) + "%"
	return sq.NotLike{quoteColumnName(cond.Lfield.OriginalName): vStr}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionContain(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.ContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	exprs := make(sq.And, len(values))
	col := quoteColumnName(cond.Lfield.OriginalName)
	for i, v := range values {
		exprs[i] = sq.Expr("? = ANY(string_to_array("+col+"::text, ','))", fmt.Sprint(v))
	}
	return exprs, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNotContain(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	exprs := make(sq.Or, len(values))
	col := quoteColumnName(cond.Lfield.OriginalName)
	for i, v := range values {
		exprs[i] = sq.Expr("NOT (? = ANY(string_to_array("+col+"::text, ',')))", fmt.Sprint(v))
	}
	return exprs, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionRange(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.RangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("range condition requires exactly 2 values")
	}

	return sq.And{
		sq.GtOrEq{quoteColumnName(cond.Lfield.OriginalName): values[0]},
		sq.LtOrEq{quoteColumnName(cond.Lfield.OriginalName): values[1]},
	}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionOutRange(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.OutRangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OutRangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [out_range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("out_range condition requires exactly 2 values")
	}

	return sq.Or{
		sq.Lt{quoteColumnName(cond.Lfield.OriginalName): values[0]},
		sq.Gt{quoteColumnName(cond.Lfield.OriginalName): values[1]},
	}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNull(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NullCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNotNull(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotNullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotNullCond")
	}

	return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionEmpty(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.EmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EmptyCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): ""}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNotEmpty(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotEmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEmptyCond")
	}

	return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): ""}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionPrefix(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.PrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.PrefixCond")
	}

	vStr := Special.Replace(cond.Value) + "%"
	return sq.Like{quoteColumnName(cond.Lfield.OriginalName): vStr}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNotPrefix(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotPrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotPrefixCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_prefix] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := Special.Replace(cond.Value) + "%"
	return sq.NotLike{quoteColumnName(cond.Lfield.OriginalName): vStr}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionBetween(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.BetweenCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BetweenCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("between condition requires exactly 2 values")
	}

	// 检查字段是否为时间类型，如果是则转换long类型值为时间戳
	fieldType := cond.Lfield.Type
	isDateType := interfaces.DataType_IsDate(fieldType)

	if isDateType {
		// 时间类型字段，使用原始SQL表达式构建BETWEEN条件
		var lowerTs, upperTs string
		if v0, ok := values[0].(float64); ok {
			lowerTs = fmt.Sprintf("%d", int64(v0))
		} else if v0, ok := values[0].(int64); ok {
			lowerTs = fmt.Sprintf("%d", v0)
		} else if v0, ok := values[0].(int); ok {
			lowerTs = fmt.Sprintf("%d", int64(v0))
		} else {
			lowerTs = fmt.Sprintf("%v", values[0])
		}

		if v1, ok := values[1].(float64); ok {
			upperTs = fmt.Sprintf("%d", int64(v1))
		} else if v1, ok := values[1].(int64); ok {
			upperTs = fmt.Sprintf("%d", v1)
		} else if v1, ok := values[1].(int); ok {
			upperTs = fmt.Sprintf("%d", int64(v1))
		} else {
			upperTs = fmt.Sprintf("%v", values[1])
		}

		// 使用SqlExpr构建完整的WHERE条件表达式
		return sq.Expr(
			quoteColumnName(cond.Lfield.OriginalName)+" >= to_timestamp(?/1000000) AND "+
				quoteColumnName(cond.Lfield.OriginalName)+" <= to_timestamp(?/1000000)",
			lowerTs, upperTs,
		), nil
	}

	// 非时间类型字段，直接使用参数化查询
	return sq.And{
		sq.GtOrEq{quoteColumnName(cond.Lfield.OriginalName): values[0]},
		sq.LtOrEq{quoteColumnName(cond.Lfield.OriginalName): values[1]},
	}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionExist(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.ExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ExistCond")
	}

	return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionNotExist(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotExistCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionRegex(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.RegexCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RegexCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return sq.Expr(quoteColumnName(cond.Lfield.OriginalName)+" ~ ?", cond.Value), nil
}

func (c *PostgresqlConnector) ConvertFilterConditionTrue(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.TrueCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.TrueCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): true}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionFalse(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.FalseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.FalseCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): false}, nil
}

func (c *PostgresqlConnector) ConvertFilterConditionBefore(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.BeforeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BeforeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("before condition requires exactly 2 values")
	}

	var n int64
	switch v := values[0].(type) {
	case float64:
		n = int64(v)
	case int:
		n = int64(v)
	case int64:
		n = v
	default:
		return nil, fmt.Errorf("condition [before] interval value should be a number")
	}
	unit, ok := values[1].(string)
	if !ok {
		return nil, fmt.Errorf("condition [before] unit value should be a string")
	}
	pgUnit, err := pgIntervalUnit(strings.TrimSpace(unit))
	if err != nil {
		return nil, err
	}
	col := quoteColumnName(cond.Lfield.OriginalName)
	return sq.Expr(fmt.Sprintf("%s < NOW() - (?::bigint * INTERVAL '1 %s')", col, pgUnit), n), nil
}

func (c *PostgresqlConnector) ConvertFilterConditionCurrent(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.CurrentCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.CurrentCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	col := quoteColumnName(cond.Lfield.OriginalName)
	var trunc string
	switch cond.Value {
	case filter_condition.CurrentYear:
		trunc = "year"
	case filter_condition.CurrentMonth:
		trunc = "month"
	case filter_condition.CurrentWeek:
		trunc = "week"
	case filter_condition.CurrentDay:
		trunc = "day"
	case filter_condition.CurrentHour:
		trunc = "hour"
	case filter_condition.CurrentMinute:
		trunc = "minute"
	default:
		return nil, fmt.Errorf("condition [current] unsupported format: %s", cond.Value)
	}

	return sq.Expr(fmt.Sprintf("date_trunc('%s', %s::timestamptz) = date_trunc('%s', CURRENT_TIMESTAMP)", trunc, col, trunc)), nil
}

// pgIntervalUnit 将 MySQL 风格 INTERVAL 单位映射为 PostgreSQL interval 乘法用的英文单数单位名。
func pgIntervalUnit(mysqlStyle string) (string, error) {
	u := strings.ToUpper(strings.TrimSpace(mysqlStyle))
	switch u {
	case "YEAR", "YEARS":
		return "year", nil
	case "MONTH", "MONTHS":
		return "month", nil
	case "DAY", "DAYS":
		return "day", nil
	case "HOUR", "HOURS":
		return "hour", nil
	case "MINUTE", "MINUTES":
		return "minute", nil
	case "SECOND", "SECONDS":
		return "second", nil
	default:
		return "", fmt.Errorf("unsupported interval unit for postgresql: %s", mysqlStyle)
	}
}
