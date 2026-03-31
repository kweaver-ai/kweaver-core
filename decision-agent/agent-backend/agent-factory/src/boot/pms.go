package boot

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
)

func initPermission() (err error) {
	if global.GConfig.SwitchFields.DisablePmsCheck {
		return
	}

	pmsSvc := dainject.NewPermissionSvc()
	ctx := context.Background()

	err = pmsSvc.InitPermission(ctx)
	if err != nil {
		return
	}

	return
}
