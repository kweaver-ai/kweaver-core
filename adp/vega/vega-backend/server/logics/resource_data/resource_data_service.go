// Package dataset provides Dataset management business logic.
package resource_data

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/dataset"
	"vega-backend/logics/filter_condition"
	"vega-backend/logics/rate"
	"vega-backend/logics/resource"
	"vega-backend/logics/resource_data/logic_view"
)

var (
	rdServiceOnce sync.Once
	rdService     interfaces.ResourceDataService
)

type resourceDataService struct {
	appSetting *common.AppSetting
	ds         interfaces.DatasetService
	cs         interfaces.CatalogService
	rs         interfaces.ResourceService
	lvs        interfaces.LogicViewService
	cl         rate.ConcurrencyLimiter
}

// NewResourceDataService creates a new ResourceDataService.
func NewResourceDataService(appSetting *common.AppSetting) interfaces.ResourceDataService {
	rdServiceOnce.Do(func() {
		rdService = &resourceDataService{
			appSetting: appSetting,
			ds:         dataset.NewDatasetService(appSetting),
			cs:         catalog.NewCatalogService(appSetting),
			rs:         resource.NewResourceService(appSetting),
			lvs:        logic_view.NewLogicViewService(appSetting),
		}

		// Initialize concurrency limiter if enabled
		if appSetting.RateLimitingSetting.Concurrency.Enabled && appSetting.RateLimitingSetting.Concurrency.Global.MaxConcurrentQueries > 0 {
			cfg := rate.ConcurrencyConfig{
				Enabled: appSetting.RateLimitingSetting.Concurrency.Enabled,
				Global: rate.GlobalConcurrencyConfig{
					MaxConcurrentQueries: appSetting.RateLimitingSetting.Concurrency.Global.MaxConcurrentQueries,
				},
			}

			rdService.(*resourceDataService).cl = rate.NewConcurrencyLimiter(cfg)
		}
	})
	return rdService
}

// Query 列出 resource 中的文档
func (rds *resourceDataService) Query(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "List resource documents")
	defer span.End()

	logger.Debugf("Query, resourceID: %s, params: %v", resource.ID, params)

	var err error
	var catalog *interfaces.Catalog
	maxConcurrentQueries := int64(0)
	if resource.Category != interfaces.ResourceCategoryLogicView {
		catalog, err = rds.cs.GetByID(ctx, resource.CatalogID, true)
		if err != nil {
			otellog.LogError(ctx, "Get catalog failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to get catalog: %v", err))
		}
		if catalog == nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
				WithErrorDetails(fmt.Sprintf("catalog %s not found", resource.CatalogID))
			otellog.LogError(ctx, "Catalog not found", httpErr)
			return nil, 0, httpErr
		}
		if concurrent, existsInCatalog := catalog.ConnectorCfg["concurrent"]; existsInCatalog {
			maxConcurrentQueries = int64(concurrent.(float64))
		}
	}

	// 并发控制
	var release func()
	if rds.cl != nil {
		// 获取并发许可
		var acquireErr error
		release, acquireErr = rds.cl.Acquire(rate.AcquireParams{
			CatalogID:            resource.CatalogID,
			MaxConcurrentQueries: maxConcurrentQueries,
		})
		if acquireErr != nil {
			logger.Warnf("Concurrency limit exceeded: catalog=%s, error=%v",
				resource.CatalogID, acquireErr)

			// 返回限流错误
			if rateErr, ok := acquireErr.(*rate.RateLimitError); ok {
				httpErr := rest.NewHTTPError(ctx, rateErr.HTTPStatus, verrors.VegaBackend_Query_ConcurrencyLimitExceeded).
					WithErrorDetails(rateErr.Message)
				otellog.LogError(ctx, "Concurrency limit exceeded", httpErr)
				return nil, 0, httpErr
			}
			httpErr := rest.NewHTTPError(ctx, http.StatusTooManyRequests, verrors.VegaBackend_Query_ConcurrencyLimitExceeded).
				WithErrorDetails("Query concurrency limit exceeded, please retry later")
			otellog.LogError(ctx, "Concurrency limit exceeded", httpErr)
			return nil, 0, httpErr
		}
		defer release() // 查询完成后释放许可
	}

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}
	actualFilterCond, err := filter_condition.NewFilterCondition(ctx, params.FilterCondCfg, fieldMap)
	if err != nil {
		otellog.LogError(ctx, "Create filter condition failed", err)
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	params.ActualFilterCond = actualFilterCond

	switch resource.Category {
	case interfaces.ResourceCategoryDataset:
		// 调用 dataset access 列出文档
		documents, total, err := rds.ds.ListDocuments(ctx, resource.ID, resource, params)
		if err != nil {
			otellog.LogError(ctx, "List dataset documents failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}
		return documents, total, nil

	case interfaces.ResourceCategoryTable:
		// 检查是否有索引名称，如果有则直接查询索引
		if resource.LocalIndexName != "" {
			// 调用 dataset access 列出文档
			documents, total, err := rds.ds.ListDocuments(ctx, resource.LocalIndexName, resource, params)
			if err != nil {
				otellog.LogError(ctx, "Query table data from local index failed", err)
				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
					WithErrorDetails(err.Error())
			}

			span.SetStatus(codes.Ok, "")
			return documents, total, nil
		}

		// 准备 sort参数
		params = rds.prepareSortParams(resource, params)
		// 准备 output
		params = rds.prepareOutputFieldsParams(resource, params)

		data, total, err := rds.QueryData(ctx, catalog, resource, params)
		if err != nil {
			otellog.LogError(ctx, "Query table data failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}

		span.SetStatus(codes.Ok, "")
		return data, total, nil

	case interfaces.ResourceCategoryIndex:
		data, total, err := rds.QueryData(ctx, catalog, resource, params)
		if err != nil {
			otellog.LogError(ctx, "Query index data failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}

		span.SetStatus(codes.Ok, "")
		return data, total, nil

	case interfaces.ResourceCategoryLogicView:
		// 准备 sort参数
		params = rds.prepareSortParams(resource, params)
		// 准备 output
		params = rds.prepareOutputFieldsParams(resource, params)

		// 逻辑视图查询数据
		data, total, err := rds.lvs.Query(ctx, resource, params)
		if err != nil {
			otellog.LogError(ctx, "Query logic view data failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}

		span.SetStatus(codes.Ok, "")
		return data, total, nil

	case interfaces.ResourceCategoryFileset:
		// 准备 sort参数
		params = rds.prepareSortParams(resource, params)
		// 准备 output
		params = rds.prepareOutputFieldsParams(resource, params)

		data, total, err := rds.QueryData(ctx, catalog, resource, params)
		if err != nil {
			otellog.LogError(ctx, "Query fileset data failed", err)
			return nil, 0, err
		}

		span.SetStatus(codes.Ok, "")
		return data, total, nil

	default:
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(resource.Category)
		otellog.LogError(ctx, "Unsupported resource category", httpErr)
		return nil, 0, httpErr
	}
}

func (rds *resourceDataService) QueryData(ctx context.Context, catalog *interfaces.Catalog, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Query data")
	defer span.End()

	logger.Debugf("QueryData, resourceID: %s, catalogID: %s, params: %v",
		resource.ID, resource.CatalogID, params)

	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		otellog.LogError(ctx, "Create connector failed", err)
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to create connector: %v", err))
	}

	if err := connector.Connect(ctx); err != nil {
		otellog.LogError(ctx, "Connect to data source failed", err)
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to connect to data source: %v", err))
	}
	defer func() { _ = connector.Close(ctx) }()

	switch resource.Category {
	case interfaces.ResourceCategoryTable:
		tableConnector, ok := connector.(connectors.TableConnector)
		if !ok {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
				WithErrorDetails(fmt.Sprintf("connector %s does not support table operations", catalog.ConnectorType))
			otellog.LogError(ctx, "Connector does not support table operations", httpErr)
			return nil, 0, httpErr
		}

		result, err := tableConnector.ExecuteQuery(ctx, resource, params)
		if err != nil {
			otellog.LogError(ctx, "Execute query failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to execute query: %v", err))
		}

		span.SetStatus(codes.Ok, "")
		return result.Rows, result.Total, nil

	case interfaces.ResourceCategoryIndex:
		indexConnector, ok := connector.(connectors.IndexConnector)
		if !ok {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
				WithErrorDetails(fmt.Sprintf("connector %s does not support index operations", catalog.ConnectorType))
			otellog.LogError(ctx, "Connector does not support index operations", httpErr)
			return nil, 0, httpErr
		}

		result, err := indexConnector.ExecuteQuery(ctx, resource.SourceIdentifier, resource, params)
		if err != nil {
			otellog.LogError(ctx, "Execute query failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to execute query: %v", err))
		}

		span.SetStatus(codes.Ok, "")
		return result.Rows, result.Total, nil

	case interfaces.ResourceCategoryFileset:
		fc, ok := connector.(connectors.FilesetConnector)
		if !ok {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
				WithErrorDetails(fmt.Sprintf("connector %s does not support fileset operations", catalog.ConnectorType))
			otellog.LogError(ctx, "Connector does not support fileset operations", httpErr)
			return nil, 0, httpErr
		}

		// 使用 ExecuteQuery 获取文件列表
		result, err := fc.ExecuteQuery(ctx, resource, params)
		if err != nil {
			otellog.LogError(ctx, "Fileset query failed", err)
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}

		span.SetStatus(codes.Ok, "")
		return result.Rows, result.Total, nil

	default:
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(connector.GetCategory())
		otellog.LogError(ctx, "Connector does not support table operations", httpErr)
		return nil, 0, httpErr
	}

}

// prepareSortParams prepares sort parameters to only include fields defined in resource SchemaDefinition
func (rds *resourceDataService) prepareSortParams(resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) *interfaces.ResourceDataQueryParams {
	if resource == nil || params == nil {
		return params
	}

	// Create a field map for quick lookup
	fieldMap := make(map[string]bool)
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = true
	}

	// Add aggregation alias to field map if aggregation is present
	if params.Aggregation != nil && params.Aggregation.Alias != "" {
		fieldMap[params.Aggregation.Alias] = true
	}
	// Add __value to field map for aggregation queries
	if params.Aggregation != nil {
		fieldMap["__value"] = true
	}
	// Add GROUP BY fields to field map for aggregation queries
	if params.GroupBy != nil {
		for _, groupByItem := range params.GroupBy {
			fieldMap[groupByItem.Property] = true
		}
	}

	filteredParams := params

	// Filter Sort fields to only include fields defined in SchemaDefinition
	if params.Sort != nil {
		filteredSort := []*interfaces.SortField{}
		for _, sortField := range params.Sort {
			if fieldMap[sortField.Field] {
				filteredSort = append(filteredSort, sortField)
			}
		}
		filteredParams.Sort = filteredSort
	}

	return filteredParams
}

// prepareOutputFieldsParams filters output fields to only include fields defined in resource SchemaDefinition.
func (rds *resourceDataService) prepareOutputFieldsParams(resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) *interfaces.ResourceDataQueryParams {
	if resource == nil || params == nil || len(params.OutputFields) == 0 {
		return params
	}

	fieldMap := make(map[string]bool, len(resource.SchemaDefinition))
	for _, prop := range resource.SchemaDefinition {
		if prop == nil || prop.Name == "" {
			continue
		}
		fieldMap[prop.Name] = true
	}

	filteredOutputFields := make([]string, 0, len(params.OutputFields))
	for _, field := range params.OutputFields {
		if fieldMap[field] || (field == "_score" && resource.Category == interfaces.ResourceCategoryIndex) {
			filteredOutputFields = append(filteredOutputFields, field)
		}
	}
	params.OutputFields = filteredOutputFields

	return params
}
