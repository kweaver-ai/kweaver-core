package daconfeo

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// DataAgent 数据智能体配置实体对象
type DataAgent struct {
	dapo.DataAgentPo

	Config *daconfvalobj.Config `json:"config"` // Agent 配置（用于创建、更新时使用）
	// ConfigForShow *daconfvalobj.ConfigForShow `json:"-"`            // Agent 配置（用于展示时使用）
	ProductName string `json:"product_name"` // 产品名称

	CreatedByName string `json:"created_by_name"` // 创建人名称
	UpdatedByName string `json:"updated_by_name"` // 更新人名称
}

// GetObjName 获取对象名称
func (da *DataAgent) GetObjName() string {
	return da.Name
}

// AuditMngLogCreate 创建数据智能体的审计日志
func (da *DataAgent) AuditMngLogCreate(ctx context.Context) {
	// 实现审计日志创建逻辑
}

// AuditMngLogUpdate 更新数据智能体的审计日志
func (da *DataAgent) AuditMngLogUpdate(ctx context.Context) {
	// 实现审计日志更新逻辑
}

// AuditMngLogDelete 删除数据智能体的审计日志
func (da *DataAgent) AuditMngLogDelete(ctx context.Context) {
	// 实现审计日志删除逻辑
}

func (da *DataAgent) SetDatasetId(datasetId string) {
	if da.Config == nil {
		return
	}

	if da.Config.DataSource == nil {
		return
	}

	if da.Config.DataSource.Doc == nil {
		return
	}

	if len(da.Config.DataSource.Doc) == 0 {
		return
	}

	docSource := da.Config.DataSource.GetBuiltInDocDataSource()
	if docSource == nil {
		return
	}

	docSource.SetDatasetId(datasetId)
}
