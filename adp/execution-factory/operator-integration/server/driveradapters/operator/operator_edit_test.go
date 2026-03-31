package operator

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestOperatorEdit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
	mockHydra := mocks.NewMockHydra(ctrl)
	mockLogger := logger.DefaultLogger()
	mockValidator := mocks.NewMockValidator(ctrl)
	handler := &operatorHandle{
		OperatorManager: mockOperatorManager,
		Hydra:           mockHydra,
		Logger:          mockLogger,
		Validator:       mockValidator,
	}
	path := "/operator/info"
	applicationJSON := "application/json"
	mockOperatorID := "b2d8baf0-e31f-4cac-851d-30ad8c2e4722"
	req := &interfaces.OperatorEditReq{
		OperatorID: mockOperatorID,
		OperatorInfoEdit: &interfaces.OperatorInfoEdit{
			Type:          "basic",
			ExecutionMode: "sync",
			Category:      "other_category",
			Source:        "unknown",
		},
		OperatorExecuteControl: &interfaces.OperatorExecuteControl{
			Timeout: 3000,
			RetryPolicy: interfaces.OperatorRetryPolicy{
				MaxAttempts:   3,
				InitialDelay:  1000,
				BackoffFactor: 2,
				MaxDelay:      6000,
				RetryConditions: interfaces.RetryConditions{
					StatusCode: nil,
					ErrorCodes: nil,
				},
			},
		},
		UserID: "mock",
	}
	Convey("TestOperatorEdit:参数校验", t, func() {
		Convey("参数为空，校验失败", func() {
			recorder := mockPostRequest(path, applicationJSON,
				http.NoBody, handler.OperatorEdit)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("参数不为空，校验失败，id、version不合法", func() {
			recorder := mockPostRequest(path, applicationJSON,
				bytes.NewBufferString(`{"operator_id": "mockoperator_id","version": "mock_version"}`), handler.OperatorEdit)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("参数校验成功，默认参数注入成功", func() {
			fmt.Println(utils.ObjectToJSON(req))
			req.MetadataType = interfaces.MetadataTypeAPI
			mockOperatorManager.EXPECT().EditOperator(gomock.Any(), gomock.Any()).Return(&interfaces.OperatorEditResp{}, nil)
			recorder := mockPostRequest(path, applicationJSON,
				bytes.NewBufferString(`{"operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "416278e0-2816-4537-a974-fbe46a3a7720", "metadata_type": "openapi"}`), func(c *gin.Context) {
					ctx := c.Request.Context()
					c.Request = c.Request.WithContext(ctx)
					handler.OperatorEdit(c)
				})
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})
		Convey("请求失败", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			mockOperatorManager.EXPECT().EditOperator(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock error"))
			recorder := mockPostRequest(path, applicationJSON,
				bytes.NewBufferString(utils.ObjectToJSON(req)), func(c *gin.Context) {
					ctx := c.Request.Context()
					c.Request = c.Request.WithContext(ctx)
					handler.OperatorEdit(c)
				})
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestOperatorDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
	mockHydra := mocks.NewMockHydra(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	handler := &operatorHandle{
		OperatorManager: mockOperatorManager,
		Hydra:           mockHydra,
		Logger:          logger.DefaultLogger(),
		Validator:       mockValidator,
	}
	path := "/operator/delete"
	applicationJSON := "application/json"
	Convey("TestOperatorDelete:参数校验", t, func() {
		Convey("参数为空，校验失败", func() {
			recorder := mockPostRequest(path, applicationJSON,
				http.NoBody, handler.OperatorDelete)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("用户信息获取失败", func() {
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(nil, errors.New("mock error"))
			recorder := mockPostRequest(path, applicationJSON,
				bytes.NewBufferString(`[{"operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "416278e0-2816-4537-a974-fbe46a3a7720"}]`), handler.OperatorDelete)
			fmt.Println(recorder.Body.String())
		})
		Convey("删除失败", func() {
			mockOperatorManager.EXPECT().DeleteOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(&interfaces.TokenInfo{}, nil).Times(1)
			recorder := mockPostRequest(path, applicationJSON,
				bytes.NewBufferString(`[{"operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "416278e0-2816-4537-a974-fbe46a3a7720"}]`),
				func(c *gin.Context) {
					ctx := c.Request.Context()
					c.Request = c.Request.WithContext(ctx)
					handler.OperatorDelete(c)
				})
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("删除成功", func() {
			mockOperatorManager.EXPECT().DeleteOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(&interfaces.TokenInfo{}, nil).Times(1)
			recorder := mockPostRequest(path, applicationJSON,
				bytes.NewBufferString(`[{"operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "416278e0-2816-4537-a974-fbe46a3a7720"}]`),
				func(c *gin.Context) {
					ctx := c.Request.Context()
					c.Request = c.Request.WithContext(ctx)
					handler.OperatorDelete(c)
				})
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})
	})
}
