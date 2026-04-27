package releasesvc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/pmsvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (svc *releaseSvc) handlePmsCtrl(ctx context.Context, pmsControl *pmsvo.PmsControlObjS, releaseID, agentID string, tx *sql.Tx) (err error) {
	// 1. 先删除现有的权限记录
	err = svc.releasePermissionRepo.DelByReleaseID(ctx, tx, releaseID)
	if err != nil {
		err = errors.Wrapf(err, "delete permissions failed")
		return
	}

	// 2. 获取Agent名称
	var agentName string

	agentName, err = svc.getAgentName(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "get agent name failed")
		return
	}

	// 3. 删除Agent使用权限
	err = svc.removeUsePmsByHTTPAcc(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "remove use pms failed")
		return
	}

	// 4. 添加权限

	if pmsControl != nil {
		// 4.1 给特定访问者授权（授权部分范围）
		err = svc.handlePmsCtrlRange(ctx, pmsControl, releaseID, agentID, tx, agentName)
		if err != nil {
			err = errors.Wrapf(err, "handlePmsCtrlRange failed")
			return
		}
	}
	// 注: pmsControl 为 nil 时，暂不授权给所有人（grantUsePmsToAll 已注释）

	return
}

func (svc *releaseSvc) handlePmsCtrlRange(ctx context.Context, pmsControl *pmsvo.PmsControlObjS, releaseID string, agentID string, tx *sql.Tx, agentName string) (err error) {
	pmsMap := make(map[cenum.PmsTargetObjType][]string)

	// 1. 添加角色权限

	// 1.1 本地添加角色权限
	rolePermissions := make([]*dapo.ReleasePermissionPO, 0)

	for _, roleID := range pmsControl.RoleIDs {
		permission := &dapo.ReleasePermissionPO{
			ReleaseId:  releaseID,
			ObjectId:   roleID,
			ObjectType: cenum.PmsTargetObjTypeRole,
		}

		rolePermissions = append(rolePermissions, permission)
	}

	err = svc.releasePermissionRepo.BatchCreate(ctx, tx, rolePermissions)
	if err != nil {
		err = errors.Wrapf(err, "batch create role permissions failed")
		return
	}

	// 1.2 远程添加角色权限
	pmsMap[cenum.PmsTargetObjTypeRole] = pmsControl.RoleIDs

	// 2. 添加用户权限

	// 2.1 本地添加用户权限
	userPermissions := make([]*dapo.ReleasePermissionPO, 0)

	for _, userID := range pmsControl.UserIDs {
		permission := &dapo.ReleasePermissionPO{
			ReleaseId:  releaseID,
			ObjectId:   userID,
			ObjectType: cenum.PmsTargetObjTypeUser,
		}

		userPermissions = append(userPermissions, permission)
	}

	err = svc.releasePermissionRepo.BatchCreate(ctx, tx, userPermissions)
	if err != nil {
		err = errors.Wrapf(err, "batch create user permissions failed")
		return
	}

	// 2.2 远程添加用户权限
	pmsMap[cenum.PmsTargetObjTypeUser] = pmsControl.UserIDs

	// 3. 添加用户组权限

	// 3.1 本地添加用户组权限
	userGroupPermissions := make([]*dapo.ReleasePermissionPO, 0)

	for _, userGroupID := range pmsControl.UserGroupIDs {
		permission := &dapo.ReleasePermissionPO{
			ReleaseId:  releaseID,
			ObjectId:   userGroupID,
			ObjectType: cenum.PmsTargetObjTypeUserGroup,
		}

		userGroupPermissions = append(userGroupPermissions, permission)
	}

	err = svc.releasePermissionRepo.BatchCreate(ctx, tx, userGroupPermissions)
	if err != nil {
		err = errors.Wrapf(err, "batch create user group permissions failed")
		return
	}

	// 3.2 远程添加用户组权限
	pmsMap[cenum.PmsTargetObjTypeUserGroup] = pmsControl.UserGroupIDs

	// 4. 添加部门（或组织）权限

	// 4.1 本地添加部门（或组织）权限
	departmentPermissions := make([]*dapo.ReleasePermissionPO, 0)

	for _, departmentID := range pmsControl.DepartmentIDs {
		permission := &dapo.ReleasePermissionPO{
			ReleaseId:  releaseID,
			ObjectId:   departmentID,
			ObjectType: cenum.PmsTargetObjTypeDep,
		}

		departmentPermissions = append(departmentPermissions, permission)
	}

	err = svc.releasePermissionRepo.BatchCreate(ctx, tx, departmentPermissions)
	if err != nil {
		err = errors.Wrapf(err, "batch create department permissions failed")
		return
	}

	// 4.2 远程添加部门（或组织）权限
	pmsMap[cenum.PmsTargetObjTypeDep] = pmsControl.DepartmentIDs

	// 5. 添加应用账号权限

	// 5.1 本地添加应用账号权限
	appAccountPermissions := make([]*dapo.ReleasePermissionPO, 0)

	for _, appAccountID := range pmsControl.AppAccountIDs {
		permission := &dapo.ReleasePermissionPO{
			ReleaseId:  releaseID,
			ObjectId:   appAccountID,
			ObjectType: cenum.PmsTargetObjTypeAppAccount,
		}

		appAccountPermissions = append(appAccountPermissions, permission)
	}

	err = svc.releasePermissionRepo.BatchCreate(ctx, tx, appAccountPermissions)
	if err != nil {
		err = errors.Wrapf(err, "batch create app account permissions failed")
		return
	}

	// 5.2 远程添加应用账号权限
	pmsMap[cenum.PmsTargetObjTypeAppAccount] = pmsControl.AppAccountIDs

	// 6. 远程添加权限（权限平台）
	err = svc.grantUsePms(ctx, agentID, agentName, pmsMap)
	if err != nil {
		err = errors.Wrapf(err, "grant use pms failed")
		return
	}

	return
}

func (svc *releaseSvc) getAgentName(ctx context.Context, agentID string) (agentName string, err error) {
	agentNameMap, err := svc.agentConfigRepo.GetIDNameMapByID(ctx, []string{agentID})
	if err != nil {
		err = errors.Wrapf(err, "get agent name map failed")
		return
	}

	var ok bool

	agentName, ok = agentNameMap[agentID]
	if !ok {
		err = errors.New("agent name not found")
		return "", err
	}

	return
}
