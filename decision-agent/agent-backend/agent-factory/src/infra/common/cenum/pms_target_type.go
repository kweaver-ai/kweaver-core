package cenum

import "errors"

type PmsTargetObjType string

// 类型 1-部门（或组织） 2-用户 3-用户组类型 1-部门（或组织） 2-用户 3-用户组
const (
	PmsTargetObjTypeDep        PmsTargetObjType = "department" // 部门（或组织）
	PmsTargetObjTypeUser       PmsTargetObjType = "user"       // 用户
	PmsTargetObjTypeUserGroup  PmsTargetObjType = "group"      // 用户组
	PmsTargetObjTypeRole       PmsTargetObjType = "role"       // 角色
	PmsTargetObjTypeAppAccount PmsTargetObjType = "app"        // 应用账号
)

func (e PmsTargetObjType) String() string {
	return string(e)
}

func (e PmsTargetObjType) EnumCheck() error {
	if e != PmsTargetObjTypeDep && e != PmsTargetObjTypeUser && e != PmsTargetObjTypeUserGroup && e != PmsTargetObjTypeRole && e != PmsTargetObjTypeAppAccount {
		return errors.New("[PmsTargetObjType][EnumCheck]: 无效的权限目标对象类型")
	}

	return nil
}
