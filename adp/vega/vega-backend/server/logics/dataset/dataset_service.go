// Package dataset provides Dataset management business logic.
package dataset

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	asynqAccess "vega-backend/drivenadapters/asynq"
	taskAccess "vega-backend/drivenadapters/build_task"
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
	client     *asynq.Client
	c          connectors.IndexConnector
	ra         interfaces.ResourceAccess
	cs         interfaces.CatalogService
	ta         interfaces.BuildTaskAccess
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
			client:     asynqAccess.NewAsynqAccess(appSetting).CreateClient(context.Background()),
			c:          connector.(connectors.IndexConnector),
			ra:         resourceAccess.NewResourceAccess(appSetting),
			cs:         catalog.NewCatalogService(appSetting),
			ta:         taskAccess.NewBuildTaskAccess(appSetting),
		}
	})
	return dsService
}

// Create a new Dataset.
func (ds *datasetService) Create(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create dataset")
	defer span.End()

	// 调用 dataset access 创建 dataset 索引，索引名称为 <res.source_identifier>-<catalog_id>
	err := ds.c.Create(ctx, res.ID, res.SchemaDefinition)
	if err != nil {
		logger.Errorf("Create dataset index failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create dataset index failed: %v", err))
		span.SetStatus(codes.Error, "Create dataset index failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Update a Dataset.
func (ds *datasetService) Update(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update dataset")
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
func (ds *datasetService) Delete(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete dataset")
	defer span.End()

	// Check dataset exist first
	exist, err := ds.c.CheckExist(ctx, res.ID)
	if err != nil {
		span.SetStatus(codes.Error, "Check dataset exist failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	if exist {
		// Delete from storage
		if err := ds.c.Delete(ctx, res.ID); err != nil {
			span.SetStatus(codes.Error, "Delete dataset failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// ListDocuments 列出 dataset 中的文档
func (ds *datasetService) ListDocuments(ctx context.Context, res *interfaces.Resource, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List dataset documents")
	defer span.End()

	// 调用 dataset access 列出文档
	queryResult, err := ds.c.ExecuteQuery(ctx, res.ID, res, params)
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Create dataset documents")
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Get dataset document")
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

// UpdateDocument 更新 dataset 文档
func (ds *datasetService) UpdateDocument(ctx context.Context, id string, docID string, document map[string]any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update dataset document")
	defer span.End()

	// 调用 dataset access 更新文档
	if err := ds.c.UpdateDocument(ctx, id, docID, document); err != nil {
		span.SetStatus(codes.Error, "Update dataset document failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteDocument 删除 dataset 文档
func (ds *datasetService) DeleteDocument(ctx context.Context, id string, docID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete dataset document")
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

// UpdateDocuments 批量更新 dataset 文档
func (ds *datasetService) UpdateDocuments(ctx context.Context, id string, updateRequests []map[string]any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update dataset documents")
	defer span.End()

	// 调用 dataset access 批量更新文档
	if err := ds.c.UpdateDocuments(ctx, id, updateRequests); err != nil {
		span.SetStatus(codes.Error, "Update dataset documents failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteDocuments 批量删除 dataset 文档
func (ds *datasetService) DeleteDocuments(ctx context.Context, id string, docIDs string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete dataset documents")
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
func (ds *datasetService) DeleteDocumentsByQuery(ctx context.Context, res *interfaces.Resource, params *interfaces.ResourceDataQueryParams) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete dataset documents by query")
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
	if err := ds.c.DeleteDocumentsByQuery(ctx, res.ID, params, res.SchemaDefinition); err != nil {
		span.SetStatus(codes.Error, "Delete dataset documents failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Build builds a resource by batch reading data from source and writing to dataset.
func (ds *datasetService) Build(ctx context.Context, id string) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Build resource")
	defer span.End()

	// Get resource
	resource, err := ds.ra.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get resource failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
	}

	// Check if resource category is dataset
	if resource.Category != interfaces.ResourceCategoryDataset {
		span.SetStatus(codes.Error, "Resource category is not dataset")
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails("Resource category must be dataset")
	}

	// Get catalog
	catalog, err := ds.cs.GetByID(ctx, resource.CatalogID, false)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
	}

	// Check if catalog connector type is mysql
	if catalog.ConnectorType != "mysql" {
		span.SetStatus(codes.Error, "Catalog connector type is not mysql")
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ConnectorType).
			WithErrorDetails("Catalog connector type must be mysql")
	}

	// Check if resource has build task
	hasUncompletedTasks, err := ds.ta.CheckResourceHasUncompletedTasks(ctx, resource.ID)
	if err != nil {
		span.SetStatus(codes.Error, "Check resource has uncompleted tasks failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if hasUncompletedTasks {
		span.SetStatus(codes.Error, "Resource has uncompleted build task")
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_Exist).
			WithErrorDetails("Resource has uncompleted build task")
	}

	// Create build task
	taskReq := &interfaces.BuildTaskRequest{
		Mode: interfaces.BuildTaskModeFull,
	}
	taskID, err := ds.CreateBuildTask(ctx, resource.ID, taskReq)
	if err != nil {
		span.SetStatus(codes.Error, "Create task failed")
		return "", err
	}

	if taskReq.Mode == interfaces.BuildTaskModeFull {
		// Full build, need delete existing dataset
		logger.Warnf("Full build, need delete existing dataset for resource: %s", resource.ID)
		delerr := ds.Delete(ctx, resource)
		if delerr != nil {
			logger.Errorf("Delete dataset failed: %v", delerr)
		}
	}

	span.SetStatus(codes.Ok, "")
	return taskID, nil
}

// CreateBuildTask creates a new BuildTask.
func (ds *datasetService) CreateBuildTask(ctx context.Context, id string, req *interfaces.BuildTaskRequest) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create build task")
	defer span.End()

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	now := time.Now().UnixMilli()
	buildTask := &interfaces.BuildTask{
		ID:              xid.New().String(),
		ResourceID:      id,
		Status:          interfaces.BuildTaskStatusPending,
		Mode:            req.Mode,
		TotalCount:      0,
		SyncedCount:     0,
		VectorizedCount: 0,
		Creator:         accountInfo,
		CreateTime:      now,
		Updater:         accountInfo,
		UpdateTime:      now,
	}

	err := ds.ta.Create(ctx, buildTask)
	if err != nil {
		logger.Errorf("Create build task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create build task failed: %v", err))
		span.SetStatus(codes.Error, "Create build task failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	payload, err := sonic.Marshal(&interfaces.BuildTaskMessage{
		TaskID: buildTask.ID,
	})
	if err != nil {
		logger.Errorf("Marshal build task message failed: %v", err)
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	asynqTask := asynq.NewTask(interfaces.BuildTaskType, payload)
	_, err = ds.client.Enqueue(asynqTask,
		asynq.Queue("default"),
		asynq.MaxRetry(10),
		asynq.Timeout(30*time.Minute),
		asynq.Deadline(time.Now().Add(24*time.Hour)),
	)
	if err != nil {
		logger.Errorf("Enqueue build task failed: %v", err)
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	return buildTask.ID, nil
}

// GetBuildTaskByID retrieves a BuildTask by ID.
func (ds *datasetService) GetBuildTaskByID(ctx context.Context, id string) (*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build task")
	defer span.End()

	buildTask, err := ds.ta.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get build task failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if buildTask == nil {
		span.SetStatus(codes.Error, "Build task not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Task_NotFound)
	}

	span.SetStatus(codes.Ok, "")
	return buildTask, nil
}

// GetBuildTasksByResourceID retrieves BuildTasks by resource ID.
func (ds *datasetService) GetBuildTasksByResourceID(ctx context.Context, resourceID string) ([]*interfaces.BuildTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get build tasks by resource ID")
	defer span.End()

	buildTasks, err := ds.ta.GetByResourceID(ctx, resourceID)
	if err != nil {
		span.SetStatus(codes.Error, "Get build tasks failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_BuildTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return buildTasks, nil
}
