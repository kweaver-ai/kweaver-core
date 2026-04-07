package drivenadapters

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockEfast(clients *HttpClientMock) Efast {
	InitARLog()
	return &efastSvc{
		baseURL:        "http://localhost:8080",
		privateBaseURL: "http://localhost:8081",
		httpClient:     clients.httpClient,
	}
}

func TestCheckPermDoc(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ef := NewMockEfast(httpClient)
	ctx := context.Background()

	Convey("TestCheckPerm", t, func() {
		mockResp := map[string]interface{}{"result": float64(1)}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ef.CheckPerm(ctx, "d", "a", "t", "i")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, float64(1))
	})
}

func TestGetEntryDocLib(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ef := NewMockEfast(httpClient)
	ctx := context.Background()

	Convey("TestGetEntryDocLib", t, func() {
		mockResp := []interface{}{
			map[string]interface{}{"id": "i1", "type": "t1", "name": "n1"},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ef.GetEntryDocLib(ctx, "t", "token", "ip")
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].ID, "i1")
	})
}

func TestCreateDir(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ef := NewMockEfast(httpClient)
	ctx := context.Background()

	Convey("TestCreateDir", t, func() {
		mockResp := map[string]interface{}{
			"docid":       "d1",
			"editor":      "e1",
			"creator":     "c1",
			"modified":    float64(1),
			"create_time": float64(1),
		}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ef.CreateDir(ctx, "p", "n", 0, "t", "i")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.DocID, "d1")
	})
}

func TestGetDocMsg(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ef := NewMockEfast(httpClient)
	ctx := context.Background()

	Convey("TestGetDocMsg", t, func() {
		mockResp := map[string]interface{}{
			"name":        "n1",
			"size":        float64(100),
			"create_time": float64(1),
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ef.GetDocMsg(ctx, "d1")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.Name, "n1")
	})
}
