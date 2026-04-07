package umarg

import "net/http"

// GetGroupMembersArgDto 获取组织架构对象的names的参数
// osn: org structure names
type GetGroupMembersArgDto struct {
	GroupIDs []string `json:"group_ids"`
}

type GetGroupMembersUMArgDto struct {
	*GetGroupMembersArgDto
	Method string `json:"method"`
}

func NewGetGroupMembersUMArgDto(getGroupMembersArgDto *GetGroupMembersArgDto) *GetGroupMembersUMArgDto {
	return &GetGroupMembersUMArgDto{
		GetGroupMembersArgDto: getGroupMembersArgDto,
		Method:                http.MethodGet,
	}
}
