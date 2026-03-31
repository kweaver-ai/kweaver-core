package toolbox

import (
	"context"
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
	Convey("TestImport:工具箱导入", t, func() {
		mockDBTx := mocks.NewMockDBTx(ctrl)
		mockToolBoxDB := mocks.NewMockIToolboxDB(ctrl)
		mockToolDB := mocks.NewMockIToolDB(ctrl)
		mockProxy := mocks.NewMockProxyHandler(ctrl)
		mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
		mockUserMgnt := mocks.NewMockUserManagement(ctrl)
		mockOperatorMgnt := mocks.NewMockOperatorManager(ctrl)
		mockIntCompConfigSvc := mocks.NewMockIIntCompConfigService(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		mockAuditLog := mocks.NewMockLogModelOperator[*metric.AuditLogBuilderParams](ctrl)
		mockMetadataService := mocks.NewMockIMetadataService(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)
		toolbox := &ToolServiceImpl{
			DBTx:             mockDBTx,
			ToolBoxDB:        mockToolBoxDB,
			ToolDB:           mockToolDB,
			Proxy:            mockProxy,
			CategoryManager:  mockCategoryManager,
			Logger:           logger.DefaultLogger(),
			UserMgnt:         mockUserMgnt,
			Validator:        mockValidator,
			OperatorMgnt:     mockOperatorMgnt,
			IntCompConfigSvc: mockIntCompConfigSvc,
			AuthService:      mockAuthService,
			AuditLog:         mockAuditLog,
			MetadataService:  mockMetadataService,
		}
		boxID := "box_id_1"
		importData := &interfaces.ComponentImpexConfigModel{
			Toolbox: &interfaces.ToolBoxImpexConfig{
				Configs: []*interfaces.ToolBoxImpexItem{
					{
						BoxID:   boxID,
						BoxName: "BoxName",
					},
				},
			},
		}
		boxList := []*model.ToolboxDB{
			{
				BoxID: boxID,
				Name:  "BoxName",
			},
		}
		accessor := &interfaces.AuthAccessor{}
		toolInfo := interfaces.ToolInfo{
			Name:         "tool_a",
			Description:  "tool_desc_a",
			MetadataType: interfaces.MetadataTypeAPI,
			Metadata: &interfaces.MetadataInfo{
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
			},
		}
		// tools := []*model.ToolDB{
		// 	{
		// 		ToolID:     "tool_id_1",
		// 		BoxID:      boxID,
		// 		Name:       "tool_name_1",
		// 		SourceType: model.SourceTypeOperator,
		// 		SourceID:   "operator_id_1",
		// 	},
		// }
		Convey("导入数据为空", func() {
			err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, nil, "")
			So(err, ShouldNotBeNil)
			err = toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, &interfaces.ComponentImpexConfigModel{}, "")
			So(err, ShouldNotBeNil)
			err = toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, &interfaces.ComponentImpexConfigModel{
				Toolbox: &interfaces.ToolBoxImpexConfig{},
			}, "")
			So(err, ShouldNotBeNil)
		})
		Convey("导入与检查", func() {
			Convey("存在同名工具", func() {
				mockToolBoxDB.EXPECT().SelectToolBoxByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, &model.ToolboxDB{}, nil)
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			})
			Convey("查询工具箱列表失败（db）", func() {
				mockToolBoxDB.EXPECT().SelectToolBoxByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil)
				mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectListByBoxIDs"))
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("存在资源冲突", func() {
				mockToolBoxDB.EXPECT().SelectToolBoxByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil)
				mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusConflict)
			})
		})
		Convey("获取accessor信息失败", func() {
			mockToolBoxDB.EXPECT().SelectToolBoxByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil)
			mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor"))
			err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			So(err, ShouldNotBeNil)
		})
		Convey("创建模式: 批量导入工具箱及工具元数据", func() {
			mockToolBoxDB.EXPECT().SelectToolBoxByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil)
			mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			Convey("校验导入的工具箱信息: 工具箱名字不合法", func() {
				importData.Toolbox.Configs[0].BoxName = " BoxName"
				mockValidator.EXPECT().ValidatorToolBoxName(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorToolBoxName"))
				mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
			})
			Convey("校验导入的工具箱信息: 工具箱desc不合法", func() {
				importData.Toolbox.Configs[0].BoxDesc = mocks.MockDescription(1000)
				mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxDesc(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorToolBoxDesc")).Times(1)
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
			})
			Convey("校验导入的工具箱信息: 工具名字不合法", func() {
				toolInfo.Name = " toolName"
				importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
					{
						ToolInfo:   toolInfo,
						SourceType: model.SourceTypeOpenAPI,
					},
				}
				mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolName(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorToolName")).Times(1)
				mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
			})
			Convey("校验导入的工具箱信息: 工具Desc不合法", func() {
				toolInfo.Description = mocks.MockDescription(999)
				importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
					{
						ToolInfo:   toolInfo,
						SourceType: model.SourceTypeOpenAPI,
					},
				}
				mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolDesc(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorToolDesc")).Times(1)
				mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
			})
			Convey("校验导入的工具箱信息: 工具元数据为空", func() {
				toolInfo.Metadata = nil
				importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
					{
						ToolInfo:   toolInfo,
						SourceType: model.SourceTypeOpenAPI,
					},
				}
				mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolBoxDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockValidator.EXPECT().ValidatorToolDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
				err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			})
			// Convey("校验导入的工具箱信息: 工具元数据解析失败", func() {
			// 	toolInfo.Metadata = nil
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// 	mockValidator.EXPECT().ValidatorToolBoxName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// 	mockValidator.EXPECT().ValidatorToolBoxDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// 	mockValidator.EXPECT().ValidatorToolName(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// 	mockValidator.EXPECT().ValidatorToolDesc(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			// })
			// Convey("校验导入的工具箱信息: 工具元数据校验未通过", func() {
			// 	toolInfo.Metadata = &interfaces.MetadataInfo{}
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}

			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			// })
			// Convey("校验导入的工具箱信息: 工具箱内存在同名工具", func() {
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			// })
			// Convey("importByCreate: 添加工具箱信息失败（db）", func() {
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	mockToolBoxDB.EXPECT().InsertToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return("", mocks.MockFuncErr("InsertToolBox"))
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			// })
			// Convey("importByCreate: 批量添加元数据失败（db）", func() {
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	mockToolBoxDB.EXPECT().InsertToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			// 	mockMetadataDB.EXPECT().InsertAPIMetadatas(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("InsertAPIMetadatas"))
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			// })
			// Convey("importByCreate: 批量添加工具失败（db）", func() {
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	mockToolBoxDB.EXPECT().InsertToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			// 	mockMetadataDB.EXPECT().InsertAPIMetadatas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			// 	mockToolDB.EXPECT().InsertTools(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("InsertTools"))
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			// })
			// Convey("导入依赖失败", func() {
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	importData.Operator = &interfaces.OperatorImpexConfig{
			// 		Configs: []*interfaces.OperatorImpexItem{{
			// 			OperatorID: "operator_id_1",
			// 		}},
			// 	}
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	mockToolBoxDB.EXPECT().InsertToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			// 	mockMetadataDB.EXPECT().InsertAPIMetadatas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			// 	mockToolDB.EXPECT().InsertTools(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			// 	mockOperatorMgnt.EXPECT().Import(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("Import"))
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldNotBeNil)
			// })
			// Convey("添加所有者权限失败", func() {
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	importData.Operator = &interfaces.OperatorImpexConfig{
			// 		Configs: []*interfaces.OperatorImpexItem{{
			// 			OperatorID: "operator_id_1",
			// 		}},
			// 	}
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	mockToolBoxDB.EXPECT().InsertToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			// 	mockMetadataDB.EXPECT().InsertAPIMetadatas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			// 	mockToolDB.EXPECT().InsertTools(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			// 	mockOperatorMgnt.EXPECT().Import(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			// 	mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("CreateOwnerPolicy"))
			// 	mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any())
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldBeNil)
			// 	time.Sleep(100 * time.Millisecond)
			// })
			// Convey("创建模式导入成功，添加审计日志", func() {
			// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
			// 		{
			// 			ToolInfo:   toolInfo,
			// 			SourceType: model.SourceTypeOpenAPI,
			// 		},
			// 	}
			// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
			// 	mockToolBoxDB.EXPECT().InsertToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			// 	mockMetadataDB.EXPECT().InsertAPIMetadatas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			// 	mockToolDB.EXPECT().InsertTools(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			// 	mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			// 	mockAuditLog.EXPECT().Logger(gomock.Any(), gomock.Any())
			// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeCreate, importData, "")
			// 	So(err, ShouldBeNil)
			// 	time.Sleep(100 * time.Millisecond)
			// })
		})
		// Convey("更新模式: 批量导入工具箱及工具元数据", func() {
		// 	importData.Toolbox.Configs[0].Tools = []*interfaces.ToolImpexItem{
		// 		{
		// 			ToolInfo:   toolInfo,
		// 			SourceType: model.SourceTypeOpenAPI,
		// 		},
		// 	}
		// 	mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 	mockToolBoxDB.EXPECT().SelectToolBoxByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, nil)
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	Convey("检查编辑权限失败", func() {
		// 		mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox).Return(mocks.MockFuncErr("CheckModifyPermission"))
		// 		err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeUpsert, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 	})
		// 	Convey("内置工具箱不允许编辑", func() {
		// 		boxList[0].IsInternal = true
		// 		mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox).Return(nil)
		// 		err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeUpsert, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
		// 	})
		// 	Convey("校验导入的工具箱信息未通过", func() {
		// 		importData.Toolbox.Configs[0].BoxName = "mock name err 不允许空格"
		// 		mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox).Return(nil)
		// 		err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeUpsert, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	})
		// 	Convey("更新工具箱信息失败（db）", func() {
		// 		importData.Toolbox.Configs[0].Status = interfaces.BizStatusPublished
		// 		mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox).Return(nil)
		// 		mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
		// 		mockToolBoxDB.EXPECT().UpdateToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("UpdateToolBox"))
		// 		err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeUpsert, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	})
		// 	Convey("查询工具箱内的工具失败（db）", func() {
		// 		mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox).Return(nil)
		// 		mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
		// 		mockToolBoxDB.EXPECT().UpdateToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockToolDB.EXPECT().SelectToolByBoxID(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectToolByBoxID"))
		// 		err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeUpsert, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	})
		// 	Convey("删除工具箱内的工具失败", func() {
		// 		mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
		// 		mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox).Return(nil)
		// 		mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
		// 		mockToolBoxDB.EXPECT().UpdateToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 		mockToolDB.EXPECT().SelectToolByBoxID(gomock.Any(), gomock.Any()).Return(tools, nil)
		// 		mockToolDB.EXPECT().DeleteBoxByIDAndTools(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("DeleteBoxByIDAndTools"))
		// 		err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeUpsert, importData, "")
		// 		So(err, ShouldNotBeNil)
		// 		httpErr, ok := err.(*myErr.HTTPError)
		// 		So(ok, ShouldBeTrue)
		// 		So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	})
		// Convey("添加元数据失败（db）", func() {
		// 	mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(boxList, nil)
		// 	mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox).Return(nil)
		// 	mockCategoryManager.EXPECT().CheckCategory(gomock.Any()).Return(false)
		// 	mockToolBoxDB.EXPECT().UpdateToolBox(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockToolDB.EXPECT().SelectToolByBoxID(gomock.Any(), gomock.Any()).Return(tools, nil)
		// 	mockToolDB.EXPECT().DeleteBoxByIDAndTools(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mockMetadataDB.EXPECT().InsertAPIMetadatas(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("InsertAPIMetadatas"))
		// 	err := toolbox.Import(context.TODO(), nil, interfaces.ImportTypeUpsert, importData, "")
		// 	So(err, ShouldNotBeNil)
		// 	httpErr, ok := err.(*myErr.HTTPError)
		// 	So(ok, ShouldBeTrue)
		// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// })
		// })
	})
}

func TestExport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	Convey("TestExport:工具箱导出", t, func() {
		mockDBTx := mocks.NewMockDBTx(ctrl)
		mockToolBoxDB := mocks.NewMockIToolboxDB(ctrl)
		mockToolDB := mocks.NewMockIToolDB(ctrl)
		mockProxy := mocks.NewMockProxyHandler(ctrl)
		mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
		mockUserMgnt := mocks.NewMockUserManagement(ctrl)
		mockOperatorMgnt := mocks.NewMockOperatorManager(ctrl)
		mockIntCompConfigSvc := mocks.NewMockIIntCompConfigService(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		mockAuditLog := mocks.NewMockLogModelOperator[*metric.AuditLogBuilderParams](ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)
		toolbox := &ToolServiceImpl{
			DBTx:             mockDBTx,
			ToolBoxDB:        mockToolBoxDB,
			ToolDB:           mockToolDB,
			Proxy:            mockProxy,
			CategoryManager:  mockCategoryManager,
			Logger:           logger.DefaultLogger(),
			UserMgnt:         mockUserMgnt,
			Validator:        mockValidator,
			OperatorMgnt:     mockOperatorMgnt,
			IntCompConfigSvc: mockIntCompConfigSvc,
			AuthService:      mockAuthService,
			AuditLog:         mockAuditLog,
		}
		boxID := "box_id_1"
		accessor := &interfaces.AuthAccessor{}
		ids := []string{boxID}
		exportReq := &interfaces.ExportReq{
			IDs: ids,
		}
		toolBoxDB := &model.ToolboxDB{
			BoxID: boxID,
		}
		metadataVersion := "metadata_version_1"
		tools := []*model.ToolDB{
			{
				ToolID:     "tool_id_1",
				BoxID:      boxID,
				SourceType: model.SourceTypeOpenAPI,
				SourceID:   metadataVersion,
				Parameters: `{}`,
			},
			{
				ToolID:     "tool_id_2",
				BoxID:      boxID,
				SourceType: model.SourceTypeOperator,
				SourceID:   "operator_id_1",
				Parameters: `{}`,
			},
		}
		// metadataDB := &model.APIMetadataDB{
		// 	Version: metadataVersion,
		// 	APISpec: `{}`,
		// }
		Convey("导出预检查", func() {
			Convey("获取accessor信息失败", func() {
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("GetAccessor"))
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
			})
			Convey("检查权限失败", func() {
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
				mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox,
					interfaces.AuthOperationTypeView).Return(nil, mocks.MockFuncErr("ResourceFilterIDs"))
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
			})
			Convey("没有查看权限", func() {
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
				mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox,
					interfaces.AuthOperationTypeView).Return([]string{}, nil)
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
			})
			Convey("查询数据失败（db）", func() {
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
				mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox,
					interfaces.AuthOperationTypeView).Return(ids, nil)
				mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectListByBoxIDs"))
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("请求工具箱不存在", func() {
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
				mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox,
					interfaces.AuthOperationTypeView).Return(ids, nil)
				mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{}, nil)
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			})
		})
		Convey("批量获取工具箱内工具信息", func() {
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox,
				interfaces.AuthOperationTypeView).Return(ids, nil)
			Convey("内置工具箱不允许导出", func() {
				toolBoxDB.IsInternal = true
				mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
			})
			Convey("获取工具箱内的全部工具失败（db）", func() {
				mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
				mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("")
				mockToolDB.EXPECT().SelectToolBoxByIDs(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectToolBoxByIDs"))
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("工具参数解析失败", func() {
				tools[0].Parameters = "A"
				mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
				mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("")
				mockToolDB.EXPECT().SelectToolBoxByIDs(gomock.Any(), gomock.Any()).Return(tools, nil)
				_, err := toolbox.Export(context.TODO(), exportReq)
				So(err, ShouldNotBeNil)
				httpErr, ok := err.(*myErr.HTTPError)
				So(ok, ShouldBeTrue)
				So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			})
			// Convey("查询元数据信息失败（db）", func() {
			// 	mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
			// 	mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("")
			// 	mockToolDB.EXPECT().SelectToolBoxByIDs(gomock.Any(), gomock.Any()).Return(tools, nil)
			// 	mockMetadataDB.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("SelectListByVersion"))
			// 	_, err := toolbox.Export(context.TODO(), exportReq)
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			// })
			// Convey("解析APISpec失败", func() {
			// 	metadataDB.APISpec = ""
			// 	mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
			// 	mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("")
			// 	mockToolDB.EXPECT().SelectToolBoxByIDs(gomock.Any(), gomock.Any()).Return(tools, nil)
			// 	mockMetadataDB.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return([]*model.APIMetadataDB{metadataDB}, nil)
			// 	_, err := toolbox.Export(context.TODO(), exportReq)
			// 	So(err, ShouldNotBeNil)
			// 	httpErr, ok := err.(*myErr.HTTPError)
			// 	So(ok, ShouldBeTrue)
			// 	So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			// })
			// Convey("批量数据获取成功, 不存在依赖算子", func() {
			// 	mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
			// 	mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("")
			// 	mockToolDB.EXPECT().SelectToolBoxByIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolDB{tools[0]}, nil)
			// 	mockMetadataDB.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return([]*model.APIMetadataDB{metadataDB}, nil)
			// 	_, err := toolbox.Export(context.TODO(), exportReq)
			// 	So(err, ShouldBeNil)
			// })
		})
		// Convey("依赖算子导出失败", func() {
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox,
		// 		interfaces.AuthOperationTypeView).Return(ids, nil)
		// 	mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
		// 	mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("")
		// 	mockToolDB.EXPECT().SelectToolBoxByIDs(gomock.Any(), gomock.Any()).Return(tools, nil)
		// 	mockMetadataDB.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return([]*model.APIMetadataDB{metadataDB}, nil)
		// 	mockOperatorMgnt.EXPECT().Export(gomock.Any(), gomock.Any()).Return(nil, mocks.MockFuncErr("OperatorMgntExport"))
		// 	_, err := toolbox.Export(context.TODO(), exportReq)
		// 	So(err, ShouldNotBeNil)
		// })
		// Convey("导出成功", func() {
		// 	mockAuthService.EXPECT().GetAccessor(gomock.Any(), gomock.Any()).Return(accessor, nil)
		// 	mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeToolBox,
		// 		interfaces.AuthOperationTypeView).Return(ids, nil)
		// 	mockToolBoxDB.EXPECT().SelectListByBoxIDs(gomock.Any(), gomock.Any()).Return([]*model.ToolboxDB{toolBoxDB}, nil)
		// 	mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("")
		// 	mockToolDB.EXPECT().SelectToolBoxByIDs(gomock.Any(), gomock.Any()).Return(tools, nil)
		// 	mockMetadataDB.EXPECT().SelectListByVersion(gomock.Any(), gomock.Any()).Return([]*model.APIMetadataDB{metadataDB}, nil)
		// 	mockOperatorMgnt.EXPECT().Export(gomock.Any(), gomock.Any()).Return(&interfaces.ComponentImpexConfigModel{}, nil)
		// 	data, err := toolbox.Export(context.TODO(), exportReq)
		// 	So(err, ShouldBeNil)
		// 	So(data, ShouldNotBeNil)
		// 	So(data.Toolbox, ShouldNotBeNil)
		// 	So(len(data.Toolbox.Configs), ShouldEqual, 1)
		// })
	})
}
