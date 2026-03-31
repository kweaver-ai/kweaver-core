package afresvo

import (
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentresperr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
)

type AgentFactoryError struct {
	Description  string `json:"Description"`
	ErrorCode    string `json:"ErrorCode"`
	ErrorDetails string `json:"ErrorDetails"`
	Solution     string `json:"Solution"`
}

func NewAgentFactoryError() *AgentFactoryError {
	return &AgentFactoryError{}
}

func IsAgentFactoryError(data []byte) (afErr AgentFactoryError, isErr bool) {
	if err := sonic.Unmarshal(data, &afErr); err != nil {
		return
	}

	isErr = afErr.ErrorCode != ""

	return
}

func HandleAFErrorForChatProcess(data []byte) (newData []byte, isErr bool) {
	afErr, isErr := IsAgentFactoryError(data)
	if !isErr {
		return
	}

	chatResponse := &agentresp.ChatResp{}
	chatResponse.Message.Ext = &conversationmsgvo.MessageExt{}

	respErr := agentresperr.NewRespError(agentresperr.RespErrorTypeAgentFactory, afErr)

	chatResponse.Message.Ext.Error = respErr

	newData, _ = sonic.Marshal(chatResponse)

	return
}
