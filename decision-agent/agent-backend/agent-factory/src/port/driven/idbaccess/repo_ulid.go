package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
)

//go:generate mockgen -source=./repo_ulid.go -destination ./dbmock/repo_ulid.go -package dbmock
type UlidRepo interface {
	GenDBID(ctx context.Context, tx *sql.Tx) (id string, err error)
	BatchGenDBID(ctx context.Context, tx *sql.Tx, num int) (ids []string, err error)

	GenUniqID(ctx context.Context, flag cconstant.UniqueIDFlag) (id string, err error)
	DelUniqID(ctx context.Context, flag cconstant.UniqueIDFlag, id string) (err error)
}
