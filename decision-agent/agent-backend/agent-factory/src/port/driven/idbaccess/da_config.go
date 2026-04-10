package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/da_config.go github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess IDataAgentConfigRepo
type IDataAgentConfigRepo interface {
	IDBAccBaseRepo

	Create(ctx context.Context, tx *sql.Tx, id string, po *dapo.DataAgentPo) (err error)
	CreateBatch(ctx context.Context, tx *sql.Tx, pos []*dapo.DataAgentPo) (err error)

	Update(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentPo) (err error)
	UpdateByKey(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentPo) (err error)

	Delete(ctx context.Context, tx *sql.Tx, id string) (err error)

	UpdateStatus(ctx context.Context, tx *sql.Tx, status cdaenum.Status, id string, uid string) (err error)

	ExistsByName(ctx context.Context, name string) (exists bool, err error)

	ExistsByID(ctx context.Context, id string) (exists bool, err error)
	ExistsByNameExcludeID(ctx context.Context, name, id string) (exists bool, err error)

	GetByID(ctx context.Context, id string) (po *dapo.DataAgentPo, err error)
	GetByIDS(ctx context.Context, ids []string) (po []*dapo.DataAgentPo, err error)

	GetMapByIDs(ctx context.Context, ids []string) (res map[string]*dapo.DataAgentPo, err error)

	GetIDNameMapByID(ctx context.Context, ids []string) (res map[string]string, err error)

	GetByKey(ctx context.Context, key string) (po *dapo.DataAgentPo, err error)
	GetByKeys(ctx context.Context, keys []string) (pos []*dapo.DataAgentPo, err error)

	// GetByIDsAndCreatedBy 根据ID列表和创建者获取agent
	GetByIDsAndCreatedBy(ctx context.Context, ids []string, createdBy string) (pos []*dapo.DataAgentPo, err error)

	ListForBenchmark(ctx context.Context, req *agentconfigreq.ListForBenchmarkReq) (pos []*dapo.ListForBenchmarkPo, count int64, err error)

	// GetAllIDs 获取所有未删除的agent ID列表
	GetAllIDs(ctx context.Context) (ids []string, err error)
}
