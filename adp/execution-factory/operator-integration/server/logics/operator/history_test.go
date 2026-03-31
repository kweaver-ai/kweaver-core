package operator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestQueryOperatorHistoryDetail(t *testing.T) {
	Convey("TestQueryOperatorHistoryDetail: 查询操作历史详情", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockDBOperatorManager := mocks.NewMockIOperatorRegisterDB(ctrl)
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
		req := &interfaces.OperatorHistoryDetailReq{}
		accessor := &interfaces.AuthAccessor{}
		historyDB := &model.OperatorReleaseHistoryDB{}
		metadataDB := &model.APIMetadataDB{}
		ctx := context.Background()
		ctx = common.SetPublicAPIToCtx(ctx, true)
		Convey("TestQueryOperatorHistoryDetail: 检查算子是否存在报错（DB）", func() {
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByOpIDAndMetdata error")).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("TestQueryOperatorHistoryDetail: 历史版本不存在", func() {
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("TestQueryOperatorHistoryDetail: 获取accessor信息失败", func() {
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock GetAccessor error")).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: 检查权限报错", func() {
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView,
				interfaces.AuthOperationTypePublicAccess).Return(false, errors.New("mock OperationCheckAny error")).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: 权限检验未通过", func() {
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView,
				interfaces.AuthOperationTypePublicAccess).Return(false, nil).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: 解析算子数据失败", func() {
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView,
				interfaces.AuthOperationTypePublicAccess).Return(true, nil).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: 获取算子元数据报错", func() {
			historyDB.OpRelease = `{"metadata_version":"1.0.1"}`
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("GetMetadataByVersion")).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: 元数据不存在", func() {
			historyDB.OpRelease = `{"metadata_version":"1.0.1"}`
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: assembleReleaseResult 组装数据报错", func() {
			historyDB.OpRelease = utils.ObjectToJSON(&model.OperatorReleaseDB{
				ExecuteControl: `{}`,
				ExtendInfo:     "SS",
			})
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadataDB, nil).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: 获取用户名报错", func() {
			historyDB.OpRelease = utils.ObjectToJSON(&model.OperatorReleaseDB{
				ExecuteControl: `{}`,
				ExtendInfo:     `{}`,
			})
			fmt.Println(historyDB.OpRelease)
			metadataDB.APISpec = `{}`
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadataDB, nil).Times(1)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("mock Name").Times(1)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock GetUserInfo error")).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("TestQueryOperatorHistoryDetail: 获取历史详情成功", func() {
			historyDB.OpRelease = utils.ObjectToJSON(&model.OperatorReleaseDB{
				ExecuteControl: `{}`,
				ExtendInfo:     `{}`,
			})
			fmt.Println(historyDB.OpRelease)
			metadataDB.APISpec = `{}`
			mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, historyDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadataDB, nil).Times(1)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("mock Name").Times(1)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil).Times(1)
			_, err := operator.QueryOperatorHistoryDetail(context.TODO(), req)
			So(err, ShouldBeNil)
		})
	})
}

func TestQueryOperatorHistoryList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDBOperatorManager := mocks.NewMockIOperatorRegisterDB(ctrl)
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
	Convey("TestQueryOperatorHistoryList:获取历史记录列表", t, func() {
		req := &interfaces.OperatorHistoryListReq{}
		accessor := &interfaces.AuthAccessor{}
		histories := []*model.OperatorReleaseHistoryDB{
			{
				ID:              0,
				OpID:            "OpID",
				MetadataVersion: "MetadataVersion0",
				MetadataType:    string(interfaces.MetadataTypeAPI),
			},
			{
				ID:              1,
				OpID:            "OpID",
				MetadataVersion: "MetadataVersion1",
				MetadataType:    string(interfaces.MetadataTypeAPI),
				OpRelease:       `{}`,
			},
			{
				ID:              2,
				OpID:            "OpID",
				MetadataVersion: "MetadataVersion2",
				MetadataType:    string(interfaces.MetadataTypeAPI),
				OpRelease:       `{}`,
			},
			{
				ID:              3,
				OpID:            "OpID",
				MetadataVersion: "MetadataVersion3",
				MetadataType:    string(interfaces.MetadataTypeFunc),
				OpRelease: utils.ObjectToJSON(&model.OperatorReleaseDB{
					ExecuteControl: `{}`,
					ExtendInfo:     `{}`,
				}),
			},
		}

		metadataList := []interfaces.IMetadataDB{
			&model.APIMetadataDB{
				ID:      0,
				Version: "MetadataVersion0",
			},
			&model.APIMetadataDB{
				ID:      1,
				Version: "MetadataVersion",
			},
			&model.APIMetadataDB{
				ID:      2,
				Version: "MetadataVersion2",
			},
			&model.APIMetadataDB{
				ID:      3,
				Version: "MetadataVersion3",
				APISpec: `{}`,
			},
		}
		Convey("查询算子信息失败（DB）", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByOperatorID error")).Times(1)
			_, err := operator.QueryOperatorHistoryList(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("算子不存在", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.QueryOperatorHistoryList(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("外部接口检查权限: 获取accessor失败", func() {
			ctx := context.Background()
			ctx = common.SetPublicAPIToCtx(ctx, true)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock GetAccessor error")).Times(1)
			_, err := operator.QueryOperatorHistoryList(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("外部接口检查权限: 检查权限报错", func() {
			ctx := context.Background()
			ctx = common.SetPublicAPIToCtx(ctx, true)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(false,
				errors.New("mock OperationCheckAny error")).Times(1)
			_, err := operator.QueryOperatorHistoryList(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("外部接口检查权限: 没有查看权限或者公开访问权限", func() {
			ctx := context.Background()
			ctx = common.SetPublicAPIToCtx(ctx, true)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(false,
				nil).Times(1)
			_, err := operator.QueryOperatorHistoryList(ctx, req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
		})
		Convey("获取历史数据失败（db）", func() {
			ctx := context.Background()
			ctx = common.SetPublicAPIToCtx(ctx, true)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(true,
				nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock SelectByOpID error")).Times(1)
			_, err := operator.QueryOperatorHistoryList(ctx, req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("历史记录为空", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			_, err := operator.QueryOperatorHistoryList(context.TODO(), req)
			So(err, ShouldBeNil)
		})
		Convey("获取元数据信息失败", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(histories, nil).Times(1)
			// mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock SelectListByVersion error")).Times(1)
			mockMetadataService.EXPECT().BatchGetMetadataBySourceIDs(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("BatchGetMetadataBySourceIDs")).Times(1)
			_, err := operator.QueryOperatorHistoryList(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("元数据信息不存在, 解析release失败，组装数据失败，组装数据成功，获取用户名报错", func() {
			sourceIDToMetadataMap := map[string]interfaces.IMetadataDB{}
			for _, metadata := range metadataList {
				sourceIDToMetadataMap[metadata.GetVersion()] = metadata
			}
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(histories, nil).Times(1)
			mockMetadataService.EXPECT().BatchGetMetadataBySourceIDs(gomock.Any(), gomock.Any()).Return(sourceIDToMetadataMap, nil).Times(1)
			// mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return(metadataList, nil).Times(1)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("MOCK").Times(1)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock GetUsersName error")).Times(1)
			_, err := operator.QueryOperatorHistoryList(context.TODO(), req)
			So(err, ShouldNotBeNil)
			fmt.Println(err)
		})
		Convey("获取元数据列表成功", func() {
			sourceIDToMetadataMap := map[string]interfaces.IMetadataDB{}
			for _, metadata := range metadataList {
				sourceIDToMetadataMap[metadata.GetVersion()] = metadata
			}
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(histories, nil).Times(1)
			mockMetadataService.EXPECT().BatchGetMetadataBySourceIDs(gomock.Any(), gomock.Any()).Return(sourceIDToMetadataMap, nil).Times(1)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("MOCK").Times(1)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil).Times(1)
			resp, err := operator.QueryOperatorHistoryList(context.TODO(), req)
			So(err, ShouldBeNil)
			So(len(resp), ShouldEqual, 1)
		})
	})
}
