// package drivenadapters 新纯文本提取器，用于从文件中提取纯文本
package drivenadapters

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

type PlainTextExtractor interface {
	ExtractPlainText(ctx context.Context, filename string, reader io.Reader) (string, error)
}

type tikaPlainTextExtractor struct {
	baseURL string
	client  *http.Client
}

func NewTikaPlainTextExtractor() PlainTextExtractor {
	config := common.NewConfig()
	return &tikaPlainTextExtractor{
		baseURL: fmt.Sprintf("http://%s:%v", config.DocConvert.Host, config.DocConvert.TikaPort),
		client:  NewOtelRawHTTPClient(),
	}
}

func (t *tikaPlainTextExtractor) ExtractPlainText(ctx context.Context, filename string, reader io.Reader) (string, error) {
	target := strings.TrimRight(t.baseURL, "/") + "/tika"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, reader)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	req.Header.Set("Accept", "text/plain")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		traceLog.WithContext(ctx).Warnf("[TikaPlainTextExtractor.ExtractPlainText] unexpected status: %d, body: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("tika extract failed, status=%d body=%s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
