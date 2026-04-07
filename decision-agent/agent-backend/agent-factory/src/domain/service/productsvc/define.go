package productsvc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/productdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/rediscmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

var (
	productSvcOnce sync.Once
	productSvcImpl iv3portdriver.IProductSvc
	// agentCreateUpdateDlmName = "agent_create_update" // 用于应用层策略创建和更新的分布式锁（创建和更新操作目前也不允许并发）
)

type productSvc struct {
	*service.SvcBase
	productRepo idbaccess.IProductRepo
	redisCmp    icmp.RedisCmp
}

var _ iv3portdriver.IProductSvc = &productSvc{}

func NewProductService() iv3portdriver.IProductSvc {
	productSvcOnce.Do(func() {
		productSvcImpl = &productSvc{
			redisCmp:    rediscmp.NewRedisCmp(),
			SvcBase:     service.NewSvcBase(),
			productRepo: productdbacc.NewProductRepo(),
		}
	})

	return productSvcImpl
}
