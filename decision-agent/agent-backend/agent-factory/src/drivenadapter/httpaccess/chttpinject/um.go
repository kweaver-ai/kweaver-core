package chttpinject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/umhttpaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	umOnce sync.Once
	umImpl iumacc.UmHttpAcc
)

func NewUmHttpAcc() iumacc.UmHttpAcc {
	umOnce.Do(func() {
		umImpl = umhttpaccess.NewUmHttpAcc(
			logger.GetLogger(),
		)
	})

	return umImpl
}
