package chelper

import (
	"os"
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
)

var appendWriteToFileLock sync.Mutex

// AppendWriteToFile 追加写入文件
// 【注意】：性能可能不好，仅用于本地调试等使用
func AppendWriteToFile(filePath, text string) (err error) {
	if !cenvhelper.IsAaronLocalDev() {
		// 非本地开发环境下直接返回
		return
	}

	appendWriteToFileLock.Lock()
	defer appendWriteToFileLock.Unlock()

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.WriteString(text)

	return
}
