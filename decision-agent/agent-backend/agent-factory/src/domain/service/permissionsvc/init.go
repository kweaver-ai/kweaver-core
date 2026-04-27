package permissionsvc

import (
	"context"

	"github.com/pkg/errors"
)

func (svc *permissionSvc) InitPermission(ctx context.Context) (err error) {
	//if cenvhelper.IsLocalDev() {
	//	return
	//}
	// 1. 更新Agent资源类型 (这里目前不能这样写 ，因为Authorization服务那边会进行覆盖 2025-12-25备注)
	//err = svc.updateAgentResourceType(ctx)
	//if err != nil {
	//	err = errors.Wrapf(err, "update agent resource type failed")
	//	return
	//}
	// 2. 授予应用管理员-所有Agent的“使用权限”
	err = svc.authZHttp.GrantAgentUsePmsForAppAdmin(ctx)
	if err != nil {
		err = errors.Wrapf(err, "grant agent use pms for app admin failed")
		return
	}

	// 3. 授予应用管理员-相关资源的管理权限（Agent、Agent模板）
	err = svc.authZHttp.GrantMgmtPmsForAppAdmin(ctx)
	if err != nil {
		err = errors.Wrapf(err, "grant agent use pms for app admin failed")
		return
	}

	return
}
