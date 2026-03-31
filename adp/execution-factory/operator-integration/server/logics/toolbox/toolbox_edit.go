package toolbox

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// UpdateToolBox 更新工具箱
func (s *ToolServiceImpl) UpdateToolBox(ctx context.Context, req *interfaces.UpdateToolBoxReq) (resp *interfaces.UpdateToolBoxResp, err error) {
	// 记录可观测性
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id":  req.UserID,
		"box_id":   req.BoxID,
		"box_name": req.BoxName,
	})

	// 校验编辑权限
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckModifyPermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 检查openapi类型的工具箱是否填写了服务地址
	if req.MetadataType == interfaces.MetadataTypeAPI {
		err = s.Validator.ValidatorURL(ctx, req.BoxSvcURL)
		if err != nil {
			return
		}
	}
	// 检查分类是否存在
	if !s.CategoryManager.CheckCategory(req.Category) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxCategoryTypeInvalid,
			fmt.Sprintf(" %s category not found", req.Category))
		return
	}
	// 检查工具是否存在
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
	// 检查工具箱名称是否存在
	isNameChanged := toolBox.Name != req.BoxName
	if isNameChanged {
		err = s.checkBoxDuplicateName(ctx, req.BoxName, toolBox.BoxID)
		if err != nil {
			return
		}
	}

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
	resp = &interfaces.UpdateToolBoxResp{}
	// 更新元数据
	switch req.MetadataType {
	case interfaces.MetadataTypeAPI:
		var metadatas []interfaces.IMetadataDB
		if req.OpenAPIInput != nil && req.OpenAPIInput.Data != nil {
			metadatas, err = s.MetadataService.ParseMetadata(ctx, req.MetadataType, req.OpenAPIInput.Data)
		}
		if len(metadatas) > 0 {
			resp.EditTools, err = s.batchUpdateOpenAPIToolMetadata(ctx, tx, toolBox.BoxID, req.UserID, metadatas)
			if err != nil {
				return
			}
		}
		toolBox.ServerURL = req.BoxSvcURL
	case interfaces.MetadataTypeFunc:
	}
	// 更新工具箱
	toolBox.Name = req.BoxName
	toolBox.Description = req.BoxDesc
	toolBox.UpdateUser = req.UserID
	toolBox.Category = string(req.Category)
	err = s.ToolBoxDB.UpdateToolBox(ctx, tx, toolBox)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("update toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 如果名称有变化，触发权限资源变更通知
	if isNameChanged {
		authResource := &interfaces.AuthResource{
			ID:   toolBox.BoxID,
			Name: toolBox.Name,
			Type: string(interfaces.AuthResourceTypeToolBox),
		}
		err = s.AuthService.NotifyResourceChange(ctx, authResource)
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[UpdateToolBox] GetAccountAuthContextFromCtx err :%v", err)
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
		})
	}()
	resp.BoxID = req.BoxID
	return
}

// 批量更新OpenAPI类型的工具元数据
func (s *ToolServiceImpl) batchUpdateOpenAPIToolMetadata(ctx context.Context, tx *sql.Tx, boxID, userID string, updateMetadatas []interfaces.IMetadataDB) (resp []*interfaces.EditToolInfo, err error) {
	resp = []*interfaces.EditToolInfo{}
	// 获取当前工具箱内全部工具
	var tools []*model.ToolDB
	tools, err = s.ToolDB.SelectToolByBoxID(ctx, boxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox tools failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 收集需要变更的元数据的工具
	metadataVersions := []string{}
	toolMap := make(map[string]*model.ToolDB)
	for _, tool := range tools {
		if tool.SourceType != model.SourceTypeOpenAPI {
			continue
		}
		metadataVersions = append(metadataVersions, tool.SourceID)
		toolMap[tool.SourceID] = tool
	}
	// 获取所有的元数据
	currentMetadatas, err := s.MetadataService.BatchGetMetadata(ctx, metadataVersions, []string{})
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select metadata failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 构建更新元数据的映射表
	updateMetadataMap := map[string]interfaces.IMetadataDB{}
	for _, metadata := range updateMetadatas {
		updateMetadataMap[validatorMethodPath(metadata.GetMethod(), metadata.GetPath())] = metadata
	}
	// 遍历所有的元数据，检查是否有变更
	var changed bool
	for _, metadata := range currentMetadatas {
		// 检查是否有变更
		_, changed = updateMetadataMap[validatorMethodPath(metadata.GetMethod(), metadata.GetPath())]
		if changed {
			break
		}
	}
	if !changed {
		// 交互设计要求返回指定错误信息：https://confluence.aishu.cn/pages/viewpage.action?pageId=280780968
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtCommonNoMatchedMethodPath,
			"no matched method path found").WithDescription(errors.ErrExtToolNotExistInFile)
		return
	}
	// 更新元数据及工具
	for _, metadata := range currentMetadatas {
		key := validatorMethodPath(metadata.GetMethod(), metadata.GetPath())
		waitUpdateMetadata, ok := updateMetadataMap[key]
		if !ok {
			continue
		}
		// 更新元数据
		metadata.SetSummary(waitUpdateMetadata.GetSummary())
		metadata.SetDescription(waitUpdateMetadata.GetDescription())
		metadata.SetPath(waitUpdateMetadata.GetPath())
		metadata.SetMethod(waitUpdateMetadata.GetMethod())
		metadata.SetServerURL(waitUpdateMetadata.GetServerURL())
		metadata.SetAPISpec(waitUpdateMetadata.GetAPISpec())
		metadata.SetUpdateInfo(userID)
		err = s.MetadataService.UpdateMetadata(ctx, tx, metadata)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("update metadata failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		// 更新工具
		toolDB, ok := toolMap[metadata.GetVersion()]
		if !ok {
			continue
		}
		toolDB.UpdateTime = time.Now().UnixNano()
		toolDB.UpdateUser = userID
		err = s.ToolDB.UpdateTool(ctx, tx, toolDB)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("update tool failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		// 收集变更的工具
		resp = append(resp, &interfaces.EditToolInfo{
			ToolID: toolDB.ToolID,
			Status: interfaces.ToolStatusType(toolDB.Status),
			Name:   toolDB.Name,
			Desc:   toolDB.Description,
		})
	}
	return
}
