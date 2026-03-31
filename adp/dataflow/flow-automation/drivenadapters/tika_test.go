package drivenadapters

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockTika(clients *HttpClientMock) Tika {
	InitARLog()
	return &tika{
		tikaURL:         "http://localhost:8080",
		TextAnalysisURL: "http://localhost:8081",
		httpClient:      clients.httpClient,
	}
}

func TestParseContent(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	tk := NewMockTika(httpClient)
	ctx := context.Background()

	Convey("TestParseContent", t, func() {
		mockResp := []byte("plain text")
		httpClient.httpClient.EXPECT().Request(ctx, gomock.Any(), http.MethodPut, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		con := []byte("data")
		res, err := tk.ParseContent(ctx, "f.txt", &con)
		assert.Equal(t, err, nil)
		assert.Equal(t, string(*res), "plain text")
	})
}

func TestMatchContent(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	tk := NewMockTika(httpClient)
	ctx := context.Background()

	Convey("TestMatchContent", t, func() {
		mockResp := map[string]interface{}{
			"has_privacy_info": true,
			"results":          map[string]interface{}{},
		}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		con := []byte("data")
		res, err := tk.MatchContent(ctx, &con, map[string]interface{}{"name": "n1"})
		assert.Equal(t, err, nil)
		assert.Equal(t, res.HasPrivateInfo, true)
	})
}

func TestGetPrivacyTemplate(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	tk := NewMockTika(httpClient)
	ctx := context.Background()

	Convey("TestGetPrivacyTemplate", t, func() {
		mockResp := []interface{}{"tpl1", "tpl2"}
		httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := tk.GetPrivacyTemplate(ctx)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 2)
		assert.Equal(t, res[0], "tpl1")
	})
}
