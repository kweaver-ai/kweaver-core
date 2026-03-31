package operator

import (
	"context"
	"database/sql"
	"fmt"
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

func TestImport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	Convey("TestImport:算子导入", t, func() {
		mockDBOperatorManager := mocks.NewMockIOperatorRegisterDB(ctrl)
		// mockDBAPIMetadataManager := mocks.NewMockIAPIMetadataDB(ctrl)
		mockDBTx := mocks.NewMockDBTx(ctrl)
		mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
		mockUserMgnt := mocks.NewMockUserManagement(ctrl)
		mockProxy := mocks.NewMockProxyHandler(ctrl)
		mockOpReleaseDB := mocks.NewMockIOperatorReleaseDB(ctrl)
		mockOpReleaseHistoryDB := mocks.NewMockIOperatorReleaseHistoryDB(ctrl)
		mockIntCompConfigSvc := mocks.NewMockIIntCompConfigService(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		mockAuditLog := mocks.NewMockLogModelOperator[*metric.AuditLogBuilderParams](ctrl)
		mockMetadataService := mocks.NewMockIMetadataService(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)
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
		metadata := &interfaces.MetadataInfo{
			Version:    "versioin_1",
			Summary:    "mock test data",
			ServerURL:  "http://127.0.0.1",
			Method:     http.MethodGet,
			Path:       "/test.path",
			CreateTime: time.Now().UnixNano(),
			CreateUser: "CreateUser",
			UpdateTime: time.Now().UnixNano(),
			UpdateUser: "UpdateUser",
			APISpec:    &interfaces.APISpec{},
		}
		importData := &interfaces.OperatorImpexConfig{
			Configs: []*interfaces.OperatorImpexItem{
				{
					OperatorID:   "1",
					OperatorName: "aa",
					Status:       interfaces.BizStatusUnpublish,
					MetadataType: "openapi",
					Metadata:     metadata,
				},
			},
		}
		operatorDB := &model.OperatorRegisterDB{OperatorID: "1"}
		accessor := &interfaces.AuthAccessor{}
		Convey("导入校验: 导入数据为空", func() {
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, nil, "")
			So(err, ShouldNotBeNil)
		})
		Convey("导入校验: 导入数据长度为0", func() {
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, &interfaces.OperatorImpexConfig{}, "")
			So(err, ShouldNotBeNil)
		})
		Convey("检查是否重名: 查询同名已发布算子失败（db）", func() {
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false,
				nil, mocks.MockFuncErr("SelectByNameAndStatus"))
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("检查是否重名: 存在同名算子", func() {
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusConflict)
		})
		Convey("检查ID资源是否冲突: 查询报错（db）", func() {
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectByOperatorIDs"))
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("算子已经存在", func() {
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusConflict)
		})
		Convey("获取accessor信息失败", func() {
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor"))
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
		})
		Convey("无算子编辑权限", func() {
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("CheckModifyPermission"))
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
		})
		Convey("内置算子不允许编辑", func() {
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{{
				OperatorID: "1",
				IsInternal: true,
			}}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
		})
		Convey("导入检查: 名字不合法", func() {
			importData.Configs[0].OperatorName = " aa"
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidateOperatorName")).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
		})
		Convey("导入检查: 异步算子数据源", func() {
			isDataSource := true
			importData.Configs[0].OperatorInfo = &interfaces.OperatorInfo{
				IsDataSource:  &isDataSource,
				ExecutionMode: interfaces.ExecutionModeAsync,
			}
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("导入检查: 元数据类型不支持", func() {
			importData.Configs[0].MetadataType = "op"
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("导入检查: 元数据为空", func() {
			importData.Configs[0].Metadata = nil
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("导入检查: 元数据解析失败", func() {
			importData.Configs[0].Metadata = nil
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("导入检查: 元数据校验失败", func() {
			importData.Configs[0].Metadata = nil
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("导入检查: 描述长度不合法", func() {
			metadata.Description = mocks.MockDescription(1000)
			importData.Configs[0].Metadata = metadata
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("更新配置: 元数据类型不一致", func() {
			operatorDB.Status = interfaces.BizStatusOffline.String()
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("RegisterMetadata")).Times(1)
			// mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("InsertAPIMetadata")).Times(1)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			fmt.Println(err)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("更新配置: 已下架算子插入元数据失败", func() {
			operatorDB.Status = interfaces.BizStatusOffline.String()
			operatorDB.MetadataType = string(interfaces.MetadataTypeAPI)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil).Times(1)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("RegisterMetadata")).Times(1)
			// mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("InsertAPIMetadata")).Times(1)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			fmt.Println(err)
			// httpErr, ok := err.(*myErr.HTTPError)
			// So(ok, ShouldBeTrue)
			// So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("更新配置: 更新算子失败（db）", func() {
			operatorDB.Status = interfaces.BizStatusPublished.String()
			operatorDB.MetadataType = string(interfaces.MetadataTypeAPI)
			mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(true, operatorDB, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().RegisterMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
			mockDBOperatorManager.EXPECT().UpdateByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("UpdateByOperatorID")).Times(1)
			err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		// Convey("更新配置: 更新元数据成功, 通知资源变更失败", func() {
		// 	operatorDB.Status = interfaces.BizStatusUnpublish.String()
		// 	mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
		// 		gomock.Any()).Return(true, operatorDB, nil)
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
		// 	mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil)
		// 	mockDBAPIMetadataManager.EXPECT().UpdateByVersion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockDBOperatorManager.EXPECT().UpdateByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockAuthService.EXPECT().NotifyResourceChange(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("NotifyResourceChange"))
		// 	mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any())
		// 	err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
		// 	So(err, ShouldBeNil)
		// 	time.Sleep(100 * time.Millisecond)
		// })
		// Convey("更新配置: 更新元数据成功, 通知资源成功，发送审计日志", func() {
		// 	operatorDB.Status = interfaces.BizStatusUnpublish.String()
		// 	mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
		// 		gomock.Any()).Return(true, operatorDB, nil)
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{operatorDB}, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
		// 	mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil)
		// 	mockDBAPIMetadataManager.EXPECT().UpdateByVersion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockDBOperatorManager.EXPECT().UpdateByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockAuthService.EXPECT().NotifyResourceChange(gomock.Any(), gomock.Any()).Return(nil)
		// 	mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any())
		// 	err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeUpsert, importData, "")
		// 	So(err, ShouldBeNil)
		// 	time.Sleep(100 * time.Millisecond)
		// })
		// Convey("添加配置", func() {
		// 	operatorDB.OperatorID = "mock1"
		// 	mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(),
		// 		gomock.Any()).Return(false, nil, nil)
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{}, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
		// 	Convey("添加元数据失败（db）", func() {
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("",
		// 			mocks.MockFuncErr("InsertAPIMetadata"))
		// 		err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	})
		// 	Convey("添加算子记录失败（db）", func() {
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
		// 		mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("InsertOperator"))
		// 		err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	})

		// 	Convey("发布失败", func() {
		// 		importData.Configs[0].Status = interfaces.BizStatusPublished
		// 		operatorDB.Status = interfaces.BizStatusPublished.String()
		// 		mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
		// 		mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, mocks.MockFuncErr("SelectByOpID"))
		// 		err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	})
		// 	Convey("发布成功, 添加所有者权限策略失败", func() {
		// 		operatorDB.Status = interfaces.BizStatusPublished.String()
		// 		importData.Configs[0].Status = interfaces.BizStatusPublished
		// 		mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
		// 		mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil)
		// 		mockOpReleaseDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("CreateOwnerPolicy"))
		// 		mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any())
		// 		err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
		// 		So(err, ShouldBeNil)
		// 		time.Sleep(100 * time.Millisecond)
		// 	})
		// 	Convey("发布成功, 添加所有者权限策略成功，发送审计日志", func() {
		// 		operatorDB.Status = interfaces.BizStatusPublished.String()
		// 		importData.Configs[0].Status = interfaces.BizStatusPublished
		// 		mockDBOperatorManager.EXPECT().InsertOperator(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
		// 		mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil)
		// 		mockOpReleaseDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any())
		// 		err := operator.Import(context.TODO(), &sql.Tx{}, interfaces.ImportTypeCreate, importData, "")
		// 		So(err, ShouldBeNil)
		// 		time.Sleep(100 * time.Millisecond)
		// 	})
		// })
	})
}

func TestExport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	Convey("TestExport:算子导出", t, func() {
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
		mockFlowAutomation := mocks.NewMockFlowAutomation(ctrl)
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
			FlowAutomation:     mockFlowAutomation,
		}
		ids := []string{"1"}
		req := &interfaces.ExportReq{
			UserID: "1",
			IDs:    ids,
		}
		operatorList1 := []*model.OperatorRegisterDB{
			{
				OperatorID:   "1",
				Name:         "1",
				OperatorType: string(interfaces.OperatorTypeComposite),
				ExtendInfo:   "",
			},
		}
		// operatorList2 := []*model.OperatorRegisterDB{
		// 	{
		// 		OperatorID:     "2",
		// 		Name:           "2",
		// 		OperatorType:   string(interfaces.OperatorTypeBase),
		// 		ExtendInfo:     `{}`,
		// 		ExecuteControl: `{}`,
		// 	},
		// }
		accessor := &interfaces.AuthAccessor{}
		Convey("检查是否有查看权限:请求获取accessor报错", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor"))
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("检查是否有查看权限:请求过滤权限报错", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(nil, mocks.MockFuncErr("ResourceFilterIDs"))
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("检查是否有查看权限:没有查看权限", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return([]string{}, nil)
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
		})
		Convey("查询算子信息报错（DB）", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
				ids, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectByOperatorIDs"))
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("请求导出算子不存在", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
				ids, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return([]*model.OperatorRegisterDB{}, nil)
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("组合算子获取拓展信息失败", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
				ids, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(operatorList1, nil)
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("导出依赖组合算子配置: 请求FlowAutomation失败", func() {
			operatorList1[0].ExtendInfo = `{"dag_id":"1"}`
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
				ids, nil)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(operatorList1, nil)
			mockFlowAutomation.EXPECT().Export(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("Export"))
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		Convey("导出依赖组合算子配置: 导出以来配置成功，预检查通过", func() {
			operatorList1[0].ExtendInfo = `{"dag_id":"1"}`
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(2)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
				ids, nil).Times(2)
			mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(operatorList1, nil).Times(2)
			mockFlowAutomation.EXPECT().Export(gomock.Any(), gomock.Any()).Return(&interfaces.FlowAutomationExportResp{
				Configs:     []any{},
				OperatorIDs: []string{"2", "2", "1"},
			}, nil)
			_, err := operator.Export(context.TODO(), req)
			So(err, ShouldNotBeNil)
		})
		// Convey("获取元数据失败（db）", func() {
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
		// 	mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
		// 		ids, nil).Times(1)
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(operatorList2, nil).Times(1)
		// 	mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectListByVersion"))
		// 	_, err := operator.Export(context.TODO(), req)
		// 	So(err, ShouldNotBeNil)
		// 	httpErr, ok := err.(*myErr.HTTPError)
		// 	So(ok, ShouldBeTrue)
		// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// })
		// Convey("解析元数据失败", func() {
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
		// 	mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
		// 		ids, nil).Times(1)
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(operatorList2, nil).Times(1)
		// 	mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return([]*model.APIMetadataDB{
		// 		{
		// 			APISpec: "",
		// 		},
		// 	}, nil)
		// 	_, err := operator.Export(context.TODO(), req)
		// 	So(err, ShouldNotBeNil)
		// 	httpErr, ok := err.(*myErr.HTTPError)
		// 	So(ok, ShouldBeTrue)
		// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// })
		// Convey("导出成功", func() {
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil).Times(1)
		// 	mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthOperationTypeView).Return(
		// 		ids, nil).Times(1)
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorIDs(gomock.Any(), gomock.Any()).Return(operatorList2, nil).Times(1)
		// 	mockDBAPIMetadataManager.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return([]*model.APIMetadataDB{
		// 		{
		// 			APISpec: `{}`,
		// 		},
		// 	}, nil)
		// 	resp, err := operator.Export(context.TODO(), req)
		// 	So(err, ShouldBeNil)
		// 	So(resp, ShouldNotBeNil)
		// 	So(resp.Operator, ShouldNotBeNil)
		// 	So(len(resp.Operator.Configs), ShouldEqual, 1)
		// 	So(len(resp.Operator.CompositeConfigs), ShouldEqual, 0)
		// })
	})
}
