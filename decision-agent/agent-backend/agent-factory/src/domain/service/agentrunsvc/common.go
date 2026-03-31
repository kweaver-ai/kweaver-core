package agentsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo/daresvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// NOTE: 将agent配置转换为agent call请求配置，支持用户传入llm配置和数据源配置代替agent配置中的llm配置和数据源配置
func AgentConfig2AgentCallConfig(ctx context.Context, agentConfig *daconfvalobj.Config, req *agentreq.ChatReq) agentexecutordto.Config {
	agentCallConfig := agentexecutordto.Config{
		Config:         *agentConfig,
		AgentID:        req.AgentID,
		ConversationID: req.ConversationID,
		SessionID:      req.AgentRunID,
	}

	if agentConfig.Skill == nil {
		agentCallConfig.Skill = &skillvalobj.Skill{
			Tools:  []*skillvalobj.SkillTool{},
			Agents: []*skillvalobj.SkillAgent{},
			MCPs:   []*skillvalobj.SkillMCP{},
		}
	}

	if agentConfig.PreDolphin == nil {
		agentCallConfig.PreDolphin = []*daconfvalobj.DolphinTpl{}
	}

	if agentConfig.PostDolphin == nil {
		agentCallConfig.PostDolphin = []*daconfvalobj.DolphinTpl{}
	}

	return agentCallConfig
}

// NOTE: 将agent配置转换为agent call请求配置，支持用户传入llm配置和数据源配置代替agent配置中的llm配置和数据源配置
func AgentConfig2AgentCallConfigDebug(ctx context.Context, agentConfig *daconfvalobj.Config, req *agentreq.DebugReq) agentexecutordto.Config {
	agentCallConfig := agentexecutordto.Config{
		Config:    *agentConfig,
		AgentID:   req.AgentID,
		SessionID: req.AgentRunID,
	}

	if agentConfig.Skill == nil {
		agentCallConfig.Skill = &skillvalobj.Skill{
			Tools:  []*skillvalobj.SkillTool{},
			Agents: []*skillvalobj.SkillAgent{},
			MCPs:   []*skillvalobj.SkillMCP{},
		}
	}

	return agentCallConfig
}

// NOTE: 生成会话消息
func GenerateAssistantMsg(ctx context.Context, req *agentreq.ChatReq, result *daresvo.DataAgentRes) (conversationmsgvo.Message, error) {
	return conversationmsgvo.Message{}, nil
}

// NOTE: 计算TTFT，单位ms
func CalculateTTFT(startTime int64, progresses []*agentrespvo.Progress, callType constant.CallType) int64 {
	switch callType {
	case constant.Chat, constant.DebugChat:
		return calculateTTFTForChat(startTime, progresses)
	default:
		return 0
	}
}

/*
*遍历progress数组中的元素进行判断：

如果是stage是llm:
判断 answer || thinking 这两个字段，任何一个有值前端会立即显示

如果是stage是skill:
只要出现，前端会立即显示这个工具。

但是前端会过滤一些工具。过滤规则如下：
规则一：
skill_info.name字段值是下面的工具，前端不会显示。
search_memory  _date  build__memory

规则二：
skill_info.args，工具的参数中，如果含有name是action并且value是show_ds的参数，此工具前端不显示
*返回值：
返回第一个有值的stage的当前时间戳- startTime
*/
func calculateTTFTForChat(startTime int64, progresses []*agentrespvo.Progress) int64 {
	if len(progresses) == 0 {
		return 0
	}

	for _, progress := range progresses {
		if progress.Stage == "llm" {
			// NOTE: 如果answer或think有值，则返回当前时间戳- startTime
			if answer, ok := progress.Answer.(string); !ok || answer == "" {
				return cutil.GetCurrentMSTimestamp() - startTime
			}

			if think, ok := progress.Think.(string); !ok || think == "" {
				return cutil.GetCurrentMSTimestamp() - startTime
			}

			return 0
		} else if progress.Stage == "skill" {
			// NOTE: 如果skill_info.name是search_memory, _date, build__memory，则continue,看下一个progress
			if progress.SkillInfo.Name == "search_memory" || progress.SkillInfo.Name == "_date" || progress.SkillInfo.Name == "build__memory" {
				continue
			}
			// NOTE: 如果skill_info.args含有name是action并且value是show_ds，则跳过,看下一个progress
			flag := false

			for _, arg := range progress.SkillInfo.Args {
				if arg.Name == "action" && arg.Type == "string" {
					if value, ok := arg.Value.(string); ok && value == "show_ds" {
						flag = true
						break
					}
				}
			}

			if flag {
				continue
			}

			return cutil.GetCurrentMSTimestamp() - startTime
		}
	}

	return 0
}
