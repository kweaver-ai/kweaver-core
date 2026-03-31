package operator

import (
	"context"
	"net/http"
	"strings"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// 排序字段与数据库字段映射
var sortFieldMap = map[string]string{
	"create_time": "f_create_time",
	"update_time": "f_update_time",
	"name":        "f_name",
}

// QueryOperatorMarketDetail 算子市场查询操作
func (m *operatorManager) QueryOperatorMarketDetail(ctx context.Context, req *interfaces.OperatorMarketDetailReq) (
	info *interfaces.OperatorDataInfo, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 检查算子是否存在
	exist, releaseDB, err := m.OpReleaseDB.SelectByOpID(ctx, req.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("query operator failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "query operator failed")
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOperatorNotFound, "operator not found")
		return
	}
	if common.IsPublicAPIFromCtx(ctx) {
		var accessor *interfaces.AuthAccessor
		accessor, err = m.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return
		}
		// 检查是否有公开访问权限
		err = m.AuthService.CheckPublicAccessPermission(ctx, accessor, req.OperatorID, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}
	}
	// 获取元数据信息
	exists, metadataDB, err := m.MetadataService.CheckMetadataExists(ctx, interfaces.MetadataType(releaseDB.MetadataType), releaseDB.MetadataVersion)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("query metadata failed, err: %v", err)
		return
	}
	if !exists {
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtMetadataNotFound, "metadata not found")
		return
	}
	// 组装算子信息结果
	var userIDs []string
	userIDs, info, err = m.assembleReleaseResult(ctx, releaseDB, metadataDB)
	if err != nil {
		return
	}
	// 获取用户信息
	userMap, err := m.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("get users failed, err: %v", err)
		return
	}
	info.CreateUser = utils.GetValueOrDefault(userMap, releaseDB.CreateUser, releaseDB.CreateUser)
	info.UpdateUser = utils.GetValueOrDefault(userMap, releaseDB.UpdateUser, releaseDB.UpdateUser)
	info.ReleaseUser = utils.GetValueOrDefault(userMap, releaseDB.ReleaseUser, releaseDB.ReleaseUser)
	return
}

// QueryOperatorMarketList 算子市场查询操作
func (m *operatorManager) QueryOperatorMarketList(ctx context.Context, req *interfaces.PageQueryOperatorMarketReq) (
	result *interfaces.PageQueryResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	result = &interfaces.PageQueryResponse{
		CommonPageResult: interfaces.CommonPageResult{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Data: []*interfaces.OperatorDataInfo{},
	}
	var releaseList []*model.OperatorReleaseDB
	authResp, resourceToBdMap, err := m.queryOperatorReleaseList(ctx, req)
	if err != nil {
		return
	}
	releaseList = authResp.Data
	result.CommonPageResult = authResp.CommonPageResult
	if len(releaseList) == 0 {
		return
	}
	// 获取元数据信息
	sourceMap := map[model.SourceType][]string{}
	for _, release := range releaseList {
		switch interfaces.MetadataType(release.MetadataType) {
		case interfaces.MetadataTypeAPI:
			sourceMap[model.SourceTypeOpenAPI] = append(sourceMap[model.SourceTypeOpenAPI], release.MetadataVersion)
		case interfaces.MetadataTypeFunc:
			sourceMap[model.SourceTypeFunction] = append(sourceMap[model.SourceTypeFunction], release.MetadataVersion)
		}
	}
	sourceIDToMetadataMap, err := m.MetadataService.BatchGetMetadataBySourceIDs(ctx, sourceMap)
	if err != nil {
		return
	}
	// 组装算子信息结果
	var userList []string
	for _, release := range releaseList {
		var userIDs []string
		var info *interfaces.OperatorDataInfo
		metadataDB, ok := sourceIDToMetadataMap[release.MetadataVersion]
		if !ok {
			m.Logger.WithContext(ctx).Errorf("select operator metadata failed, err: %v", err)
			continue
		}
		userIDs, info, err = m.assembleReleaseResult(ctx, release, metadataDB)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("assemble release result failed, err: %v", err)
			continue
		}
		info.BusinessDomainID = utils.GetValueOrDefault(resourceToBdMap, info.OperatorID, req.BusinessDomainID)
		result.Data = append(result.Data, info)
		userList = append(userList, userIDs...)
	}
	userMap, err := m.UserMgnt.GetUsersName(ctx, userList)
	if err != nil {
		return
	}
	for i := range result.Data {
		result.Data[i].CreateUser = utils.GetValueOrDefault(userMap, result.Data[i].CreateUser, interfaces.UnknownUser)
		result.Data[i].UpdateUser = utils.GetValueOrDefault(userMap, result.Data[i].UpdateUser, interfaces.UnknownUser)
		result.Data[i].ReleaseUser = utils.GetValueOrDefault(userMap, result.Data[i].ReleaseUser, interfaces.UnknownUser)
	}
	return
}

// 根据请求参数查询并过滤算子发布列表
func (m *operatorManager) queryOperatorReleaseList(ctx context.Context, req *interfaces.PageQueryOperatorMarketReq) (
	authResp *interfaces.QueryResponse[model.OperatorReleaseDB], resourceToBdMap map[string]string, err error) {
	filter := make(map[string]interface{})
	filter["all"] = req.All
	if req.Name != "" {
		filter["name"] = req.Name
	}
	if req.Category != "" { // 检查分类是否合法
		if !m.CategoryManager.CheckCategory(req.Category) {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCategoryTypeInvalid, "invalid operator category")
			return
		}
		filter["category"] = req.Category
	}
	if req.CreateUser != "" {
		filter["create_user"] = req.CreateUser
	}
	if req.ReleaseUser != "" {
		filter["release_user"] = req.ReleaseUser
	}
	if req.OperatorType != "" {
		filter["operator_type"] = req.OperatorType
	}
	if req.Status != "" {
		filter["status"] = req.Status
	}
	if req.IsDataSource != nil {
		filter["is_data_source"] = *req.IsDataSource
	}
	if req.ExecutionMode != "" {
		filter["execution_mode"] = req.ExecutionMode
	}
	if req.MetadataType != "" {
		filter["metadata_type"] = req.MetadataType
	}
	// 构造排序字段
	sort := &ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{
				Field: sortFieldMap[req.SortBy],
				Order: ormhelper.SortOrder(req.SortOrder),
			},
		},
	}
	// 构建查询执行器
	queryTotal := func(newCtx context.Context) (int64, error) {
		var count int64
		count, err = m.OpReleaseDB.CountByWhereClause(newCtx, filter)
		if err != nil {
			m.Logger.WithContext(newCtx).Errorf("count operator release list error: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "count operator release list error")
			return 0, err
		}
		return count, nil
	}
	queryBatch := func(newCtx context.Context, pageSize int, offset int, cursorValue *model.OperatorReleaseDB) ([]*model.OperatorReleaseDB, error) {
		var list []*model.OperatorReleaseDB
		var cursor *ormhelper.CursorParams
		if cursorValue != nil {
			cursor = &ormhelper.CursorParams{
				Field:     sortFieldMap[req.SortBy],
				Direction: ormhelper.SortOrder(req.SortOrder),
			}
			switch req.SortBy {
			case "update_time":
				cursor.Value = cursorValue.UpdateTime
			case "create_time":
				cursor.Value = cursorValue.CreateTime
			case "name":
				cursor.Value = cursorValue.Name
			}
			// 如果使用游标不需要offset
			offset = 0
		}
		filter["limit"] = pageSize
		filter["offset"] = offset
		list, err = m.OpReleaseDB.SelectByWhereClause(newCtx, filter, sort, cursor)
		if err != nil {
			m.Logger.WithContext(newCtx).Errorf("select operator release list error: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "select operator release list error")
			return nil, err
		}
		return list, nil
	}

	businessDomainIds := strings.Split(req.BusinessDomainID, ",")
	resourceToBdMap, err = m.BusinessDomainService.BatchResourceList(ctx, businessDomainIds, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}

	queryBuilder := auth.NewQueryBuilder[model.OperatorReleaseDB]().
		SetPage(req.Page, req.PageSize).SetAll(req.All).
		SetQueryFunctions(queryTotal, queryBatch).
		SetFilteredQueryFunctions(func(newCtx context.Context, ids []string) (int64, error) {
			filter["in"] = ids
			return queryTotal(newCtx)
		}, func(newCtx context.Context, pageSize int, offset int, ids []string, cursorValue *model.OperatorReleaseDB) ([]*model.OperatorReleaseDB, error) {
			filter["in"] = ids
			return queryBatch(newCtx, pageSize, offset, cursorValue)
		}).
		SetBusinessDomainFilter(func(newCtx context.Context) ([]string, error) {
			resourceIDs := make([]string, 0, len(resourceToBdMap))
			for resourceID := range resourceToBdMap {
				resourceIDs = append(resourceIDs, resourceID)
			}
			return resourceIDs, nil
		})
	if common.IsPublicAPIFromCtx(ctx) {
		// 设置公共访问权限过滤
		queryBuilder.SetAuthFilter(func(newCtx context.Context) ([]string, error) {
			var accessor *interfaces.AuthAccessor
			accessor, err = m.AuthService.GetAccessor(newCtx, req.UserID)
			if err != nil {
				return nil, err
			}
			return m.AuthService.ResourceListIDs(newCtx, accessor, interfaces.AuthResourceTypeOperator, interfaces.AuthOperationTypePublicAccess)
		})
	}
	authResp, err = queryBuilder.Execute(ctx)
	return
}
