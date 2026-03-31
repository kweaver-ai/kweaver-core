package agentconfigreq

type TestTmpReq struct {
	TestFlag string `json:"test_flag" binding:"required"`

	Params interface{} `json:"params" binding:"required"`
}

func (p *TestTmpReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"TestFlag.required": "test_flag is required",
		"Params.required":   "params is required",
	}
}
