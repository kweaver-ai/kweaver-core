package dainject

import (
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/cconf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func getModelApiUrlPrefix(_conf *cconf.ModelFactoryConf) (urlPrefix string) {
	apiSvc := _conf.ModelApiSvc
	host := cutil.ParseHost(apiSvc.Host)

	urlPrefix = fmt.Sprintf("%s://%s:%d/api/private/mf-model-api/v1", apiSvc.Protocol, host, apiSvc.Port)

	return
}
