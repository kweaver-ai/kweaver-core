package drivenadapters

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockAuthentication(clients *HttpClientMock) Authentication {
	InitARLog()
	return &auth{
		publicAddress:  "http://localhost:8080",
		privateAddress: "http://localhost:8081",
		log:            commonLog.NewLogger(),
		httpClient:     clients.httpClient1,
	}
}

func TestConfigAuthPerm(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	auth := NewMockAuthentication(httpClient)

	Convey("TestConfigAuthPerm", t, func() {
		Convey("Success", func() {
			httpClient.httpClient1.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil, nil)
			err := auth.ConfigAuthPerm("appID")
			assert.Equal(t, err, nil)
		})

		Convey("Error", func() {
			httpClient.httpClient1.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			err := auth.ConfigAuthPerm("appID")
			assert.NotEqual(t, err, nil)
		})
	})
}

func TestGetAssertion(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	auth := NewMockAuthentication(httpClient)

	Convey("TestGetAssertion", t, func() {
		Convey("Success", func() {
			mockResp := map[string]interface{}{
				"assertion": "jwt-token",
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			assertion, err := auth.GetAssertion("userID", "token")
			assert.Equal(t, err, nil)
			assert.Equal(t, assertion, "jwt-token")
		})

		Convey("Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			assertion, err := auth.GetAssertion("userID", "token")
			assert.NotEqual(t, err, nil)
			assert.Equal(t, assertion, "")
		})

		Convey("Invalid Response", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return("not a map", nil)
			assertion, err := auth.GetAssertion("userID", "token")
			assert.Equal(t, err, nil)
			assert.Equal(t, assertion, "")
		})
	})
}
