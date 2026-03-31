package operator

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

var errMock = errors.New("mock error")

func TestEditOperator(t *testing.T) {
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
	m := &operatorManager{
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
	mockOperatorID := "b2d8baf0-e31f-4cac-851d-30ad8c2e4722"
	mockReq := &interfaces.OperatorEditReq{
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
	// mockAPIEditReq := &interfaces.APIMetadataEdit{
	// 	Summary:     mockName,
	// 	Description: "Description",
	// 	Path:        "/path",
	// 	Method:      "GET",
	// 	ServerURL:   "http://127.0.0.1",
	// 	APISpec: &interfaces.APISpec{
	// 		Tags: []string{"tag1", "tag2"},
	// 		Parameters: []*interfaces.Parameter{
	// 			{
	// 				Name:        "param1",
	// 				In:          "path",
	// 				Description: "Description",
	// 				Required:    true,
	// 			},
	// 		},
	// 		Responses: []*interfaces.Response{
	// 			{
	// 				StatusCode:  "200",
	// 				Description: "Success",
	// 				Content:     openapi3.Content{},
	// 			},
	// 		},
	// 	},
	// }
	operatorDB := &model.OperatorRegisterDB{}
	Convey("TestEditOperator:算子编辑校验", t, func() {
		Convey("查询算子失败", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil, errMock)
			_, err := m.EditOperator(context.TODO(), &interfaces.OperatorEditReq{})
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("算子不存在", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil, nil)
			_, err := m.EditOperator(context.TODO(), &interfaces.OperatorEditReq{})
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("权限校验: 获取accessor信息失败", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, &model.OperatorRegisterDB{
					UpdateUser: "AAAA",
				}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock GetAccessor error")).Times(1)
			_, err := m.EditOperator(context.TODO(), mockReq)
			So(err, ShouldNotBeNil)
		})
		Convey("权限校验: 检查编辑权限失败", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, &model.OperatorRegisterDB{
					UpdateUser: "AAAA",
				}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock CheckModifyPermission error")).Times(1)
			_, err := m.EditOperator(context.TODO(), mockReq)
			So(err, ShouldNotBeNil)
		})
		Convey("OpenAPI类型解析元数据失败", func() {
			operatorDB.MetadataType = string(interfaces.MetadataTypeAPI)
			mockReq.OpenAPIInput = &interfaces.OpenAPIInput{
				Data: json.RawMessage(`{"openapi": "3.0.0"}`),
			}
			mockReq.MetadataType = interfaces.MetadataTypeAPI
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, operatorDB, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("ParseMetadata")).Times(1)
			_, err := m.EditOperator(context.TODO(), mockReq)
			So(err, ShouldNotBeNil)
		})
		Convey("OpenAPI类型解析元数据: 基础算子: 未匹配到带更新算子", func() {
			operatorDB.MetadataType = string(interfaces.MetadataTypeAPI)
			operatorDB.OperatorType = string(interfaces.OperatorTypeBase)
			mockReq.OpenAPIInput = &interfaces.OpenAPIInput{
				Data: json.RawMessage(`{"openapi": "3.0.0"}`),
			}
			mockReq.MetadataType = interfaces.MetadataTypeAPI
			metadataList := []interfaces.IMetadataDB{
				&model.APIMetadataDB{
					Path:   "/path",
					Method: "GET",
				},
			}
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, operatorDB, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataList, nil).Times(1)
			_, err := m.EditOperator(context.TODO(), mockReq)
			So(err, ShouldNotBeNil)
		})
		Convey("OpenAPI类型解析元数据: 组合算子更新, 编辑算子获取TX失败", func() {
			operatorDB.MetadataType = string(interfaces.MetadataTypeAPI)
			operatorDB.OperatorType = string(interfaces.OperatorTypeComposite)
			mockReq.OpenAPIInput = &interfaces.OpenAPIInput{
				Data: json.RawMessage(`{"openapi": "3.0.0"}`),
			}
			mockReq.MetadataType = interfaces.MetadataTypeAPI
			metadataList := []interfaces.IMetadataDB{
				&model.APIMetadataDB{
					Path:        "/path",
					Method:      "GET",
					Summary:     "测试",
					APISpec:     `{}`,
					Description: "测试",
				},
			}
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, operatorDB, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{
				Description: "测试2",
			}, nil).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataList, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, mocks.MockFuncErr("GetTx")).Times(1)
			_, err := m.EditOperator(context.TODO(), mockReq)
			So(err, ShouldNotBeNil)
		})
		Convey("Function类型解析元数据: 编辑算子获取TX失败", func() {
			operatorDB.MetadataType = string(interfaces.MetadataTypeFunc)
			operatorDB.OperatorType = string(interfaces.OperatorTypeComposite)
			mockReq.FunctionInputEdit = &interfaces.FunctionInputEdit{}
			mockReq.MetadataType = interfaces.MetadataTypeFunc
			metadataList := []interfaces.IMetadataDB{
				&model.FunctionMetadataDB{
					Path:        "/path",
					Method:      "GET",
					Summary:     "测试",
					APISpec:     `{}`,
					Description: "测试",
				},
			}
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, operatorDB, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockMetadataService.EXPECT().CheckMetadataExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{
				Description: "测试2",
			}, nil).Times(1)
			mockMetadataService.EXPECT().ParseMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return(metadataList, nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
			p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
				return nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				return nil
			})
			defer p1.Reset()
			_, err := m.EditOperator(context.TODO(), mockReq)
			So(err, ShouldNotBeNil)
		})
		// Convey("算子名称不合法", func() {
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 		Return(true, operatorDB, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
		// 	mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 	mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(errors.New("mock ValidateName error")).Times(1)
		// 	_, err := m.EditOperator(context.TODO(), mockReq)
		// 	So(err, ShouldNotBeNil)
		// })
		// 	Convey("算子元数据查询报错", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockName,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 			}, nil)
		// 		mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByVersion error"))
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 		fmt.Println(err.Error())
		// 		httpErr := &myErr.HTTPError{}
		// 		So(errors.As(err, &httpErr), ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	})
		// 	Convey("算子元数据不存在", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockName,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 			}, nil)
		// 		mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(false, nil, nil)
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 		fmt.Println(err.Error())
		// 		httpErr := &myErr.HTTPError{}
		// 		So(errors.As(err, &httpErr), ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		// 	})
		// 	Convey("检查元数据参数变更,Description不合法", func() {
		// 		// mockAPIEditReq.Summary = mockName
		// 		// mockAPIEditReq.Description = "超出默认字符限制,默认10个"
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockName,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 			}, nil)
		// 		mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil)
		// 		mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidateOperatorDesc")).Times(1)
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// 	Convey("检查元数据参数变更,元数据传参无效", func() {
		// 		mockReq.MetadataType = interfaces.MetadataTypeAPI
		// 		mockReq.Data = []byte(`{"name": "mockName", "description": "mockDesc"}`)
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockName,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 			}, nil)
		// 		// mockOpenAPIParser.EXPECT().GetPathItemContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetPathItemContent")).Times(1)
		// 		mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil)
		// 		mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 		fmt.Println(err.Error())
		// 	})
		// 	Convey("检查元数据参数变更,元数据校验未通过", func() {
		// 		mockReq.MetadataType = interfaces.MetadataTypeAPI
		// 		mockReq.Data = []byte(`{}`)
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockName,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 			}, nil)
		// 		// mockOpenAPIParss
		// 		mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).Times(1)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).Times(2)
		// 		mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil)
		// 		mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorStruct")).Times(1)
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// })
		// Convey("TestEditOperator:编辑未发布算子", t, func() {
		// 	p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
		// 		return nil
		// 	})
		// 	defer p.Reset()
		// 	p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
		// 		return nil
		// 	})
		// 	defer p1.Reset()
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).AnyTimes()
		// 	mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockDBTx.EXPECT().GetTx(gomock.Any()).Return(&sql.Tx{}, nil)
		// 	// mockOpenAPIParser.EXPECT().GetPathItemContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.PathItemContent{
		// 	// 	Summary: mockName,
		// 	// }, nil).AnyTimes()
		// 	Convey("检查算子重名: 查询重名算子失败（db）", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 				Status:       string(interfaces.BizStatusUnpublish),
		// 			}, nil)

		// 		mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByNameAndStatus error"))
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// 	Convey("检查算子重名: 存在重名算子", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 				Status:       string(interfaces.BizStatusUnpublish),
		// 			}, nil)
		// 		mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil)
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// 	Convey("更新算子信息失败", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockReq.Name,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 				Status:       string(interfaces.BizStatusUnpublish),
		// 			}, nil)
		// 		mockDBOperatorManager.EXPECT().UpdateByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateByOperatorID error"))
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// 	Convey("更新版本信息失败", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockReq.Name,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 				Status:       string(interfaces.BizStatusUnpublish),
		// 			}, nil)
		// 		mockDBOperatorManager.EXPECT().UpdateByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockDBAPIMetadataManager.EXPECT().UpdateByVersion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateByVersion error"))
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// })
		// Convey("TestEditOperator:编辑已发布算子", t, func() {
		// 	p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
		// 		return nil
		// 	})
		// 	defer p.Reset()
		// 	p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
		// 		return nil
		// 	})
		// 	defer p1.Reset()
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).AnyTimes()
		// 	mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockDBTx.EXPECT().GetTx(gomock.Any()).Return(&sql.Tx{}, nil).AnyTimes()
		// 	// mockOpenAPIParser.EXPECT().GetPathItemContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.PathItemContent{
		// 	// 	Summary: mockName,
		// 	// }, nil).AnyTimes()
		// 	Convey("新增元数据信息失败", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				Name:         mockReq.Name,
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 				Status:       string(interfaces.BizStatusPublished),
		// 			}, nil)
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("mock InsertAPIMetadata error"))
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// 	Convey("名字变更，通知所有订阅者失败", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 				Status:       string(interfaces.BizStatusPublished),
		// 			}, nil)
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
		// 		mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil)
		// 		mockDBOperatorManager.EXPECT().UpdateByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockAuthService.EXPECT().NotifyResourceChange(gomock.Any(), gomock.Any()).Return(errors.New("mock NotifyResourceChange error"))
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// 	})
		// 	Convey("更新算子信息成功", func() {
		// 		mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Return(true, &model.OperatorRegisterDB{
		// 				UpdateUser:   mockReq.UserID,
		// 				MetadataType: string(interfaces.MetadataTypeAPI),
		// 				Status:       string(interfaces.BizStatusPublished),
		// 			}, nil)
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).Times(1)
		// 		mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
		// 		mockDBOperatorManager.EXPECT().UpdateByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockAuthService.EXPECT().NotifyResourceChange(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 		mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any()).Times(1)
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		time.Sleep(100 * time.Millisecond)
		// 		So(err, ShouldBeNil)
		// 	})
		// })
		// Convey("TestEditOperator:编辑已下架算子", t, func() {
		// 	p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
		// 		return nil
		// 	})
		// 	defer p.Reset()
		// 	p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
		// 		return nil
		// 	})
		// 	defer p1.Reset()
		// 	// mockOpenAPIParser.EXPECT().GetPathItemContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.PathItemContent{
		// 	// 	Summary: mockName,
		// 	// }, nil).AnyTimes()
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).AnyTimes()
		// 	mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidateOperatorName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockDBAPIMetadataManager.EXPECT().SelectByVersion(gomock.Any(), gomock.Any()).Return(true, &model.APIMetadataDB{}, nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockValidator.EXPECT().ValidateOperatorDesc(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		// 	mockDBTx.EXPECT().GetTx(gomock.Any()).Return(&sql.Tx{}, nil).AnyTimes()
		// 	mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).
		// 		Return(true, &model.OperatorRegisterDB{
		// 			UpdateUser:   mockReq.UserID,
		// 			MetadataType: string(interfaces.MetadataTypeAPI),
		// 			Status:       string(interfaces.BizStatusOffline),
		// 		}, nil)
		// 	Convey("升级元数据失败（db）", func() {
		// 		mockDBAPIMetadataManager.EXPECT().InsertAPIMetadata(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("InsertAPIMetadata")).Times(1)
		// 		_, err := m.EditOperator(context.TODO(), mockReq)
		// 		So(err, ShouldNotBeNil)
		// })
	})
}
func TestUpdateOperatorStatus(t *testing.T) {
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
	m := &operatorManager{
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
	}
	Convey("TestUpdateOperatorStatus:更新算子状态", t, func() {
		mockDBTx.EXPECT().GetTx(gomock.Any()).Return(&sql.Tx{}, nil).AnyTimes()
		p := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
			return nil
		})
		defer p.Reset()
		p1 := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
			return nil
		})
		defer p1.Reset()
		req := &interfaces.OperatorStatusUpdateReq{
			StatusItems: []*interfaces.OperatorStatusItem{
				{
					OperatorID: "mockOperatorID",
					Status:     interfaces.BizStatusPublished,
				},
			},
		}
		Convey("获取算子信息失败（DB）", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByOperatorID error"))
			err := m.UpdateOperatorStatus(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("算子不存在", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil)
			err := m.UpdateOperatorStatus(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("当前状态不支持转换", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{
				Status: interfaces.BizStatusPublished.String(),
			}, nil)
			err := m.UpdateOperatorStatus(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
			httpErr := &myErr.HTTPError{}
			So(errors.As(err, &httpErr), ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("获取accessor信息失败", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{
				Status: interfaces.BizStatusEditing.String(),
			}, nil).Times(1)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock GetAccessor error")).Times(1)
			err := m.UpdateOperatorStatus(context.TODO(), req, "")
			So(err, ShouldNotBeNil)
		})
		Convey("发布算子", func() {
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{
				Status: interfaces.BizStatusEditing.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).AnyTimes()
			Convey("没有发布权限", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock CheckPublishPermission error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("检查是否重名:SelectByNameAndStatus获取失败（db）", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByNameAndStatus error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				fmt.Println(err.Error())
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("检查是否重名:存在同名已发布算子", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{}, nil).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				fmt.Println(err.Error())
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusConflict)
			})
			Convey("更新配置信息失败（db）", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateOperatorStatus error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("获取已发布版本失败（DB）", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByOpID error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("存在已发布版本: 更新已发布版本失败（DB）", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
				mockOpReleaseDB.EXPECT().UpdateByOpID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateByOpID error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("存在已发布版本: 添加历史版本失败（DB）", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
				mockOpReleaseDB.EXPECT().UpdateByOpID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock Insert error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("不存在已发布版本: 添加Release记录失败（DB）", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockOpReleaseDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock insert release error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
				httpErr := &myErr.HTTPError{}
				So(errors.As(err, &httpErr), ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("不存在已发布版本: 添加历史版本成功", func() {
				mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().SelectByNameAndStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockOpReleaseDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any()).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldBeNil)
				time.Sleep(100 * time.Millisecond)
			})
		})
		Convey("编辑算子", func() {
			req.StatusItems[0].Status = interfaces.BizStatusEditing
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{
				Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).AnyTimes()
			Convey("没有编辑权限", func() {
				mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock CheckModifyPermission error"))
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("更新状态失败（DB）", func() {
				mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateOperatorStatus error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("更新成功，记录审计日志", func() {
				mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any()).MaxTimes(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldBeNil)
				time.Sleep(100 * time.Millisecond)
			})
		})
		Convey("下架算子", func() {
			req.StatusItems[0].Status = interfaces.BizStatusOffline
			mockDBOperatorManager.EXPECT().SelectByOperatorID(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorRegisterDB{
				Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(&interfaces.AuthAccessor{}, nil).AnyTimes()
			Convey("没有下架权限", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock CheckUnpublishPermission eror")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("更新配置失败", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateOperatorStatus error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("下架算子操作: 获取release记录失败（db）", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByOpID error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("下架算子操作: 当前无release记录，无需更新，直接下架记录审计日志", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any())
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldBeNil)
				time.Sleep(100 * time.Millisecond)
			})
			Convey("下架算子操作: 存在release记录, 获取指定版本历史数据信息失败（DB）", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("mock SelectByOpIDAndMetdata error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("下架算子操作: 存在release记录, 存在历史记录，更新历史记录失败（db）", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseHistoryDB{}, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().UpdateReleaseHistoryByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateReleaseHistoryByID error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("下架算子操作: 存在release记录, 存在历史记录，更新release记录失败（db）", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseHistoryDB{}, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().UpdateReleaseHistoryByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().UpdateByOpID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock UpdateByOpID eror")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("下架算子操作: 存在release记录, 不存在历史记录，添加历史记录失败（db）", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mock insert history error")).Times(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldNotBeNil)
			})
			Convey("下架算子操作: 存在release记录, 不存在历史记录，添加历史记录成功", func() {
				mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().SelectByOpID(gomock.Any(), gomock.Any()).Return(true, &model.OperatorReleaseDB{}, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().SelectByOpIDAndMetdata(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				mockOpReleaseHistoryDB.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockOpReleaseDB.EXPECT().UpdateByOpID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any()).MaxTimes(1)
				err := m.UpdateOperatorStatus(context.TODO(), req, "")
				So(err, ShouldBeNil)
				time.Sleep(100 * time.Millisecond)
			})
		})
	})
}

// 组装OpenAPI 3.0 格式文档
// func assembleOpenAPIDoc(editReq *interfaces.APIMetadataEdit) *openapi3.T {
// 	doc := &openapi3.T{
// 		OpenAPI: "3.0.0",
// 		Info: &openapi3.Info{
// 			Title:   "AISHU DevOps Agent Operator API",
// 			Version: "1.0.0",
// 		},
// 		Paths: openapi3.NewPaths(),
// 		Components: &openapi3.Components{
// 			Schemas: make(map[string]*openapi3.SchemaRef),
// 		},
// 	}

// 	// 创建一个PathItemContent实例，使用editReq中的信息
// 	pathItem := &interfaces.PathItemContent{
// 		Summary:     editReq.Summary,
// 		Path:        editReq.Path,
// 		Method:      editReq.Method,
// 		Description: editReq.Description,
// 		APISpec:     *editReq.APISpec,
// 		ServerURL:   editReq.ServerURL,
// 	}

// 	// 创建操作对象
// 	operation := openapi3.NewOperation()
// 	operation.Summary = pathItem.Summary
// 	operation.Description = pathItem.Description

// 	// 添加参数
// 	if pathItem.APISpec.Parameters != nil {
// 		for _, param := range pathItem.APISpec.Parameters {
// 			openapiParam := &openapi3.Parameter{
// 				Name:        param.Name,
// 				In:          param.In,
// 				Description: param.Description,
// 				Required:    param.Required,
// 				Schema:      param.Schema,
// 				Example:     param.Example,
// 				Examples:    param.Examples,
// 				Content:     param.Content,
// 			}
// 			operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: openapiParam})
// 		}
// 	}

// 	// 添加请求体
// 	if pathItem.APISpec.RequestBody != nil {
// 		requestBody := &openapi3.RequestBody{
// 			Description: pathItem.APISpec.RequestBody.Description,
// 			Content:     pathItem.APISpec.RequestBody.Content,
// 			Required:    pathItem.APISpec.RequestBody.Required,
// 		}
// 		operation.RequestBody = &openapi3.RequestBodyRef{Value: requestBody}
// 	}

// 	// 添加响应
// 	if pathItem.APISpec.Responses != nil {
// 		responses := openapi3.NewResponses()
// 		for _, resp := range pathItem.APISpec.Responses {
// 			response := &openapi3.Response{
// 				Description: &resp.Description,
// 				Content:     resp.Content,
// 			}
// 			responses.Set(resp.StatusCode, &openapi3.ResponseRef{Value: response})
// 		}
// 		operation.Responses = responses
// 	}

// 	// 创建路径项并添加到文档
// 	pathItemObj := &openapi3.PathItem{}

// 	// 根据HTTP方法设置操作
// 	switch pathItem.Method {
// 	case "GET":
// 		pathItemObj.Get = operation
// 	case "POST":
// 		pathItemObj.Post = operation
// 	case "PUT":
// 		pathItemObj.Put = operation
// 	case "DELETE":
// 		pathItemObj.Delete = operation
// 	case "PATCH":
// 		pathItemObj.Patch = operation
// 	case "HEAD":
// 		pathItemObj.Head = operation
// 	case "OPTIONS":
// 		pathItemObj.Options = operation
// 	}

// 	// 添加路径项到文档
// 	doc.Paths.Set(pathItem.Path, pathItemObj)
// 	return doc
// }
