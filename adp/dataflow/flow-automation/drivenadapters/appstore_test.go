package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockAppStore(clients *HttpClientMock) Appstore {
	InitARLog()
	return &appList{
		log:        commonLog.NewLogger(),
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestGetWhiteListStatus(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	appStore := NewMockAppStore(httpClient)

	ctx := context.Background()
	appName := "testApp"
	token := "Bearer testToken"

	Convey("TestGetWhiteListStatus", t, func() {
		Convey("Get WhiteList Status Success", func() {
			mockResp := map[string]interface{}{
				"accessible": true,
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

			res, err := appStore.GetWhiteListStatus(ctx, appName, token)
			assert.Equal(t, err, nil)
			assert.Equal(t, res["accessible"], true)
		})

		Convey("Get WhiteList Status Error", func() {
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))

			res, err := appStore.GetWhiteListStatus(ctx, appName, token)
			assert.NotEqual(t, err, nil)
			assert.Equal(t, res, nil)
		})
	})
}
