package daconfeo

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// DataAgentTpl 数据智能体模板实体对象
type DataAgentTpl struct {
	dapo.DataAgentTplPo

	Config *daconfvalobj.Config `json:"config"` // Agent 配置（用于创建、更新时使用）

	ProductName string `json:"product_name"` // 产品名称
}

// GetObjName 获取对象名称
func (dat *DataAgentTpl) GetObjName() string {
	return dat.Name
}

// AuditMngLogCreate 创建数据智能体模板的审计日志
func (dat *DataAgentTpl) AuditMngLogCreate(ctx context.Context) {
	// 实现审计日志创建逻辑
}

// AuditMngLogUpdate 更新数据智能体模板的审计日志
func (dat *DataAgentTpl) AuditMngLogUpdate(ctx context.Context) {
	// 实现审计日志更新逻辑
}

// AuditMngLogDelete 删除数据智能体模板的审计日志
func (dat *DataAgentTpl) AuditMngLogDelete(ctx context.Context) {
	// 实现审计日志删除逻辑
}
