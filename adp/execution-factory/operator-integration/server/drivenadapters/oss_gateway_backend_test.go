package drivenadapters

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestOSSGatewayBackendCurrentStorageIDRecoversAfterDefaultStorageCreated(t *testing.T) {
	Convey("CurrentStorageID retries loading default storage when it was missing at startup", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ossGatewayBackendClient{
			httpClient:     mockHTTPClient,
			baseURL:        "http://oss-gateway/api/v1",
			refreshDefault: true,
			logger:         mockLogger,
			stopCh:         make(chan struct{}),
		}

		mockHTTPClient.EXPECT().
			GetNoUnmarshal(gomock.Any(), "http://oss-gateway/api/v1/storages?enabled=true&is_default=true", gomock.Any(), gomock.Any()).
			Return(http.StatusOK, []byte(`{"count":0,"data":[]}`), nil)
		err := client.initStorageID(context.Background())
		So(err, ShouldNotBeNil)
		So(client.IsReady(), ShouldBeFalse)

		mockHTTPClient.EXPECT().
			GetNoUnmarshal(gomock.Any(), "http://oss-gateway/api/v1/storages?enabled=true&is_default=true", gomock.Any(), gomock.Any()).
			Return(http.StatusOK, []byte(`{"count":1,"data":[{"storage_id":"storage-1","is_default":true,"is_enabled":true}]}`), nil)

		storageID, err := client.CurrentStorageID(context.Background())
		So(err, ShouldBeNil)
		So(storageID, ShouldEqual, "storage-1")
		So(client.IsReady(), ShouldBeTrue)
	})
}

func TestOSSGatewayBackendIsReadyDependsOnStorageIDOnly(t *testing.T) {
	Convey("IsReady only reflects whether a storage id is currently available", t, func() {
		client := &ossGatewayBackendClient{}
		So(client.IsReady(), ShouldBeFalse)

		client.storeStorageID("storage-1")
		So(client.IsReady(), ShouldBeTrue)

		client.storageMu.Lock()
		client.storageID = ""
		client.storageMu.Unlock()
		So(client.IsReady(), ShouldBeFalse)
	})
}
