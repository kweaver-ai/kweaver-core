package agentconfigreq

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// CreateReq 表示创建agent的请求
type CreateReq struct {
	*UpdateReq

	Key string `json:"key" binding:"max=50"` // agent 标识

	IsSystemAgent *cenum.YesNoInt8 `json:"is_system_agent"` // 是否是系统agent
}

func (p *CreateReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Key.max": `"key"长度不能超过50`,
	}
}

func (p *CreateReq) D2e() (eo *daconfeo.DataAgent, err error) {
	// 1. 生成allowed_file_types
	err = HandleConfig(p.Config)
	if err != nil {
		err = errors.Wrap(err, "[CreateReq]: HandleConfig failed")
		return
	}

	// 2. dto to eo
	eo = &daconfeo.DataAgent{}

	err = cutil.CopyStructUseJSON(eo, p)
	if err != nil {
		return
	}

	// 3. 生成eo的key
	if eo.Key == "" {
		eo.Key = cutil.UlidMake()
	}

	// 4. d2e后处理
	D2eCommonAfterD2e(eo)

	return
}

func (p *CreateReq) ReqCheckWithCtx(ctx context.Context) (err error) {
	// 1. 验证update_req
	if err = p.UpdateReq.ReqCheckWithCtx(ctx); err != nil {
		err = errors.Wrap(err, "[CreateReq]: update_req is invalid")
		return
	}
	// 3. 验证is_private_api相关
	if !p.IsInternalAPI {
		if p.CreatedBy != "" {
			err = errors.New("[CreateReq]: created_by is valid when is_private_api is false")
		}
	} else {
		if p.CreatedBy == "" {
			err = errors.New("[CreateReq]: created_by is required when is_private_api is true")
		}
	}

	// 4. 验证is_built_in（内部接口有效，外部接口不允许设置）
	if !p.IsInternalAPI {
		if p.IsBuiltIn.IsBuiltIn() {
			err = errors.New("[UpdateReq]: is_built_in is invalid when is_private_api is false")
			return
		}
	}

	return
}
