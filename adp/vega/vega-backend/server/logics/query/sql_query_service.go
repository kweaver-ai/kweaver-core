// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package query

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/local/table/mariadb"
	"vega-backend/logics/connectors/local/table/postgresql"
	"vega-backend/logics/resource"

	"vega-backend/common"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/query/sqlglot"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"
)

var (
	sqlQueryServiceOnce     sync.Once
	sqlQueryServiceInstance interfaces.SQLQueryService
)

type sqlQueryService struct {
	cs interfaces.CatalogService
	rs interfaces.ResourceService
}

// NewSQLQueryService 创建SQL查询服务（单例模式）
func NewSQLQueryService(appSetting *common.AppSetting) interfaces.SQLQueryService {
	sqlQueryServiceOnce.Do(func() {
		sqlQueryServiceInstance = &sqlQueryService{
			cs: catalog.NewCatalogService(appSetting),
			rs: resource.NewResourceService(appSetting),
		}
	})
	return sqlQueryServiceInstance
}

// NewSQLQueryServiceWithDeps 创建SQL查询服务（用于测试）
func NewSQLQueryServiceWithDeps(cs interfaces.CatalogService, rs interfaces.ResourceService) interfaces.SQLQueryService {
	return &sqlQueryService{cs: cs, rs: rs}
}

// Execute 执行SQL查询
func (s *sqlQueryService) Execute(ctx context.Context, req *interfaces.SQLQueryRequest) (*interfaces.SQLQueryResponse, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "SQLQueryExecute")
	defer span.End()

	// 1. 校验请求
	if err := s.validateRequest(ctx, req); err != nil {
		span.SetStatus(codes.Error, "validate request failed")
		return nil, err
	}

	// 2. 判断查询类型
	// 如果是流式查询，调用executeStreamQuery方法
	if req.QueryType == "stream" {
		// OpenSearch流式查询直接调用executeOpenSearchQuery
		if req.ResourceType == "opensearch" {
			// 从query中获取resource_id
			queryMap, ok := req.Query.(map[string]any)
			if !ok {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
					WithErrorDetails("query must be a JSON object for opensearch queries")
			}

			var resourceID string
			if rid, ok := queryMap["resource_id"].(string); ok && rid != "" {
				resourceID = rid
			} else {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
					WithErrorDetails("resource_id is required for opensearch queries")
			}

			// 获取资源信息
			resource, err := s.rs.GetByID(ctx, resourceID)
			if err != nil {
				span.SetStatus(codes.Error, "get resource failed")
				return nil, err.(*rest.HTTPError)
			}
			if resource == nil {
				span.SetStatus(codes.Error, "resource not found")
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
					WithErrorDetails(fmt.Sprintf("resource %s not found", resourceID))
			}

			// 获取catalog
			catalog, err := s.cs.GetByID(ctx, resource.CatalogID, true)
			if err != nil {
				span.SetStatus(codes.Error, "get catalog failed")
				return nil, err.(*rest.HTTPError)
			}
			if catalog == nil {
				span.SetStatus(codes.Error, "catalog not found")
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_CatalogNotFound).
					WithErrorDetails(fmt.Sprintf("catalog %s not found", resource.CatalogID))
			}

			return s.executeOpenSearchQuery(ctx, req, []string{}, catalog)
		}
		// SQL流式查询
		return s.executeStreamQuery(ctx, req)
	}

	// 优先检查resource_type，因为OpenSearch查询的query是JSON对象，不包含resource_id占位符
	if req.ResourceType == "opensearch" {
		// OpenSearch查询，跳过resource_id提取
		// 从query中获取resource_id
		queryMap, ok := req.Query.(map[string]any)
		if !ok {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
				WithErrorDetails("query must be a JSON object for opensearch queries")
		}

		var resourceID string
		if rid, ok := queryMap["resource_id"].(string); ok && rid != "" {
			resourceID = rid
		} else {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
				WithErrorDetails("resource_id is required for opensearch queries")
		}

		// 获取资源信息
		resource, err := s.rs.GetByID(ctx, resourceID)
		if err != nil {
			span.SetStatus(codes.Error, "get resource failed")
			return nil, err.(*rest.HTTPError)
		}
		if resource == nil {
			span.SetStatus(codes.Error, "resource not found")
			return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
				WithErrorDetails(fmt.Sprintf("resource %s not found", resourceID))
		}

		// 获取catalog
		catalog, err := s.cs.GetByID(ctx, resource.CatalogID, true)
		if err != nil {
			span.SetStatus(codes.Error, "get catalog failed")
			return nil, err.(*rest.HTTPError)
		}
		if catalog == nil {
			span.SetStatus(codes.Error, "catalog not found")
			return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_CatalogNotFound).
				WithErrorDetails(fmt.Sprintf("catalog %s not found", resource.CatalogID))
		}

		return s.executeOpenSearchQuery(ctx, req, []string{}, catalog)
	}

	// 3. 从SQL中提取所有{{.resource_id}}占位符
	resourceIds, err := s.extractResourceIds(req.Query)
	if err != nil {
		span.SetStatus(codes.Error, "extract resource ids failed")
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("extract resource ids failed: %v", err))
	}

	// 4. 判断查询类型
	if len(resourceIds) == 0 {
		// 没有resource_id，直接执行原生SQL（需要指定resource_type）
		if req.ResourceType == "" {
			span.SetStatus(codes.Error, "resource_type is required")
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
				WithErrorDetails("resource_type is required when no resource_id in query")
		}
		return s.executeNativeSQL(ctx, req)
	}

	// 5. 判断所有resource_id是否来自同一个数据源
	// 获取catalog（从第一个resource_id获取）
	if len(resourceIds) > 0 {
		// 如果有resource_id，使用第一个resource_id获取catalog
		resource, err := s.rs.GetByID(ctx, resourceIds[0])
		if err != nil {
			span.SetStatus(codes.Error, "get resource failed")
			return nil, err.(*rest.HTTPError)
		}
		if resource == nil {
			span.SetStatus(codes.Error, "resource not found")
			return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
				WithErrorDetails(fmt.Sprintf("resource %s not found", resourceIds[0]))
		}
		// 获取catalog
		catalog, err := s.cs.GetByID(ctx, resource.CatalogID, true)
		if err != nil {
			span.SetStatus(codes.Error, "get catalog failed")
			return nil, err.(*rest.HTTPError)
		}
		if catalog == nil {
			span.SetStatus(codes.Error, "catalog not found")
			return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_CatalogNotFound).
				WithErrorDetails(fmt.Sprintf("catalog %s not found", resource.CatalogID))
		}

		// 根据catalog的ConnectorType来决定调用哪个方法
		if catalog.ConnectorType == "opensearch" {
			return s.executeOpenSearchQuery(ctx, req, resourceIds, catalog)
		}

		// 如果指定了resource_type为mysql/mariadb/postgresql，则不进行SQL转换，直接执行
		if req.ResourceType == "mysql" || req.ResourceType == "mariadb" || req.ResourceType == "postgresql" {
			// 将resource_id替换为catalog.schema.table格式
			replacedSQL, err := s.replaceResourceIdWithSchemaTable(ctx, req.Query, resourceIds, catalog)
			if err != nil {
				span.SetStatus(codes.Error, "replace resource id failed")
				return nil, err
			}

			// 直接执行SQL，不进行转换
			result, err := s.executeSQL(ctx, catalog, replacedSQL)
			if err != nil {
				return nil, err
			}

			// standard模式下，限制最大返回数据量为10000
			if req.QueryType == "" || req.QueryType == "standard" {
				if len(result.Entries) > 10000 {
					result.Entries = result.Entries[:10000]
					result.Stats.HasMore = true
				}
				// 更新TotalCount为实际返回的数据条数
				result.TotalCount = int64(len(result.Entries))
			}

			return result, nil
		}

		// 对于非OpenSearch查询，继续执行下面的SQL处理逻辑
	}

	// 6. 判断所有resource_id是否来自同一个数据源
	dataSource, err := s.checkSameDataSource(ctx, resourceIds)
	if err != nil {
		span.SetStatus(codes.Error, "check data source failed")
		return nil, err
	}

	// 7. 将resource_id替换为catalog.schema.table格式
	replacedSQL, err := s.replaceResourceIdWithSchemaTable(ctx, req.Query, resourceIds, dataSource)
	if err != nil {
		span.SetStatus(codes.Error, "replace resource id failed")
		return nil, err
	}

	// 8. 根据catalog的ConnectorType决定目标SQL方言
	var targetDialect string
	switch dataSource.ConnectorType {
	case "mariadb", "mysql":
		targetDialect = "mysql"
	case "postgresql":
		targetDialect = "postgres"
	default:
		span.SetStatus(codes.Error, "unsupported connector type")
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("unsupported connector type: %s", dataSource.ConnectorType))
	}

	// 9. 使用sqlglot将Trino SQL转换为目标SQL方言
	sqlParseResult, err := sqlglot.TranspileSQL(replacedSQL, "trino", targetDialect)
	if err != nil {
		span.SetStatus(codes.Error, "transpile SQL failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(fmt.Sprintf("transpile SQL failed: %v", err))
	}

	// 10. 为standard模式查询添加LIMIT 10000限制
	// 使用正则表达式检查SQL是否已经包含LIMIT子句（不区分大小写）
	finalSQL := sqlParseResult.SQL
	limitRegex := regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)\s*(?:,\s*\d+)?\s*$`)
	if !limitRegex.MatchString(finalSQL) {
		// 如果没有包含LIMIT子句，添加LIMIT 10000
		finalSQL = fmt.Sprintf("%s LIMIT 10000", finalSQL)
	} else {
		// 如果已经包含LIMIT子句，检查是否超过10000
		matches := limitRegex.FindStringSubmatch(finalSQL)
		if len(matches) > 1 {
			var limit int
			_, err := fmt.Sscanf(matches[1], "%d", &limit)
			if err == nil && limit > 10000 {
				// 如果LIMIT超过10000，替换为10000
				finalSQL = limitRegex.ReplaceAllString(finalSQL, "LIMIT 10000")
			}
		}
	}

	// 11. 执行转换后的SQL
	result, err := s.executeSQL(ctx, dataSource, finalSQL)
	if err != nil {
		return nil, err
	}
	// standard查询模式下，如果返回数据条数等于10000，则has_more设置为true
	if len(result.Entries) >= 10000 {
		result.Stats.HasMore = true
	}
	return result, nil
}

// validateRequest 校验请求
func (s *sqlQueryService) validateRequest(ctx context.Context, req *interfaces.SQLQueryRequest) error {
	// 校验查询类型
	if req.QueryType != "" && req.QueryType != "standard" && req.QueryType != "stream" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("query_type must be either 'standard' or 'stream'")
	}

	// 当不存在query_id时，query参数必填
	if req.QueryID == "" && req.Query == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("query is required when query_id is not provided")
	}

	// 如果提供了query参数，需要进行类型校验
	if req.Query != nil {
		// 如果是OpenSearch查询，query应该是map类型
		if req.ResourceType == "opensearch" {
			if _, ok := req.Query.(map[string]any); !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
					WithErrorDetails("query must be a JSON object for opensearch queries")
			}
			// 如果是OpenSearch流式查询，校验query中是否包含sort参数
			if req.QueryType == "stream" {
				if queryMap, ok := req.Query.(map[string]any); ok {
					if _, ok := queryMap["sort"]; !ok {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
							WithErrorDetails("sort is required for opensearch stream query")
					}
				}
			}
		} else {
			// 其他类型的查询，query应该是字符串类型
			if queryStr, ok := req.Query.(string); !ok || queryStr == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
					WithErrorDetails("query cannot be empty")
			}
		}
	}

	// 流式查询时，query_id和query不能同时出现
	if req.QueryType == "stream" && req.QueryID != "" && req.Query != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("query_id and query cannot be provided at the same time for stream query")
	}

	// 流式查询时，如果提供了query_id，则不需要校验query
	if req.QueryType == "stream" && req.QueryID != "" {
		return nil
	}

	// 流式查询时，stream_size必填（OpenSearch流式查询除外，使用size参数）
	if req.QueryType == "stream" && req.ResourceType != "opensearch" && req.StreamSize <= 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("stream_size is required for stream query and must be greater than 0")
	}

	return nil
}

// extractResourceIds 从SQL中提取所有{{.resource_id}}占位符
func (s *sqlQueryService) extractResourceIds(query any) ([]string, error) {
	// 如果query是字符串类型，使用正则表达式提取resource_id
	if queryStr, ok := query.(string); ok {
		re := regexp.MustCompile(`\{\{\.?(\w+)\}\}`)
		matches := re.FindAllStringSubmatch(queryStr, -1)

		resourceIds := make([]string, 0, len(matches))
		seen := make(map[string]bool)

		for _, match := range matches {
			if len(match) > 1 {
				resourceId := match[1]
				if !seen[resourceId] {
					seen[resourceId] = true
					resourceIds = append(resourceIds, resourceId)
				}
			}
		}

		return resourceIds, nil
	}

	// 如果query是map类型（OpenSearch DSL），返回空数组
	// OpenSearch查询通过resource_id参数指定索引
	return []string{}, nil
}

// checkSameDataSource 检查所有resource_id是否来自同一个数据源
func (s *sqlQueryService) checkSameDataSource(ctx context.Context, resourceIds []string) (*interfaces.Catalog, error) {
	if len(resourceIds) == 0 {
		return nil, fmt.Errorf("no resource ids provided")
	}

	// 获取所有资源
	resources, err := s.rs.GetByIDs(ctx, resourceIds)
	if err != nil {
		return nil, err.(*rest.HTTPError)
	}
	if len(resources) != len(resourceIds) {
		resourceMap := make(map[string]bool)
		for _, r := range resources {
			resourceMap[r.ID] = true
		}
		for _, id := range resourceIds {
			if !resourceMap[id] {
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
					WithErrorDetails(fmt.Sprintf("resource %s not found", id))
			}
		}
	}

	// 检查是否来自同一个catalog
	catalogIDs := make(map[string]bool)
	for _, r := range resources {
		catalogIDs[r.CatalogID] = true
	}
	if len(catalogIDs) > 1 {
		return nil, rest.NewHTTPError(ctx, http.StatusNotImplemented, verrors.VegaBackend_Query_MultiCatalogNotSupported).
			WithErrorDetails("暂不支持多数据源 JOIN，计划使用 Trino/DuckDB 实现。")
	}

	// 获取catalog
	var catalogID string
	for id := range catalogIDs {
		catalogID = id
		break
	}

	catalog, err := s.cs.GetByID(ctx, catalogID, true)
	if err != nil {
		return nil, err.(*rest.HTTPError)
	}
	if catalog == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", catalogID))
	}

	return catalog, nil
}

// replaceResourceIdWithSchemaTable 将resource_id替换为catalog.schema.table格式
func (s *sqlQueryService) replaceResourceIdWithSchemaTable(ctx context.Context, sql any, resourceIds []string, catalog *interfaces.Catalog) (string, error) {
	replacedSQL := sql.(string)

	for _, resourceId := range resourceIds {
		// 获取资源信息
		resource, err := s.rs.GetByID(ctx, resourceId)
		if err != nil {
			return "", err.(*rest.HTTPError)
		}
		if resource == nil {
			return "", rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
				WithErrorDetails(fmt.Sprintf("resource %s not found", resourceId))
		}

		// 构建catalog.schema.table格式
		// 使用catalogName + resource.SourceIdentifier
		// schemaTable := fmt.Sprintf(`%s.%s`, catalog.Name, resource.SourceIdentifier)

		// 替换{{.resource_id}}和{{resource_id}}为schema.table
		placeholder1 := fmt.Sprintf("{{.%s}}", resourceId)
		placeholder2 := fmt.Sprintf("{{%s}}", resourceId)
		replacedSQL = regexp.MustCompile(regexp.QuoteMeta(placeholder1)).ReplaceAllString(replacedSQL, resource.SourceIdentifier)
		replacedSQL = regexp.MustCompile(regexp.QuoteMeta(placeholder2)).ReplaceAllString(replacedSQL, resource.SourceIdentifier)
	}

	return replacedSQL, nil
}

// executeSQL 执行SQL查询
func (s *sqlQueryService) executeSQL(ctx context.Context, catalog *interfaces.Catalog, sql string) (*interfaces.SQLQueryResponse, error) {
	// 创建connector
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		logger.Errorf("create connector failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}

	// 根据connector类型执行SQL
	var result *interfaces.SQLQueryResponse
	switch catalog.ConnectorType {
	case "mariadb", "mysql":
		mariadbConnector := connector.(*mariadb.MariaDBConnector)
		result, err = mariadbConnector.ExecuteRawSQL(ctx, sql)
	case "postgresql":
		postgresqlConnector := connector.(*postgresql.PostgresqlConnector)
		result, err = postgresqlConnector.ExecuteRawSQL(ctx, sql)
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("unsupported connector type: %s", catalog.ConnectorType))
	}

	if err != nil {
		logger.Errorf("execute SQL failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}

	return result, nil
}

// executeNativeSQL 执行原生SQL（不包含resource_id）
func (s *sqlQueryService) executeNativeSQL(ctx context.Context, req *interfaces.SQLQueryRequest) (*interfaces.SQLQueryResponse, error) {
	// 获取SQL语句
	sql, ok := req.Query.(string)
	if !ok {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("query must be a string for native SQL execution")
	}

	// 根据resource_type确定connector类型
	var connectorType string
	switch req.ResourceType {
	case "mysql", "mariadb":
		connectorType = "mariadb"
	case "postgresql":
		connectorType = "postgresql"
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("unsupported resource_type: %s", req.ResourceType))
	}

	// 创建catalog对象用于执行SQL
	catalog := &interfaces.Catalog{
		ConnectorType: connectorType,
		ConnectorCfg:  map[string]any{},
	}

	// 直接执行SQL，不添加任何LIMIT限制
	result, err := s.executeSQL(ctx, catalog, sql)
	if err != nil {
		return nil, err
	}

	// standard模式下，限制最大返回数据量为10000
	if req.QueryType == "" || req.QueryType == "standard" {
		if len(result.Entries) > 10000 {
			result.Entries = result.Entries[:10000]
			result.Stats.HasMore = true
		}
		// 更新TotalCount为实际返回的数据条数
		result.TotalCount = int64(len(result.Entries))
	}

	return result, nil
}

// executeOpenSearchQuery 执行OpenSearch DSL查询
func (s *sqlQueryService) executeOpenSearchQuery(ctx context.Context, req *interfaces.SQLQueryRequest, resourceIds []string, catalog *interfaces.Catalog) (*interfaces.SQLQueryResponse, error) {
	// 验证query字段是否为有效的JSON对象
	queryMap, ok := req.Query.(map[string]any)
	if !ok {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("query must be a JSON object for opensearch queries")
	}

	// 确定resource_id
	var resourceID string
	// 优先从query参数中获取resource_id
	if rid, ok := queryMap["resource_id"].(string); ok && rid != "" {
		resourceID = rid
		// 从query中移除resource_id，避免传递给OpenSearch
		delete(queryMap, "resource_id")
	} else if len(resourceIds) > 0 {
		// 如果query中没有resource_id，从resourceIds中获取
		resourceID = resourceIds[0]
	} else {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("resource_id is required for opensearch queries")
	}

	// 创建connector
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		logger.Errorf("create connector failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}

	// 获取资源信息
	resource, err := s.rs.GetByID(ctx, resourceID)
	if err != nil {
		return nil, err.(*rest.HTTPError)
	}
	if resource == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
			WithErrorDetails(fmt.Sprintf("resource %s not found", resourceID))
	}

	// 使用resource.SourceIdentifier作为索引名
	indexName := resource.SourceIdentifier

	// 执行OpenSearch查询
	opensearchConnector := connector.(connectors.IndexConnector)
	if !ok {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails("connector is not an IndexConnector")
	}

	// 判断是否为流式查询
	isStreamQuery := req.QueryType == "stream"

	// 如果是流式查询，确保query中包含sort参数
	if isStreamQuery {
		if _, ok := queryMap["sort"]; !ok {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
				WithErrorDetails("sort is required for opensearch stream query")
		}
	} else {
		// 如果是standard模式查询，限制最多返回10000条数据
		const maxStandardQuerySize = 10000
		if size, ok := queryMap["size"]; ok {
			// 如果已经设置了size，检查是否超过10000
			if sizeFloat, ok := size.(float64); ok && sizeFloat > maxStandardQuerySize {
				queryMap["size"] = maxStandardQuerySize
			}
		} else {
			// 如果没有设置size，设置为10000
			queryMap["size"] = maxStandardQuerySize
		}
	}

	result, err := opensearchConnector.ExecuteRawQuery(ctx, indexName, queryMap)
	if err != nil {
		logger.Errorf("execute OpenSearch query failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}

	// 如果是流式查询，设置has_more
	if isStreamQuery {
		// 判断是否还有更多数据
		size := 10 // 默认值
		if s, ok := queryMap["size"].(float64); ok {
			size = int(s)
		}
		result.Stats.HasMore = len(result.Entries) >= size
		// search_after值已经由OpenSearch连接器在ExecuteRawQuery中设置
	}

	return result, nil
}

// executeStreamQuery 执行流式查询
func (s *sqlQueryService) executeStreamQuery(ctx context.Context, req *interfaces.SQLQueryRequest) (*interfaces.SQLQueryResponse, error) {
	// 1. 如果提供了query_id，则使用已有会话
	if req.QueryID != "" {
		return s.executeStreamQueryWithSession(ctx, req)
	}

	// 2. 如果没有提供query_id，则创建新会话
	return s.executeStreamQueryNewSession(ctx, req)
}

// executeStreamQueryNewSession 创建新会话并执行流式查询
func (s *sqlQueryService) executeStreamQueryNewSession(ctx context.Context, req *interfaces.SQLQueryRequest) (*interfaces.SQLQueryResponse, error) {
	// 1. 从SQL中提取resource_id
	resourceIds, err := s.extractResourceIds(req.Query)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("extract resource ids failed: %v", err))
	}

	if len(resourceIds) == 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("resource_id is required for stream query")
	}

	// 2. 获取资源信息
	resource, err := s.rs.GetByID(ctx, resourceIds[0])
	if err != nil {
		return nil, err.(*rest.HTTPError)
	}
	if resource == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
			WithErrorDetails(fmt.Sprintf("resource %s not found", resourceIds[0]))
	}

	// 3. 获取catalog
	catalog, err := s.cs.GetByID(ctx, resource.CatalogID, true)
	if err != nil {
		return nil, err.(*rest.HTTPError)
	}
	if catalog == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", resource.CatalogID))
	}

	// 4. 检查是否支持流式查询
	if catalog.ConnectorType != "mariadb" && catalog.ConnectorType != "mysql" && catalog.ConnectorType != "postgresql" {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("stream query is not supported for connector type: %s", catalog.ConnectorType))
	}

	// 5. 获取原始SQL
	var originalSQL string
	if queryStr, ok := req.Query.(string); ok {
		originalSQL = queryStr
	}

	// 6. 创建流式查询会话
	streamManager := GetStreamQueryManager()
	session, err := streamManager.CreateSession(catalog.ConnectorType, catalog.Name, catalog.ID, catalog, req.StreamSize, originalSQL, resourceIds)
	if err != nil {
		logger.Errorf("create stream query session failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}

	// 记录query_id和query的对应关系
	logger.Infof("Created stream query session - query_id: %s, query: %s, resource_ids: %v", session.QueryID, originalSQL, resourceIds)

	// 7. 执行查询
	return s.executeSQLWithSession(ctx, req, resourceIds, session)
}

// executeStreamQueryWithSession 使用已有会话执行流式查询
func (s *sqlQueryService) executeStreamQueryWithSession(ctx context.Context, req *interfaces.SQLQueryRequest) (*interfaces.SQLQueryResponse, error) {
	// 1. 获取会话
	streamManager := GetStreamQueryManager()
	session, ok := streamManager.GetSession(req.QueryID)
	if !ok {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("stream query session not found: %s", req.QueryID))
	}

	// 2. 如果提供了query参数，从SQL中提取resource_id
	var resourceIds []string
	if req.Query != nil {
		var err error
		resourceIds, err = s.extractResourceIds(req.Query)
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("extract resource ids failed: %v", err))
		}
	}

	// 记录使用已有会话时的query_id和query的对应关系
	logger.Infof("Using existing stream query session - query_id: %s, query: %s, resource_ids: %v, offset: %d",
		session.QueryID, session.OriginalSQL, session.ResourceIds, session.Offset)

	// 3. 执行查询
	return s.executeSQLWithSession(ctx, req, resourceIds, session)
}

// executeSQLWithSession 使用会话执行SQL查询
func (s *sqlQueryService) executeSQLWithSession(ctx context.Context, req *interfaces.SQLQueryRequest, resourceIds []string, session *StreamQuerySession) (*interfaces.SQLQueryResponse, error) {
	// 1. 获取catalog和resourceIds
	var catalog *interfaces.Catalog
	var err error
	var effectiveResourceIds []string
	var effectiveQuery any

	// 如果请求中提供了resourceIds，则使用请求中的
	if len(resourceIds) > 0 {
		effectiveResourceIds = resourceIds
		effectiveQuery = req.Query
		// 从resourceIds获取catalog
		catalog, err = s.checkSameDataSource(ctx, effectiveResourceIds)
		if err != nil {
			return nil, err
		}
	} else {
		// 否则使用会话中的catalog和resourceIds
		catalog = session.Catalog
		effectiveResourceIds = session.ResourceIds
		// 使用会话中保存的原始SQL
		effectiveQuery = session.OriginalSQL
	}

	// 2. 将resource_id替换为catalog.schema.table格式
	replacedSQL, err := s.replaceResourceIdWithSchemaTable(ctx, effectiveQuery, effectiveResourceIds, catalog)
	if err != nil {
		return nil, err
	}

	// 3. 根据catalog的ConnectorType决定目标SQL方言
	var targetDialect string
	switch catalog.ConnectorType {
	case "mariadb", "mysql":
		targetDialect = "mysql"
	case "postgresql":
		targetDialect = "postgres"
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("unsupported connector type: %s", catalog.ConnectorType))
	}

	// 4. 使用sqlglot将Trino SQL转换为目标SQL方言
	sqlParseResult, err := sqlglot.TranspileSQL(replacedSQL, "trino", targetDialect)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(fmt.Sprintf("transpile SQL failed: %v", err))
	}

	// 5. 获取当前偏移量
	streamManager := GetStreamQueryManager()
	currentSession, ok := streamManager.GetSession(session.QueryID)
	if !ok {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("stream query session not found: %s", session.QueryID))
	}

	// 6. 添加或替换LIMIT和OFFSET子句
	// 使用正则表达式检查SQL是否已经包含LIMIT子句（不区分大小写）
	var finalSQL string
	// 匹配 LIMIT 子句（可能包含 OFFSET）
	limitRegex := regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)\s*(?:OFFSET\s+(\d+))?\s*$`)
	if limitRegex.MatchString(sqlParseResult.SQL) {
		// 如果已经包含LIMIT子句，保留用户指定的LIMIT值，只添加或替换OFFSET
		matches := limitRegex.FindStringSubmatch(sqlParseResult.SQL)
		if len(matches) >= 2 {
			userLimit := matches[1]
			// 使用用户指定的LIMIT值，并添加或替换OFFSET
			finalSQL = limitRegex.ReplaceAllString(sqlParseResult.SQL, fmt.Sprintf("LIMIT %s OFFSET %d", userLimit, currentSession.Offset))
		} else {
			// 如果正则匹配失败，使用StreamSize
			finalSQL = limitRegex.ReplaceAllString(sqlParseResult.SQL, fmt.Sprintf("LIMIT %d OFFSET %d", currentSession.StreamSize, currentSession.Offset))
		}
	} else {
		// 如果没有包含LIMIT子句，使用StreamSize添加它
		finalSQL = fmt.Sprintf("%s LIMIT %d OFFSET %d", sqlParseResult.SQL, currentSession.StreamSize, currentSession.Offset)
	}

	// 记录执行查询的详细信息
	logger.Infof("Executing stream query - query_id: %s, offset: %d, stream_size: %d, sql: %s",
		currentSession.QueryID, currentSession.Offset, currentSession.StreamSize, finalSQL)

	// 7. 使用会话中的catalog执行查询
	result, err := s.executeSQL(ctx, currentSession.Catalog, finalSQL)
	if err != nil {
		return nil, err
	}

	// 8. 设置流式查询响应
	result.Stats.QueryID = currentSession.QueryID
	// 如果返回的数据量小于stream_size，说明这是最后一页
	result.Stats.HasMore = len(result.Entries) >= currentSession.StreamSize
	// 设置已获取到的数据总数
	result.Stats.Offset = currentSession.Offset

	// 9. 只有在还有更多数据时才更新偏移量，为下一次查询做准备
	if result.Stats.HasMore {
		currentSession.Offset += currentSession.StreamSize
	} else {
		// 如果没有更多数据，删除会话，防止重复请求
		streamManager.RemoveSession(currentSession.QueryID)
	}

	return result, nil
}
