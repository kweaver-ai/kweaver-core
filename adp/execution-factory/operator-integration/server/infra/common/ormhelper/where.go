package ormhelper

import (
	"fmt"
	"strings"
)

// WhereBuilder WHERE条件构建器
type WhereBuilder struct {
	conditions []string
	args       []interface{}
	operator   string // AND 或 OR
}

// NewWhere 创建WHERE构建器
func NewWhere() *WhereBuilder {
	return &WhereBuilder{
		operator: "AND",
	}
}

// Condition 添加条件
func (w *WhereBuilder) Condition(field, op string, value interface{}) *WhereBuilder {
	w.conditions = append(w.conditions, fmt.Sprintf("%s %s ?", field, op))
	w.args = append(w.args, value)
	return w
}

// Eq 等于条件
func (w *WhereBuilder) Eq(field string, value interface{}) *WhereBuilder {
	return w.Condition(field, "=", value)
}

// Ne 不等于条件
func (w *WhereBuilder) Ne(field string, value interface{}) *WhereBuilder {
	return w.Condition(field, "!=", value)
}

// Gt 大于条件
func (w *WhereBuilder) Gt(field string, value interface{}) *WhereBuilder {
	return w.Condition(field, ">", value)
}

// Gte 大于等于条件
func (w *WhereBuilder) Gte(field string, value interface{}) *WhereBuilder {
	return w.Condition(field, ">=", value)
}

// Lt 小于条件
func (w *WhereBuilder) Lt(field string, value interface{}) *WhereBuilder {
	return w.Condition(field, "<", value)
}

// Lte 小于等于条件
func (w *WhereBuilder) Lte(field string, value interface{}) *WhereBuilder {
	return w.Condition(field, "<=", value)
}

// In IN条件
func (w *WhereBuilder) In(field string, values ...interface{}) *WhereBuilder {
	if len(values) == 0 {
		return w
	}

	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1]

	w.conditions = append(w.conditions, fmt.Sprintf("%s IN (%s)", field, placeholders))
	w.args = append(w.args, values...)
	return w
}

// NotIn NOT IN条件
func (w *WhereBuilder) NotIn(field string, values ...interface{}) *WhereBuilder {
	if len(values) == 0 {
		return w
	}

	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1]

	w.conditions = append(w.conditions, fmt.Sprintf("%s NOT IN (%s)", field, placeholders))
	w.args = append(w.args, values...)
	return w
}

// Like LIKE条件
func (w *WhereBuilder) Like(field, pattern string) *WhereBuilder {
	return w.Condition(field, "LIKE", pattern)
}

// NotLike NOT LIKE条件
func (w *WhereBuilder) NotLike(field, pattern string) *WhereBuilder {
	return w.Condition(field, "NOT LIKE", pattern)
}

// Between BETWEEN条件
func (w *WhereBuilder) Between(field string, start, end interface{}) *WhereBuilder {
	w.conditions = append(w.conditions, fmt.Sprintf("%s BETWEEN ? AND ?", field))
	w.args = append(w.args, start, end)
	return w
}

// NotBetween NOT BETWEEN条件
func (w *WhereBuilder) NotBetween(field string, start, end interface{}) *WhereBuilder {
	w.conditions = append(w.conditions, fmt.Sprintf("%s NOT BETWEEN ? AND ?", field))
	w.args = append(w.args, start, end)
	return w
}

// IsNull IS NULL条件
func (w *WhereBuilder) IsNull(field string) *WhereBuilder {
	w.conditions = append(w.conditions, field+" IS NULL")
	return w
}

// IsNotNull IS NOT NULL条件
func (w *WhereBuilder) IsNotNull(field string) *WhereBuilder {
	w.conditions = append(w.conditions, field+" IS NOT NULL")
	return w
}

// And 添加AND条件组
func (w *WhereBuilder) And(fn func(*WhereBuilder)) *WhereBuilder {
	subWhere := NewWhere()
	fn(subWhere)

	if len(subWhere.conditions) > 0 {
		subClause, subArgs := subWhere.Build()
		w.conditions = append(w.conditions, "("+subClause+")")
		w.args = append(w.args, subArgs...)
	}
	return w
}

// Or 添加OR条件组
func (w *WhereBuilder) Or(fn func(*WhereBuilder)) *WhereBuilder {
	subWhere := &WhereBuilder{operator: "OR"}
	fn(subWhere)

	if len(subWhere.conditions) > 0 {
		subClause, subArgs := subWhere.Build()
		w.conditions = append(w.conditions, "("+subClause+")")
		w.args = append(w.args, subArgs...)
	}
	return w
}

// Cursor 游标条件
func (w *WhereBuilder) Cursor(field string, value interface{}, isAsc bool) *WhereBuilder {
	if isAsc {
		w.Gte(field, value)
	} else {
		w.Lt(field, value)
	}
	return w
}

// Raw 原始条件
func (w *WhereBuilder) Raw(condition string, args ...interface{}) *WhereBuilder {
	w.conditions = append(w.conditions, condition)
	w.args = append(w.args, args...)
	return w
}

// Build 构建WHERE子句
func (w *WhereBuilder) Build() (query string, args []interface{}) {
	if len(w.conditions) == 0 {
		return "", nil
	}
	query = strings.Join(w.conditions, " "+w.operator+" ")
	args = w.args
	return
}
