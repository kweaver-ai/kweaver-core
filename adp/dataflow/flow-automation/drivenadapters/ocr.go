package drivenadapters

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/ocr.go -destination ../tests/mock_drivenadapters/ocr_mock.go

// Ocr method interface
type Ocr interface {
	// 提交任务
	SubmitTask(ctx context.Context, filename string, con *[]byte) (*SubmitTaskRes, error)

	// 查询任务结果详情
	GetResultDetails(ctx context.Context, taskID string) (interface{}, error)

	GetOCRPredict(ctx context.Context, image *string) (interface{}, error)

	BatchDeleteTask(ctx context.Context, taskIDs []string) (interface{}, error)

	RecognizeText(ctx context.Context, fileUrl, fileName string) (string, error)
}

// SubmitTaskRes 提交结果返回
type SubmitTaskRes struct {
	Code int64  `json:"code"`
	Data string `json:"data"`
}

type ocr struct {
	t4thURL    string
	ocrBaseURL string
	httpClient otelHttp.HTTPClient
	rawClient  *http.Client
}

var (
	oOnce sync.Once
	o     Ocr
)

// NewOcr 创建ocr服务
func NewOcr() Ocr {
	oOnce.Do(func() {
		config := common.NewConfig()
		o = &ocr{
			t4thURL:    fmt.Sprintf("%s://%s:%v", config.T4th.Protocol, config.T4th.Host, config.T4th.Port),
			ocrBaseURL: fmt.Sprintf("http://%s:%v", config.OCR.PrivateHost, config.OCR.PrivatePort),
			httpClient: NewOtelHTTPClient(),
			rawClient: &http.Client{
				Transport: otelhttp.NewTransport(&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}),
			},
		}
	})
	return o
}

// 提交任务
func (o *ocr) SubmitTask(ctx context.Context, filename string, con *[]byte) (*SubmitTaskRes, error) {
	target := fmt.Sprintf("%s/service/submitTask?taskType=1&uri=/lab/ocr/predict/general&scene=handwriting", o.t4thURL)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// 添加二进制数据字段
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("Error creating form field: %s", err.Error())
		return nil, err
	}
	_, err = part.Write(*con)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("Error writing binary data: %s", err.Error())
		return nil, err
	}

	// 必须在写入完毕后关闭 writer
	writer.Close()
	// 创建 HTTP 请求
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target, &body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("Error creating request: %s", err.Error())
		return nil, err
	}

	// 设置请求头部信息，包括 Content-Type
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Accept", " */*")

	// 发送请求
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("Error sending request: %s", err.Error())
		return nil, err
	}

	defer func() {
		closeErr := response.Body.Close()
		if closeErr != nil {
			traceLog.WithContext(ctx).Warnln(closeErr)
		}
	}()
	resbody, err := io.ReadAll(response.Body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("Error reading response: %s", err.Error())
		return nil, err
	}
	respCode := response.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = errors.ExHTTPError{
			Body:   string(resbody),
			Status: respCode,
		}
		return nil, err
	}

	// 处理响应
	var submitRes SubmitTaskRes
	err = json.Unmarshal(resbody, &submitRes)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("json parse error:", err)
		return nil, err
	}

	return &submitRes, nil
}

// 查询任务结果详情
func (o *ocr) GetResultDetails(ctx context.Context, taskID string) (interface{}, error) {
	target := fmt.Sprintf("%s/service/getResultDetail?taskNo=%s", o.t4thURL, taskID)

	_, res, err := o.httpClient.Get(ctx, target, map[string]string{})

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetResultDetails failed: %v, url: %v", err, target)
		return nil, err
	}

	return res, nil
}

// 批量删除任务
func (o *ocr) BatchDeleteTask(ctx context.Context, taskIDs []string) (interface{}, error) {
	target := fmt.Sprintf("%s/service/batchDeleteTask", o.t4thURL)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	_, res, err := o.httpClient.Post(ctx, target, headers, taskIDs)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("BatchDeleteTask failed: %v, url: %v", err, target)
		return nil, err
	}

	return res, nil
}

// 普通OCR识别
func (o *ocr) GetOCRPredict(ctx context.Context, image *string) (interface{}, error) {
	target := fmt.Sprintf("%s/lab/ocr/predict/general", o.t4thURL)
	body := map[string]interface{}{"scene": "handwriting", "image": *image, "vis_flag": false}
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	_, res, err := o.httpClient.Post(ctx, target, headers, body)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetOCRPredict failed: %v, url: %v", err, target)
		return nil, err
	}

	return res, nil
}

func (o *ocr) RecognizeText(ctx context.Context, fileUrl, fileName string) (string, error) {

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		var err error
		defer func() {
			writer.Close()
			if err != nil {
				pw.CloseWithError(err)
			} else {
				pw.Close()
			}
		}()

		resp, err := o.rawClient.Get(fileUrl)

		if err != nil {
			return
		}

		defer resp.Body.Close()
		part, err := writer.CreateFormFile("image_file", fileName)

		if err != nil {
			return
		}

		_, err = io.Copy(part, resp.Body)

		if err != nil {
			return
		}

		writer.WriteField("use_det", "true")
		writer.WriteField("use_cls", "true")
		writer.WriteField("use_rec", "true")
	}()

	target := fmt.Sprintf("%s/ocr", o.ocrBaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", target, pr)

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := o.rawClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", string(data))
	}

	parts := make([][]any, 0)
	err = json.Unmarshal(data, &parts)

	if err != nil {
		return "", err
	}

	text := ""

	for _, part := range parts {
		if len(part) < 2 {
			continue
		}

		if content, ok := part[1].(string); ok {
			if text != "" {
				text += " "
			}
			text += content
		}
	}

	return text, nil
}
