package drivenadapters

import (
	"context"
	"sync"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockBusinessDomain(clients *HttpClientMock) BusinessDomain {
	InitARLog()
	return &businessDomain{
		baseURL:    "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestBindResourceInternal(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	bd := NewMockBusinessDomain(httpClient)
	ctx := context.Background()
	params := BizDomainResourceParams{BizDomainID: "bd1"}

	Convey("TestBindResourceInternal", t, func() {
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), params).Return(200, nil, nil)
		err := bd.BindResourceInternal(ctx, params)
		assert.Equal(t, err, nil)
	})
}

func TestUnBindResourceInternal(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	bd := NewMockBusinessDomain(httpClient)
	ctx := context.Background()
	params := BizDomainResourceParams{BizDomainID: "bd1"}

	Convey("TestUnBindResourceInternal", t, func() {
		httpClient.httpClient.EXPECT().Delete(ctx, gomock.Any(), gomock.Any()).Return(nil, nil)
		err := bd.UnBindResourceInternal(ctx, params)
		assert.Equal(t, err, nil)
	})
}

func TestListResourceBD(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	bd := NewMockBusinessDomain(httpClient)
	ctx := context.Background()
	params := BizDomainResourceQuery{}
	token := "token"

	Convey("TestListResource", t, func() {
		mockResp := map[string]interface{}{
			"total": int64(1),
			"items": []interface{}{
				map[string]interface{}{"bd_id": "bd1", "id": "r1", "type": "data-flow"},
			},
		}
		httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		res, err := bd.ListResource(ctx, params, token)
		assert.Equal(t, err, nil)
		assert.Equal(t, res.Total, int64(1))
		assert.Equal(t, res.Items[0].BizDomainID, "bd1")
	})
}

func TestCheckerResource(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	bd := NewMockBusinessDomain(httpClient)
	ctx := context.Background()
	params := BizDomainResourceParams{BizDomainID: "bd1"}
	token := "token"

	Convey("TestCheckerResource", t, func() {
		Convey("Exist", func() {
			mockResp := map[string]interface{}{
				"total": 1,
				"items": []interface{}{
					map[string]interface{}{"bd_id": "bd1", "id": "r1"},
				},
			}
			httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			exist, err := bd.CheckerResource(ctx, params, token)
			assert.Equal(t, err, nil)
			assert.Equal(t, exist, true)
		})

		Convey("Not Exist", func() {
			mockResp := map[string]interface{}{
				"total": 0,
				"items": []interface{}{},
			}
			httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			exist, err := bd.CheckerResource(ctx, params, token)
			assert.Equal(t, err, nil)
			assert.Equal(t, exist, false)
		})
	})
}

func TestBusinessDomainDisabled(t *testing.T) {
	ctx := context.Background()
	params := BizDomainResourceParams{BizDomainID: "bd1"}
	token := "token"

	origin := common.NewConfig().Server.BusinessDomainEnabled
	originBD := bd
	common.NewConfig().Server.BusinessDomainEnabled = "false"
	bOnce = sync.Once{}
	bd = nil
	defer func() {
		common.NewConfig().Server.BusinessDomainEnabled = origin
		bOnce = sync.Once{}
		bd = originBD
	}()

	Convey("TestBusinessDomainDisabled", t, func() {
		Convey("NewBusinessDomain should return mock implementation", func() {
			bizDomain := NewBusinessDomain()
			_, ok := bizDomain.(*mockBusinessDomain)
			assert.Equal(t, ok, true)
		})

		Convey("BindResourceInternal should noop", func() {
			err := NewBusinessDomain().BindResourceInternal(ctx, params)
			assert.Equal(t, err, nil)
		})

		Convey("UnBindResourceInternal should noop", func() {
			err := NewBusinessDomain().UnBindResourceInternal(ctx, params)
			assert.Equal(t, err, nil)
		})

		Convey("ListResource should return empty result", func() {
			res, err := NewBusinessDomain().ListResource(ctx, BizDomainResourceQuery{
				BizDomainResourceParams: params,
			}, token)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Total, int64(0))
			assert.Equal(t, len(res.Items), 0)
		})

		Convey("CheckerResource should bypass remote check", func() {
			exist, err := NewBusinessDomain().CheckerResource(ctx, params, token)
			assert.Equal(t, err, nil)
			assert.Equal(t, exist, true)
		})
	})
}

func TestBizDomainResources_GetIDs(t *testing.T) {
	res := BizDomainResources{
		{ResourceID: "id1:type1"},
		{ResourceID: "id2:type2"},
		{ResourceID: "id3"},
		{ResourceID: ""},
		{ResourceID: "id5:type5:1"},
	}

	Convey("TestGetIDs", t, func() {
		Convey("All", func() {
			ids := res.GetIDs("")
			assert.Equal(t, len(ids), 5)
			assert.Equal(t, ids[0], "id1")
		})
		Convey("Filter Type", func() {
			ids := res.GetIDs("type1")
			assert.Equal(t, len(ids), 1)
			assert.Equal(t, ids[0], "id1")
		})
	})
}
