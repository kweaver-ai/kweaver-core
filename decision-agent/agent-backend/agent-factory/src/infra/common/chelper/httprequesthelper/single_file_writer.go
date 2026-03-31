package httprequesthelper

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// SingleFileWriter 单文件写入器，保留最近N条日志
type SingleFileWriter struct {
	config     *Config
	file       *os.File
	filePath   string
	mu         sync.Mutex
	lineCount  int
	maxEntries int
	// threshold 超过 maxEntries + threshold 时才触发截断
	threshold int
}

// NewSingleFileWriter 创建单文件写入器
func NewSingleFileWriter(config *Config) (*SingleFileWriter, error) {
	if err := config.EnsureSingleFileDir(); err != nil {
		return nil, fmt.Errorf("failed to create single log dir: %w", err)
	}

	filePath := config.GetSingleFilePath()
	maxEntries := config.SingleFileMaxEntries

	writer := &SingleFileWriter{
		config:     config,
		filePath:   filePath,
		maxEntries: maxEntries,
		threshold:  100, // 超过 maxEntries + 100 时才截断
	}

	// 统计现有文件的行数
	if lineCount, err := writer.countLines(); err == nil {
		writer.lineCount = lineCount
	}

	return writer, nil
}

// Write 写入日志内容
func (w *SingleFileWriter) Write(content string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 打开文件（如果尚未打开）
	if w.file == nil {
		file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open single log file: %w", err)
		}

		w.file = file
	}

	// 写入内容
	_, err := w.file.WriteString(content)
	if err != nil {
		return err
	}

	// 更新行数（假设每次写入一条记录）
	w.lineCount++

	// 检查是否需要截断
	if w.lineCount > w.maxEntries+w.threshold {
		if err := w.truncateLocked(); err != nil {
			// 截断失败不影响正常写入，只记录错误
			fmt.Fprintf(os.Stderr, "failed to truncate single log file: %v\n", err)
		}
	}

	return nil
}

// Close 关闭写入器
func (w *SingleFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}

	return nil
}

// countLines 统计文件行数
func (w *SingleFileWriter) countLines() (int, error) {
	file, err := os.Open(w.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}

		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		count++
	}

	return count, scanner.Err()
}

// truncateLocked 截断文件，保留最新的 maxEntries 条记录
// 注意：调用此方法前必须持有锁
// 【注意】如果当每次写入的不只一行时，可能不是符合预期的（因为是debug时才开启的，用于开发和测试的工具，暂时先这样处理）
func (w *SingleFileWriter) truncateLocked() error {
	// 关闭当前文件
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}

	// 读取所有行
	lines, err := w.readAllLines()
	if err != nil {
		return err
	}

	// 保留最新的 maxEntries 条
	if len(lines) > w.maxEntries {
		lines = lines[len(lines)-w.maxEntries:]
	}

	// 重写文件
	if err := w.rewriteFile(lines); err != nil {
		return err
	}

	// 更新行数
	w.lineCount = len(lines)

	// 重新打开文件
	file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	w.file = file

	return nil
}

// readAllLines 读取文件所有行
func (w *SingleFileWriter) readAllLines() ([]string, error) {
	file, err := os.Open(w.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以处理较长的日志行
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// rewriteFile 重写文件
func (w *SingleFileWriter) rewriteFile(lines []string) error {
	file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}
