package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockContentPipeline(clients *HttpClientMock) ContentPipeline {
	return &contentpipeline{
		privateBaseURL: "http://localhost:8080",
		httpClient2:    clients.httpClient2,
	}
}

func TestNewJob(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cp := NewMockContentPipeline(httpClient)
	ctx := context.Background()
	req := &NewJobReq{Source: &SourceData{Type: "t"}}

	Convey("TestNewJob", t, func() {
		Convey("Success", func() {
			httpClient.httpClient2.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), req, gomock.Any()).DoAndReturn(
				func(ctx, url, headers, req, resp any) (int, error) {
					r := resp.(*JobItem)
					r.ID = "job1"
					return 200, nil
				})
			res, err := cp.NewJob(ctx, req)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.ID, "job1")
		})

		Convey("Error", func() {
			httpClient.httpClient2.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), req, gomock.Any()).Return(500, fmt.Errorf("error"))
			_, err := cp.NewJob(ctx, req)
			assert.NotEqual(t, err, nil)
		})
	})
}
