package drivenadapters

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	os "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type OpenSearch interface {
	BulkUpsert(ctx context.Context, index string, documents []map[string]any, settings map[string]interface{}, mappings map[string]interface{}) error
	CreateIndex(ctx context.Context, index string, settings map[string]interface{}, mappings map[string]interface{}) error
	IndexExists(ctx context.Context, index string) (bool, error)
}

var (
	openSearchOnce     sync.Once
	openSearchInstance OpenSearch
)

func NewOpenSearch() OpenSearch {
	openSearchOnce.Do(func() {
		config := common.NewConfig()
		address := fmt.Sprintf("%s://%s:%s", config.OpenSearch.Protocol, config.OpenSearch.Host, config.OpenSearch.Port)
		retryBackoff := backoff.NewExponentialBackOff()
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second, // 连接超时时间
				KeepAlive: 60 * time.Second, // 保持长连接的时间
			}).DialContext, // 设置连接的参数
			MaxIdleConns:          1000,             // 最大空闲连接
			IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
			ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
			MaxIdleConnsPerHost:   500,              // 每个host保持的空闲连接数
			TLSHandshakeTimeout:   30 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		client, _ := os.NewClient(os.Config{
			Addresses: []string{
				address,
			},
			Username:      config.OpenSearch.User,
			Password:      config.OpenSearch.Password,
			Transport:     transport,
			RetryOnStatus: []int{502, 503, 504, 429},
			RetryBackoff: func(attempt int) time.Duration {
				if attempt == 1 {
					retryBackoff.Reset()
				}
				return retryBackoff.NextBackOff()
			},
			MaxRetries: 1,
		})

		openSearchInstance = &openSearch{
			client: client,
		}
	})

	return openSearchInstance
}

type openSearch struct {
	client *os.Client
}

func (o *openSearch) IndexExists(ctx context.Context, index string) (bool, error) {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, o.client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return false, nil
	}
	if res.IsError() {
		return false, fmt.Errorf("check index exists err: %s", res.String())
	}

	return true, nil
}

func (o *openSearch) BulkUpsert(ctx context.Context, index string, documents []map[string]any, settings map[string]interface{}, mappings map[string]interface{}) error {
	var body strings.Builder
	var failedEncodeDocs []int
	var encodeErrs []error

	// 如果提供了 settings 和 mappings，检查索引是否存在，不存在则创建
	if settings != nil && mappings != nil {
		exists, err := o.IndexExists(ctx, index)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("bulk upsert err: failed to check index existence: %s", err.Error())
			return fmt.Errorf("bulk upsert err: failed to check index existence: %w", err)
		}

		if !exists {
			traceLog.WithContext(ctx).Infof("bulk upsert: index %s does not exist, creating it", index)
			if err := o.CreateIndex(ctx, index, settings, mappings); err != nil {
				traceLog.WithContext(ctx).Warnf("bulk upsert err: failed to create index: %s", err.Error())
				return fmt.Errorf("bulk upsert err: failed to create index: %w", err)
			}
		}
	}

	for i, doc := range documents {
		id, ok := doc["__id"]

		if ok && id != nil {
			if _, isString := id.(string); !isString {
				id = fmt.Sprintf("%v", id)
			}
		}

		var metaBytes []byte
		var docBytes []byte
		var err error

		if !ok || id == "" {
			meta := map[string]any{
				"create": map[string]any{
					"_index": index,
				},
			}
			metaBytes, err = json.Marshal(meta)
			if err != nil {
				traceLog.WithContext(ctx).Warnf("bulk upsert marshal meta failed, doc %d: %v", i, err)
				failedEncodeDocs = append(failedEncodeDocs, i)
				encodeErrs = append(encodeErrs, fmt.Errorf("[doc %d meta] %w", i, err))
				continue
			}
			docBytes, err = json.Marshal(doc)
		} else {
			meta := map[string]any{
				"update": map[string]any{
					"_index": index,
					"_id":    id,
				},
			}
			metaBytes, err = json.Marshal(meta)
			if err != nil {
				traceLog.WithContext(ctx).Warnf("bulk upsert marshal meta failed, doc %d: %v", i, err)
				failedEncodeDocs = append(failedEncodeDocs, i)
				encodeErrs = append(encodeErrs, fmt.Errorf("[doc %d meta] %w", i, err))
				continue
			}
			docBytes, err = json.Marshal(map[string]any{
				"doc":           doc,
				"doc_as_upsert": true,
			})
		}

		if err != nil {
			traceLog.WithContext(ctx).Warnf("bulk upsert marshal doc failed, doc %d: %v", i, err)
			failedEncodeDocs = append(failedEncodeDocs, i)
			encodeErrs = append(encodeErrs, fmt.Errorf("[doc %d] %w", i, err))
			continue
		}

		body.Write(metaBytes)
		body.WriteString("\n")
		body.Write(docBytes)
		body.WriteString("\n")
	}

	if body.Len() == 0 {
		traceLog.WithContext(ctx).Warnf("bulk upsert err: empty document")
		if len(encodeErrs) > 0 {
			return fmt.Errorf("bulk upsert err: all documents encoding failed: %v", encodeErrs)
		}
		return fmt.Errorf("bulk upsert err: empty document")
	}

	req := opensearchapi.BulkRequest{
		Body: strings.NewReader(body.String()),
	}

	res, err := req.Do(ctx, o.client)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("bulk upsert err: %s", err.Error())
		return fmt.Errorf("bulk upsert request err: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		traceLog.WithContext(ctx).Warnf("bulk upsert err: %s", res.String())
		return fmt.Errorf("bulk upsert err: %s", res.String())
	}

	// 检查OpenSearch Bulk API响应中的单条失败项
	var bulkResp struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Status int             `json:"status"`
			Error  json.RawMessage `json:"error"`
		} `json:"items"`
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&bulkResp)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("bulk upsert fail to unmarshal response: %v", err)
		return fmt.Errorf("bulk upsert fail to unmarshal response: %w", err)
	}

	if bulkResp.Errors {
		failedItems := []string{}
		for i, item := range bulkResp.Items {
			for _, op := range item {
				if op.Status > 299 {
					errMsg := string(op.Error)
					failedItems = append(failedItems, fmt.Sprintf("[docIdx %d, status %d, err: %s]", i, op.Status, errMsg))
				}
			}
		}
		if len(failedItems) > 0 {
			errMsg := fmt.Sprintf("bulk upsert: %d items failed: %v", len(failedItems), failedItems)
			traceLog.WithContext(ctx).Warnf(errMsg)
			if len(encodeErrs) > 0 {
				return fmt.Errorf("%s; encoding failed docs: %v, errors: %v", errMsg, failedEncodeDocs, encodeErrs)
			}
			return fmt.Errorf("%s", errMsg)
		}
	}

	// 编码失败也要抛出异常
	if len(encodeErrs) > 0 {
		return fmt.Errorf("bulk upsert partial doc encoding failed: %v, errors: %v", failedEncodeDocs, encodeErrs)
	}

	return nil
}

func (o *openSearch) CreateIndex(ctx context.Context, index string, settings map[string]interface{}, mappings map[string]interface{}) error {
	indexBody := map[string]interface{}{
		"settings": settings,
		"mappings": mappings,
	}

	bodyBytes, err := json.Marshal(indexBody)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("create index err: failed to marshal index body: %v", err)
		return fmt.Errorf("create index err: failed to marshal index body: %w", err)
	}

	req := opensearchapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(string(bodyBytes)),
	}

	res, err := req.Do(ctx, o.client)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("create index err: %s", err.Error())
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		traceLog.WithContext(ctx).Warnf("create index err: %s", res.String())
		return fmt.Errorf("create index err: %s", res.String())
	}

	return nil
}
