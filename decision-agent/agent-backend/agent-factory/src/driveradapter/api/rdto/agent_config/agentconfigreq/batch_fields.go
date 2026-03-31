package agentconfigreq

import (
	"errors"
)

type BatchFieldsReqField string

const (
	BatchFieldsReqFieldName BatchFieldsReqField = "name"
)

func (f BatchFieldsReqField) String() string {
	return string(f)
}

func (f BatchFieldsReqField) ValObjCheck() (err error) {
	if f != BatchFieldsReqFieldName {
		err = errors.New("[BatchFieldsReqField] invalid field: " + string(f))
		return
	}

	return
}

// BatchFieldsReq 批量获取agent指定字段的请求
type BatchFieldsReq struct {
	AgentIDs []string              `json:"agent_ids" binding:"required" description:"agent id列表"`
	Fields   []BatchFieldsReqField `json:"fields" binding:"required" description:"agent字段列表，目前支持：name"`
}

// GetErrMsgMap 返回错误消息映射
func (req *BatchFieldsReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"AgentIDs.required": "agent_ids不能为空",
		"Fields.required":   "fields不能为空",
	}
}

// Validate 校验请求参数
func (req *BatchFieldsReq) Validate() (err error) {
	if len(req.AgentIDs) == 0 {
		err = errors.New("agent_ids不能为空")
		return
	}

	if len(req.Fields) == 0 {
		err = errors.New("fields不能为空")
		return
	}

	for _, field := range req.Fields {
		if err = field.ValObjCheck(); err != nil {
			return
		}
	}

	return nil
}
