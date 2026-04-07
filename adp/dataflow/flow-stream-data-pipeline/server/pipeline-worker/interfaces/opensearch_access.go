package interfaces

import (
	"context"

	osutil "github.com/opensearch-project/opensearch-go/v2/opensearchutil"
)

type BulkActionMetadata struct {
	Index           string  `json:"_index,omitempty"`
	DocumentID      string  `json:"_id,omitempty"`
	Routing         *string `json:"routing,omitempty"`
	RequireAlias    *bool   `json:"require_alias,omitempty"`
	RetryOnConflict *int    `json:"retry_on_conflict,omitempty"`
}

type BulkIndexer interface {
	BufLen() int
	ItemsLen() int
	Reset()
	Add(item *osutil.BulkIndexerItem) error
	Flush(ctx context.Context) ([]string, error)
}

//go:generate mockgen -source ../interfaces/opensearch_access.go -destination ../interfaces/mock/mock_opensearch_access.go
type OpenSearchAccess interface {
	NewBulkIndexer(flushBytes int) BulkIndexer
}
