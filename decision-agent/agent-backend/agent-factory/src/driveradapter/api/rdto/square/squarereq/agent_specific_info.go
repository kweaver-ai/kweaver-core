package squarereq

type GetInfoType string

const (
	OutputConfGetInfoType GetInfoType = "output_conf"
)

type AgentSpecificInfoReq struct {
	GetInfoType GetInfoType `json:"get_info_type" binding:"required"`
}

func (p *AgentSpecificInfoReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"GetInfoType.required": `"get_info_type"不能为空`,
	}
}
