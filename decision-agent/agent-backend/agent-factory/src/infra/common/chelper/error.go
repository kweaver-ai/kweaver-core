package chelper

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
)

//go:noinline
func RecordErrLogWithPos(logger icmp.Logger, err error, positions ...string) {
	sb := strings.Builder{}
	for _, pos := range positions {
		sb.WriteString("[")
		sb.WriteString(pos)
		sb.WriteString("]")
	}

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		fn := runtime.FuncForPC(pc)
		fnName := filepath.Base(fn.Name())

		msg := fmt.Errorf("%v error: %v, loc: %s:%d, func:%s\n", sb.String(), err, file, line, fnName)
		if cenvhelper.IsAaronLocalDev() {
			log.Println(msg)
		}

		logger.Errorln(msg)
	} else {
		msg := fmt.Errorf("%v error: %v\n", sb.String(), err)

		if cenvhelper.IsAaronLocalDev() {
			log.Println(msg)
		}

		logger.Errorln(msg)
	}
}
