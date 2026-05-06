// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package opensearch provides OpenSearch/ElasticSearch connector implementation.
package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

type opensearchConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	IndexPattern string `mapstructure:"index_pattern"`
}

// OpenSearchConnector implements IndexConnector for OpenSearch/ElasticSearch.
type OpenSearchConnector struct {
	enabled bool
	Config  *opensearchConfig
	client  *opensearch.Client
}

func (c *OpenSearchConnector) ExecuteQueryWithDsl(ctx context.Context, resourceName string, dsl string) (*interfaces.QueryResult, error) {
	// Ensure the connector is enabled
	if !c.enabled {
		return nil, fmt.Errorf("OpenSearch connector is not enabled")
	}
	// Ensure we have a connection
	if err := c.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}
	// Validate DSL
	if dsl == "" {
		return nil, fmt.Errorf("DSL query is empty")
	}
	// Parse the DSL to ensure it's valid JSON
	var dslMap map[string]any
	if err := json.Unmarshal([]byte(dsl), &dslMap); err != nil {
		return nil, fmt.Errorf("invalid DSL JSON: %w", err)
	}

	// Execute search request with the provided DSL
	// resourceId is used as the index name
	req := opensearchapi.SearchRequest{
		Index: []string{resourceName},
		Body:  strings.NewReader(dsl),
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return nil, fmt.Errorf("search failed: %s", resp.String())
	}

	// Parse response
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	hits, ok := result["hits"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid search result format: missing hits")
	}

	// Extract total count
	var total int64
	if totalMap, ok := hits["total"].(map[string]any); ok {
		if value, ok := totalMap["value"].(float64); ok {
			total = int64(value)
		} else if value, ok := totalMap["value"].(int64); ok {
			total = value
		}
	}

	hitsArray, ok := hits["hits"].([]any)
	if !ok {
		return &interfaces.QueryResult{
			Rows:  []map[string]any{},
			Total: total,
		}, nil
	}

	// Extract documents from hits
	documents := make([]map[string]any, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]any)
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]any)
		if !ok {
			// If _source is not present, create an empty map
			source = make(map[string]any)
		}

		// Add _id to the source
		source["_id"] = hitMap["_id"]

		// Add _score field if present
		if score, ok := hitMap["_score"].(float64); ok {
			source["_score"] = score
		}

		documents = append(documents, source)
	}

	return &interfaces.QueryResult{
		Rows:  documents,
		Total: total,
	}, nil
}

// NewOpenSearchConnector 创建 OpenSearch connector 构建器
func NewOpenSearchConnector() connectors.IndexConnector {
	return &OpenSearchConnector{}
}

// GetType returns the data source type.
func (c *OpenSearchConnector) GetType() string {
	return interfaces.ConnectorTypeOpenSearch
}

// GetName returns the data source name.
func (c *OpenSearchConnector) GetName() string {
	return interfaces.ConnectorTypeOpenSearch
}

// GetMode returns the connector mode.
func (c *OpenSearchConnector) GetMode() string {
	return interfaces.ConnectorModeLocal
}

// GetCategory returns the connector category.
func (c *OpenSearchConnector) GetCategory() string {
	return interfaces.ConnectorCategoryIndex
}

// GetEnabled returns the enabled status.
func (c *OpenSearchConnector) GetEnabled() bool {
	return c.enabled
}

// SetEnabled sets the enabled status.
func (c *OpenSearchConnector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// GetSensitiveFields returns the sensitive fields for OpenSearch connector.
func (c *OpenSearchConnector) GetSensitiveFields() []string {
	return []string{"password"}
}

// GetFieldConfig returns the field configuration for OpenSearch connector.
func (c *OpenSearchConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return map[string]interfaces.ConnectorFieldConfig{
		"host":          {Name: "主机地址", Type: "string", Description: "OpenSearch 服务器主机地址", Required: true, Encrypted: false},
		"port":          {Name: "端口号", Type: "integer", Description: "OpenSearch 服务器端口", Required: true, Encrypted: false},
		"username":      {Name: "用户名", Type: "string", Description: "认证用户名", Required: false, Encrypted: false},
		"password":      {Name: "密码", Type: "string", Description: "认证密码", Required: false, Encrypted: true},
		"index_pattern": {Name: "索引模式", Type: "string", Description: "索引匹配模式（可选，如 log-*）", Required: false, Encrypted: false},
	}
}

// New creates a new OpenSearch connector.
func (c *OpenSearchConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	var osCfg opensearchConfig
	if err := mapstructure.Decode(cfg, &osCfg); err != nil {
		return nil, fmt.Errorf("failed to decode opensearch config: %w", err)
	}

	return &OpenSearchConnector{
		Config: &osCfg,
	}, nil
}

// Connect establishes connection to OpenSearch.
func (c *OpenSearchConnector) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil
	}

	cfg := opensearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%d", c.Config.Host, c.Config.Port)},
		Username:  c.Config.Username,
		Password:  c.Config.Password,
	}
	// TODO: Handle SSL/TLS options if needed

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create opensearch client: %w", err)
	}

	c.client = client
	return nil
}

// Close closes the connection.
func (c *OpenSearchConnector) Close(ctx context.Context) error {
	c.client = nil
	return nil
}

// Ping checks the connection.
func (c *OpenSearchConnector) Ping(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	req := opensearchapi.InfoRequest{}
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.IsError() {
		return fmt.Errorf("ping failed: %s", resp.String())
	}
	return nil
}

// TestConnection tests the connection to OpenSearch.
func (c *OpenSearchConnector) TestConnection(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	return c.Ping(ctx)
}

// GetMetadata returns the metadata for the catalog.
// GetMetadata 方法用于获取OpenSearch的元数据信息
// 参数:
//   - ctx: 上下文，用于控制请求的超时和取消
//
// 返回值:
//   - map[string]any: 包含OpenSearch元数据的键值对映射
//   - error: 如果操作过程中发生错误，返回相应的错误信息
func (c *OpenSearchConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	// 检查客户端是否已初始化连接
	if c.client == nil {
		return nil, fmt.Errorf("connector not connected")
	}

	// 创建OpenSearch信息请求
	req := opensearchapi.InfoRequest{}
	// 发送请求到OpenSearch服务器
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	// 确保响应体被关闭，以释放资源
	defer func() { _ = resp.Body.Close() }()
	// 检查响应是否包含错误
	if resp.IsError() {
		return nil, fmt.Errorf("get metadata failed: %s", resp.String())
	}

	// 用于存储解析后的元数据信息
	var info map[string]any
	// 将响应体中的JSON数据解码到info变量中
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	// 返回解析后的元数据信息
	return info, nil
}

// ListIndexes lists all indices.
func (c *OpenSearchConnector) ListIndexes(ctx context.Context) ([]*interfaces.IndexMeta, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	req := opensearchapi.CatIndicesRequest{
		Format: "json",
	}
	if c.Config.IndexPattern != "" {
		req.Index = []string{c.Config.IndexPattern}
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list indices: %s", resp.String())
	}

	var catIndices []struct {
		Index     string `json:"index"`
		DocsCount string `json:"docs.count"`
		StoreSize string `json:"store.size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&catIndices); err != nil {
		return nil, err
	}

	var indices []*interfaces.IndexMeta
	for _, idx := range catIndices {
		if strings.HasPrefix(idx.Index, ".") {
			continue // Skip system indices
		}

		indices = append(indices, &interfaces.IndexMeta{
			Name: idx.Index,
			Properties: map[string]any{
				"docs.count": idx.DocsCount,
				"store.size": idx.StoreSize,
			},
		})
	}
	return indices, nil
}

// GetIndexMeta retrieves index metadata (mappings, settings).
// GetIndexMeta 获取指定索引的元数据信息，包括映射和设置
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - index: 指向接口 IndexMeta 的指针，用于存储获取到的元数据
//
// 返回值:
//   - error: 如果操作过程中发生错误，则返回错误信息
func (c *OpenSearchConnector) GetIndexMeta(ctx context.Context, index *interfaces.IndexMeta) error {
	// 首先确保连接器已连接到 OpenSearch 服务
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// 检查索引的属性映射是否为空，如果为空则初始化一个空的 map
	if index.Properties == nil {
		index.Properties = make(map[string]any)
	}

	// 1. Get Mappings
	if err := c.fetchMappings(ctx, index); err != nil {
		return fmt.Errorf("failed to fetch mappings: %w", err)
	}

	// 2. Get Settings
	if err := c.fetchSettings(ctx, index); err != nil {
		return fmt.Errorf("failed to fetch settings: %w", err)
	}

	return nil
}

// fetchMappings retrieves and parses index mappings.
func (c *OpenSearchConnector) fetchMappings(ctx context.Context, index *interfaces.IndexMeta) error {
	req := opensearchapi.IndicesGetMappingRequest{
		Index: []string{index.Name},
	}
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return fmt.Errorf("opensearch API error: %s", resp.String())
	}
	//{
	//	"product_index" : {
	//	"mappings" : {
	//		"properties" : {
	//			"age" : {
	//				"type" : "integer"
	//			},
	//			"create_time" : {
	//				"type" : "date"
	//			},
	//			"description" : {
	//				"type" : "text",
	//				"fields" : {
	//					"keyword" : {
	//						"type" : "keyword",
	//						"ignore_above" : 256
	//					}
	//				}
	//			}
	//		}
	//	}
	//}
	//}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	// 解析 JSON
	err = json.Unmarshal(bodyBytes, &dataMapping)
	if err != nil {
		panic(err)
	}
	result := make(map[string]map[string]interface{})
	if idxData, ok := dataMapping[index.Name]; ok {
		parseProperties("", idxData.Mappings.Properties, result)
	}
	// Parse mappings:这里是存储的字段元数据，包括type映射
	fieldMap := make(map[string]interfaces.IndexFieldMeta)
	// 遍历：key 和 value 都拿到
	for fieldName, value := range result {
		// value: {"ignore_above":256,"type":"keyword"}
		fieldType, ok := value["type"].(string)
		if !ok {
			return fmt.Errorf("failed to read fieldType: indexName:%s,%w", index.Name, err)
		}
		fieldMap[fieldName] = interfaces.IndexFieldMeta{
			Name:       fieldName,
			Type:       MapType(fieldType),
			OrigType:   fieldType,
			Searchable: true, // Default to true for now
			Features:   value,
		}
	}
	index.Mapping = fieldMap
	return nil
}

// fetchSettings retrieves index settings.
func (c *OpenSearchConnector) fetchSettings(ctx context.Context, index *interfaces.IndexMeta) error {
	flatSettings := true
	req := opensearchapi.IndicesGetSettingsRequest{
		Index:        []string{index.Name},
		FlatSettings: &flatSettings,
	}
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return fmt.Errorf("opensearch API error: %s", resp.String())
	}

	var settingsResp map[string]struct {
		Settings map[string]any `json:"settings"`
	}
	//{
	//	"test-index" : {
	//	"settings" : {
	//		"index.creation_date" : "1772682337114",
	//			"index.number_of_replicas" : "1",
	//			"index.number_of_shards" : "1",
	//			"index.provided_name" : "test-index",
	//			"index.uuid" : "2G4vPna8SIC0vTEzZ0NK3Q",
	//			"index.version.created" : "136287827"
	//	}
	//}
	//}
	if err := json.NewDecoder(resp.Body).Decode(&settingsResp); err != nil {
		return err
	}
	if idxData, ok := settingsResp[index.Name]; ok {
		for k, v := range idxData.Settings {
			index.Properties[k] = v
		}
	}
	return nil
}

// fetchMappingsForQuery 获取索引的mapping信息并构建字段类型映射
func (c *OpenSearchConnector) fetchMappingsForQuery(ctx context.Context, indexName string, fieldTypeMap map[string]string) error {
	req := opensearchapi.IndicesGetMappingRequest{
		Index: []string{indexName},
	}
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return fmt.Errorf("opensearch API error: %s", resp.String())
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析 JSON
	var dataMapping map[string]struct {
		Mappings struct {
			Properties map[string]Property `json:"properties"`
		} `json:"mappings"`
	}
	if err := json.Unmarshal(bodyBytes, &dataMapping); err != nil {
		return fmt.Errorf("failed to unmarshal mappings: %w", err)
	}

	// 解析字段类型
	result := make(map[string]map[string]any)
	if idxData, ok := dataMapping[indexName]; ok {
		parseProperties("", idxData.Mappings.Properties, result)
	}

	// 构建字段类型映射
	for fieldName, value := range result {
		if fieldType, ok := value["type"].(string); ok {
			fieldTypeMap[fieldName] = MapType(fieldType)
		}
	}

	return nil
}

// ExecuteRawQuery executes a raw OpenSearch DSL query on the specified index.
func (c *OpenSearchConnector) ExecuteRawQuery(ctx context.Context, index string, query map[string]any) (*interfaces.SQLQueryResponse, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	// Convert query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Log the DSL query
	logger.Infof("[OpenSearch DSL Query] Index: %s, Query: %s", index, string(queryJSON))

	// Create search request
	req := opensearchapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(queryJSON)),
	}

	// Execute search
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("execute query failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return nil, fmt.Errorf("opensearch API error: %s", resp.String())
	}

	// Parse response
	var searchResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]any `json:"_source"`
				Sort   []any          `json:"sort"` // 添加sort字段
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// If no hits, return empty result
	if len(searchResp.Hits.Hits) == 0 {
		return &interfaces.SQLQueryResponse{
			Columns:    []interfaces.ColumnInfo{},
			Entries:    []map[string]any{},
			TotalCount: 0,
			Stats: interfaces.QueryStats{
				IsTimeout: false,
			},
		}, nil
	}

	// 获取索引的mapping信息以确定字段类型
	fieldTypeMap := make(map[string]string)
	if err := c.fetchMappingsForQuery(ctx, index, fieldTypeMap); err != nil {
		// 如果获取mapping失败，使用默认的string类型
		logger.Warnf("failed to fetch index mappings, using default string type: %v", err)
	}

	// Collect all field names from the first hit
	firstHit := searchResp.Hits.Hits[0].Source
	columns := make([]interfaces.ColumnInfo, 0, len(firstHit))
	for fieldName := range firstHit {
		fieldType := "string" // 默认类型
		if mappedType, ok := fieldTypeMap[fieldName]; ok {
			fieldType = mappedType
		}
		columns = append(columns, interfaces.ColumnInfo{
			Name: fieldName,
			Type: fieldType,
		})
	}

	// Convert hits to entries
	entries := make([]map[string]any, 0, len(searchResp.Hits.Hits))
	for _, hit := range searchResp.Hits.Hits {
		entries = append(entries, hit.Source)
	}

	// 构建响应
	// total_count设置为OpenSearch返回的总数据量
	totalCount := searchResp.Hits.Total.Value

	response := &interfaces.SQLQueryResponse{
		Columns:    columns,
		Entries:    entries,
		TotalCount: totalCount,
		Stats: interfaces.QueryStats{
			IsTimeout: false,
		},
	}

	// 如果有结果，检查是否需要返回search_after
	if len(searchResp.Hits.Hits) > 0 {
		lastHit := searchResp.Hits.Hits[len(searchResp.Hits.Hits)-1]
		// 如果最后一条记录有sort值，将其作为search_after返回
		if len(lastHit.Sort) > 0 {
			response.Stats.SearchAfter = lastHit.Sort
		}
	}

	return response, nil
}

// ExecuteQuery executes a query on the OpenSearch index.
// ExecuteQuery 执行OpenSearch查询并返回结果
// 参数:
//   - ctx: 上下文信息
//   - resource: 资源信息，包含索引名称等
//   - params: 查询参数，包括输出字段、排序、分页等
//
// 返回值:
//   - *interfaces.QueryResult: 查询结果，包含行数据和总数
//   - error: 错误信息
func (c *OpenSearchConnector) ExecuteQuery(ctx context.Context, indexName string, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {

	// Ensure we have a connection
	if err := c.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}

	if indexName == "" {
		return nil, fmt.Errorf("index name is empty in resource")
	}

	// 聚合查询：当Aggregation、GroupBy或Having任一参数存在时执行
	if params.Aggregation != nil || len(params.GroupBy) > 0 || params.Having != nil {
		// 构建OpenSearch聚合查询
		query := map[string]any{
			"size": 0, // 聚合查询不需要返回文档
		}

		// 处理过滤条件
		if params.ActualFilterCond != nil {
			filterQuery, err := c.ConvertFilterCondition(params.ActualFilterCond, resource.SchemaDefinition)
			if err != nil {
				return nil, fmt.Errorf("failed to build filter query: %w", err)
			}
			if filterQuery != nil {
				query["query"] = filterQuery
			}
		} else {
			query["query"] = map[string]any{
				"match_all": map[string]any{},
			}
		}

		// 构建聚合查询
		aggs := map[string]any{}

		// 确定聚合函数和别名
		var aggAlias string
		var metricBody map[string]any
		if params.Aggregation != nil {
			if params.Aggregation.Alias != "" {
				aggAlias = params.Aggregation.Alias
			} else {
				aggAlias = "__value"
			}

			aggField := params.Aggregation.Property
			aggFunc := params.Aggregation.Aggr

			switch aggFunc {
			case "count":
				metricBody = map[string]any{
					"value_count": map[string]any{
						"field": aggField,
					},
				}
			case "count_distinct":
				metricBody = map[string]any{
					"cardinality": map[string]any{
						"field": aggField,
					},
				}
			case "sum":
				metricBody = map[string]any{
					"sum": map[string]any{
						"field": aggField,
					},
				}
			case "avg":
				metricBody = map[string]any{
					"avg": map[string]any{
						"field": aggField,
					},
				}
			case "max":
				metricBody = map[string]any{
					"max": map[string]any{
						"field": aggField,
					},
				}
			case "min":
				metricBody = map[string]any{
					"min": map[string]any{
						"field": aggField,
					},
				}
			}
		}

		// 分组：自内向外嵌套 terms / date_histogram；度量与 HAVING 挂在最内层桶下。
		if len(params.GroupBy) > 0 {
			leafAggs := make(map[string]any)
			if metricBody != nil {
				leafAggs[aggAlias] = metricBody
			}
			if params.Having != nil && params.Aggregation != nil {
				leafAggs["having_filter"] = c.buildHavingBucketSelector(params.Having, aggAlias)
			}

			innerNode := leafAggs
			n := len(params.GroupBy)
			for i := n - 1; i >= 0; i-- {
				gb := params.GroupBy[i]
				name := "group_by_" + gb.Property
				var bucket map[string]any
				if gb.CalendarInterval != "" {
					bucket = map[string]any{
						"date_histogram": map[string]any{
							"field":             gb.Property,
							"calendar_interval": gb.CalendarInterval,
						},
					}
				} else {
					bucket = map[string]any{
						"terms": map[string]any{
							"field": gb.Property,
							"size":  nestedTermsSize(i, n, params.Limit),
						},
					}
				}
				if len(innerNode) > 0 {
					bucket["aggs"] = innerNode
				}
				innerNode = map[string]any{name: bucket}
			}
			for k, v := range innerNode {
				aggs[k] = v
			}
			// 对每一层 terms 应用 sort 映射到的 order（第二维度排序写在内层 terms 上）
			for _, v := range aggs {
				if node, ok := v.(map[string]any); ok {
					c.applyTermsOrderToGroupAggNode(node, params, aggAlias)
				}
			}
		} else if metricBody != nil {
			aggs[aggAlias] = metricBody
		}

		// 将聚合添加到查询
		query["aggs"] = aggs

		// 序列化查询
		queryJSON, err := json.Marshal(query)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize aggregate query: %w", err)
		}

		logger.Debugf("OpenSearch aggregate query: %s", string(queryJSON))

		// 执行搜索请求
		req := opensearchapi.SearchRequest{
			Index: []string{indexName},
			Body:  bytes.NewReader(queryJSON),
		}

		resp, err := req.Do(ctx, c.client)
		if err != nil {
			return nil, fmt.Errorf("failed to execute aggregate search: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.IsError() {
			return nil, fmt.Errorf("aggregate search failed: %s", resp.String())
		}

		// 读取响应体用于日志记录
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		logger.Debugf("OpenSearch aggregate response: %s", string(bodyBytes))

		// 解析响应
		var result map[string]any
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to decode aggregate search result: %w", err)
		}

		// 提取文档总数
		var totalCount int64
		if hits, ok := result["hits"].(map[string]any); ok {
			if totalMap, ok := hits["total"].(map[string]any); ok {
				if value, ok := totalMap["value"].(float64); ok {
					totalCount = int64(value)
				} else if value, ok := totalMap["value"].(int64); ok {
					totalCount = value
				}
			}
		}

		// 提取聚合结果
		aggregations, ok := result["aggregations"].(map[string]any)
		if !ok {
			return &interfaces.QueryResult{
				Rows:  []map[string]any{},
				Total: totalCount,
			}, nil
		}

		// 处理分组聚合结果（支持多层 group_by 嵌套桶展平）
		var rows []map[string]any
		if len(params.GroupBy) > 0 {
			groupByAggName := "group_by_" + params.GroupBy[0].Property
			if groupByAgg, ok := aggregations[groupByAggName].(map[string]any); ok {
				rows = c.flattenNestedGroupByRows(groupByAgg, params, aggAlias)
			}
		} else {
			// 没有分组，只有聚合
			if params.Aggregation != nil {
				row := make(map[string]any)
				if aggResult, ok := aggregations[aggAlias].(map[string]any); ok {
					if value, ok := aggResult["value"]; ok {
						row[aggAlias] = value
					}
				}
				rows = append(rows, row)
			}
		}

		return &interfaces.QueryResult{
			Rows:  rows,
			Total: totalCount,
		}, nil
	}

	// 明细查询
	// Build the OpenSearch query
	query := map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"from": 0,
		"size": 100,
	}

	// Handle output fields (_source)
	if params != nil && len(params.OutputFields) > 0 {
		// Filter out _score field as it's not a source field but a calculated score
		sourceFields := []string{}
		includeScore := false
		for _, field := range params.OutputFields {
			if field != "_score" {
				sourceFields = append(sourceFields, field)
			} else {
				includeScore = true
			}
		}
		if len(sourceFields) > 0 {
			query["_source"] = sourceFields
		}
		// Ensure track_scores is true to get _score when needed
		if includeScore {
			query["track_scores"] = true
		}
	}

	// Handle sorting
	if params != nil && len(params.Sort) > 0 {
		sort := make([]map[string]any, 0, len(params.Sort))
		for _, s := range params.Sort {
			keyword, _ := c.getKeywordSuffix(s.Field, resource.SchemaDefinition)
			sort = append(sort, map[string]any{
				s.Field + keyword: map[string]any{
					"order": s.Direction,
				},
			})
		}
		query["sort"] = sort
	}

	// Handle pagination
	if params != nil {
		if params.Offset > 0 && params.SearchAfter == nil {
			query["from"] = params.Offset
		}

		if params.Limit > 0 {
			query["size"] = params.Limit
		}

		// Handle search_after
		if len(params.SearchAfter) > 0 {
			query["search_after"] = params.SearchAfter
		}
	}

	// Handle filter conditions
	if params != nil && params.ActualFilterCond != nil {
		// Build filter condition query
		filterQuery, err := c.ConvertFilterCondition(params.ActualFilterCond, resource.SchemaDefinition)
		if err != nil {
			return nil, fmt.Errorf("failed to build filter query: %w", err)
		}
		if filterQuery != nil {
			query["query"] = filterQuery
		}
	}

	// Serialize query
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}
	logger.Debugf("Executing query: %s", string(queryJSON))

	// Execute search request
	req := opensearchapi.SearchRequest{
		Index: []string{indexName},
		Body:  bytes.NewReader(queryJSON),
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return nil, fmt.Errorf("search failed: %s", resp.String())
	}

	// Parse response
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	hits, ok := result["hits"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid search result format: missing hits")
	}

	total, ok := hits["total"].(map[string]any)["value"].(float64)
	if !ok {
		total = 0
	}

	hitsArray, ok := hits["hits"].([]any)
	if !ok {
		return &interfaces.QueryResult{
			Rows:  []map[string]any{},
			Total: int64(total),
		}, nil
	}

	// Extract documents from hits
	documents := make([]map[string]any, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]any)
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]any)
		if !ok {
			continue
		}

		source["_id"] = hitMap["_id"]
		// Add _score field if present
		if score, ok := hitMap["_score"].(float64); ok {
			source["_score"] = score
		}
		documents = append(documents, source)
	}

	return &interfaces.QueryResult{
		Rows:  documents,
		Total: int64(total),
	}, nil
}

// nestedTermsSize 为嵌套 group_by 中每一层 terms 设置 size：最内层用 limit 控制「每个父桶下」的行数，外层用较大上限以展开组合。
func nestedTermsSize(levelIndex, numLevels, limit int) int {
	if numLevels <= 1 {
		if limit > 0 {
			return limit
		}
		return 10
	}
	if levelIndex == numLevels-1 {
		if limit > 0 {
			return limit
		}
		return 10
	}
	outer := 1000
	if limit > 0 {
		if x := limit * 100; x > 10000 {
			outer = 10000
		} else if x < 100 {
			outer = 100
		} else {
			outer = x
		}
	}
	return outer
}

// applyTermsOrderToGroupAggNode 递归为子树中每个 terms 桶写入 order（多维度时第二维 sort 落在内层 terms）。
// 按度量排序仅在该 terms 的直接子 aggs 中包含度量名时生效，避免外层 terms 引用嵌套过深的子聚合导致 DSL 非法。
func (c *OpenSearchConnector) applyTermsOrderToGroupAggNode(node map[string]any, params *interfaces.ResourceDataQueryParams, aggAlias string) {
	if terms, ok := node["terms"].(map[string]any); ok {
		field, _ := terms["field"].(string)
		sub, _ := node["aggs"].(map[string]any)
		metricDirectChild := aggAlias != "" && sub != nil && sub[aggAlias] != nil

		var orderList []map[string]any
		for _, sortItem := range params.Sort {
			dir := strings.ToLower(sortItem.Direction)
			if dir != "asc" && dir != "desc" {
				dir = "asc"
			}
			if params.Aggregation != nil && metricDirectChild && (sortItem.Field == aggAlias || sortItem.Field == "__value") {
				orderList = append(orderList, map[string]any{aggAlias: dir})
			}
			if sortItem.Field == field {
				orderList = append(orderList, map[string]any{"_key": dir})
			}
		}
		if len(orderList) > 0 {
			terms["order"] = orderList
		}
	}
	sub, ok := node["aggs"].(map[string]any)
	if !ok {
		return
	}
	for name, child := range sub {
		if name == "having_filter" {
			continue
		}
		if childMap, ok := child.(map[string]any); ok {
			c.applyTermsOrderToGroupAggNode(childMap, params, aggAlias)
		}
	}
}

func (c *OpenSearchConnector) mergeMetricIntoRowFromBucket(bucket map[string]any, row map[string]any, aggAlias string) {
	if aggAlias == "" {
		return
	}
	if value, ok := bucket[aggAlias]; ok {
		if valueMap, ok := value.(map[string]any); ok {
			if val, ok := valueMap["value"]; ok {
				row[aggAlias] = val
			}
		} else {
			row[aggAlias] = value
		}
	}
}

// collectGroupByRowsFromBucket 自外层桶递归展开为多行（每行包含各维度键与可选度量）。
func (c *OpenSearchConnector) collectGroupByRowsFromBucket(bucket map[string]any, level int, params *interfaces.ResourceDataQueryParams, aggAlias string, rowSoFar map[string]any) []map[string]any {
	if level < 0 || level >= len(params.GroupBy) {
		return nil
	}
	gb := params.GroupBy[level]
	row := make(map[string]any, len(rowSoFar)+2)
	for k, v := range rowSoFar {
		row[k] = v
	}
	if key, ok := bucket["key"]; ok {
		row[gb.Property] = key
	} else if keyStr, ok := bucket["key_as_string"]; ok {
		row[gb.Property] = keyStr
	}

	if level == len(params.GroupBy)-1 {
		if params.Aggregation != nil {
			c.mergeMetricIntoRowFromBucket(bucket, row, aggAlias)
		}
		return []map[string]any{row}
	}

	nextName := "group_by_" + params.GroupBy[level+1].Property
	// OpenSearch bucket 的子聚合结果直接平铺在 bucket 下，而不是挂在 bucket["aggs"] 中。
	childAgg, ok := bucket[nextName].(map[string]any)
	if !ok {
		return []map[string]any{row}
	}
	nextBuckets, ok := childAgg["buckets"].([]any)
	if !ok {
		return []map[string]any{row}
	}
	var out []map[string]any
	for _, nb := range nextBuckets {
		nbm, ok := nb.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, c.collectGroupByRowsFromBucket(nbm, level+1, params, aggAlias, row)...)
	}
	return out
}

// flattenNestedGroupByRows 读取最外层 group_by 聚合并展平为结果行，最后按 limit 截断。
func (c *OpenSearchConnector) flattenNestedGroupByRows(rootAgg map[string]any, params *interfaces.ResourceDataQueryParams, aggAlias string) []map[string]any {
	buckets, ok := rootAgg["buckets"].([]any)
	if !ok {
		return []map[string]any{}
	}
	var rows []map[string]any
	for _, b := range buckets {
		bm, ok := b.(map[string]any)
		if !ok {
			continue
		}
		rows = append(rows, c.collectGroupByRowsFromBucket(bm, 0, params, aggAlias, nil)...)
	}
	if params.Limit > 0 && len(rows) > params.Limit {
		rows = rows[:params.Limit]
	}
	return rows
}

// buildHavingBucketSelector 构建HAVING条件的bucket_selector聚合
func (c *OpenSearchConnector) buildHavingBucketSelector(having *interfaces.HavingClause, aggAlias string) map[string]any {
	// OpenSearch使用bucket_selector聚合实现HAVING
	script := ""
	switch having.Operation {
	case "==":
		script = fmt.Sprintf("params.%s == %v", aggAlias, having.Value)
	case "!=":
		script = fmt.Sprintf("params.%s != %v", aggAlias, having.Value)
	case ">":
		script = fmt.Sprintf("params.%s > %v", aggAlias, having.Value)
	case ">=":
		script = fmt.Sprintf("params.%s >= %v", aggAlias, having.Value)
	case "<":
		script = fmt.Sprintf("params.%s < %v", aggAlias, having.Value)
	case "<=":
		script = fmt.Sprintf("params.%s <= %v", aggAlias, having.Value)
	case "in":
		if values, ok := having.Value.([]any); ok {
			script = fmt.Sprintf("%s.contains(params.%s.toString())", formatInValuesForScript(values), aggAlias)
		}
	case "not_in":
		if values, ok := having.Value.([]any); ok {
			script = fmt.Sprintf("!%s.contains(params.%s.toString())", formatInValuesForScript(values), aggAlias)
		}
	case "range":
		if values, ok := having.Value.([]any); ok && len(values) == 2 {
			script = fmt.Sprintf("params.%s >= %v && params.%s <= %v", aggAlias, values[0], aggAlias, values[1])
		}
	case "out_range":
		if values, ok := having.Value.([]any); ok && len(values) == 2 {
			script = fmt.Sprintf("params.%s < %v || params.%s > %v", aggAlias, values[0], aggAlias, values[1])
		}
	}

	return map[string]any{
		"bucket_selector": map[string]any{
			"buckets_path": map[string]any{
				aggAlias: aggAlias,
			},
			"script": map[string]any{
				"source": script,
			},
		},
	}
}

// formatInValuesForScript 格式化IN操作的值列表为Painless脚本格式
func formatInValuesForScript(values []any) string {
	if len(values) == 0 {
		return "[]"
	}

	var strValues []string
	for _, v := range values {
		switch val := v.(type) {
		case string:
			strValues = append(strValues, fmt.Sprintf("'%s'", val))
		default:
			strValues = append(strValues, fmt.Sprintf("%v", val))
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(strValues, ", "))
}

// buildFieldMappings 构建字段映射
func (c *OpenSearchConnector) buildFieldMappings(schemaDefinition []*interfaces.Property) (map[string]any, bool, error) {
	properties := map[string]any{}
	hasVectorField := false

	for _, column := range schemaDefinition {
		fieldType := column.Type
		switch column.Type {
		case "integer":
			fieldType = "long"
		case "unsigned_integer":
			fieldType = "unsigned_long"
		case "float":
			fieldType = "double"
		case "decimal":
			fieldType = "scaled_float"
		case "string":
			fieldType = "keyword"
		case "datetime":
			fieldType = "date"
		case "time":
			fieldType = "keyword"
		case "json":
			fieldType = "object"
		case "vector":
			hasVectorField = true
			fieldType = "knn_vector"
		case "point":
			fieldType = "geo_point"
		case "shape":
			fieldType = "geo_shape"
		default:
			// 保持 fieldType 不变
		}

		// 创建字段属性映射
		fieldProps := map[string]any{
			"type": fieldType,
		}

		// 为decimal类型添加scaling_factor参数
		if column.Type == "decimal" {
			fieldProps["scaling_factor"] = 1000000000000000000.0 // 18位小数
		}

		// 处理字段特性
		if column.Features != nil {
			for _, feature := range column.Features {
				if feature.Config != nil {
					switch feature.FeatureType {
					case "keyword":
						fieldsAdded := false
						for k, v := range feature.Config {
							if column.Type == "text" {
								if !fieldsAdded {
									// 添加子字段
									fieldProps["fields"] = map[string]any{
										feature.FeatureName: map[string]any{
											"type": "keyword",
										},
									}
									fieldsAdded = true
								}
								// 添加到子字段属性中
								if fields, ok := fieldProps["fields"].(map[string]any); ok {
									if subField, ok := fields[feature.FeatureName].(map[string]any); ok {
										subField[k] = v
									}
								}
							} else {
								// 直接添加到字段属性中
								fieldProps[k] = v
							}
						}
					case "vector":
						for k, v := range feature.Config {
							fieldProps[k] = v
						}
					case "fulltext":
						continue
					default:
						return nil, false, fmt.Errorf("unsupported feature type: %s", feature.FeatureType)
					}
				}
			}
		}

		properties[column.Name] = fieldProps
	}

	return properties, hasVectorField, nil
}

// Create index
func (c *OpenSearchConnector) Create(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	exist, err := c.indexExist(ctx, name)
	if err != nil {
		return err
	}
	// index exist
	if exist {
		return fmt.Errorf("index %s already exist", name)
	}

	// 构建字段映射
	properties, hasVectorField, err := c.buildFieldMappings(schemaDefinition)
	if err != nil {
		return err
	}

	mappings := map[string]any{
		"properties": properties,
	}

	mapping := map[string]any{
		"mappings": mappings,
	}

	mapping["settings"] = map[string]any{
		"index": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	// 如果有vector字段，开启knn
	if hasVectorField {
		indexSettings := mapping["settings"].(map[string]any)["index"].(map[string]any)
		indexSettings["knn"] = true
	}

	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	createReq := opensearchapi.IndicesCreateRequest{
		Index: name,
		Body:  bytes.NewReader(data),
	}

	createResp, err := createReq.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = createResp.Body.Close() }()

	if createResp.IsError() {
		return fmt.Errorf("failed to create index: %s", createResp.String())
	}

	return nil
}

// Update index.
func (c *OpenSearchConnector) Update(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	exist, err := c.indexExist(ctx, name)
	if err != nil {
		return err
	}
	// index not exist
	if !exist {
		return fmt.Errorf("index %s not exist", name)
	}

	// 构建字段映射
	properties, _, err := c.buildFieldMappings(schemaDefinition)
	if err != nil {
		return err
	}

	// 构建properties映射
	mappings := map[string]any{
		"properties": properties,
	}

	// 构建 JSON 字符串
	data, err := json.Marshal(mappings)
	if err != nil {
		return err
	}
	updateReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{name},
		Body:  bytes.NewReader(data),
	}
	updateResp, err := updateReq.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = updateResp.Body.Close() }()

	if updateResp.IsError() {
		return fmt.Errorf("failed to update index mapping: %s", updateResp.String())
	}

	return nil
}

// Delete a Dataset.
func (c *OpenSearchConnector) Delete(ctx context.Context, name string) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	exist, err := c.CheckExist(ctx, name)
	if err != nil {
		return err
	}
	// index not exist
	if !exist {
		return nil
	}

	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{name},
	}

	deleteResp, err := deleteReq.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = deleteResp.Body.Close() }()

	if deleteResp.IsError() {
		return fmt.Errorf("failed to delete index: %s", deleteResp.String())
	}

	return nil
}

// Check Index Exist
func (c *OpenSearchConnector) CheckExist(ctx context.Context, name string) (bool, error) {
	if err := c.Connect(ctx); err != nil {
		return false, err
	}

	return c.indexExist(ctx, name)
}

// Create Documents
func (c *OpenSearchConnector) CreateDocuments(ctx context.Context, name string, documents []map[string]any) ([]string, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	var bulkBody strings.Builder
	for _, doc := range documents {
		opMeta := map[string]map[string]string{
			"index": {
				"_index": name,
			},
		}
		// if _id in doc, use it as document id
		if docID, ok := doc["_id"].(string); ok {
			opMeta["index"]["_id"] = docID
			delete(doc, "_id")
		}

		if err := json.NewEncoder(&bulkBody).Encode(opMeta); err != nil {
			return nil, err
		}
		if err := json.NewEncoder(&bulkBody).Encode(doc); err != nil {
			return nil, err
		}
	}

	req := opensearchapi.BulkRequest{
		Body:    strings.NewReader(bulkBody.String()),
		Refresh: "true",
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return nil, fmt.Errorf("failed to create documents: %s", resp.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if errors, ok := result["errors"].(bool); ok && errors {
		// 遍历所有操作结果，检查是否有失败
		if items, ok := result["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if indexResult, ok := itemMap["index"].(map[string]interface{}); ok {
						if errorObj, ok := indexResult["error"].(map[string]interface{}); ok {
							// 找到失败的文档，返回错误
							return nil, fmt.Errorf("failed to create document, error type: %s, reason: %s", errorObj["type"].(string), errorObj["reason"].(string))
						}
					}
				}
			}
		}
	}

	var docIDs []string
	if items, ok := result["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if indexResult, ok := itemMap["index"].(map[string]interface{}); ok {
					if docID, ok := indexResult["_id"].(string); ok {
						docIDs = append(docIDs, docID)
					}
				}
			}
		}
	}

	return docIDs, nil
}

// Get Document
func (c *OpenSearchConnector) GetDocument(ctx context.Context, name string, docID string) (map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	req := opensearchapi.GetRequest{
		Index:      name,
		DocumentID: docID,
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get document: %s", resp.String())
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	source, ok := result["_source"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document not found")
	}

	source["_id"] = result["_id"]

	return source, nil
}

// Delete Document
func (c *OpenSearchConnector) DeleteDocument(ctx context.Context, name string, docID string) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	req := opensearchapi.DeleteRequest{
		Index:      name,
		DocumentID: docID,
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return fmt.Errorf("failed to delete document: %s", resp.String())
	}

	return nil
}

// Update Documents
func (c *OpenSearchConnector) UpsertDocuments(ctx context.Context, name string, updateRequests []map[string]any) ([]string, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	var bulkBody bytes.Buffer
	for _, updateReq := range updateRequests {
		docID, ok := updateReq["id"].(string)
		if !ok {
			continue
		}
		document := updateReq["document"]
		if document == nil {
			continue
		}

		metadata := map[string]map[string]string{
			"update": {
				"_index": name,
				"_id":    docID,
			},
		}
		if err := json.NewEncoder(&bulkBody).Encode(metadata); err != nil {
			return nil, err
		}

		// 写入更新操作的文档，添加upsert功能
		updateDoc := map[string]any{
			"doc":    document,
			"upsert": document, // 当文档不存在时，使用整个document作为新文档
		}
		if err := json.NewEncoder(&bulkBody).Encode(updateDoc); err != nil {
			return nil, err
		}
	}

	req := opensearchapi.BulkRequest{
		Body:    &bulkBody,
		Refresh: "true",
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return nil, fmt.Errorf("failed to update documents: %s", resp.String())
	}

	// 检查是否有部分文档更新失败
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var successDocIDs []string
	var errMsg string
	if items, ok := result["items"].([]interface{}); ok {
		for i, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if updateResult, ok := itemMap["update"].(map[string]interface{}); ok {
					if status, ok := updateResult["status"].(float64); ok {
						if status < 400 {
							// 提取成功的文档ID
							if docID, ok := updateRequests[i]["id"].(string); ok {
								successDocIDs = append(successDocIDs, docID)
							}
						} else {
							// 记录错误信息
							if errMsg == "" {
								errMsg = fmt.Sprintf("error type: %s, reason: %s", updateResult["error"].(map[string]interface{})["type"].(string), updateResult["error"].(map[string]interface{})["reason"].(string))
							}
						}
					}
				}
			}
		}
	}

	if errMsg != "" {
		return successDocIDs, fmt.Errorf("%s", errMsg)
	}

	return successDocIDs, nil
}

// Delete Documents
func (c *OpenSearchConnector) DeleteDocuments(ctx context.Context, name string, docIDs string) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	docIDList := strings.Split(docIDs, ",")

	var bulkBody bytes.Buffer
	for _, docID := range docIDList {
		docID = strings.TrimSpace(docID)
		if docID == "" {
			continue
		}

		metadata := map[string]map[string]string{
			"delete": {
				"_index": name,
				"_id":    docID,
			},
		}
		if err := json.NewEncoder(&bulkBody).Encode(metadata); err != nil {
			return err
		}
	}

	req := opensearchapi.BulkRequest{
		Body:    &bulkBody,
		Refresh: "true",
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return fmt.Errorf("failed to delete documents: %s", resp.String())
	}

	return nil
}

// Delete Documents By Query
func (c *OpenSearchConnector) DeleteDocumentsByQuery(ctx context.Context, name string, params *interfaces.ResourceDataQueryParams, schemaDefinition []*interfaces.Property) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	query := map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
	}

	if params != nil && params.ActualFilterCond != nil {
		filterQuery, err := c.ConvertFilterCondition(params.ActualFilterCond, schemaDefinition)
		if err != nil {
			return err
		}
		if filterQuery != nil {
			query["query"] = filterQuery
		}
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return err
	}

	refresh := true
	req := opensearchapi.DeleteByQueryRequest{
		Index:   []string{name},
		Body:    bytes.NewReader(queryBytes),
		Refresh: &refresh,
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.IsError() {
		return fmt.Errorf("failed to delete documents: %s", resp.String())
	}

	return nil
}

// index exist
func (c *OpenSearchConnector) indexExist(ctx context.Context, name string) (bool, error) {
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{name},
	}

	existsResp, err := existsReq.Do(ctx, c.client)
	if err != nil {
		return false, err
	}
	defer func() { _ = existsResp.Body.Close() }()

	return existsResp.StatusCode == 200, nil
}

// 映射结构定义
var dataMapping map[string]struct {
	Mappings struct {
		Properties map[string]Property `json:"properties"`
	} `json:"mappings"`
}

// Property 定义完整的字段属性
type Property struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"` // object 嵌套
	Fields     map[string]Property `json:"fields"`     // multi-fields 子字段
	// 使用 map[string]any 存储所有其他动态属性
	Attributes map[string]any `json:"-"`
}

// UnmarshalJSON 自定义反序列化方法
func (p *Property) UnmarshalJSON(data []byte) error {
	// 解析所有字段到一个临时的 map
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 初始化 Attributes
	if p.Attributes == nil {
		p.Attributes = make(map[string]any)
	}

	// 处理 type 字段
	if typeVal, ok := raw["type"]; ok {
		p.Type = fmt.Sprintf("%v", typeVal)
	}

	// 将除 type、properties、fields 之外的所有字段复制到 Attributes
	for key, value := range raw {
		switch key {
		case "properties", "fields":
			continue
		default:
			p.Attributes[key] = value
		}
	}
	// 处理 properties 字段（递归解析）
	if propsVal, ok := raw["properties"].(map[string]any); ok {
		p.Properties = make(map[string]Property)
		for propName, propValue := range propsVal {
			propJSON, _ := json.Marshal(propValue)
			var prop Property
			if err := json.Unmarshal(propJSON, &prop); err == nil {
				p.Properties[propName] = prop
			}
		}
	}
	// 处理 fields 字段（递归解析）
	if fieldsVal, ok := raw["fields"].(map[string]any); ok {
		p.Fields = make(map[string]Property)
		for fieldName, fieldValue := range fieldsVal {
			fieldJSON, _ := json.Marshal(fieldValue)
			var field Property
			if err := json.Unmarshal(fieldJSON, &field); err == nil {
				p.Fields[fieldName] = field
			}
		}
	}

	return nil
}

// 递归解析字段（支持 object 嵌套 + fields 子字段）
func parseProperties(parentPath string, props map[string]Property, result map[string]map[string]any) {
	for name, prop := range props {
		// 拼接完整路径：user.address.city
		currentPath := name
		if parentPath != "" {
			currentPath = parentPath + "." + name
		}
		// 如果不是 object 类型，输出完整字段属性
		if prop.Type != "object" {
			if len(prop.Attributes) > 0 {
				result[currentPath] = prop.Attributes
			}
		}
		// 解析 fields 子字段（title.keyword）
		if len(prop.Fields) > 0 {
			for fieldName, fieldProp := range prop.Fields {
				if len(fieldProp.Attributes) > 0 {
					result[currentPath+"."+fieldName] = fieldProp.Attributes
				}
			}
		}
		// 递归解析 object 嵌套字段
		if len(prop.Properties) > 0 {
			parseProperties(currentPath, prop.Properties, result)
		}
	}
}
