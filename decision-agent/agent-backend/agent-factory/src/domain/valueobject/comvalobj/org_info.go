package comvalobj

// RoleInfo 角色信息
type RoleInfo struct {
	RoleID   string `json:"role_id"`   // 角色ID
	RoleName string `json:"role_name"` // 角色名称
}

// UserInfo 用户信息
type UserInfo struct {
	UserID   string `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名称
}

// UserGroupInfo 用户组信息
type UserGroupInfo struct {
	UserGroupID   string `json:"user_group_id"`   // 用户组ID
	UserGroupName string `json:"user_group_name"` // 用户组名称
}

// DepartmentInfo 部门信息
type DepartmentInfo struct {
	DepartmentID   string `json:"department_id"`   // 部门ID
	DepartmentName string `json:"department_name"` // 部门名称
}

type AppAccountInfo struct {
	AppAccountID   string `json:"app_account_id"`   // 应用账号ID
	AppAccountName string `json:"app_account_name"` // 应用账号名称
}
