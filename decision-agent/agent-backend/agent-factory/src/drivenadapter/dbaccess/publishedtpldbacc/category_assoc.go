package publishedtpldbacc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// 批量创建 agent模板与分类关联
func (repo *PubedTplRepo) BatchCreateCategoryAssoc(ctx context.Context, tx *sql.Tx, pos []*dapo.PubTplCatAssocPo) (err error) {
	if len(pos) == 0 {
		return
	}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(&dapo.PubTplCatAssocPo{})
	_, err = sr.InsertStructs(pos)

	return
}

// 删除 agent模板与分类关联
func (repo *PubedTplRepo) DelCategoryAssocByTplID(ctx context.Context, tx *sql.Tx, tplID int64) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.PubTplCatAssocPo{}
	sr.FromPo(po)

	_, err = sr.WhereEqual("f_published_tpl_id", tplID).Delete()
	if err != nil {
		return errors.Wrapf(err, "delete category by tpl id %d", tplID)
	}

	return nil
}

func (repo *PubedTplRepo) GetCategoryAssocByTplID(ctx context.Context, tx *sql.Tx, tplID int64) (pos []*dapo.PubTplCatAssocPo, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.PubTplCatAssocPo{}
	sr.FromPo(po)

	err = sr.WhereEqual("f_published_tpl_id", tplID).Find(&pos)
	if err != nil {
		return nil, errors.Wrapf(err, "get category assoc by tpl id %d", tplID)
	}

	return pos, nil
}

func (repo *PubedTplRepo) GetCategoryJoinPosByTplID(ctx context.Context, tx *sql.Tx, tplID int64) (pos []*dapo.DataAgentTplCategoryJoinPo, err error) {
	pos = make([]*dapo.DataAgentTplCategoryJoinPo, 0)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	_pos := make([]dapo.DataAgentTplCategoryJoinPo, 0)
	assocPo := &dapo.PubTplCatAssocPo{}
	categoryPo := &dapo.CategoryPO{}

	rawSQL := fmt.Sprintf(`
		SELECT a.f_id, a.f_published_tpl_id, a.f_category_id, c.f_name as f_category_name
		FROM %s a
		LEFT JOIN %s c ON a.f_category_id = c.f_id
		WHERE a.f_published_tpl_id = ?
	`, assocPo.TableName(), categoryPo.TableName())

	err = sr.Raw(rawSQL, tplID).Find(&_pos)
	if err != nil {
		err = errors.Wrapf(err, "get category join pos by tpl id %d", tplID)
		return
	}

	pos = cutil.SliceToPtrSlice(_pos)

	return
}
