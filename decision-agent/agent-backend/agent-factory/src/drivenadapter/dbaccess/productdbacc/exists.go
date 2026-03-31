package productdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// ExistsByName 根据名称检查是否存在
func (r *ProductRepo) ExistsByName(ctx context.Context, name string) (exists bool, err error) {
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(&dapo.ProductPo{})
	exists, err = sr.WhereEqual("f_name", name).
		WhereEqual("f_deleted_at", 0).
		Exists()

	return
}

// ExistsByKey 根据Key检查是否存在
func (r *ProductRepo) ExistsByKey(ctx context.Context, key string) (exists bool, err error) {
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(&dapo.ProductPo{})
	exists, err = sr.WhereEqual("f_key", key).
		WhereEqual("f_deleted_at", 0).
		Exists()

	return
}

// ExistsByID 根据ID检查是否存在
func (r *ProductRepo) ExistsByID(ctx context.Context, id int64) (exists bool, err error) {
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(&dapo.ProductPo{})

	exists, err = sr.WhereEqual("f_id", id).
		WhereEqual("f_deleted_at", 0).
		Exists()

	return
}

// ExistsByNameExcludeID 根据名称检查是否存在（排除指定ID）
func (r *ProductRepo) ExistsByNameExcludeID(ctx context.Context, name string, id int64) (exists bool, err error) {
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(&dapo.ProductPo{})
	exists, err = sr.WhereEqual("f_name", name).
		WhereNotEqual("f_id", id).
		WhereEqual("f_deleted_at", 0).
		Exists()

	return
}

// ExistsByKeyExcludeID 根据Key检查是否存在（排除指定ID）
func (r *ProductRepo) ExistsByKeyExcludeID(ctx context.Context, key string, id int64) (exists bool, err error) {
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(&dapo.ProductPo{})
	exists, err = sr.WhereEqual("f_key", key).
		WhereNotEqual("f_id", id).
		WhereEqual("f_deleted_at", 0).
		Exists()

	return
}
