// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	"ontology-query/interfaces"
)

var (
	osAccessOnce sync.Once
	osAccess     interfaces.OpenSearchAccess
	//osAddress    string
)

type openSearchAccess struct {
	appSetting *common.AppSetting
	client     *opensearch.Client
}

func NewOpenSearchAccess(appSetting *common.AppSetting) interfaces.OpenSearchAccess {
	osAccessOnce.Do(func() {
		osAccess = &openSearchAccess{
			appSetting: appSetting,
			client:     rest.NewOpenSearchClient(appSetting.OpenSearchSetting),
		}
	})

	return osAccess
}

// CreateIndex 创建指定名称和配置的索引
// 根据提供的索引名称和body配置创建新的OpenSearch索引
// 参数：
//   - ctx: 上下文对象，用于控制请求生命周期
//   - indexName: 要创建的索引名称
//   - body: 索引配置，包括settings和mappings等
//
// 返回：创建成功返回nil，失败返回具体错误信息
func (o *openSearchAccess) CreateIndex(ctx context.Context, indexName string, body any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateIndex", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("index_name").String(indexName))

	// 将body转换为JSON字节
	bodyBytes, err := sonic.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal index body: %w", err)
	}

	// 创建索引请求
	req := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  bytes.NewBuffer(bodyBytes),
	}

	// 执行创建索引请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexName, err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("create index %s failed: %s, %s", indexName, res.Status(), res.String())
	}

	return nil
}

// IndexExists 检查指定索引是否存在
// 通过发送索引存在性检查请求来确定指定的索引是否已存在于OpenSearch中
// 参数：
//   - ctx: 上下文对象，用于控制请求生命周期
//   - indexName: 要检查的索引名称
//
// 返回：索引存在返回true，不存在返回false；发生错误时返回false和错误信息
// 示例：
//
//	exists, err := client.IndexExists(ctx, "my-index")
//	if err != nil {
//	    // 处理错误
//	}
//	if exists {
//	    // 索引已存在，可以跳过创建步骤
//	} else {
//	    // 索引不存在，需要创建
//	}
func (o *openSearchAccess) IndexExists(ctx context.Context, indexName string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "IndexExists", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("index_name").String(indexName))

	// 创建索引存在性检查请求
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	// 执行请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	// 根据响应状态码判断索引是否存在
	// 200 - 索引存在
	// 404 - 索引不存在
	// 其他状态码 - 错误
	if res.StatusCode == 200 {
		return true, nil
	} else if res.StatusCode == 404 {
		return false, nil
	} else {
		return false, fmt.Errorf("check index existence failed: %s, %s", res.Status(), res.String())
	}
}

func (o *openSearchAccess) DeleteIndex(ctx context.Context, indexName string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteIndex", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("index_name").String(indexName))

	// 创建删除索引请求
	req := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	// 执行删除索引请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("failed to delete index %s: %w", indexName, err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("delete index %s failed: %s, %s", indexName, res.Status(), res.String())
	}

	return nil
}

// InsertData 向指定索引写入数据，并指定文档ID
// 将单个文档数据插入到指定的OpenSearch索引中
// 参数：
//   - ctx: 上下文对象，用于控制请求生命周期
//   - indexName: 目标索引名称
//   - id: 文档的唯一标识符
//   - data: 要插入的文档数据，可以是任意可序列化的结构体或map
//
// 返回：插入成功返回nil，失败返回具体错误信息
// 注意：数据插入后会立即刷新索引，使数据立即可搜索
func (o *openSearchAccess) InsertData(ctx context.Context, indexName string, docID string, data any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "InsertData", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("index_name").String(indexName),
		attr.Key("doc_id").String(docID))

	// 将数据编码为JSON
	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// 创建索引请求，指定文档ID
	req := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(jsonData),
		Refresh:    "true", // 立即刷新，使数据可搜索
	}

	// 执行请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("failed to insert data with ID: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("insert data with ID failed: %s, %s", res.Status(), res.String())
	}

	return nil
}

// BulkInsertData 批量写入数据到指定索引
// 高效地将多个文档批量插入到指定的OpenSearch索引中
// 使用批量API可以显著提高大量数据的插入效率，比单条插入性能提升10-100倍
// 参数：
//   - ctx: 上下文对象，用于控制请求生命周期
//   - indexName: 目标索引名称
//   - dataList: 文档数据列表，每个元素必须包含"id"字段作为文档ID
//
// 返回：批量插入成功返回nil，失败返回具体错误信息
// 注意：数据插入后会立即刷新索引，使数据立即可搜索
// 性能：建议单次批量插入的文档数量控制在合理范围内（如1000-5000条）
// 示例：
//
//	dataList := []any{
//	  map[string]interface{}{"id": "doc1", "title": "文档1"},
//	  map[string]interface{}{"id": "doc2", "title": "文档2"},
//	}
func (o *openSearchAccess) BulkInsertData(ctx context.Context, indexName string, dataList []any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "BulkInsertData", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("index_name").String(indexName))

	if len(dataList) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, data := range dataList {
		// 准备元数据
		meta := map[string]any{
			"index": map[string]any{
				"_index": indexName,
				"_id":    data.(map[string]any)["__id"],
			},
		}

		// 写入元数据行
		metaJSON, err := sonic.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal bulk metadata: %w", err)
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// 写入数据行
		dataJSON, err := sonic.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal bulk data: %w", err)
		}
		buf.Write(dataJSON)
		buf.WriteByte('\n')
	}

	// 创建批量请求
	req := opensearchapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	// 执行请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("failed to bulk insert data: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk insert data failed: %s, %s", res.Status(), res.String())
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var resp struct {
		Took   int             `json:"took"`
		Errors bool            `json:"errors"`
		Items  json.RawMessage `json:"items"`
	}

	if err := sonic.Unmarshal(resBody, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if resp.Errors {
		return fmt.Errorf("bulk insert data failed: %s", resp.Items)
	}

	return nil
}

// SearchData 搜索指定索引中的数据
// 根据提供的查询条件在指定索引中执行搜索操作
// 支持复杂的查询DSL，包括全文搜索、过滤、聚合等
// 参数：
//   - ctx: 上下文对象，用于控制请求生命周期
//   - indexName: 要搜索的索引名称
//   - query: 查询条件，可以是OpenSearch查询DSL的任意结构
//
// 返回：搜索结果列表，每个元素是一个文档的完整内容；失败返回错误信息
// 示例：
//
//	query := map[string]interface{}{
//	  "query": map[string]interface{}{
//	    "match": map[string]interface{}{"title": "搜索关键词"},
//	  },
//	}
func (o *openSearchAccess) SearchData(ctx context.Context, indexName string, query any) ([]interfaces.Hit, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "SearchData", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("index_name").String(indexName))

	// 将查询条件编码为JSON
	queryJSON, err := sonic.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	logger.Debug(string(queryJSON))

	// 创建搜索请求
	req := opensearchapi.SearchRequest{
		Index: []string{indexName},
		Body:  bytes.NewReader(queryJSON),
	}

	// 执行请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return nil, fmt.Errorf("failed to search data: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search data failed: %s, %s", res.Status(), res.String())
	}

	// 解析响应
	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
				Sort   []any                  `json:"sort"`
				Score  float64                `json:"_score"`
			} `json:"hits"`
		} `json:"hits"`
	}

	cfg := sonic.Config{UseInt64: true}.Froze()
	if err := cfg.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// 提取搜索结果
	results := make([]interfaces.Hit, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		results = append(results, hit)
	}

	return results, nil
}

// DeleteData 删除指定索引中的单条数据
// 根据文档ID从指定索引中删除单个文档
// 如果文档不存在（404错误），不会返回错误
// 参数：
//   - ctx: 上下文对象，用于控制请求生命周期
//   - indexName: 目标索引名称
//   - id: 要删除的文档ID
//
// 返回：删除成功返回nil，失败返回具体错误信息
// 注意：删除操作会立即刷新索引，使删除结果立即可见
func (o *openSearchAccess) DeleteData(ctx context.Context, indexName string, docID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteData", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("index_name").String(indexName),
		attr.Key("doc_id").String(docID))

	req := opensearchapi.DeleteRequest{
		Index:      indexName,
		DocumentID: docID,
		Refresh:    "true", // 立即刷新，使删除操作立即可见
	}

	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("failed to delete data %s from index %s: %w", docID, indexName, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 404错误表示文档不存在，不视为错误
		if res.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("delete data %s from index %s failed: %s, %s", docID, indexName, res.Status(), res.String())
	}

	return nil
}

// BulkDeleteData 批量删除指定索引中的数据
// 高效地批量删除指定索引中的多个文档
// 使用批量API可以显著提高大量数据的删除效率
// 参数：
//   - ctx: 上下文对象，用于控制请求生命周期
//   - indexName: 目标索引名称
//   - idList: 要删除的文档ID列表
//
// 返回：批量删除成功返回nil，失败返回具体错误信息
// 注意：删除操作会立即刷新索引，使删除结果立即可见
// 性能：建议单次批量删除的文档数量控制在合理范围内（如1000-5000条）
// 容错：如果某个ID对应的文档不存在，不会影响其他文档的删除
func (o *openSearchAccess) BulkDeleteData(ctx context.Context, indexName string, docIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "BulkDeleteData", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("index_name").String(indexName))

	if len(docIDs) == 0 {
		return nil // 空列表直接返回，避免不必要的网络请求
	}

	var buf bytes.Buffer

	// 构建批量删除请求，每行包含删除操作元数据
	for _, docID := range docIDs {
		// 创建删除操作元数据（delete操作）
		action := map[string]interface{}{
			"delete": map[string]interface{}{
				"_index": indexName,
				"_id":    docID,
			},
		}

		// 写入操作元数据行
		actionBytes, err := sonic.Marshal(action)
		if err != nil {
			return fmt.Errorf("failed to marshal delete action: %w", err)
		}
		buf.Write(actionBytes)
		buf.WriteByte('\n')
	}

	// 创建批量请求，设置立即刷新使删除结果立即可见
	req := opensearchapi.BulkRequest{
		Body:    &buf,
		Refresh: "true", // 立即刷新，使删除操作立即可见
	}

	// 执行批量删除请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return fmt.Errorf("failed to bulk delete data: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("bulk delete data failed: %s, %s", res.Status(), res.String())
	}

	return nil
}

func (o *openSearchAccess) Count(ctx context.Context, indexName string, query any) ([]byte, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Count", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("index_name").String(indexName))

	// 将查询条件编码为JSON
	queryJSON, err := sonic.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// 创建搜索请求
	ignoreUnavailable := true
	req := opensearchapi.CountRequest{
		Index:             []string{indexName},
		Body:              bytes.NewReader(queryJSON),
		IgnoreUnavailable: &ignoreUnavailable,
	}

	// 执行请求
	res, err := req.Do(ctx, o.client)
	if err != nil {
		return nil, fmt.Errorf("failed to Count: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Count failed: %s, %s", res.Status(), res.String())
	}

	// 解析响应
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resBytes, nil
}
