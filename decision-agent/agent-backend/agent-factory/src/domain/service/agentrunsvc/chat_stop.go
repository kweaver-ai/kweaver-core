package agentsvc

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// NOTE: 处理终止信号，对话终止时，进行 助手消息的持久化
func (agentSvc *agentSvc) HandleStopChan(ctx context.Context, req *agentreq.ChatReq, session *Session) error {
	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
		attribute.String(otelconst.AttrUserID, req.UserID),
	)
	oteltrace.SetConversationID(ctx, req.ConversationID)

	msgResp := session.GetTempMsgResp()

	if msgResp.Message.Content == nil {
		otellog.LogInfo(ctx, "[HandleStopChan] msgResp.Message.Content is nil")
		agentSvc.logger.Infof("[HandleStopChan] msgResp.Message.Content is nil")
	} else {
		contentBytes, err := sonic.Marshal(msgResp.Message.Content)
		if err != nil {
			otellog.LogError(ctx, "[HandleStopChan] marshal msgResp.Message.Content err", err)
			return errors.Wrapf(err, "[HandleStopChan] marshal msgResp.Message.Content err")
		}

		otellog.LogInfo(ctx, fmt.Sprintf("[HandleStopChan] msgResp.Message.Content: %s", string(contentBytes)))
	}

	existingMsgPO, err := agentSvc.conversationMsgRepo.GetByID(ctx, req.AssistantMessageID)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[HandleStopChan] get message %s err", req.AssistantMessageID), err)
		return errors.Wrapf(err, "[HandleStopChan] get message err")
	}

	if existingMsgPO == nil {
		otellog.LogInfo(ctx, "[HandleStopChan] message does not exist, creating new message")
		agentSvc.logger.Infof("[HandleStopChan] message does not exist, creating new message")

		msgPO, _, err := agentSvc.MsgResp2MsgPO(ctx, msgResp, req)
		if err != nil {
			otellog.LogError(ctx, "[HandleStopChan] convert msgResp to msgPO err", err)
			return errors.Wrapf(err, "[HandleStopChan] convert msgResp to msgPO err")
		}

		msgPO.Status = cdaenum.MsgStatusCancelled
		msgPO.UpdateTime = cutil.GetCurrentMSTimestamp()

		_, err = agentSvc.conversationMsgRepo.Create(ctx, &msgPO)
		if err != nil {
			otellog.LogError(ctx, "[HandleStopChan] create message err", err)
			return errors.Wrapf(err, "[HandleStopChan] create message err")
		}
	} else {
		if existingMsgPO.Content != nil {
			otellog.LogInfo(ctx, fmt.Sprintf("[HandleStopChan] existingMsgPO.Content: %s", *existingMsgPO.Content))
		} else {
			otellog.LogInfo(ctx, "[HandleStopChan] existingMsgPO.Content is nil")
			agentSvc.logger.Infof("[HandleStopChan] existingMsgPO.Content is nil")
		}

		msgPO, _, err := agentSvc.MsgResp2MsgPO(ctx, msgResp, req)
		if err != nil {
			otellog.LogError(ctx, "[HandleStopChan] convert msgResp to msgPO err", err)
			return errors.Wrapf(err, "[HandleStopChan] convert msgResp to msgPO err")
		}

		if msgPO.Content != nil {
			otellog.LogInfo(ctx, fmt.Sprintf("[HandleStopChan] msgPO.Content: %s", *msgPO.Content))
		} else {
			otellog.LogInfo(ctx, "[HandleStopChan] msgPO.Content is nil")
			agentSvc.logger.Infof("[HandleStopChan] msgPO.Content is nil")
		}

		existingMsgPO.Content = msgPO.Content
		existingMsgPO.ContentType = msgPO.ContentType
		existingMsgPO.Ext = msgPO.Ext
		existingMsgPO.Status = cdaenum.MsgStatusCancelled
		existingMsgPO.UpdateTime = cutil.GetCurrentMSTimestamp()

		otellog.LogInfo(ctx, "[HandleStopChan] message exists, updating content and status to cancelled")
		agentSvc.logger.Infof("[HandleStopChan] message exists, updating content and status to cancelled")

		err = agentSvc.conversationMsgRepo.Update(ctx, existingMsgPO)
		if err != nil {
			otellog.LogError(ctx, "[HandleStopChan] update message err", err)
			return errors.Wrapf(err, "[HandleStopChan] update message err")
		}
	}

	conversationPO, err := agentSvc.conversationRepo.GetByID(ctx, req.ConversationID)
	if err != nil {
		otellog.LogError(ctx, "[HandleStopChan] get conversationPO err", err)
		return errors.Wrapf(err, "[HandleStopChan] get conversationPO err")
	}

	conversationPO.UpdateTime = cutil.GetCurrentMSTimestamp()
	conversationPO.MessageIndex = req.AssistantMessageIndex

	// 更新会话
	err = agentSvc.conversationRepo.Update(ctx, conversationPO)
	if err != nil {
		otellog.LogError(ctx, "[HandleStopChan] update conversationPO err", err)
		return errors.Wrapf(err, "[HandleStopChan] update conversationPO err")
	}

	otellog.LogInfo(ctx, "[HandleStopChan] terminate chat success")

	return nil
}
