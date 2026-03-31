package drivenadapters

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockPersonalConfig(clients *HttpClientMock) PersonalConfig {
	return &PersonalConfigImpl{
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestGetModuleByName(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	pc := NewMockPersonalConfig(httpClient)
	ctx := context.Background()

	Convey("TestGetModuleByName", t, func() {
		mockResp := map[string]interface{}{
			"name": "m1",
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := pc.GetModuleByName(ctx, "m1")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.Name, "m1")
	})
}
