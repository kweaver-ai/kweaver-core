package drivenadapters

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockEcoTag(clients *HttpClientMock) EcoTag {
	InitARLog()
	return &ecoTag{
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestGetTags(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	et := NewMockEcoTag(httpClient)
	ctx := context.Background()

	Convey("TestGetTags", t, func() {
		mockResp := []interface{}{
			map[string]interface{}{
				"id":      "t1",
				"name":    "n1",
				"path":    "p1",
				"version": float64(1),
			},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := et.GetTags(ctx, map[string][]string{"id": {"t1"}})
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].ID, "t1")
		assert.Equal(t, res[0].Version, 1)
	})
}

func TestGetTagTrees(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	et := NewMockEcoTag(httpClient)
	ctx := context.Background()

	Convey("TestGetTagTrees", t, func() {
		mockResp := []interface{}{
			map[string]interface{}{
				"id":   "t1",
				"name": "n1",
				"child_tags": []interface{}{
					map[string]interface{}{
						"id":   "t2",
						"name": "n2",
					},
				},
			},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := et.GetTagTrees(ctx)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].ID, "t1")
		assert.Equal(t, len(res[0].ChildTags), 1)
	})
}
