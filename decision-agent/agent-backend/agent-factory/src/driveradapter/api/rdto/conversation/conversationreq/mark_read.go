package conversationreq

type MarkReadReq struct {
	LastestReadIdx int `json:"latest_read_index" binding:"required"`
}

func (p *MarkReadReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"LastestReadIdx.required": `"latest_read_index"不能为空`,
	}
}

func (p *MarkReadReq) ReqCheck() (err error) {
	return nil
}
