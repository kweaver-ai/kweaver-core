package proxy

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

const (
	defaultBufferSize = 4096
)

// StreamProcessor 流式处理器
type StreamProcessor struct {
	logger interfaces.Logger
}

// NewStreamProcessor 创建流式处理器
func NewStreamProcessor(logger interfaces.Logger) *StreamProcessor {
	return &StreamProcessor{
		logger: logger,
	}
}

// ProcessSSE 处理SSE流
func (p *StreamProcessor) ProcessSSE(ctx context.Context, reader io.Reader, writer io.Writer, isSSE bool) error {
	bufReader := bufio.NewReader(reader)
	var (
		// 记录最后一行是否是结束标记
		receivedDone bool
		// 记录是否接收到任何数据
		receivedData bool
		buffer       bytes.Buffer
	)

	if !isSSE {
		if _, err := buffer.ReadFrom(bufReader); err != nil {
			return fmt.Errorf("read non sse data failed: %w", err)
		}
		rawContent := buffer.String()
		// 将整个响应内容作为单个SSE数据返回
		// 确保JSON等格式数据不被分割，保持完整性
		trimmedContent := strings.TrimRight(rawContent, "\n")
		// 将多行内容压缩为单行，保持JSON完整性
		singleLineContent := strings.ReplaceAll(trimmedContent, "\n", "")
		if _, err := fmt.Fprintf(writer, "data: %s\n\n", singleLineContent); err != nil {
			return err
		}
		receivedData = true
	} else {
		// 原有流式处理逻辑
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				line, err := bufReader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						return p.checkReceivedData(writer, receivedData, receivedDone)
					}
					return fmt.Errorf("read sse data failed: %w", err)
				}

				if len(line) == 0 {
					continue
				}
				receivedData = true
				// 如果包含[data: [DONE]]，则标记为完成
				if strings.Contains(string(line), "data: [DONE]") {
					lineStr := strings.TrimRight(string(line), "\n")
					if strings.TrimSpace(lineStr) == "data: [DONE]" {
						receivedDone = true
					}
				}
				if _, err = fmt.Fprintf(writer, "%s", line); err != nil {
					return err
				}
			}
		}
	}

	if flusher, ok := writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return p.checkReceivedData(writer, receivedData, receivedDone)
}

// ProcessHTTPStream 处理HTTP流
func (p *StreamProcessor) ProcessHTTPStream(ctx context.Context, reader io.Reader, writer io.Writer) error {
	buffer := make([]byte, defaultBufferSize) // 4KB 缓冲区
	// 记录是否接收到任何数据
	var receivedData bool
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := reader.Read(buffer)
			if n > 0 {
				receivedData = true
				if _, writeErr := writer.Write(buffer[:n]); writeErr != nil {
					return writeErr
				}

				if flusher, ok := writer.(http.Flusher); ok {
					flusher.Flush()
				}
			}
			if err == nil {
				continue
			}
			if err != io.EOF {
				return fmt.Errorf("read http stream data failed: %w", err)
			}
			return p.checkReceivedData(writer, receivedData, true)
		}
	}
}

// 检查接收数据
func (p *StreamProcessor) checkReceivedData(writer io.Writer, receivedData, receivedDone bool) (err error) {
	if !receivedData {
		// 发送错误消息
		err = errors.New("server does not support streaming or not data")
		return
	}
	// 流结束时，只有在没有接收到[DONE]标记时才发送
	if receivedDone {
		return
	}
	if _, err := fmt.Fprintf(writer, "%s\n\n", "data: [DONE]"); err != nil {
		return fmt.Errorf("write done message failed: %w", err)
	}
	if flusher, ok := writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return
}
