package pubedeo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// DataAgentTpl 数据智能体模板实体对象
type PublishedTpl struct {
	dapo.PublishedTplPo

	Config *daconfvalobj.Config `json:"config"` // Agent 配置（用于创建、更新时使用）

	ProductName string `json:"product_name"` // 产品名称
}
