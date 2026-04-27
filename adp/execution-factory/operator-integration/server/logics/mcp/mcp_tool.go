package mcp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
)

func (s *mcpServiceImpl) getMCPToolConfig(ctx context.Context, mcpID string, mcpVersion int) (toolConfigs []*interfaces.MCPToolConfigInfo, err error) {
	tools, err := s.DBMCPTool.SelectListByMCPIDAndVersion(ctx, nil, mcpID, mcpVersion)
	if err != nil {
		return nil, err
	}

	toolConfigs = make([]*interfaces.MCPToolConfigInfo, len(tools))
	for i, tool := range tools {
		toolConfigs[i] = s.modelToMCPToolConfigInfo(tool)
	}
	return toolConfigs, nil
}

func (s *mcpServiceImpl) getMCPToolConfigs(ctx context.Context, mcpConfigs []*model.MCPServerConfigDB) (toolConfigMap map[string][]*interfaces.MCPToolConfigInfo, err error) {
	toolConfigMap = make(map[string][]*interfaces.MCPToolConfigInfo)

	mcpIDs := []string{}
	for _, mcpConfig := range mcpConfigs {
		if mcpConfig.CreationType != interfaces.MCPCreationTypeToolImported.String() {
			continue
		}
		mcpIDs = append(mcpIDs, mcpConfig.MCPID)
	}

	if len(mcpIDs) == 0 {
		return toolConfigMap, nil
	}

	tools, err := s.DBMCPTool.SelectListByMCPIDS(ctx, nil, mcpIDs)
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		key := s.genToolConfigMapKey(tool.MCPID, tool.MCPVersion)
		toolConfigMap[key] = append(toolConfigMap[key], s.modelToMCPToolConfigInfo(tool))
	}
	return toolConfigMap, nil
}

func (s *mcpServiceImpl) genToolConfigMapKey(mcpID string, mcpVersion int) string {
	return fmt.Sprintf("%s-%d", mcpID, mcpVersion)
}

func (s *mcpServiceImpl) modelToMCPToolConfigInfo(tool *model.MCPToolDB) *interfaces.MCPToolConfigInfo {
	if tool == nil {
		return nil
	}
	return &interfaces.MCPToolConfigInfo{
		BoxID:           tool.BoxID,
		BoxName:         tool.BoxName,
		ToolID:          tool.ToolID,
		ToolName:        tool.Name,
		ToolDescription: tool.Description,
		UseRule:         tool.UseRule,
	}
}
