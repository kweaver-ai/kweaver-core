package operator

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestDebugOperator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	Convey("TestDebugOperator:Debug 单元测试", t, func() {
		mockDBOperatorManager := mocks.NewMockIOperatorRegisterDB(ctrl)
		// mockDBAPIMetadataManager := mocks.NewMockIAPIMetadataDB(ctrl)
		mockDBTx := mocks.NewMockDBTx(ctrl)
		mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
		mockUserMgnt := mocks.NewMockUserManagement(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)
		mockProxy := mocks.NewMockProxyHandler(ctrl)
		mockOpReleaseDB := mocks.NewMockIOperatorReleaseDB(ctrl)
		mockOpReleaseHistoryDB := mocks.NewMockIOperatorReleaseHistoryDB(ctrl)
		mockIntCompConfigSvc := mocks.NewMockIIntCompConfigService(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		mockAuditLog := mocks.NewMockLogModelOperator[*metric.AuditLogBuilderParams](ctrl)
		mockMetadataService := mocks.NewMockIMetadataService(ctrl)
		operator := &operatorManager{
			Logger:             logger.DefaultLogger(),
			DBOperatorManager:  mockDBOperatorManager,
			DBTx:               mockDBTx,
			CategoryManager:    mockCategoryManager,
			UserMgnt:           mockUserMgnt,
			Validator:          mockValidator,
			Proxy:              mockProxy,
			OpReleaseDB:        mockOpReleaseDB,
			OpReleaseHistoryDB: mockOpReleaseHistoryDB,
			IntCompConfigSvc:   mockIntCompConfigSvc,
			AuthService:        mockAuthService,
			AuditLog:           mockAuditLog,
			MetadataService:    mockMetadataService,
		}
		operatorDB := &model.OperatorRegisterDB{}
		accessor := &interfaces.AuthAccessor{}
		req := &interfaces.DebugOperatorReq{}
		metadata := &model.APIMetadataDB{
			ServerURL: "http://localhost:8080",
			Path:      "/api/v1/debug",
			Method:    "GET",
		}
		Convey("TestDebugOperator: 查询算子信息失败（DB）", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil, errors.New("mock SelectByOperatorID error")).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.Code, ShouldEqual, "Public.InternalServerError")
		})
		Convey("TestDebugOperator: 算子不存在", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil, nil).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			So(httpErr.Code, ShouldContainSubstring, myErr.ErrExtOperatorNotFound.String())
		})
		Convey("TestDebugOperator: 获取accessor信息失败", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("mock GetAccessor error")).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 检查执行权限失败", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock CheckExecutePermission error")).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 元数据格式不支持Debug", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 调试版本与线上版本不一致,读取历史版本记录失败（DB）", func() {
			operatorDB.MetadataType = "openapi"
			operatorDB.MetadataVersion = "1.0.1"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByOperatorID error")).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 调试版本与线上版本不一致,历史版本记录不存在", func() {
			operatorDB.MetadataType = "openapi"
			operatorDB.MetadataVersion = "1.0.1"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 调试版本与线上版本不一致,解析历史版本信息报错", func() {
			operatorDB.MetadataType = string(interfaces.MetadataTypeFunc)
			operatorDB.MetadataVersion = "1.0.1"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseHistoryDB{}, nil).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 执行模式不支持", func() {
			operatorDB.MetadataType = string(interfaces.MetadataTypeAPI)
			operatorDB.MetadataVersion = "1.0.1"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseHistoryDB{
				OpRelease: `{"execution_mode":"async"}`,
			}, nil).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 获取元数据失败", func() {
			operatorDB.MetadataType = "openapi"
			operatorDB.MetadataVersion = "1.0.1"
			req.Version = operatorDB.MetadataVersion
			operatorDB.ExecutionMode = "sync"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("CheckMetadataExists")).Times(1)
			// mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByVersion error")).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 元数据不存在", func() {
			operatorDB.MetadataType = "openapi"
			operatorDB.MetadataVersion = "1.0.1"
			req.Version = operatorDB.MetadataVersion
			operatorDB.ExecutionMode = "sync"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: 代理请求失败", func() {
			operatorDB.MetadataType = "openapi"
			operatorDB.MetadataVersion = "1.0.1"
			req.Version = operatorDB.MetadataVersion
			operatorDB.ExecutionMode = "sync"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadata, nil).Times(1)
			mockProxy.EXPECT().HandlerRequest(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock HandlerRequest error")).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDebugOperator: Deubg调试成功，记录审计日志", func() {
			operatorDB.MetadataType = "openapi"
			operatorDB.MetadataVersion = "1.0.1"
			req.Version = operatorDB.MetadataVersion
			operatorDB.ExecutionMode = "sync"
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadata, nil).Times(1)
			mockProxy.EXPECT().HandlerRequest(gomock.Any(), gomock.Any()).Return(&interfaces.HTTPResponse{
				StatusCode: 200,
				Body:       []byte("mock Body"),
			}, nil).Times(1)
			mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any()).Times(1)
			_, err := operator.DebugOperator(context.TODO(), req)
			So(err, ShouldBeNil)
			time.Sleep(100 * time.Millisecond)
		})
	})
}
