package iv3portdriver

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplresp"
)

//go:generate mockgen -source=./da_config_tpl_svc.go -destination ./v3portdrivermock/da_config_tpl_svc.go -package v3portdrivermock
type IDataAgentTplSvc interface {
	Update(ctx context.Context, req *agenttplreq.UpdateReq, id int64) (auditloginfo auditlogdto.AgentTemplateUpdateAuditLogInfo, err error)
	Detail(ctx context.Context, id int64) (res *agenttplresp.DetailRes, err error)
	DetailByKey(ctx context.Context, key string) (res *agenttplresp.DetailRes, err error)
	Delete(ctx context.Context, id int64, uid string, isPrivate bool) (auditloginfo auditlogdto.AgentTemplateDeleteAuditLogInfo, err error)

	// Publish 发布模板
	Publish(ctx context.Context, tx *sql.Tx, req *agenttplreq.PublishReq, id int64, isFromCopy2TplAndPublish bool) (res *agenttplresp.PublishUpsertResp, auditloginfo auditlogdto.AgentTemplatePublishAuditLogInfo, err error)
	// Unpublish 取消发布模板
	Unpublish(ctx context.Context, id int64) (auditloginfo auditlogdto.AgentTemplateUnpublishAuditLogInfo, err error)

	// GetPublishInfo 获取模板发布信息
	GetPublishInfo(ctx context.Context, id int64) (res *agenttplresp.PublishInfoRes, err error)
	// UpdatePublishInfo 更新模板发布信息
	UpdatePublishInfo(ctx context.Context, req *agenttplreq.UpdatePublishInfoReq, id int64) (res *agenttplresp.PublishUpsertResp, auditloginfo auditlogdto.AgentTemplateModifyPublishAuditLogInfo, err error)
	// Copy 复制智能体模板
	Copy(ctx context.Context, id int64) (res *agenttplresp.CopyResp, auditloginfo auditlogdto.AgentTemplateCopyAuditLogInfo, err error)
}
