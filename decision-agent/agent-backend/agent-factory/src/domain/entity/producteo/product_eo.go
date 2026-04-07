package producteo

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"

// Product 产品实体对象
type Product struct {
	dapo.ProductPo
}
