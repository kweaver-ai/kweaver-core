package agentsvc

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/agentrunsvc/chatlogrecord"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/session/sessionreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/panichelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/ctype"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/attribute"
	otelsdklog "go.opentelemetry.io/otel/log"
)

var (
	// NOTE: 终止channel map， 用于终止会话，key为会话ID，value为终止channel
	stopChanMap sync.Map = sync.Map{}
	// NOTE: session map，用于对话恢复，key为会话ID，value为session
	SessionMap sync.Map = sync.Map{}

	// NOTE: key 为assistantMessageID，value 为progress的数组,存储所有状态不为processing的progress，不重复
	// progressMap map[string][]*agentrespvo.Progress = make(map[string][]*agentrespvo.Progress)
	progressMap sync.Map = sync.Map{}

	// NOTE: key 为assistantMessageID，value 为map[srting]bool ,判断一个progress的ID是否已经存在
	// progressSet map[string]map[string]bool = make(map[string]map[string]bool)
	progressSet sync.Map = sync.Map{}
)

const (
	CHANNEL_SIZE = 100
)

// NOTE: 统一的chat服务
func (agentSvc *agentSvc) Chat(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
	var err error

	// NOTE: 使用 invoke_agent span（按 OTel Gen AI Agent Spans 规范）
	newCtx, _ := oteltrace.StartInvokeAgentSpan(ctx, "")
	defer oteltrace.EndSpan(newCtx, err)
	oteltrace.SetAttributes(newCtx,
		attribute.String(otelconst.AttrGenAIOperationName, "invoke_agent"),
		attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
		attribute.String(otelconst.AttrGenAIAgentVersion, req.AgentVersion),
		attribute.String(otelconst.AttrUserID, req.UserID),
	)
	oteltrace.SetConversationID(newCtx, req.ConversationID)

	startAttrs := []otelsdklog.KeyValue{
		otelsdklog.String(otelconst.AttrGenAIAgentID, req.AgentID),
	}
	if req.ConversationID != "" {
		startAttrs = append(startAttrs, otelsdklog.String(otelconst.AttrGenAIConversationID, req.ConversationID))
	}

	otellog.LogDebug(newCtx, "[chat] started", startAttrs...)

	defer func() {
		if err != nil {
			chatlogrecord.LogFailedExecution(ctx, req, err, nil)
		}
	}()

	// NOTE: 1. 根据agentID 和agentVersion 获取agent配置
	// NOTE: Chat接口请求时，agentID 实际值为agentID, APIChat接口请求时，agentID 实际值为agentKey

	// 1.1 通过agent id获取agent信息
	agentInfo, err := agentSvc.squareSvc.GetAgentInfoByIDOrKey(newCtx, &squarereq.AgentInfoReq{
		AgentID:      req.AgentID,
		AgentVersion: req.AgentVersion,
	})
	if err != nil {
		otellog.LogError(newCtx, "[chat] get agent failed", err)

		return nil, rest.NewHTTPError(newCtx, http.StatusInternalServerError,
			apierr.AgentAPP_Agent_GetAgentFailed).WithErrorDetails(fmt.Sprintf("[chat] get agent failed: %v", err))
	}

	// 1.2 传递给AgentExecutor的agentID 前确保实际值为agentID
	req.AgentID = agentInfo.DataAgent.ID

	otellog.LogDebug(newCtx, "[chat] agent info loaded",
		otelsdklog.String(otelconst.AttrGenAIAgentName, agentInfo.DataAgent.Name),
		otelsdklog.String(otelconst.AttrGenAIAgentID, agentInfo.DataAgent.ID),
	)

	// NOTE: 如果是apichat,但是没有发布成api agent，则返回403
	if req.CallType == constant.APIChat && agentInfo.PublishInfo.IsAPIAgent == 0 {
		httpErr := capierr.NewCustom403Err(newCtx, apierr.AgentAPP_Forbidden_PermissionDenied, "[Chat] apichat is not published")
		return nil, httpErr
	}

	// NOTE: 设置历史上下文限制，从 Agent 配置中获取
	historyLimit := constant.DefaultHistoryLimit
	if agentInfo.Config.ConversationHistoryConfig != nil && agentInfo.Config.ConversationHistoryConfig.CountParams != nil && agentInfo.Config.ConversationHistoryConfig.CountParams.CountLimit > 0 {
		historyLimit = agentInfo.Config.ConversationHistoryConfig.CountParams.CountLimit
	}

	conversationPO, contexts, msgIndex, err := agentSvc.GetHistoryAndMsgIndex(newCtx, req, historyLimit, agentInfo.Config.ConversationHistoryConfig)
	if err != nil {
		otellog.LogError(newCtx, "[chat] get history and msg index failed", err)
		return nil, err
	}

	oteltrace.SetConversationID(newCtx, req.ConversationID)

	// NOTE: 3. 插入用户消息和助手消息, 并返回userMessageID, assistantMessageID, assistantMessageIndex
	req.UserMessageID, req.AssistantMessageID, req.AssistantMessageIndex, err = agentSvc.UpsertUserAndAssistantMsg(newCtx, req, msgIndex, conversationPO)
	if err != nil {
		otellog.LogError(newCtx, "[chat] upsert user and assistant msg failed", err)
		return nil, err
	}

	// NOTE: 4.  创建一个stop_channel 关联conversationID
	stopChan := make(chan struct{})
	stopChanMap.Store(req.ConversationID, stopChan)

	// NOTE: 5. 创建一个session 关联conversationID 用于会话恢复
	session := &Session{
		RWMutex:        sync.RWMutex{},
		ConversationID: req.ConversationID,
		TempMsgResp:    agentresp.ChatResp{},
		Signal:         nil,
		IsResuming:     false,
	}
	SessionMap.Store(req.ConversationID, session)

	progressMap.Store(req.AssistantMessageID, make([]*agentrespvo.Progress, 0))
	progressSet.Store(req.AssistantMessageID, make(map[string]bool, 0))

	// NOTE: 创建一个session
	manageReq := sessionreq.ManageReq{
		Action:         sessionreq.SessionManageActionGetInfoOrCreate,
		AgentID:        req.AgentID,
		AgentVersion:   req.AgentVersion,
		ConversationID: req.ConversationID,
	}

	startTime, _, err := agentSvc.sessionSvc.HandleGetInfoOrCreate(ctx, manageReq, &ctype.VisitorInfo{
		XAccountID:        req.XAccountID,
		XAccountType:      req.XAccountType,
		XBusinessDomainID: cenum.BizDomainID(req.XBusinessDomainID),
	}, false)
	if err != nil {
		return nil, err
	}

	// NOTE: 确保 Sandbox Session 存在并就绪（仅在启用沙箱时执行）
	var sandboxSessionID string

	if agentSvc.sandboxPlatformConf.Enable {
		sessionID := cutil.GetSandboxSessionID()

		var sandboxErr error

		sandboxSessionID, sandboxErr = agentSvc.EnsureSandboxSession(newCtx, sessionID, req)
		if sandboxErr != nil {
			otellog.LogWarn(newCtx, fmt.Sprintf("[chat] ensure sandbox session failed: %v", sandboxErr))
		}
	}

	// 将 sandbox_session_id 传递给 Agent Executor
	req.SandboxSessionID = sandboxSessionID

	// NOTE: 生成ConversationSessionID
	req.ConversationSessionID = fmt.Sprintf("%s-%d", req.ConversationID, startTime)

	// NOTE: 6. 生成agent call请求
	agentCallReq, err := agentSvc.GenerateAgentCallReq(newCtx, req, contexts, agentInfo)
	if err != nil {
		agentSvc.logger.Errorf("[Chat] generate agent call req err: %v", err)
		otellog.LogError(newCtx, "[chat] generate agent call req err", err)

		return nil, err
	}

	// NOTE: 7. 调用agent-executor
	// 创建一个不带取消的ctx，复制可观测性信息
	callCtx := context.WithoutCancel(ctx)
	// 创建一个带取消的ctx，用于终止对话时取消agent-executor的请求
	cancelCtx, cancel := context.WithCancel(callCtx)

	agentCall := &AgentCall{
		callCtx:         cancelCtx,
		req:             agentCallReq,
		agentExecutorV1: agentSvc.agentExecutorV1,
		agentExecutorV2: agentSvc.agentExecutorV2,
		cancelFunc:      cancel,
	}

	var messageChan chan string

	var errChan chan error

	// 统一调用 Call 方法（Resume 信息通过 _options 传递）
	// 原有逻辑分两个分支调用 Resume/Call，现统一为 Call
	messageChan, errChan, err = agentCall.Call()
	if err != nil {
		// NOTE: 发生错误，将assistantMessage 状态设置为failed
		conversationAssistantMsgPO, _ := agentSvc.conversationMsgRepo.GetByID(callCtx, req.AssistantMessageID)
		conversationAssistantMsgPO.Status = cdaenum.MsgStatusFailed
		_ = agentSvc.conversationMsgRepo.Update(callCtx, conversationAssistantMsgPO)
		agentSvc.logger.Errorf("[Chat] call agent executor err: %v", err)
		otellog.LogError(newCtx, "[chat] call agent executor err", err)

		return nil, rest.NewHTTPError(newCtx, http.StatusInternalServerError,
			apierr.AgentAPP_Agent_CallAgentExecutorFailed).WithErrorDetails(fmt.Sprintf("[chat] call agent executor err: %v", err))
	}

	// NOTE: 8. 流式响应处理
	channel := make(chan []byte, CHANNEL_SIZE)

	processAttrs := []otelsdklog.KeyValue{
		otelsdklog.String(otelconst.AttrGenAIAssistantMsgID, req.AssistantMessageID),
	}
	if req.ConversationID != "" {
		processAttrs = append(processAttrs, otelsdklog.String(otelconst.AttrGenAIConversationID, req.ConversationID))
	}

	otellog.LogDebug(newCtx, "[chat] starting Process goroutine", processAttrs...)

	go func() {
		defer panichelper.Recovery(agentSvc.logger)
		// NOTE: 使用 WithoutCancel 继承 trace context 但不继承 cancel signal
		traceCtx := context.WithoutCancel(newCtx)
		_ = agentSvc.Process(traceCtx, req, agentInfo, stopChan, channel, messageChan, errChan, agentCall.Cancel)
	}()

	// NOTE: 9. 异步恢复会话
	go func() {
		defer panichelper.Recovery(agentSvc.logger)

		manageReq := sessionreq.ManageReq{
			Action:         sessionreq.SessionManageActionRecoverLifetimeOrCreate,
			AgentID:        req.AgentID,
			AgentVersion:   req.AgentVersion,
			ConversationID: req.ConversationID,
		}

		_ctx := context.Background()

		_, _, err := agentSvc.sessionSvc.HandleRecoverLifetimeOrCreate(_ctx, manageReq, &ctype.VisitorInfo{
			XAccountID:        req.XAccountID,
			XAccountType:      req.XAccountType,
			XBusinessDomainID: cenum.BizDomainID(req.XBusinessDomainID),
		}, false)
		if err != nil {
			agentSvc.logger.Errorf("[Chat] SessionManage RecoverLifetimeOrCreate err: %v", err)
		}
	}()

	return channel, nil
}
