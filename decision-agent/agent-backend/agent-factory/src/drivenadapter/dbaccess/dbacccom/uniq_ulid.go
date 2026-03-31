package dbacccom

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type UniqUlidHelper struct {
	Po     dbhelper2.ITable
	Pk     string
	Tx     *sql.Tx // 比DB优先级高
	DB     dbhelper2.ISQLRunner
	Logger icmp.Logger
}

func NewUniqUlidHelper(dto *UniqUlidHelper) *UniqUlidHelper {
	if dto == nil {
		panic("[UniqUlidHelper][NewUniqUlidHelper]: dto is nil")
	}

	if dto.Po == nil {
		panic("[UniqUlidHelper][NewUniqUlidHelper]: dto.Po is nil")
	}

	if dto.Pk == "" {
		dto.Pk = "f_id"
	}

	if dto.DB == nil && dto.Tx == nil {
		panic("[UniqUlidHelper][NewUniqUlidHelper]: dto.DB is nil and dto.Tx is nil")
	}

	return &UniqUlidHelper{
		Po:     dto.Po,
		Pk:     dto.Pk,
		DB:     dto.DB,
		Tx:     dto.Tx,
		Logger: dto.Logger,
	}
}

func (h *UniqUlidHelper) GenDBID(ctx context.Context) (id string, err error) {
	maxRetry := 50
	for i := 0; i < maxRetry; i++ {
		id, err = h.genDBID(ctx)
		if err != nil {
			continue
		}

		if id != "" {
			break
		}
	}

	if id == "" {
		err = errors.New("[UniqUlidHelper][GenDBID]: failed to generate unique id")
	}

	return
}

func (h *UniqUlidHelper) genDBID(ctx context.Context) (id string, err error) {
	sr := dbhelper2.NewSQLRunner(h.DB, h.Logger)

	if h.Tx != nil {
		sr = dbhelper2.TxSr(h.Tx, h.Logger)
	}

	id = cutil.UlidMake()

	exists, err := sr.FromPo(h.Po).
		WhereEqual(h.Pk, id).
		Exists()
	if err != nil {
		id = ""
	}

	if exists {
		id = ""
	}

	return
}
