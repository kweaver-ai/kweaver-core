// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package catalog provides Catalog management business logic.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	kwcrypto "github.com/kweaver-ai/kweaver-go-lib/crypto"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	catalogAccess "vega-backend/drivenadapters/catalog"
	resourceAccess "vega-backend/drivenadapters/resource"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/permission"
	"vega-backend/logics/user_mgmt"
)

const (
	// EncryptedPrefix is the prefix for encrypted values.
	EncryptedPrefix = "ENC:"
)

var (
	cServiceOnce sync.Once
	cService     interfaces.CatalogService
)

type catalogService struct {
	appSetting *common.AppSetting
	cipher     kwcrypto.Cipher
	ca         interfaces.CatalogAccess
	ra         interfaces.ResourceAccess
	ps         interfaces.PermissionService
	ums        interfaces.UserMgmtService
}

// NewCatalogService creates a new CatalogService.
func NewCatalogService(appSetting *common.AppSetting) interfaces.CatalogService {
	cServiceOnce.Do(func() {
		var cipher kwcrypto.Cipher
		if appSetting.CryptoSetting.Enabled {
			var err error
			cipher, err = kwcrypto.NewRSACipher(appSetting.CryptoSetting.PrivateKey, appSetting.CryptoSetting.PublicKey)
			if err != nil {
				logger.Fatalf("Failed to create RSA cipher: %v", err)
			}
		}
		cService = &catalogService{
			appSetting: appSetting,
			cipher:     cipher,
			ca:         catalogAccess.NewCatalogAccess(appSetting),
			ra:         resourceAccess.NewResourceAccess(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			ums:        user_mgmt.NewUserMgmtService(appSetting),
		}
	})
	return cService
}

// Create creates a new Catalog.
func (cs *catalogService) Create(ctx context.Context, req *interfaces.CatalogRequest) (string, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Create catalog")
	defer span.End()

	// 判断userid是否有创建业务知识网络的权限（策略决策）
	err := cs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_CATALOG,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return "", err
	}

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	catalogType := interfaces.CatalogTypePhysical
	if req.ConnectorType == "" {
		catalogType = interfaces.CatalogTypeLogical
	} else {
		// 验证敏感字段是否为合法 RSA 密文，获取明文用于连接测试
		sensitiveFields := factory.GetFactory().GetSensitiveFields(req.ConnectorType)
		decryptedConfig, err := cs.validateAndDecryptSensitiveFields(sensitiveFields, req.ConnectorCfg)
		if err != nil {
			otellog.LogError(ctx, "Failed to validate sensitive fields", err)
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}

		// 用解密后的明文 config 创建 connector 并测试连接
		connectorCfg := interfaces.ConnectorConfig(decryptedConfig)
		connector, err := factory.GetFactory().CreateConnectorInstance(ctx, req.ConnectorType, connectorCfg)
		if err != nil {
			otellog.LogError(ctx, "Failed to create connector", err)
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_CreateFailed).
				WithErrorDetails(err.Error())
		}

		if err := connector.TestConnection(ctx); err != nil {
			otellog.LogError(ctx, "Failed to test connection to data source", err)
			_ = connector.Close(ctx)
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_TestConnectionFailed).
				WithErrorDetails(err.Error())
		}
		defer func() { _ = connector.Close(ctx) }()
	}

	now := time.Now().UnixMilli()
	id := req.ID
	if id == "" {
		id = xid.New().String()
	}
	catalog := &interfaces.Catalog{
		ID:                 id,
		Name:               req.Name,
		Tags:               req.Tags,
		Description:        req.Description,
		Type:               catalogType,
		ConnectorType:      req.ConnectorType,
		ConnectorCfg:       req.ConnectorCfg,
		HealthCheckEnabled: true,
		CatalogHealthCheckStatus: interfaces.CatalogHealthCheckStatus{
			HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
			LastCheckTime:     now,
		},
		Creator:    accountInfo,
		CreateTime: now,
		Updater:    accountInfo,
		UpdateTime: now,
	}

	err = cs.ca.Create(ctx, catalog)
	if err != nil {
		otellog.LogError(ctx, "Create catalog failed", err)
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	// 注册资源
	err = cs.ps.CreateResources(ctx, []interfaces.PermissionResource{{
		ID:   catalog.ID,
		Type: interfaces.RESOURCE_TYPE_CATALOG,
		Name: catalog.Name,
	}}, interfaces.COMMON_OPERATIONS)
	if err != nil {
		logger.Errorf("CreateResources error: %s", err.Error())
		span.SetStatus(codes.Error, "创建目录资源失败")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Catalog_InternalError_CreateResourcesFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return catalog.ID, nil
}

// Get retrieves a Catalog by ID.
func (cs *catalogService) GetByID(ctx context.Context, id string, withSensitiveFields bool) (*interfaces.Catalog, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Get catalog")
	defer span.End()

	catalog, err := cs.ca.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG, []string{catalog.ID},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return nil, err
	}

	if resrc, exist := matchResoucesMap[catalog.ID]; exist {
		catalog.Operations = resrc.Operations // 用户当前有权限的操作
	} else {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", interfaces.OPERATION_TYPE_VIEW_DETAIL))
	}

	accountInfos := []*interfaces.AccountInfo{&catalog.Creator, &catalog.Updater}
	err = cs.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Catalog_InternalError_GetAccountNamesFailed).WithErrorDetails(err.Error())
	}

	if !withSensitiveFields {
		// 移除敏感字段，不返回给前端
		cs.removeSensitiveFields(catalog)
	} else {
		// 验证敏感字段是否为合法 RSA 密文，获取明文用于连接测试
		sensitiveFields := factory.GetFactory().GetSensitiveFields(catalog.ConnectorType)
		decryptedConfig, err := cs.decryptSensitiveFields(sensitiveFields, catalog.ConnectorCfg)
		if err != nil {
			otellog.LogError(ctx, "Failed to validate sensitive fields", err)
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}
		catalog.ConnectorCfg = decryptedConfig
	}

	span.SetStatus(codes.Ok, "")
	return catalog, nil
}

// GetByIDs retrieves a Catalog by IDs.
func (cs *catalogService) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Catalog, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Get catalogs")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.Catalog{}, nil
	}

	catalogs, err := cs.ca.GetByIDs(ctx, ids)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	// 移除敏感字段，不返回给前端
	for _, c := range catalogs {
		cs.removeSensitiveFields(c)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG, ids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return nil, err
	}

	accountInfos := make([]*interfaces.AccountInfo, 0)
	for _, c := range catalogs {
		if resrc, exist := matchResoucesMap[c.ID]; exist {
			c.Operations = resrc.Operations // 用户当前有权限的操作
		} else {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", interfaces.OPERATION_TYPE_VIEW_DETAIL))
		}
		accountInfos = append(accountInfos, &c.Creator, &c.Updater)
	}

	err = cs.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Catalog_InternalError_GetAccountNamesFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return catalogs, nil
}

// List lists Catalogs with filters.
func (cs *catalogService) List(ctx context.Context, params interfaces.CatalogsQueryParams) ([]*interfaces.Catalog, int64, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "List catalogs")
	defer span.End()

	// 查询所有catalog的ID
	ids, err := cs.ca.ListIDs(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List catalog IDs failed")
		return []*interfaces.Catalog{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.Catalog{}, 0, nil
	}

	// 使用分批处理的方式过滤权限，每批处理1万个ID
	batchSize := 10000
	// 所有有权限的catalog及其操作权限
	matchResourceOpsMap := make(map[string]interfaces.PermissionResourceOps)

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batchIDs := ids[i:end]

		var batchMatchResources map[string]interfaces.PermissionResourceOps
		// 校验权限管理的操作权限
		batchMatchResources, err = cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG,
			batchIDs, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
		if err != nil {
			span.SetStatus(codes.Error, "Filter resources error")
			return []*interfaces.Catalog{}, 0, err
		}

		// 合并结果
		for _, resourceOps := range batchMatchResources {
			matchResourceOpsMap[resourceOps.ResourceID] = resourceOps
		}
	}

	// 提取有权限的catalog ID，保持与ids的顺序一致
	authorizedIDs := make([]string, 0, len(matchResourceOpsMap))
	for _, id := range ids {
		if _, exist := matchResourceOpsMap[id]; exist {
			authorizedIDs = append(authorizedIDs, id)
		}
	}
	total := int64(len(authorizedIDs))

	// 如果没有有权限的catalog，直接返回空结果
	if total == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.Catalog{}, total, nil
	}

	// 根据有权限的ID数组应用分页
	if params.Limit != -1 {
		// 分页处理authorizedIDs
		// 检查起始位置是否越界
		if params.Offset < 0 || params.Offset >= len(authorizedIDs) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.Catalog{}, total, nil
		}
		// 计算结束位置
		end := params.Offset + params.Limit
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		// 只查询当前页的catalog ID
		authorizedIDs = authorizedIDs[params.Offset:end]
	}

	// 根据有权限的ID数组查询完整catalog
	// 分批处理，每批10000个ids, 避免prepared statement contains too many placeholders错误
	catalogs := make([]*interfaces.Catalog, 0, len(authorizedIDs))
	queryBatchSize := 10000
	for i := 0; i < len(authorizedIDs); i += queryBatchSize {
		end := i + queryBatchSize
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		batchIDs := authorizedIDs[i:end]

		batchCatalogs, err := cs.ca.GetByIDs(ctx, batchIDs)
		if err != nil {
			span.SetStatus(codes.Error, "Get catalogs by IDs failed")
			return []*interfaces.Catalog{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
				WithErrorDetails(err.Error())
		}

		catalogs = append(catalogs, batchCatalogs...)
	}

	// 设置catalog操作权限
	for _, c := range catalogs {
		if resrc, exist := matchResourceOpsMap[c.ID]; exist {
			c.Operations = resrc.Operations // 用户当前有权限的操作
		}
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(catalogs)*2)
	for _, c := range catalogs {
		accountInfos = append(accountInfos, &c.Creator, &c.Updater)
	}

	err = cs.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.Catalog{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	// 移除敏感字段，不返回给前端
	for _, c := range catalogs {
		cs.removeSensitiveFields(c)
	}

	span.SetStatus(codes.Ok, "")
	return catalogs, total, nil
}

// Update updates a Catalog.
func (cs *catalogService) Update(ctx context.Context, id string, req *interfaces.CatalogRequest) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Update catalog")
	defer span.End()

	catalog := req.OriginCatalog
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
	}

	// 判断userid是否有修改权限
	err := cs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_CATALOG,
		ID:   catalog.ID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// Apply updates
	catalog.Name = req.Name
	catalog.Tags = req.Tags
	catalog.Description = req.Description

	if req.ConnectorType != "" {
		// 验证敏感字段是否为合法 RSA 密文，获取明文用于连接测试
		sensitiveFields := factory.GetFactory().GetSensitiveFields(req.ConnectorType)
		decryptedConfig, err := cs.validateAndDecryptSensitiveFields(sensitiveFields, req.ConnectorCfg)
		if err != nil {
			otellog.LogError(ctx, "Failed to validate sensitive fields", err)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}

		// 用解密后的明文 config 创建 connector 并测试连接
		connectorCfg := interfaces.ConnectorConfig(decryptedConfig)
		connector, err := factory.GetFactory().CreateConnectorInstance(ctx, req.ConnectorType, connectorCfg)
		if err != nil {
			otellog.LogError(ctx, "Failed to create connector", err)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_CreateFailed).
				WithErrorDetails(err.Error())
		}

		if err := connector.TestConnection(ctx); err != nil {
			otellog.LogError(ctx, "Failed to test connection to data source", err)
			_ = connector.Close(ctx)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_TestConnectionFailed).
				WithErrorDetails(err.Error())
		}
		defer func() { _ = connector.Close(ctx) }()

		// req.ConnectorConfig 已在 validateAndDecryptSensitiveFields 中加上 ENC: 前缀
		catalog.ConnectorCfg = req.ConnectorCfg
	}

	// Get account info
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	now := time.Now().UnixMilli()
	catalog.Updater = accountInfo
	catalog.UpdateTime = now
	catalog.CatalogHealthCheckStatus = interfaces.CatalogHealthCheckStatus{
		HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
		LastCheckTime:     now,
	}

	if err := cs.ca.Update(ctx, catalog); err != nil {
		span.SetStatus(codes.Error, "Update catalog failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	// 请求更新资源名称的接口，更新资源的名称
	if req.IfNameModify {
		err = cs.ps.UpdateResource(ctx, interfaces.PermissionResource{
			ID:   catalog.ID,
			Type: interfaces.RESOURCE_TYPE_CATALOG,
			Name: catalog.Name,
		})
		if err != nil {
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteByIDs deletes Catalogs by IDs.
func (cs *catalogService) DeleteByIDs(ctx context.Context, ids []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete catalogs")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	// 判断userid是否有删除权限
	matchResoucesMap, err := cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG, ids,
		[]string{interfaces.OPERATION_TYPE_DELETE}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return err
	}

	// 检查是否有删除权限
	if len(matchResoucesMap) != len(ids) {
		// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
		for _, id := range ids {
			if _, exist := matchResoucesMap[id]; !exist {
				return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
					WithErrorDetails("Access denied: insufficient permissions for catalog's delete operation.")
			}
		}
	}

	if err := cs.ca.DeleteByIDs(ctx, ids); err != nil {
		span.SetStatus(codes.Error, "Delete catalogs failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	// 清理关联resource数据
	if err := cs.ra.DeleteByCatalogIDs(ctx, ids); err != nil {
		span.SetStatus(codes.Error, "Delete resources failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = cs.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_CATALOG, ids)
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExistByID checks if a Catalog exists by ID.
func (cs *catalogService) CheckExistByID(ctx context.Context, id string) (bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Check catalog exist by ID")
	defer span.End()

	catalog, err := cs.ca.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "GetByID failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return catalog != nil, nil
}

// CheckExistByName checks if a Catalog exists by name.
func (cs *catalogService) CheckExistByName(ctx context.Context, name string) (bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Check catalog exist by name")
	defer span.End()

	catalog, err := cs.ca.GetByName(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, "GetByName failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return catalog != nil, nil
}

// TestConnection tests catalog connection.
func (cs *catalogService) TestConnection(ctx context.Context, catalog *interfaces.Catalog) (*interfaces.CatalogHealthCheckStatus, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Test catalog connection")
	defer span.End()

	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
	}

	result := catalog.CatalogHealthCheckStatus
	span.SetStatus(codes.Ok, "")
	return &result, nil
}

// validateAndDecryptSensitiveFields 验证敏感字段是否为合法 RSA 密文，
// 返回解密后的明文 config（用于连接测试），同时在原始 config 中加上 ENC: 前缀（用于存储）。
// 如果 cipher 为 nil（加密未启用），直接返回原始 config 的拷贝作为 decryptedConfig，不做验证。
func (cs *catalogService) validateAndDecryptSensitiveFields(
	sensitiveFields []string, config map[string]any,
) (decryptedConfig map[string]any, err error) {
	// 拷贝 config 作为 decryptedConfig
	decryptedConfig = make(map[string]any, len(config))
	for k, v := range config {
		decryptedConfig[k] = v
	}

	if cs.cipher == nil {
		return decryptedConfig, nil
	}

	for _, field := range sensitiveFields {
		val, ok := config[field].(string)
		if !ok || val == "" {
			continue
		}
		// 尝试用私钥解密，验证是否为合法密文
		decrypted, decryptErr := cs.cipher.Decrypt(val)
		if decryptErr != nil {
			return nil, fmt.Errorf("field %s: %w", field, decryptErr)
		}
		// 解密成功：明文放入 decryptedConfig，原始 config 加上 ENC: 前缀
		decryptedConfig[field] = decrypted
		config[field] = EncryptedPrefix + val
	}
	return decryptedConfig, nil
}

// removeSensitiveFields 从 ConnectorConfig 中移除敏感字段，用于 GET/List 返回
func (cs *catalogService) removeSensitiveFields(catalog *interfaces.Catalog) {
	if catalog == nil || catalog.ConnectorType == "" {
		return
	}
	sensitiveFields := factory.GetFactory().GetSensitiveFields(catalog.ConnectorType)
	for _, field := range sensitiveFields {
		delete(catalog.ConnectorCfg, field)
	}
}

// decryptSensitiveFields 验证敏感字段是否为合法 RSA 密文，
// 返回解密后的明文 config（用于连接），数据从数据库获取而来，需要先去除ENC前缀，再解密
// 如果 cipher 为 nil（加密未启用），直接返回原始 config 的拷贝作为 decryptedConfig，不做验证。
func (cs *catalogService) decryptSensitiveFields(sensitiveFields []string,
	config map[string]any) (decryptedConfig map[string]any, err error) {

	// 拷贝 config 作为 decryptedConfig
	decryptedConfig = make(map[string]any, len(config))
	for k, v := range config {
		decryptedConfig[k] = v
	}

	if cs.cipher == nil {
		return decryptedConfig, nil
	}

	for _, field := range sensitiveFields {
		val, ok := config[field].(string)
		if !ok || val == "" {
			continue
		}
		// 尝试用私钥解密，验证是否为合法密文
		if !strings.HasPrefix(val, EncryptedPrefix) {
			return nil, fmt.Errorf("field %s: %w", field, errors.New("not encrypted"))
		} else {
			val = val[len(EncryptedPrefix):]
		}
		decrypted, decryptErr := cs.cipher.Decrypt(val)
		if decryptErr != nil {
			return nil, fmt.Errorf("field %s: %w", field, decryptErr)
		}
		// 解密成功：明文放入 decryptedConfig，原始 config 加上 ENC: 前缀
		decryptedConfig[field] = decrypted
		config[field] = EncryptedPrefix + val
	}
	return decryptedConfig, nil
}

func (cs *catalogService) UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "UpdateMetadata")
	defer span.End()

	err := cs.ca.UpdateMetadata(ctx, id, metadata)
	if err != nil {
		otellog.LogError(ctx, "Update metadata failed", err)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	return nil
}

// ListCatalogSrcs lists Catalog Sources with filters.
func (cs *catalogService) ListCatalogSrcs(ctx context.Context, params interfaces.ListCatalogsQueryParams) ([]*interfaces.ListCatalogEntry, int64, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "ListCatalogSrcs catalogs")
	defer span.End()

	// 先查询所有catalog源的ID
	ids, err := cs.ca.ListCatalogSrcsIDs(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "ListCatalogSrcsIDs failed")
		return []*interfaces.ListCatalogEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	if len(ids) == 0 {
		return []*interfaces.ListCatalogEntry{}, 0, nil
	}

	// 使用分批处理的方式过滤权限，每批处理1万个ID
	batchSize := 10000
	// 所有有权限的catalog及其操作权限
	matchResourceOpsMap := make(map[string]interfaces.PermissionResourceOps)

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batchIDs := ids[i:end]

		var batchMatchResources map[string]interfaces.PermissionResourceOps
		// 校验权限管理的操作权限
		batchMatchResources, err = cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG,
			batchIDs, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false, interfaces.COMMON_OPERATIONS)
		if err != nil {
			return []*interfaces.ListCatalogEntry{}, 0, err
		}

		// 合并结果
		for _, resourceOps := range batchMatchResources {
			matchResourceOpsMap[resourceOps.ResourceID] = resourceOps
		}
	}

	// 提取有权限的catalog ID，保持与ids的顺序一致
	authorizedIDs := make([]string, 0, len(matchResourceOpsMap))
	for _, id := range ids {
		if _, exist := matchResourceOpsMap[id]; exist {
			authorizedIDs = append(authorizedIDs, id)
		}
	}
	total := int64(len(authorizedIDs))

	// 如果没有有权限的catalog，直接返回空结果
	if total == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.ListCatalogEntry{}, total, nil
	}

	// 根据有权限的ID数组应用分页
	if params.Limit != -1 {
		// 分页处理authorizedIDs
		// 检查起始位置是否越界
		if params.Offset < 0 || params.Offset >= len(authorizedIDs) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ListCatalogEntry{}, total, nil
		}
		// 计算结束位置
		end := params.Offset + params.Limit
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		// 只查询当前页的catalog ID
		authorizedIDs = authorizedIDs[params.Offset:end]
	}

	// 根据有权限的ID数组查询完整catalog
	// 分批处理，每批10000个ids, 避免prepared statement contains too many placeholders错误
	entries := make([]*interfaces.ListCatalogEntry, 0, len(authorizedIDs))
	queryBatchSize := 10000
	for i := 0; i < len(authorizedIDs); i += queryBatchSize {
		end := i + queryBatchSize
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		batchIDs := authorizedIDs[i:end]

		batchEntries, err := cs.ca.ListCatalogSrcsByIDs(ctx, batchIDs)
		if err != nil {
			span.SetStatus(codes.Error, "ListCatalogSrcsByIDs failed")
			return []*interfaces.ListCatalogEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
				WithErrorDetails(err.Error())
		}

		entries = append(entries, batchEntries...)
	}

	// 据有权限的ID过滤结果
	results := make([]*interfaces.ListCatalogEntry, 0)
	for _, entry := range entries {
		if _, exist := matchResourceOpsMap[entry.ID]; exist {
			results = append(results, entry)
		}
	}

	span.SetStatus(codes.Ok, "")
	return results, total, nil
}
