package umhttpaccess

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/cconf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc"
)

type umHttpAcc struct {
	um     *umcmp.Um
	logger icmp.Logger
}

var _ iumacc.UmHttpAcc = &umHttpAcc{}

func NewUmHttpAcc(
	logger icmp.Logger,
) iumacc.UmHttpAcc {
	umConf := &cconf.UserMgntCfg{
		Host:     cglobal.GConfig.Hydra.UserMgnt.Host,
		Port:     cglobal.GConfig.Hydra.UserMgnt.Port,
		Protocol: "http",
	}

	_um := umcmp.NewUmCmp(
		umConf,
		logger,
	)

	umImpl := &umHttpAcc{
		um:     _um,
		logger: logger,
	}

	return umImpl
}
