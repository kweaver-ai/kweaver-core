package httpinject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iuniqueryhttp"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

var (
	uniqueryOnce sync.Once
	uniqueryImpl iuniqueryhttp.IUniquery
)

func NewUniqueryHttpAcc() iuniqueryhttp.IUniquery {
	uniqueryOnce.Do(func() {
		uniqueryConf := global.GConfig.UniqueryConf
		uniqueryImpl = uniqueryaccess.NewUniqueryHttpAcc(
			logger.GetLogger(),
			uniqueryConf,
			rest.NewHTTPClient(),
		)
	})

	return uniqueryImpl
}
