// Package dataset provides Dataset management business logic.
package resource_data

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
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
	})
	return rdService
}

// Query 列出 resource 中的文档
func (rds *resourceDataService) Query(ctx context.Context, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List resource documents")
	defer span.End()

	logger.Debugf("Query, resourceID: %s, params: %v", resource.ID, params)

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}
	actualFilterCond, err := filter_condition.NewFilterCondition(ctx, params.FilterCondCfg, fieldMap)
	if err != nil {
		span.SetStatus(codes.Error, "Create filter condition failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	params.ActualFilterCond = actualFilterCond

	switch resource.Category {
	case interfaces.ResourceCategoryDataset:
		// 调用 dataset access 列出文档
		documents, total, err := rds.ds.ListDocuments(ctx, resource.ID, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "List dataset documents failed")
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
				span.SetStatus(codes.Error, "Query table data from local index failed")
				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
					WithErrorDetails(err.Error())
			}
			return documents, total, nil
		}
		// 准备 sort参数
		params = rds.prepareSortParams(resource, params)
		data, total, err := rds.QueryData(ctx, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Query table data failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}
		return data, total, nil

	case interfaces.ResourceCategoryIndex:
		data, total, err := rds.QueryData(ctx, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Query index data failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}
		return data, total, nil

	case interfaces.ResourceCategoryLogicView:
		// 准备 sort参数
		params = rds.prepareSortParams(resource, params)
		// 逻辑视图查询数据
		data, total, err := rds.lvs.Query(ctx, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Query logic view data failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}
		return data, total, nil

	case interfaces.ResourceCategoryFileset:
		data, total, err := rds.QueryData(ctx, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Query fileset data failed")
			return nil, 0, err
		}
		return data, total, nil

	default:
		span.SetStatus(codes.Error, "Unsupported resource category")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(resource.Category)
	}
}

func (rds *resourceDataService) QueryData(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query data")
	defer span.End()

	logger.Debugf("QueryData, resourceID: %s, catalogID: %s, params: %v",
		resource.ID, resource.CatalogID, params)

	catalog, err := rds.cs.GetByID(ctx, resource.CatalogID, true)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to get catalog: %v", err))
	}
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", resource.CatalogID))
	}

	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		span.SetStatus(codes.Error, "Create connector failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to create connector: %v", err))
	}

	if err := connector.Connect(ctx); err != nil {
		span.SetStatus(codes.Error, "Connect to data source failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to connect to data source: %v", err))
	}
	defer connector.Close(ctx)

	switch resource.Category {
	case interfaces.ResourceCategoryTable:
		tableConnector, ok := connector.(connectors.TableConnector)
		if !ok {
			span.SetStatus(codes.Error, "Connector does not support table operations")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
				WithErrorDetails(fmt.Sprintf("connector %s does not support table operations", catalog.ConnectorType))
		}

		result, err := tableConnector.ExecuteQuery(ctx, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Execute query failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to execute query: %v", err))
		}
		return result.Rows, result.Total, nil

	case interfaces.ResourceCategoryIndex:
		indexConnector, ok := connector.(connectors.IndexConnector)
		if !ok {
			span.SetStatus(codes.Error, "Connector does not support index operations")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
				WithErrorDetails(fmt.Sprintf("connector %s does not support index operations", catalog.ConnectorType))
		}

		result, err := indexConnector.ExecuteQuery(ctx, resource.SourceIdentifier, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Execute query failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to execute query: %v", err))
		}
		return result.Rows, result.Total, nil

	case interfaces.ResourceCategoryFileset:
		fc, ok := connector.(connectors.FilesetConnector)
		if !ok {
			span.SetStatus(codes.Error, "Connector does not support fileset operations")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
				WithErrorDetails(fmt.Sprintf("connector %s does not support fileset operations", catalog.ConnectorType))
		}

		// 使用 ExecuteQuery 获取文件列表
		result, err := fc.ExecuteQuery(ctx, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Fileset query failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}
		return result.Rows, result.Total, nil

	default:
		span.SetStatus(codes.Error, "Connector does not support table operations")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(connector.GetCategory())
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
