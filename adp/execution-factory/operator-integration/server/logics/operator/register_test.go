package operator

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestRegisterOperatorByOpenAPI(t *testing.T) {
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
	req := &interfaces.OperatorRegisterReq{
		OperatorInfo: &interfaces.OperatorInfo{
			Category: "test_category",
		},
	}
	accessor := &interfaces.AuthAccessor{}
	Convey("TestRegisterOperatorByOpenAPI:算子注册", t, func() {
		Convey("权限检查：获取accessor信息失败", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor")).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("权限检查：无新建权限", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("CheckCreatePermission")).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("异步：不支持添加为数据源", func() {
			isDataSource := true
			req.OperatorInfo = &interfaces.OperatorInfo{
				IsDataSource:  &isDataSource,
				ExecutionMode: interfaces.ExecutionModeAsync,
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("流式：不支持添加为数据源", func() {
			isDataSource := true
			req.OperatorInfo = &interfaces.OperatorInfo{
				IsDataSource:  &isDataSource,
				ExecutionMode: interfaces.ExecutionModeStream,
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("同步添加为数据源，分类不存在", func() {
			isDataSource := true
			req.OperatorInfo = &interfaces.OperatorInfo{
				IsDataSource:  &isDataSource,
				ExecutionMode: interfaces.ExecutionModeSync,
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("元数据类型不存在", func() {
			req.MetadataType = ""
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("解析OpenAPI类型元数据", func() {
			req.MetadataType = interfaces.MetadataTypeAPI
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("ParseMetadata")).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("解析Function类型元数据, Items 校验未通过", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.IMetadataDB{}, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidateOperatorImportCount")).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("直接发布多个算子", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{
				&model.FunctionMetadataDB{},
				&model.FunctionMetadataDB{},
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			_, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.Code, ShouldContainSubstring, myErr.ErrExtOperatorDirectPublishErr.String())
		})
		Convey("注册算子-validateOperator：解析失败", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{
				ErrMessage: "ParseMetadata error",
			}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("注册算子-validateOperator：算子名称为空", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidateOperatorName")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("注册算子-validateOperator：算子描述为空", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidateOperatorDesc")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("checkDuplicateName-检查是否重名：获取算子名称失败", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("CheckDuplicateName")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("checkDuplicateName-检查是否重名：存在重名", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("获取锁失败（GetTx）", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, mocks.MockFuncErr("GetTx")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("注册元数据失败", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("RegisterMetadata")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("添加算子信息失败", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("InsertOperator")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("关联业务域失败", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				mocks.MockFuncErr("AssociateResource")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("触发新建策略失败", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("CreateOwnerPolicy")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("检查是否存在已发布版本失败(db)", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("SelectByOpID")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("存在已发布版本，更新算子(db)", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
			mockOpReleaseDB.EXPECT().UpdateByOpID(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("UpdateByOpID")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("不存在已发布版本，新建算子(db)", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockOpReleaseDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("Insert")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("添加历史记录失败（db）", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockOpReleaseDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("Insert")).Times(1)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusFailed)
		})
		Convey("添加成功,发送审计日志", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.DirectPublish = true
			metadataDBs := []interfaces.IMetadataDB{&model.FunctionMetadataDB{}}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			mockOpReleaseDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any()).Times(2)
			resp, err := operator.RegisterOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].Status, ShouldEqual, interfaces.ResultStatusSuccess)
			time.Sleep(100 * time.Millisecond)
		})
	})
}

func TestUpdateOperatorByOpenAPI(t *testing.T) {
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
	Convey("TestUpdateOperatorByOpenAPI: 更新算子", t, func() {
		req := &interfaces.OperatorUpdateReq{
			OperatorRegisterReq: &interfaces.OperatorRegisterReq{
				OperatorInfo:           &interfaces.OperatorInfo{},
				OperatorExecuteControl: &interfaces.OperatorExecuteControl{},
			},
		}
		Convey("解析算子失败(分类不存在)", func() {
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false).Times(1)
			_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("不允许编辑多个算子", func() {
			req.MetadataType = interfaces.MetadataTypeAPI
			metadataDBs := []interfaces.IMetadataDB{
				&model.APIMetadataDB{},
				&model.APIMetadataDB{},
			}
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("待编辑元数据为空", func() {
			req.MetadataType = interfaces.MetadataTypeAPI
			metadataDBs := []interfaces.IMetadataDB{}
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("查询算子失败（db）", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.FunctionInput = &interfaces.FunctionInput{}
			metadataDBs := []interfaces.IMetadataDB{
				&model.FunctionMetadataDB{},
			}
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("SelectByOperatorID")).Times(1)
			_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("算子不存在", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.FunctionInput = &interfaces.FunctionInput{}
			metadataDBs := []interfaces.IMetadataDB{
				&model.FunctionMetadataDB{},
			}
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("名字不合法", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.FunctionInput = &interfaces.FunctionInput{}
			metadataDBs := []interfaces.IMetadataDB{
				&model.FunctionMetadataDB{
					Summary: "func name",
				},
			}
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidateOperatorName")).Times(1)
			_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("描述不合法", func() {
			req.MetadataType = interfaces.MetadataTypeFunc
			req.FunctionInput = &interfaces.FunctionInput{}
			metadataDBs := []interfaces.IMetadataDB{
				&model.FunctionMetadataDB{
					Description: "Mock Description",
				},
			}
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidateOperatorDesc")).Times(1)
			_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("权限校验", func() {
			Convey("获取accessor失败", func() {
				req.MetadataType = interfaces.MetadataTypeFunc
				req.FunctionInput = &interfaces.FunctionInput{}
				metadataDBs := []interfaces.IMetadataDB{
					&model.FunctionMetadataDB{},
				}
				mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
				mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
				mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor")).Times(1)
				_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("直接发布：校验编辑、发布权限", func() {
				req.MetadataType = interfaces.MetadataTypeFunc
				req.FunctionInput = &interfaces.FunctionInput{}
				metadataDBs := []interfaces.IMetadataDB{
					&model.FunctionMetadataDB{},
				}
				req.DirectPublish = true
				mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
				mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
				mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
				mockAuthService.EXPECT().MultiCheckOperationPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeModify, interfaces.AuthOperationTypePublish).Return(mocks.MockFuncErr("MultiCheckOperationPermission")).Times(1)
				_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("检查编辑权限成功，获取元数据失败", func() {
				req.MetadataType = interfaces.MetadataTypeFunc
				req.FunctionInput = &interfaces.FunctionInput{}
				metadataDBs := []interfaces.IMetadataDB{
					&model.FunctionMetadataDB{},
				}
				req.DirectPublish = false
				mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
				mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
				mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
				mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("CheckMetadataExists")).Times(1)
				_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("元数据不存在", func() {
				req.MetadataType = interfaces.MetadataTypeFunc
				req.FunctionInput = &interfaces.FunctionInput{}
				metadataDBs := []interfaces.IMetadataDB{
					&model.FunctionMetadataDB{},
				}
				req.DirectPublish = false
				mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
				mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
				mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
				mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			})
			// Convey("组装待更新数据", func() {
			// 	Convey("元数据不存在", func() {
			// 		req.MetadataType = interfaces.MetadataTypeFunc
			// 		req.FunctionInput = &interfaces.FunctionInput{}
			// 		metadataDBs := []interfaces.IMetadataDB{
			// 			&model.FunctionMetadataDB{},
			// 		}
			// 		req.DirectPublish = false
			// 		mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(true).Times(1)
			// 		mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataDBs, nil).Times(1)
			// 		mockValidator.EXPECT().ValidateOperatorImportCount(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
			// 		mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
			// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// 		mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
			// 		_, err := operator.UpdateOperatorByOpenAPI(context.TODO(), req, "")
			// 		So(err, ShouldNotBeNil)
			// 		httpErr := &myErr.HTTPError{}
			// 		So(errors.As(err, &httpErr), ShouldBeTrue)
			// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			// 	})
			// })
		})
	})
}
