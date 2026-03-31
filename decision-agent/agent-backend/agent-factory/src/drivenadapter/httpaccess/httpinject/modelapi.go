package httpinject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/modelfactoryacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/httpclient"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/imodelfactoryacc"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

var (
	modelApiOnce sync.Once
	modelApiImpl imodelfactoryacc.IModelApiAcc
)

func NewModelApiAcc() imodelfactoryacc.IModelApiAcc {
	modelApiOnce.Do(func() {
		modelApiImpl = modelfactoryacc.NewModelApiAcc(
			httpclient.NewHTTPClient(),
			rest.NewHTTPClient(),
			logger.GetLogger(),
		)
	})

	return modelApiImpl
}
