package cmpopenai

import (
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/sashabaranov/go-openai"
)

// OpenAICmp AI相关的功能封装
type OpenAICmp struct {
	client  *openai.Client
	baseURL string
	apiKey  string
	model   string
}

func NewOpenAICmp(apiKey, baseURL, model string, isTlsInsecureSkipVerify bool) icmp.IOpenAI {
	// if apiKey == "" || baseURL == "" || model == "" {
	// 	panic("[NewOpenAICmp]: apiKey and baseURL and model is required")
	// }
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	if isTlsInsecureSkipVerify {
		tran := http.DefaultTransport
		cutil.SetTpTlsInsecureSkipVerify(tran.(*http.Transport))
		config.HTTPClient = &http.Client{
			Transport: tran,
		}
	}

	client := openai.NewClientWithConfig(config)

	return &OpenAICmp{
		client:  client,
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
	}
}
