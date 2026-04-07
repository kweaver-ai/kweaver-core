package authzhttp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc"
)

type authZHttpAcc struct {
	logger         icmp.Logger
	privateBaseURL string
	publicBaseURL  string
}

var _ iauthzacc.AuthZHttpAcc = &authZHttpAcc{}

func NewAuthZHttpAcc(
	logger icmp.Logger,
) iauthzacc.AuthZHttpAcc {
	// 从配置中获取授权服务的地址
	authZPrivConf := cglobal.GConfig.Authorization.PrivateSvc
	authZPubConf := cglobal.GConfig.Authorization.PublicSvc

	privateBaseURL := cutil.GetHTTPAccess(authZPrivConf.Host, authZPrivConf.Port, authZPrivConf.Protocol)

	publicBaseURL := cutil.GetHTTPAccess(authZPubConf.Host, authZPubConf.Port, authZPubConf.Protocol)

	authZImpl := &authZHttpAcc{
		logger:         logger,
		privateBaseURL: privateBaseURL,
		publicBaseURL:  publicBaseURL,
	}

	return authZImpl
}
