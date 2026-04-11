// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package sql

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

var Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)

// quoteColumnName 将列名转为 SQL 标识符；支持 "alias.col" -> "`alias`.`col`"
func quoteColumnName(name string) string {
	if name == "" {
		return "``"
	}
	if idx := strings.Index(name, "."); idx >= 0 {
		alias := strings.TrimSpace(name[:idx])
		col := strings.TrimSpace(name[idx+1:])
		return "`" + strings.ReplaceAll(alias, "`", "``") + "`." + "`" + strings.ReplaceAll(col, "`", "``") + "`"
	}
	return "`" + strings.ReplaceAll(strings.TrimSpace(name), "`", "``") + "`"
}

func (g *logicViewSQLGenerator) ConvertFilterCondition(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationAnd:
		return g.ConvertFilterConditionAnd(ctx, condition, fieldsMap)

	case filter_condition.OperationOr:
		return g.ConvertFilterConditionOr(ctx, condition, fieldsMap)

	default:
		return g.ConvertFilterConditionWithOpr(ctx, condition, fieldsMap)
	}
}

func (g *logicViewSQLGenerator) ConvertFilterConditionAnd(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	condAnd, ok := condition.(*filter_condition.AndCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.AndCond")
	}

	convertedConds := sq.And{}
	for _, subCond := range condAnd.SubConds {
		convertedCond, err := g.ConvertFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		convertedConds = append(convertedConds, convertedCond)
	}

	return convertedConds, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionOr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	condOr, ok := condition.(*filter_condition.OrCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OrCond")
	}

	convertedConds := sq.Or{}
	for _, subCond := range condOr.SubConds {
		convertedCond, err := g.ConvertFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		convertedConds = append(convertedConds, convertedCond)
	}

	return convertedConds, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionWithOpr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationEqual, filter_condition.OperationEqual2:
		return g.ConvertFilterConditionEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEqual, filter_condition.OperationNotEqual2:
		return g.ConvertFilterConditionNotEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationGt, filter_condition.OperationGt2:
		return g.ConvertFilterConditionGt(ctx, condition, fieldsMap)
	case filter_condition.OperationGte, filter_condition.OperationGte2:
		return g.ConvertFilterConditionGte(ctx, condition, fieldsMap)
	case filter_condition.OperationLt, filter_condition.OperationLt2:
		return g.ConvertFilterConditionLt(ctx, condition, fieldsMap)
	case filter_condition.OperationLte, filter_condition.OperationLte2:
		return g.ConvertFilterConditionLte(ctx, condition, fieldsMap)
	case filter_condition.OperationIn:
		return g.ConvertFilterConditionIn(ctx, condition, fieldsMap)
	case filter_condition.OperationNotIn:
		return g.ConvertFilterConditionNotIn(ctx, condition, fieldsMap)
	case filter_condition.OperationLike:
		return g.ConvertFilterConditionLike(ctx, condition, fieldsMap)
	case filter_condition.OperationNotLike:
		return g.ConvertFilterConditionNotLike(ctx, condition, fieldsMap)
	case filter_condition.OperationContain:
		return g.ConvertFilterConditionContain(ctx, condition, fieldsMap)
	case filter_condition.OperationNotContain:
		return g.ConvertFilterConditionNotContain(ctx, condition, fieldsMap)
	case filter_condition.OperationRange:
		return g.ConvertFilterConditionRange(ctx, condition, fieldsMap)
	case filter_condition.OperationOutRange:
		return g.ConvertFilterConditionOutRange(ctx, condition, fieldsMap)
	case filter_condition.OperationNull:
		return g.ConvertFilterConditionNull(ctx, condition, fieldsMap)
	case filter_condition.OperationNotNull:
		return g.ConvertFilterConditionNotNull(ctx, condition, fieldsMap)
	case filter_condition.OperationEmpty:
		return g.ConvertFilterConditionEmpty(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEmpty:
		return g.ConvertFilterConditionNotEmpty(ctx, condition, fieldsMap)
	case filter_condition.OperationPrefix:
		return g.ConvertFilterConditionPrefix(ctx, condition, fieldsMap)
	case filter_condition.OperationNotPrefix:
		return g.ConvertFilterConditionNotPrefix(ctx, condition, fieldsMap)
	case filter_condition.OperationBetween:
		return g.ConvertFilterConditionBetween(ctx, condition, fieldsMap)
	case filter_condition.OperationExist:
		return g.ConvertFilterConditionExist(ctx, condition, fieldsMap)
	case filter_condition.OperationNotExist:
		return g.ConvertFilterConditionNotExist(ctx, condition, fieldsMap)
	case filter_condition.OperationRegex:
		return g.ConvertFilterConditionRegex(ctx, condition, fieldsMap)
	case filter_condition.OperationTrue:
		return g.ConvertFilterConditionTrue(ctx, condition, fieldsMap)
	case filter_condition.OperationFalse:
		return g.ConvertFilterConditionFalse(ctx, condition, fieldsMap)
	case filter_condition.OperationBefore:
		return g.ConvertFilterConditionBefore(ctx, condition, fieldsMap)
	case filter_condition.OperationCurrent:
		return g.ConvertFilterConditionCurrent(ctx, condition, fieldsMap)
	default:
		return nil, fmt.Errorf("operation %s is not supported", condition.GetOperation())
	}
}

func (g *logicViewSQLGenerator) ConvertFilterConditionEqual(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionNotEqual(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionGt(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionGte(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionLt(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionLte(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionIn(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionNotIn(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionLike(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionNotLike(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionContain(ctx context.Context, condition interfaces.FilterCondition,
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
	for i, v := range values {
		exprs[i] = sq.Expr("FIND_IN_SET(?, "+quoteColumnName(cond.Lfield.OriginalName)+") > 0", v)
	}
	return exprs, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionNotContain(ctx context.Context, condition interfaces.FilterCondition,
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
	for i, v := range values {
		exprs[i] = sq.Expr("FIND_IN_SET(?, "+quoteColumnName(cond.Lfield.OriginalName)+") = 0", v)
	}
	return exprs, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionRange(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionOutRange(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionNull(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NullCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionNotNull(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotNullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotNullCond")
	}

	return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionEmpty(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.EmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EmptyCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): ""}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionNotEmpty(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotEmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEmptyCond")
	}

	return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): ""}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionPrefix(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.PrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.PrefixCond")
	}

	vStr := Special.Replace(cond.Value) + "%"
	return sq.Like{quoteColumnName(cond.Lfield.OriginalName): vStr}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionNotPrefix(ctx context.Context, condition interfaces.FilterCondition,
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

func (g *logicViewSQLGenerator) ConvertFilterConditionBetween(ctx context.Context, condition interfaces.FilterCondition,
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

	return sq.And{
		sq.GtOrEq{quoteColumnName(cond.Lfield.OriginalName): values[0]},
		sq.LtOrEq{quoteColumnName(cond.Lfield.OriginalName): values[1]},
	}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionExist(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.ExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ExistCond")
	}

	return sq.NotEq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionNotExist(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotExistCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): nil}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionRegex(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.RegexCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RegexCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return sq.Expr(quoteColumnName(cond.Lfield.OriginalName)+" REGEXP ?", cond.Value), nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionTrue(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.TrueCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.TrueCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): true}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionFalse(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.FalseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.FalseCond")
	}

	return sq.Eq{quoteColumnName(cond.Lfield.OriginalName): false}, nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionBefore(ctx context.Context, condition interfaces.FilterCondition,
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

	interval, ok := values[0].(float64)
	if !ok {
		return nil, fmt.Errorf("condition [before] interval value should be a number")
	}
	unit, ok := values[1].(string)
	if !ok {
		return nil, fmt.Errorf("condition [before] unit value should be a string")
	}

	return sq.Expr(quoteColumnName(cond.Lfield.OriginalName)+" < DATE_SUB(NOW(), INTERVAL ? "+unit+")", int(interval)), nil
}

func (g *logicViewSQLGenerator) ConvertFilterConditionCurrent(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.CurrentCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.CurrentCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	var dateFormat string
	switch cond.Value {
	case filter_condition.CurrentYear:
		dateFormat = "%Y"
	case filter_condition.CurrentMonth:
		dateFormat = "%Y-%m"
	case filter_condition.CurrentWeek:
		dateFormat = "%Y-%u"
	case filter_condition.CurrentDay:
		dateFormat = "%Y-%m-%d"
	case filter_condition.CurrentHour:
		dateFormat = "%Y-%m-%d %H"
	case filter_condition.CurrentMinute:
		dateFormat = "%Y-%m-%d %H:%i"
	default:
		return nil, fmt.Errorf("condition [current] unsupported format: %s", cond.Value)
	}

	return sq.Expr("DATE_FORMAT(" + quoteColumnName(cond.Lfield.OriginalName) + ", '" + dateFormat + "') = DATE_FORMAT(NOW(), '" + dateFormat + "')"), nil
}
