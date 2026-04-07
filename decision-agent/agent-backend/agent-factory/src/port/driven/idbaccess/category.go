package idbaccess

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/category.go github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess ICategoryRepo
type ICategoryRepo interface {
	GetByReleaseId(ctx context.Context, releaaseId string) (rt []*dapo.CategoryPO, err error)
	List(ctx context.Context, req interface{}) (rt []*dapo.CategoryPO, err error)

	GetIDNameMap(ctx context.Context, ids []string) (m map[string]string, err error)
}
