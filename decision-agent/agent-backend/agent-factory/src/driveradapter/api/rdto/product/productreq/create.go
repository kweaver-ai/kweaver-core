package productreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/producteo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// CreateReq 表示创建产品的请求
type CreateReq struct {
	Name    string `json:"name" binding:"required,max=50" example:"智能客服"`         // 产品名称
	Profile string `json:"profile" binding:"max=100" example:"这是一个智能客服产品"`        // 产品简介
	Key     string `json:"key" binding:"max=50" example:"smart-customer-service"` // 产品标识，唯一
}

func (p *CreateReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Name.required": `"name"不能为空`,
		"Name.max":      `"name"长度不能超过50`,
		"Profile.max":   `"profile"长度不能超过100`,
		"Key.max":       `"key"长度不能超过50`,
	}
}

func (p *CreateReq) D2e() (eo *producteo.Product, err error) {
	eo = &producteo.Product{}

	err = cutil.CopyStructUseJSON(eo, p)
	if err != nil {
		return
	}

	// 如果Key为空，生成一个唯一标识
	if eo.Key == "" {
		eo.Key = cutil.UlidMake()
	}

	return
}

func (p *CreateReq) CustomCheck() (err error) {
	return
}
