package idbaccess

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/product.go github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess IProductRepo
type IProductRepo interface {
	// 基本CRUD操作
	Create(ctx context.Context, po *dapo.ProductPo) (key string, err error)
	Update(ctx context.Context, po *dapo.ProductPo) (err error)
	Delete(ctx context.Context, id int64) (err error)

	// 查询操作
	GetByID(ctx context.Context, id int64) (po *dapo.ProductPo, err error)
	GetByKey(ctx context.Context, key string) (po *dapo.ProductPo, err error)
	GetByKeys(ctx context.Context, keys []string) (pos []*dapo.ProductPo, err error)
	GetByNameMapByKeys(ctx context.Context, keys []string) (m map[string]string, err error)

	List(ctx context.Context, offset, limit int) (pos []*dapo.ProductPo, total int, err error)

	// 存在性检查
	ExistsByName(ctx context.Context, name string) (exists bool, err error)
	ExistsByKey(ctx context.Context, key string) (exists bool, err error)
	ExistsByID(ctx context.Context, id int64) (exists bool, err error)
	ExistsByNameExcludeID(ctx context.Context, name string, id int64) (exists bool, err error)
	ExistsByKeyExcludeID(ctx context.Context, key string, id int64) (exists bool, err error)
}
