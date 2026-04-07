package umret

type IDName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetOsnRetDto 获取组织架构对象的names的返回值
// osn: org structure names
type GetOsnRetDto struct {
	UserNames       []IDName `json:"user_names"`
	DepartmentNames []IDName `json:"department_names"`
	GroupNames      []IDName `json:"group_names"`
	AppNames        []IDName `json:"app_names"`
}
