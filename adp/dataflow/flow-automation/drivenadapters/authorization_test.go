package drivenadapters

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockAuthorization(clients *HttpClientMock) AuthorizationDriven {
	InitARLog()
	return &authorization{
		privateAddress: "http://localhost:8080",
		httpClient:     clients.httpClient,
	}
}

func TestCreatePolicy(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	authz := NewMockAuthorization(httpClient)
	ctx := context.Background()
	params := []PermPolicyParams{{Accessor: &Vistor{ID: "u1"}}}

	Convey("TestCreatePolicy", t, func() {
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), params).Return(200, nil, nil)
		err := authz.CreatePolicy(ctx, params)
		assert.Equal(t, err, nil)
	})
}

func TestUpdatePolicy(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	authz := NewMockAuthorization(httpClient)
	ctx := context.Background()
	params := []PermPolicyParams{{Accessor: &Vistor{ID: "u1"}}}

	Convey("TestUpdatePolicy", t, func() {
		httpClient.httpClient.EXPECT().Put(ctx, gomock.Any(), gomock.Any(), params).Return(200, nil, nil)
		err := authz.UpdatePolicy(ctx, []string{"id1"}, params)
		assert.Equal(t, err, nil)
	})
}

func TestDeletePolicy(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	authz := NewMockAuthorization(httpClient)
	ctx := context.Background()
	params := PermPolicyDeleteParams{Method: "m"}

	Convey("TestDeletePolicy", t, func() {
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), params).Return(200, nil, nil)
		err := authz.DeletePolicy(ctx, params)
		assert.Equal(t, err, nil)
	})
}

func TestResourceFilter(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	authz := NewMockAuthorization(httpClient)
	ctx := context.Background()
	params := ResourceFilterParmas{Method: "m"}

	Convey("TestResourceFilter", t, func() {
		mockResp := []interface{}{
			map[string]interface{}{"id": "r1", "type": "t1"},
		}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), params).Return(200, mockResp, nil)
		res, err := authz.ResourceFilter(ctx, params)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].ID, "r1")
	})
}

func TestListResourceOperation(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	authz := NewMockAuthorization(httpClient)
	ctx := context.Background()
	params := ListResourceOperationParmas{Method: "m"}

	Convey("TestListResourceOperation", t, func() {
		mockResp := []interface{}{
			map[string]interface{}{"id": "r1", "operation": []interface{}{"op1"}},
		}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), params).Return(200, mockResp, nil)
		res, err := authz.ListResourceOperation(ctx, params)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].ID, "r1")
	})
}

func TestOperationPermCheck(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	authz := NewMockAuthorization(httpClient)
	ctx := context.Background()
	params := OperationPermCheckParams{Method: "m"}

	Convey("TestOperationPermCheck", t, func() {
		mockResp := map[string]interface{}{"result": true}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), params).Return(200, mockResp, nil)
		res, err := authz.OperationPermCheck(ctx, params)
		assert.Equal(t, err, nil)
		assert.Equal(t, res, true)
	})
}

func TestListResourcePolicy(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	authz := NewMockAuthorization(httpClient)
	ctx := context.Background()
	params := ListResourceParams{Method: "m"}

	Convey("TestListResource", t, func() {
		mockResp := []interface{}{
			map[string]interface{}{"id": "r1"},
		}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), params).Return(200, mockResp, nil)
		res, err := authz.ListResource(ctx, params)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].ID, "r1")
	})
}
