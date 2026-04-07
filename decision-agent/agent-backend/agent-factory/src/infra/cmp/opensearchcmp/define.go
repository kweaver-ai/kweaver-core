package opensearchcmp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/opensearch-project/opensearch-go"
)

type OpsCmp struct {
	address  string
	username string
	password string

	logger icmp.Logger

	client *opensearch.Client
}

type OpsCmpConf struct {
	Address  string
	Username string
	Password string

	Logger icmp.Logger
}

var _ icmp.IOpsCmp = &OpsCmp{}

func NewOpsCmp(conf *OpsCmpConf) (cmp icmp.IOpsCmp, err error) {
	o := &OpsCmp{
		address:  conf.Address,
		username: conf.Username,
		password: conf.Password,

		logger: conf.Logger,
	}

	err = o.newClient()
	if err != nil {
		return
	}

	cmp = o

	return
}
