package toolbox

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

func (s *ToolServiceImpl) toolBoxDBToToolBoxInfo(ctx context.Context, toolBox *model.ToolboxDB) (boxInfo *interfaces.ToolBoxInfo) {
	boxInfo = &interfaces.ToolBoxInfo{
		BoxID:        toolBox.BoxID,
		BoxName:      toolBox.Name,
		BoxDesc:      toolBox.Description,
		BoxSvcURL:    toolBox.ServerURL,
		Status:       interfaces.BizStatus(toolBox.Status),
		CategoryType: toolBox.Category,
		CategoryName: s.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(toolBox.Category)),
		IsInternal:   toolBox.IsInternal,
		Source:       toolBox.Source,
		CreateTime:   toolBox.CreateTime,
		UpdateTime:   toolBox.UpdateTime,
		CreateUser:   toolBox.CreateUser,
		UpdateUser:   toolBox.UpdateUser,
		ReleaseUser:  toolBox.ReleaseUser,
		ReleaseTime:  toolBox.ReleaseTime,
		MetadataType: interfaces.MetadataType(toolBox.MetadataType),
	}
	return
}

func (s *ToolServiceImpl) toolBoxDBToToolBoxToolInfo(ctx context.Context, toolBox *model.ToolboxDB) (boxInfo *interfaces.ToolBoxToolInfo) {
	boxInfo = &interfaces.ToolBoxToolInfo{
		BoxID:        toolBox.BoxID,
		BoxName:      toolBox.Name,
		BoxDesc:      toolBox.Description,
		Status:       interfaces.BizStatus(toolBox.Status),
		BoxSvcURL:    toolBox.ServerURL,
		CategoryType: toolBox.Category,
		CategoryName: s.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(toolBox.Category)),
		IsInternal:   toolBox.IsInternal,
		Source:       toolBox.Source,
		Tools:        []*interfaces.ToolInfo{},
		CreateUser:   toolBox.CreateUser,
		CreateTime:   toolBox.CreateTime,
		UpdateUser:   toolBox.UpdateUser,
		UpdateTime:   toolBox.UpdateTime,
		ReleaseUser:  toolBox.ReleaseUser,
		ReleaseTime:  toolBox.ReleaseTime,
		MetadataType: interfaces.MetadataType(toolBox.MetadataType),
	}
	return
}

// toolDB 转换成ToolInfo
func (s *ToolServiceImpl) toolDBToToolInfo(ctx context.Context, toolDB *model.ToolDB) (toolInfo *interfaces.ToolInfo, err error) {
	globalParameters := &interfaces.ParametersStruct{}
	if toolDB.Parameters != "" {
		err = utils.StringToObject(toolDB.Parameters, globalParameters)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("parse global parameters failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}
	extendInfo := map[string]interface{}{}
	// 解析扩展信息
	if toolDB.ExtendInfo != "" {
		err = utils.StringToObject(toolDB.ExtendInfo, &extendInfo)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("parse extend info failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}
	var resourceObject interfaces.ResourceObjectType
	switch toolDB.SourceType {
	case model.SourceTypeFunction, model.SourceTypeOpenAPI:
		resourceObject = interfaces.ResourceObjectTool
	case model.SourceTypeOperator:
		resourceObject = interfaces.ResourceObjectOperator
	default:
		resourceObject = interfaces.ResourceObjectTool
	}
	toolInfo = &interfaces.ToolInfo{
		ToolID:           toolDB.ToolID,
		Name:             toolDB.Name,
		Description:      toolDB.Description,
		Status:           interfaces.ToolStatusType(toolDB.Status),
		UseRule:          toolDB.UseRule,
		GlobalParameters: globalParameters,
		ExtendInfo:       extendInfo,
		UpdateTime:       toolDB.UpdateTime,
		CreateTime:       toolDB.CreateTime,
		UpdateUser:       toolDB.UpdateUser,
		CreateUser:       toolDB.CreateUser,
		ResourceObject:   resourceObject,
	}
	return
}

func (s *ToolServiceImpl) getToolBoxList(ctx context.Context, toolBoxDBList []*model.ToolboxDB, resourceToBdMap map[string]string) (toolBoxInfoList []*interfaces.ToolBoxInfo, err error) {
	// 组装工具箱信息结果
	toolBoxInfoList = []*interfaces.ToolBoxInfo{}
	var userIDs, boxIDs []string
	for _, toolBox := range toolBoxDBList {
		toolBoxInfoList = append(toolBoxInfoList, s.toolBoxDBToToolBoxInfo(ctx, toolBox))
		userIDs = append(userIDs, toolBox.CreateUser, toolBox.UpdateUser, toolBox.ReleaseUser)
		boxIDs = append(boxIDs, toolBox.BoxID)
	}
	toolNameMap := make(map[string][]string)
	for i := 0; i < len(boxIDs); i += interfaces.DefaultBatchSize {
		end := i + interfaces.DefaultBatchSize
		if end > len(boxIDs) {
			end = len(boxIDs)
		}
		// 查询工具箱下的工具
		var toolNameList map[string][]string
		toolNameList, err = s.ToolDB.SelectToolNameListByBoxID(ctx, boxIDs[i:end])
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("select toolbox tools failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		for boxID, toolNames := range toolNameList {
			toolNameMap[boxID] = toolNames
		}
	}
	// 获取用户名称
	userMap, err := s.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	businessDomainIDStr, _ := common.GetBusinessDomainFromCtx(ctx)
	for i, toolBox := range toolBoxInfoList {
		toolBoxInfoList[i].BusinessDomainID = utils.GetValueOrDefault(resourceToBdMap, toolBox.BoxID, businessDomainIDStr)
		toolBoxInfoList[i].Tools = toolNameMap[toolBox.BoxID]
		toolBoxInfoList[i].CreateUser = utils.GetValueOrDefault(userMap, toolBox.CreateUser, interfaces.UnknownUser)
		toolBoxInfoList[i].UpdateUser = utils.GetValueOrDefault(userMap, toolBox.UpdateUser, interfaces.UnknownUser)
		toolBoxInfoList[i].ReleaseUser = utils.GetValueOrDefault(userMap, toolBox.ReleaseUser, interfaces.UnknownUser)
	}
	return
}
