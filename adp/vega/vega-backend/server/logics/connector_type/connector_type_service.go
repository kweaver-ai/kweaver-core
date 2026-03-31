// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package connectortype provides ConnectorType management business logic.
package connector_type

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	connectorTypeAccess "vega-backend/drivenadapters/connector_type"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/permission"
)

var (
	ctServiceOnce sync.Once
	ctService     interfaces.ConnectorTypeService
)

type connectorTypeService struct {
	appSetting *common.AppSetting
	cta        interfaces.ConnectorTypeAccess
	cf         *factory.ConnectorFactory
	ps         interfaces.PermissionService
}

// NewConnectorTypeService creates a new ConnectorTypeService.
func NewConnectorTypeService(appSetting *common.AppSetting) interfaces.ConnectorTypeService {
	ctServiceOnce.Do(func() {
		ctService = &connectorTypeService{
			appSetting: appSetting,
			cta:        connectorTypeAccess.NewConnectorTypeAccess(appSetting),
			cf:         factory.GetFactory(),
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return ctService
}

// Register register a new ConnectorType.
func (cts *connectorTypeService) Register(ctx context.Context, req *interfaces.ConnectorTypeReq) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Register connector type")
	defer span.End()

	// 判断userid是否有创建业务知识网络的权限（策略决策）
	err := cts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_CONNECTOR_TYPE,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return err
	}

	ct := &interfaces.ConnectorType{
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Mode:        req.Mode,
		Category:    req.Category,
		Endpoint:    req.Endpoint,
		FieldConfig: req.FieldConfig,
		Enabled:     req.Enabled,
	}

	err = cts.cta.Create(ctx, ct)
	if err != nil {
		logger.Errorf("Register connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Register connector type failed: %v", err))
		span.SetStatus(codes.Error, "Register connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_RegisterFailed).
			WithErrorDetails(err.Error())
	}

	err = cts.cf.RegisterConnector(ctx, ct.Type, ct)
	if err != nil {
		logger.Errorf("Register connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Register connector type failed: %v", err))
		span.SetStatus(codes.Error, "Register connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_RegisterFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByType retrieves a ConnectorType by Type.
func (cts *connectorTypeService) GetByType(ctx context.Context, tp string) (*interfaces.ConnectorType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get connector type")
	defer span.End()

	ct, err := cts.cta.GetByType(ctx, tp)
	if err != nil {
		span.SetStatus(codes.Error, "Get connector type failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if ct == nil {
		span.SetStatus(codes.Error, "Connector type not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_ConnectorType_NotFound)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := cts.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CONNECTOR_TYPE, []string{ct.Type},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return nil, err
	}

	if resrc, exist := matchResoucesMap[ct.Type]; exist {
		ct.Operations = resrc.Operations // 用户当前有权限的操作
	} else {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", interfaces.OPERATION_TYPE_VIEW_DETAIL))
	}

	span.SetStatus(codes.Ok, "")
	return ct, nil
}

// List lists ConnectorTypes with filters.
func (cts *connectorTypeService) List(ctx context.Context, params interfaces.ConnectorTypesQueryParams) ([]*interfaces.ConnectorType, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List connector types")
	defer span.End()

	connectorTypesArr, _, err := cts.cta.List(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List connector types failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	// 处理资源id
	types := make([]string, 0)
	for _, m := range connectorTypesArr {
		types = append(types, m.Type)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := cts.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CONNECTOR_TYPE, types,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return []*interfaces.ConnectorType{}, 0, err
	}

	connectorTypes := make([]*interfaces.ConnectorType, 0)
	for _, c := range connectorTypesArr {
		// 只留下有权限的模型
		if resrc, exist := matchResoucesMap[c.Type]; exist {
			c.Operations = resrc.Operations // 用户当前有权限的操作
			connectorTypes = append(connectorTypes, c)
		}
	}
	total := int64(len(connectorTypes))

	// limit = -1,则返回所有
	if params.Limit != -1 {
		// 分页
		// 检查起始位置是否越界
		if params.Offset < 0 || params.Offset >= len(connectorTypes) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ConnectorType{}, total, nil
		}
		// 计算结束位置
		end := params.Offset + params.Limit
		if end > len(connectorTypes) {
			end = len(connectorTypes)
		}

		connectorTypes = connectorTypes[params.Offset:end]
	}

	span.SetStatus(codes.Ok, "")
	return connectorTypes, total, nil
}

// Update updates a ConnectorType.
func (cts *connectorTypeService) Update(ctx context.Context, req *interfaces.ConnectorTypeReq) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update connector type")
	defer span.End()

	ct := req.OriginConnectorType
	if ct == nil {
		span.SetStatus(codes.Error, "Connector type not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_ConnectorType_NotFound)
	}

	// 判断userid是否有创建业务知识网络的权限（策略决策）
	err := cts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_CONNECTOR_TYPE,
		ID:   ct.Type,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// Apply updates
	if req.Type != ct.Type {
		span.SetStatus(codes.Error, "can not change connector type")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Type)
	}
	if req.Mode != ct.Mode {
		span.SetStatus(codes.Error, "can not change connector mode")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Mode)
	}
	if req.Category != ct.Category {
		span.SetStatus(codes.Error, "can not change connector category")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Category)
	}

	ct.Type = req.Type
	ct.Name = req.Name
	ct.Tags = req.Tags
	ct.Description = req.Description
	ct.Mode = req.Mode
	ct.Category = req.Category
	ct.Endpoint = req.Endpoint
	ct.FieldConfig = req.FieldConfig
	ct.Enabled = req.Enabled

	if err := cts.cta.Update(ctx, ct); err != nil {
		span.SetStatus(codes.Error, "Update connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	if err := cts.cf.RegisterConnector(ctx, ct.Type, ct); err != nil {
		logger.Errorf("Register connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Register connector type failed: %v", err))
		span.SetStatus(codes.Error, "Register connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_RegisterFailed).
			WithErrorDetails(err.Error())
	}

	// 请求更新资源名称的接口，更新资源的名称
	if req.IfNameModify {
		err = cts.ps.UpdateResource(ctx, interfaces.PermissionResource{
			ID:   ct.Type,
			Type: interfaces.RESOURCE_TYPE_CONNECTOR_TYPE,
			Name: ct.Name,
		})
		if err != nil {
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Delete deletes a ConnectorType.
func (cts *connectorTypeService) DeleteByType(ctx context.Context, tp string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete connector type")
	defer span.End()

	// 判断userid是否有删除权限
	err := cts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_CONNECTOR_TYPE,
		ID:   tp,
	}, []string{interfaces.OPERATION_TYPE_DELETE})
	if err != nil {
		return err
	}

	if err := cts.cta.DeleteByType(ctx, tp); err != nil {
		span.SetStatus(codes.Error, "Delete connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	if err := cts.cf.DeleteConnector(ctx, tp); err != nil {
		logger.Errorf("Delete connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Delete connector type failed: %v", err))
		span.SetStatus(codes.Error, "Delete connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = cts.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_CONNECTOR_TYPE, []string{tp})
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// SetEnabled sets the enabled status of a ConnectorType.
func (cts *connectorTypeService) SetEnabled(ctx context.Context, tp string, enabled bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Set enabled connector type")
	defer span.End()

	// 判断userid是否有修改权限
	err := cts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_CONNECTOR_TYPE,
		ID:   tp,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if err := cts.cta.SetEnabled(ctx, tp, enabled); err != nil {
		span.SetStatus(codes.Error, "Set enabled connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExistByType checks if a ConnectorType exists by Type.
func (cts *connectorTypeService) CheckExistByType(ctx context.Context, tp string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check connector type exist by Type")
	defer span.End()

	ct, err := cts.cta.GetByType(ctx, tp)
	if err != nil {
		span.SetStatus(codes.Error, "GetByType failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ct != nil, nil
}

// CheckExistByName checks if a ConnectorType exists by Name.
func (cts *connectorTypeService) CheckExistByName(ctx context.Context, name string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check connector type exist by Name")
	defer span.End()

	ct, err := cts.cta.GetByName(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, "GetByName failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ct != nil, nil
}
