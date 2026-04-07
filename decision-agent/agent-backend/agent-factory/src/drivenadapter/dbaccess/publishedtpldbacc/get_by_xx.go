package publishedtpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *PubedTplRepo) GetByID(ctx context.Context, id int64) (po *dapo.PublishedTplPo, err error) {
	po = &dapo.PublishedTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.
		WhereEqual("f_id", id).
		FindOne(po)

	return
}

func (repo *PubedTplRepo) GetByKey(ctx context.Context, key string) (po *dapo.PublishedTplPo, err error) {
	po = &dapo.PublishedTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.
		WhereEqual("f_key", key).
		FindOne(po)

	return
}

func (repo *PubedTplRepo) GetByTplID(ctx context.Context, tplID int64) (po *dapo.PublishedTplPo, err error) {
	po = &dapo.PublishedTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.
		WhereEqual("f_tpl_id", tplID).
		FindOne(po)

	return
}

func (repo *PubedTplRepo) GetByIDWithTx(ctx context.Context, tx *sql.Tx, id int64) (po *dapo.PublishedTplPo, err error) {
	po = &dapo.PublishedTplPo{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	err = sr.
		WhereEqual("f_id", id).
		FindOne(po)
	if err != nil {
		return nil, errors.Wrapf(err, "[GetByIDWithTx]: get by id %d", id)
	}

	return
}

func (repo *PubedTplRepo) GetByKeyWithTx(ctx context.Context, tx *sql.Tx, key string) (po *dapo.PublishedTplPo, err error) {
	po = &dapo.PublishedTplPo{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	err = sr.
		WhereEqual("f_key", key).
		FindOne(po)
	if err != nil {
		return nil, errors.Wrapf(err, "[GetByKeyWithTx]: get by key %s", key)
	}

	return
}
