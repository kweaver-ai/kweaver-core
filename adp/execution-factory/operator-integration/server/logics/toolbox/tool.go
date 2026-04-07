package toolbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// CreateTool 工具管理
func (s *ToolServiceImpl) CreateTool(ctx context.Context, req *interfaces.CreateToolReq) (resp *interfaces.CreateToolResp, err error) {
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
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolBoxNotFound,
			fmt.Sprintf("toolbox %s not found", req.BoxID))
		return
	}
	// 内置工具箱不允许添加工具
	if toolBox.IsInternal {
		err = errors.DefaultHTTPError(ctx, http.StatusForbidden, "internal toolbox cannot add tools")
		return
	}
	// 解析导入数据
	var metadataList []interfaces.IMetadataDB
	switch req.MetadataType {
	case interfaces.MetadataTypeFunc:
		metadataList, err = s.MetadataService.ParseMetadata(ctx, req.MetadataType, req.FunctionInput)
	case interfaces.MetadataTypeAPI:
		metadataList, err = s.MetadataService.ParseMetadata(ctx, req.MetadataType, req.OpenAPIInput)
	default:
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("metadata type %s not found", req.MetadataType))
	}
	if err != nil {
		return
	}
	// 检查导入工具中是否存在重复的工具
	tools, validatorNameMap, validatorMethodPathMap, err := s.parseOpenAPIToMetadata(ctx, req.BoxID, req.UserID, metadataList, false)
	if err != nil {
		return
	}
	// 去除掉重复的
	failuresVailMap, err := s.checkToolConflict(ctx, req.BoxID, validatorNameMap, validatorMethodPathMap)
	if err != nil {
		return
	}
	resp = &interfaces.CreateToolResp{
		BoxID:      req.BoxID,
		SuccessIDs: []string{},
		Failures:   []interfaces.CreateToolFailureResult{},
	}
	// 组装信息并保存工具
	extendInfo := utils.ObjectToJSON(req.ExtendInfo)
	globalParameters := utils.ObjectToJSON(req.GlobalParameters)
	useRule := req.UseRule
	var detils []metric.AuditLogToolDetil
	for i, tool := range tools {
		// 记录失败信息
		if failuresVailMap[tool.Name] != nil {
			resp.FailureCount++
			resp.Failures = append(resp.Failures, interfaces.CreateToolFailureResult{Error: failuresVailMap[tool.Name], ToolName: tool.Name})
			continue
		}
		// 保存工具
		tool.ExtendInfo = extendInfo
		tool.UseRule = useRule
		tool.Parameters = globalParameters
		toolID, err := s.saveToolToBox(ctx, tool, metadataList[i])
		if err != nil {
			resp.FailureCount++
			resp.Failures = append(resp.Failures, interfaces.CreateToolFailureResult{Error: err, ToolName: tool.Name})
			continue
		}
		// 记录成功信息
		resp.SuccessCount++
		resp.SuccessIDs = append(resp.SuccessIDs, toolID)
		detils = append(detils, metric.AuditLogToolDetil{ToolID: toolID, ToolName: tool.Name})
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[CreateTool] GetAccountAuthContextFromCtx err :%v", err)
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
				OperationCode: metric.AddTool,
			},
		})
	}()
	return resp, nil
}

// 检查新增工具是否和已存在工具冲突
func (s *ToolServiceImpl) checkToolConflict(ctx context.Context, boxID string, validatorNameMap, validatorMethodPathMap map[string]bool) (
	failuresVailMap map[string]error, err error) {
	// 检查工具是否存在
	toolList, err := s.ToolDB.SelectToolByBoxID(ctx, boxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 去除掉重复的
	failuresVailMap = map[string]error{}
	for _, tool := range toolList {
		if validatorNameMap[tool.Name] {
			failuresVailMap[tool.Name] = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolExists,
				fmt.Sprintf("tool name %s exist", tool.Name), tool.Name)
			continue
		}
		if tool.SourceType == model.SourceTypeFunction {
			continue
		}
		// 获取元数据
		var has bool
		var metadata interfaces.IMetadataDB
		has, metadata, err = s.MetadataService.GetMetadataBySource(ctx, tool.SourceID, tool.SourceType)
		if err != nil {
			failuresVailMap[tool.Name] = err
			continue
		}
		if !has {
			continue
		}
		val := validatorMethodPath(metadata.GetMethod(), metadata.GetPath())
		if validatorMethodPathMap[val] {
			failuresVailMap[tool.Name] = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolExists,
				fmt.Sprintf("tool %s exist", val), val)
			continue
		}
	}
	return
}

// saveToolToBox 向工具箱内添加工具
func (s *ToolServiceImpl) saveToolToBox(ctx context.Context, tool *model.ToolDB, metadata interfaces.IMetadataDB) (toolID string, err error) {
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
	var sourceID string
	sourceID, err = s.MetadataService.RegisterMetadata(ctx, tx, metadata)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("insert metadata failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	tool.SourceID = sourceID
	toolID, err = s.ToolDB.InsertTool(ctx, tx, tool)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("insert tool failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}
