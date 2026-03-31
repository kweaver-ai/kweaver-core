package logic_view

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/filter_condition"
	"vega-backend/logics/permission"
	"vega-backend/logics/resource"
)

var (
	lvServiceOnce sync.Once
	lvService     interfaces.LogicViewService
)

type logicViewService struct {
	appSetting *common.AppSetting
	cs         interfaces.CatalogService
	rs         interfaces.ResourceService
	ps         interfaces.PermissionService
}

// NewLogicViewService creates a new ResourceDataService.
func NewLogicViewService(appSetting *common.AppSetting) interfaces.LogicViewService {
	lvServiceOnce.Do(func() {
		lvService = &logicViewService{
			appSetting: appSetting,
			cs:         catalog.NewCatalogService(appSetting),
			rs:         resource.NewResourceService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return lvService
}

func (lvs *logicViewService) Query(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query logic view")
	defer span.End()

	logger.Debugf("Query logic view, resourceID: %s, params: %v",
		resource.ID, params)

	switch resource.LogicType {
	case interfaces.LogicType_Derived:
		return lvs.queryDerivedLogicView(ctx, resource, params)
	case interfaces.LogicType_Composite:
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_NotImplemented).
			WithErrorDetails(fmt.Sprintf("Composite view '%s' is not supported yet", resource.ID))
	default:
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(fmt.Sprintf("The logic type of the custom view '%s' is not supported", resource.ID))
	}
}

func (lvs *logicViewService) queryDerivedLogicView(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query derived logic view")
	defer span.End()

	var inputNode *interfaces.LogicDefinitionNode
	for _, node := range resource.LogicDefinition {
		if node.Type == interfaces.LogicDefinitionNodeType_Resource {
			inputNode = node
			break
		}
	}

	var nodeCfg interfaces.ResourceNodeCfg
	if err := mapstructure.Decode(inputNode.Config, &nodeCfg); err != nil {
		span.SetStatus(codes.Error, "Decode resource node config failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to decode resource node config: %v", err))
	}
	fromResourceFilterCond := nodeCfg.Filters

	fromResource, err := lvs.rs.GetByID(ctx, nodeCfg.ResourceID)
	if err != nil {
		span.SetStatus(codes.Error, "Get source resource failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to get source resource %s: %v", nodeCfg.ResourceID, err))
	}
	if fromResource == nil {
		span.SetStatus(codes.Error, "Source resource not found")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
			WithErrorDetails(fmt.Sprintf("source resource %s not found", nodeCfg.ResourceID))
	}

	catalog, err := lvs.cs.GetByID(ctx, fromResource.CatalogID, true)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to get catalog: %v", err))
	}
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", fromResource.CatalogID))
	}

	fieldMap := map[string]*interfaces.Property{}
	outputFields := make([]string, 0, len(resource.SchemaDefinition))
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
		outputFields = append(outputFields, prop.Name)
	}
	params.OutputFields = outputFields

	// 合并资源和查询的 FilterCondCfg, 需要判断下是否为nil
	var mergedFilterCond *interfaces.FilterCondCfg
	if fromResourceFilterCond != nil && params.FilterCondCfg != nil {
		mergedFilterCond = &interfaces.FilterCondCfg{
			Operation: filter_condition.OperationAnd,
			SubConds:  []*interfaces.FilterCondCfg{fromResourceFilterCond, params.FilterCondCfg},
		}
	} else if fromResourceFilterCond != nil {
		mergedFilterCond = fromResourceFilterCond
	} else if params.FilterCondCfg != nil {
		mergedFilterCond = params.FilterCondCfg
	}

	actualFilterCond, err := filter_condition.NewFilterCondition(ctx, mergedFilterCond, fieldMap)
	if err != nil {
		span.SetStatus(codes.Error, "Create filter condition failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	params.ActualFilterCond = actualFilterCond

	// 交给 querySingleSourceData Connector 处理 SQL push-down
	return querySingleSourceData(ctx, catalog, fromResource, params)
}

func querySingleSourceData(ctx context.Context, catalog *interfaces.Catalog, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query data")
	defer span.End()

	logger.Debugf("QueryData, resourceID: %s, catalogID: %s, params: %v",
		resource.ID, resource.CatalogID, params)

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

	default:
		span.SetStatus(codes.Error, "Connector does not support table operations")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(connector.GetCategory())
	}
}
