package agentsvc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// NOTE: 将msgResp转换为msgPO
func (agentSvc *agentSvc) MsgResp2MsgPO(ctx context.Context, msgResp agentresp.ChatResp, req *agentreq.ChatReq) (dapo.ConversationMsgPO, bool, error) {
	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
		attribute.String(otelconst.AttrUserID, req.UserID),
	)
	oteltrace.SetConversationID(ctx, req.ConversationID)
	oteltrace.SetConversationID(ctx, msgResp.ConversationID)

	content, err := sonic.Marshal(msgResp.Message.Content)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[MsgResp2MsgPO] marshal msgResp.Message.Content err: %v", err), err)
		return dapo.ConversationMsgPO{}, false, errors.Wrapf(err, "[MsgResp2MsgPO] marshal msgResp.Message.Content err")
	}

	ext, err := sonic.Marshal(msgResp.Message.Ext)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[MsgResp2MsgPO] marshal msgResp.Message.Ext err: %v", err), err)
		return dapo.ConversationMsgPO{}, false, errors.Wrapf(err, "[MsgResp2MsgPO] marshal msgResp.Message.Ext err")
	}

	contentStr := string(content)
	extStr := string(ext)
	msgPO := dapo.ConversationMsgPO{
		ID:             req.AssistantMessageID,
		AgentAPPKey:    req.AgentAPPKey,
		ConversationID: msgResp.ConversationID,
		AgentID:        req.AgentID,
		AgentVersion:   req.AgentVersion,
		ReplyID:        req.UserMessageID,
		Role:           cdaenum.MsgRoleAssistant,
		Index:          req.AssistantMessageIndex,

		// Repo更新字段
		Content:     &contentStr,
		ContentType: cdaenum.ConversationMsgContentType(msgResp.Message.ContentType),
		Status:      cdaenum.MsgStatusSucceded,
		Ext:         &extStr,
		UpdateTime:  cutil.GetCurrentMSTimestamp(),
		UpdateBy:    req.UserID,
	}

	return msgPO, false, nil
}

// NOTE: 获取会话中的上下文、会话中消息的最大下标、更新req.ConversationID
func (agentSvc *agentSvc) GetHistoryAndMsgIndex(ctx context.Context, req *agentreq.ChatReq, historyLimit int, historyConfig *daconfvalobj.ConversationHistoryConfig) (*dapo.ConversationPO, []*comvalobj.LLMMessage, int, error) {
	var contexts []*comvalobj.LLMMessage

	var conversationPO *dapo.ConversationPO

	var msgIndex int

	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
		attribute.String(otelconst.AttrUserID, req.UserID),
	)
	oteltrace.SetConversationID(ctx, req.ConversationID)
	// NOTE: 从前端请求的conversationID不为空，接口可能为空;
	// NOTE: 如果会话ID为空，则创建新会话；
	if req.ConversationID == "" {
		conversationPO = &dapo.ConversationPO{
			AgentAPPKey: req.AgentAPPKey,
			Title:       "新会话", // todo
			CreateBy:    req.UserID,
			UpdateBy:    req.UserID,
			Ext:         new(string),
		}
		// NOTE: 如果query不为空，则更新会话标题
		if req.Query != "" {
			// NOTE: 用query 的前50个字符作为会话标题，如果query长度小于50个字符，则用query作为会话标题
			// 使用 []rune 来处理 Unicode 字符
			runes := []rune(req.Query)
			if len(runes) < 50 {
				conversationPO.Title = string(runes)
			} else {
				conversationPO.Title = string(runes[:50])
			}
		}

		conversationPO, err = agentSvc.conversationRepo.Create(ctx, conversationPO)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[GetHistoryAndMsgIndex] create conversation failed: %v", err), err)
			return nil, nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_CreateConversationFailed).WithErrorDetails(fmt.Sprintf("[GetHistoryAndMsgIndex] create conversation failed: %v", err))
		}

		req.ConversationID = conversationPO.ID
		oteltrace.SetConversationID(ctx, req.ConversationID)
	} else {
		oteltrace.SetConversationID(ctx, req.ConversationID)
		// 获取对话
		conversationPO, err = agentSvc.conversationRepo.GetByID(ctx, req.ConversationID)
		if err != nil {
			if chelper.IsSqlNotFound(err) {
				otellog.LogWarn(ctx, fmt.Sprintf("[GetHistoryAndMsgIndex] conversation not found: %v", err))
				return nil, nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound,
					apierr.AgentAPP_Agent_GetConversationFailed).WithErrorDetails(fmt.Sprintf("[GetHistoryAndMsgIndex] conversation not found: %v", err))
			}

			otellog.LogError(ctx, fmt.Sprintf("[GetHistoryAndMsgIndex] get conversation failed: %v", err), err)

			return nil, nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_GetConversationFailed).WithErrorDetails(fmt.Sprintf("[GetHistoryAndMsgIndex] get conversation failed: %v", err))
		}
		// NOTE: 获取会话中消息的最大下标，后续创建新消息时需要在这基础上递增
		msgIndex, err = agentSvc.conversationMsgRepo.GetMaxIndexByID(ctx, req.ConversationID)
		if err != nil {
			// NOTE：当前会话未产生消息
			if chelper.IsSqlNotFound(err) {
				msgIndex = 0
			} else {
				otellog.LogError(ctx, fmt.Sprintf("[GetHistoryAndMsgIndex] get max index failed: %v", err), err)
				return nil, nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					apierr.AgentAPP_Agent_GetMaxIndexFailed).WithErrorDetails(fmt.Sprintf("[GetHistoryAndMsgIndex] get max index failed: %v", err))
			}
		}

		if req.ChatOption.IsNeedHistory {
			if historyConfig != nil {
				contexts, err = agentSvc.conversationSvc.GetHistoryV2(ctx, req.ConversationID, historyConfig, req.RegenerateUserMsgID, req.RegenerateAssistantMsgID)
			} else {
				contexts, err = agentSvc.conversationSvc.GetHistory(ctx, req.ConversationID, historyLimit, req.RegenerateUserMsgID, req.RegenerateAssistantMsgID)
			}

			if err != nil {
				otellog.LogError(ctx, fmt.Sprintf("[GetHistoryAndMsgIndex] get conversation messages history failed: %v", err), err)
				return nil, nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					apierr.AgentAPP_Agent_GetHistoryFailed).WithErrorDetails(fmt.Sprintf("[GetHistoryAndMsgIndex] get conversation messages history failed: %v", err))
			}
		}
	}

	return conversationPO, contexts, msgIndex, nil
}

// NOTE: 插入用户消息和助手消息
func (agentSvc *agentSvc) UpsertUserAndAssistantMsg(ctx context.Context, req *agentreq.ChatReq,
	msgIndex int, conversationPO *dapo.ConversationPO,
) (string, string, int, error) {
	userMessageID := ""
	assistantMessageID := ""
	assistantMessageIndex := 0

	var conversationUserMsgPO *dapo.ConversationMsgPO

	var conversationAssistantMsgPO *dapo.ConversationMsgPO

	var err error
	// NOTE: ctx变量名
	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
		attribute.String(otelconst.AttrUserID, req.UserID),
	)
	oteltrace.SetConversationID(ctx, req.ConversationID)
	// NOTE: 普通对话则创建userMessage,状态为recieved
	if IsNormalChat(req) {
		userContent := conversationmsgvo.UserContent{
			Text:          req.Query,
			SelectedFiles: req.SelectedFiles,
		}
		userContentBytes, _ := sonic.Marshal(userContent)
		userContentStr := string(userContentBytes)
		conversationUserMsgPO = &dapo.ConversationMsgPO{
			ConversationID: req.ConversationID,
			AgentAPPKey:    req.AgentAPPKey,
			AgentID:        req.AgentID,
			AgentVersion:   req.AgentVersion,
			Index:          msgIndex + 1,
			Role:           cdaenum.MsgRoleUser,
			Content:        &userContentStr,
			ContentType:    cdaenum.MsgText,
			Status:         cdaenum.MsgStatusProcessed,
			Ext:            new(string),
			CreateBy:       req.UserID,
			UpdateBy:       req.UserID,
		}

		userMessageID, err = agentSvc.conversationMsgRepo.Create(ctx, conversationUserMsgPO)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] create conversation user message failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_CreateMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] create conversation user message failed: %v", err))
		}
		// 更新会话下标
		conversationPO.MessageIndex = conversationUserMsgPO.Index

		err = agentSvc.conversationRepo.Update(ctx, conversationPO)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_UpdateConversationFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation failed: %v", err))
		}
	} else if req.RegenerateUserMsgID != "" {
		// 如果是编辑问题，则更新userMessage
		userMessageID = req.RegenerateUserMsgID

		conversationUserMsgPO, err = agentSvc.conversationMsgRepo.GetByID(ctx, req.RegenerateUserMsgID)
		if err != nil {
			otellog.LogWarn(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation user message [%s] failed: %v", req.RegenerateUserMsgID, err))
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_GetMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation user message [%s] failed: %v", req.RegenerateUserMsgID, err))
		}

		userContent := conversationmsgvo.UserContent{
			Text:          req.Query,
			SelectedFiles: req.SelectedFiles,
		}
		userContentBytes, _ := sonic.Marshal(userContent)
		userContentStr := string(userContentBytes)
		conversationUserMsgPO.Content = &userContentStr
		conversationUserMsgPO.Status = cdaenum.MsgStatusReceived
		conversationUserMsgPO.UpdateBy = req.UserID
		conversationUserMsgPO.UpdateTime = cutil.GetCurrentMSTimestamp()

		err = agentSvc.conversationMsgRepo.Update(ctx, conversationUserMsgPO)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation user message failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_UpdateMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation user message failed: %v", err))
		}
	} else {
		// NOTE: 如果是重新生成或者中断，则获取userMessageID
		if req.RegenerateAssistantMsgID != "" {
			conversationAssistantMsgPO, err = agentSvc.conversationMsgRepo.GetByID(ctx, req.RegenerateAssistantMsgID)
			if err != nil {
				otellog.LogWarn(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.RegenerateAssistantMsgID, err))
				return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					apierr.AgentAPP_Agent_GetMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.RegenerateAssistantMsgID, err))
			}
		} else {
			conversationAssistantMsgPO, err = agentSvc.conversationMsgRepo.GetByID(ctx, req.InterruptedAssistantMsgID)
			if err != nil {
				otellog.LogWarn(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.InterruptedAssistantMsgID, err))
				return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					apierr.AgentAPP_Agent_GetMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.InterruptedAssistantMsgID, err))
			}
		}

		userMessageID = conversationAssistantMsgPO.ReplyID
	}

	// NOTE: 如果req.RegenerateAssistantMsgID 和 req.InterruptedAssistantMsgID 和req.RegenerateUserMsgID == ""都为空，说明当前为普通对话，只需要创建assistantMessage,状态为processing 持久化
	if IsNormalChat(req) {
		conversationAssistantMsgPO = &dapo.ConversationMsgPO{
			ConversationID: req.ConversationID,
			AgentAPPKey:    req.AgentAPPKey,
			AgentID:        req.AgentID,
			AgentVersion:   req.AgentVersion,
			ReplyID:        conversationUserMsgPO.ID,
			Index:          conversationUserMsgPO.Index + 1,
			Role:           cdaenum.MsgRoleAssistant,
			Content:        new(string),
			ContentType:    cdaenum.MsgText,
			Status:         cdaenum.MsgStatusProcessing,
			Ext:            new(string),
			CreateBy:       req.UserID,
			UpdateBy:       req.UserID,
			UpdateTime:     cutil.GetCurrentMSTimestamp(),
		}
		// NOTE: 只创建assistantMessage 不更新会话下标，会话下标在对话完成时更新
		assistantMessageID, err = agentSvc.conversationMsgRepo.Create(ctx, conversationAssistantMsgPO)
		assistantMessageIndex = conversationAssistantMsgPO.Index

		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] create conversation assistant message failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_CreateMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] create conversation assistant message failed: %v", err))
		}
	} else if req.RegenerateAssistantMsgID != "" {
		// NOTE: 如果是重新生成
		conversationAssistantMsgPO, err = agentSvc.conversationMsgRepo.GetByID(ctx, req.RegenerateAssistantMsgID)
		if err != nil {
			otellog.LogWarn(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.RegenerateAssistantMsgID, err))
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_GetMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.RegenerateAssistantMsgID, err))
		}
		// NOTE: 重新生成将assistantMessage 状态设置为processing
		conversationAssistantMsgPO.Status = cdaenum.MsgStatusProcessing

		err = agentSvc.conversationMsgRepo.Update(ctx, conversationAssistantMsgPO)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation assistant message failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_UpdateMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation assistant message failed: %v", err))
		}

		assistantMessageID = req.RegenerateAssistantMsgID
		assistantMessageIndex = conversationAssistantMsgPO.Index
	} else if req.InterruptedAssistantMsgID != "" {
		// NOTE: 如果是中断
		conversationAssistantMsgPO, err = agentSvc.conversationMsgRepo.GetByID(ctx, req.InterruptedAssistantMsgID)
		if err != nil {
			otellog.LogWarn(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.InterruptedAssistantMsgID, err))
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_GetMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", req.InterruptedAssistantMsgID, err))
		}
		// NOTE: 中断将将assistantMessage 状态设置为processing
		conversationAssistantMsgPO.Status = cdaenum.MsgStatusProcessing

		// NOTE: 清除 Ext 中的 InterruptInfo，因为中断恢复后旧的中断信息不再有效
		if conversationAssistantMsgPO.Ext != nil && *conversationAssistantMsgPO.Ext != "" {
			var msgExt conversationmsgvo.MessageExt
			if err = sonic.Unmarshal([]byte(*conversationAssistantMsgPO.Ext), &msgExt); err != nil {
				otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] unmarshal ext err: %v", err), err)
				return userMessageID, assistantMessageID, assistantMessageIndex, errors.Wrapf(err, "[UpsertUserAndAssistantMsg] unmarshal ext err")
			}

			msgExt.InterruptInfo = nil

			extBytes, marshalErr := sonic.Marshal(msgExt)
			if marshalErr != nil {
				otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] marshal ext err: %v", marshalErr), marshalErr)
				return userMessageID, assistantMessageID, assistantMessageIndex, errors.Wrapf(marshalErr, "[UpsertUserAndAssistantMsg] marshal ext err")
			}

			extStr := string(extBytes)
			conversationAssistantMsgPO.Ext = &extStr
		}

		err = agentSvc.conversationMsgRepo.Update(ctx, conversationAssistantMsgPO)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation assistant message failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_UpdateMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation assistant message failed: %v", err))
		}

		assistantMessageID = req.InterruptedAssistantMsgID
		assistantMessageIndex = conversationAssistantMsgPO.Index
	} else if req.RegenerateUserMsgID != "" {
		// NOTE: 如果是编辑用户消息
		// TODO: 后续版本优，同时考虑多版本消息设计
		conversation, err := agentSvc.conversationSvc.Detail(ctx, req.ConversationID)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_GetConversationFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation failed: %v", err))
		}

		for index, msg := range conversation.Messages {
			if msg.ID == req.RegenerateUserMsgID {
				assistantMessageID = conversation.Messages[index+1].ID
				assistantMessageIndex = conversation.Messages[index+1].Index

				break
			}
		}
		// NOTE: 编辑用户消息将assistantMessage 状态设置为processing
		conversationAssistantMsgPO, err = agentSvc.conversationMsgRepo.GetByID(ctx, assistantMessageID)
		if err != nil {
			otellog.LogWarn(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", assistantMessageID, err))
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_GetMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] get conversation assistant message [%s] failed: %v", assistantMessageID, err))
		}

		conversationAssistantMsgPO.Status = cdaenum.MsgStatusProcessing

		err = agentSvc.conversationMsgRepo.Update(ctx, conversationAssistantMsgPO)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation assistant message failed: %v", err), err)
			return userMessageID, assistantMessageID, assistantMessageIndex, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				apierr.AgentAPP_Agent_UpdateMessageFailed).WithErrorDetails(fmt.Sprintf("[UpsertUserAndAssistantMsg] update conversation assistant message failed: %v", err))
		}
	}

	return userMessageID, assistantMessageID, assistantMessageIndex, nil
}

// NOTE: 如果req.RegenerateAssistantMsgID 和 req.InterruptedAssistantMsgID 和req.RegenerateUserMsgID == ""都为空，
func IsNormalChat(req *agentreq.ChatReq) bool {
	return req.RegenerateAssistantMsgID == "" && req.InterruptedAssistantMsgID == "" && req.RegenerateUserMsgID == ""
}
