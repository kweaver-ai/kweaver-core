package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
)

//go:generate mockgen -source=./personal_space_svc.go -destination ./v3portdrivermock/personal_space_svc.go -package v3portdrivermock

// IPersonalSpaceService 个人空间服务接口
type IPersonalSpaceService interface {
	// AgentTplList 获取个人空间Agent模板列表
	AgentTplList(ctx context.Context, req *personalspacereq.AgentTplListReq) (resp *personalspaceresp.AgentTplListResp, err error)

	// AgentList 获取个人空间Agent列表
	AgentList(ctx context.Context, req *personalspacereq.AgentListReq) (resp *personalspaceresp.AgentListResp, err error)
}

// GetPersonalSpaceService 获取个人空间服务实例
func GetPersonalSpaceService() IPersonalSpaceService {
	// 这里需要避免循环依赖，通过延迟初始化解决
	// 实际实现会在personalspacesvc包中提供
	panic("GetPersonalSpaceService should be implemented in personalspacesvc package")
}
