package sandboxplatformhttp

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/isandboxhtpp"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type sandboxPlatformHttpAcc struct {
	logger              icmp.Logger
	client              rest.HTTPClient
	sandboxPlatformConf *conf.SandboxPlatformConf
	baseURL             string
}

var (
	sandboxPlatformOnce sync.Once
	sandboxPlatformImpl isandboxhtpp.ISandboxPlatform
)

func NewSandboxPlatformHttpAcc(sandboxPlatformConf *conf.SandboxPlatformConf, client rest.HTTPClient, logger icmp.Logger) isandboxhtpp.ISandboxPlatform {
	sandboxPlatformOnce.Do(func() {
		sandboxPlatformImpl = &sandboxPlatformHttpAcc{
			logger:              logger,
			client:              client,
			sandboxPlatformConf: sandboxPlatformConf,
			baseURL:             cutil.GetHTTPAccess(sandboxPlatformConf.PrivateSvc.Host, sandboxPlatformConf.PrivateSvc.Port, sandboxPlatformConf.PublicSvc.Protocol),
		}
	})

	return sandboxPlatformImpl
}
