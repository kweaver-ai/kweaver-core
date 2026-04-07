package releaseacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetByAgentId implements idbaccess.ReleaseRepo.
func (repo *releaseRepo) GetByAgentID(ctx context.Context, agentID string) (rt *dapo.ReleasePO, err error) {
	po := &dapo.ReleasePO{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_agent_id", agentID).
		FindOne(po)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "get release by agent id %s", agentID)
	}

	return po, nil
}
