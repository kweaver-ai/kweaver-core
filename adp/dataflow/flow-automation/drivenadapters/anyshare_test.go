package drivenadapters

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockAnyshare(clients *HttpClientMock) Anyshare {
	InitARLog()
	return &anyshareClient{
		log:        commonLog.NewLogger(),
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient1,
	}
}

func TestClusterAccess(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	anyshare := NewMockAnyshare(httpClient)

	Convey("TestClusterAccess", t, func() {
		Convey("HTTP Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("network error"))
			_, err := anyshare.ClusterAccess()
			assert.NotEqual(t, err, nil)
		})

		Convey("Nil Response", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, nil)
			_, err := anyshare.ClusterAccess()
			assert.NotEqual(t, err, nil)
			assert.Equal(t, err.Error(), "cluster access response is nil")
		})

		Convey("Invalid Response Type", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return("not a map", nil)
			_, err := anyshare.ClusterAccess()
			assert.NotEqual(t, err, nil)
			assert.Equal(t, err.Error(), "cluster access response is not a map")
		})

		Convey("Missing Host or Port", func() {
			mockResp := map[string]interface{}{
				"host": "localhost",
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			_, err := anyshare.ClusterAccess()
			assert.NotEqual(t, err, nil)
			assert.Equal(t, err.Error(), "cluster access response missing host or port")
		})

		Convey("Nil Host or Port", func() {
			mockResp := map[string]interface{}{
				"host": "localhost",
				"port": nil,
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			_, err := anyshare.ClusterAccess()
			assert.NotEqual(t, err, nil)
			assert.Equal(t, err.Error(), "cluster access host or port is nil")
		})

		Convey("Host or Port Not String", func() {
			mockResp := map[string]interface{}{
				"host": "localhost",
				"port": 8080,
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			_, err := anyshare.ClusterAccess()
			assert.NotEqual(t, err, nil)
			assert.Equal(t, err.Error(), "cluster access host or port is not a string")
		})

		Convey("Success", func() {
			mockResp := map[string]interface{}{
				"host": "127.0.0.1",
				"port": "8080",
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			access, err := anyshare.ClusterAccess()
			assert.Equal(t, err, nil)
			assert.Equal(t, access.Host, "127.0.0.1")
			assert.Equal(t, access.Port, "8080")
		})
	})
}
