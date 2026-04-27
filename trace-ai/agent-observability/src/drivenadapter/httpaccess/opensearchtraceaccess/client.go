package opensearchtraceaccess

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/kweaver-core/trace-ai/agent-observability/src/domain/valueobject/opensearchvo"
	"github.com/kweaver-ai/kweaver-core/trace-ai/agent-observability/src/infra/opensearch"
)

type Client struct {
	openSearchClient *opensearch.Client
	traceIndex       string
}

func New(openSearchClient *opensearch.Client, traceIndex string) *Client {
	return &Client{
		openSearchClient: openSearchClient,
		traceIndex:       traceIndex,
	}
}

func (c *Client) SearchTraces(ctx context.Context, query json.RawMessage) (opensearchvo.SearchResult, error) {
	resp, err := c.openSearchClient.Search(ctx, c.traceIndex, query)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(resp), nil
}
