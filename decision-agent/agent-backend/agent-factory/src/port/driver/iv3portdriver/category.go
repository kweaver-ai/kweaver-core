package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/category/categoryresp"
)

//go:generate mockgen -source=./category.go -destination ./v3portdrivermock/category.go -package v3portdrivermock
type ICategorySvc interface {
	List(ctx context.Context) (res categoryresp.ListResp, err error)
}
