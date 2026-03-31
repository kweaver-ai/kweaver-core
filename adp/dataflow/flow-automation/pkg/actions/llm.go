package actions

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	liburl "net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type MediaType string

const (
	MediaTypeVideo MediaType = "video"
	MediaTypeImage MediaType = "image"
)

type LLMAttachement struct {
	SourceType SourceType `json:"source_type"`
	DocID      string     `json:"docid"`
	Version    string     `json:"version"`
	Url        string     `json:"url"`
	MediaType  MediaType  `json:"media_type"`
}

type LLMChatCompletion struct {
	Model            string            `json:"model"`
	SystemMessage    string            `json:"system_message,omitempty"`
	UserMessage      string            `json:"user_message"`
	Temperature      float64           `json:"temperature,omitempty"`
	TopP             float64           `json:"top_p,omitempty"`
	MaxTokens        int               `json:"max_tokens,omitempty"`
	TopK             int               `json:"top_k,omitempty"`
	PresencePenalty  float64           `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64           `json:"frequency_penalty,omitempty"`
	Attachements     []*LLMAttachement `json:"attachements"`
}

func (a *LLMChatCompletion) Name() string {
	return common.OpLLMChatCompletion
}

func (a *LLMChatCompletion) ParameterNew() interface{} {
	return &LLMChatCompletion{}
}

func (a *LLMChatCompletion) Run(ctx entity.ExecuteContext, input interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)
	log := traceLog.WithContext(ctx.Context())

	params := input.(*LLMChatCompletion)
	ad := drivenadapters.NewAnyData()

	messages := []*drivenadapters.ChatMessage{}

	if len(params.SystemMessage) > 0 {
		messages = append(messages, &drivenadapters.ChatMessage{
			Role:    "system",
			Content: params.SystemMessage,
		})
	}

	if len(params.Attachements) > 0 {
		contents := []map[string]any{}

		for _, item := range params.Attachements {

			var url string

			switch item.SourceType {
			case SourceTypeUrl:
				name := DetectFilenameFromURL(item.Url)
				url, err = ToDataURL(item.Url, name)
				if err != nil {
					log.Warnf("[LLMChatCompletion] convert attachment to data url err %s, url %s", err.Error(), url)
					return nil, err
				}
			default:
				if item.DocID == "" {
					continue
				}
				downloadInfo, err := GetFileDownloadInfo(ctx.Context(), ctx, item.DocID, item.Version)
				if err != nil {
					log.Warnf("[LLMChatCompletion] get file download info err %s, docid %s, version %s", err.Error(), item.DocID, item.Version)
					return nil, err
				}
				url, err = ToDataURL(downloadInfo.URL, downloadInfo.Filename)
				if err != nil {
					log.Warnf("[LLMChatCompletion] convert attachment to data url err %s, docid %s, version", err.Error(), item.DocID, item.Version)
					return nil, err
				}
			}

			switch item.MediaType {
			case MediaTypeImage:
				contents = append(contents, map[string]any{
					"type": "image_url",
					"image_url": map[string]string{
						"url": url,
					},
				})
			case MediaTypeVideo:
				contents = append(contents, map[string]any{
					"type": "video_url",
					"video_url": map[string]string{
						"url": url,
					},
				})
			default:
				continue
			}
		}

		if len(contents) > 0 {
			messages = append(messages, &drivenadapters.ChatMessage{
				Role:    common.User.ToString(),
				Content: contents,
			})
		}
	}

	if len(params.UserMessage) > 0 {
		messages = append(messages, &drivenadapters.ChatMessage{
			Role:    common.User.ToString(),
			Content: params.UserMessage,
		})
	}

	res, err := ad.ChatCompletion(newCtx, &drivenadapters.ChatCompletionRequest{
		Model:            params.Model,
		Messages:         messages,
		Temperature:      params.Temperature,
		TopP:             params.TopP,
		MaxTokens:        params.MaxTokens,
		TopK:             params.TopK,
		PresencePenalty:  params.PresencePenalty,
		FrequencyPenalty: params.FrequencyPenalty,
	}, token.Token)

	if err != nil {
		log.Warnf("[LLMChatCompletion] ChatCompletion err %s, model %s", err.Error(), params.Model)
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		return nil, err
	}

	result := make(map[string]interface{})

	if len(res.Choices) == 0 {
		result["answer"] = ""
	} else {
		result["answer"] = res.Choices[0].Message.Content
	}

	if answer, ok := result["answer"].(string); ok {
		result["json"] = ParseJsonValue(answer)
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)

	ctx.Trace(newCtx, "run end")
	return result, nil
}

func ToDataURL(url string, filename string) (string, error) {
	client := &http.Client{
		Transport: otelhttp.NewTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}),
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch file: %s", string(body))
	}

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	mimeType := DetectContentType(filename, buf, n)
	if n == 0 {
		return "data:application/octet-stream;base64,", nil
	}

	var sb strings.Builder
	sb.Grow(64 + n*4/3)
	sb.WriteString("data:")
	sb.WriteString(mimeType)
	sb.WriteString(";base64,")

	encoder := base64.NewEncoder(base64.StdEncoding, &sb)
	_, err = encoder.Write(buf[:n])
	if err != nil {
		return "", err
	}

	_, err = io.Copy(encoder, resp.Body)
	if err != nil {
		encoder.Close()
		return "", err
	}
	encoder.Close()

	return sb.String(), nil
}

var customContentTypes = map[string]string{
	".wmv": "video/x-ms-wmv",
	".mov": "video/quicktime",
	".flv": "video/x-flv",
	".mkv": "video/x-matroska",
}

func DetectContentType(filename string, buf []byte, n int) string {
	if filename != "" {
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != "" {
			if mimeType, ok := customContentTypes[ext]; ok {
				return mimeType
			}

			if mimeType := mime.TypeByExtension(ext); mimeType != "" {
				return mimeType
			}
		}
	}

	if n > 0 {
		if sniffType := http.DetectContentType(buf[:n]); sniffType != "" {
			return sniffType
		}
	}

	return "application/octet-stream"
}

func DetectFilenameFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	u, err := liburl.Parse(rawURL)
	if err != nil {
		return ""
	}

	queryCandidates := []string{"filename", "file", "name", "fileName"}
	for _, key := range queryCandidates {
		if value := u.Query().Get(key); value != "" {
			if decoded, err := liburl.QueryUnescape(value); err == nil {
				return decoded
			}
			return value
		}
	}

	base := path.Base(u.Path)
	if base == "." || base == "/" {
		return ""
	}

	if decoded, err := liburl.PathUnescape(base); err == nil {
		return decoded
	}

	return base
}

func ParseJsonValue(s string) any {
	var result any

	if len(s) == 0 {
		return nil
	}

	// 首先尝试直接解析整个字符串
	if err := json.Unmarshal([]byte(s), &result); err == nil {
		return result
	}

	pattern := "(?s)```json\\s*(.*?)\\s*```"
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(s)

	if len(matches) > 1 {
		jsonContent := strings.TrimSpace(matches[1])
		if err := json.Unmarshal([]byte(jsonContent), &result); err == nil {
			return result
		}
	}

	return nil
}

type LLMEmbedding struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

func (a *LLMEmbedding) Name() string {
	return common.OpLLmEmbedding
}

func (a *LLMEmbedding) ParameterNew() interface{} {
	return &LLMEmbedding{}
}

func (a *LLMEmbedding) Run(ctx entity.ExecuteContext, input interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)
	log := traceLog.WithContext(ctx.Context())

	ad := drivenadapters.NewAnyData()

	param := input.(*LLMEmbedding)
	res, err := ad.Embedding(ctx.Context(), param.Model, param.Input, token.Token)

	if err != nil {
		log.Warnf("[LLMEmbedding] ad.Embedding err %s, model %s", err.Error(), param.Model)
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		return nil, err
	}

	result := map[string]any{
		"data": res.Data,
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	return result, nil
}

type LLMReranker struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
}

func (a *LLMReranker) Name() string {
	return common.OpLLMReranker
}

func (a *LLMReranker) ParameterNew() interface{} {
	return &LLMReranker{}
}

func (a *LLMReranker) Run(ctx entity.ExecuteContext, input interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)
	log := traceLog.WithContext(ctx.Context())

	ad := drivenadapters.NewAnyData()

	param := input.(*LLMReranker)
	res, err := ad.Reranker(ctx.Context(), param.Model, param.Query, param.Documents, token.Token)

	if err != nil {
		log.Warnf("[LLMEmbedding] ad.Reranker err %s, model %s", err.Error(), param.Model)
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		return nil, err
	}

	var documents []string

	for _, result := range res.Results {
		if result.Index < len(param.Documents) {
			documents = append(documents, param.Documents[result.Index])
		}
	}

	result := map[string]any{
		"results":   res.Results,
		"documents": documents,
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	return result, nil
}
