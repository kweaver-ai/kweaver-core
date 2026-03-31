package ibizdomainacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpres"
)

//go:generate mockgen -source=./bizdomain.go -destination ./bizdomainaccmock/bizdomain_mock.go -package bizdomainaccmock
type BizDomainHttpAcc interface {
	// 资源关联
	AssociateResource(ctx context.Context, req *bizdomainhttpreq.AssociateResourceReq) (err error)

	// 批量资源关联
	AssociateResourceBatch(ctx context.Context, req bizdomainhttpreq.AssociateResourceBatchReq) (err error)

	// 资源取消关联
	DisassociateResource(ctx context.Context, req *bizdomainhttpreq.DisassociateResourceReq) (err error)

	// 关联关系查询
	QueryResourceAssociations(ctx context.Context, req *bizdomainhttpreq.QueryResourceAssociationsReq) (res *bizdomainhttpres.QueryResourceAssociationsRes, err error)

	// 获取所有Agent ID列表
	GetAllAgentIDList(ctx context.Context, bdIDs []string) (agentIDs []string, agentID2BdIDMap map[string]string, err error)

	// 获取所有Agent Tpl ID列表
	GetAllAgentTplIDList(ctx context.Context, bdIDs []string) (agentIDs []string, err error)

	// 单个资源关联关系查询
	HasResourceAssociation(ctx context.Context, req *bizdomainhttpreq.QueryResourceAssociationSingleReq) (hasAssociation bool, err error)
}
