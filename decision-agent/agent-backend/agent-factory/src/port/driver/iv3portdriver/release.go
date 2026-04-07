package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releaseresp"
)

//go:generate mockgen -source=./release.go -destination ./v3portdrivermock/release.go -package v3portdrivermock
type IReleaseSvc interface {
	Publish(ctx context.Context, req *releasereq.PublishReq) (res *releaseresp.PublishUpsertResp, auditloginfo auditlogdto.AgentPublishAuditLogInfo, err error)
	UnPublish(ctx context.Context, agentID string) (auditloginfo auditlogdto.AgentUnPublishAuditLogInfo, err error)

	GetPublishHistoryList(ctx context.Context, agentID string) (res releaseresp.HistoryListResp, total int64, err error)
	GetPublishHistoryInfo(ctx context.Context, req interface{}) (res string, err error)

	// 发布信息相关接口
	GetPublishInfo(ctx context.Context, agentID string) (res *releaseresp.PublishInfoResp, err error)
	UpdatePublishInfo(ctx context.Context, agentID string, req *releasereq.UpdatePublishInfoReq) (res *releaseresp.PublishUpsertResp,
		auditloginfo auditlogdto.AgentModifyPublishAuditLogInfo, err error)
}
