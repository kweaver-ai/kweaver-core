package producteo

import "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"

// Product 产品实体对象
type Product struct {
	dapo.ProductPo
}
