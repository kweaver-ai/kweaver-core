package cenum

import "errors"

type OrgObjType string

// 类型 1-部门（或组织） 2-用户 3-用户组类型 1-部门（或组织） 2-用户 3-用户组
const (
	OrgObjTypeDep   OrgObjType = "dept"       // 部门（或组织）
	OrgObjTypeUser  OrgObjType = "user"       // 用户
	OrgObjTypeGroup OrgObjType = "user_group" // 用户组
)

func (e OrgObjType) String() string {
	switch e {
	case OrgObjTypeUser:
		return "user"
	case OrgObjTypeDep:
		return "dept"
	case OrgObjTypeGroup:
		return "user_group"
	}

	return ""
}

func (e OrgObjType) EnumCheck() error {
	if e != OrgObjTypeDep && e != OrgObjTypeUser && e != OrgObjTypeGroup {
		return errors.New("[OrgObjType][EnumCheck]: 无效的组织对象类型")
	}

	return nil
}
