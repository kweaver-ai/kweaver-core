package panichelper

import (
	"bytes"
	"fmt"
	"runtime"
)

// PanicTrace 返回panic的堆栈信息
func PanicTrace() []byte {
	def := []byte("An unknown error")
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")

	line := []byte("\n")
	stack := make([]byte, (2<<10)*2) // 4KB
	length := runtime.Stack(stack, false)

	start := bytes.Index(stack, s)
	if start == -1 || start > len(stack) || length > len(stack) {
		return def
	}

	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1

	if start >= len(stack) {
		return def
	}

	stack = stack[start:]
	end := bytes.LastIndex(stack, line)

	if end > len(stack) {
		return def
	}

	if end != -1 {
		stack = stack[:end]
	}

	end = bytes.Index(stack, e)
	if end > len(stack) {
		return def
	}

	if end != -1 {
		stack = stack[:end]
	}

	stack = bytes.TrimRight(stack, "\n")

	return stack
}

// PanicTraceErrLog 返回panic的堆栈信息，用于日志输出
func PanicTraceErrLog(err interface{}) (logMessage string) {
	logMessage = fmt.Sprintf("Panic: %s\n", err)
	logMessage += fmt.Sprintf("At stack: %s\n", string(PanicTrace()))

	return
}
