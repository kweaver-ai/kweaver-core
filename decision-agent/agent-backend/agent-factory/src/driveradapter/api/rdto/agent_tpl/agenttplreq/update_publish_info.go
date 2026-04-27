package agenttplreq

type UpdatePublishInfoReq struct {
	CategoryIDs []string `json:"category_ids"`
}

func (p *UpdatePublishInfoReq) GetErrMsgMap() map[string]string {
	return map[string]string{}
}
