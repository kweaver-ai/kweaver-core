package ormhelper

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// InsertBuilder INSERT语句构建器
type InsertBuilder struct {
	db             *DB
	table          string
	columns        []string
	values         [][]interface{}
	data           map[string]interface{}
	onDuplicateKey map[string]interface{}
	ignore         bool
}

// Into 指定表名
func (i *InsertBuilder) Into(table string) *InsertBuilder {
	i.table = fmt.Sprintf("`%s`.`%s`", i.db.dbName, table)
	return i
}

// Values 指定单行数据
func (i *InsertBuilder) Values(data map[string]interface{}) *InsertBuilder {
	i.data = data
	return i
}

// BatchValues 指定多行数据
func (i *InsertBuilder) BatchValues(columns []string, values [][]interface{}) *InsertBuilder {
	i.columns = columns
	i.values = values
	return i
}

// OnDuplicateKeyUpdate 重复键更新
func (i *InsertBuilder) OnDuplicateKeyUpdate(data map[string]interface{}) *InsertBuilder {
	i.onDuplicateKey = data
	return i
}

// Ignore INSERT IGNORE
func (i *InsertBuilder) Ignore() *InsertBuilder {
	i.ignore = true
	return i
}

// Build 构建SQL语句
func (i *InsertBuilder) Build() (query string, args []interface{}) {
	insertType := "INSERT"
	if i.ignore {
		insertType = "INSERT IGNORE"
	}

	if i.data != nil {
		// 单行插入
		fields := make([]string, 0, len(i.data))
		placeholders := make([]string, 0, len(i.data))
		args := make([]interface{}, 0, len(i.data))

		for field, value := range i.data {
			fields = append(fields, field)
			placeholders = append(placeholders, "?")
			args = append(args, value)
		}

		query = fmt.Sprintf("%s INTO %s (%s) VALUES (%s)",
			insertType,
			i.table,
			strings.Join(fields, ", "),
			strings.Join(placeholders, ", "))

		// ON DUPLICATE KEY UPDATE
		if i.onDuplicateKey != nil {
			updates := make([]string, 0, len(i.onDuplicateKey))
			for field, value := range i.onDuplicateKey {
				updates = append(updates, field+" = ?")
				args = append(args, value)
			}
			query += " ON DUPLICATE KEY UPDATE " + strings.Join(updates, ", ")
		}

		return query, args
	} else if len(i.values) > 0 {
		// 批量插入
		placeholderGroups := make([]string, len(i.values))
		args := make([]interface{}, 0, len(i.values)*len(i.columns))

		for idx, row := range i.values {
			placeholders := strings.Repeat("?,", len(row))
			placeholders = placeholders[:len(placeholders)-1]
			placeholderGroups[idx] = "(" + placeholders + ")"
			args = append(args, row...)
		}

		query = fmt.Sprintf("%s INTO %s (%s) VALUES %s",
			insertType,
			i.table,
			strings.Join(i.columns, ", "),
			strings.Join(placeholderGroups, ", "))

		return query, args
	}

	return "", nil
}

// Execute 执行插入
func (i *InsertBuilder) Execute(ctx context.Context) (sql.Result, error) {
	query, args := i.Build()
	return i.db.executor.ExecContext(ctx, query, args...)
}

// ExecuteAndReturnID 执行插入并返回最后插入的ID
// @deprecated 废弃： 人大建仓、达梦、TiDB 等数据库不支持LastInsertId使用
func (i *InsertBuilder) ExecuteAndReturnID(ctx context.Context) (int64, error) {
	result, err := i.Execute(ctx)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ExecuteAndReturnAffected 执行插入并返回影响的行数
func (i *InsertBuilder) ExecuteAndReturnAffected(ctx context.Context) (int64, error) {
	result, err := i.Execute(ctx)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
