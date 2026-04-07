package daconfdbacc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigRepo) UpdateStatus(ctx context.Context, tx *sql.Tx, status cdaenum.Status, id string, uid string) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.DataAgentPo{}
	sr.FromPo(po)

	var (
		updateFields []string
		updateStruct *dapo.DataAgentPo
	)

	switch status {
	case cdaenum.StatusPublished, cdaenum.StatusUnpublished:
		updateFields = []string{"f_status"}
		updateStruct = &dapo.DataAgentPo{
			Status: status,
		}
	default:
		err = errors.New("invalid status")
		return
	}

	_, err = sr.WhereEqual("f_id", id).
		SetUpdateFields(updateFields).
		UpdateByStruct(updateStruct)

	return
}
