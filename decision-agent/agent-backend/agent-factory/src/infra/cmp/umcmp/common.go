package umcmp

import (
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func (u *Um) getPrivateURLPrefix() string {
	protocol := u.umConf.Protocol
	if protocol == "" {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s:%d/api/user-management", protocol, cutil.ParseHost(u.umConf.Host), u.umConf.Port)
}
