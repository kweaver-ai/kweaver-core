package agentsvc

import "github.com/bytedance/sonic"

func (agentSvc *agentSvc) originalChatResp(data []byte) (originalRes map[string]interface{}, err error) {
	err = sonic.Unmarshal(data, &originalRes)

	return
}
