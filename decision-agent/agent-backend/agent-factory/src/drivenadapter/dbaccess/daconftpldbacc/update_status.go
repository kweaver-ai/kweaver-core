package daconftpldbacc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigTplRepo) UpdateStatus(ctx context.Context, tx *sql.Tx, status cdaenum.Status, id int64, uid string, publishedAt int64) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.DataAgentTplPo{}
	sr.FromPo(po)

	var (
		updateFields []string
		updateStruct *dapo.DataAgentTplPo
	)

	switch status {
	case cdaenum.StatusPublished:
		updateFields = []string{
			"f_status",
			"f_published_at",
			"f_published_by",
		}
		updateStruct = &dapo.DataAgentTplPo{
			Status:      status,
			PublishedAt: &publishedAt,
			PublishedBy: &uid,
		}
	case cdaenum.StatusUnpublished:
		updateFields = []string{
			"f_status",
			"f_published_at",
			"f_published_by",
		}
		updateStruct = &dapo.DataAgentTplPo{
			Status: status,
		}
		updateStruct.SetPublishedAt(0)
		updateStruct.SetPublishedBy("")
	default:
		err = errors.New("invalid status")
		return
	}

	_, err = sr.WhereEqual("f_id", id).
		SetUpdateFields(updateFields).
		UpdateByStruct(updateStruct)

	return
}
