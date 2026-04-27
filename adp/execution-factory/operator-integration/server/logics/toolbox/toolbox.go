// Package toolbox 工具箱、工具管理
// @file internal_tool.go
// @description: 管理实现
package toolbox

import (
	"context"
	"fmt"
	"net/http"

	infracommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metadata"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// GetToolBox 获取工具箱信息
func (s *ToolServiceImpl) GetToolBox(ctx context.Context, req *interfaces.GetToolBoxReq, isMarket bool) (resp *interfaces.ToolBoxToolInfo, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 如果是公开接口，检查查看权限
	if infracommon.IsPublicAPIFromCtx(ctx) {
		var accessor *interfaces.AuthAccessor
		accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return
		}
		if isMarket {
			err = s.AuthService.CheckPublicAccessPermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
		} else {
			err = s.AuthService.CheckViewPermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
		}
		if err != nil {
			return
		}
	}

	// 检查工具箱是否存在
	exist, toolBox, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtToolBoxNotFound,
			fmt.Sprintf("toolbox %s not found", req.BoxID))
		return
	}
	// 如果时市场接口，只能获取已发布工具详情
	if isMarket && toolBox.Status != interfaces.BizStatusPublished.String() {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound,
			fmt.Sprintf("toolbox %s is not published", req.BoxID))
		return
	}

	// 转换工具箱数据库模型到工具箱信息
	resp = s.toolBoxDBToToolBoxToolInfo(ctx, toolBox)
	userIDs := []string{toolBox.CreateUser, toolBox.UpdateUser, toolBox.ReleaseUser}

	// 获取工具箱下的工具
	tools, err := s.ToolDB.SelectToolByBoxID(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	toolInfos, userMap, err := s.batchGetToolInfoAndUserInfo(ctx, tools, userIDs, toolBox.ServerURL, interfaces.MetadataType(toolBox.MetadataType))
	if err != nil {
		return
	}
	resp.Tools = append(resp.Tools, toolInfos...)
	resp.CreateUser = utils.GetValueOrDefault(userMap, toolBox.CreateUser, interfaces.UnknownUser)
	resp.UpdateUser = utils.GetValueOrDefault(userMap, toolBox.UpdateUser, interfaces.UnknownUser)
	resp.ReleaseUser = utils.GetValueOrDefault(userMap, toolBox.ReleaseUser, interfaces.UnknownUser)
	return
}

// DeleteBoxByID 删除工具箱
func (s *ToolServiceImpl) DeleteBoxByID(ctx context.Context, req *interfaces.DeleteBoxReq) (resp *interfaces.DeleteBoxResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"box_id":  req.BoxID,
		"user_id": req.UserID,
	})
	// 校验删除权限
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckDeletePermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 检查工具箱是否存在
	exist, toolBox, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound, "toolbox not found")
		return
	}
	// 删除工具箱
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	err = s.deleteToolBox(ctx, tx, req.BoxID)
	if err != nil {
		return
	}

	// 取消关联业务域
	err = s.BusinessDomainService.DisassociateResource(ctx, req.BusinessDomainID, req.BoxID, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 删除资源权限策略
	err = s.AuthService.DeletePolicy(ctx, []string{req.BoxID}, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}

	// 记录审计日志
	go func() {
		accountAuthContext, ok := infracommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[DeleteToolBox] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationDelete,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectTool,
				Name: toolBox.Name,
				ID:   toolBox.BoxID,
			},
		})
	}()
	return
}

// QueryToolBoxList 工具箱管理
func (s *ToolServiceImpl) QueryToolBoxList(ctx context.Context, req *interfaces.QueryToolBoxListReq) (resp *interfaces.QueryToolBoxListResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 构造查询条件
	filter := make(map[string]interface{})
	filter["all"] = req.All
	if req.BoxName != "" {
		filter["name"] = req.BoxName
	}
	if req.BoxCategory != "" {
		// 检查分类是否合法
		if !s.CategoryManager.CheckCategory(req.BoxCategory) {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxCategoryTypeInvalid,
				fmt.Sprintf(" %s category not found", req.BoxCategory))
			return
		}
		filter["category"] = req.BoxCategory
	}
	if req.CreateUser != "" {
		filter["create_user"] = req.CreateUser
	}
	if req.ReleaseUser != "" {
		filter["release_user"] = req.ReleaseUser
	}
	if req.Status != "" {
		filter["status"] = req.Status
	}
	operations := interfaces.AuthOperationTypeView
	resp = &interfaces.QueryToolBoxListResp{
		Data: []*interfaces.ToolBoxInfo{},
	}
	authResp, resourceToBdMap, err := s.getToolBoxListPage(ctx, filter, req.CommonPageParams, req.UserID, operations)
	if err != nil {
		return
	}
	resp.CommonPageResult = authResp.CommonPageResult
	toolBoxList := authResp.Data
	if len(toolBoxList) == 0 {
		return
	}
	// 组装工具箱信息结果
	toolBoxInfoList, err := s.getToolBoxList(ctx, toolBoxList, resourceToBdMap)
	if err != nil {
		return
	}
	resp.Data = toolBoxInfoList
	return
}

// UpdateToolBoxStatus 修改工具箱状态
func (s *ToolServiceImpl) UpdateToolBoxStatus(ctx context.Context, req *interfaces.UpdateToolBoxStatusReq) (resp *interfaces.UpdateToolBoxStatusResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"box_id":  req.BoxID,
		"user_id": req.UserID,
	})
	// 检查工具箱是否存在
	exist, toolBox, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound,
			fmt.Sprintf("toolbox %s not found", req.BoxID))
		return
	}
	// 检查请求转换参数是否合法
	if !common.CheckStatusTransition(interfaces.BizStatus(toolBox.Status), req.Status) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxStatusInvalid,
			fmt.Sprintf("toolbox %s status can not be transition to %s", req.BoxID, req.Status))
		return
	}
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	var operation metric.AuditLogOperationType
	switch req.Status {
	case interfaces.BizStatusPublished:
		operation = metric.AuditLogOperationPublish
		// 校验发布权限
		err = s.AuthService.CheckPublishPermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
		if err != nil {
			return
		}
		// 检查是否重名
		err = s.checkBoxDuplicateName(ctx, toolBox.Name, toolBox.BoxID)
	case interfaces.BizStatusUnpublish, interfaces.BizStatusEditing:
	case interfaces.BizStatusOffline:
		operation = metric.AuditLogOperationUnpublish
		// 校验下架权限、校验编辑权限
		err = s.AuthService.CheckUnpublishPermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
	default:
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxStatusInvalid,
			fmt.Sprintf("invalid toolbox status: %s", req.Status))
	}
	if err != nil {
		return
	}
	err = s.ToolBoxDB.UpdateToolBoxStatus(ctx, nil, req.BoxID, string(req.Status), req.UserID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("update toolbox status failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "update toolbox status failed")
		return
	}
	// 记录审计日志
	if operation != "" {
		go func() {
			accountAuthContext, ok := infracommon.GetAccountAuthContextFromCtx(ctx)
			if !ok {
				s.Logger.WithContext(ctx).Warnf("[UpdateToolBoxStatus] GetAccountAuthContextFromCtx err :%v", err)
				return
			}
			s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
				TokenInfo: accountAuthContext.TokenInfo,
				Accessor:  accessor,
				Operation: operation,
				Object: &metric.AuditLogObject{
					Type: metric.AuditLogObjectTool,
					Name: toolBox.Name,
					ID:   toolBox.BoxID,
				},
			})
		}()
	}
	resp = &interfaces.UpdateToolBoxStatusResp{
		BoxID:  req.BoxID,
		Status: req.Status,
	}
	return
}

// GetBoxTool 获取工具信息
func (s *ToolServiceImpl) GetBoxTool(ctx context.Context, req *interfaces.GetToolReq) (resp *interfaces.ToolInfo, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"box_id":  req.BoxID,
		"tool_id": req.ToolID,
	})
	// 如果是外部接口，校验是否拥有所属工具的查看、公开访问权限
	if infracommon.IsPublicAPIFromCtx(ctx) {
		var accessor *interfaces.AuthAccessor
		accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return
		}
		var authorized bool
		authorized, err = s.AuthService.OperationCheckAny(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox, interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess)
		if err != nil {
			return
		}
		if !authorized {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden, nil)
			return
		}
	}
	// 检查工具箱是否存在
	exist, boxDB, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound, fmt.Sprintf("toolbox %s not found", req.BoxID))
		return
	}
	exist, tool, err := s.ToolDB.SelectTool(ctx, req.ToolID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolNotFound,
			fmt.Sprintf("tool %s not found", req.ToolID))
		return
	}
	resp, err = s.getToolInfo(ctx, tool, boxDB.ServerURL, interfaces.MetadataType(boxDB.MetadataType))
	return
}

// DeleteBoxTool 批量删除工具箱内工具
func (s *ToolServiceImpl) DeleteBoxTool(ctx context.Context, req *interfaces.BatchDeleteToolReq) (resp *interfaces.BatchDeleteToolResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 权限校验
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckModifyPermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 检查工具箱是否存在
	exist, toolBox, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound, "toolbox not found")
		return
	}
	// 内置工具不允许删除工具
	if toolBox.IsInternal {
		err = errors.DefaultHTTPError(ctx, http.StatusForbidden, "internal toolbox cannot delete tools")
		return
	}
	// 检查工具是否存在
	tools, err := s.ToolDB.SelectToolBoxByID(ctx, req.BoxID, req.ToolIDs)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if len(tools) != len(req.ToolIDs) {
		checkTools := []string{}
		for _, v := range tools {
			checkTools = append(checkTools, v.ToolID)
		}
		clist := utils.FindMissingElements(req.ToolIDs, checkTools)
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolNotFound,
			fmt.Sprintf("tools %v not found", clist))
		return
	}
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("get tx failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	err = s.deleteTools(ctx, tx, req.BoxID, tools)
	if err != nil {
		return
	}
	// 记录审计日志
	go func() {
		var detils []metric.AuditLogToolDetil
		for _, tool := range tools {
			detils = append(detils, metric.AuditLogToolDetil{
				ToolID:   tool.ToolID,
				ToolName: tool.Name,
			})
		}
		accountAuthContext, ok := infracommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[DeleteBoxTool] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationEdit,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectTool,
				Name: toolBox.Name,
				ID:   toolBox.BoxID,
			},
			Detils: &metric.AuditLogToolDetils{
				Infos:         detils,
				OperationCode: metric.DeleteTool,
			},
		})
	}()
	return
}

// QueryToolList 查询工具列表(获取工具箱内工具列表)
func (s *ToolServiceImpl) QueryToolList(ctx context.Context, req *interfaces.QueryToolListReq) (resp *interfaces.QueryToolListResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 如果外部接口，校验是否拥有所属工具箱的查看、公开访问权限
	if infracommon.IsPublicAPIFromCtx(ctx) {
		var accessor *interfaces.AuthAccessor
		accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return
		}
		var authorized bool
		authorized, err = s.AuthService.OperationCheckAny(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox, interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess)
		if err != nil {
			return
		}
		if !authorized {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden, nil)
			return
		}
	}
	// 检查工具箱是否存在
	exist, boxDB, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound, "toolbox not found")
		return
	}
	// 构造查询条件
	filter := make(map[string]interface{})
	filter["all"] = req.All
	if req.ToolName != "" {
		filter["name"] = req.ToolName
	}
	if req.Status != "" {
		filter["status"] = req.Status
	}
	if req.QueryUserID != "" {
		filter["user_id"] = req.QueryUserID
	}
	// 查询工具箱总数
	total, err := s.ToolDB.CountToolByBoxID(ctx, req.BoxID, filter)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("count tool failed by id: %s, err: %v", req.BoxID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	resp = &interfaces.QueryToolListResp{
		BoxID: req.BoxID,
		CommonPageResult: interfaces.CommonPageResult{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: int(total),
		},
		Tools: []*interfaces.ToolInfo{},
	}
	if total == 0 {
		return
	}
	// 计算偏移量
	var offset int
	if req.PageSize > 0 {
		offset = (req.Page - 1) * req.PageSize
		resp.TotalPage = int(total) / req.PageSize
		if int(total)%req.PageSize > 0 {
			resp.TotalPage++
		}
		resp.HasNext = req.Page < resp.TotalPage
		resp.HasPrev = req.Page > 1
	} else {
		resp.TotalPage = 1
		resp.PageSize = int(total)
	}
	// 构造排序条件
	filter["sort_by"] = req.SortBy
	filter["sort_order"] = req.SortOrder
	filter["limit"] = req.PageSize
	filter["offset"] = offset
	// 查询工具箱列表
	tools, err := s.ToolDB.SelectToolLisByBoxID(ctx, req.BoxID, filter)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool list failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 收集工具相关信息
	userIDs := []string{}
	toolInfos, _, err := s.batchGetToolInfoAndUserInfo(ctx, tools, userIDs, boxDB.ServerURL, interfaces.MetadataType(boxDB.MetadataType))
	if err != nil {
		return
	}
	resp.Tools = append(resp.Tools, toolInfos...)
	return
}

// UpdateToolStatus 更新工具状态
func (s *ToolServiceImpl) UpdateToolStatus(ctx context.Context, req *interfaces.UpdateToolStatusReq) (resp []*interfaces.ToolStatus, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"box_id":  req.BoxID,
		"user_id": req.UserID,
	})
	// 权限校验
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckModifyPermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 检查工具箱是否存在
	exist, toolBox, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound, "toolbox not found")
		return
	}
	// 检查工具是否存在
	var toolIDs []string
	for _, v := range req.ToolStatusList {
		toolIDs = append(toolIDs, v.ToolID)
	}
	tools, err := s.ToolDB.SelectToolBoxByID(ctx, req.BoxID, toolIDs)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	checkTools := []string{}
	sourceMap := map[model.SourceType][]string{}
	sourceMap[model.SourceTypeOperator] = []string{}
	for _, v := range tools {
		checkTools = append(checkTools, v.ToolID)
		if v.SourceType == model.SourceTypeOperator {
			sourceMap[v.SourceType] = append(sourceMap[v.SourceType], v.SourceID)
		}
	}
	//  比较工具ID是否存在
	clist := utils.FindMissingElements(toolIDs, checkTools)
	if len(clist) > 0 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolNotFound,
			fmt.Sprintf("tools %v not found", clist))
		return
	}
	if len(sourceMap[model.SourceTypeOperator]) > 0 {
		// 检查依赖资源是否存在
		var sourceIDToMetadataMap map[string]interfaces.IMetadataDB
		sourceIDToMetadataMap, err = s.MetadataService.BatchGetMetadataBySourceIDs(ctx, sourceMap)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("batch get metadata failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		for _, v := range tools {
			if v.SourceType == model.SourceTypeOperator {
				if _, ok := sourceIDToMetadataMap[v.SourceID]; !ok {
					err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolRefOperatorNotFound,
						fmt.Sprintf("tool %s ref operator %s not found", v.ToolID, v.SourceID), v.Name)
					return
				}
			}
		}
	}

	// 更新工具状态
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	resp = []*interfaces.ToolStatus{}
	for _, tool := range req.ToolStatusList {
		err = s.ToolDB.UpdateToolStatus(ctx, tx, tool.ToolID, string(tool.Status), req.UserID)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("update tool status failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		resp = append(resp, &interfaces.ToolStatus{
			ToolID: tool.ToolID,
			Status: tool.Status,
		})
	}
	// 记录审计日志
	go func() {
		var detils []metric.AuditLogToolDetil
		for _, tool := range tools {
			detils = append(detils, metric.AuditLogToolDetil{
				ToolID:   tool.ToolID,
				ToolName: tool.Name,
			})
		}
		accountAuthContext, ok := infracommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[UpdateToolStatus] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationEdit,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectTool,
				Name: toolBox.Name,
				ID:   toolBox.BoxID,
			},
			Detils: &metric.AuditLogToolDetils{
				Infos:         detils,
				OperationCode: metric.UpdateToolStatus,
			},
		})
	}()
	return resp, nil
}

// getToolInfo 获取工具信息
func (s *ToolServiceImpl) getToolInfo(ctx context.Context, tool *model.ToolDB, boxSvcURL string, boxMetadataType interfaces.MetadataType) (toolInfo *interfaces.ToolInfo, err error) {
	toolInfo, err = s.toolDBToToolInfo(ctx, tool)
	if err != nil {
		return
	}
	// 获取元数据p
	has, metadataDB, err := s.MetadataService.GetMetadataBySource(ctx, tool.SourceID, tool.SourceType)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("get metadata failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !has {
		s.Logger.WithContext(ctx).Errorf("metadata type: %s source_id: %s not found", tool.SourceType, tool.SourceID)
		toolInfo.MetadataType = boxMetadataType
		toolInfo.Metadata = metadata.DefaultMetadataInfo(boxMetadataType)
		return
	}
	// 若为OpenAPI类型，ServerURL和工具箱配置的boxSvcURL保持一致
	metadataDB.SetServerURL(boxSvcURL)
	// 转换为结构体
	toolInfo.MetadataType = interfaces.MetadataType(metadataDB.GetType())
	toolInfo.Metadata = metadata.MetadataDBToStruct(metadataDB)
	return
}

// batchGetToolInfoAndUserInfo 批量获取工具及用户信息
func (s *ToolServiceImpl) batchGetToolInfoAndUserInfo(ctx context.Context, tools []*model.ToolDB, userIDs []string,
	boxSvcURL string, boxMetadataType interfaces.MetadataType) (toolInfos []*interfaces.ToolInfo, userMap map[string]string, err error) {
	toolInfos = []*interfaces.ToolInfo{}
	sourceMap := map[model.SourceType][]string{}
	toolIDSourceMap := map[string]string{}
	// 组装工具信息
	for _, toolDB := range tools {
		toolIDSourceMap[toolDB.ToolID] = toolDB.SourceID
		sourceMap[toolDB.SourceType] = append(sourceMap[toolDB.SourceType], toolDB.SourceID)
		userIDs = append(userIDs, toolDB.CreateUser, toolDB.UpdateUser)
		var toolInfo *interfaces.ToolInfo
		toolInfo, err = s.toolDBToToolInfo(ctx, toolDB)
		if err != nil {
			return
		}
		toolInfos = append(toolInfos, toolInfo)
	}
	// 获取用户名称
	userMap, err = s.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	// 批量获取工具元数据
	sourceIDToMetadataMap, err := s.MetadataService.BatchGetMetadataBySourceIDs(ctx, sourceMap)
	if err != nil {
		return
	}
	// 填充元数据信息
	for _, toolInfo := range toolInfos {
		toolInfo.CreateUser = utils.GetValueOrDefault(userMap, toolInfo.CreateUser, interfaces.UnknownUser)
		toolInfo.UpdateUser = utils.GetValueOrDefault(userMap, toolInfo.UpdateUser, interfaces.UnknownUser)
		metadataDB, ok := sourceIDToMetadataMap[toolIDSourceMap[toolInfo.ToolID]]
		if !ok {
			s.Logger.WithContext(ctx).Errorf("metadata not found, toolID: %s", toolInfo.ToolID)
			toolInfo.MetadataType = boxMetadataType
			toolInfo.Metadata = metadata.DefaultMetadataInfo(boxMetadataType)
			continue
		}
		metadataDB.SetServerURL(boxSvcURL)
		toolInfo.MetadataType = interfaces.MetadataType(metadataDB.GetType())
		toolInfo.Metadata = metadata.MetadataDBToStruct(metadataDB)
	}
	return
}

// checkBoxDuplicateName 检查工具箱名称是否重复
func (s *ToolServiceImpl) checkBoxDuplicateName(ctx context.Context, name, boxID string) (err error) {
	has, boxDB, err := s.ToolBoxDB.SelectToolBoxByName(ctx, name, []string{string(interfaces.BizStatusPublished)})
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox by name failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select toolbox by name failed")
		return
	}
	if !has || (boxID != "" && boxDB.BoxID == boxID) {
		return
	}
	err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNameExists,
		fmt.Sprintf("toolbox name %s already exists", name), name)
	return
}
