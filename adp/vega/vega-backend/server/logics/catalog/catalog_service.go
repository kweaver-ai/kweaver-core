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

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	kwcrypto "github.com/kweaver-ai/kweaver-go-lib/crypto"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Create catalog")
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
			logger.Errorf("Failed to validate sensitive fields: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to validate sensitive fields: %v", err))
			span.SetStatus(codes.Error, "Validate sensitive fields failed")
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}

		// 用解密后的明文 config 创建 connector 并测试连接
		connectorCfg := interfaces.ConnectorConfig(decryptedConfig)
		connector, err := factory.GetFactory().CreateConnectorInstance(ctx, req.ConnectorType, connectorCfg)
		if err != nil {
			logger.Errorf("Failed to create connector: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to create connector: %v", err))
			span.SetStatus(codes.Error, "Create connector failed")
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_CreateFailed).
				WithErrorDetails(err.Error())
		}

		if err := connector.TestConnection(ctx); err != nil {
			logger.Errorf("Failed to test connection to data source: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to test connection to data source: %v", err))
			span.SetStatus(codes.Error, "Connection failed")
			connector.Close(ctx)
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_TestConnectionFailed).
				WithErrorDetails(err.Error())
		}
		defer connector.Close(ctx)
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
		logger.Errorf("Create catalog failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create catalog failed: %v", err))
		span.SetStatus(codes.Error, "Create catalog failed")
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Get catalog")
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
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
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
			logger.Errorf("Failed to validate sensitive fields: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to validate sensitive fields: %v", err))
			span.SetStatus(codes.Error, "Validate sensitive fields failed")
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Get catalogs")
	defer span.End()

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
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
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
	ctx, span := ar_trace.Tracer.Start(ctx, "List catalogs")
	defer span.End()

	catalogsArr, _, err := cs.ca.List(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List catalogs failed")
		return []*interfaces.Catalog{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	// 处理资源id
	ids := make([]string, 0)
	for _, m := range catalogsArr {
		ids = append(ids, m.ID)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG, ids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return []*interfaces.Catalog{}, 0, err
	}

	catalogs := make([]*interfaces.Catalog, 0)
	for _, c := range catalogsArr {
		// 只留下有权限的模型
		if resrc, exist := matchResoucesMap[c.ID]; exist {
			c.Operations = resrc.Operations // 用户当前有权限的操作
			catalogs = append(catalogs, c)
		}
	}
	total := int64(len(catalogs))

	// limit = -1,则返回所有
	if params.Limit != -1 {
		// 分页
		// 检查起始位置是否越界
		if params.Offset < 0 || params.Offset >= len(catalogs) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.Catalog{}, total, nil
		}
		// 计算结束位置
		end := params.Offset + params.Limit
		if end > len(catalogs) {
			end = len(catalogs)
		}

		catalogs = catalogs[params.Offset:end]
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Update catalog")
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
			logger.Errorf("Failed to validate sensitive fields: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to validate sensitive fields: %v", err))
			span.SetStatus(codes.Error, "Validate sensitive fields failed")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}

		// 用解密后的明文 config 创建 connector 并测试连接
		connectorCfg := interfaces.ConnectorConfig(decryptedConfig)
		connector, err := factory.GetFactory().CreateConnectorInstance(ctx, req.ConnectorType, connectorCfg)
		if err != nil {
			logger.Errorf("Failed to create connector: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to create connector: %v", err))
			span.SetStatus(codes.Error, "Create connector failed")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_CreateFailed).
				WithErrorDetails(err.Error())
		}

		if err := connector.TestConnection(ctx); err != nil {
			logger.Errorf("Failed to test connection to data source: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to test connection to data source: %v", err))
			span.SetStatus(codes.Error, "Connection failed")
			connector.Close(ctx)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InternalError_TestConnectionFailed).
				WithErrorDetails(err.Error())
		}
		defer connector.Close(ctx)

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
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete catalogs")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	// 判断userid是否有删除权限
	matchResoucesMap, err := cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG, ids,
		[]string{interfaces.OPERATION_TYPE_DELETE}, true)
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Check catalog exist by ID")
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Check catalog exist by name")
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
	ctx, span := ar_trace.Tracer.Start(ctx, "Test catalog connection")
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
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateMetadata")
	defer span.End()

	err := cs.ca.UpdateMetadata(ctx, id, metadata)
	if err != nil {
		logger.Errorf("Update metadata failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Update metadata failed: %v", err))
		span.SetStatus(codes.Error, "Update metadata failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	return nil
}

// ListCatalogSrcs lists Catalog Sources with filters.
func (cs *catalogService) ListCatalogSrcs(ctx context.Context, params interfaces.ListCatalogsQueryParams) ([]*interfaces.ListCatalogEntry, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListCatalogSrcs catalogs")
	defer span.End()

	// 使用ca.ListCatalogSrcs函数查询catalogs
	entries, total, err := cs.ca.ListCatalogSrcs(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "ListCatalogSrcs catalogs failed")
		return []*interfaces.ListCatalogEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	if total == 0 {
		return []*interfaces.ListCatalogEntry{}, 0, nil
	}

	// 根据权限过滤有显示权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	ids := make([]string, 0)
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}

	// 分批处理，每批1万个ids, fix权限接口报错prepared statement contains too many placeholders
	batchSize := 10000
	// 所有有权限的catalog id
	matchResourceIDMap := make(map[string]bool)

	for i := 0; i < int(total); i += batchSize {
		end := i + batchSize
		if end > int(total) {
			end = int(total)
		}
		batchIDs := ids[i:end]

		var batchMatchResources map[string]interfaces.PermissionResourceOps
		// 校验权限管理的操作权限
		batchMatchResources, err = cs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_CATALOG,
			batchIDs, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
		if err != nil {
			return nil, 0, err
		}

		// 合并结果
		for _, resourceOps := range batchMatchResources {
			matchResourceIDMap[resourceOps.ResourceID] = true
		}
	}

	// 遍历对象
	results := make([]*interfaces.ListCatalogEntry, 0)
	for _, entry := range entries {
		if matchResourceIDMap[entry.ID] {
			results = append(results, entry)
		}
	}

	resTotal := len(results)
	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= resTotal {
		return []*interfaces.ListCatalogEntry{}, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > resTotal {
		end = resTotal
	}

	span.SetStatus(codes.Ok, "")
	return results[params.Offset:end], int64(resTotal), nil
}
