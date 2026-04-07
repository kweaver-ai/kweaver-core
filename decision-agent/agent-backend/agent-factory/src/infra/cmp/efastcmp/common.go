package efastcmp

import (
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func (e *EFast) getUrlPrefix() string {
	return fmt.Sprintf("%s://%s:%d/api/efast", e.privateScheme, cutil.ParseHost(e.privateHost), e.privatePort)
}

func (e *EFast) getPublicUrlPrefix() string {
	return fmt.Sprintf("%s://%s:%d/api/efast", e.publicScheme, cutil.ParseHost(e.publicHost), e.publicPort)
}
