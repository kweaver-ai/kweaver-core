// Package dataset provides Dataset management business logic.
package dataset

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	resourceAccess "vega-backend/drivenadapters/resource"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	opensearchConnector "vega-backend/logics/connectors/local/index/opensearch"
	"vega-backend/logics/filter_condition"
)

var (
	dsServiceOnce sync.Once
	dsService     interfaces.DatasetService
)

type datasetService struct {
	appSetting *common.AppSetting
	c          connectors.IndexConnector
	ra         interfaces.ResourceAccess
	cs         interfaces.CatalogService
}

// NewDatasetService creates a new DatasetService.
func NewDatasetService(appSetting *common.AppSetting) interfaces.DatasetService {
	dsServiceOnce.Do(func() {
		// Get OpenSearch config from depServices
		opensearchSetting, ok := appSetting.DepServices["opensearch"]
		if !ok {
			panic("opensearch service not found in depServices")
		}

		// Create connector config
		cfg := interfaces.ConnectorConfig{
			"host":          opensearchSetting["host"],
			"port":          opensearchSetting["port"],
			"username":      opensearchSetting["user"],
			"password":      opensearchSetting["password"],
			"index_pattern": opensearchSetting["index_pattern"],
		}

		// Create OpenSearch connector
		connector, err := opensearchConnector.NewOpenSearchConnector().New(cfg)
		if err != nil {
			panic(fmt.Sprintf("failed to create OpenSearch connector: %v", err))
		}

		dsService = &datasetService{
			appSetting: appSetting,
			c:          connector.(connectors.IndexConnector),
			ra:         resourceAccess.NewResourceAccess(appSetting),
			cs:         catalog.NewCatalogService(appSetting),
		}
	})
	return dsService
}

// Create a new Dataset.
func (ds *datasetService) Create(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Create dataset")
	defer span.End()

	// 调用 dataset access 创建 dataset 索引，索引名称为 <res.source_identifier>-<catalog_id>
	err := ds.c.Create(ctx, res.ID, res.SchemaDefinition)
	if err != nil {
		otellog.LogError(ctx, "Create dataset index failed", err)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Update a Dataset.
func (ds *datasetService) Update(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Update dataset")
	defer span.End()

	// 调用 dataset access 更新 dataset 索引，索引名称为 <res.source_identifier>-<id>
	if err := ds.c.Update(ctx, fmt.Sprintf("%s-%s", res.SourceIdentifier, res.ID), res.SchemaDefinition); err != nil {
		span.SetStatus(codes.Error, "Update dataset failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Delete a Dataset.
func (ds *datasetService) Delete(ctx context.Context, id string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete dataset")
	defer span.End()

	// Check dataset exist first
	exist, err := ds.c.CheckExist(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Check dataset exist failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	if exist {
		// Delete from storage
		if err := ds.c.Delete(ctx, id); err != nil {
			span.SetStatus(codes.Error, "Delete dataset failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExist checks if a dataset exists.
func (ds *datasetService) CheckExist(ctx context.Context, id string) (bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Check dataset exist")
	defer span.End()

	exist, err := ds.c.CheckExist(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Check dataset exist failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return exist, nil
}

// ListDocuments 列出 dataset 中的文档
func (ds *datasetService) ListDocuments(ctx context.Context, indexName string, res *interfaces.Resource, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "List dataset documents")
	defer span.End()

	// 调用 dataset access 列出文档
	queryResult, err := ds.c.ExecuteQuery(ctx, indexName, res, params)
	if err != nil {
		span.SetStatus(codes.Error, "List dataset documents failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return queryResult.Rows, queryResult.Total, nil
}

// CreateDocuments 批量创建 dataset 文档
func (ds *datasetService) CreateDocuments(ctx context.Context, id string, documents []map[string]any) ([]string, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Create dataset documents")
	defer span.End()

	// 调用 dataset access 批量创建文档
	docIDs, err := ds.c.CreateDocuments(ctx, id, documents)
	if err != nil {
		span.SetStatus(codes.Error, "Create dataset documents failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return docIDs, nil
}

// GetDocument 获取 dataset 文档
func (ds *datasetService) GetDocument(ctx context.Context, id string, docID string) (map[string]any, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Get dataset document")
	defer span.End()

	// 调用 dataset access 获取文档
	document, err := ds.c.GetDocument(ctx, id, docID)
	if err != nil {
		span.SetStatus(codes.Error, "Get dataset document failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return document, nil
}

// DeleteDocument 删除 dataset 文档
func (ds *datasetService) DeleteDocument(ctx context.Context, id string, docID string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete dataset document")
	defer span.End()

	// 调用 dataset access 删除文档
	if err := ds.c.DeleteDocument(ctx, id, docID); err != nil {
		span.SetStatus(codes.Error, "Delete dataset document failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpsertDocuments 批量更新 dataset 文档
func (ds *datasetService) UpsertDocuments(ctx context.Context, id string, updateRequests []map[string]any) ([]string, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Update dataset documents")
	defer span.End()

	// 调用 dataset access 批量更新文档
	docIDs, err := ds.c.UpsertDocuments(ctx, id, updateRequests)
	if err != nil {
		span.SetStatus(codes.Error, "Update dataset documents failed")
		return docIDs, err
	}

	span.SetStatus(codes.Ok, "")
	return docIDs, nil
}

// DeleteDocuments 批量删除 dataset 文档
func (ds *datasetService) DeleteDocuments(ctx context.Context, id string, docIDs string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete dataset documents")
	defer span.End()

	// 调用 dataset access 批量删除文档
	if err := ds.c.DeleteDocuments(ctx, id, docIDs); err != nil {
		span.SetStatus(codes.Error, "Delete dataset documents failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteDocumentsByQuery 批量删除 dataset 文档
func (ds *datasetService) DeleteDocumentsByQuery(ctx context.Context, indexName string, res *interfaces.Resource, params *interfaces.ResourceDataQueryParams) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete dataset documents by query")
	defer span.End()

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range res.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}

	// 创建实际的过滤条件
	actualFilterCond, err := filter_condition.NewFilterCondition(ctx, params.FilterCondCfg, fieldMap)
	if err != nil {
		span.SetStatus(codes.Error, "Create filter condition failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	params.ActualFilterCond = actualFilterCond

	// 调用 dataset access 批量删除文档
	if err := ds.c.DeleteDocumentsByQuery(ctx, indexName, params, res.SchemaDefinition); err != nil {
		span.SetStatus(codes.Error, "Delete dataset documents failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
