package panichelper

import (
	"fmt"
	"log"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
)

func Recovery(logger icmp.Logger) {
	if err := recover(); err != nil {
		// 1、记录日志
		panicLogMsg := PanicTraceErrLog(err)

		if cenvhelper.IsDebugMode() {
			log.Println(panicLogMsg)
		}

		logger.Errorln(panicLogMsg)

		return
	}
}

func RecoveryAndSetErr(logger icmp.Logger, err *error) {
	if r := recover(); r != nil {
		// 1、记录日志
		panicLogMsg := PanicTraceErrLog(r)
		logger.Errorln(panicLogMsg)

		// 2、设置错误
		if e, ok := r.(error); ok {
			*err = e
		} else {
			*err = fmt.Errorf("%v", r)
		}

		return
	}
}
