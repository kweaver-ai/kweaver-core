package chttpinject

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/umhttpaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	umOnce sync.Once
	umImpl iumacc.UmHttpAcc
)

func NewUmHttpAcc() iumacc.UmHttpAcc {
	umOnce.Do(func() {
		if global.GConfig != nil &&
			global.GConfig.SwitchFields != nil &&
			global.GConfig.SwitchFields.Mock != nil &&
			global.GConfig.SwitchFields.Mock.MockUserManagerModule {
			umImpl = umhttpaccess.NewMockUmHttpAcc()
			return
		}

		umImpl = umhttpaccess.NewUmHttpAcc(logger.GetLogger())
	})

	return umImpl
}
