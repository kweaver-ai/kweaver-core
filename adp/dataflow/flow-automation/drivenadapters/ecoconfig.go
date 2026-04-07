package drivenadapters

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

type Ecoconfig interface {
	Reindex(ctx context.Context, docID string, partType string) (int, error)
}

type EcoconfigImpl struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

var (
	ecoconfig     Ecoconfig
	ecoconfigOnce sync.Once
)

func NewEcoconfig() Ecoconfig {

	ecoconfigOnce.Do(func() {
		config := common.NewConfig()
		ecoconfig = &EcoconfigImpl{
			baseURL:    fmt.Sprintf("http://%s:%v", config.Ecoconfig.PrivateHost, config.Ecoconfig.PrivatePort),
			httpClient: NewOtelHTTPClient(),
		}
	})

	return ecoconfig
}

func (e *EcoconfigImpl) Reindex(ctx context.Context, docID string, partType string) (int, error) {
	log := traceLog.WithContext(ctx)

	target := fmt.Sprintf("%s/api/ecoconfig/v2/reindex", e.baseURL)

	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	code, _, err := e.httpClient.Post(ctx, target, headers, []map[string]interface{}{{
		"doc_id":    docID,
		"part_type": partType,
	}})

	if err != nil {
		log.Warnf("Reindex failed %v", err)
	}
	return code, err
}
