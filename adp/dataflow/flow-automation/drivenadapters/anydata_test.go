package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
	
)

func NewMockAnyData(clients *HttpClientMock) AnyData {
	InitARLog()
	return &AnyDataImpl{
		baseURL:              "http://localhost:8080",
		httpClient:           clients.httpClient,
		httpClient2:          clients.httpClient2,
		appID:                "appid",
		model:                "model",
		agentFactoryBaseURL:  "http://localhost:8081",
		modelManagerBaseURL:  "http://localhost:8082",
		knowledgeDataBaseURL: "http://localhost:8083",
		modelApiBaseURL:      "http://localhost:8084",
	}
}

func TestGetAgentByID(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestGetAgentByID", t, func() {
		Convey("Get Agent By ID Error", func() {
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := anyData.GetAgentByID(context.Background(), "id")
			assert.NotEqual(t, err, nil)
		})
		Convey("Get Agent By ID Success", func() {
			mockResp := map[string]interface{}{
				"res": map[string]interface{}{
					"agent_id": "agent_id",
				},
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			agent, err := anyData.GetAgentByID(context.Background(), "id")
			assert.Equal(t, err, nil)
			assert.Equal(t, agent.AgentID, "agent_id")
		})
	})
}

func TestAddAgent(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)
	info := &AgentInfo{Name: "test"}

	Convey("TestAddAgent", t, func() {
		Convey("Add Agent Error", func() {
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := anyData.AddAgent(context.Background(), info)
			assert.NotEqual(t, err, nil)
		})
		Convey("Add Agent Success", func() {
			mockResp := map[string]interface{}{
				"res": "new_id",
			}
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			id, err := anyData.AddAgent(context.Background(), info)
			assert.Equal(t, err, nil)
			assert.Equal(t, id, "new_id")
		})
	})
}

func TestUpdateAgent(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)
	info := &AgentInfo{AgentID: "id", Name: "test"}

	Convey("TestUpdateAgent", t, func() {
		Convey("Update Agent Success", func() {
			mockResp := map[string]interface{}{
				"res": "ok",
			}
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			err := anyData.UpdateAgent(context.Background(), info)
			assert.Equal(t, err, nil)
		})
	})
}

func TestGetModelList(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestGetModelList", t, func() {
		Convey("GetModelList Success", func() {
			mockResp := map[string]interface{}{
				"model_list": []string{"m1"},
				"total":      1,
				"res": []interface{}{
					map[string]interface{}{"model_id": "m1"},
				},
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			res, err := anyData.GetModelList(context.Background(), nil)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Total, 1)
		})
	})
}

func TestGetLLMSource(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestGetLLMSource", t, func() {
		Convey("GetLLMSource Success", func() {
			mockResp := map[string]interface{}{
				"model_list": []string{"s1"},
				"res": map[string]interface{}{
					"data": []interface{}{
						map[string]interface{}{"model_id": "s1"},
					},
					"total": 1,
				},
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			res, err := anyData.GetLLMSource(context.Background(), nil)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Res.Total, 1)
		})
	})
}

func TestCallAgent(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestCallAgent", t, func() {
		Convey("CallAgent Success", func() {
			mockResp := map[string]interface{}{
				"res": map[string]interface{}{
					"answer": map[string]interface{}{"text": "done"},
					"status": "True",
				},
			}
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			res, ch, err := anyData.CallAgent(context.Background(), "key", map[string]interface{}{}, nil, "token")
			assert.Equal(t, err, nil)
			assert.Equal(t, ch, nil)
			assert.Equal(t, res.Status, "True")
		})
	})
}

func TestGetAgents(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestGetAgents", t, func() {
		Convey("GetAgents Success", func() {
			mockResp := map[string]interface{}{
				"res": map[string]interface{}{
					"agents": []interface{}{
						map[string]interface{}{"agent_id": "a1"},
					},
					"count": 1,
				},
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			agents, err := anyData.GetAgents(context.Background(), "", nil, nil)
			assert.Equal(t, err, nil)
			assert.Equal(t, len(agents), 1)
		})
	})
}

func TestGetAgent(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestGetAgent", t, func() {
		Convey("GetAgent Success", func() {
			mockResp := map[string]interface{}{
				"res": map[string]interface{}{
					"agent_id": "a1",
				},
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			agent, err := anyData.GetAgent(context.Background(), "a1")
			assert.Equal(t, err, nil)
			assert.Equal(t, agent.AgentID, "a1")
		})
	})
}

func TestChatCompletionAnyData(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestChatCompletion", t, func() {
		Convey("ChatCompletion Success", func() {
			mockResp := map[string]interface{}{
				"id": "chat-1",
			}
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			res, err := anyData.ChatCompletion(context.Background(), &ChatCompletionRequest{}, "token")
			assert.Equal(t, err, nil)
			assert.Equal(t, res.ID, "chat-1")
		})
	})
}

func TestGetGraphInfo(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestGetGraphInfo", t, func() {
		Convey("GetGraphInfo Success", func() {
			mockResp := map[string]interface{}{
				"res": map[string]interface{}{
					"graph_name": "g1",
				},
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			res, err := anyData.GetGraphInfo(context.Background(), 1, "token")
			assert.Equal(t, err, nil)
			assert.Equal(t, res.GraphName, "g1")
		})
	})
}

func TestEmbedding(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestEmbedding", t, func() {
		Convey("Embedding Success", func() {
			httpClient.httpClient2.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx, url, headers, req, resp any) (int, error) {
					r := resp.(*EmbeddingRes)
					r.Data = []EmbeddingResItem{{Embedding: []float64{0.1}}}
					return 200, nil
				})
			res, err := anyData.Embedding(context.Background(), "m", []string{"i"}, "t")
			assert.Equal(t, err, nil)
			assert.Equal(t, len(res.Data), 1)
		})
	})
}

func TestReranker(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyData := NewMockAnyData(httpClient)

	Convey("TestReranker", t, func() {
		Convey("Reranker Success", func() {
			httpClient.httpClient2.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx, url, headers, req, resp any) (int, error) {
					r := resp.(*RerankerRes)
					r.Results = []RerankerResult{{RelevanceScore: 0.9}}
					return 200, nil
				})
			res, err := anyData.Reranker(context.Background(), "m", "q", []string{"d"}, "t")
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Results[0].RelevanceScore, 0.9)
		})
	})
}

func TestCallAgent_Stream(t *testing.T) {
	httpClient := NewHttpClientMock(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"status\": \"False\", \"answer\": {\"text\": \"Thinking...\"}}\n")
		fmt.Fprint(w, "data: {\"status\": \"True\", \"answer\": {\"text\": \"Finished\"}}\n")
	}))
	defer server.Close()

	anyData := &AnyDataImpl{
		agentFactoryBaseURL: server.URL,
		httpClient:          httpClient.httpClient,
		appID:               "appid",
	}

	Convey("Test CallAgent Stream", t, func() {
		res, ch, err := anyData.CallAgent(context.Background(), "key", map[string]interface{}{}, &CallAgentOptions{Stream: true}, "token")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, nil)
		assert.NotEqual(t, ch, nil)

		var results []*CallAgentRes
		for r := range ch {
			results = append(results, r)
		}

		assert.Equal(t, len(results), 2)
		assert.Equal(t, results[0].Status, "False")
		assert.Equal(t, results[1].Status, "True")
		assert.Equal(t, results[1].Answer["text"], "Finished")
	})
}
