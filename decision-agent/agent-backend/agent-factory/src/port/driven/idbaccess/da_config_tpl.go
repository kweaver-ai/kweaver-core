package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/da_config_tpl.go github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess IDataAgentTplRepo
type IDataAgentTplRepo interface {
	IDBAccBaseRepo

	Create(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentTplPo) (err error)

	Update(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentTplPo) (err error)
	Delete(ctx context.Context, tx *sql.Tx, id int64) (err error)

	UpdateStatus(ctx context.Context, tx *sql.Tx, status cdaenum.Status, id int64, uid string, publishedAt int64) (err error)

	ExistsByName(ctx context.Context, name string) (exists bool, err error)
	ExistsByKey(ctx context.Context, key string) (exists bool, err error)

	ExistsByID(ctx context.Context, id int64) (exists bool, err error)
	ExistsByNameExcludeID(ctx context.Context, name string, id int64) (exists bool, err error)
	ExistsByKeyExcludeID(ctx context.Context, key string, id int64) (exists bool, err error)

	GetByID(ctx context.Context, id int64) (po *dapo.DataAgentTplPo, err error)
	GetByIDS(ctx context.Context, ids []int64) (po []*dapo.DataAgentTplPo, err error)

	GetByIDWithTx(ctx context.Context, tx *sql.Tx, id int64) (po *dapo.DataAgentTplPo, err error)

	GetMapByIDs(ctx context.Context, ids []int64) (res map[int64]*dapo.DataAgentTplPo, err error)

	GetByKey(ctx context.Context, key string) (po *dapo.DataAgentTplPo, err error)
	GetByKeys(ctx context.Context, keys []string) (po []*dapo.DataAgentTplPo, err error)

	GetByKeyWithTx(ctx context.Context, tx *sql.Tx, key string) (po *dapo.DataAgentTplPo, err error)

	// GetAllIDs 获取所有未删除的agent模板ID列表
	GetAllIDs(ctx context.Context) (ids []int64, err error)
}
