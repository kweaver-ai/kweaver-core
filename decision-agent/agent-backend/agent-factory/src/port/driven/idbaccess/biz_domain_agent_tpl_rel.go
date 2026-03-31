package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/biz_domain_agent_tpl_rel.go github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess IBizDomainAgentTplRelRepo

// IBizDomainAgentTplRelRepo 业务域与agent模板关联表操作接口
type IBizDomainAgentTplRelRepo interface {
	IDBAccBaseRepo

	// BatchCreate 批量创建业务域与agent模板关联
	BatchCreate(ctx context.Context, tx *sql.Tx, pos []*dapo.BizDomainAgentTplRelPo) error

	// DeleteByBizDomainID 根据业务域ID删除关联
	DeleteByBizDomainID(ctx context.Context, tx *sql.Tx, bizDomainID string) error

	// DeleteByAgentTplID 根据agent模板ID删除关联
	DeleteByAgentTplID(ctx context.Context, tx *sql.Tx, agentTplID int64) error

	// GetByBizDomainID 根据业务域ID获取关联列表
	GetByBizDomainID(ctx context.Context, tx *sql.Tx, bizDomainID string) ([]*dapo.BizDomainAgentTplRelPo, error)

	// GetByAgentTplID 根据agent模板ID获取关联列表
	GetByAgentTplID(ctx context.Context, tx *sql.Tx, agentTplID int64) ([]*dapo.BizDomainAgentTplRelPo, error)
}
