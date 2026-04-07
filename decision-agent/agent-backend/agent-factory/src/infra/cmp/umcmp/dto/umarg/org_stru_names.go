package umarg

import (
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// GetOsnArgDto 获取组织架构对象的names的参数
// osn: org structure names
type GetOsnArgDto struct {
	UserIDs       []string `json:"user_ids,omitempty"`
	DepartmentIDs []string `json:"department_ids,omitempty"`
	GroupIDs      []string `json:"group_ids,omitempty"`
	AppIDs        []string `json:"app_ids,omitempty"`
}

// DeDupl 去重
func (d *GetOsnArgDto) DeDupl() {
	d.UserIDs = cutil.DeduplGeneric(d.UserIDs)
	d.DepartmentIDs = cutil.DeduplGeneric(d.DepartmentIDs)
	d.GroupIDs = cutil.DeduplGeneric(d.GroupIDs)
}

func (d *GetOsnArgDto) RemoveEmptyStrFromSlice() {
	d.UserIDs = cutil.RemoveEmptyStrFromSlice(d.UserIDs)
	d.DepartmentIDs = cutil.RemoveEmptyStrFromSlice(d.DepartmentIDs)
	d.GroupIDs = cutil.RemoveEmptyStrFromSlice(d.GroupIDs)
}

func (d *GetOsnArgDto) ToSfgKey() (key string, err error) {
	// 去重
	d.DeDupl()

	// 去除空字符串
	d.RemoveEmptyStrFromSlice()

	key, err = cutil.JSON().MarshalToString(d)

	return
}

type GetOsnUMArgDto struct {
	*GetOsnArgDto
	Method string `json:"method"`
}

func NewGetOsnUMArgDto(getOsnArgDto *GetOsnArgDto) *GetOsnUMArgDto {
	// 去重
	getOsnArgDto.DeDupl()

	// 去除空字符串
	getOsnArgDto.RemoveEmptyStrFromSlice()

	return &GetOsnUMArgDto{
		GetOsnArgDto: getOsnArgDto,
		Method:       http.MethodGet,
	}
}
