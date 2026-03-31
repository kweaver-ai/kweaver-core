package conversationsvc

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

func (svc *conversationSvc) GetHistoryV2(ctx context.Context, id string, historyConfig *daconfvalobj.ConversationHistoryConfig, regenerateUserMsgID string,
	regenerateAssistantMsgID string,
) ([]*comvalobj.LLMMessage, error) {
	var err error

	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("conversation_id", id))

	if historyConfig == nil {
		return nil, errors.New("[GetHistoryV2] history_config is required")
	}

	if err = historyConfig.Strategy.EnumCheck(); err != nil {
		o11y.Error(ctx, fmt.Sprintf("[GetHistoryV2] invalid history strategy, err: %v", err))
		return nil, errors.Wrapf(err, "[GetHistoryV2] invalid history strategy, err: %v", err)
	}

	switch historyConfig.Strategy {
	case cdaenum.HistoryStrategyNone:
		return []*comvalobj.LLMMessage{}, nil
	case cdaenum.HistoryStrategyCount:
		countLimit := constant.DefaultHistoryLimit
		if historyConfig.CountParams != nil && historyConfig.CountParams.CountLimit > 0 {
			countLimit = historyConfig.CountParams.CountLimit
		}

		return svc.GetHistory(ctx, id, countLimit, regenerateUserMsgID, regenerateAssistantMsgID)
	case cdaenum.HistoryStrategyTimeWindow:
		return nil, errors.New("[GetHistoryV2] time_window strategy is not implemented yet")
	case cdaenum.HistoryStrategyToken:
		return nil, errors.New("[GetHistoryV2] token strategy is not implemented yet")
	default:
		return nil, errors.New("[GetHistoryV2] unsupported history strategy")
	}
}

func (svc *conversationSvc) GetHistory(ctx context.Context, id string, limit int, regenerateUserMsgID string,
	regenerateAssistantMsgID string,
) ([]*comvalobj.LLMMessage, error) {
	var err error

	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("conversation_id", id))

	// NOTE: 如果不需要regenerate，则使用DetailWithLimit减少数据库查询
	if regenerateUserMsgID == "" && regenerateAssistantMsgID == "" {
		conversation, err := svc.DetailWithLimit(ctx, id, limit)
		if err != nil {
			o11y.Error(ctx, fmt.Sprintf("[GetHistory] get conversation detail error, id: %s, err: %v", id, err))
			return nil, errors.Wrapf(err, "[GetHistory] get conversation detail error, id: %s, err: %v", id, err)
		}

		history := make([]*comvalobj.LLMMessage, 0)

		for _, msg := range conversation.Messages {
			if msg.Role == cdaenum.MsgRoleAssistant {
				content := conversationmsgvo.AssistantContent{}
				if msg.Content != nil && *msg.Content != "" {
					err := sonic.Unmarshal([]byte(*msg.Content), &content)
					if err != nil {
						o11y.Error(ctx, fmt.Sprintf("[GetHistory] unmarshal assistant content error, id: %s, err: %v", id, err))
						return nil, errors.Wrapf(err, "[GetHistory] unmarshal assistant content error, id: %s, err: %v", id, err)
					}
				}

				if content.FinalAnswer.Answer.Text != "" {
					history = append(history, &comvalobj.LLMMessage{
						Role:    string(msg.Role),
						Content: content.FinalAnswer.Answer.Text,
					})
				} else if len(content.FinalAnswer.SkillProcess) > 0 {
					history = append(history, &comvalobj.LLMMessage{
						Role:    string(msg.Role),
						Content: content.FinalAnswer.SkillProcess[len(content.FinalAnswer.SkillProcess)-1].Text,
					})
				} else {
					other := content.FinalAnswer.AnswerTypeOther
					if otherStr, ok := other.(string); ok {
						history = append(history, &comvalobj.LLMMessage{
							Role:    string(msg.Role),
							Content: otherStr,
						})
					} else {
						byt, _ := sonic.Marshal(other)
						history = append(history, &comvalobj.LLMMessage{
							Role:    string(msg.Role),
							Content: string(byt),
						})
					}
				}
			} else {
				userContent := conversationmsgvo.UserContent{}
				if msg.Content != nil && *msg.Content != "" {
					err := sonic.Unmarshal([]byte(*msg.Content), &userContent)
					if err != nil {
						o11y.Error(ctx, fmt.Sprintf("[GetHistory] unmarshal user content error, id: %s, err: %v", id, err))
						return nil, errors.Wrapf(err, "[GetHistory] unmarshal user content error, id: %s, err: %v", id, err)
					}
				}

				if len(userContent.SelectedFiles) > 0 {
					contextMsg := &comvalobj.LLMMessage{
						Role:    "user",
						Content: buildWorkspaceContextMessage(msg.ConversationID, conversation.CreateBy, userContent.SelectedFiles),
					}
					history = append(history, contextMsg)
				}

				history = append(history, &comvalobj.LLMMessage{
					Role:    string(msg.Role),
					Content: userContent.Text,
				})
			}
		}

		if len(history) == 0 {
			return history, nil
		}

		if limit >= len(history) {
			return history, nil
		}

		return history[len(history)-limit:], nil
	}

	// NOTE: 需要regenerate时，使用原来的全量查询逻辑
	conversation, err := svc.Detail(ctx, id)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[GetHistory] get conversation detail error, id: %s, err: %v", id, err))
		return nil, errors.Wrapf(err, "[GetHistory] get conversation detail error, id: %s, err: %v", id, err)
	}

	history := make([]*comvalobj.LLMMessage, 0)

	userMsgID, assistantMsgID := GetID(ctx, conversation.Messages, regenerateUserMsgID, regenerateAssistantMsgID)
	for _, msg := range conversation.Messages {
		if msg.ID == userMsgID || msg.ID == assistantMsgID {
			break
		}

		if msg.Role == cdaenum.MsgRoleAssistant {
			content := conversationmsgvo.AssistantContent{}
			// NOTE: 不能将空字符串反序列化，否则会报错
			if msg.Content != nil && *msg.Content != "" {
				err := sonic.Unmarshal([]byte(*msg.Content), &content)
				if err != nil {
					o11y.Error(ctx, fmt.Sprintf("[GetHistory] unmarshal assistant content error, id: %s, err: %v", id, err))
					return nil, errors.Wrapf(err, "[GetHistory] unmarshal assistant content error, id: %s, err: %v", id, err)
				}
			}
			// NOTE: 如果最终输出变量是是prompt，则将answer.text作为content
			if content.FinalAnswer.Answer.Text != "" {
				history = append(history, &comvalobj.LLMMessage{
					Role:    string(msg.Role),
					Content: content.FinalAnswer.Answer.Text,
				})
			} else if len(content.FinalAnswer.SkillProcess) > 0 {
				// NOTE:如果最终输出变量是是explore, 如果技能执行过程大于0，则将技能执行过程的最后一个技能的answer.text作为content
				history = append(history, &comvalobj.LLMMessage{
					Role:    string(msg.Role),
					Content: content.FinalAnswer.SkillProcess[len(content.FinalAnswer.SkillProcess)-1].Text,
				})
			} else {
				// NOTE: 如果是other类型，则将other变量序列化为json字符串
				other := content.FinalAnswer.AnswerTypeOther
				if otherStr, ok := other.(string); ok {
					history = append(history, &comvalobj.LLMMessage{
						Role:    string(msg.Role),
						Content: otherStr,
					})
				} else {
					byt, _ := sonic.Marshal(other)
					history = append(history, &comvalobj.LLMMessage{
						Role:    string(msg.Role),
						Content: string(byt),
					})
				}
			}
		} else {
			userContent := conversationmsgvo.UserContent{}
			if msg.Content != nil && *msg.Content != "" {
				err := sonic.Unmarshal([]byte(*msg.Content), &userContent)
				if err != nil {
					o11y.Error(ctx, fmt.Sprintf("[GetHistory] unmarshal user content error, id: %s, err: %v", id, err))
					return nil, errors.Wrapf(err, "[GetHistory] unmarshal user content error, id: %s, err: %v", id, err)
				}
			}

			// NOTE: 如果用户选中了文件，先插入工作区上下文消息
			// 这样可以在历史消息中重建完整的上下文，即使退出重进也不会丢失
			if len(userContent.SelectedFiles) > 0 {
				contextMsg := &comvalobj.LLMMessage{
					Role:    "user",
					Content: buildWorkspaceContextMessage(msg.ConversationID, conversation.CreateBy, userContent.SelectedFiles),
				}
				history = append(history, contextMsg)
			}

			// 然后添加实际的用户查询
			history = append(history, &comvalobj.LLMMessage{
				Role:    string(msg.Role),
				Content: userContent.Text,
			})
		}
	}

	if len(history) == 0 {
		return history, nil
	}

	if limit >= len(history) {
		return history, nil
	}

	return history[len(history)-limit:], nil
}

// NOTE: 如果不是普通对话，则获取用户消息和助手消息的ID
func GetID(ctx context.Context, messages []*dapo.ConversationMsgPO, regenerateUserMsgID string, regenerateAssistantMsgID string) (userMsgID string, assistantMsgID string) {
	if regenerateAssistantMsgID == "" && regenerateUserMsgID == "" {
		return "", ""
	}

	for index, msg := range messages {
		if msg.ID == regenerateUserMsgID {
			userMsgID = msg.ID
			assistantMsgID = messages[index+1].ID

			return userMsgID, assistantMsgID
		}

		if msg.ID == regenerateAssistantMsgID {
			assistantMsgID = msg.ID
			userMsgID = msg.ReplyID

			return userMsgID, assistantMsgID
		}
	}

	return "", ""
}
