package drivenadapters

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockUie(clients *HttpClientMock) Uie {
	InitARLog()
	return &uie{
		privateURL: "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestListTrainLog(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ui := NewMockUie(httpClient)
	ctx := context.Background()

	Convey("TestListTrainLog", t, func() {
		mockResp := map[string]interface{}{
			"res": []interface{}{
				map[string]interface{}{"id": 1, "status": "ok"},
			},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ui.ListTrainLog(ctx)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].ID, 1)
	})
}

func TestStartInfer(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ui := NewMockUie(httpClient)
	ctx := context.Background()

	Convey("TestStartInfer", t, func() {
		mockResp := map[string]interface{}{"entities": []interface{}{"e1"}}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ui.StartInfer(ctx, "schema", "text1")
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res["entities"].([]interface{})), 1)
	})
}
