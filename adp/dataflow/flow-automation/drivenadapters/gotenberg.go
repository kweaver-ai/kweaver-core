package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

type GotenbergConvertRequest struct {
	FileName        string
	File            io.Reader
	WebhookURL      string
	WebhookErrorURL string
	WebhookHeaders  map[string]string
}

type PDFConverter interface {
	ConvertToPDF(ctx context.Context, req *GotenbergConvertRequest) error
}

type gotenberg struct {
	baseURL string
	client  *http.Client
}

func NewGotenberg() PDFConverter {
	config := common.NewConfig()
	return newGotenberg(
		fmt.Sprintf("http://%s:%v", config.DocConvert.Host, config.DocConvert.GotenbergPort),
		NewOtelRawHTTPClient(),
	)
}

func newGotenberg(baseURL string, client *http.Client) PDFConverter {
	return &gotenberg{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

func (g *gotenberg) ConvertToPDF(ctx context.Context, req *GotenbergConvertRequest) error {
	if req == nil {
		return fmt.Errorf("gotenberg request is nil")
	}
	if req.File == nil {
		return fmt.Errorf("gotenberg request file is nil")
	}

	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	go func() {
		defer pipeWriter.Close()
		defer writer.Close()

		part, err := writer.CreateFormFile("files", req.FileName)
		if err != nil {
			_ = pipeWriter.CloseWithError(err)
			return
		}

		if _, err = io.Copy(part, req.File); err != nil {
			_ = pipeWriter.CloseWithError(err)
		}
	}()

	target := fmt.Sprintf("%s/forms/libreoffice/convert", g.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, pipeReader)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Gotenberg-Webhook-Url", req.WebhookURL)
	httpReq.Header.Set("Gotenberg-Webhook-Error-Url", req.WebhookErrorURL)
	httpReq.Header.Set("Gotenberg-Webhook-Method", http.MethodPost)
	httpReq.Header.Set("Gotenberg-Webhook-Error-Method", http.MethodPost)
	if len(req.WebhookHeaders) > 0 {
		headerBytes, err := json.Marshal(req.WebhookHeaders)
		if err != nil {
			return err
		}
		httpReq.Header.Set("Gotenberg-Webhook-Extra-Http-Headers", string(headerBytes))
	}

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		traceLog.WithContext(ctx).Warnf("[Gotenberg.ConvertToPDF] unexpected status: %d, body: %s", resp.StatusCode, string(body))
		return fmt.Errorf("gotenberg convert failed, status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}
