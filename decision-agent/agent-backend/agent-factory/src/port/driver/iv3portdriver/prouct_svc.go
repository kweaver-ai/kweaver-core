package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productresp"
)

//go:generate mockgen -source=./prouct_svc.go -destination ./v3portdrivermock/prouct_svc.go -package v3portdrivermock
type IProductSvc interface {
	Create(ctx context.Context, req *productreq.CreateReq) (key string, err error)
	Update(ctx context.Context, req *productreq.UpdateReq, id int64) (auditloginfo auditlogdto.ProductUpdateAuditLogInfo, err error)
	Detail(ctx context.Context, id int64) (res *productresp.DetailRes, err error)
	Delete(ctx context.Context, id int64) (auditloginfo auditlogdto.ProductDeleteAuditLogInfo, err error)
	List(ctx context.Context, offset, limit int) (res *productresp.ListRes, err error)
	GetByKey(ctx context.Context, key string) (res *productresp.DetailRes, err error)
}
