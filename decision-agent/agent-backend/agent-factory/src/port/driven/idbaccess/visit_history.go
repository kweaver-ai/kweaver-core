package idbaccess

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/visit_history.go github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess IVisitHistoryRepo
type IVisitHistoryRepo interface {
	IncVisitCount(ctx context.Context, po *dapo.VisitHistoryPO) (err error)
}
