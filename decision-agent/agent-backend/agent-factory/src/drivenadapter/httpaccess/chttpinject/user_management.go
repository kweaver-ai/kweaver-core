package chttpinject

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/usermanagementacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
)

var (
	userManagementOnce sync.Once
	userManagementImpl iusermanagementacc.UserMgnt
)

func NewUserManagementClient() iusermanagementacc.UserMgnt {
	userManagementOnce.Do(func() {
		if global.GConfig != nil &&
			global.GConfig.SwitchFields != nil &&
			global.GConfig.SwitchFields.Mock != nil &&
			global.GConfig.SwitchFields.Mock.MockUserManagerModule {
			userManagementImpl = usermanagementacc.NewMockClient()
			return
		}

		userManagementImpl = usermanagementacc.NewClient()
	})

	return userManagementImpl
}
