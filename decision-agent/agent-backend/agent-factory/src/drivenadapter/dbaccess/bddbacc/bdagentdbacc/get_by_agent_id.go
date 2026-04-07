package bdagentdbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetByAgentID 根据agent ID获取关联列表
func (repo *BizDomainAgentRelRepo) GetByAgentID(ctx context.Context, tx *sql.Tx, agentID string) (pos []*dapo.BizDomainAgentRelPo, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.BizDomainAgentRelPo{}
	sr.FromPo(po)

	poList := make([]dapo.BizDomainAgentRelPo, 0)

	err = sr.WhereEqual("f_agent_id", agentID).Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by agent id %s", agentID)
	}

	pos = cutil.SliceToPtrSlice(poList)

	return pos, nil
}
