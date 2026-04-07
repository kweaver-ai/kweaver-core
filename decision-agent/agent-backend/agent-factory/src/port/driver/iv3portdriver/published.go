package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedresp"
)

//go:generate mockgen -source=./published.go -destination ./v3portdrivermock/published.go -package v3portdrivermock
type IPublishedSvc interface {
	// 获取已发布智能体列表
	GetPublishedAgentList(ctx context.Context, req *pubedreq.PubedAgentListReq) (res *pubedresp.PubedAgentListResp, err error)

	// 获取已发布智能体模板列表
	GetPubedTplList(ctx context.Context, req *pubedreq.PubedTplListReq) (res *pubedresp.PublishedAgentTplListResp, err error)

	PubedTplDetail(ctx context.Context, publishedTplID int64) (res *pubedresp.DetailRes, err error)

	GetPubedAgentInfoList(ctx context.Context, req *pubedreq.PAInfoListReq) (res *pubedresp.PAInfoListResp, err error)
}
