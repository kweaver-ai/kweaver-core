package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/published_tpl.go github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess IPublishedTplRepo
type IPublishedTplRepo interface {
	IDBAccBaseRepo

	Create(ctx context.Context, tx *sql.Tx, po *dapo.PublishedTplPo) (id int64, err error)

	Delete(ctx context.Context, tx *sql.Tx, id int64) (err error)

	DeleteByTplID(ctx context.Context, tx *sql.Tx, tplID int64) (err error)

	ExistsByKey(ctx context.Context, key string) (exists bool, err error)

	ExistsByID(ctx context.Context, id int64) (exists bool, err error)

	GetByID(ctx context.Context, id int64) (po *dapo.PublishedTplPo, err error)

	GetByIDWithTx(ctx context.Context, tx *sql.Tx, id int64) (po *dapo.PublishedTplPo, err error)

	GetByTplID(ctx context.Context, tplID int64) (po *dapo.PublishedTplPo, err error)

	GetByKey(ctx context.Context, key string) (po *dapo.PublishedTplPo, err error)

	GetByKeyWithTx(ctx context.Context, tx *sql.Tx, key string) (po *dapo.PublishedTplPo, err error)

	GetByCategoryID(ctx context.Context, categoryID string) (po []*dapo.PublishedTplPo, err error)

	// 新增已发布模板列表方法
	GetPubTplList(ctx context.Context, req *pubedreq.PubedTplListReq) (rt []*dapo.PublishedTplPo, err error)

	//--- category start ---
	BatchCreateCategoryAssoc(ctx context.Context, tx *sql.Tx, pos []*dapo.PubTplCatAssocPo) (err error)
	DelCategoryAssocByTplID(ctx context.Context, tx *sql.Tx, tplID int64) (err error)

	GetCategoryAssocByTplID(ctx context.Context, tx *sql.Tx, tplID int64) (pos []*dapo.PubTplCatAssocPo, err error)

	GetCategoryJoinPosByTplID(ctx context.Context, tx *sql.Tx, tplID int64) (pos []*dapo.DataAgentTplCategoryJoinPo, err error)
	//--- category end ---
}
