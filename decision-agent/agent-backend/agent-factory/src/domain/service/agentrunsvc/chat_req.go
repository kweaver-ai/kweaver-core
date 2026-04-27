package agentsvc

import (
	"context"
	"math"
	"slices"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req/chatopt"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"go.opentelemetry.io/otel/attribute"
)

func (agentSvc *agentSvc) GenerateAgentCallReq(ctx context.Context, req *agentreq.ChatReq, contexts []*comvalobj.LLMMessage, agent *squareresp.AgentMarketAgentInfoResp) (*agentexecutordto.AgentCallReq, error) {
	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
		attribute.String(otelconst.AttrUserID, req.UserID),
	)
	oteltrace.SetConversationID(ctx, req.ConversationID)
	// NOTE: 如果req.ChatMode不为空，则设置req.ChatMode
	if req.ChatMode != constant.DeepThinkingMode {
		req.ChatMode = constant.NormalMode
	}

	if contexts == nil {
		contexts = []*comvalobj.LLMMessage{}
	}
	// NOTE: 如果req.History不为空，应该直接使用req.History
	if len(req.History) > 0 {
		contexts = req.History
	}

	// 新增：将当前消息的上下文添加到 contexts
	// 如果有选中的文件，先插入工作区上下文消息
	// 注意：当前用户查询不需要添加到 contexts，因为它已经通过 Input["query"] 单独传递
	if len(req.SelectedFiles) > 0 {
		contextMsg := &comvalobj.LLMMessage{
			Role:    "user",
			Content: buildWorkspaceContextMessage(req.ConversationID, req.UserID, req.SelectedFiles),
		}
		contexts = append(contexts, contextMsg)
	}

	// NOTE: 动态字段 file  和 自定义变量
	agentCallReq := &agentexecutordto.AgentCallReq{
		ID:           req.AgentID,
		AgentVersion: req.AgentVersion,
		Config:       AgentConfig2AgentCallConfig(ctx, &agent.Config, req),
		Input: map[string]interface{}{
			"query":   req.Query,
			"history": contexts,
			//"confirm_plan": req.ConfirmPlan,
		},
		CallType:          req.CallType,
		ExecutorVersion:   req.ExecutorVersion,
		XAccountID:        req.XAccountID,
		XAccountType:      req.XAccountType,
		XBusinessDomainID: req.XBusinessDomainID,
		// ConversationSessionID: req.ConversationSessionID,
		ChatOption: chatopt.ChatOption{
			EnableDependencyCache: req.ChatOption.EnableDependencyCache,
			// IsNeedDocRetrivalPostProcess: req.ChatOption.IsNeedDocRetrivalPostProcess,
			IsNeedHistory:  req.ChatOption.IsNeedHistory,
			IsNeedProgress: req.ChatOption.IsNeedProgress,
		},
	}
	// NOTE: 将agent.Config.Input.Fields 转换为map，排除一些内置参数
	excludeFields := []string{"history", "query", "header", "tool", "self_config"}

	for _, field := range agent.Config.Input.Fields {
		if field.Type == "file" {
			agentCallReq.Input[field.Name] = req.TempFiles
			continue
		}
		// NOTE: 如果field.Name为内置参数则不进行处理
		if slices.Contains(excludeFields, field.Name) {
			continue
		}
		// NOTE: 如果field.Name为自定义参数，则将req.CustomQuerys[field.Name]赋值给agentCallReq.Input[field.Name]
		agentCallReq.Input[field.Name] = req.CustomQuerys[field.Name]
	}

	// NOTE:根据请求参数切换深度思考大模型
	if req.ChatMode == constant.DeepThinkingMode {
		agentSvc.logger.Infof("[GenerateAgentCallReq] deep_thinking")
		// NOTE: 先将默认的llm设置为false
		for _, llm := range agentCallReq.Config.Llms {
			if llm.IsDefault && llm.LlmConfig.ModelType == cdaenum.ModelTypeLlm {
				llm.IsDefault = false
			}
		}
		// NOTE: 将rlm设置为默认
		for _, llm := range agentCallReq.Config.Llms {
			if llm.LlmConfig.ModelType == cdaenum.ModelTypeRlm {
				llm.IsDefault = true
				break
			}
		}
	}
	// NOTE: 重新生成时调整大模型温度参数
	if req.RegenerateAssistantMsgID != "" {
		// NOTE: 如果传了modelname，则修改对应大模型的温度，否则修改默认大模型的温度
		if req.ModelName != "" {
			for _, llm := range agentCallReq.Config.Llms {
				if llm.LlmConfig.Name == req.ModelName {
					llm.LlmConfig.Temperature = math.Max(llm.LlmConfig.Temperature, 0.8)
					if llm.LlmConfig.TopK < 10 {
						llm.LlmConfig.TopK = 10
					}

					break
				}
			}
		} else {
			for _, llm := range agentCallReq.Config.Llms {
				if llm.IsDefault {
					llm.LlmConfig.Temperature = math.Max(llm.LlmConfig.Temperature, 0.8)
					if llm.LlmConfig.TopK < 10 {
						llm.LlmConfig.TopK = 10
					}

					break
				}
			}
		}
	}
	// NOTE: 鉴权
	agentCallReq.UserID = req.UserID
	agentCallReq.Token = req.Token
	agentCallReq.VisitorType = req.VisitorType

	// 新增：传递中断恢复信息（统一 Run 接口支持恢复执行）
	agentCallReq.ResumeInterruptInfo = req.ResumeInterruptInfo

	return agentCallReq, nil
}
