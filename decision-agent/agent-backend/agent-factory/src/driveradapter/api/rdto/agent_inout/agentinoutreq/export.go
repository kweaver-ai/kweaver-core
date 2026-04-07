package agentinoutreq

import (
	"errors"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// ExportReq 导出agent请求
type ExportReq struct {
	AgentIDs []string `json:"agent_ids" binding:"required" label:"要导出的agent ID列表"`
}

func NewExportReq() *ExportReq {
	return &ExportReq{
		AgentIDs: make([]string, 0),
	}
}

// GetErrMsgMap 返回错误信息映射
func (r *ExportReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"AgentIDs.required": `"agent_ids"字段不能为空`,
	}
}

// CustomCheckAndDedupl 自定义校验和去重
func (r *ExportReq) CustomCheckAndDedupl() error {
	// 1. 校验
	if len(r.AgentIDs) == 0 {
		return errors.New("[ExportReq][CustomCheck]: agent_ids不能为空")
	}

	// 2. 去重
	r.AgentIDs = cutil.DeduplGeneric[string](r.AgentIDs)

	// 3. 检查去重后是否还有数据
	if len(r.AgentIDs) == 0 {
		return errors.New("[ExportReq][CustomCheck]: 去重后agent_ids为空")
	}

	// 4. 校验单次导入最多导入xx个agent
	maxSize := daconstant.AgentInoutMaxSize
	if len(r.AgentIDs) > maxSize {
		return fmt.Errorf("[ExportReq][CustomCheck]: 单次导入最多导入%d个agent", maxSize)
	}

	return nil
}
