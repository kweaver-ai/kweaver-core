package drivenadapters

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-playground/assert/v2"
	jsoniter "github.com/json-iterator/go"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockDocShare(clients *HttpClientMock) DocShare {
	InitARLog()
	return &docShare{
		address:        "http://localhost:8080",
		privateAddress: "http://localhost:8081",
		httpClient:     clients.httpClient,
	}
}

func TestGetDocConfig(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ds := NewMockDocShare(httpClient)
	ctx := context.Background()

	Convey("TestGetDocConfig", t, func() {
		mockResp := map[string]interface{}{
			"result": map[string]interface{}{"conf": "ok"},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ds.GetDocConfig(ctx, "u", "i", "d", "l")
		assert.Equal(t, err, nil)
		assert.Equal(t, res["conf"], "ok")
	})
}

func TestSetPerm(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ds := NewMockDocShare(httpClient)
	ctx := context.Background()

	Convey("TestSetPerm", t, func() {
		mockResp := map[string]interface{}{"result": float64(1)}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ds.SetPerm(ctx, "d", nil, true, "t")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, float64(1))
	})
}

func TestGetPerm(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ds := NewMockDocShare(httpClient)
	ctx := context.Background()

	Convey("TestGetPerm", t, func() {
		mockResp := GetPermRes{Inherit: true}
		respBody, _ := json.Marshal(mockResp)
		httpClient.httpClient.EXPECT().Request(ctx, gomock.Any(), "POST", gomock.Any(), gomock.Any()).Return(200, respBody, nil)
		res, err := ds.GetPerm(ctx, "d", "t")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.Inherit, true)
	})
}

func TestBatchGetDocPerm(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ds := NewMockDocShare(httpClient)
	ctx := context.Background()

	Convey("TestBatchGetDocPerm", t, func() {
		mockResp := []interface{}{
			map[string]interface{}{"id": "d1", "allow": []interface{}{"a"}},
		}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ds.BatchGetDocPerm(ctx, "m", nil, nil)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res["d1"]["allow"][0], "a")
	})
}

func TestCheckOwnerDoc(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ds := NewMockDocShare(httpClient)
	ctx := context.Background()

	Convey("TestCheckOwner", t, func() {
		mockResp := map[string]interface{}{
			"value": map[string]interface{}{
				"result": true,
			},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := ds.CheckOwner(ctx, "d", "o")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, true)
	})
}

func TestGetDocOwners(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ds := NewMockDocShare(httpClient)
	ctx := context.Background()

	Convey("TestGetDocOwners", t, func() {
		owners := []DocOwner{{DocID: "d1"}}
		respBody, _ := jsoniter.Marshal(owners)
		httpClient.httpClient.EXPECT().Request(ctx, gomock.Any(), "GET", gomock.Any(), gomock.Any()).Return(200, respBody, nil)
		res, err := ds.GetDocOwners(ctx, "d")
		assert.Equal(t, err, nil)
		assert.Equal(t, res[0].DocID, "d1")
	})
}
