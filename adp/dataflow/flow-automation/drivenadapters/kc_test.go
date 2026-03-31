package drivenadapters

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockKcmc(clients *HttpClientMock) Kcmc {
	InitARLog()
	return &kcmc{
		url:        "http://localhost:8080",
		urlPrivate: "http://localhost:8081",
		httpClient: clients.httpClient,
	}
}

func TestGetUserEntity(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	kc := NewMockKcmc(httpClient)
	ctx := context.Background()

	Convey("TestGetUserEntity", t, func() {
		mockResp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"id": "u1"},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := kc.GetUserEntity(ctx, "u1", "token")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.Code, 0)
		assert.Equal(t, res.Data.ID, "u1")
	})
}

func TestGetArticleByProxyDirID(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	kc := NewMockKcmc(httpClient)
	ctx := context.Background()

	Convey("TestGetArticleByProxyDirID", t, func() {
		mockResp := map[string]interface{}{
			"code": 200033800,
			"data": map[string]interface{}{
				"article_id": "123",
				"title":      "title1",
			},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := kc.GetArticleByProxyDirID(ctx, "p1")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.Title, "title1")
		assert.Equal(t, res.ArticleID, uint64(123))
	})
}

func TestKcSetPerm(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	kc := NewMockKcmc(httpClient)
	ctx := context.Background()

	Convey("TestSetPerm", t, func() {
		mockResp := map[string]interface{}{
			"code": 200033800,
		}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := kc.SetPerm(ctx, KcPerm{}, "token")
		assert.Equal(t, err, nil)
		assert.Equal(t, res, float64(0))
	})
}
