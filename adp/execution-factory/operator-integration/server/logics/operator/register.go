package operator

import (
	"context"
	inErr "errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// RegisterOperatorByOpenAPI 算子注册
func (m *operatorManager) RegisterOperatorByOpenAPI(ctx context.Context, req *interfaces.OperatorRegisterReq, userID string) (resultList []*interfaces.OperatorRegisterResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 检查是否有新建权限
	var accessor *interfaces.AuthAccessor
	accessor, err = m.AuthService.GetAccessor(ctx, userID)
	if err != nil {
		return
	}
	err = m.AuthService.CheckCreatePermission(ctx, accessor, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}
	// 检查请求信息
	isDataSource, err := checkIsDataSource(ctx, req.OperatorInfo.ExecutionMode, req.OperatorInfo.IsDataSource)
	if err != nil {
		return
	}
	// 解析API文档
	metadataDBs, err := m.checkAndParserOpenAPIOperator(ctx, req)
	if err != nil {
		return
	}
	// 初始化算子注册状态
	operatorRegisterStatus := interfaces.BizStatusUnpublish
	// 只允许单个算子直接注册发布
	if req.DirectPublish && len(metadataDBs) > 1 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorDirectPublishErr, "direct_publish only support one api")
		return
	} else if req.DirectPublish && len(metadataDBs) == 1 {
		operatorRegisterStatus = interfaces.BizStatusPublished
	}
	resultList = []*interfaces.OperatorRegisterResp{}
	// 遍历算子列表，解析元数据
	for _, metadataDB := range metadataDBs {
		resultList = append(resultList, m.registerOperator(ctx, req, metadataDB, accessor, operatorRegisterStatus, isDataSource))
	}
	return resultList, nil
}

// UpdateOperatorByOpenAPI 算子更新
func (m *operatorManager) UpdateOperatorByOpenAPI(ctx context.Context, req *interfaces.OperatorUpdateReq, userID string) (resultList []*interfaces.OperatorRegisterResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	var isDataSource bool
	isDataSource, err = checkIsDataSource(ctx, req.OperatorInfo.ExecutionMode, req.OperatorInfo.IsDataSource)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("check is data source failed, err: %v", err)
		return
	}
	// 解析API文档
	metadataDBs, err := m.checkAndParserOpenAPIOperator(ctx, req.OperatorRegisterReq)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("check and parser openapi operator failed, err: %v", err)
		return
	}
	// 编辑仅允许单个算子
	if len(metadataDBs) > 1 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorEditLimit, "edit operator only support one api")
		return
	} else if len(metadataDBs) == 0 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorEditLimit, "edit operator failed, no api found")
		return
	}
	resultList = []*interfaces.OperatorRegisterResp{}
	result := &interfaces.OperatorRegisterResp{
		Status:     interfaces.ResultStatusFailed,
		OperatorID: req.OperatorID,
	}
	funcInputEdit := &interfaces.FunctionInputEdit{}
	if req.FunctionInput != nil {
		funcInputEdit = &interfaces.FunctionInputEdit{
			Inputs:          req.FunctionInput.Inputs,
			Outputs:         req.FunctionInput.Outputs,
			ScriptType:      req.FunctionInput.ScriptType,
			Code:            req.FunctionInput.Code,
			Dependencies:    req.FunctionInput.Dependencies,
			DependenciesURL: req.FunctionInput.DependenciesURL,
		}
	}
	updateReq := &interfaces.OperatorEditReq{
		OperatorID:  req.OperatorID,
		Name:        metadataDBs[0].GetSummary(),
		Description: metadataDBs[0].GetDescription(),
		OperatorInfoEdit: &interfaces.OperatorInfoEdit{
			Type:          req.OperatorInfo.Type,
			ExecutionMode: req.OperatorInfo.ExecutionMode,
			Category:      req.OperatorInfo.Category,
			Source:        req.OperatorInfo.Source,
			IsDataSource:  req.OperatorInfo.IsDataSource,
		},
		OperatorExecuteControl: req.OperatorExecuteControl,
		ExtendInfo:             req.ExtendInfo,
		MetadataType:           req.MetadataType,
		UserID:                 userID,
		OpenAPIInput: &interfaces.OpenAPIInput{
			Data: []byte(req.Data),
		},
		FunctionInputEdit: funcInputEdit,
	}
	operator, metadataDB, accessor, needUpdateMetadata, err := m.preCheckEdit(ctx, updateReq, req.DirectPublish)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("[UpdateOperatorByOpenAPI] pre check edit failed, err: %v", err)
		return
	}
	metadataDB.SetMethod(metadataDBs[0].GetMethod())
	metadataDB.SetPath(metadataDBs[0].GetPath())
	editRes, err := m.editOperator(ctx, updateReq, operator, metadataDB, needUpdateMetadata, req.DirectPublish, isDataSource)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("edit operator failed, err: %v", err)
		httpErr := &errors.HTTPError{}
		if inErr.As(err, &httpErr) {
			result.Error = err
		} else {
			result.Error = errors.NewHTTPError(ctx, http.StatusConflict, errors.ErrExtOperatorEditFailed, err.Error())
		}
	} else {
		result.OperatorID = editRes.OperatorID
		result.Status = interfaces.ResultStatusSuccess
		result.Version = editRes.Version
	}
	resultList = append(resultList, result)
	// 记录审计日志
	go func() {
		accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			m.Logger.WithContext(ctx).Warnf("[UpdateOperatorByOpenAPI] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationEdit,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OperatorID,
				Name: operator.Name,
			},
		})
		if operator.Status != interfaces.BizStatusPublished.String() {
			return
		}
		// 发布操作
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationPublish,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OperatorID,
				Name: operator.Name,
			},
		})
	}()
	return resultList, nil
}

// validateOperator 校验算子信息
func (m *operatorManager) validateOperator(ctx context.Context, metadataDB interfaces.IMetadataDB) (err error) {
	if metadataDB.GetErrMessage() != "" {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, metadataDB.GetErrMessage())
		return
	}
	// 校验算子名称
	err = m.Validator.ValidateOperatorName(ctx, metadataDB.GetSummary())
	if err != nil {
		return
	}
	// 校验算子描述
	err = m.Validator.ValidateOperatorDesc(ctx, metadataDB.GetDescription())
	return
}

// checkAndParserOpenAPIOperator 检查并解析OpenAPI算子
func (m *operatorManager) checkAndParserOpenAPIOperator(ctx context.Context, req *interfaces.OperatorRegisterReq) (metadataDBs []interfaces.IMetadataDB, err error) {
	// 检查算子类型
	if !m.CategoryManager.CheckCategory(req.OperatorInfo.Category) {
		m.Logger.WithContext(ctx).Warnf("invalid operator category, category: %s", req.OperatorInfo.Category)
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCategoryTypeInvalid, "invalid operator category")
		return
	}
	switch req.MetadataType {
	case interfaces.MetadataTypeAPI:
		// 解析API数据
		metadataDBs, err = m.MetadataService.ParseMetadata(ctx, req.MetadataType, &interfaces.OpenAPIInput{
			Data: []byte(req.Data),
		})
	case interfaces.MetadataTypeFunc:
		metadataDBs, err = m.MetadataService.ParseMetadata(ctx, req.MetadataType, req.FunctionInput)
	default:
		m.Logger.WithContext(ctx).Warnf("invalid metadata type, metadata_type: %s", req.MetadataType)
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "invalid metadata type")
	}
	if err != nil {
		return
	}
	// 检查Items长度
	err = m.Validator.ValidateOperatorImportCount(ctx, int64(len(metadataDBs)))
	return
}

func (m *operatorManager) registerOperator(ctx context.Context, req *interfaces.OperatorRegisterReq, metadataDB interfaces.IMetadataDB,
	accessor *interfaces.AuthAccessor, status interfaces.BizStatus, isDataSource bool) (result *interfaces.OperatorRegisterResp) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, nil)
	result = &interfaces.OperatorRegisterResp{
		Status: interfaces.ResultStatusFailed,
	}
	var operator *model.OperatorRegisterDB
	var err error
	defer func() {
		if err != nil {
			result.Error = err
			return
		}
		result.OperatorID = operator.OperatorID
		result.Version = operator.MetadataVersion
		result.Status = interfaces.ResultStatusSuccess
	}()
	err = m.validateOperator(ctx, metadataDB)
	if err != nil {
		return
	}
	// 设置创建人和更新人
	metadataDB.SetCreateInfo(accessor.ID)
	metadataDB.SetUpdateInfo(accessor.ID)
	metadataDB.SetVersion(uuid.New().String())
	operator = &model.OperatorRegisterDB{
		Name:            metadataDB.GetSummary(),
		MetadataVersion: metadataDB.GetVersion(),
		MetadataType:    metadataDB.GetType(),
		Status:          status.String(),
		OperatorType:    string(req.OperatorInfo.Type),
		ExecutionMode:   string(req.OperatorInfo.ExecutionMode),
		Category:        string(req.OperatorInfo.Category),
		Source:          req.OperatorInfo.Source,
		ExecuteControl:  utils.ObjectToJSON(req.OperatorExecuteControl),
		ExtendInfo:      utils.ObjectToJSON(req.ExtendInfo),
		CreateUser:      accessor.ID,
		CreateTime:      time.Now().UnixNano(),
		UpdateUser:      accessor.ID,
		UpdateTime:      time.Now().UnixNano(),
		IsDataSource:    isDataSource,
	}
	// 1. 检查算子是否存在
	err = m.checkDuplicateName(ctx, operator.Name, operator.OperatorID)
	if err != nil {
		return
	}
	tx, err := m.DBTx.GetTx(ctx)
	if err != nil {
		err = fmt.Errorf("get tx failed, err: %v", err)
		m.Logger.WithContext(ctx).Errorf("get tx failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed, "get tx failed")
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	// 2. 插入元数据
	version, err := m.MetadataService.RegisterMetadata(ctx, tx, metadataDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("insert metadata failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed, "insert metadata failed")
		return
	}
	// 3. 插入算子
	operator.MetadataVersion = version
	opID, err := m.DBOperatorManager.InsertOperator(ctx, tx, operator)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("insert operator failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed, fmt.Errorf("insert operator failed, err: %v", err))
		return
	}
	// 查找
	operator.OperatorID = opID

	// 关联业务域
	businessDomainID, _ := common.GetBusinessDomainFromCtx(ctx)
	err = m.BusinessDomainService.AssociateResource(ctx, businessDomainID, opID, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}

	// 触发新建策略，创建人默认拥有对当前资源的所有操作权限
	err = m.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
		ID:   operator.OperatorID,
		Type: string(interfaces.AuthResourceTypeOperator),
		Name: operator.Name,
	})
	if err != nil {
		return
	}
	// 注册填写了 direct_publish 为 true，直接发布
	if operator.Status == interfaces.BizStatusPublished.String() {
		// 发布操作
		err = m.publishRelease(ctx, tx, operator, operator.UpdateUser)
		if err != nil {
			return
		}
	}
	go func() {
		accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			m.Logger.WithContext(ctx).Warnf("[registerOperator] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationCreate,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OperatorID,
				Name: operator.Name,
			},
		})
		if operator.Status != interfaces.BizStatusPublished.String() {
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationPublish,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OperatorID,
				Name: operator.Name,
			},
		})
	}()
	return
}
