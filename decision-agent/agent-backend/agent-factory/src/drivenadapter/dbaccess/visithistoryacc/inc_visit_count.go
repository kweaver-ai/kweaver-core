package visithistoryacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// IncVisitCount implements idbaccess.IVisitHistoryRepo.
func (repo *visitHistoryRepo) IncVisitCount(ctx context.Context, po *dapo.VisitHistoryPO) (err error) {
	if po.AgentID == "" || po.AgentVersion == "" || po.CreateBy == "" {
		return errors.Wrapf(errors.New("agentID or agentVersion or userID is empty"), "inc visit count")
	}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	exist, err := sr.WhereEqual("f_agent_id", po.AgentID).WhereEqual("f_agent_version", po.AgentVersion).WhereEqual("f_create_by", po.CreateBy).Exists()
	if err != nil {
		return errors.Wrapf(err, "check agent id %s exist", po.AgentID)
	}

	if !exist {
		po.ID = cutil.UlidMake()

		_, err := sr.InsertStruct(po)
		if err != nil {
			return errors.Wrapf(err, "insert agent id %s", po.AgentID)
		}

		return nil
	}

	_, err = sr.RawExec(
		"UPDATE "+po.TableName()+" SET f_visit_count = f_visit_count + 1, f_update_time= ? WHERE f_agent_id = ? AND f_agent_version = ? AND f_create_by = ?",
		po.UpdateTime,
		po.AgentID,
		po.AgentVersion,
		po.CreateBy,
	)

	return
}
