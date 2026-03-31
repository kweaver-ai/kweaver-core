package cutil

import "time"

// WaitForGoroutine 等待goroutine执行完成（单元测试中使用）
func WaitForGoroutine(ds ...time.Duration) {
	if len(ds) == 0 {
		//nolint:gomnd
		ds = append(ds, time.Millisecond*100)
	}

	time.Sleep(ds[0])
}
