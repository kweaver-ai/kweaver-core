package dbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type baseTx struct {
	dbPool *sqlx.DB
}

func NewBaseTx() model.DBTx {
	return &baseTx{
		dbPool: db.NewDBPool(),
	}
}

func (b *baseTx) GetTx(ctx context.Context) (*sql.Tx, error) {
	return b.dbPool.BeginTx(ctx, nil)
}
