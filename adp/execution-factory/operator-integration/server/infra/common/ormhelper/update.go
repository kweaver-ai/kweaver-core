package ormhelper

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// UpdateBuilder UPDATE语句构建器
type UpdateBuilder struct {
	db    *DB
	table string
	sets  map[string]interface{}
	where *WhereBuilder
	limit int
}

// rawExpression 原始表达式
type rawExpression struct {
	expr string
}

// Set 设置字段值
func (u *UpdateBuilder) Set(field string, value interface{}) *UpdateBuilder {
	if u.sets == nil {
		u.sets = make(map[string]interface{})
	}
	u.sets[field] = value
	return u
}

// SetData 批量设置字段值
func (u *UpdateBuilder) SetData(data map[string]interface{}) *UpdateBuilder {
	if u.sets == nil {
		u.sets = make(map[string]interface{})
	}
	for field, value := range data {
		u.sets[field] = value
	}
	return u
}

// SetRaw 设置原始SQL表达式
func (u *UpdateBuilder) SetRaw(field, expr string) *UpdateBuilder {
	if u.sets == nil {
		u.sets = make(map[string]interface{})
	}
	// 使用特殊标记表示这是原始表达式
	u.sets[field] = &rawExpression{expr: expr}
	return u
}

// Increment 字段自增
func (u *UpdateBuilder) Increment(field string, value interface{}) *UpdateBuilder {
	return u.SetRaw(field, fmt.Sprintf("%s + %v", field, value))
}

// Decrement 字段自减
func (u *UpdateBuilder) Decrement(field string, value interface{}) *UpdateBuilder {
	return u.SetRaw(field, fmt.Sprintf("%s - %v", field, value))
}

// Where 添加WHERE条件
func (u *UpdateBuilder) Where(field, op string, value interface{}) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.Condition(field, op, value)
	return u
}

// WhereEq 等于条件的简写
func (u *UpdateBuilder) WhereEq(field string, value interface{}) *UpdateBuilder {
	return u.Where(field, "=", value)
}

// WhereNe 不等于条件
func (u *UpdateBuilder) WhereNe(field string, value interface{}) *UpdateBuilder {
	return u.Where(field, "!=", value)
}

// WhereIn IN条件
func (u *UpdateBuilder) WhereIn(field string, values ...interface{}) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.In(field, values...)
	return u
}

// WhereNotIn NOT IN条件
func (u *UpdateBuilder) WhereNotIn(field string, values ...interface{}) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.NotIn(field, values...)
	return u
}

// WhereLike LIKE条件
func (u *UpdateBuilder) WhereLike(field, pattern string) *UpdateBuilder {
	return u.Where(field, "LIKE", pattern)
}

// WhereBetween BETWEEN条件
func (u *UpdateBuilder) WhereBetween(field string, start, end interface{}) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.Between(field, start, end)
	return u
}

// WhereNull IS NULL条件
func (u *UpdateBuilder) WhereNull(field string) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.IsNull(field)
	return u
}

// WhereNotNull IS NOT NULL条件
func (u *UpdateBuilder) WhereNotNull(field string) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.IsNotNull(field)
	return u
}

// And 开始AND条件组
func (u *UpdateBuilder) And(fn func(*WhereBuilder)) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.And(fn)
	return u
}

// Or 开始OR条件组
func (u *UpdateBuilder) Or(fn func(*WhereBuilder)) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.Or(fn)
	return u
}

// WhereRaw 原始WHERE条件
func (u *UpdateBuilder) WhereRaw(condition string, args ...interface{}) *UpdateBuilder {
	if u.where == nil {
		u.where = NewWhere()
	}
	u.where.Raw(condition, args...)
	return u
}

// Limit 限制更新数量
func (u *UpdateBuilder) Limit(limit int) *UpdateBuilder {
	u.limit = limit
	return u
}

// Build 构建SQL语句
func (u *UpdateBuilder) Build() (query string, args []interface{}) {
	sets := make([]string, 0, len(u.sets))
	args = make([]interface{}, 0, len(u.sets))

	for field, value := range u.sets {
		if raw, ok := value.(*rawExpression); ok {
			// 原始表达式，不需要占位符
			sets = append(sets, field+" = "+raw.expr)
		} else {
			sets = append(sets, field+" = ?")
			args = append(args, value)
		}
	}

	query = fmt.Sprintf("UPDATE %s SET %s", u.table, strings.Join(sets, ", "))

	// WHERE条件
	if u.where != nil {
		whereClause, whereArgs := u.where.Build()
		if whereClause != "" {
			query += " WHERE " + whereClause
			args = append(args, whereArgs...)
		}
	}

	// LIMIT
	if u.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", u.limit)
	}

	return query, args
}

// Execute 执行更新
func (u *UpdateBuilder) Execute(ctx context.Context) (sql.Result, error) {
	query, args := u.Build()
	return u.db.executor.ExecContext(ctx, query, args...)
}

// ExecuteAndReturnAffected 执行更新并返回影响的行数
func (u *UpdateBuilder) ExecuteAndReturnAffected(ctx context.Context) (int64, error) {
	result, err := u.Execute(ctx)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
