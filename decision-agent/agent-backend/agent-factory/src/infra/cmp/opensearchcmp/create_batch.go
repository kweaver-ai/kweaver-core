package opensearchcmp

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (o *OpsCmp) BatchCreate(ctx context.Context, index string, docs []map[string]interface{}, isWithID bool) (err error) {
	defer func() {
		if err != nil {
			chelper.RecordErrLogWithPos(o.logger, err, "BatchCreate", "BatchCreate")
		}
	}()

	var buf bytes.Buffer

	for i, item := range docs {
		// 1. Write the meta data
		if isWithID {
			if id, ok := item["id"]; ok {
				meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s", "_id" : "%v" } }%s`, index, id, "\n"))
				buf.Write(meta)
			} else {
				err = fmt.Errorf("ID is required for docs item at index %d", i)
				return
			}
		} else {
			meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s" } }%s`, index, "\n"))
			buf.Write(meta)
		}

		// Remove the "id" field if present
		delete(item, "id")

		// 2. Write the docs
		var dataJSON []byte

		dataJSON, err = cutil.JSON().Marshal(item)
		if err != nil {
			err = fmt.Errorf("failed to marshal item at index %d: %s", i, err)
			return
		}

		buf.Write(dataJSON)
		buf.WriteString("\n")
	}

	// Print the request body for debugging
	o.logger.Debugf("[BatchCreateInterfaces]: Bulk Request Body:\n%s\n", buf.String())

	bulkRequest := opensearchapi.BulkRequest{
		Body: bytes.NewReader(buf.Bytes()),
	}

	res, err := bulkRequest.Do(ctx, o.client)
	if err != nil {
		err = fmt.Errorf("error executing bulk request: %s", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		err = fmt.Errorf("bulk request error: %s", res.String())
		return
	}

	return
}

// BatchCreateInterface If `batchSize` is 0 or negative, it calls the `BatchCreate` function with all docs at once; otherwise, it processes the docs in batches of the specified size.
func (o *OpsCmp) BatchCreateInterface(ctx context.Context, index string, docs interface{}, isWithID bool, batchSize int) (err error) {
	defer func() {
		if err != nil {
			chelper.RecordErrLogWithPos(o.logger, err, "BatchCreateInterface", "BatchCreateInterface")
		}
	}()

	// 1. interface{} to []map[string]interface{}
	// 1.1 interface{} to []byte
	bys, err := cutil.JSON().Marshal(docs)
	if err != nil {
		err = fmt.Errorf("failed to marshal docs: %s", err)
		return
	}

	// 1.2 []byte to []map[string]interface{}
	var maps []map[string]interface{}

	err = cutil.JSON().Unmarshal(bys, &maps)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal docs: %s", err)
		return
	}

	if batchSize <= 0 {
		// Create all in one batch if batchSize is 0 or negative
		err = o.BatchCreate(ctx, index, maps, isWithID)
		if err != nil {
			err = fmt.Errorf("failed to BatchCreate: %s", err)
			return
		}
	} else {
		// Create in batches
		for start := 0; start < len(maps); start += batchSize {
			end := start + batchSize
			if end > len(maps) {
				end = len(maps)
			}

			batch := maps[start:end]

			err = o.BatchCreate(ctx, index, batch, isWithID)
			if err != nil {
				err = fmt.Errorf("failed to BatchCreate in batch starting at index %d: %s", start, err)
				return
			}
		}
	}

	return
}
