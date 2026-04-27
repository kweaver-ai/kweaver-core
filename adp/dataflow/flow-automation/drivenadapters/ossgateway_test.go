package drivenadapters

import (
	"context"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockOssGateWay(clients *HttpClientMock) OssGateWay {
	InitARLog()
	return &ossGatetway{
		address: "http://localhost:8080",
		client:  clients.httpClient,
	}
}

func TestSimpleUpload(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	og := NewMockOssGateWay(httpClient)
	ctx := context.Background()

	Convey("TestSimpleUpload", t, func() {
		mockBody := map[string]interface{}{
			"url":    "http://oss/u",
			"method": "PUT",
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), nil).Return(200, mockBody, nil)
		httpClient.httpClient.EXPECT().OSSClient(ctx, "http://oss/u", "PUT", gomock.Any(), gomock.Any()).Return(nil, nil, nil)

		err := og.SimpleUpload(ctx, "oss1", "key1", true, strings.NewReader("data"))
		assert.Equal(t, err, nil)
	})
}

func TestGetDownloadURL(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	og := NewMockOssGateWay(httpClient)
	ctx := context.Background()

	Convey("TestGetDownloadURL", t, func() {
		mockResp := map[string]interface{}{"url": "http://download"}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), nil).Return(200, mockResp, nil)
		res, err := og.GetDownloadURL(ctx, "o", "k", 0, true)
		assert.Equal(t, err, nil)
		assert.Equal(t, res, "http://download")
	})
}
