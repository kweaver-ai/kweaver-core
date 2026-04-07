package pmsvo

// PmsControlObjS 权限控制对象
// S means: Struct
type PmsControlObjS struct {
	RoleIDs       []string `json:"role_ids"`        // 角色ID列表
	UserIDs       []string `json:"user_ids"`        // 用户ID列表
	UserGroupIDs  []string `json:"user_group_ids"`  // 用户组ID列表
	DepartmentIDs []string `json:"department_ids"`  // 部门ID列表
	AppAccountIDs []string `json:"app_account_ids"` // 应用账号ID列表
}

// func NewPmsControlObjS() *PmsControlObjS {
//	return &PmsControlObjS{
//		RoleIDs:       make([]string, 0),
//		UserIDs:       make([]string, 0),
//		UserGroupIDs:  make([]string, 0),
//		DepartmentIDs: make([]string, 0),
//		AppAccountIDs: make([]string, 0),
//	}
//}
