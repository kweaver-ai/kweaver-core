package bdagenttpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// DeleteByAgentTplID 根据agent模板ID删除关联
func (repo *BizDomainAgentTplRelRepo) DeleteByAgentTplID(ctx context.Context, tx *sql.Tx, agentTplID int64) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.BizDomainAgentTplRelPo{}
	sr.FromPo(po)

	_, err = sr.WhereEqual("f_agent_tpl_id", agentTplID).Delete()
	if err != nil {
		return errors.Wrapf(err, "delete by agent tpl id %d", agentTplID)
	}

	return nil
}
