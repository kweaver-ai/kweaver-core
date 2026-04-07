package agentconfigreq

import "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/common"

type ListForBenchmarkReq struct {
	Name string `json:"name" form:"name"`
	common.PageSize

	// AgentIDsByBizDomain 用于业务域过滤，不从请求参数获取，由服务层设置
	AgentIDsByBizDomain []string `json:"-" form:"-"`
}
