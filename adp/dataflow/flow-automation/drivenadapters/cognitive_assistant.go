package drivenadapters

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

var (
	cognitiveAssistantOnce sync.Once
	cognitiveAssistant     CognitiveAssistant
)

func NewCognitiveAssistant() CognitiveAssistant {
	cognitiveAssistantOnce.Do(func() {

		config := common.NewConfig()
		cognitiveAssistant = &CognitiveAssistantImpl{
			baseURL:    fmt.Sprintf("http://%s:%d", config.CognitiveAssistant.Host, config.CognitiveAssistant.Port),
			httpClient: otelHttp.NewOtelHttpClient(),
		}
	})
	return cognitiveAssistant
}

type CognitiveAssistant interface {
	GetCustomPrompts(ctx context.Context) (result interface{}, err error)
	CustomPrompt(ctx context.Context, serviceID string, prompt string) (string, error)
	CustomPromptWithFile(ctx context.Context, serviceID string, DocID string) (result string, err error)
}

type CognitiveAssistantImpl struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

func (c *CognitiveAssistantImpl) GetCustomPrompts(ctx context.Context) (result interface{}, err error) {
	target := fmt.Sprintf("%s/private/cognitive-assistant/v1/custom-prompt", c.baseURL)
	_, result, err = c.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetCustomPrompts failed: %v, url: %v", err, target)
	}

	return
}

type CustomPromptResult struct {
	ResultID string `json:"result_id"`
	Results  string `json:"results"`
	Status   string `json:"status"`
}

func (c *CognitiveAssistantImpl) CustomPrompt(ctx context.Context, serviceID string, content string) (result string, err error) {

	target := fmt.Sprintf("%s/private/cognitive-assistant/v1/custom-prompt", c.baseURL)

	// 截取前2500字
	if len(content) > 2500 {
		content = content[:2500]
	}

	body := map[string]interface{}{
		"id":      serviceID,
		"content": content,
	}

	fmt.Printf("CustomPrompt service: %v, content: %v\n", serviceID, content)

	data, _ := json.Marshal(body)

	resp, err := http.Post(target, "application/json;charset=UTF-8", bytes.NewBuffer(data))

	if err != nil {
		traceLog.WithContext(ctx).Warnf("CustomPrompt failed: %v, url: %v", err, target)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.NewIError(errors.InternalError, "", map[string]interface{}{
			"prompt_service_id": serviceID,
			"content":           content,
		})
		traceLog.WithContext(ctx).Warnf("CustomPrompt failed: %v, url: %v", err, target)
		return
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		var res CustomPromptResult
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimPrefix(line, "data:")
			err = json.Unmarshal([]byte(line), &res)
			if err != nil {
				traceLog.WithContext(ctx).Warnf("CustomPrompt parse result error: %v, line: %v", err, string(line))
				return
			}

			if res.Status == "completed" {
				return res.Results, nil
			} else if res.Status == "processing" {
				continue
			} else {
				return "", errors.NewIError(errors.InternalError, "", map[string]interface{}{
					"prompt_service_id": serviceID,
					"content":           content,
				})
			}
		}
	}

	if err = scanner.Err(); err != nil {
		traceLog.WithContext(ctx).Warnf("CustomPrompt scan error: %v", err)
		return
	}

	return
}

func (c *CognitiveAssistantImpl) CustomPromptWithFile(ctx context.Context, serviceID string, DocID string) (result string, err error) {
	content, err := NewEfast().GetDocSetSubdocContent(ctx, DocSetSubdocParams{
		DocID: DocID,
		Type:  "full_text",
	}, 0, time.Second, 10*1024*1024)
	if err != nil {
		return
	}

	if len(content) == 0 {
		return content, nil
	}

	result, err = c.CustomPrompt(ctx, serviceID, content)

	return
}
