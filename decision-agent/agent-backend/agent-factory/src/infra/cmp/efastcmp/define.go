package efastcmp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
)

type EFast struct {
	privateScheme string
	privateHost   string
	privatePort   int

	publicScheme string
	publicHost   string
	publicPort   int

	logger icmp.Logger
}

type EFastConf struct {
	PrivateScheme string
	PrivateHost   string
	PrivatePort   int

	PublicScheme string
	PublicHost   string
	PublicPort   int

	Logger icmp.Logger

	// HttpClient icmp.IHttpClient
}

var _ icmp.IEFast = &EFast{}

func NewEFast(conf *EFastConf) icmp.IEFast {
	return &EFast{
		privateScheme: conf.PrivateScheme,
		privateHost:   conf.PrivateHost,
		privatePort:   conf.PrivatePort,

		publicScheme: conf.PublicScheme,
		publicHost:   conf.PublicHost,
		publicPort:   conf.PublicPort,

		logger: conf.Logger,
		// httpClient: conf.HttpClient,
	}
}
