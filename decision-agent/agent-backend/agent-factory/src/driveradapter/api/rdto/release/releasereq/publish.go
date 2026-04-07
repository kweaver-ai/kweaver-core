package releasereq

import (
	"github.com/pkg/errors"
)

// PublishReq 发布Agent 请求对象
type PublishReq struct {
	UserID  string
	AgentID string `json:"agent_id" binding:"required"`
	// AgentConfig string `json:"agent_config"` // 草稿配置

	BusinessDomainID string `json:"business_domain_id"` // 业务域ID

	*UpdatePublishInfoReq

	IsInternalAPI bool `json:"-"` // 是否是内部接口api。此字段不是前端传入，后端自动设置
}

func NewPublishReq() *PublishReq {
	return &PublishReq{
		UpdatePublishInfoReq: &UpdatePublishInfoReq{},
	}
}

// GetErrMsgMap implements helpers.IErrMsgBindStruct.
func (p *PublishReq) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"AgentID.required": `"Agent ID 不能为空`,
	}
}

func (p *PublishReq) ReqCheck() (err error) {
	if p.AgentID == "" {
		return errors.New("[PublishReq]: agent_id is required")
	}

	if err = p.UpdatePublishInfoReq.CustomCheck(); err != nil {
		return
	}

	return
}
