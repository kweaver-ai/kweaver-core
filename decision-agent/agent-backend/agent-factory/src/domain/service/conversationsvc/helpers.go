package conversationsvc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/util"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

// buildWorkspaceContextMessage 生成独立的工作区上下文消息
// 实际逻辑已移至 util.BuildWorkspaceContextMessage
func buildWorkspaceContextMessage(conversationID string, userID string, selectedFiles []agentreq.SelectedFile) string {
	return util.BuildWorkspaceContextMessage(conversationID, userID, selectedFiles)
}
