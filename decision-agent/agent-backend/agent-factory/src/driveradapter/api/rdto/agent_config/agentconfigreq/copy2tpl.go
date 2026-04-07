package agentconfigreq

import (
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// Copy2TplReq 复制Agent为模板请求
type Copy2TplReq struct {
	Name string `json:"name"` // 新模板名称（可选，如果不提供则自动生成）
}

// GetErrMsgMap 获取错误信息映射
func (req *Copy2TplReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		// name字段是可选的，不需要required验证
	}
}

// ReqCheck 请求参数校验
func (req *Copy2TplReq) ReqCheck() error {
	// 名称长度校验（如果提供了名称）
	nameMaxLength := cconstant.NameMaxLength
	if req.Name != "" && cutil.RuneLength(req.Name) > nameMaxLength {
		return errors.New(fmt.Sprintf("模板名称长度不能超过%d个字符", nameMaxLength))
	}

	return nil
}
