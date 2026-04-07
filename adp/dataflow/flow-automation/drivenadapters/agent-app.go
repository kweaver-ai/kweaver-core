package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

type ChatCompletionReq struct {
	AgentID      string         `json:"agent_id"`
	Stream       bool           `json:"stream"`
	Query        string         `json:"query"`
	CustomQuerys map[string]any `json:"custom_querys,omitempty"`
	BizDomainID  string         `json:"-"`
}

type ChatCompletionAnswer struct {
	Text string `json:"text"`
}

type ChatCompletionSkillProcessItem struct {
	AgentName string `json:"agent_name"`
	Text      string `json:"text"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	Thinking  string `json:"thinking"`
}

type ChatCompletionFinalAnswer struct {
	Answer       ChatCompletionAnswer              `json:"answer"`
	Thinking     string                            `json:"thinking"`
	SkillProcess []*ChatCompletionSkillProcessItem `json:"skill_process"`
}

type ChatCompletionContent struct {
	FinalAnswer ChatCompletionFinalAnswer `json:"final_answer"`
}

type ChatCompletionMessage struct {
	Content     ChatCompletionContent `json:"content"`
	ContentType string                `json:"content_type"`
}

type ChatCompletionRes struct {
	Message ChatCompletionMessage `json:"message"`
}

type AgentApp interface {
	ChatCompletion(ctx context.Context, appKey string, req *ChatCompletionReq, token string) (answer, thinking string, err error)
}

type agentApp struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

var (
	agentAppIns  AgentApp
	agentAppOnce sync.Once
)

func NewAgentApp() AgentApp {
	agentAppOnce.Do(func() {
		config := common.NewConfig()
		agentAppIns = &agentApp{
			baseURL:    fmt.Sprintf("http://%s:%s", config.AgentApp.Host, config.AgentApp.Port),
			httpClient: NewOtelHTTPClient(),
		}
	})

	return agentAppIns
}

func (a *agentApp) ChatCompletion(ctx context.Context, appKey string, req *ChatCompletionReq, token string) (answer, thinking string, err error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/agent-app/v1/app/%s/chat/completion", a.baseURL, appKey)
	headers := map[string]string{
		"Authorization":     fmt.Sprintf("Bearer %s", token),
		"Content-Type":      "application/json;charset=UTF-8",
		"X-Business-Domain": req.BizDomainID,
	}
	req.Stream = false

	_, resp, err := a.httpClient.Post(ctx, target, headers, req)

	if err != nil {
		log.Warnf("ChatCompletion failed: %v, url: %v", err, target)
		return "", "", err
	}

	bytes, _ := json.Marshal(resp)
	var res ChatCompletionRes
	err = json.Unmarshal(bytes, &res)

	if err != nil {
		return "", "", err
	}

	switch res.Message.ContentType {
	case "explore":
		for i, item := range res.Message.Content.FinalAnswer.SkillProcess {
			if i > 0 {
				answer += "\n"
				thinking += "\n"
			}
			answer += item.Text
			thinking += item.Thinking
		}
	default:
		answer = res.Message.Content.FinalAnswer.Answer.Text
		thinking = res.Message.Content.FinalAnswer.Thinking
	}

	return
}
