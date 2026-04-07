package publishedtpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *PubedTplRepo) Create(ctx context.Context, tx *sql.Tx, po *dapo.PublishedTplPo) (id int64, err error) {
	sr := dbhelper2.TxSr(tx, repo.logger)

	// 1. 新建po
	sr.FromPo(po)

	_, err = sr.InsertStruct(po)
	if err != nil {
		return
	}

	// 2. 根据key和状态获得id
	_po, err := repo.GetByKeyWithTx(ctx, tx, po.Key)
	if err != nil {
		return
	}

	id = _po.ID

	return
}
