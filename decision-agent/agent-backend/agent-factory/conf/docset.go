package conf

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/cconf"
)

type DocsetConf struct {
	PublicSvc  cconf.SvcConf `yaml:"public_svc"`
	PrivateSvc cconf.SvcConf `yaml:"private_svc"`
}
