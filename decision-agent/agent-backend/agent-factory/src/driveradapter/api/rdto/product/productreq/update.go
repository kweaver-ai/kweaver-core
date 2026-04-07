package productreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/producteo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// UpdateReq 表示更新产品的请求
type UpdateReq struct {
	Name    string `json:"name" binding:"required,max=50" example:"智能客服"`  // 产品名称
	Profile string `json:"profile" binding:"max=100" example:"这是一个智能客服产品"` // 产品简介
}

func (p *UpdateReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Name.required": `"name"不能为空`,
		"Name.max":      `"name"长度不能超过50`,
		"Profile.max":   `"profile"长度不能超过100`,
	}
}

func (p *UpdateReq) D2e() (eo *producteo.Product, err error) {
	eo = &producteo.Product{}

	err = cutil.CopyStructUseJSON(eo, p)
	if err != nil {
		return
	}

	return
}

func (p *UpdateReq) CustomCheck() (err error) {
	return
}
