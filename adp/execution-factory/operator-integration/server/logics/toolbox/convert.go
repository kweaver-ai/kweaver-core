// Package toolbox 工具箱、工具管理
// @file convert.go
// @description: 转换算子为工具
package toolbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// ConvertOperatorToTool 算子转换成工具
func (s *ToolServiceImpl) ConvertOperatorToTool(ctx context.Context, req *interfaces.ConvertOperatorToToolReq) (resp *interfaces.ConvertOperatorToToolResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 校验是否拥有所属工具箱的编辑权限
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
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound,
			fmt.Sprintf("toolbox %s not found", req.BoxID))
		return
	}
	// TODO : 内置工具不允许添加工具
	if toolBox.IsInternal {
		err = errors.DefaultHTTPError(ctx, http.StatusForbidden, "internal toolbox cannot add tools")
		return
	}
	operatorCheckInfo, err := s.OperatorMgnt.CheckAddAsTool(ctx, req.OperatorID, req.UserID)
	if err != nil {
		return
	}
	// 检查算子元数据类型和工具是否一致
	if toolBox.MetadataType != operatorCheckInfo.Metadata.GetType() {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolConvertMetadataTypeNotMatch,
			fmt.Sprintf("operator %s metadata type %s not match toolbox metadata type %s", operatorCheckInfo.OperatorID, operatorCheckInfo.Metadata.GetType(), toolBox.MetadataType))
		return
	}

	resp = &interfaces.ConvertOperatorToToolResp{
		BoxID: req.BoxID,
	}
	switch interfaces.MetadataType(operatorCheckInfo.Metadata.GetType()) {
	case interfaces.MetadataTypeAPI, interfaces.MetadataTypeFunc:
		metadataDB := operatorCheckInfo.Metadata
		err = s.checkBoxToolSame(ctx, req.BoxID, operatorCheckInfo.Name, metadataDB.GetMethod(), metadataDB.GetPath())
		if err != nil {
			return
		}
		// 转换算子为工具
		tool := &model.ToolDB{
			BoxID:       req.BoxID,
			Name:        operatorCheckInfo.Name,
			Description: metadataDB.GetDescription(),
			SourceID:    operatorCheckInfo.OperatorID,
			SourceType:  model.SourceTypeOperator,
			Status:      string(interfaces.ToolStatusTypeDisabled),
			UseRule:     req.UseRule,
			ExtendInfo:  utils.ObjectToJSON(req.ExtendInfo),
			Parameters:  utils.ObjectToJSON(req.GlobalParameters),
			CreateUser:  req.UserID,
			UpdateUser:  req.UserID,
		}
		// 插入工具
		resp.ToolID, err = s.ToolDB.InsertTool(ctx, nil, tool)
		if err != nil {
			s.Logger.WithContext(ctx).Warnf("insert tool failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolConvertOnlySupportAPI,
			"only api operators can be published as tools")
		return
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[ConvertOperatorToTool] GetAccountAuthContextFromCtx err :%v", err)
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
				Infos: []metric.AuditLogToolDetil{
					{
						ToolID:   resp.ToolID,
						ToolName: operatorCheckInfo.Name,
					},
				},
				OperationCode: metric.ImportToolFromOperator,
			},
		})
	}()
	return
}

// checkBoxToolSame 检查工具箱内是否存在同名工具
func (s *ToolServiceImpl) checkBoxToolSame(ctx context.Context, boxID, name, method, path string) (err error) {
	// 检查工具是否存在
	toolList, err := s.ToolDB.SelectToolByBoxID(ctx, boxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	for _, tool := range toolList {
		if tool.Name == name {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolExists,
				fmt.Sprintf("tool name %s exist", tool.Name), tool.Name)
			return
		}
		var toolInfo *interfaces.ToolInfo
		toolInfo, err = s.getToolInfo(ctx, tool, "", "")
		if err != nil {
			return
		}
		if toolInfo.Metadata == nil {
			s.Logger.WithContext(ctx).Warnf("toolbox %s tool %s:%s metadata is nil", boxID, tool.Name, tool.ToolID)
			continue
		}
		val := validatorMethodPath(toolInfo.Metadata.Method, toolInfo.Metadata.Path)
		if val == validatorMethodPath(method, path) {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolExists,
				fmt.Sprintf("tool %s exist", val), val)
			return
		}
	}
	return
}
