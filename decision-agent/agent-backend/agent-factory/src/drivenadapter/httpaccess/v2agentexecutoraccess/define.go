package v2agentexecutoraccess

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/httpclient"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iv2agentexecutorhttp"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type v2AgentExecutorHttpAcc struct {
	logger            icmp.Logger
	client            rest.HTTPClient
	agentExecutorConf *conf.AgentExecutorConf
	streamClient      httpclient.HTTPClient

	privateAddress string
}

var _ iv2agentexecutorhttp.IV2AgentExecutor = &v2AgentExecutorHttpAcc{}

func NewV2AgentExecutorHttpAcc(logger icmp.Logger, agentExecutorConf *conf.AgentExecutorConf, client rest.HTTPClient, streamClient httpclient.HTTPClient) iv2agentexecutorhttp.IV2AgentExecutor {
	impl := &v2AgentExecutorHttpAcc{
		logger:            logger,
		client:            client,
		agentExecutorConf: agentExecutorConf,
		streamClient:      streamClient,
		privateAddress:    cutil.GetHTTPAccess(agentExecutorConf.PrivateSvc.Host, agentExecutorConf.PrivateSvc.Port, agentExecutorConf.PrivateSvc.Protocol),
	}

	return impl
}
