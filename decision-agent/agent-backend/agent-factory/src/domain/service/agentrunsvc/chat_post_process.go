package agentsvc

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/agentrunsvc/chatlogrecord"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentresperr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo/daresvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// NOTE: 对话后处理模块
// chunkIndex: 流式周期索引（0-based），仅首(0)尾(isEnd=true)周期创建独立 span
func (agentSvc *agentSvc) AfterProcess(ctx context.Context, callResult []byte, req *agentreq.ChatReq, agent *squareresp.AgentMarketAgentInfoResp, chunkIndex int) ([]byte, bool, error) {
	var err error

	var newData []byte

	var isEnd bool

	// A+D 融合：仅首个周期创建独立 span，末尾由 isEnd 在返回后由调用方处理
	if chunkIndex == 0 {
		ctx, _ = oteltrace.StartInternalSpan(ctx)
		defer oteltrace.EndSpan(ctx, err)
		oteltrace.SetAttributes(ctx,
			attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
			attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
			attribute.String(otelconst.AttrUserID, req.UserID),
			attribute.String("stream.chunk_position", "first"),
		)
		oteltrace.SetConversationID(ctx, req.ConversationID)
	}

	var chatResponse agentresp.ChatResp
	// // 1. 获取agentV3

	// 2. 获取output配置中的variables
	outputVariablesS := agentconfigvo.NewOutputVariablesS()

	err = outputVariablesS.LoadFromConfig(&agent.Config)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] load output variables err: %v", err))
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(err, "[AfterProcess] load output variables err: %v", err)
	}

	if outputVariablesS.AnswerVar == "" {
		otellog.LogWarn(ctx, "[AfterProcess] outputVariablesS.AnswerVar is empty")

		err = errors.New("[getChatDataProcessV3]: outputVariablesS.AnswerVar is empty")
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(err, "[AfterProcess] outputVariablesS.AnswerVar is empty")
	}

	// 3. 转换为VariableV3
	var outputVariables *agentconfigvo.Variable

	outputVariables, err = outputVariablesS.ToVariable()
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] to variable err: %v", err))
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(err, "[AfterProcess] to variable err: %v", err)
	}

	// 1. 解析data
	result, err := daresvo.NewDataAgentRes(ctx, callResult, outputVariablesS)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] new data agent res err: %v", err))
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(err, "[AfterProcess] new data agent res err: %v", err)
	}

	// 4. 检查状态是否为end
	if result.Status == "False" {
		isEnd = false
	} else if result.Status == "True" {
		isEnd = true
	} else if result.Status == "Error" {
		isEnd = true
	}

	var (
		answerType        = chatresenum.AnswerTypeOther
		answer            = &conversationmsgvo.Answer{}
		answerTypeOther   interface{}
		otherVariablesMap map[string]interface{}
		// middleOutputVars  *agentrespvo.MiddleOutputVarRes
	)

	// 7. 拿到思考过程（explor模式）
	// 拿到思考过程
	// 最终answer 由思考过程
	exploreAnswerList, ok := result.GetExploreAnswerList()

	if ok {
		answerType = chatresenum.AnswerTypeExplore
	}

	// 8. 处理思考过程数据
	// 生成 name 到 type 的map
	var (
		thinking      string
		skillsProcess []*conversationmsgvo.SkillsProcessItem
	)

	nameToTypeMap := make(map[string]string)

	if len(exploreAnswerList) > 0 {
		dto := handleExploreDto{
			exploreAnswerList: exploreAnswerList,
			nameToTypeMap:     nameToTypeMap,
		}

		thinking, skillsProcess, err = agentSvc.handleExplore(ctx, dto)
		if err != nil {
			otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] handle explore err: %v", err))
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
			chatResponse.Error = httpErr
			bytes, _ := sonic.Marshal(chatResponse)

			return bytes, false, errors.Wrapf(err, "[AfterProcess] handle explore err: %v", err)
		}
	} else {
		if tmp, ok1 := result.IsPromptType(); ok1 {
			answerType = chatresenum.AnswerTypePrompt

			answerPromptTxt := &agentrespvo.AnswerPromptText{}

			answerPromptTxt.Text = tmp.Answer

			thinking = tmp.Think

			answer.Text = answerPromptTxt.Text
		} else {
			answerTypeOther = result.GetFinalAnswer()
		}
	}

	// 9. 中断信息（从 Executor 响应中获取）
	interruptInfo := result.InterruptInfo

	// 10. 相关问题
	var qs []string
	for _, q := range result.RelatedQueries() {
		qs = append(qs, q.Query)
	}

	// 11. 其他字段
	otherVariablesMap, err = result.GetOtherVarsMap()
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] get other vars map err: %v", err))
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(err, "[AfterProcess] get other vars map err: %v", err)
	}
	// 13. 构建chatResponse
	progresses := result.Answer.Progress

	// DEBUG: 记录 _evidence 字段状态
	for idx, pg := range progresses {
		if pg.Evidence != nil && len(pg.Evidence) > 0 {
			agentSvc.logger.Infof("[AfterProcess] Progress[%d] has _evidence: stage=%s, count=%d", idx, pg.Stage, len(pg.Evidence))
		} else {
			agentSvc.logger.Debugf("[AfterProcess] Progress[%d] no _evidence: stage=%s, Evidence=%v", idx, pg.Stage, pg.Evidence)
		}
	}

	if req.ChatOption.IsNeedDocRetrivalPostProcess {
		// NOTE:先对progresses进行处理
		progresses = agentSvc.addCitesToProgress(ctx, progresses, true)
	}

	// TODO: 这里progress 的处理应该还是需要的，只是结果可以不返回
	progressAns, err := agentSvc.handleProgress(ctx, req, progresses, -1)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] handle progress err: %v", err))
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(err, "[AfterProcess] handle progress err: %v", err)
	}

	// NOTE: 当状态为Error时，如果progressAns为空，尝试从progressMap中获取
	if result.Status == "Error" && len(progressAns) == 0 {
		if v, ok := progressMap.Load(req.AssistantMessageID); ok {
			if pgs, ok := v.([]*agentrespvo.Progress); ok {
				progressAns = pgs
				agentSvc.logger.Debugf("[AfterProcess] status is Error, loaded progress from progressMap, count: %d", len(progressAns))
			} else {
				agentSvc.logger.Warnf("[AfterProcess] progressMap has wrong type for assistantMessageID: %s", req.AssistantMessageID)
				// 提供回退机制：创建一个默认的progress对象
				progressAns = []*agentrespvo.Progress{
					{
						ID:     "fallback-" + req.AssistantMessageID,
						Status: "failed",
						Answer: map[string]interface{}{
							"text": "[AfterProcess] agent error, please try again",
						},
					},
				}
			}
			// 清理progressMap，避免内存泄漏
			progressMap.Delete(req.AssistantMessageID)
		} else {
			agentSvc.logger.Debugf("[AfterProcess] status is Error, progressMap is empty for assistantMessageID: %s", req.AssistantMessageID)
			// 提供回退机制：创建一个默认的progress对象
			progressAns = []*agentrespvo.Progress{
				{
					ID:     "fallback-" + req.AssistantMessageID,
					Status: "failed",
					Answer: map[string]interface{}{
						"text": "[AfterProcess] agent error, please try again",
					},
				},
			}
		}
	}
	// NOTE: 计算TTFT，单位ms
	if req.TTFT == 0 {
		req.TTFT = CalculateTTFT(req.ReqStartTime, progressAns, req.CallType)
	}

	if !req.ChatOption.IsNeedProgress {
		progressAns = []*agentrespvo.Progress{}
	}

	var (
		totalTime   float64 = 0.0
		totalTokens int64   = 0
	)

	if isEnd {
		// NOTE: 计算总时长=最后一个Progress的EndTime-第一个Progress的StartTime
		if len(progressAns) > 0 {
			totalTime = progressAns[len(progressAns)-1].EndTime - progressAns[0].StartTime
			// NOTE: 保留两位小数
			totalTime = math.Round(totalTime*100) / 100
			if totalTime < 0 {
				totalTime = 0
			}
		}

		for _, progress := range progressAns {
			totalTokens += progress.TokenUsage.TotalTokens
		}
	}

	content := conversationmsgvo.AssistantContent{
		FinalAnswer: conversationmsgvo.FinalAnswer{
			Query:                 req.Query,
			Answer:                *answer,
			SkillProcess:          skillsProcess,
			Thinking:              thinking,
			AnswerTypeOther:       answerTypeOther,
			OutputVariablesConfig: outputVariables,
		},
		MiddleAnswer: &conversationmsgvo.MiddleAnswer{
			Progress: progressAns,
			// DocRetrieval:   docRetrievalField,
			// GraphRetrieval: graphRetrievalField,
			OtherVariables: otherVariablesMap,
		},
	}
	messageVO := conversationmsgvo.Message{
		ID:             req.AssistantMessageID,
		ConversationID: req.ConversationID,
		Role:           cdaenum.MsgRoleAssistant,
		Content:        content,
		ContentType:    answerType,
		// Status: cdaenum.MsgStatusSuccess,
		ReplyID: req.UserMessageID,
		AgentInfo: valueobject.AgentInfo{
			AgentID:      req.AgentID,
			AgentVersion: req.AgentVersion,
			AgentName:    agent.DataAgent.Name,
		},
		Index: req.AssistantMessageIndex,
		Ext: &conversationmsgvo.MessageExt{
			InterruptInfo:  interruptInfo,
			RelatedQueries: qs,
			TotalTime:      totalTime,
			TotalTokens:    totalTokens,
			TTFT:           req.TTFT,
			AgentRunID:     result.AgentRunID,
			Error:          convertErrorToRespError(result.Error),
		},
	}
	chatResponse = agentresp.ChatResp{
		ConversationID:     req.ConversationID,
		AgentRunID:         result.AgentRunID, // 从 Executor 响应中传递到顶层
		UserMessageID:      req.UserMessageID,
		AssistantMessageID: req.AssistantMessageID,
		Message:            messageVO,
	}

	// 14. 结束时的处理
	// NOTE: 绑定临时区，更新会话和消息
	if isEnd {
		err = agentSvc.handleMessageAndTempArea(ctx, req, messageVO)
		if err != nil {
			otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] handle message and temp area err: %v", err))
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
			chatResponse.Error = httpErr
			bytes, _ := sonic.Marshal(chatResponse)

			return bytes, false, errors.Wrapf(err, "[AfterProcess] handle message and temp area err: %v", err)
		}

		if result.Status == "True" {
			chatlogrecord.LogSuccessExecution(ctx, req, progressAns, totalTime, totalTokens)
		}
	}
	// 检查状态是否为error
	if result.Status == "Error" {
		// 如果报错，记录错误码，直接返回
		otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] agent call failed, error: %v", result.Error))
		httpErr := TransformErrorToHTTPError(ctx, result.Error)
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(httpErr, "[AfterProcess]: agent call failed, error: %v", result.Error)
	}

	startTime := time.Now()
	// 15. 将chatResponse序列化
	newData, err = sonic.Marshal(chatResponse)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[AfterProcess] marshal chat response err: %v", err))
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
		chatResponse.Error = httpErr
		bytes, _ := sonic.Marshal(chatResponse)

		return bytes, false, errors.Wrapf(err, "[AfterProcess] marshal chat response err: %v", err)
	}

	// DEBUG: 检查序列化后的 JSON 是否包含 _evidence
	jsonStr := string(newData)
	if strings.Contains(jsonStr, "_evidence") {
		agentSvc.logger.Infof("[AfterProcess] ✅ Marshaled JSON contains '_evidence' field")
		// 查找第一个 _evidence 的位置并打印上下文
		if idx := strings.Index(jsonStr, "_evidence"); idx > 0 {
			start := idx - 50
			if start < 0 {
				start = 0
			}
			end := idx + 100
			if end > len(jsonStr) {
				end = len(jsonStr)
			}
			agentSvc.logger.Infof("[AfterProcess] Context: ...%s...", jsonStr[start:end])
		}
	} else {
		agentSvc.logger.Warnf("[AfterProcess] ⚠️ Marshaled JSON does NOT contain '_evidence' field")
	}

	marshalTime := time.Since(startTime)
	// NOTE: 打印序列化时间，ms
	agentSvc.logger.Debugf("[AfterProcess] marshal chat response time: %d ms", marshalTime.Milliseconds())

	return newData, isEnd, err
}

// NOTE: 将助手消息持久化，并绑定临时区
func (agentSvc *agentSvc) handleMessageAndTempArea(ctx context.Context, req *agentreq.ChatReq, messageVO conversationmsgvo.Message) error {
	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, nil)
	oteltrace.SetAttributes(ctx, attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID))
	oteltrace.SetAttributes(ctx, attribute.String(otelconst.AttrGenAIAgentID, req.AgentID))
	oteltrace.SetAttributes(ctx, attribute.String(otelconst.AttrUserID, req.UserID))
	oteltrace.SetConversationID(ctx, req.ConversationID)
	// NOTE: VO-PO
	content, err := sonic.Marshal(messageVO.Content)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[handleMessageAndTempArea] marshal msgResp.Message.Content err: %v", err))
		agentSvc.logger.Errorf("[handleMessageAndTempArea] marshal msgResp.Message.Content err: %v", err)

		return err
	}

	ext, err := sonic.Marshal(messageVO.Ext)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[handleMessageAndTempArea] marshal msgResp.Message.Ext err: %v", err))
		agentSvc.logger.Errorf("[handleMessageAndTempArea] marshal msgResp.Message.Ext err: %v", err)

		return err
	}

	contentStr := string(content)
	extStr := string(ext)
	agentSvc.logger.Debugf("[handleMessageAndTempArea] extStr: %s", extStr)
	msgPO := dapo.ConversationMsgPO{
		ID:             req.AssistantMessageID,
		AgentAPPKey:    req.AgentAPPKey,
		ConversationID: req.ConversationID,
		AgentID:        req.AgentID,
		AgentVersion:   req.AgentVersion,
		ReplyID:        req.UserMessageID,
		Role:           cdaenum.MsgRoleAssistant,
		Index:          req.AssistantMessageIndex,
		// Repo更新字段
		Content:     &contentStr,
		ContentType: cdaenum.ConversationMsgContentType(messageVO.ContentType),
		Status:      cdaenum.MsgStatusSucceded,
		Ext:         &extStr,
		UpdateTime:  cutil.GetCurrentMSTimestamp(),
		UpdateBy:    req.UserID,
	}

	if messageVO.IsInterrupted() {
		msgPO.Status = cdaenum.MsgStatusProcessing
	}

	err = agentSvc.conversationMsgRepo.Update(ctx, &msgPO)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[handleMessageAndTempArea] update msgPO err: %v", err))
		agentSvc.logger.Errorf("[handleMessageAndTempArea] update msgPO err: %v", err)

		return err
	}
	// NOTE: 获取消息的下标，更新会话的更新时间和最大下标
	conversationPO, err := agentSvc.conversationRepo.GetByID(ctx, req.ConversationID)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[handleMessageAndTempArea] get conversationPO err: %v", err))
		agentSvc.logger.Errorf("[handleMessageAndTempArea] get conversationPO err: %v", err)

		return err
	}

	conversationPO.MessageIndex = msgPO.Index
	conversationPO.UpdateTime = cutil.GetCurrentMSTimestamp()

	err = agentSvc.conversationRepo.Update(ctx, conversationPO)
	if err != nil {
		otellog.LogWarn(ctx, fmt.Sprintf("[handleMessageAndTempArea] update conversationPO err: %v", err))
		agentSvc.logger.Errorf("[handleMessageAndTempArea] update conversationPO err: %v", err)

		return err
	}

	return nil
}

func (agentSvc *agentSvc) addCitesToProgress(ctx context.Context, progresses []*agentrespvo.Progress, markCite bool) []*agentrespvo.Progress {
	for _, progress := range progresses {
		// NOTE: 如果agentName为doc_qa，则需要将结果加上引用tag
		if progress.AgentName == "doc_qa" && progress.Status == "completed" {
			bytes, err := sonic.Marshal(progress.Answer)
			if err != nil {
				otellog.LogWarn(ctx, fmt.Sprintf("[addCitesToProgress] marshal progress answer err: %v", err))
				agentSvc.logger.Errorf("[addCitesToProgress] marshal progress answer err: %v", err)

				continue
			}
			// NOTE: 将doc_qa的answer反序列化
			var docRetrievalAns agentrespvo.DocRetrievalAnswer

			err = sonic.Unmarshal(bytes, &docRetrievalAns)
			if err != nil {
				otellog.LogWarn(ctx, fmt.Sprintf("[addCitesToProgress] unmarshal progress answer err: %v", err))
				agentSvc.logger.Errorf("[addCitesToProgress] unmarshal progress answer err: %v", err)

				continue
			}

			answer := docRetrievalAns.FullResult.Text

			cites := make([]*agentrespvo.AnswerCite, 0)
			for _, reference := range docRetrievalAns.FullResult.References {
				cites = append(cites, &agentrespvo.AnswerCite{ //nolint:staticcheck // SA4010 暂忽略
					Content:  reference.Content,
					Meta:     reference.Meta,
					CiteType: reference.RetrieveSourceType,
					Score:    reference.Score,
				})
			}

			docRetrievalField := &agentrespvo.DocRetrievalField{
				Text: answer,
			}

			progress.Answer.(map[string]interface{})["full_result"].(map[string]interface{})["text"] = docRetrievalField.Text
			progress.Answer.(map[string]interface{})["full_result"].(map[string]interface{})["cites"] = docRetrievalField.Cites
			// NOTE: 将references清空
			progress.Answer.(map[string]interface{})["full_result"].(map[string]interface{})["references"] = []interface{}{}
		}
	}

	return progresses
}

func TransformErrorToHTTPError(ctx context.Context, err interface{}) *rest.HTTPError {
	errMap, ok := err.(map[string]interface{})
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(fmt.Sprintf("[AfterProcess]: agent call failed, error: %v", err))
	}

	if errMap != nil {
		if errCode, ok := errMap["error_code"]; ok {
			errCodeStr, ok := errCode.(string)
			if !ok {
				otellog.LogWarn(ctx, fmt.Sprintf("[TransformErrorToHTTPError] errCode is not a string: %v", errCode))
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(fmt.Sprintf("[AfterProcess]: agent call failed, error: %v", err))
			}

			switch errCodeStr {
			case "AgentExecutor.DolphinSDKException.ModelExecption":
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_Agent_ModelExecption).WithErrorDetails(fmt.Sprintf("[AfterProcess]: agent call failed, cause: %v", errMap["error_details"]))
			case "AgentExecutor.DolphinSDKException.SkillExecption":
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_Agent_SkillExecption).WithErrorDetails(fmt.Sprintf("[AfterProcess]: agent call failed, cause: %v", errMap["error_details"]))
			case "AgentExecutor.DolphinSDKException.BaseExecption":
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_Agent_DolphinSDKExecption).WithErrorDetails(fmt.Sprintf("[AfterProcess]: agent call failed, cause: %v", errMap["error_details"]))
			default:
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_Agent_ExecutorExecption).WithErrorDetails(fmt.Sprintf("[AfterProcess]: agent call failed, cause: %v", errMap["error_details"]))
			}
		} else {
			otellog.LogWarn(ctx, fmt.Sprintf("[TransformErrorToHTTPError] error code is nil: %v", err))
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(fmt.Sprintf("[AfterProcess]: agent call failed, error: %v", err))
		}
	}

	return nil
}

func convertErrorToRespError(err interface{}) *agentresperr.RespError {
	if err == nil {
		return nil
	}

	errMap, ok := err.(map[string]interface{})
	if !ok {
		return agentresperr.NewRespError(agentresperr.RespErrorTypeAgentFactory, err)
	}

	return agentresperr.NewRespError(agentresperr.RespErrorTypeAgentExecutor, errMap)
}
