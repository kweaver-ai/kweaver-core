package umret

// GetGroupMembersRetDto 获取用户组成员的返回值
type GetGroupMembersRetDto struct {
	UserIDs       []string `json:"user_ids"`       // 用户成员id数组
	DepartmentIDs []string `json:"department_ids"` // 部门成员id数组
}

func NewGetGroupMembersRetDto() *GetGroupMembersRetDto {
	return &GetGroupMembersRetDto{
		UserIDs:       make([]string, 0),
		DepartmentIDs: make([]string, 0),
	}
}
