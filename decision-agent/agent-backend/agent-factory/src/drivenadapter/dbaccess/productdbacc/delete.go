package productdbacc

import (
	"context"
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// Delete 删除产品（软删除）
func (r *ProductRepo) Delete(ctx context.Context, id int64) (err error) {
	po := &dapo.ProductPo{}

	sr := dbhelper2.NewSQLRunner(r.db, r.logger)

	uid := chelper.GetUserIDFromCtx(ctx)
	if uid == "" {
		err = errors.New("[ProductRepo][Delete]: uid is empty")
		return
	}

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_id", id).
		SetUpdateFields([]string{
			"f_deleted_at",
			"f_deleted_by",
		}).
		UpdateByStruct(struct {
			DeletedAt int64  `db:"f_deleted_at"`
			DeletedBy string `db:"f_deleted_by"`
		}{
			DeletedAt: cutil.GetCurrentMSTimestamp(),
			DeletedBy: uid,
		})

	return
}
