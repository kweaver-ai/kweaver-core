package grhelper

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/panichelper"
)

// GoSafe 安全地执行一个 goroutine，会自动捕获和处理 panic
func GoSafe(logger icmp.Logger, f func() (err error)) {
	go func() {
		//defer func() {
		//	if r := recover(); r != nil {
		//		logger.Errorf("goroutine panic: %v\nstack:\n%s\n", r, debug.Stack())
		//	}
		//}()
		defer panichelper.Recovery(logger)

		err1 := f()
		if err1 != nil {
			logger.Errorf("[GoSafe] err: %v", err1)
		}
	}()
}
