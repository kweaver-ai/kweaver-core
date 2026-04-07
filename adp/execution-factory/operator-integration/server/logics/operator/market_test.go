package operator

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestQueryOperatorMarketDetail(t *testing.T) {
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
	Convey("TestQueryOperatorMarketDetail:算子市场查看详情", t, func() {
		req := &interfaces.OperatorMarketDetailReq{}
		releaseDB := &model.OperatorReleaseDB{
			ExecuteControl: `{}`,
			ExtendInfo:     `{}`,
		}
		accessor := &interfaces.AuthAccessor{}
		metadataDB := &model.APIMetadataDB{
			APISpec: `{}`,
		}
		ctx := context.Background()
		ctx = common.SetPublicAPIToCtx(ctx, true)
		Convey("检查算子是否存在报错（db）", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("SelectByOpID")).Times(1)
			_, err := operator.QueryOperatorMarketDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("算子不存在", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.QueryOperatorMarketDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("公开接口: 获取accessor信息失败", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, releaseDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor")).Times(1)
			_, err := operator.QueryOperatorMarketDetail(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("公开接口: 没有公开访问权限", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, releaseDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("CheckPublicAccessPermission")).Times(1)
			_, err := operator.QueryOperatorMarketDetail(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("获取元数据失败（db）", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, releaseDB, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("CheckMetadataExists")).Times(1)
			// mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("SelectByVersion")).Times(1)
			_, err := operator.QueryOperatorMarketDetail(ctx, req)
			So(err, ShouldNotBeNil)
		})
		Convey("元数据信息不存在", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, releaseDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.QueryOperatorMarketDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("组装数据失败", func() {
			releaseDB.ExtendInfo = "ss"
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, releaseDB, nil).Times(1)
			// mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadataDB, nil).Times(1)
			_, err := operator.QueryOperatorMarketDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("获取用户名失败", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, releaseDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadataDB, nil).Times(1)
			// mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, metadataDB, nil).Times(1)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").Times(1)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetUsersName")).Times(1)
			_, err := operator.QueryOperatorMarketDetail(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("获取历史信息详情成功", func() {
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, releaseDB, nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, metadataDB, nil).Times(1)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").Times(1)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil).Times(1)
			_, err := operator.QueryOperatorMarketDetail(context.TODO(), req)
			So(err, ShouldBeNil)
		})
	})
}

func TestQueryOperatorMarketList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
	operator := &operatorManager{
		Logger:                logger.DefaultLogger(),
		DBOperatorManager:     mockDBOperatorManager,
		DBTx:                  mockDBTx,
		CategoryManager:       mockCategoryManager,
		UserMgnt:              mockUserMgnt,
		Validator:             mockValidator,
		Proxy:                 mockProxy,
		OpReleaseDB:           mockOpReleaseDB,
		OpReleaseHistoryDB:    mockOpReleaseHistoryDB,
		IntCompConfigSvc:      mockIntCompConfigSvc,
		AuthService:           mockAuthService,
		AuditLog:              mockAuditLog,
		MetadataService:       mockMetadataService,
		BusinessDomainService: mockBusinessDomainService,
	}
	Convey("TestQueryOperatorMarketList: 算子市场查询列表", t, func() {
		req := &interfaces.PageQueryOperatorMarketReq{
			Name:         "test",
			CreateUser:   "CreateUser",
			ReleaseUser:  "ReleaseUser",
			OperatorType: interfaces.OperatorTypeBase,
			Status:       interfaces.BizStatusPublished,
			PageSize:     10,
			Page:         1,
		}
		// ctx := context.Background()
		// ctx = common.SetPublicAPIToCtx(ctx, true)
		// opID := "OpID0"
		// releaseList := []*model.OperatorReleaseDB{
		// 	{
		// 		ID:              0,
		// 		OpID:            opID,
		// 		Name:            "Name0",
		// 		MetadataVersion: "MetadataVersion0",
		// 		MetadataType:    string(interfaces.MetadataTypeAPI),
		// 		Status:          interfaces.BizStatusPublished.String(),
		// 		OperatorType:    string(interfaces.OperatorTypeBase),
		// 		Category:        interfaces.CategoryTypeOther.String(),
		// 	},
		// 	{
		// 		ID:              1,
		// 		OpID:            opID,
		// 		Name:            "Name1",
		// 		MetadataVersion: "MetadataVersion1",
		// 		MetadataType:    string(interfaces.MetadataTypeAPI),
		// 		Status:          interfaces.BizStatusPublished.String(),
		// 		OperatorType:    string(interfaces.OperatorTypeBase),
		// 		Category:        interfaces.CategoryTypeOther.String(),
		// 		ExecuteControl:  `{}`,
		// 	},
		// 	{
		// 		ID:              2,
		// 		OpID:            opID,
		// 		Name:            "Name2",
		// 		MetadataVersion: "MetadataVersion2",
		// 		MetadataType:    string(interfaces.MetadataTypeAPI),
		// 		Status:          interfaces.BizStatusPublished.String(),
		// 		OperatorType:    string(interfaces.OperatorTypeBase),
		// 		Category:        interfaces.CategoryTypeOther.String(),
		// 		ExecuteControl:  `{}`,
		// 		ExtendInfo:      `{}`,
		// 	},
		// }
		// metadataList := []*model.APIMetadataDB{
		// 	{
		// 		ID:      1,
		// 		Version: "MetadataVersion1",
		// 	},
		// 	{
		// 		ID:      2,
		// 		Version: "MetadataVersion2",
		// 		APISpec: `{}`,
		// 	},
		// }
		// accessor := &interfaces.AuthAccessor{}
		Convey("分类不合法", func() {
			req.Category = "invalid"
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false).Times(1)
			_, err := operator.QueryOperatorMarketList(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		// Convey("外部接口: 拉取数据失败（db）", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]string{}, nil).Times(1)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectByWhereClause"))
		// 	_, err := operator.QueryOperatorMarketList(ctx, req)
		// 	So(err, ShouldNotBeNil)
		// })
		// Convey("外部接口: 获取accessor信息失败", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]string{}, nil).Times(1)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(releaseList, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor"))
		// 	_, err := operator.QueryOperatorMarketList(ctx, req)
		// 	So(err, ShouldNotBeNil)
		// })
		// Convey("外部接口: 检查公开访问权限，资源过滤失败", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(releaseList, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockAuthService.EXPECT().ResourceListIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypePublicAccess).Return(nil, mocks.MockFuncErr("ResourceListIDs")).Times(1)
		// 	_, err := operator.QueryOperatorMarketList(ctx, req)
		// 	So(err, ShouldNotBeNil)
		// })
		// Convey("外部接口: 匹配到的资源为空", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(releaseList, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockAuthService.EXPECT().ResourceListIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypePublicAccess).Return([]string{}, nil).Times(1)
		// 	_, err := operator.QueryOperatorMarketList(ctx, req)
		// 	So(err, ShouldBeNil)
		// })
		// Convey("外部接口: 获取元数据失败（db）", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(releaseList, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockAuthService.EXPECT().ResourceListIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypePublicAccess).Return([]string{opID}, nil).Times(1)
		// 	mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectListByVersion"))
		// 	_, err := operator.QueryOperatorMarketList(ctx, req)
		// 	So(err, ShouldNotBeNil)
		// 	httpErr := &myErr.HTTPError{}
		// 	So(errors.As(err, &httpErr), ShouldBeTrue)
		// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// })
		// Convey("内部接口: 根据条件统计失败（db）", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().CountByWhereClause(gomock.Any(), gomock.Any()).Return(int64(0), mocks.MockFuncErr("CountByWhereClause"))
		// 	_, err := operator.QueryOperatorMarketList(context.TODO(), req)
		// 	So(err, ShouldNotBeNil)
		// 	httpErr := &myErr.HTTPError{}
		// 	So(errors.As(err, &httpErr), ShouldBeTrue)
		// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// })
		// Convey("内部接口: 未匹配到数据", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().CountByWhereClause(gomock.Any(), gomock.Any()).Return(int64(0), nil)
		// 	_, err := operator.QueryOperatorMarketList(context.TODO(), req)
		// 	So(err, ShouldBeNil)
		// })
		// Convey("内部接口: 条件查询失败（db）", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().CountByWhereClause(gomock.Any(), gomock.Any()).Return(int64(1), nil)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectByWhereClause"))
		// 	_, err := operator.QueryOperatorMarketList(context.TODO(), req)
		// 	So(err, ShouldNotBeNil)
		// 	httpErr := &myErr.HTTPError{}
		// 	So(errors.As(err, &httpErr), ShouldBeTrue)
		// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// })
		// Convey("内部接口: 未匹配到元数据, 解析数据失败, 解析元数据失败, 获取用户名失败", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().CountByWhereClause(gomock.Any(), gomock.Any()).Return(int64(len(releaseList)), nil)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(releaseList, nil)
		// 	mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return(metadataList, nil)
		// 	mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return(interfaces.CategoryTypeOther.String()).Times(1)
		// 	mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetUsersName"))
		// 	_, err := operator.QueryOperatorMarketList(context.TODO(), req)
		// 	So(err, ShouldNotBeNil)
		// })
		// Convey("内部接口: 获取市场列表成功", func() {
		// 	req.Category = interfaces.CategoryTypeOther
		// 	req.All = true
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
		// 	mockOpReleaseDB.EXPECT().CountByWhereClause(gomock.Any(), gomock.Any()).Return(int64(len(releaseList)), nil)
		// 	mockOpReleaseDB.EXPECT().SelectByWhereClause(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(releaseList, nil)
		// 	mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return(metadataList, nil)
		// 	mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return(interfaces.CategoryTypeOther.String()).Times(1)
		// 	mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
		// 	resp, err := operator.QueryOperatorMarketList(context.TODO(), req)
		// 	So(err, ShouldBeNil)
		// 	So(resp, ShouldNotBeNil)
		// 	So(len(resp.Data), ShouldEqual, 1)
		// })
	})
}
