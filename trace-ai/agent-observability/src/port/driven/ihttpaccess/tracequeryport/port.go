package tracequeryport

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/kweaver-core/trace-ai/agent-observability/src/domain/valueobject/opensearchvo"
)

type TraceQueryPort interface {
	SearchTraces(ctx context.Context, query json.RawMessage) (opensearchvo.SearchResult, error)
}
