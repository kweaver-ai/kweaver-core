package umtypes

type ParentDepInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // 固定为"department"
}

type DepartmentInfo struct {
	DepartmentId string           `json:"department_id"` // 部门ID
	Name         string           `json:"name"`          // 部门名称
	ParentDeps   []*ParentDepInfo `json:"parent_deps"`   // 父部门信息
}
