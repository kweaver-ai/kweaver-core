package httprequesthelper

import (
	"fmt"
	"os"
	"sync"
)

// Writer 日志写入器接口
type Writer interface {
	Write(content string, userID string) error
	Close() error
}

// ConsoleWriter 控制台写入器
type ConsoleWriter struct{}

// NewConsoleWriter 创建控制台写入器
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

func (w *ConsoleWriter) Write(content string, userID string) error {
	fmt.Print(content)
	return nil
}

func (w *ConsoleWriter) Close() error {
	return nil
}

// FileWriter 文件写入器
type FileWriter struct {
	config   *Config
	file     *os.File
	filePath string
	mu       sync.Mutex
}

// NewFileWriter 创建文件写入器
func NewFileWriter(config *Config) (*FileWriter, error) {
	if err := config.EnsureLogDir(); err != nil {
		return nil, fmt.Errorf("failed to create log dir: %w", err)
	}

	return &FileWriter{
		config: config,
	}, nil
}

func (w *FileWriter) Write(content string, userID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 检查是否需要切换文件（按日期或用户ID）
	currentPath := w.config.GetLogFilePath(userID)
	if w.file == nil || w.filePath != currentPath {
		if w.file != nil {
			_ = w.file.Close()
		}

		file, err := os.OpenFile(currentPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		w.file = file
		w.filePath = currentPath
	}

	_, err := w.file.WriteString(content)

	return err
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}

	return nil
}

// MultiWriter 多写入器（同时写入多个目标）
type MultiWriter struct {
	writers []Writer
}

// NewMultiWriter 创建多写入器
func NewMultiWriter(writers ...Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

func (w *MultiWriter) Write(content string, userID string) error {
	var lastErr error

	for _, writer := range w.writers {
		if err := writer.Write(content, userID); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func (w *MultiWriter) Close() error {
	var lastErr error

	for _, writer := range w.writers {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
