package drivenadapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

// SpeechModel 音频转文字模型接口
type SpeechModel interface {
	// AudioTransfer 音频转文字
	AudioTransfer(ctx context.Context, filename, wbbHook string, content *[]byte) (string, error)
	// GetAudioTransferResult 获取音频转文字结果
	GetAudioTransferResult(ctx context.Context, taskID string) (int, map[string]interface{}, error)
	// CheckSpeechModel 检查SpeechModel服务是否可用
	CheckSpeechModel(ctx context.Context) error
}

type speechModel struct {
	privateURL string
	httpClient otelHttp.HTTPClient
}

var (
	sOnce sync.Once
	sM    SpeechModel
)

// NewSpeechModel SpeechModel实例化
func NewSpeechModel() SpeechModel {
	sOnce.Do(func() {
		config := common.NewConfig()
		sM = &speechModel{
			privateURL: fmt.Sprintf("%s:%v", getSpeechHost(config), config.SpeechModel.PrivatePort),
			httpClient: NewOtelHTTPClient(),
		}
	})

	return sM
}

// AudioTransfer 音频转文字
func (s *speechModel) AudioTransfer(ctx context.Context, filename, wbbHook string, content *[]byte) (string, error) {
	target := fmt.Sprintf("%s/api/speech-model/v1/convert/task", s.privateURL)
	var body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加二进制数据字段
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[AudioTransfer] Error creating form field: %s", err.Error())
		return "", err
	}
	_, err = part.Write(*content)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[AudioTransfer] Error writing binary data: %s", err.Error())
		return "", err
	}

	_ = writer.WriteField("web_hook", wbbHook)
	// 必须在写入完毕后关闭 writer
	_ = writer.Close()

	header := map[string]string{"Content-Type": writer.FormDataContentType()}
	payload := body.Bytes()
	_, resp, err := s.httpClient.Request(ctx, target, http.MethodPost, header, &payload)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[AudioTransfer] Request failed, detail: %s", err.Error())
		return "", err
	}

	var result map[string]string
	_ = json.Unmarshal(resp, &result)
	return result["id"], nil
}

// GetAudioTransferResult 获取音频转文字结果
func (s *speechModel) GetAudioTransferResult(ctx context.Context, taskID string) (int, map[string]interface{}, error) {
	target := fmt.Sprintf("%s/api/speech-model/v1/convert/result/%s", s.privateURL, taskID)
	header := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	respCode, resp, err := s.httpClient.Get(ctx, target, header)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[AudioTransfer] Request failed, detail: %s", err.Error())
		return respCode, nil, err
	}

	if respCode == http.StatusAccepted {
		return respCode, nil, nil
	}

	return respCode, resp.(map[string]interface{}), nil
}

// CheckSpeechModel 检查SpeechModel服务是否可用
func (s *speechModel) CheckSpeechModel(ctx context.Context) error {
	target := fmt.Sprintf("%s/api/speech-model/v1/health/ready", s.privateURL)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, _, err := s.httpClient.Get(ctx, target, map[string]string{})

	return err
}

func getSpeechHost(config *common.Config) (host string) {
	privateHost := config.SpeechModel.PrivateHost
	if len(privateHost) >= 4 && privateHost[:4] == "http" {
		host = privateHost
	} else {
		host = "http://" + privateHost
	}

	return host
}
