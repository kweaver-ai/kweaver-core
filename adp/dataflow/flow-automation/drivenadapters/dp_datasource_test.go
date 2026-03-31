package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockDPDataSource(clients *HttpClientMock) IDataSource {
	InitARLog()
	return &DataSourceImpl{
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestGetDataSourceCatalog(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ds := NewMockDPDataSource(httpClient)
	ctx := context.Background()

	Convey("TestGetDataSourceCatalog", t, func() {
		Convey("Success", func() {
			mockResp := map[string]interface{}{
				"id":   "ds1",
				"name": "name1",
			}
			httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			res, err := ds.GetDataSourceCatalog(ctx, "ds1", "token", "ip")
			assert.Equal(t, err, nil)
			assert.Equal(t, res.ID, "ds1")
			assert.Equal(t, res.Name, "name1")
		})

		Convey("HTTP Error", func() {
			httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := ds.GetDataSourceCatalog(ctx, "ds1", "token", "ip")
			assert.NotEqual(t, err, nil)
		})
	})
}
