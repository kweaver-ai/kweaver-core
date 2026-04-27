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

func NewMockOcr(clients *HttpClientMock, raw *http.Client) Ocr {
	InitARLog()
	return &ocr{
		t4thURL:    "http://localhost:8080",
		ocrBaseURL: "http://localhost:8081",
		httpClient: clients.httpClient,
		rawClient:  raw,
	}
}

func TestSubmitTask(t *testing.T) {
	Convey("TestSubmitTask", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "POST")
			fmt.Fprint(w, `{"code":0, "data":"task123"}`)
		}))
		defer server.Close()

		// ocr uses a local http.Client in SubmitTask, but it's hardcoded.
		// Wait, look at ocr.go:107: client := &http.Client{}.
		// This means I can't inject it without changing the code.
		// However, I can point t4thURL to the test server.

		o := &ocr{
			t4thURL: server.URL,
		}
		data := []byte("content")
		res, err := o.SubmitTask(context.Background(), "f.txt", &data)
		assert.Equal(t, err, nil)
		assert.Equal(t, res.Data, "task123")
	})
}

func TestGetResultDetails(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	o := NewMockOcr(httpClient, nil)

	Convey("TestGetResultDetails", t, func() {
		mockResp := map[string]interface{}{"status": "done"}
		httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := o.GetResultDetails(context.Background(), "t1")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.(map[string]interface{})["status"], "done")
	})
}

func TestRecognizeText(t *testing.T) {
	Convey("TestRecognizeText", t, func() {
		ocrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[[1, "hello"], [2, "world"]]`)
		}))
		defer ocrServer.Close()

		fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "image data")
		}))
		defer fileServer.Close()

		o := &ocr{
			ocrBaseURL: ocrServer.URL,
			rawClient:  ocrServer.Client(),
		}

		res, err := o.RecognizeText(context.Background(), fileServer.URL, "i.png")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, "hello world")
	})
}
