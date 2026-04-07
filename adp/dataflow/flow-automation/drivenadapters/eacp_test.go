package drivenadapters

import (
	"testing"

	"github.com/go-playground/assert/v2"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockEacp(clients *HttpClientMock) Eacp {
	InitARLog()
	return &eacpSvc{
		baseURL:    "http://localhost:8080",
		log:        commonLog.NewLogger(),
		httpClient: clients.httpClient1,
	}
}

func TestGetUserInfoEacp(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	eacp := NewMockEacp(httpClient)

	Convey("TestGetUserInfo", t, func() {
		mockResp := map[string]interface{}{"userid": "u1", "name": "n1"}
		httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
		id, name, err := eacp.GetUserInfo("t")
		assert.Equal(t, err, nil)
		assert.Equal(t, id, "u1")
		assert.Equal(t, name, "n1")
	})
}

func TestCheckOwnerEacp(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	eacp := NewMockEacp(httpClient)

	Convey("TestCheckOwner", t, func() {
		mockResp := map[string]interface{}{"isowner": true}
		httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := eacp.CheckOwner("d", "t")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, true)
	})
}

func TestCheckPermEacp(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	eacp := NewMockEacp(httpClient)

	Convey("TestCheckPerm", t, func() {
		mockResp := map[string]interface{}{"result": float64(1)}
		httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := eacp.CheckPerm("d", "a", "t")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, float64(1))
	})
}
