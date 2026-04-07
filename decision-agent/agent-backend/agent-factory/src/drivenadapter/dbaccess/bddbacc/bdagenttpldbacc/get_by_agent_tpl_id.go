package bdagenttpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetByAgentTplID 根据agent模板ID获取关联列表
func (repo *BizDomainAgentTplRelRepo) GetByAgentTplID(ctx context.Context, tx *sql.Tx, agentTplID int64) (pos []*dapo.BizDomainAgentTplRelPo, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.BizDomainAgentTplRelPo{}
	sr.FromPo(po)

	poList := make([]dapo.BizDomainAgentTplRelPo, 0)

	err = sr.WhereEqual("f_agent_tpl_id", agentTplID).Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by agent tpl id %d", agentTplID)
	}

	pos = cutil.SliceToPtrSlice(poList)

	return pos, nil
}
