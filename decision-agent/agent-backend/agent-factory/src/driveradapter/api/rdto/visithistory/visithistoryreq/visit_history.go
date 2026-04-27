package visithistoryreq

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/common"
)

// VisitHistoryReq 访问日志列表请求
type VisitHistoryReq struct {
	UserID    string
	StartTime int64 // ms unix time
	EndTime   int64 // ms unix time
	common.PageSize
}

// GetErrMsgMap implements helpers.IErrMsgBindStruct.
