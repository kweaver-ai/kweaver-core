// Package model 定义数据库操作接口
// @file tx.go
// @description: 定义数据库事务操作接口
package model

//go:generate mockgen -source=tx.go -destination=../../mocks/model_tx.go -package=mocks
import (
	"context"
	"database/sql"
)

type DBTx interface {
	GetTx(ctx context.Context) (*sql.Tx, error)
}
