package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/biz_domain_agent_rel.go github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess IBizDomainAgentRelRepo

// IBizDomainAgentRelRepo 业务域与agent关联表操作接口
type IBizDomainAgentRelRepo interface {
	IDBAccBaseRepo

	// BatchCreate 批量创建业务域与agent关联
	BatchCreate(ctx context.Context, tx *sql.Tx, pos []*dapo.BizDomainAgentRelPo) error

	// DeleteByBizDomainID 根据业务域ID删除关联
	DeleteByBizDomainID(ctx context.Context, tx *sql.Tx, bizDomainID string) error

	// DeleteByAgentID 根据agent ID删除关联
	DeleteByAgentID(ctx context.Context, tx *sql.Tx, agentID string) error

	// GetByBizDomainID 根据业务域ID获取关联列表
	GetByBizDomainID(ctx context.Context, tx *sql.Tx, bizDomainID string) ([]*dapo.BizDomainAgentRelPo, error)

	// GetByAgentID 根据agent ID获取关联列表
	GetByAgentID(ctx context.Context, tx *sql.Tx, agentID string) ([]*dapo.BizDomainAgentRelPo, error)
}
