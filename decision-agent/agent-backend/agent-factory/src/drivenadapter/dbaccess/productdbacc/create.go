package productdbacc

import (
	"context"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// Create 创建产品
func (r *ProductRepo) Create(ctx context.Context, po *dapo.ProductPo) (key string, err error) {
	sr := dbhelper2.NewSQLRunner(r.db, r.logger)

	// 设置创建时间
	now := time.Now().UnixMilli()
	po.CreatedAt = now
	po.UpdatedAt = now

	// 生成Key如果为空
	if po.Key == "" {
		po.Key = cutil.UlidMake()
	}

	sr.FromPo(po)

	_, err = sr.InsertStruct(po)
	if err != nil {
		return
	}

	key = po.Key

	return
}
