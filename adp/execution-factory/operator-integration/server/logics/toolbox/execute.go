package toolbox

import (
	"context"
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

// DebugTool 工具调试
func (s *ToolServiceImpl) DebugTool(ctx context.Context, req *interfaces.ExecuteToolReq) (resp *interfaces.HTTPResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"box_id":  req.BoxID,
		"tool_id": req.ToolID,
		"user_id": req.UserID,
	})

	// 权限校验
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckExecutePermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 检查工具箱是否存在
	exist, toolBox, err := s.ToolBoxDB.SelectToolBox(ctx, req.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox failed	, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtToolBoxNotFound, "toolbox not found")
		return
	}
	// 检查工具是否存在
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
	resp, err = s.executeTool(ctx, req, tool, toolBox.ServerURL)
	if err != nil {
		return
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[ExecuteTool] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationExecute,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectTool,
				Name: toolBox.Name,
				ID:   toolBox.BoxID,
			},
			Detils: &metric.AuditLogToolDetils{
				Infos: []metric.AuditLogToolDetil{
					{
						ToolID:   tool.ToolID,
						ToolName: tool.Name,
					},
				},
				OperationCode: metric.DebugTool,
			},
		})
	}()
	return
}

// ExecuteTool 工具执行
func (s *ToolServiceImpl) ExecuteTool(ctx context.Context, req *interfaces.ExecuteToolReq) (resp *interfaces.HTTPResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"box_id":  req.BoxID,
		"tool_id": req.ToolID,
		"user_id": req.UserID,
	})
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckExecutePermission(ctx, accessor, req.BoxID, interfaces.AuthResourceTypeToolBox)
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
	// 检查工具是否可用
	if tool.Status != string(interfaces.ToolStatusTypeEnabled) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolNotAvailable,
			"tool not available", tool.Name)
		return
	}
	resp, err = s.executeTool(ctx, req, tool, toolBox.ServerURL)
	if err != nil {
		return
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[ExecuteTool] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationExecute,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectTool,
				Name: toolBox.Name,
				ID:   toolBox.BoxID,
			},
			Detils: &metric.AuditLogToolDetils{
				Infos: []metric.AuditLogToolDetil{
					{
						ToolID:   tool.ToolID,
						ToolName: tool.Name,
					},
				},
				OperationCode: metric.ExecuteTool,
			},
		})
	}()
	return resp, nil
}

// ExecuteToolCore 执行工具核心逻辑（不包含权限校验和审计日志）
func (s *ToolServiceImpl) ExecuteToolCore(ctx context.Context, req *interfaces.ExecuteToolReq) (resp *interfaces.HTTPResponse, err error) {
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
	// 检查工具是否可用
	if tool.Status != string(interfaces.ToolStatusTypeEnabled) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolNotAvailable,
			"tool not available", tool.Name)
		return
	}
	resp, err = s.executeTool(ctx, req, tool, toolBox.ServerURL)
	return
}

func (s *ToolServiceImpl) executeTool(ctx context.Context, req *interfaces.ExecuteToolReq, tool *model.ToolDB, toolBoxURL string) (resp *interfaces.HTTPResponse, err error) {
	// 获取元数据
	exist, metadata, err := s.MetadataService.GetMetadataBySource(ctx, tool.SourceID, tool.SourceType)
	if err != nil {
		return
	}
	if !exist {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtMetadataNotFound,
			fmt.Sprintf("metadata type: %s id: %s not found", tool.SourceType, tool.SourceID))
		return
	}
	var url string
	switch tool.SourceType {
	case model.SourceTypeOpenAPI:
		if toolBoxURL == "" {
			toolBoxURL = metadata.GetServerURL()
		}
		url = fmt.Sprintf("%s%s", toolBoxURL, metadata.GetPath())
	case model.SourceTypeOperator:
		url = fmt.Sprintf("%s%s", metadata.GetServerURL(), metadata.GetPath())
	case model.SourceTypeFunction:
		url = fmt.Sprintf("%s%s", metadata.GetServerURL(), metadata.GetPath())
	}
	proxyReq := &interfaces.HTTPRequest{
		ClientID: req.ToolID,
		HTTPRouter: interfaces.HTTPRouter{
			URL:    url,
			Method: metadata.GetMethod(),
		},
		HTTPRequestParams: req.HTTPRequestParams,
		Timeout:           time.Duration(req.Timeout) * time.Second,
	}
	resp, err = s.Proxy.HandlerRequest(ctx, proxyReq)
	return
}
