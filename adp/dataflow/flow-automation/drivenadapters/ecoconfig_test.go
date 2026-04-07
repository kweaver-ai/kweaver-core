package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockEcoconfig(clients *HttpClientMock) Ecoconfig {
	InitARLog()
	return &EcoconfigImpl{
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestReindex(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	eco := NewMockEcoconfig(httpClient)
	ctx := context.Background()

	Convey("TestReindex", t, func() {
		Convey("Success", func() {
			httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil, nil)
			code, err := eco.Reindex(ctx, "d1", "t1")
			assert.Equal(t, err, nil)
			assert.Equal(t, code, 200)
		})

		Convey("Error", func() {
			httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := eco.Reindex(ctx, "d1", "t1")
			assert.NotEqual(t, err, nil)
		})
	})
}
