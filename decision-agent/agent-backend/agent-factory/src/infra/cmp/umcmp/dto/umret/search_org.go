package umret

// SearchOrgRetDto 组织范围搜索的返回值
type SearchOrgRetDto struct {
	UserIDs       []string `json:"user_ids"`       // 符合条件的用户ID数组
	DepartmentIDs []string `json:"department_ids"` // 符合条件的部门ID数组
}
