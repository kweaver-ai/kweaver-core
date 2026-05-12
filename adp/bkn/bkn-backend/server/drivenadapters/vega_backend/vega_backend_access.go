// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package vega_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	vbAccessOnce sync.Once
	vbAccess     interfaces.VegaBackendAccess
)

type vegaBackendAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
	baseUrl    string
}

// NewVegaBackendAccess creates a new vega-backend access instance
func NewVegaBackendAccess(appSetting *common.AppSetting) interfaces.VegaBackendAccess {
	vbAccessOnce.Do(func() {
		vbAccess = &vegaBackendAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
			baseUrl:    appSetting.VegaBackendUrl,
		}
	})

	return vbAccess
}

func (vba *vegaBackendAccess) buildHeaders(ctx context.Context) map[string]string {
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	// accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo := ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		headers[interfaces.HTTP_HEADER_ACCOUNT_ID] = accountInfo.ID
		headers[interfaces.HTTP_HEADER_ACCOUNT_TYPE] = accountInfo.Type
	} else {
		headers[interfaces.HTTP_HEADER_ACCOUNT_ID] = interfaces.ADMIN_ACCOUNT_ID
		headers[interfaces.HTTP_HEADER_ACCOUNT_TYPE] = interfaces.ADMIN_ACCOUNT_TYPE
	}

	return headers
}

func (vba *vegaBackendAccess) GetCatalogByID(ctx context.Context, id string) (*interfaces.Catalog, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Get catalog by ID")
	defer span.End()

	span.SetAttributes(attr.Key("catalog_id").String(id))

	httpUrl := fmt.Sprintf("%s/catalogs/%s", vba.baseUrl, url.PathEscape(id))
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := vba.buildHeaders(ctx)
	respCode, respData, err := vba.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	logger.Debugf("GetCatalogByID [%s] finished, response code is [%d], result is [%s], error is [%v]",
		httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("GetCatalogByID http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get catalog by ID failed")
		return nil, fmt.Errorf("GetCatalogByID http request failed: %s", err)
	}

	if respCode == http.StatusNotFound {
		return nil, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("GetCatalogByID failed: %s", respData)
		return nil, fmt.Errorf("GetCatalogByID failed: %s", respData)
	}

	var catalog interfaces.Catalog
	if err := json.Unmarshal([]byte(respData), &catalog); err != nil {
		logger.Errorf("Failed to unmarshal GetCatalogByID response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal GetCatalogByID response: %v", err)
	}

	return &catalog, nil
}

func (vba *vegaBackendAccess) CreateCatalog(ctx context.Context, req *interfaces.CatalogRequest) (*interfaces.Catalog, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Create catalog")
	defer span.End()

	span.SetAttributes(attr.Key("catalog_name").String(req.Name))

	httpUrl := fmt.Sprintf("%s/catalogs", vba.baseUrl)
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := vba.buildHeaders(ctx)
	respCode, respData, err := vba.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, req)
	logger.Debugf("CreateCatalog [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("CreateCatalog http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http create catalog failed")
		return nil, fmt.Errorf("CreateCatalog http request failed: %s", err)
	}

	if respCode != http.StatusCreated && respCode != http.StatusOK {
		logger.Errorf("CreateCatalog failed: %s", respData)
		return nil, fmt.Errorf("CreateCatalog failed: %s", respData)
	}

	var catalog interfaces.Catalog
	if err := json.Unmarshal([]byte(respData), &catalog); err != nil {
		logger.Errorf("Failed to unmarshal CreateCatalog response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal CreateCatalog response: %v", err)
	}

	return &catalog, nil
}

func (vba *vegaBackendAccess) GetResourceByID(ctx context.Context, id string) (*interfaces.VegaResource, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Get resource by ID")
	defer span.End()

	span.SetAttributes(attr.Key("resource_id").String(id))

	httpUrl := fmt.Sprintf("%s/resources/%s", vba.baseUrl, url.PathEscape(id))
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := vba.buildHeaders(ctx)
	respCode, respData, err := vba.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	logger.Debugf("GetResourceByID [%s] finished, response code is [%d], result is [%s], error is [%v]",
		httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("GetResourceByID http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get resource by ID failed")
		return nil, fmt.Errorf("GetResourceByID http request failed: %s", err)
	}

	if respCode == http.StatusNotFound {
		return nil, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("GetResourceByID failed: %s", respData)
		return nil, fmt.Errorf("GetResourceByID failed: %s", respData)
	}

	var resourceData struct {
		Entries []*interfaces.VegaResource `json:"entries"`
	}
	if err := json.Unmarshal([]byte(respData), &resourceData); err != nil {
		logger.Errorf("Failed to unmarshal GetResourceByID response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal GetResourceByID response: %v", err)
	}

	if len(resourceData.Entries) == 0 {
		return nil, nil
	}

	return resourceData.Entries[0], nil
}

func (vba *vegaBackendAccess) CreateResource(ctx context.Context, req *interfaces.VegaResource) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Create resource")
	defer span.End()

	span.SetAttributes(attr.Key("resource_name").String(req.Name))

	httpUrl := fmt.Sprintf("%s/resources", vba.baseUrl)
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := vba.buildHeaders(ctx)
	respCode, respData, err := vba.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, req)
	logger.Debugf("CreateResource [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("CreateResource http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http create resource failed")
		return fmt.Errorf("CreateResource http request failed: %s", err)
	}

	if respCode != http.StatusCreated && respCode != http.StatusOK {
		logger.Errorf("CreateResource failed: %s", respData)
		return fmt.Errorf("CreateResource failed: %s", respData)
	}

	var resource interfaces.VegaResource
	if err := json.Unmarshal([]byte(respData), &resource); err != nil {
		logger.Errorf("Failed to unmarshal CreateResource response: %v", err)
		return fmt.Errorf("failed to unmarshal CreateResource response: %v", err)
	}

	return nil
}

func (vba *vegaBackendAccess) DeleteResource(ctx context.Context, id string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Delete resource")
	defer span.End()

	span.SetAttributes(attr.Key("resource_id").String(id))

	httpUrl := fmt.Sprintf("%s/resources/%s", vba.baseUrl, url.PathEscape(id))
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodDelete,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := vba.buildHeaders(ctx)
	respCode, respData, err := vba.httpClient.DeleteNoUnmarshal(ctx, httpUrl, headers)
	logger.Debugf("DeleteResource [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("DeleteResource http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http delete resource failed")
		return fmt.Errorf("DeleteResource http request failed: %s", err)
	}

	if respCode != http.StatusNoContent && respCode != http.StatusOK {
		logger.Errorf("DeleteResource failed: %s", respData)
		return fmt.Errorf("DeleteResource failed: %s", respData)
	}

	return nil
}

func (vba *vegaBackendAccess) DeleteDatasetDocumentByID(ctx context.Context, datasetID string, docID string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Delete dataset document by ID")
	defer span.End()

	span.SetAttributes(attr.Key("dataset_id").String(datasetID))
	span.SetAttributes(attr.Key("doc_id").String(docID))

	httpUrl := fmt.Sprintf("%s/resources/%s/data/%s", vba.baseUrl, url.PathEscape(datasetID), url.PathEscape(docID))
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:    httpUrl,
		HttpMethod: http.MethodDelete,
	})

	headers := vba.buildHeaders(ctx)
	respCode, respData, err := vba.httpClient.DeleteNoUnmarshal(ctx, httpUrl, headers)
	logger.Debugf("DeleteDatasetDocumentByID [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("DeleteDatasetDocumentByID http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http delete dataset document by ID failed")
		return fmt.Errorf("DeleteDatasetDocumentByID http request failed: %s", err)
	}

	if respCode != http.StatusNoContent && respCode != http.StatusOK {
		logger.Errorf("DeleteDatasetDocumentByID failed: %s", respData)
		return fmt.Errorf("DeleteDatasetDocumentByID failed: %s", respData)
	}

	return nil
}

func (vba *vegaBackendAccess) DeleteDatasetDocumentsByQuery(ctx context.Context, datasetID string, filterCondition map[string]any) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Delete dataset documents by query")
	defer span.End()

	span.SetAttributes(attr.Key("dataset_id").String(datasetID))

	httpUrl := fmt.Sprintf("%s/resources/%s/data", vba.baseUrl, url.PathEscape(datasetID))
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	reqBody := map[string]any{
		"filter_condition": filterCondition,
	}

	headers := vba.buildHeaders(ctx)
	headers[oteltrace.HTTP_HEADER_METHOD_OVERRIDE] = http.MethodDelete
	reqBodyJson, _ := sonic.Marshal(reqBody)
	respCode, respData, err := vba.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, reqBody)
	logger.Debugf("DeleteDatasetDocumentsByQuery [%s] finished, request is [%s], response code is [%d], result is [%s], error is [%v]",
		httpUrl, string(reqBodyJson), respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("DeleteDatasetDocumentsByQuery http request failed: %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http delete dataset documents by query failed")
		return fmt.Errorf("DeleteDatasetDocumentsByQuery http request failed: %s", err)
	}

	if respCode != http.StatusNoContent && respCode != http.StatusOK {
		logger.Errorf("DeleteDatasetDocumentsByQuery failed: %s", respData)
		return fmt.Errorf("DeleteDatasetDocumentsByQuery failed: %s", respData)
	}

	return nil
}

// 从 vega resource中获取数据
func (vba *vegaBackendAccess) QueryResourceData(ctx context.Context, resourceID string, params *interfaces.ResourceDataQueryParams) (*interfaces.DatasetQueryResponse, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Query dataset data")
	defer span.End()

	span.SetAttributes(attr.Key("dataset_id").String(resourceID))

	httpUrl := fmt.Sprintf("%s/resources/%s/data", vba.baseUrl, url.PathEscape(resourceID))
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := vba.buildHeaders(ctx)
	headers[oteltrace.HTTP_HEADER_METHOD_OVERRIDE] = http.MethodGet
	paramsJson, _ := sonic.Marshal(params)
	respCode, respData, err := vba.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, params)
	logger.Debugf("QueryDatasetData [%s] finished, request is [%s], response code is [%d],  error is [%v]",
		httpUrl, string(paramsJson), respCode, err)

	if err != nil {
		errDetails := fmt.Sprintf("QueryDatasetData http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http query dataset data failed")
		return nil, fmt.Errorf("QueryDatasetData http request failed: %s", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("QueryDatasetData failed: %s", respData)
		return nil, fmt.Errorf("QueryDatasetData failed: %s", respData)
	}

	var response interfaces.DatasetQueryResponse
	if err := json.Unmarshal([]byte(respData), &response); err != nil {
		logger.Errorf("Failed to unmarshal QueryDatasetData response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal QueryDatasetData response: %v", err)
	}

	return &response, nil
}

func (vba *vegaBackendAccess) WriteDatasetDocuments(ctx context.Context, datasetID string, documents []map[string]any) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "driven layer: Write dataset documents")
	defer span.End()

	span.SetAttributes(attr.Key("dataset_id").String(datasetID))
	span.SetAttributes(attr.Key("documents_count").Int(len(documents)))

	httpUrl := fmt.Sprintf("%s/resources/%s/data", vba.baseUrl, url.PathEscape(datasetID))
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := vba.buildHeaders(ctx)
	headers[oteltrace.HTTP_HEADER_METHOD_OVERRIDE] = http.MethodPost
	reqBodyJson, _ := sonic.Marshal(documents)
	respCode, respData, err := vba.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, documents)
	logger.Debugf("WriteDatasetDocuments [%s] finished,	 request is [%s], response code is [%d], result is [%s], error is [%v]",
		httpUrl, string(reqBodyJson), respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("WriteDatasetDocuments http request failed: %s", err.Error())
		otellog.LogError(ctx, errDetails, nil)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http write dataset documents failed")
		return fmt.Errorf("WriteDatasetDocuments http request failed: %s", err)
	}

	if respCode != http.StatusCreated && respCode != http.StatusOK {
		logger.Errorf("WriteDatasetDocuments failed: %s", respData)
		return fmt.Errorf("WriteDatasetDocuments failed: %s", respData)
	}

	return nil
}
