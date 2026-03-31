package drivenadapters

import (
	"bytes"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	opensearch "github.com/opensearch-project/opensearch-go/v2"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

type openSearchAccess struct {
	appSetting *common.AppSetting
	client     *opensearch.Client
}

var (
	osAccessOnce sync.Once
	osAccess     interfaces.OpenSearchAccess
)

func NewOpenSearchAccess(appSetting *common.AppSetting) interfaces.OpenSearchAccess {
	osAccessOnce.Do(func() {
		osAccess = &openSearchAccess{
			appSetting: appSetting,
			client:     rest.NewOpenSearchClient(appSetting.OpenSearchSetting),
		}
	})

	return osAccess
}

func (osa *openSearchAccess) NewBulkIndexer(flushBytes int) interfaces.BulkIndexer {
	indexer := BulkIndexer{
		Client: osa.client,
		Buf:    bytes.NewBuffer(make([]byte, 0, flushBytes)),
		Aux:    make([]byte, 0, 512),
	}
	return &indexer
}
