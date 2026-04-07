package v3agentconfigsvc

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/imodelfactoryacc"
	"github.com/sashabaranov/go-openai"
)

// AiAutogen AI自动生成内容
// 参数:
//   - ctx: Gin上下文
//   - req: AI自动生成内容请求
//
// 返回:
//   - *agentconfigresp.AiAutogenRes: AI自动生成内容响应
//   - error: 错误信息
//   TODO
//    1. 新增简介
//    2. 人设 systemPrompt 换成思琪提供的提示词

func (s *dataAgentConfigSvc) AIAutogenNotStream(ctx *gin.Context, req *agentconfigreq.AiAutogenReq) (questions agentconfigresp.PreSetQuestions, err error) {
	var chatRes openai.ChatCompletionResponse

	chatRes, err = s.doAiChat(ctx, req)
	if err != nil {
		return
	}

	content := chatRes.Choices[0].Message.Content

	switch req.From {
	case daenum.AiAutogenFromPreSetQuestion:
		tryNum := 5
		for i := 0; i < tryNum; i++ {
			var ok bool
			if questions, ok = s.checkPreSetQuestionResFormat(content); ok {
				return
			}

			// 如果格式不正确，重新生成
			chatRes, err = s.doAiChat(ctx, req)
			if err != nil {
				return
			}

			content = chatRes.Choices[0].Message.Content
		}
	}

	return
}

func (s *dataAgentConfigSvc) doAiChat(ctx *gin.Context, req *agentconfigreq.AiAutogenReq) (chatRes openai.ChatCompletionResponse, err error) {
	// 1. 根据内容来源类型和提示词生成内容
	switch req.From {
	case daenum.AiAutogenFromPreSetQuestion:
		var sysMsg string

		switch req.Language {
		case enUS:
			sysMsg = `Generate preset questions according to the user's requirements.
Format requirements: Return in pure JSON string format, do not output any other content. The format is an array, each string representing one question. For example: ["Question 1", "Question 2", "Question 3", "Question 4"].
Quantity requirements: If the user doesn't specify a quantity, generate 4 preset questions.
Language requirements for preset questions: generate preset questions should  be in English.`
		case zhTW:
			sysMsg = `根據用戶的要求生成預設問題。
格式要求：以純JSON字符串格式返回，不要輸出任何其他內容。格式為數組，每個字符串代表一個問題。如：["問題1", "問題2", "問題3", "問題4"]。
數量要求：如果用戶沒有指定數量，則生成4個預設問題。
生成預設問題的語言要求：生成的預設問題應該使用繁體中文。`
		default:
			sysMsg = `根据用户的要求生成预设问题。
格式要求：以纯JSON字符串格式返回，不要输出任何其他内容。格式为数组，每个字符串代表一个问题。如：["问题1", "问题2", "问题3", "问题4"]。
数量要求：如果用户没有指定数量，则生成4个预设问题。
生成预设问题的语言要求：生成的预设问题应该使用简体中文。`
		}

		userPrompt := userPromptForPresetQuestion(req.Language, req.Params.Name, req.Params.Profile, req.Params.Skills, req.Params.Sources)
		newReq := &imodelfactoryacc.ChatCompletionReq{
			// NOTE: 大模型名称传空，使用默认大模型
			Model:       "",
			Messages:    []imodelfactoryacc.Message{{Role: "system", Content: sysMsg}, {Role: "user", Content: userPrompt}},
			Stream:      false,
			UserID:      req.UserID,
			AccountType: req.AccountType,
		}
		chatRes, err = s.modelFactoryAcc.ChatCompletion(ctx, newReq)
	default:
		err = fmt.Errorf("[AiAutogenNotStream][doAiChat]: 不支持的来源类型: %v", req.From)
	}

	if err != nil {
		err = fmt.Errorf("[AiAutogenNotStream][doAiChat]: 生成内容失败: %v", err)
		s.logger.Errorln(err)

		return
	}

	return
}

func (s *dataAgentConfigSvc) checkPreSetQuestionResFormat(content string) (questions agentconfigresp.PreSetQuestions, ok bool) {
	err := cutil.JSON().UnmarshalFromString(content, &questions)
	if err != nil || len(questions) == 0 {
		return
	}

	ok = true

	return
}

func (s *dataAgentConfigSvc) AIAutogenV3(ctx *gin.Context, req *agentconfigreq.AiAutogenReq) (chan string, chan error, error) {
	// 1. 根据内容来源类型和提示词生成内容
	switch req.From {
	case daenum.AiAutogenFromSystemPrompt:
		// 生成系统提示词
		sysMsg := systemPrompt(req.Language)
		userPrompt := userPromptForSystem(req.Language, req.Params.Name, req.Params.Profile, req.Params.Skills, req.Params.Sources)

		messageChan, errorChan, err := s.chatCompletion(ctx, userPrompt, sysMsg, req.UserID, req.AccountType, true)
		if err != nil {
			return nil, nil, err
		}

		return messageChan, errorChan, nil
	case daenum.AiAutogenFromOpeningRemarks:
		// 生成开场白
		sysMsg := openingRemarksSystemPrompt(req.Language)
		userPrompt := userPromptForOpenRemarks(req.Language, req.Params.Name, req.Params.Profile, req.Params.Skills, req.Params.Sources)

		messageChan, errorChan, err := s.chatCompletion(ctx, userPrompt, sysMsg, req.UserID, req.AccountType, true)
		if err != nil {
			return nil, nil, err
		}

		return messageChan, errorChan, nil
	default:
		err := fmt.Errorf("[AiAutogenV3]: 不支持的类型: %v", req.From)
		return nil, nil, err
	}
}

func (s *dataAgentConfigSvc) chatCompletion(ctx *gin.Context, prompt string, sysMsg string, userID string, accountType cenum.AccountType, stream bool) (chan string, chan error, error) {
	req := &imodelfactoryacc.ChatCompletionReq{
		// NOTE: 大模型名称传空，使用默认大模型
		Model:       "",
		Messages:    []imodelfactoryacc.Message{{Role: "system", Content: sysMsg}, {Role: "user", Content: prompt}},
		Stream:      stream,
		UserID:      userID,
		AccountType: accountType,
	}

	return s.modelFactoryAcc.StreamChatCompletion(ctx, req)
}
