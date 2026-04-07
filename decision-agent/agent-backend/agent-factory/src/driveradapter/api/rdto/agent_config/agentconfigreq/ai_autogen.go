package agentconfigreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/pkg/errors"
)

// Params 表示 agent 自动生成内容的详细参数
type Params struct {
	Name    string   `json:"name"`    // agent名称
	Profile string   `json:"profile"` // agent简介
	Skills  []string `json:"skills"`  // agent技能列表
	Sources []string `json:"sources"` // agent知识来源列表
}

// ReqCheck 验证请求参数
func (p *Params) ReqCheck() (err error) {
	// 当全部为空时，返回错误
	if p.Name == "" && p.Profile == "" && len(p.Skills) == 0 && len(p.Sources) == 0 {
		err = errors.New("[Params]: name、profile、skills和sources至少有一个不能为空")
		return
	}

	return
}

// AiAutogenReq AI自动生成内容请求
type AiAutogenReq struct {
	Language string               `json:"-"`
	Params   *Params              `json:"params" binding:"required"`                                                    // agent详细参数
	From     daenum.AiAutogenFrom `json:"from" binding:"omitempty,oneof=system_prompt opening_remarks preset_question"` // 内容来源类型
	Stream   bool                 `json:"stream" binding:"omitempty"`                                                   // 是否流式输出

	UserID string `json:"-"` // 用户ID

	AccountType cenum.AccountType `json:"-"` // 账户类型
}

// GetErrMsgMap 获取错误信息映射
func (req *AiAutogenReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Params.required": "参数对象(params)不能为空",
		"From.oneof":      "内容来源类型无效，只能为preset_question、system_prompt或opening_remarks",
	}
}

func (req *AiAutogenReq) IsNotStream() bool {
	return req.From == daenum.AiAutogenFromPreSetQuestion
}

func (req *AiAutogenReq) ReqCheck() (err error) {
	if req.Params == nil {
		err = errors.New("[AiAutogenReq]: params is required")
		return
	}

	if err = req.Params.ReqCheck(); err != nil {
		err = errors.Wrap(err, "[AiAutogenReq]: params is invalid")
		return
	}

	if err = req.From.EnumCheck(); err != nil {
		err = errors.Wrap(err, "[AiAutogenReq]: from is invalid")
		return
	}

	return
}
