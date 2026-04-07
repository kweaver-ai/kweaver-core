package drivenadapters

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/anydata.go -destination ../tests/mock_drivenadapters/anydata_mock.go

type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ChatCompletionRequest struct {
	Model            string         `json:"model"`
	Messages         []*ChatMessage `json:"messages"`
	Temperature      float64        `json:"temperature,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	TopK             int            `json:"top_k,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
}

type ChatCompletionChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	ID      string                  `json:"id"`
	Object  string                  `json:"object"`
	Created int64                   `json:"created"`
	Model   string                  `json:"model"`
	Choices []*ChatCompletionChoice `json:"choices"`
}

type EmbeddingResItem struct {
	Embedding []float64 `json:"embedding"`
}

type EmbeddingRes struct {
	Data []EmbeddingResItem `json:"data"`
}

type RerankerResult struct {
	RelevanceScore float64 `json:"relevance_score"`
	Index          int     `json:"index"`
}

type RerankerRes struct {
	Results []RerankerResult `json:"results"`
}

type AnyData interface {
	GetBaseURL() string
	GetAppID() string
	GetModelList(ctx context.Context, page *AnyDataPagination) (*GetModelListRes, error)
	GetLLMSource(ctx context.Context, page *AnyDataPagination) (*GetLLMSourceRes, error)
	GetAgentByID(ctx context.Context, agentID string) (agent *AgentInfo, err error)
	AddAgent(ctx context.Context, info *AgentInfo) (agentID string, err error)
	UpdateAgent(ctx context.Context, info *AgentInfo) (err error)
	CallAgent(ctx context.Context, agentKey string, inputs map[string]interface{}, options *CallAgentOptions, token string) (res *CallAgentRes, ch chan *CallAgentRes, err error)
	GetAgents(ctx context.Context, keyword string, status []string, page *AnyDataPagination) ([]*AgentInfo, error)
	GetAgent(ctx context.Context, id string) (agent *AgentInfo, err error)
	ChatCompletion(ctx context.Context, req *ChatCompletionRequest, token string) (*ChatCompletionResponse, error)
	GetGraphInfo(ctx context.Context, id uint64, token string) (*GraphInfo, error)
	Embedding(ctx context.Context, model string, input []string, token string) (*EmbeddingRes, error)
	Reranker(ctx context.Context, model string, query string, documents []string, token string) (*RerankerRes, error)
}

type AnyDataImpl struct {
	baseURL              string
	httpClient           otelHttp.HTTPClient
	httpClient2          HTTPClient2
	appID                string
	model                string
	agentFactoryBaseURL  string
	modelManagerBaseURL  string
	knowledgeDataBaseURL string
	modelApiBaseURL      string
}

var (
	anyDataOnce sync.Once
	ad          AnyData
)

func NewAnyData() AnyData {
	anyDataOnce.Do(func() {
		config := common.NewConfig()
		baseURL := ""

		if config.AnyData.Host != "" {
			baseURL = fmt.Sprintf("%s://%s:%d", config.AnyData.Protocol, config.AnyData.Host, config.AnyData.Port)
		}

		ad = &AnyDataImpl{
			baseURL:              baseURL,
			agentFactoryBaseURL:  fmt.Sprintf("http://%s:%s", config.AgentFactory.Host, config.AgentFactory.Port),
			modelManagerBaseURL:  fmt.Sprintf("http://%s:%s", config.MfModelManager.Host, config.MfModelManager.Port),
			knowledgeDataBaseURL: fmt.Sprintf("http://%s:%s", config.KnKnowledgeData.Host, config.KnKnowledgeData.Port),
			modelApiBaseURL:      fmt.Sprintf("http://%s:%s", config.MfModelApi.Host, config.MfModelApi.Port),
			model:                config.AnyData.Model,
			appID:                config.AnyData.AppID,
			httpClient:           NewOtelHTTPClient(),
			httpClient2:          NewHTTPClient2(),
		}
	})
	return ad
}

type AgentInfo struct {
	AgentID       string      `json:"agent_id,omitempty"`       // Agent ID
	Name          string      `json:"name,omitempty"`           // 名称
	Color         string      `json:"color,omitempty"`          // 图标颜色
	Description   string      `json:"description,omitempty"`    // 描述
	Template      string      `json:"template,omitempty"`       // 模板，可选值 kbqa, doc_qa, adv_qa
	DraftConfig   interface{} `json:"draft_config,omitempty"`   // 草稿配置
	ReleaseConfig interface{} `json:"release_config,omitempty"` // 发布配置
	ReleaseInfo   string      `json:"release_info,omitempty"`   // 发布版本信息
}

func (a *AnyDataImpl) GetAgentByID(ctx context.Context, agentID string) (*AgentInfo, error) {
	target := fmt.Sprintf("%s/api/agent-factory/v2/agent?agent_id=%s", a.baseURL, agentID)

	status, resp, err := a.httpClient.Get(ctx, target, map[string]string{
		"content-type":    "application/json;charset=UTF-8",
		"appid":           a.appID,
		"accept-language": "zh-CN",
	})

	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", status)
	}

	var response struct {
		Res *AgentInfo `json:"res"`
	}
	bytes, _ := json.Marshal(resp)
	err = json.Unmarshal(bytes, &response)

	if err != nil {
		return nil, err
	}

	return response.Res, nil
}

func (a *AnyDataImpl) AddAgent(ctx context.Context, info *AgentInfo) (agentID string, err error) {

	target := fmt.Sprintf("%s/api/agent-factory/v2/agent/add", a.baseURL)

	_, resp, err := a.httpClient.Post(ctx, target, map[string]string{
		"content-type":    "application/json;charset=UTF-8",
		"appid":           a.appID,
		"accept-language": "zh-CN",
	}, info)

	if err != nil {
		return "", err
	}

	bytes, _ := json.Marshal(resp)

	var result struct {
		Res       string `json:"res"`
		ErrorCode string `json:"ErrorCode"`
	}

	_ = json.Unmarshal(bytes, &result)

	if result.ErrorCode != "" {
		return "", fmt.Errorf("%s", result.ErrorCode)
	}

	return result.Res, nil
}

func (a *AnyDataImpl) UpdateAgent(ctx context.Context, info *AgentInfo) (err error) {
	target := fmt.Sprintf("%s/api/agent-factory/v2/agent/update", a.baseURL)

	_, resp, err := a.httpClient.Post(ctx, target, map[string]string{
		"content-type":    "application/json;charset=UTF-8",
		"appid":           a.appID,
		"accept-language": "zh-CN",
	}, info)

	if err != nil {
		return err
	}

	bytes, _ := json.Marshal(resp)

	var result struct {
		Res string `json:"res"`
	}

	_ = json.Unmarshal(bytes, &result)

	if result.Res != "ok" {
		return fmt.Errorf("unexpected response: %v", resp)
	}

	return nil
}

type AnyDataPagination struct {
	Page  int    `json:"page"`  // 页码
	Rule  string `json:"rule"`  // 排序 desc, asc
	Order string `json:"order"` // 排序方式 create_time
	Size  int    `json:"size"`  // 分页大小
}

type ModelItem struct {
	ModelID            string `json:"model_id"`
	ModelName          string `json:"model_name"`
	APIModel           string `json:"api_model"`
	InputTokens        int64  `json:"input_tokens"`
	CreateTime         string `json:"create_time"`
	InputTokensUsed    int64  `json:"input_tokens_used"`
	InputTokensRemain  int64  `json:"input_tokens_remain"`
	Icon               string `json:"icon"`
	BillingType        int    `json:"billing_type"`
	Quota              bool   `json:"quota"`
	MaxModelLen        int    `json:"max_model_len"`
	ModelParameters    any    `json:"model_parameters"`
	OutputTokens       int64  `json:"output_tokens"`
	OutputTokensUsed   int64  `json:"output_tokens_used"`
	OutputTokensRemain int64  `json:"output_tokens_remain"`
}

type GetModelListRes struct {
	ModelList []string     `json:"model_list"`
	Res       []*ModelItem `json:"res"`
	Total     int          `json:"total"`
}

func (a *AnyDataImpl) GetModelList(ctx context.Context, page *AnyDataPagination) (*GetModelListRes, error) {
	var err error

	if page == nil {
		page = &AnyDataPagination{
			Page:  1,
			Rule:  "desc",
			Order: "create_time",
			Size:  -1,
		}
	}

	getAll := false
	if page.Size == -1 {
		page.Size = 100
		getAll = true
	}

	target := fmt.Sprintf("%s/api/model-factory/v1/user-quota/model-list?page=%d&rule=%s&order=%s&size=%d", a.baseURL, page.Page, page.Rule, page.Order, page.Size)

	_, respParam, err := a.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8", "appid": a.appID})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetModelList failed: %v, url: %v", err, target)
		return nil, err
	}
	bytes, _ := json.Marshal(respParam)

	res := &GetModelListRes{}

	err = json.Unmarshal(bytes, res)

	if err != nil {
		return nil, err
	}

	if getAll && res.Total > len(res.ModelList) {
		page.Size = res.Total
		res, err = a.GetModelList(ctx, page)
	}

	return res, err
}

type LLMSourceItem struct {
	Model     string `json:"model"`
	ModelID   string `json:"model_id"`
	ModelName string `json:"model_name"`
}

type GetLLMSourceRes struct {
	ModelList []string `json:"model_list"`
	Res       struct {
		Data  []*LLMSourceItem `json:"data"`
		Total int              `json:"total"`
	} `json:"res"`
}

func (a *AnyDataImpl) GetLLMSource(ctx context.Context, page *AnyDataPagination) (*GetLLMSourceRes, error) {
	var err error

	if page == nil {
		page = &AnyDataPagination{
			Page:  1,
			Rule:  "update_time",
			Order: "desc",
			Size:  -1,
		}
	}

	getAll := false
	if page.Size == -1 {
		page.Size = 100
		getAll = true
	}

	target := fmt.Sprintf("%s/api/model-factory/v1/llm-source?page=%d&rule=%s&order=%s&size=%d", a.baseURL, page.Page, page.Rule, page.Order, page.Size)

	_, respParam, err := a.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8", "appid": a.appID})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetLLMSource failed: %v, url: %v", err, target)
		return nil, err
	}
	bytes, _ := json.Marshal(respParam)

	res := &GetLLMSourceRes{}

	err = json.Unmarshal(bytes, res)

	if err != nil {
		return nil, err
	}

	if getAll && res.Res.Total > len(res.ModelList) {
		page.Size = res.Res.Total
		res, err = a.GetLLMSource(ctx, page)
	}

	return res, err
}

type CallAgentRes struct {
	Answer      map[string]interface{} `json:"answer"`
	BlockAnswer map[string]interface{} `json:"block_answer"`
	Status      string                 `json:"status"` // True， False
}

type CallAgentResWrapper struct {
	Res CallAgentRes `json:"res"`
}

type CallAgentOptions struct {
	Debug  bool `json:"debug"`
	Retry  bool `json:"retry"`
	Stream bool `json:"stream"`
}

func (a *AnyDataImpl) CallAgent(ctx context.Context, agentKey string, inputs map[string]interface{}, options *CallAgentOptions, token string) (res *CallAgentRes, ch chan *CallAgentRes, err error) {
	if options == nil {
		options = &CallAgentOptions{Stream: false}
	}

	inputs["_options"] = options

	target := fmt.Sprintf("%s/api/agent-factory/v2/agent/%s/version/latest", a.agentFactoryBaseURL, agentKey)

	if options.Stream {
		ch = make(chan *CallAgentRes)

		payload, _ := json.Marshal(inputs)

		req, err := http.NewRequest("POST", target, bytes.NewBuffer(payload))
		if err != nil {
			traceLog.WithContext(ctx).Warnf("CallAgent create request error: %v", err)
			return nil, nil, err
		}
		req.Header.Set("accept", "text/event-stream")
		req.Header.Set("content-type", "application/json;charset=UTF-8")
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("accept-language", "zh-CN")
		client := NewRawHTTPClient()

		resp, err := client.Do(req)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("CallAgent request failed: %v, url: %v", err, target)
			return nil, nil, err
		}

		go func() {
			defer close(ch)
			defer resp.Body.Close()
			reader := bufio.NewReader(resp.Body)
			for {
				var res CallAgentRes
				line, err := reader.ReadBytes('\n')
				if err == io.EOF {
					break
				} else if err != nil {
					traceLog.WithContext(ctx).Warnf("CallAgent decode error: %v", err)
					return
				}

				text := string(line)

				if !strings.HasPrefix(text, "data: ") {
					continue
				}

				data := strings.TrimPrefix(text, "data: ")

				if err := json.Unmarshal([]byte(data), &res); err != nil {
					traceLog.WithContext(ctx).Warnf("CallAgent decode error: %v", err)
					return
				}

				ch <- &res
				if res.Status == "True" {
					break
				}
			}
		}()

		return nil, ch, nil
	} else {
		_, respParam, err := a.httpClient.Post(ctx, target, map[string]string{
			"Content-Type":  "application/json;charset=UTF-8",
			"authorization": fmt.Sprintf("Bearer %s", token),
		}, inputs)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("CallAgent request failed: %v, url: %v", err, target)
			return nil, nil, err
		}
		bytes, _ := json.Marshal(respParam)
		wrapper := &CallAgentResWrapper{}
		err = json.Unmarshal(bytes, wrapper)

		if err != nil {
			traceLog.WithContext(ctx).Warnf("CallAgent decode response error: %v", err)
			return nil, nil, err
		}

		return &wrapper.Res, nil, nil
	}
}

func (a *AnyDataImpl) GetBaseURL() string {
	return a.baseURL
}

func (a *AnyDataImpl) GetAppID() string {
	return a.appID
}

type GetAgentsRes struct {
	Agents []*AgentInfo `json:"agents"`
	Count  int          `json:"count"`
}

type GetAgentsResWrapper struct {
	Res GetAgentsRes `json:"res"`
}

func (a *AnyDataImpl) GetAgents(ctx context.Context, keyword string, status []string, page *AnyDataPagination) (agents []*AgentInfo, err error) {

	if page == nil {
		page = &AnyDataPagination{
			Page:  1,
			Rule:  "create_time",
			Order: "desc",
			Size:  -1,
		}
	}

	getAll := false
	if page.Size == -1 {
		page.Size = 100
		getAll = true
	}

	query := url.Values{
		"page":  []string{fmt.Sprintf("%d", page.Page)},
		"rule":  []string{page.Rule},
		"order": []string{page.Order},
		"size":  []string{fmt.Sprintf("%d", page.Size)},
	}

	if keyword != "" {
		query["name"] = []string{keyword}
	}

	if len(status) > 0 {
		query["status"] = []string{strings.Join(status, "+")}
	}

	target := fmt.Sprintf("%s/api/agent-factory/v2/agent/list?%s", a.baseURL, query.Encode())

	_, resp, err := a.httpClient.Get(ctx, target, map[string]string{
		"content-type":    "application/json;charset=UTF-8",
		"appid":           a.appID,
		"accept-language": "zh-CN",
	})

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetAgent decode response error: %v", err)
		return
	}

	var wrapper GetAgentsResWrapper
	bytes, _ := json.Marshal(resp)
	err = json.Unmarshal(bytes, &wrapper)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetAgent decode response error: %v", err)
		return nil, err
	}

	if getAll && wrapper.Res.Count > page.Size {
		page.Size = wrapper.Res.Count
		return a.GetAgents(ctx, keyword, status, page)
	}

	agents = wrapper.Res.Agents

	return
}

func (a *AnyDataImpl) GetAgent(ctx context.Context, id string) (agent *AgentInfo, err error) {
	target := fmt.Sprintf("%s/api/agent-factory/v2/agent?agent_id=%s", a.baseURL, id)

	_, resp, err := a.httpClient.Get(ctx, target, map[string]string{
		"content-type":    "application/json;charset=UTF-8",
		"appid":           a.appID,
		"accept-language": "zh-CN",
	})

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetAgent error: %v", err)
		return
	}

	var wrapper struct {
		Res *AgentInfo `json:"res"`
	}

	bytes, _ := json.Marshal(resp)
	err = json.Unmarshal(bytes, &wrapper)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetAgent decode response error: %v", err)
		return nil, err
	}
	agent = wrapper.Res

	return
}

func (a *AnyDataImpl) ChatCompletion(ctx context.Context, req *ChatCompletionRequest, token string) (*ChatCompletionResponse, error) {

	target := fmt.Sprintf("%s/api/mf-model-api/v1/chat/completions", a.modelApiBaseURL)
	_, resp, err := a.httpClient.Post(ctx, target, map[string]string{
		"content-type":  "application/json",
		"authorization": fmt.Sprintf("Bearer %s", token),
	}, req)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("ChatCompletion error: %v", err)
		return nil, err
	}

	res := &ChatCompletionResponse{}
	bytes, _ := json.Marshal(resp)
	err = json.Unmarshal(bytes, res)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("ChatCompletion decode response error: %v", err)
		return nil, err
	}

	return res, nil
}

type GraphInfo struct {
	Entity        []*Entity `json:"entity,omitempty"`
	Edge          []*Edge   `json:"edge,omitempty"`
	DBName        string    `json:"dbname,omitempty"`
	KnwID         int       `json:"knw_id,omitempty"`
	GraphName     string    `json:"graph_name,omitempty"`
	QuantizedFlag int       `json:"quantized_flag,omitempty"`
}

type Entity struct {
	EntityID   string      `json:"entity_id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Properties []*Property `json:"properties,omitempty"`
}

type Property struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Alias string `json:"alias,omitempty"`
}

type Edge struct {
	EdgeID     string      `json:"edge_id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Properties []*Property `json:"properties,omitempty"`
	Relations  []string    `json:"relations,omitempty"`
	Relation   []string    `json:"relation,omitempty"`
}

type GraphInfoRes struct {
	Res *GraphInfo `json:"res,omitempty"`
}

func (a *AnyDataImpl) GetGraphInfo(ctx context.Context, id uint64, token string) (*GraphInfo, error) {
	target := fmt.Sprintf("%s/api/kn-knowledge-data/v1/graph/info/onto?graph_id=%d", a.knowledgeDataBaseURL, id)
	_, resp, err := a.httpClient.Get(ctx, target, map[string]string{
		"content-type":  "application/json",
		"authorization": fmt.Sprintf("Bearer %s", token),
	})

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetGraphInfo error: %v", err)
		return nil, err
	}
	res := &GraphInfoRes{}
	bytes, _ := json.Marshal(resp)
	err = json.Unmarshal(bytes, res)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetGraphInfo decode response error: %v", err)
		return nil, err
	}
	return res.Res, nil
}

func (a *AnyDataImpl) Embedding(ctx context.Context, model string, input []string, token string) (*EmbeddingRes, error) {
	target := fmt.Sprintf("%s/api/mf-model-api/v1/small-model/embedding", a.modelApiBaseURL)
	headers := map[string]string{
		"content-type":  "application/json",
		"authorization": fmt.Sprintf("Bearer %s", token),
	}
	var result EmbeddingRes
	_, err := a.httpClient2.Post(ctx, target, headers, map[string]any{
		"model": model,
		"input": input,
	}, &result)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("Embedding response error: %v", err)
		return nil, err
	}

	return &result, nil
}

func (a *AnyDataImpl) Reranker(ctx context.Context, model string, query string, documents []string, token string) (*RerankerRes, error) {
	target := fmt.Sprintf("%s/api/mf-model-api/v1/small-model/reranker", a.modelApiBaseURL)
	headers := map[string]string{
		"content-type":  "application/json",
		"authorization": fmt.Sprintf("Bearer %s", token),
	}
	var result RerankerRes
	_, err := a.httpClient2.Post(ctx, target, headers, map[string]any{
		"model":     model,
		"query":     query,
		"documents": documents,
	}, &result)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("Reranker response error: %v", err)
		return nil, err
	}

	return &result, nil
}
