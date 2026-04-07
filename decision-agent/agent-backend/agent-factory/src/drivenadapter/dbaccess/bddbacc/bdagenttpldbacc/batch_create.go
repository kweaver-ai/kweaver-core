package bdagenttpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

const batchSize = 5000

// BatchCreate 批量创建业务域与agent模板关联（分批写入，每批5000条）
func (repo *BizDomainAgentTplRelRepo) BatchCreate(ctx context.Context, tx *sql.Tx, pos []*dapo.BizDomainAgentTplRelPo) (err error) {
	if len(pos) == 0 {
		return
	}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(&dapo.BizDomainAgentTplRelPo{})
	err = sr.InsertStructsInBatches(pos, batchSize)

	return
}
