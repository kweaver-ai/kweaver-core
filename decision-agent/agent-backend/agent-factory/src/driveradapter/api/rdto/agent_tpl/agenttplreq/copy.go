package agenttplreq

type CopyReq struct{}

func (p *CopyReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		// name字段现在是可选的，不需要required验证
	}
}
