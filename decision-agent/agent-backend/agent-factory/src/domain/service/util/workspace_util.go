package util

import (
	"fmt"
	"path"
	"strings"

	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// buildWorkspaceContextMessage 生成独立的工作区上下文消息
// 作为单独的 "user" 角色消息插入，与用户问题分离
// selectedFiles.FileName 现在包含完整路径，例如：/workspace/conv-123/uploads/temparea/data.csv
// 生成的格式：
// 【System auto-generated context - not user query】
//
// Current workspace path: /workspace/{conversation_id}/uploads/temparea/
// Sandbox Session ID: sess-agent-default (IMPORTANT: This ID is required for code execution tool calls)
//
// Available files:
// - data.csv (/workspace/conv-123/uploads/temparea/data.csv)
// - config.json (/workspace/conv-123/uploads/temparea/config.json)
func BuildWorkspaceContextMessage(conversationID string, userID string, selectedFiles []agentreq.SelectedFile) string {
	if len(selectedFiles) == 0 {
		return ""
	}

	var fileList strings.Builder

	for _, file := range selectedFiles {
		// file.FileName 现在是完整路径，例如：/workspace/conv-123/uploads/temparea/data.csv
		// 提取文件名用于显示，完整路径直接使用
		fullPath := file.FileName
		fileName := path.Base(fullPath)
		fileList.WriteString(fmt.Sprintf("- %s (%s)\n", fileName, fullPath))
	}

	rootPath := fmt.Sprintf("/workspace/%s/uploads/temparea/", conversationID)
	sandboxSessionID := cutil.GetSandboxSessionID()
	contextMsg := fmt.Sprintf(`【System auto-generated context - not user query】

Current workspace path: %s
Sandbox Session ID: %s (IMPORTANT: This ID MUST be passed as the 'session_id' parameter when calling code execution tools. The LLM should automatically generate and include this parameter in tool calls.)

Available files:
%s`,
		rootPath,
		sandboxSessionID,
		fileList.String(),
	)

	return contextMsg
}
