package releasesvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/locale"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releaseresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetPublishInfo 获取发布信息
func (svc *releaseSvc) GetPublishInfo(ctx context.Context, agentID string) (res *releaseresp.PublishInfoResp, err error) {
	// 1. 检查Agent是否存在
	exists, err := svc.agentConfigRepo.ExistsByID(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "get agent by id failed, agentID: %s", agentID)
		return
	}

	if !exists {
		err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent not found")
		return
	}

	// 2. 获取发布信息
	releasePo, err := svc.releaseRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "get release by agent id failed, agentID: %s", agentID)
		return
	}

	if releasePo == nil {
		err = capierr.NewCustom404Err(ctx, apierr.ReleaseNotFound, "release not found")
		return
	}

	// 3. 构建响应
	res = releaseresp.NewPublishInfoResp()

	// 3.1 设置基本信息
	res.Description = releasePo.AgentDesc

	// 3.2 设置发布为标识
	res.SetPublishedToBes(releasePo)

	// 3.3 设置发布到标识
	res.SetPublishToWhere(releasePo)

	// 3.4 获取分类信息
	categoryRels, err := svc.releaseCategoryRelRepo.GetByReleaseID(ctx, releasePo.ID)
	if err != nil {
		err = errors.Wrapf(err, "get category relations failed, releaseID: %s", releasePo.ID)
		return
	}

	if len(categoryRels) > 0 {
		// 构建分类ID字符串
		categoryIDs := make([]string, 0, len(categoryRels))
		for _, rel := range categoryRels {
			categoryIDs = append(categoryIDs, rel.CategoryID)
		}

		// res.CategoryID = strings.Join(categoryIDs, ",")

		var categoryNameMap map[string]string

		categoryNameMap, err = svc.categoryRepo.GetIDNameMap(ctx, categoryIDs)
		if err != nil {
			err = errors.Wrapf(err, "get category name map failed, categoryIDs: %v", categoryIDs)
			return
		}

		for _, rel := range categoryRels {
			res.Categories = append(res.Categories, &releaseresp.CategoryInfo{
				ID:   rel.CategoryID,
				Name: categoryNameMap[rel.CategoryID],
			})
		}
	}

	// 3.5 获取权限信息

	if releasePo.IsPmsCtrlBool() {
		res.PmsControl = releaseresp.NewPmsControlResp()

		// --start 原有代码：从repo获取权限信息 --
		var permissions []*dapo.ReleasePermissionPO

		permissions, err = svc.releasePermissionRepo.GetByReleaseID(ctx, releasePo.ID)
		if err != nil {
			err = errors.Wrapf(err, "get permissions failed, releaseID: %s", releasePo.ID)
			return
		}

		// 3.6 处理权限信息
		if len(permissions) > 0 {
			res.PmsControl, err = svc.genPmsControlResp(ctx, permissions)
			if err != nil {
				return
			}
		}
		// --end 原有代码：从repo获取权限信息 --
		// --start 新代码：使用GetPolicyOfAgentUse获取权限策略 --
		// res.PmsControl, err = svc.genPmsControlRespFromPolicy(ctx, agentID)
		//if err != nil {
		//	return
		//}
		// --end 新代码：使用GetPolicyOfAgentUse获取权限策略 --
	}

	return
}

func (svc *releaseSvc) genPmsControlResp(ctx context.Context, pos []*dapo.ReleasePermissionPO) (resp *releaseresp.PmsControlResp, err error) {
	resp = releaseresp.NewPmsControlResp()

	// 1. 收集需要查询的ID
	roleIds := make([]string, 0, len(pos))
	userIds := make([]string, 0, len(pos))
	userGroupIds := make([]string, 0, len(pos))
	departmentIds := make([]string, 0, len(pos))
	appAccountIds := make([]string, 0, len(pos))

	for _, po := range pos {
		switch po.ObjectType {
		case cenum.PmsTargetObjTypeRole:
			roleIds = append(roleIds, po.ObjectId) //nolint:staticcheck // SA4010 暂忽略
		case cenum.PmsTargetObjTypeUser:
			userIds = append(userIds, po.ObjectId)
		case cenum.PmsTargetObjTypeUserGroup:
			userGroupIds = append(userGroupIds, po.ObjectId)
		case cenum.PmsTargetObjTypeDep:
			departmentIds = append(departmentIds, po.ObjectId)
		case cenum.PmsTargetObjTypeAppAccount:
			appAccountIds = append(appAccountIds, po.ObjectId)
		}
	}

	// 2. 查询用户信息
	arg := &umarg.GetOsnArgDto{
		UserIDs:       userIds,
		DepartmentIDs: departmentIds,
		GroupIDs:      userGroupIds,
		AppIDs:        appAccountIds,
	}

	ret := umtypes.NewOsnInfoMapS()

	if cenvhelper.IsLocalDev() {
		// 本地开发环境模拟数据
		for _, userID := range arg.UserIDs {
			ret.UserNameMap[userID] = userID + "_name"
		}

		for _, departmentID := range arg.DepartmentIDs {
			ret.DepartmentNameMap[departmentID] = departmentID + "_name"
		}

		for _, groupID := range arg.GroupIDs {
			ret.GroupNameMap[groupID] = groupID + "_name"
		}

		for _, appID := range arg.AppIDs {
			ret.AppNameMap[appID] = appID + "_name"
		}
	} else {
		ret, err = svc.umHttp.GetOsnNames(ctx, arg)
		if err != nil {
			return
		}
	}

	unknownUserName := locale.GetI18nByCtx(ctx, locale.UnknownUser)

	// 3. 构建响应
	for _, po := range pos {
		switch po.ObjectType {
		case cenum.PmsTargetObjTypeRole:
			resp.Roles = append(resp.Roles, comvalobj.RoleInfo{
				RoleID: po.ObjectId,
			})
		case cenum.PmsTargetObjTypeUser:
			userName, ok := ret.UserNameMap[po.ObjectId]
			if !ok {
				userName = unknownUserName
			}

			resp.Users = append(resp.Users, comvalobj.UserInfo{
				UserID:   po.ObjectId,
				Username: userName,
			})
		case cenum.PmsTargetObjTypeUserGroup:
			resp.UserGroups = append(resp.UserGroups, comvalobj.UserGroupInfo{
				UserGroupID:   po.ObjectId,
				UserGroupName: ret.GroupNameMap[po.ObjectId],
			})
		case cenum.PmsTargetObjTypeDep:
			resp.Departments = append(resp.Departments, comvalobj.DepartmentInfo{
				DepartmentID:   po.ObjectId,
				DepartmentName: ret.DepartmentNameMap[po.ObjectId],
			})
		case cenum.PmsTargetObjTypeAppAccount:
			resp.AppAccounts = append(resp.AppAccounts, comvalobj.AppAccountInfo{
				AppAccountID:   po.ObjectId,
				AppAccountName: ret.AppNameMap[po.ObjectId],
			})
		}
	}

	return
}

// genPmsControlRespFromPolicy 从权限策略生成权限控制响应
func (svc *releaseSvc) genPmsControlRespFromPolicy(ctx context.Context, agentID string) (resp *releaseresp.PmsControlResp, err error) {
	resp = releaseresp.NewPmsControlResp()

	// 1. 通过权限服务获取智能体使用权限策略
	policyRes, err := svc.pmsSvc.GetPolicyOfAgentUse(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "获取智能体[%s]权限策略失败", agentID)
		return
	}

	if policyRes == nil || len(policyRes.Entries) == 0 {
		// 没有策略条目，返回空的权限控制响应
		return
	}

	// 2. 构建响应
	for _, entry := range policyRes.Entries {
		if entry == nil || entry.Accessor == nil {
			continue
		}

		switch entry.Accessor.Type {
		case cenum.PmsTargetObjTypeRole:
			resp.Roles = append(resp.Roles, comvalobj.RoleInfo{
				RoleID:   entry.Accessor.ID,
				RoleName: entry.Accessor.Name,
			})
		case cenum.PmsTargetObjTypeUser:
			resp.Users = append(resp.Users, comvalobj.UserInfo{
				UserID:   entry.Accessor.ID,
				Username: entry.Accessor.Name,
			})
		case cenum.PmsTargetObjTypeUserGroup:
			resp.UserGroups = append(resp.UserGroups, comvalobj.UserGroupInfo{
				UserGroupID:   entry.Accessor.ID,
				UserGroupName: entry.Accessor.Name,
			})
		case cenum.PmsTargetObjTypeDep:
			resp.Departments = append(resp.Departments, comvalobj.DepartmentInfo{
				DepartmentID:   entry.Accessor.ID,
				DepartmentName: entry.Accessor.Name,
			})
		case cenum.PmsTargetObjTypeAppAccount:
			resp.AppAccounts = append(resp.AppAccounts, comvalobj.AppAccountInfo{
				AppAccountID:   entry.Accessor.ID,
				AppAccountName: entry.Accessor.Name,
			})
		}
	}

	return
}
