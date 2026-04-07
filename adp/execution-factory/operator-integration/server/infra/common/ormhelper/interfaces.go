package ormhelper

import (
	"context"
	"database/sql"
)

// Executor 数据库执行器接口
// 兼容 *sql.DB, *sql.Tx, *sqlx.DB, *sqlx.Tx 等所有标准数据库接口
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Scanner 结果扫描器接口
type Scanner interface {
	ScanOne(row *sql.Row, dest interface{}) error
	ScanOneWithColumns(row *sql.Row, dest interface{}, columns []string) error
	ScanMany(rows *sql.Rows, dest interface{}) error
}
