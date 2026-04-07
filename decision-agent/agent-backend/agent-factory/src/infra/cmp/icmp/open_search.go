package icmp

import (
	"context"
	"io"
	//"github.com/kweaver-ai/decision-agent/agent-factory/src/models"
)

//go:generate mockgen -package cmpmock -source open_search.go -destination ./cmpmock/open_search_mock.go
type IOpsCmp interface {
	Create(c context.Context, index, docID string, docReader io.Reader) (err error)

	CreateInterfaceNoID(ctx context.Context, index string, i interface{}) (err error)

	BatchCreate(ctx context.Context, index string, data []map[string]interface{}, isWithID bool) (err error)

	BatchCreateInterface(ctx context.Context, index string, docs interface{}, isWithID bool, batchSize int) (err error)

	CreateIndex(ctx context.Context, index string, mapping, setting string) (err error)

	DeleteIndex(ctx context.Context, index string) (err error)

	IndexExists(ctx context.Context, index string) (bool, error)

	DeleteDocByField(ctx context.Context, index string, field string, value interface{}) (err error)

	DeleteDocsByFieldRange(ctx context.Context, index string, field string, from, to interface{}) (err error)

	// Query(ctx context.Context, dslQuery string, index string) (*models.OSResp, error)
}
