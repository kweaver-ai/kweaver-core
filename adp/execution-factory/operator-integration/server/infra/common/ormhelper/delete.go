// Package ormhelper provides a simple and efficient way to interact with databases in Go applications.
package ormhelper

import (
	"context"
	"database/sql"
	"fmt"
)

// DeleteBuilder DELETE语句构建器
type DeleteBuilder struct {
	db    *DB
	table string
	where *WhereBuilder
	limit int
}

// From 指定表名
func (d *DeleteBuilder) From(table string) *DeleteBuilder {
	d.table = fmt.Sprintf("`%s`.`%s`", d.db.dbName, table)
	return d
}

// Where 添加WHERE条件
func (d *DeleteBuilder) Where(field, op string, value interface{}) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.Condition(field, op, value)
	return d
}

// WhereEq 等于条件的简写
func (d *DeleteBuilder) WhereEq(field string, value interface{}) *DeleteBuilder {
	return d.Where(field, "=", value)
}

// WhereNe 不等于条件
func (d *DeleteBuilder) WhereNe(field string, value interface{}) *DeleteBuilder {
	return d.Where(field, "!=", value)
}

// WhereGt 大于条件
func (d *DeleteBuilder) WhereGt(field string, value interface{}) *DeleteBuilder {
	return d.Where(field, ">", value)
}

// WhereGte 大于等于条件
func (d *DeleteBuilder) WhereGte(field string, value interface{}) *DeleteBuilder {
	return d.Where(field, ">=", value)
}

// WhereLt 小于条件
func (d *DeleteBuilder) WhereLt(field string, value interface{}) *DeleteBuilder {
	return d.Where(field, "<", value)
}

// WhereLte 小于等于条件
func (d *DeleteBuilder) WhereLte(field string, value interface{}) *DeleteBuilder {
	return d.Where(field, "<=", value)
}

// WhereIn IN条件
func (d *DeleteBuilder) WhereIn(field string, values ...interface{}) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.In(field, values...)
	return d
}

// WhereNotIn NOT IN条件
func (d *DeleteBuilder) WhereNotIn(field string, values ...interface{}) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.NotIn(field, values...)
	return d
}

// WhereLike LIKE条件
func (d *DeleteBuilder) WhereLike(field, pattern string) *DeleteBuilder {
	return d.Where(field, "LIKE", pattern)
}

// WhereNotLike NOT LIKE条件
func (d *DeleteBuilder) WhereNotLike(field, pattern string) *DeleteBuilder {
	return d.Where(field, "NOT LIKE", pattern)
}

// WhereBetween BETWEEN条件
func (d *DeleteBuilder) WhereBetween(field string, start, end interface{}) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.Between(field, start, end)
	return d
}

// WhereNotBetween NOT BETWEEN条件
func (d *DeleteBuilder) WhereNotBetween(field string, start, end interface{}) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.NotBetween(field, start, end)
	return d
}

// WhereNull IS NULL条件
func (d *DeleteBuilder) WhereNull(field string) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.IsNull(field)
	return d
}

// WhereNotNull IS NOT NULL条件
func (d *DeleteBuilder) WhereNotNull(field string) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.IsNotNull(field)
	return d
}

// And 开始AND条件组
func (d *DeleteBuilder) And(fn func(*WhereBuilder)) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.And(fn)
	return d
}

// Or 开始OR条件组
func (d *DeleteBuilder) Or(fn func(*WhereBuilder)) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.Or(fn)
	return d
}

// WhereRaw 原始WHERE条件
func (d *DeleteBuilder) WhereRaw(condition string, args ...interface{}) *DeleteBuilder {
	if d.where == nil {
		d.where = NewWhere()
	}
	d.where.Raw(condition, args...)
	return d
}

// Limit 限制删除数量
func (d *DeleteBuilder) Limit(limit int) *DeleteBuilder {
	d.limit = limit
	return d
}

// Build 构建SQL语句
func (d *DeleteBuilder) Build() (query string, args []interface{}) {
	query = fmt.Sprintf("DELETE FROM %s", d.table)

	// WHERE条件
	if d.where != nil {
		whereClause, whereArgs := d.where.Build()
		if whereClause != "" {
			query += " WHERE " + whereClause
			args = append(args, whereArgs...)
		}
	}

	// LIMIT
	if d.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", d.limit)
	}

	return query, args
}

// Execute 执行删除
func (d *DeleteBuilder) Execute(ctx context.Context) (sql.Result, error) {
	query, args := d.Build()
	return d.db.executor.ExecContext(ctx, query, args...)
}

// ExecuteAndReturnAffected 执行删除并返回影响的行数
func (d *DeleteBuilder) ExecuteAndReturnAffected(ctx context.Context) (int64, error) {
	result, err := d.Execute(ctx)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
