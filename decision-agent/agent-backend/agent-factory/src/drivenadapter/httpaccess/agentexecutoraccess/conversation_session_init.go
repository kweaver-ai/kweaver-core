package agentexecutoraccess

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccreq"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccres"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/ctype"
// )

// // ConversationSessionInit 只有 v1 接口，没有 v2 接口
// func (ae *agentExecutorHttpAcc) ConversationSessionInit(ctx context.Context, req *agentexecutoraccreq.ConversationSessionInitReq, visitorInfo *ctype.VisitorInfo) (result agentexecutoraccres.ConversationSessionInitResp, err error) {

// 	result = agentexecutoraccres.ConversationSessionInitResp{}

// 	// 1. 构建请求
// 	url := fmt.Sprintf("%s/api/agent-executor/v1/agent/conversation-session/init", ae.privateAddress)

// 	headers := make(map[string]string)
// 	chelper.SetAccountInfoToHeaderMap(headers, visitorInfo.XAccountID, visitorInfo.XAccountType)
// 	headers["x-business-domain"] = visitorInfo.XBusinessDomainID.ToString()

// 	// 2. 发起请求
// 	respCode, respBody, err := ae.restClient.PostNoUnmarshal(ctx, url, headers, req)
// 	if err != nil {
// 		ae.logger.Errorf("failed to initialize conversation session: %v", err)
// 		return
// 	}

// 	// 3. 解析响应
// 	err = json.Unmarshal(respBody, &result)
// 	if err != nil {
// 		ae.logger.Errorf("failed to unmarshal response body: %v", err)
// 		return
// 	}

// 	// 4. 检查响应状态码
// 	if respCode != http.StatusOK {
// 		err = fmt.Errorf("failed to initialize conversation session: %d", respCode)
// 		return
// 	}

// 	return
// }
