package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/other/otherreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/other/otherresp"
)

//go:generate mockgen -source=./other_svc.go -destination ./v3portdrivermock/other_svc.go -package v3portdrivermock
type IOtherSvc interface {
	DolphinTplList(ctx context.Context, req *otherreq.DolphinTplListReq) (*otherresp.DolphinTplListResp, error)
}
