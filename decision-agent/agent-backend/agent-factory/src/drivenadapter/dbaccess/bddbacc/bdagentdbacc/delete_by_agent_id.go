package bdagentdbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// DeleteByAgentID 根据agent ID删除关联
func (repo *BizDomainAgentRelRepo) DeleteByAgentID(ctx context.Context, tx *sql.Tx, agentID string) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.BizDomainAgentRelPo{}
	sr.FromPo(po)

	_, err = sr.WhereEqual("f_agent_id", agentID).Delete()
	if err != nil {
		return errors.Wrapf(err, "delete by agent id %s", agentID)
	}

	return nil
}
