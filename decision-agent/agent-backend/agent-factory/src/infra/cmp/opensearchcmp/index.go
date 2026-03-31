package opensearchcmp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (o *OpsCmp) CreateIndex(ctx context.Context, index string, mapping, setting string) (err error) {
	// Create a map to hold the request body
	bodyMap := make(map[string]interface{})
	if mapping != "" {
		bodyMap["mappings"] = json.RawMessage(mapping)
	}

	if setting != "" {
		bodyMap["settings"] = json.RawMessage(setting)
	}

	// Convert the map to JSON
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %w", err)
	}

	req := opensearchapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(bodyBytes),
	}

	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	return nil
}

func (o *OpsCmp) IndexExists(ctx context.Context, index string) (bool, error) {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, o.client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

func (o *OpsCmp) DeleteIndex(ctx context.Context, index string) (err error) {
	req := opensearchapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, o.client)
	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.IsError() {
		err = fmt.Errorf("error deleting index: %s", res.String())
	}

	return
}
