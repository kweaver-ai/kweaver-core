package drivenadapters

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockIntelliinfo(clients *HttpClientMock) Intelliinfo {
	InitARLog()
	return &intelliinfo{
		privateURL: "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestTransferData(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ii := NewMockIntelliinfo(httpClient)
	ctx := context.Background()

	Convey("TestTransferData", t, func() {
		Convey("V1", func() {
			req := &DataTransferReqV1{RuleID: "r1", Data: "d1"}
			mockResp := map[string]interface{}{"result": "ok"}
			httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

			res, err := ii.TransferData(ctx, req, V1, "u1", "t1")
			assert.Equal(t, err, nil)
			assert.Equal(t, res.(map[string]interface{})["result"], "ok")
		})

		Convey("V2", func() {
			req := &DataTransferReqV2{Datas: []*ADGraphTansferWrapper{{GrapID: 1}}}
			mockResp := map[string]interface{}{"result": "ok_v2"}
			httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), req).Return(200, mockResp, nil)

			res, err := ii.TransferData(ctx, req, V2, "u1", "t1")
			assert.Equal(t, err, nil)
			assert.Equal(t, res.(map[string]interface{})["result"], "ok_v2")
		})
	})
}
