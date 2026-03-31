package dbaccess

import (
	"context"
	"database/sql"
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	baseRepoOnce sync.Once
	baseRepoImpl idbaccess.IDBAccBaseRepo
)

var _ idbaccess.IDBAccBaseRepo = &DBAccBase{}

type DBAccBase struct {
	db *sqlx.DB
}

func NewDBAccBase() idbaccess.IDBAccBaseRepo {
	baseRepoOnce.Do(func() {
		db := global.GDB
		baseRepoImpl = &DBAccBase{
			db: db,
		}
	})

	return baseRepoImpl
}

func (r *DBAccBase) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *DBAccBase) GetDB() *sqlx.DB {
	return r.db
}
