package agentsvc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/util"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

// buildUserQuery 将文件信息注入到用户问题中（已废弃，保留用于向后兼容）
// Deprecated: 使用 BuildWorkspaceContextMessage 代替
func buildUserQuery(originalQuery string, conversationID string, selectedFiles []agentreq.SelectedFile) string {
	return util.BuildWorkspaceContextMessage(conversationID, "", selectedFiles) + originalQuery
}

// buildWorkspaceContextMessage 生成独立的工作区上下文消息
// 实际逻辑已移至 util.BuildWorkspaceContextMessage
func buildWorkspaceContextMessage(conversationID string, userID string, selectedFiles []agentreq.SelectedFile) string {
	return util.BuildWorkspaceContextMessage(conversationID, userID, selectedFiles)
}
