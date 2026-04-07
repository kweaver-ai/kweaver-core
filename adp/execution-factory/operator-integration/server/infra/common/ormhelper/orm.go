package ormhelper

import (
	"context"
	"database/sql"
	"fmt"
)

// DB ORM核心类
type DB struct {
	executor Executor
	scanner  Scanner
	dbName   string
	// 为事务支持添加原始执行器引用
	originalExecutor Executor
}

// New 创建ORM实例
func New(executor Executor, dbName string) *DB {
	return &DB{
		executor:         executor,
		scanner:          NewScanner(),
		dbName:           dbName,
		originalExecutor: executor,
	}
}

// WithTx 使用事务创建新的ORM实例（推荐方式）
func (db *DB) WithTx(tx *sql.Tx) *DB {
	return &DB{
		executor:         tx,
		scanner:          db.scanner,
		dbName:           db.dbName,
		originalExecutor: db.originalExecutor,
	}
}

// SetExecutor 设置执行器（兼容现有代码模式）
func (db *DB) SetExecutor(ctx context.Context, tx *sql.Tx) *DB {
	if tx != nil {
		return db.WithTx(tx)
	}
	return db
}

// WithTxIfProvided 便民方法：根据tx参数自动选择执行器
func (db *DB) WithTxIfProvided(tx *sql.Tx) *DB {
	if tx != nil {
		return db.WithTx(tx)
	}
	return db
}

// Select SELECT语句入口
func (db *DB) Select(columns ...string) *SelectBuilder {
	return &SelectBuilder{
		db:      db,
		columns: columns,
	}
}

// Insert INSERT语句入口
func (db *DB) Insert() *InsertBuilder {
	return &InsertBuilder{
		db: db,
	}
}

// Update UPDATE语句入口
func (db *DB) Update(table string) *UpdateBuilder {
	fullTableName := fmt.Sprintf("`%s`.`%s`", db.dbName, table)
	return &UpdateBuilder{
		db:    db,
		table: fullTableName,
	}
}

// Delete DELETE语句入口
func (db *DB) Delete() *DeleteBuilder {
	return &DeleteBuilder{
		db: db,
	}
}

// GetExecutor 获取当前执行器（用于原生SQL）
func (db *DB) GetExecutor() Executor {
	return db.executor
}

// IsInTransaction 检查是否在事务中
func (db *DB) IsInTransaction() bool {
	_, isTx := db.executor.(*sql.Tx)
	return isTx
}
