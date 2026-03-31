package umtypes

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umret"

type OsnInfoMapS struct {
	UserNameMap       map[string]string `json:"user_name_map"`
	DepartmentNameMap map[string]string `json:"department_name_map"`
	GroupNameMap      map[string]string `json:"group_name_map"`
	AppNameMap        map[string]string `json:"app_name_map"`
}

func NewOsnInfoMapS() *OsnInfoMapS {
	return &OsnInfoMapS{
		UserNameMap:       make(map[string]string),
		DepartmentNameMap: make(map[string]string),
		GroupNameMap:      make(map[string]string),
		AppNameMap:        make(map[string]string),
	}
}

func (p *OsnInfoMapS) FromGetOsnRetDto(dto *umret.GetOsnRetDto) {
	p.UserNameMap = make(map[string]string, len(dto.UserNames))
	for _, v := range dto.UserNames {
		p.UserNameMap[v.ID] = v.Name
	}

	p.DepartmentNameMap = make(map[string]string, len(dto.DepartmentNames))
	for _, v := range dto.DepartmentNames {
		p.DepartmentNameMap[v.ID] = v.Name
	}

	p.GroupNameMap = make(map[string]string, len(dto.GroupNames))
	for _, v := range dto.GroupNames {
		p.GroupNameMap[v.ID] = v.Name
	}

	p.AppNameMap = make(map[string]string, len(dto.AppNames))
	for _, v := range dto.AppNames {
		p.AppNameMap[v.ID] = v.Name
	}
}
