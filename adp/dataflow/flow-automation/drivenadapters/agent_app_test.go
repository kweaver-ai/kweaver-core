package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockAgentApp(clients *HttpClientMock) AgentApp {
	InitARLog()
	return &agentApp{
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestChatCompletion(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	app := NewMockAgentApp(httpClients)

	appKey := "test-app-key"
	token := "test-token"
	req := &ChatCompletionReq{
		AgentID:     "agent-1",
		Query:       "hello",
		BizDomainID: "domain-1",
	}

	Convey("TestChatCompletion", t, func() {
		Convey("Success Case - Default Content Type", func() {
			mockResp := map[string]interface{}{
				"message": map[string]interface{}{
					"content_type": "text",
					"content": map[string]interface{}{
						"final_answer": map[string]interface{}{
							"answer": map[string]interface{}{
								"text": "this is the answer",
							},
							"thinking": "thinking process",
						},
					},
				},
			}
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

			answer, thinking, err := app.ChatCompletion(context.Background(), appKey, req, token)

			assert.Equal(t, err, nil)
			assert.Equal(t, answer, "this is the answer")
			assert.Equal(t, thinking, "thinking process")
		})

		Convey("Success Case - Explore Content Type", func() {
			mockResp := map[string]interface{}{
				"message": map[string]interface{}{
					"content_type": "explore",
					"content": map[string]interface{}{
						"final_answer": map[string]interface{}{
							"skill_process": []interface{}{
								map[string]interface{}{
									"text":     "step 1",
									"thinking": "think 1",
								},
								map[string]interface{}{
									"text":     "step 2",
									"thinking": "think 2",
								},
							},
						},
					},
				},
			}
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

			answer, thinking, err := app.ChatCompletion(context.Background(), appKey, req, token)

			assert.Equal(t, err, nil)
			assert.Equal(t, answer, "step 1\nstep 2")
			assert.Equal(t, thinking, "think 1\nthink 2")
		})

		Convey("HTTP Client Error", func() {
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("network error"))

			_, _, err := app.ChatCompletion(context.Background(), appKey, req, token)

			assert.NotEqual(t, err, nil)
			assert.Equal(t, err.Error(), "network error")
		})

		Convey("Unmarshal Error", func() {
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, "invalid", nil)

			_, _, err := app.ChatCompletion(context.Background(), appKey, req, token)

			assert.NotEqual(t, err, nil)
		})
	})
}
