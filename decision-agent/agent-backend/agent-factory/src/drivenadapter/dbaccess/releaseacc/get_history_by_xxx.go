package releaseacc

import (
	"context"
	"strconv"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *releaseHistoryRepo) GetLatestVersionByAgentID(ctx context.Context, agentID string) (rt *dapo.ReleaseHistoryPO, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	list := make([]dapo.ReleaseHistoryPO, 0)
	po := &dapo.ReleaseHistoryPO{}
	sr.FromPo(po)

	err = sr.WhereEqual("f_agent_id", agentID).Order("f_create_time DESC").Find(&list)
	if err != nil {
		err = errors.Wrapf(err, "get release history by agent id %s failed", agentID)
		return
	}

	if len(list) <= 0 {
		return nil, nil
	}

	var maxVersion int

	for _, row := range list {
		rowTmp := row
		versionStr := strings.TrimPrefix(row.AgentVersion, "v")

		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return nil, errors.Wrapf(err, "convert version to int failed")
		}

		if version > maxVersion {
			maxVersion = version
			rt = &rowTmp
		}
	}

	return rt, err
}

// GetByAgentIdVersion implements idbaccess.ReleaseHistoryRepo.
func (repo *releaseHistoryRepo) GetByAgentIdVersion(ctx context.Context, agentID string, version string) (rt *dapo.ReleaseHistoryPO, err error) {
	po := &dapo.ReleaseHistoryPO{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_agent_id", agentID).WhereEqual("f_agent_version", version).FindOne(po)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "get release by agent id %s", agentID)
	}

	return po, nil
}
