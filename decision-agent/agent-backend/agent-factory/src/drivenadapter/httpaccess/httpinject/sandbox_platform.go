package httpinject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/isandboxhtpp"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

var (
	sandboxPlatformOnce sync.Once
	sandboxPlatformImpl isandboxhtpp.ISandboxPlatform
)

func NewSandboxPlatformHttpAcc() isandboxhtpp.ISandboxPlatform {
	sandboxPlatformOnce.Do(func() {
		sandboxPlatformConf := global.GConfig.SandboxPlatformConf

		if global.GConfig.SwitchFields.Mock.MockSandboxPlatform {
			sandboxPlatformImpl = sandboxplatformhttp.NewMockSandboxPlatform(logger.GetLogger())
		} else {
			sandboxPlatformImpl = sandboxplatformhttp.NewSandboxPlatformHttpAcc(
				sandboxPlatformConf,
				rest.NewHTTPClient(),
				logger.GetLogger(),
			)
		}
	})

	return sandboxPlatformImpl
}
