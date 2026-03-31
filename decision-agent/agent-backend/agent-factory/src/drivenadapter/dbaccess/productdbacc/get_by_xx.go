package productdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// GetByID 根据ID获取产品
func (r *ProductRepo) GetByID(ctx context.Context, id int64) (po *dapo.ProductPo, err error) {
	po = &dapo.ProductPo{}
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(po)
	err = sr.WhereEqual("f_id", id).
		WhereEqual("f_deleted_at", 0).
		FindOne(po)

	return
}

// GetByKey 根据Key获取产品
func (r *ProductRepo) GetByKey(ctx context.Context, key string) (po *dapo.ProductPo, err error) {
	po = &dapo.ProductPo{}
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(po)
	err = sr.WhereEqual("f_key", key).
		WhereEqual("f_deleted_at", 0).
		FindOne(po)

	return
}

func (r *ProductRepo) GetByKeys(ctx context.Context, keys []string) (pos []*dapo.ProductPo, err error) {
	pos = make([]*dapo.ProductPo, 0)

	if len(keys) == 0 {
		return
	}

	po := &dapo.ProductPo{}
	poList := make([]dapo.ProductPo, 0)
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		In("f_key", keys).
		Find(&poList)
	if err != nil {
		return nil, err
	}

	pos = cutil.SliceToPtrSlice(poList)

	return
}

func (r *ProductRepo) GetByNameMapByKeys(ctx context.Context, keys []string) (m map[string]string, err error) {
	m = make(map[string]string)

	if len(keys) == 0 {
		return
	}

	pos, err := r.GetByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}

	for i := range pos {
		m[pos[i].Name] = pos[i].Key
	}

	return
}

// List 获取产品列表
func (r *ProductRepo) List(ctx context.Context, offset, limit int) (pos []*dapo.ProductPo, total int, err error) {
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)

	po := &dapo.ProductPo{}
	sr.FromPo(po).WhereEqual("f_deleted_at", 0)

	count, err := sr.Count()
	if err != nil {
		return
	}

	total = int(count)

	if count == 0 {
		return
	}

	poList := make([]dapo.ProductPo, 0)

	sr.ResetSelect().
		Order("f_id DESC").
		Offset(offset).
		Limit(limit)

	err = sr.Find(&poList)
	if err != nil {
		return
	}

	pos = cutil.SliceToPtrSlice(poList)

	return
}
