package umarg

import "net/http"

// SearchOrgArgDto 组织范围搜索的参数
type SearchOrgArgDto struct {
	UserIDs       []string `json:"user_ids"`       // 需要检查的用户ID数组
	DepartmentIDs []string `json:"department_ids"` // 需要检查的部门ID数组
	Scope         []string `json:"scope"`          // 范围数组，是部门ID数组
}

type SearchOrgUMArgDto struct {
	*SearchOrgArgDto
	Method string `json:"method"`
}

func NewSearchOrgUMArgDto(dto *SearchOrgArgDto) *SearchOrgUMArgDto {
	return &SearchOrgUMArgDto{
		SearchOrgArgDto: dto,
		Method:          http.MethodGet,
	}
}
