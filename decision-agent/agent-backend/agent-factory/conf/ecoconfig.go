package conf

import "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"

type EcoConfigConf struct {
	PublicSvc  cconf.SvcConf `yaml:"public_svc"`
	PrivateSvc cconf.SvcConf `yaml:"private_svc"`
}
