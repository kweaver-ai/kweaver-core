package agentsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"go.opentelemetry.io/otel/attribute"
)

// TerminateChat 终止聊天
// 如果 agentRunID 不为空，先调用 Executor 终止，再执行原有逻辑
// 如果 interruptedAssistantMessageID 不为空，更新消息状态为 cancelled
func (agentSvc *agentSvc) TerminateChat(ctx context.Context, conversationID string, agentRunID string, interruptedAssistantMessageID string) (err error) {
	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String(otelconst.AttrGenAIAgentRunID, agentRunID),
		attribute.String(otelconst.AttrGenAIAssistantMsgID, interruptedAssistantMessageID),
	)
	oteltrace.SetConversationID(ctx, conversationID)

	otellog.LogDebug(ctx, "[TerminateChat] started")

	// 1. 如果提供了 agentRunID，先调用 Executor 终止
	if agentRunID != "" {
		otellog.LogInfo(ctx, fmt.Sprintf("[TerminateChat] calling executor terminate, agentRunID: %s", agentRunID))

		req := &v2agentexecutordto.AgentTerminateReq{
			AgentRunID: agentRunID,
		}
		if err := agentSvc.agentExecutorV2.Terminate(ctx, req); err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[TerminateChat] executor terminate failed: %v", err), err)
			// 继续执行原有逻辑，不阻止 channel 关闭
		}
	}

	// 2. 执行原有的 channel 关闭逻辑
	stopchan, ok := stopChanMap.Load(conversationID)
	if ok && stopchan != nil {
		// 找到 stopchan 且不为 nil，执行关闭操作
		close(stopchan.(chan struct{}))
		stopChanMap.Delete(conversationID)
		otellog.LogInfo(ctx, fmt.Sprintf("[TerminateChat] terminate chat success, conversationID: %s", conversationID))
		agentSvc.logger.Infof("terminate chat success, conversationID: %s", conversationID)
	} else {
		// 找不到 stopchan 或为 nil
		// 只有当 interruptedAssistantMessageID 为空时才返回错误
		if interruptedAssistantMessageID == "" {
			if !ok {
				err = capierr.New500Err(ctx, "stopchan not found in map")
			} else {
				otellog.LogError(ctx, fmt.Sprintf("[TerminateChat] terminate chat failed, conversationID: %s, stopchan is nil", conversationID), nil)
				agentSvc.logger.Errorf("terminate chat failed, conversationID: %s, stopchan is nil", conversationID)

				err = capierr.New500Err(ctx, "stopchan is nil")
			}

			return
		}
		// interruptedAssistantMessageID 不为空时，静默继续执行后续逻辑
	}

	// 3. 如果提供了 interruptedAssistantMessageID，更新消息状态为 cancelled
	if interruptedAssistantMessageID != "" {
		otellog.LogInfo(ctx, fmt.Sprintf("[TerminateChat] updating message status to cancelled, messageID: %s", interruptedAssistantMessageID))

		msgPO, getErr := agentSvc.conversationMsgRepo.GetByID(ctx, interruptedAssistantMessageID)
		if getErr != nil {
			otellog.LogError(ctx, fmt.Sprintf("[TerminateChat] get message failed: %v", getErr), getErr)
			agentSvc.logger.Errorf("[TerminateChat] get message failed, messageID: %s, error: %v", interruptedAssistantMessageID, getErr)
			err = getErr

			return
		}

		if msgPO != nil {
			msgPO.Status = cdaenum.MsgStatusCancelled
			msgPO.UpdateTime = time.Now().Unix()

			if updateErr := agentSvc.conversationMsgRepo.Update(ctx, msgPO); updateErr != nil {
				otellog.LogError(ctx, fmt.Sprintf("[TerminateChat] update message status failed: %v", updateErr), updateErr)
				agentSvc.logger.Errorf("[TerminateChat] update message status failed, messageID: %s, error: %v", interruptedAssistantMessageID, updateErr)
				err = updateErr

				return
			}

			otellog.LogInfo(ctx, fmt.Sprintf("[TerminateChat] message status updated to cancelled, messageID: %s", interruptedAssistantMessageID))
		}
	}

	return
}
