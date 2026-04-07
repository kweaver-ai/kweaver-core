package operator

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metadata"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// GetOperatorInfoByOperatorID 根据算子ID或版本获取算子(外部接口) -- 查询算子详情
func (m *operatorManager) GetOperatorInfoByOperatorID(ctx context.Context, req *interfaces.GetOperatorInfoByOperatorIDReq) (result *interfaces.OperatorDataInfo, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 获取算子信息
	operator, metadataDB, err := m.getOperatorRegisterInfo(ctx, req.OperatorID)
	if err != nil {
		return
	}
	if common.IsPublicAPIFromCtx(ctx) {
		// 检查查看权限
		var accessor *interfaces.AuthAccessor
		accessor, err = m.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		err = m.AuthService.CheckViewPermission(ctx, accessor, req.OperatorID, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}
	}
	userIDs, result, err := m.assembleOperatorResult(ctx, operator, metadataDB)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	userMap, err := m.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("get users info failed, err: %v", err)
		return
	}
	result.CreateUser = utils.GetValueOrDefault(userMap, result.CreateUser, interfaces.UnknownUser)
	result.UpdateUser = utils.GetValueOrDefault(userMap, result.UpdateUser, interfaces.UnknownUser)
	return
}

// getOperatorInfo 获取算子信息
func (m *operatorManager) getOperatorRegisterInfo(ctx context.Context, operatorID string) (operator *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB, err error) {
	// 查询算子信息
	has, operator, err := m.DBOperatorManager.SelectByOperatorID(ctx, nil, operatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select operator failed, OperatorID: %s, err: %v", operatorID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !has {
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOperatorNotFound, fmt.Sprintf("operator not found, OperatorID: %s", operatorID))
		return
	}
	// 查询算子元数据信息
	metadataDB, err = m.MetadataService.GetMetadataByVersion(ctx, interfaces.MetadataType(operator.MetadataType), operator.MetadataVersion)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select metadata failed, OperatorID: %s, Version: %s, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("select metadata failed, OperatorID: %s, Version: %s, Err: %s",
			operator.OperatorID, operator.MetadataVersion, err.Error()))
	}
	return
}

// GetOperatorQueryPage 获取算子列表
func (m *operatorManager) GetOperatorQueryPage(ctx context.Context, req *interfaces.PageQueryRequest) (result *interfaces.PageQueryResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	result = &interfaces.PageQueryResponse{
		CommonPageResult: interfaces.CommonPageResult{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}
	var operatorList []*model.OperatorRegisterDB
	authResp, resourceToBdMap, err := m.queryOperatorConfigList(ctx, req)
	if err != nil {
		return nil, err
	}
	operatorList = authResp.Data
	result.CommonPageResult = authResp.CommonPageResult
	if len(operatorList) == 0 {
		result.Data = []*interfaces.OperatorDataInfo{}
		return
	}
	// 批量遍历
	sourceMap := map[model.SourceType][]string{}
	var userList []string
	for _, operator := range operatorList {
		switch operator.MetadataType {
		case string(interfaces.MetadataTypeAPI):
			sourceMap[model.SourceTypeOpenAPI] = append(sourceMap[model.SourceTypeOpenAPI], operator.MetadataVersion)
		case string(interfaces.MetadataTypeFunc):
			sourceMap[model.SourceTypeFunction] = append(sourceMap[model.SourceTypeFunction], operator.MetadataVersion)
		}
		userList = append(userList, operator.CreateUser, operator.UpdateUser)
	}
	sourceIDToMetadataMap, err := m.MetadataService.BatchGetMetadataBySourceIDs(ctx, sourceMap)
	if err != nil {
		return
	}
	userMap, err := m.UserMgnt.GetUsersName(ctx, utils.UniqueStrings(userList))
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("get user info failed, err: %v", err)
		return
	}
	// var userList []string
	result.Data = []*interfaces.OperatorDataInfo{}
	for _, operator := range operatorList {
		// 查询算子元数据信息
		var metadataDB interfaces.IMetadataDB
		var ok bool
		metadataDB, ok = sourceIDToMetadataMap[operator.MetadataVersion]
		if !ok {
			// TODO: 这里需要处理没有元数据的情况,暂时不处理直接跳过
			m.Logger.WithContext(ctx).Errorf("metadata not found, query: %v", req)
			continue
		}
		// 组装算子信息结果
		var operatorInfo *interfaces.OperatorDataInfo
		// var userIDs []string
		_, operatorInfo, err = m.assembleOperatorResult(ctx, operator, metadataDB)
		if err != nil {
			m.Logger.WithContext(ctx).Warnf("assemble operator result failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		operatorInfo.BusinessDomainID = utils.GetValueOrDefault(resourceToBdMap, operator.OperatorID, req.BusinessDomainID)
		operatorInfo.CreateUser = utils.GetValueOrDefault(userMap, operatorInfo.CreateUser, interfaces.UnknownUser)
		operatorInfo.UpdateUser = utils.GetValueOrDefault(userMap, operatorInfo.UpdateUser, interfaces.UnknownUser)
		result.Data = append(result.Data, operatorInfo)
	}
	return
}

func (m *operatorManager) queryOperatorConfigList(ctx context.Context, req *interfaces.PageQueryRequest) (
	authResp *interfaces.QueryResponse[model.OperatorRegisterDB], resourceToBdMap map[string]string, err error) {
	// 查询算子列表
	conditions := map[string]interface{}{}
	// 将请求参数转换为条件
	conditions["all"] = req.All
	if req.Name != "" {
		conditions["name"] = req.Name
	}
	if req.Category != "" { // 检查分类是否合法
		if !m.CategoryManager.CheckCategory(req.Category) {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCategoryTypeInvalid, "invalid operator category")
			return
		}
		conditions["category"] = req.Category
	}
	if req.CreateUser != "" {
		conditions["create_user"] = req.CreateUser
	}
	if req.OperatorType != "" {
		conditions["operator_type"] = req.OperatorType
	}
	if req.Status != "" {
		conditions["status"] = req.Status
	}
	if req.IsDataSource != nil {
		conditions["is_data_source"] = req.IsDataSource
	}

	// 构建查询执行器
	sortField := sortFieldMap[req.SortBy]
	sort := &ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{Field: sortField, Order: ormhelper.SortOrder(req.SortOrder)},
		},
	}
	queryTotal := func(newCtx context.Context) (int64, error) {
		var count int64
		count, err = m.DBOperatorManager.CountByWhereClause(newCtx, conditions)
		if err != nil {
			m.Logger.WithContext(newCtx).Errorf("count operator list failed, err: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "count operator list failed")
			return 0, err
		}
		return count, nil
	}

	queryBatch := func(newCtx context.Context, pageSize int, offset int, cursorValue *model.OperatorRegisterDB) ([]*model.OperatorRegisterDB, error) {
		var operatorList []*model.OperatorRegisterDB
		var cursor *ormhelper.CursorParams
		if cursorValue != nil {
			cursor = &ormhelper.CursorParams{
				Field:     sortField,
				Direction: ormhelper.SortOrder(req.SortOrder),
			}
			switch sortField {
			case "f_update_time":
				cursor.Value = cursorValue.UpdateTime
			case "f_create_time":
				cursor.Value = cursorValue.CreateTime
			case "f_name":
				cursor.Value = cursorValue.Name
			}
			// 如果使用游标不需要offset
			offset = 0
		}
		conditions["limit"] = pageSize
		conditions["offset"] = offset
		operatorList, err = m.DBOperatorManager.SelectListPage(newCtx, conditions, sort, cursor)
		if err != nil {
			m.Logger.WithContext(newCtx).Errorf("select operator list failed, err: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "select operator list failed")
			return nil, err
		}
		return operatorList, nil
	}

	businessDomainIds := strings.Split(req.BusinessDomainID, ",")
	resourceToBdMap, err = m.BusinessDomainService.BatchResourceList(ctx, businessDomainIds, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}

	queryBuilder := auth.NewQueryBuilder[model.OperatorRegisterDB]().
		SetPage(req.Page, req.PageSize).SetAll(req.All).
		SetQueryFunctions(queryTotal, queryBatch).
		SetFilteredQueryFunctions(
			func(newCtx context.Context, ids []string) (int64, error) {
				conditions["in"] = ids
				return queryTotal(newCtx)
			},
			func(newCtx context.Context, pageSize int, offset int, ids []string, cursorValue *model.OperatorRegisterDB) ([]*model.OperatorRegisterDB, error) {
				conditions["in"] = ids
				return queryBatch(newCtx, pageSize, offset, cursorValue)
			},
		).
		SetBusinessDomainFilter(func(newCtx context.Context) ([]string, error) {
			resourceIDs := make([]string, 0, len(resourceToBdMap))
			for resourceID := range resourceToBdMap {
				resourceIDs = append(resourceIDs, resourceID)
			}
			return resourceIDs, nil
		})
	if common.IsPublicAPIFromCtx(ctx) {
		queryBuilder.SetAuthFilter(func(newCtx context.Context) ([]string, error) {
			// 检查查看权限
			var accessor *interfaces.AuthAccessor
			accessor, err = m.AuthService.GetAccessor(newCtx, req.UserID)
			if err != nil {
				return nil, err
			}
			return m.AuthService.ResourceListIDs(newCtx, accessor, interfaces.AuthResourceTypeOperator, interfaces.AuthOperationTypeView)
		})
	}
	authResp, err = queryBuilder.Execute(ctx)
	return
}

// 组装算子信息结果
func (m *operatorManager) assembleOperatorResult(ctx context.Context, operator *model.OperatorRegisterDB, metadataDB interfaces.IMetadataDB) (
	userIDs []string, result *interfaces.OperatorDataInfo, err error) {
	executeControl := &interfaces.OperatorExecuteControl{}
	err = jsoniter.Unmarshal([]byte(operator.ExecuteControl), executeControl)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("unmarshal execute control failed, err: %v", err)
		return
	}
	var extendInfo map[string]interface{}
	err = jsoniter.Unmarshal([]byte(operator.ExtendInfo), &extendInfo)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("unmarshal extend info failed, err: %v", err)
		return
	}
	result = &interfaces.OperatorDataInfo{
		Name:         operator.Name,
		OperatorID:   operator.OperatorID,
		Version:      operator.MetadataVersion,
		Status:       interfaces.BizStatus(operator.Status),
		MetadataType: interfaces.MetadataType(operator.MetadataType),
		Metadata:     metadata.MetadataDBToStruct(metadataDB),
		ExtendInfo:   extendInfo,
		OperatorInfo: &interfaces.OperatorInfo{
			Type:          interfaces.OperatorType(operator.OperatorType),
			ExecutionMode: interfaces.ExecutionMode(operator.ExecutionMode),
			Category:      interfaces.BizCategory(operator.Category),
			CategoryName:  m.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(operator.Category)),
			Source:        operator.Source,
			IsDataSource:  &operator.IsDataSource,
		},
		OperatorExecuteControl: executeControl,
		CreateTime:             operator.CreateTime,
		UpdateTime:             operator.UpdateTime,
		UpdateUser:             operator.UpdateUser,
		CreateUser:             operator.CreateUser,
		IsInternal:             operator.IsInternal,
	}
	userIDs = []string{operator.CreateUser, operator.UpdateUser, metadataDB.GetCreateUser(), metadataDB.GetUpdateUser()}
	return
}
