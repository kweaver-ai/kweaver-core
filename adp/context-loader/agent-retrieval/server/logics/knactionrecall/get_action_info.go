// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knactionrecall

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// GetActionInfo 获取行动信息（行动召回）
func (s *knActionRecallServiceImpl) GetActionInfo(ctx context.Context, req *interfaces.KnActionRecallRequest) (*interfaces.KnActionRecallResponse, error) {
	// 1. 参数合并：_instance_identities 优先，回退到 _instance_identity 包装为数组
	instanceIdentities := make([]map[string]any, 0)
	if len(req.InstanceIdentities) > 0 {
		for _, id := range req.InstanceIdentities {
			if len(id) > 0 {
				instanceIdentities = append(instanceIdentities, id)
			}
		}
	} else if len(req.InstanceIdentity) > 0 {
		instanceIdentities = append(instanceIdentities, req.InstanceIdentity)
	}

	// 2. 调用行动查询接口
	actionsReq := &interfaces.QueryActionsRequest{
		KnID:               req.KnID,
		AtID:               req.AtID,
		InstanceIdentities: instanceIdentities,
		IncludeTypeInfo:    false, // 不需要类型信息
	}

	actionsResp, err := s.ontologyQuery.QueryActions(ctx, actionsReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] QueryActions failed, err: %v", err)
		return nil, err
	}

	// 3. 检查返回结果
	if actionsResp.ActionSource == nil {
		s.logger.WithContext(ctx).Warnf("[KnActionRecall#GetActionInfo] ActionSource is nil")
		return &interfaces.KnActionRecallResponse{
			DynamicTools: []interfaces.KnDynamicTool{},
		}, nil
	}

	if len(actionsResp.Actions) == 0 {
		s.logger.WithContext(ctx).Warnf("[KnActionRecall#GetActionInfo] Actions is empty")
		return &interfaces.KnActionRecallResponse{
			DynamicTools: []interfaces.KnDynamicTool{},
		}, nil
	}

	// 4. 检查 action_source.type
	if actionsResp.ActionSource.Type != interfaces.ActionSourceTypeTool && actionsResp.ActionSource.Type != interfaces.ActionSourceTypeMCP {
		s.logger.WithContext(ctx).Warnf("[KnActionRecall#GetActionInfo] Unsupported action_source type: %s", actionsResp.ActionSource.Type)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadRequest,
			fmt.Sprintf("当前仅支持 type=%s 或 %s 的行动源。当前类型: %s",
				interfaces.ActionSourceTypeTool, interfaces.ActionSourceTypeMCP, actionsResp.ActionSource.Type))
	}

	// 5. 仅处理 actions[0]
	firstAction := actionsResp.Actions[0]

	// 6. 统一构造行动驱动 API URL
	apiURL := s.buildActionDriverAPIURL(req.KnID, req.AtID)

	// 7. 统一构造行动驱动 fixed_params
	fixedParams := interfaces.ActionDriverFixedParams{
		DynamicParams:      firstAction.Parameters,
		InstanceIdentities: instanceIdentities,
	}

	var dynamicTool interfaces.KnDynamicTool

	if actionsResp.ActionSource.Type == interfaces.ActionSourceTypeTool {
		// 8a. Tool 类型：获取工具详情
		toolDetailReq := &interfaces.GetToolDetailRequest{
			BoxID:  actionsResp.ActionSource.BoxID,
			ToolID: actionsResp.ActionSource.ToolID,
		}

		toolDetail, err := s.operatorIntegration.GetToolDetail(ctx, toolDetailReq)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] GetToolDetail failed, err: %v", err)
			return nil, err
		}

		// 9a. 将 Tool Schema 转换为行动驱动参数结构
		parameters, err := s.convertToolSchemaToActionDriver(ctx, toolDetail.Metadata.APISpec)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] ConvertToolSchemaToActionDriver failed, err: %v", err)
			return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError,
				fmt.Sprintf("Tool Schema 转换为行动驱动结构失败: %v", err))
		}

		// 10a. 构建 KnDynamicTool
		dynamicTool = interfaces.KnDynamicTool{
			Name:            toolDetail.Name,
			Description:     toolDetail.Description,
			Parameters:      parameters,
			APIURL:          apiURL,
			FixedParams:     fixedParams,
			APICallStrategy: interfaces.ResultProcessStrategyKnActionRecall,
		}
	} else {
		// 8b. MCP 类型：获取 MCP 工具详情
		mcpReq := &interfaces.GetMCPToolDetailRequest{
			McpID:    actionsResp.ActionSource.McpID,
			ToolName: actionsResp.ActionSource.ToolName,
		}

		toolDetail, err := s.operatorIntegration.GetMCPToolDetail(ctx, mcpReq)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] GetMCPToolDetail failed, err: %v", err)
			return nil, err
		}

		// 9b. 将 MCP Schema 转换为行动驱动参数结构
		parameters, err := s.convertMCPSchemaToActionDriver(ctx, toolDetail.InputSchema)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] ConvertMCPSchemaToActionDriver failed, err: %v", err)
			return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError,
				fmt.Sprintf("MCP Schema 转换为行动驱动结构失败: %v", err))
		}

		// 10b. 构建 KnDynamicTool
		dynamicTool = interfaces.KnDynamicTool{
			Name:            toolDetail.Name,
			Description:     toolDetail.Description,
			Parameters:      parameters,
			APIURL:          apiURL,
			FixedParams:     fixedParams,
			APICallStrategy: interfaces.ResultProcessStrategyKnActionRecall,
		}
	}

	// 11. 构建headers
	headers := common.GetHeaderFromCtx(ctx)

	return &interfaces.KnActionRecallResponse{
		Headers:      headers,
		DynamicTools: []interfaces.KnDynamicTool{dynamicTool},
	}, nil
}

// buildActionDriverAPIURL 统一生成行动驱动内部执行接口地址
// Tool 和 MCP 类型均调用此方法生成相同格式的 api_url
func (s *knActionRecallServiceImpl) buildActionDriverAPIURL(knID, atID string) string {
	servicePath := fmt.Sprintf("/api/ontology-query/in/v1/knowledge-networks/%s/action-types/%s/execute", knID, atID)
	return s.config.OntologyQuery.BuildURL(servicePath)
}
