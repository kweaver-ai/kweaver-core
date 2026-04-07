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

func NewMockCognitiveAssistant(clients *HttpClientMock) CognitiveAssistant {
	InitARLog()
	return &CognitiveAssistantImpl{
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestGetCustomPrompts(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ca := NewMockCognitiveAssistant(httpClient)
	ctx := context.Background()

	Convey("TestGetCustomPrompts", t, func() {
		mockResp := map[string]interface{}{"prompts": []interface{}{"p1"}}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := ca.GetCustomPrompts(ctx)
		assert.Equal(t, err, nil)
		assert.Equal(t, res.(map[string]interface{})["prompts"].([]interface{})[0], "p1")
	})
}

func TestCustomPrompt(t *testing.T) {
	Convey("TestCustomPrompt", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data:{\"status\":\"processing\",\"results\":\"\"}\n")
			fmt.Fprint(w, "data:{\"status\":\"completed\",\"results\":\"final answer\"}\n")
		}))
		defer server.Close()

		ca := &CognitiveAssistantImpl{
			baseURL: server.URL,
		}

		res, err := ca.CustomPrompt(context.Background(), "serviceID", "content")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, "final answer")
	})

	Convey("TestCustomPrompt Error Status", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data:{\"status\":\"failed\",\"results\":\"\"}\n")
		}))
		defer server.Close()

		ca := &CognitiveAssistantImpl{
			baseURL: server.URL,
		}

		_, err := ca.CustomPrompt(context.Background(), "serviceID", "content")
		assert.NotEqual(t, err, nil)
	})
}
