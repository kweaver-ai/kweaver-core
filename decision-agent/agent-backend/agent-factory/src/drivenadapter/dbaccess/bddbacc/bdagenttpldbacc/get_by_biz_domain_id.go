package bdagenttpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetByBizDomainID 根据业务域ID获取关联列表
func (repo *BizDomainAgentTplRelRepo) GetByBizDomainID(ctx context.Context, tx *sql.Tx, bizDomainID string) (pos []*dapo.BizDomainAgentTplRelPo, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.BizDomainAgentTplRelPo{}
	sr.FromPo(po)

	poList := make([]dapo.BizDomainAgentTplRelPo, 0)

	err = sr.WhereEqual("f_biz_domain_id", bizDomainID).Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by biz domain id %s", bizDomainID)
	}

	pos = cutil.SliceToPtrSlice(poList)

	return pos, nil
}
