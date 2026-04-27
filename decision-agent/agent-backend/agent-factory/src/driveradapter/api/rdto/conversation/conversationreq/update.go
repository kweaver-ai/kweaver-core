package conversationreq

// UpdateReq 表示更新agent的请求
type UpdateReq struct {
	ID         string `json:"id" binding:"required"`     // conversation id
	Title      string `json:"title"  binding:"required"` // conversation 标题
	TempareaId string `json:"temparea_id"`               // 临时区域ID
}

func (p *UpdateReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"ID.required":    `"name"不能为空`,
		"Title.required": `"title"不能为空`,
	}
}

func (p *UpdateReq) ReqCheck() (err error) {
	return nil
}
