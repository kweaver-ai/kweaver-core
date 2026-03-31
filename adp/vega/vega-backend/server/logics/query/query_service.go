// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package query

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/filter_condition"
)

const (
	QueryLimitMax = 10000
)

type queryService struct {
	cs interfaces.CatalogService
	rs interfaces.ResourceService
	ss interfaces.QuerySessionStore
}

// NewQueryService 创建查询服务
func NewQueryService(cs interfaces.CatalogService, rs interfaces.ResourceService, ss interfaces.QuerySessionStore) *queryService {
	return &queryService{cs: cs, rs: rs, ss: ss}
}

// Execute 执行统一查询
func (qs *queryService) Execute(ctx context.Context, req *interfaces.QueryExecuteRequest) (*interfaces.QueryExecuteResponse, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "QueryExecute")
	defer span.End()

	// 1. 校验
	if err := qs.validateRequest(ctx, req); err != nil {
		span.SetStatus(codes.Error, "validate request failed")
		return nil, err
	}

	// 1.1 首页允许不传 query_id，后端生成并回传
	if req.QueryID == "" && req.Offset == 0 {
		req.QueryID = xid.New().String()
	}

	// 2. 获取 resources
	resourceIDs := make([]string, len(req.Tables))
	resourceIDToAlias := make(map[string]string)
	for i, t := range req.Tables {
		resourceIDs[i] = t.ResourceID
		if t.Alias != "" {
			resourceIDToAlias[t.ResourceID] = t.Alias
		}
	}

	resources, err := qs.rs.GetByIDs(ctx, resourceIDs)
	if err != nil {
		span.SetStatus(codes.Error, "get resources failed")
		return nil, err.(*rest.HTTPError)
	}
	if len(resources) != len(resourceIDs) {
		resourceMap := make(map[string]bool)
		for _, r := range resources {
			resourceMap[r.ID] = true
		}
		for _, id := range resourceIDs {
			if !resourceMap[id] {
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_ResourceNotFound).
					WithErrorDetails(fmt.Sprintf("resource %s not found", id))
			}
		}
	}

	// 3. 解析 catalog_id 集合
	catalogIDs := make(map[string]bool)
	for _, r := range resources {
		catalogIDs[r.CatalogID] = true
	}
	if len(catalogIDs) > 1 {
		return nil, rest.NewHTTPError(ctx, http.StatusNotImplemented, verrors.VegaBackend_Query_MultiCatalogNotSupported).
			WithErrorDetails("暂不支持多数据源 JOIN，计划使用 Trino/DuckDB 实现。")
	}

	var catalogID string
	for id := range catalogIDs {
		catalogID = id
		break
	}

	// 4. 校验 joins 中 alias 均在 tables 中
	aliasSet := make(map[string]bool)
	for _, t := range req.Tables {
		alias := t.Alias
		if alias == "" {
			alias = t.ResourceID
		}
		aliasSet[alias] = true
	}
	for _, j := range req.Joins {
		if !aliasSet[j.LeftTableAlias] || !aliasSet[j.RightTableAlias] {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter_JoinTableNotInTables).
				WithErrorDetails(fmt.Sprintf("join alias %s or %s not in tables", j.LeftTableAlias, j.RightTableAlias))
		}
	}

	// 5. 获取 catalog
	cat, err := qs.cs.GetByID(ctx, catalogID, true)
	if err != nil {
		span.SetStatus(codes.Error, "get catalog failed")
		return nil, err.(*rest.HTTPError)
	}
	if cat == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Query_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", catalogID))
	}

	// 6. 构建 filter condition
	var filterCond interfaces.FilterCondition
	fieldMap := make(map[string]*interfaces.Property)
	for _, r := range resources {
		alias := resourceIDToAlias[r.ID]
		if alias == "" {
			alias = r.ID
		}
		for _, prop := range r.SchemaDefinition {
			key := alias + "." + prop.Name
			fieldMap[key] = &interfaces.Property{
				Name:         prop.Name,
				Type:         prop.Type,
				DisplayName:  prop.DisplayName,
				OriginalName: key,
				Description:  prop.Description,
			}
			fieldMap[prop.Name] = prop
		}
	}

	var filterCondCfg *interfaces.FilterCondCfg
	if req.FilterCondition != nil {
		if err := mapstructure.Decode(req.FilterCondition, &filterCondCfg); err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
				WithErrorDetails(err.Error())
		}
		if filterCondCfg != nil && (filterCondCfg.Name != "" || filterCondCfg.Operation != "" || len(filterCondCfg.SubConds) > 0) {
			filterCond, err = filter_condition.NewFilterCondition(ctx, filterCondCfg, fieldMap)
			if err != nil {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
					WithErrorDetails(err.Error())
			}
		}
	}

	// 7. 获取游标
	prevOffset := req.Offset - req.Limit
	if prevOffset < 0 {
		prevOffset = -1
	}
	cursorEncoded := ""
	if prevOffset >= 0 {
		cursorEncoded, _ = qs.ss.GetCursor(ctx, req.QueryID, prevOffset)
	}
	_ = qs.ss.Touch(ctx, req.QueryID)

	// 8. 创建 connector 并执行
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, cat.ConnectorType, cat.ConnectorCfg)
	if err != nil {
		span.SetStatus(codes.Error, "create connector failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}
	if err := connector.Connect(ctx); err != nil {
		span.SetStatus(codes.Error, "connect failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}
	defer connector.Close(ctx)

	tableConnector, ok := connector.(connectors.TableConnector)
	if !ok {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails("connector does not support table join query")
	}

	joinParams := &interfaces.JoinQueryParams{
		Resources:         resources,
		ResourceIDToAlias: resourceIDToAlias,
		Joins:             ptrSlice(req.Joins),
		OutputFields:      req.OutputFields,
		FilterCondCfg:     filterCondCfg,
		ActualFilterCond:  filterCond,
		Sort:              req.Sort,
		Offset:            req.Offset,
		Limit:             req.Limit,
		NeedTotal:         req.NeedTotal,
		CursorEncoded:     cursorEncoded,
	}

	result, err := tableConnector.ExecuteJoinQuery(ctx, cat, joinParams)
	if err != nil {
		logger.Errorf("ExecuteJoinQuery failed: %v", err)
		span.SetStatus(codes.Error, "execute query failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Query_ExecuteFailed).
			WithErrorDetails(err.Error())
	}

	// 9. 写回游标
	if len(result.Rows) > 0 && len(req.Sort) > 0 {
		lastRow := result.Rows[len(result.Rows)-1]
		cursorVals := make([]any, 0, len(req.Sort))
		for _, sf := range req.Sort {
			if v, ok := lastRow[sf.Field]; ok {
				cursorVals = append(cursorVals, v)
			} else {
				// 尝试无别名
				for _, r := range resources {
					alias := resourceIDToAlias[r.ID]
					if alias == "" {
						alias = r.ID
					}
					key := alias + "." + sf.Field
					if v, ok := lastRow[key]; ok {
						cursorVals = append(cursorVals, v)
						break
					}
				}
			}
		}
		if len(cursorVals) == len(req.Sort) {
			encoded, _ := json.Marshal(cursorVals)
			cursorB64 := base64.StdEncoding.EncodeToString(encoded)
			_ = qs.ss.SetCursor(ctx, req.QueryID, req.Offset, cursorB64)
		}
	}

	// 10. 构建响应
	resp := &interfaces.QueryExecuteResponse{
		QueryID:    req.QueryID,
		Entries:    result.Rows,
		NextOffset: req.Offset + req.Limit,
		HasMore:    int64(len(result.Rows)) == int64(req.Limit),
	}
	if req.NeedTotal {
		resp.TotalCount = &result.Total
	}

	return resp, nil
}

func (qs *queryService) validateRequest(ctx context.Context, req *interfaces.QueryExecuteRequest) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(req.Tables) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter).
			WithErrorDetails("tables cannot be empty")
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > QueryLimitMax {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter_LimitExceeded).
			WithErrorDetails(fmt.Sprintf("limit cannot exceed %d", QueryLimitMax))
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	// 首页（offset=0）允许不传 query_id；非首页必须带 query_id 才能命中 session 游标
	if req.QueryID == "" && req.Offset > 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Query_InvalidParameter_QueryIDRequired).
			WithErrorDetails("query_id is required for non-first page requests")
	}
	return nil
}

func ptrSlice[T any](s []T) []*T {
	if s == nil {
		return nil
	}
	out := make([]*T, len(s))
	for i := range s {
		out[i] = &s[i]
	}
	return out
}
