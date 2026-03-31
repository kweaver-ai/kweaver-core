package drivenadapters

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	osapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	osutil "github.com/opensearch-project/opensearch-go/v2/opensearchutil"

	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

type BulkIndexer struct {
	Client *opensearch.Client
	Buf    *bytes.Buffer
	Aux    []byte
	Items  []*osutil.BulkIndexerItem
}

func (bi *BulkIndexer) BufLen() int {
	return bi.Buf.Len()
}

func (bi *BulkIndexer) ItemsLen() int {
	return len(bi.Items)
}

func (bi *BulkIndexer) Reset() {
	bi.Items = bi.Items[:0]
	bi.Buf.Reset()
}

func (bi *BulkIndexer) Add(item *osutil.BulkIndexerItem) error {
	err := bi.WriteMeta(item)
	if err != nil {
		return err
	}

	err = bi.WriteBody(item)
	if err != nil {
		return err
	}

	bi.Items = append(bi.Items, item)
	return nil
}

func (bi *BulkIndexer) WriteMeta(item *osutil.BulkIndexerItem) error {
	var err error

	meta := interfaces.BulkActionMetadata{
		Index:           item.Index,
		DocumentID:      item.DocumentID,
		Routing:         item.Routing,
		RequireAlias:    item.RequireAlias,
		RetryOnConflict: item.RetryOnConflict,
	}
	bi.Aux, err = sonic.Marshal(map[string]interfaces.BulkActionMetadata{
		item.Action: meta,
	})
	if err != nil {
		return err
	}
	_, err = bi.Buf.Write(bi.Aux)
	if err != nil {
		return err
	}
	bi.Aux = bi.Aux[:0]
	_, err = bi.Buf.WriteRune('\n')
	if err != nil {
		return err
	}
	return nil
}

// WriteBody writes the item body to the buffer; it must be called under a lock.
func (bi *BulkIndexer) WriteBody(item *osutil.BulkIndexerItem) error {
	if item.Body != nil {
		_, err := bi.Buf.ReadFrom(item.Body)
		if err != nil {
			return err
		}

		_, err = item.Body.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		bi.Buf.WriteRune('\n')
	}
	return nil
}

// Flush writes out the worker buffer; it must be called under a lock.
func (bi *BulkIndexer) Flush(ctx context.Context) ([]string, error) {
	if bi.Buf.Len() < 1 {
		return []string{}, nil
	}

	var (
		err error
		blk osutil.BulkIndexerResponse
	)

	ln := bi.Buf.Len()
	old_buf := bi.Buf.Bytes()
	new_buf := make([]byte, ln)
	n := copy(new_buf, old_buf)
	if n != ln {
		return nil, fmt.Errorf("copy len error, wanted: %d, actual: %d", ln, n)
	}
	clonedBuffer := bytes.NewBuffer(new_buf)
	req := osapi.BulkRequest{
		Body: clonedBuffer,
	}

	res, err := req.Do(ctx, bi.Client)
	if err != nil {
		return nil, fmt.Errorf("Flush: do the request failed, error: %v", err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.IsError() {
		return nil, fmt.Errorf("Flush: response is error: %s", res.String())
	}

	err = decoder.NewStreamDecoder(res.Body).Decode(&blk)
	if err != nil {
		return nil, fmt.Errorf("Flush: error parsing response body: %s", err)
	}

	failedDocIds := make([]string, 0)
	for _, blkItem := range blk.Items {
		for _, info := range blkItem {

			if info.Error.Type != "" || info.Status > 201 {
				logger.Errorf("failed to flush indexer, doc id: %s, error: %v", info.DocumentID, info.Error)
				// 记录失败的文档 ID，后续写入 error topic
				failedDocIds = append(failedDocIds, info.DocumentID)
				// return fmt.Errorf("%v", info.Error)
			}
		}
	}

	if len(failedDocIds) > 0 {
		logger.Errorf("failed to flush indexer, Items size: %d, failed doc count: %d, failed doc ids: %v",
			len(bi.Items), len(failedDocIds), failedDocIds)

		return failedDocIds, fmt.Errorf("flush to opensearch failed, failed doc count: %d", len(failedDocIds))
	}

	logger.Debugf("success to flush indexer, Items size: %d", len(bi.Items))

	bi.Items = bi.Items[:0]
	bi.Buf.Reset()

	return []string{}, nil
}
