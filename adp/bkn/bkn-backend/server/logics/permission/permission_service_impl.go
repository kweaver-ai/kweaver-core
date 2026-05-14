// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	mqclient "github.com/kweaver-ai/proton-mq-sdk-go"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

type PermissionServiceImpl struct {
	appSetting *common.AppSetting
	mqClient   mqclient.ProtonMQClient
	pa         interfaces.PermissionAccess
}

func NewPermissionServiceImpl(appSetting *common.AppSetting) interfaces.PermissionService {
	mqSetting := appSetting.MQSetting
	client, err := mqclient.NewProtonMQClient(mqSetting.MQHost, mqSetting.MQPort,
		mqSetting.MQHost, mqSetting.MQPort, mqSetting.MQType,
		mqclient.UserInfo(mqSetting.Auth.Username, mqSetting.Auth.Password),
		mqclient.AuthMechanism(mqSetting.Auth.Mechanism),
	)
	if err != nil {
		logger.Fatal("failed to create a proton mq client:", err)
	}
	return &PermissionServiceImpl{
		appSetting: appSetting,
		mqClient:   client,
		pa:         logics.PA,
	}
}

func (ps *PermissionServiceImpl) CheckPermission(ctx context.Context, resource interfaces.PermissionResource, ops []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "CheckPermission")
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	if accountInfo.ID == "" || accountInfo.Type == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: missing account ID or type")
		otellog.LogError(ctx, "CheckPermission missing account ID or type", httpErr)
		return httpErr
	}

	// todo: 暂时先去掉权限校验
	ok, err := ps.pa.CheckPermission(ctx, interfaces.PermissionCheck{
		Accessor: interfaces.PermissionAccessor{
			ID:   accountInfo.ID,
			Type: accountInfo.Type,
		},
		Resource:   resource,
		Operations: ops,
	})
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_InternalError_CheckPermissionFailed).WithErrorDetails(err)
		otellog.LogError(ctx, "CheckPermission failed", httpErr)
		return httpErr
	}
	if !ok {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", ops))
		otellog.LogError(ctx, "CheckPermission denied", httpErr)
		return httpErr
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// 添加资源权限（新建决策）
func (ps *PermissionServiceImpl) CreateResources(ctx context.Context, resources []interfaces.PermissionResource, ops []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "CreatePermissionResources")
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	if accountInfo.ID == "" || accountInfo.Type == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: missing account ID or type")
		otellog.LogError(ctx, "CreateResources missing account ID or type", httpErr)
		return httpErr
	}

	// todo: 创建资源权限暂时先去掉
	allowOps := []interfaces.PermissionOperation{}
	for _, op := range ops {
		allowOps = append(allowOps, interfaces.PermissionOperation{
			Operation: op,
		})
	}

	policies := []interfaces.PermissionPolicy{}
	for _, resource := range resources {
		policies = append(policies, interfaces.PermissionPolicy{
			Accessor: interfaces.PermissionAccessor{
				Type: accountInfo.Type,
				ID:   accountInfo.ID,
			},
			Resource: resource,
			Operations: interfaces.PermissionPolicyOps{
				Allow: allowOps,
				Deny:  []interfaces.PermissionOperation{},
			},
		})
	}

	err := ps.pa.CreateResources(ctx, policies)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_InternalError_CreateResourcesFailed).WithErrorDetails(err.Error())
		otellog.LogError(ctx, "CreateResources failed", httpErr)
		return httpErr
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除策略
func (ps *PermissionServiceImpl) DeleteResources(ctx context.Context, resourceType string, ids []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "DeletePermissionResources")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}
	// todo：删除权限资源暂时先去掉
	// 清除资源策略
	resources := []interfaces.PermissionResource{}
	for _, id := range ids {
		resources = append(resources, interfaces.PermissionResource{
			Type: resourceType,
			ID:   id,
		})
	}

	err := ps.pa.DeleteResources(ctx, resources)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_InternalError_DeleteResourcesFailed).WithErrorDetails(err)
		otellog.LogError(ctx, "DeleteResources failed", httpErr)
		return httpErr
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// 过滤资源列表
func (ps *PermissionServiceImpl) FilterResources(ctx context.Context, resourceType string, ids []string,
	ops []string, allowOperation bool, fullOps []string) (map[string]interfaces.PermissionResourceOps, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "FilterPermissionResources")
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	if accountInfo.ID == "" || accountInfo.Type == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: missing account ID or type")
		otellog.LogError(ctx, "FilterResources missing account ID or type", httpErr)
		return nil, httpErr
	}

	resources := []interfaces.PermissionResource{}
	for _, id := range ids {
		resources = append(resources, interfaces.PermissionResource{
			ID:   id,
			Type: resourceType,
		})
	}

	// todo: 权限过滤先去掉，进来多少个id就返回多少个id
	matchResouces, err := ps.pa.FilterResources(ctx, interfaces.PermissionResourcesFilter{
		Accessor: interfaces.PermissionAccessor{
			ID:   accountInfo.ID,
			Type: accountInfo.Type,
		},
		Resources:      resources,
		Operations:     ops,
		AllowOperation: allowOperation,
	})
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_InternalError_FilterResourcesFailed).WithErrorDetails(err)
		otellog.LogError(ctx, "FilterResources failed", httpErr)
		return nil, httpErr
	}

	// id转map
	idMap := map[string]interfaces.PermissionResourceOps{}
	for _, resourceOps := range matchResouces {
		idMap[resourceOps.ResourceID] = resourceOps
	}

	span.SetStatus(codes.Ok, "")
	return idMap, nil
}

// 更新资源名称
func (ps *PermissionServiceImpl) UpdateResource(ctx context.Context, resource interfaces.PermissionResource) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "UpdatePermissionResource")
	defer span.End()

	bytes, err := sonic.Marshal(resource)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_InternalError_MarshalDataFailed).WithErrorDetails(err)
		otellog.LogError(ctx, "UpdateResource marshal failed", httpErr)
		return httpErr
	}

	err = ps.mqClient.Pub(interfaces.AUTHORIZATION_RESOURCE_NAME_MODIFY, bytes)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_InternalError_UpdateResourceFailed).WithErrorDetails(err)
		otellog.LogError(ctx, "UpdateResource publish failed", httpErr)
		return httpErr
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
