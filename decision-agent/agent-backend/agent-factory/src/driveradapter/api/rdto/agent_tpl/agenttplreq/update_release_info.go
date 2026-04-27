package agenttplreq

type UpdateReleaseInfoReq struct {
	ReleaseNote string `json:"release_note"` // 发布说明
	Version     string `json:"version"`      // 版本号
}

func (p *UpdateReleaseInfoReq) GetErrMsgMap() map[string]string {
	return map[string]string{}
}
