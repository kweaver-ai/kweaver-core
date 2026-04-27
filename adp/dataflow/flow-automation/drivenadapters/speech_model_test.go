package drivenadapters

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockSpeechModel(clients *HttpClientMock) SpeechModel {
	InitARLog()
	return &speechModel{
		privateURL: "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestAudioTransfer(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	sm := NewMockSpeechModel(httpClient)
	ctx := context.Background()

	Convey("TestAudioTransfer", t, func() {
		mockResp := map[string]string{"id": "task1"}
		respByte, _ := json.Marshal(mockResp)
		httpClient.httpClient.EXPECT().Request(ctx, gomock.Any(), http.MethodPost, gomock.Any(), gomock.Any()).Return(200, respByte, nil)

		content := []byte("audio")
		res, err := sm.AudioTransfer(ctx, "a.wav", "hook", &content)
		assert.Equal(t, err, nil)
		assert.Equal(t, res, "task1")
	})
}

func TestGetAudioTransferResult(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	sm := NewMockSpeechModel(httpClient)
	ctx := context.Background()

	Convey("TestGetAudioTransferResult", t, func() {
		mockResp := map[string]interface{}{"text": "hello"}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		code, res, err := sm.GetAudioTransferResult(ctx, "task1")
		assert.Equal(t, err, nil)
		assert.Equal(t, code, 200)
		assert.Equal(t, res["text"], "hello")
	})
}
