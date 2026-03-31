package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/category/categoryresp"
)

//go:generate mockgen -source=./category.go -destination ./v3portdrivermock/category.go -package v3portdrivermock
type ICategorySvc interface {
	List(ctx context.Context) (res categoryresp.ListResp, err error)
}
