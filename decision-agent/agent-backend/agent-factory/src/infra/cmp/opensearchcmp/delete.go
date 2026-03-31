package opensearchcmp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// DeleteDocByField 根据字段值删除文档
func (o *OpsCmp) DeleteDocByField(ctx context.Context, index string, field string, value interface{}) (err error) {
	// 1. 构建查询条件
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				field: value,
			},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error marshalling query: %w", err)
	}

	// 2. 构建删除请求
	req := opensearchapi.DeleteByQueryRequest{
		Index: []string{index},
		Body:  bytes.NewReader(queryBytes),
	}

	// 3. 执行删除请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("error performing delete by query: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting documents: %s", res.String())
	}

	return nil
}

// DeleteDocsByFieldRange 根据字段范围删除文档，例如时间范围
func (o *OpsCmp) DeleteDocsByFieldRange(ctx context.Context, index string, field string, from, to interface{}) (err error) {
	// 1. 判断from和to是否都为空
	if from == nil && to == nil {
		return fmt.Errorf("both from and to are nil")
	}

	// 2. 处理from和to单个为空的情况
	var (
		f map[string]interface{}
	)

	f = map[string]interface{}{
		"gte": from,
		"lte": to,
	}

	if from == nil {
		f = map[string]interface{}{
			"lte": to,
		}
	}

	if to == nil {
		f = map[string]interface{}{
			"gte": from,
		}
	}

	// 3. 构建查询条件
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				field: f,
			},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error marshalling query: %w", err)
	}

	// 2. 构建删除请求
	req := opensearchapi.DeleteByQueryRequest{
		Index: []string{index},
		Body:  bytes.NewReader(queryBytes),
	}

	// 3. 执行删除请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("error performing delete by query: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting documents: %s", res.String())
	}

	return nil
}
