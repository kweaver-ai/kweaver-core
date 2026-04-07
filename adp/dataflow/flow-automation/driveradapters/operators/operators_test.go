package operators

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	i18n "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/i18n"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func initErrorInfo() {
	i18n.InitI18nTranslator("../../" + common.MultiResourcePath)
	ierr.InitServiceName(common.ErrCodeServiceName)
}

// setGinMode 设置gin模式为Test模式，使用完成后还原gin模式
func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

type MockDependency struct {
	mgnt       *mock_logics.MockMgntHandler
	hydraAdmin *mock_drivenadapters.MockHydraAdmin
	policy     *mock_logics.MockHandler
	userMgnt   *mock_drivenadapters.MockUserManagement
}

func NewDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return MockDependency{
		mgnt:       mock_logics.NewMockMgntHandler(ctrl),
		hydraAdmin: mock_drivenadapters.NewMockHydraAdmin(ctrl),
		policy:     mock_logics.NewMockHandler(ctrl),
		userMgnt:   mock_drivenadapters.NewMockUserManagement(ctrl),
	}
}

func MockRestHandlerRouter(t *testing.T) (MockDependency, *gin.Engine) {
	initErrorInfo()
	test := setGinMode()
	defer test()
	engine := gin.New()

	// 设置 visitor 对象
	engine.Use(gin.Recovery())

	var h RESTHandler
	dep := NewDependency(t)
	group := engine.Group("/api/automation/v1")
	h = &OperatorsRESTHandler{
		mgnt:   dep.mgnt,
		policy: dep.policy,
	}
	middleware.SetMiddlewareMock(dep.hydraAdmin, dep.userMgnt)
	h.RegisterAPI(group)
	h.RegisterPrivateAPI(group)
	return dep, engine
}

func TestRegisterOperator(t *testing.T) {
	dep, engine := MockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:      false,
		UserID:      "UserID",
		UdID:        "UdID",
		LoginIP:     "LoginIP",
		VisitorType: "realname",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	Convey("RegisterOperator", t, func() {
		Convey("Toekn Expired", func() {
			// token expired
			params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/operators", bytes.NewReader(paramsByte))
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("Invalid Params", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"params": []string{"title: Invalid type, expected string"}})
			})
			defer patch.Reset()
			params := map[string]interface{}{"title": 123, "description": "", "status": "normal"}
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/operators", bytes.NewReader(paramsByte))
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Param Unmarshal Err", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return nil
			})
			defer patch.Reset()
			invalidJSON := []byte(`{"title": 123, "description": "", "status": "normal`)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/operators", bytes.NewReader(invalidJSON))
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("CreateComboOperator Err", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return nil
			})
			defer patch.Reset()
			dep.mgnt.EXPECT().CreateComboOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("", "", ierr.NewPublicRestError(context.Background(), ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil))

			params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/operators", bytes.NewReader(paramsByte))
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Success", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return nil
			})
			defer patch.Reset()
			dep.mgnt.EXPECT().CreateComboOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("dagID", "operatorID", nil)

			params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/operators", bytes.NewReader(paramsByte))
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 201, resp.Code)
		})
	})
}

func TestUpdateOperator(t *testing.T) {
	dep, engine := MockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:      false,
		UserID:      "UserID",
		UdID:        "UdID",
		LoginIP:     "LoginIP",
		VisitorType: "realname",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	Convey("UpdateOperator", t, func() {
		Convey("Toekn Expired", func() {
			// token expired
			params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/1", bytes.NewReader(paramsByte))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("Invalid Params", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"params": []string{"title: Invalid type, expected string"}})
			})
			defer patch.Reset()
			params := map[string]interface{}{"title": 123, "description": "", "status": "normal"}
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/1", bytes.NewReader(paramsByte))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Param Unmarshal Err", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return nil
			})
			defer patch.Reset()
			invalidJSON := []byte(`{"title": 123, "description": "", "status": "normal`)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/1", bytes.NewReader(invalidJSON))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("UpdateComboOperator Err", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return nil
			})
			defer patch.Reset()
			dep.mgnt.EXPECT().UpdateComboOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(ierr.NewPublicRestError(context.Background(), ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil))

			params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/1", bytes.NewReader(paramsByte))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Success", func() {
			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
				return nil
			})
			defer patch.Reset()
			dep.mgnt.EXPECT().UpdateComboOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)

			params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
			paramsByte, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/1", bytes.NewReader(paramsByte))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 204, resp.Code)
		})
	})
}

// func TestListOperator(t *testing.T) {
// 	dep, engine := MockRestHandlerRouter(t)
// 	userInfo := drivenadapters.TokenIntrospectInfo{
// 		Active:      false,
// 		UserID:      "UserID",
// 		UdID:        "UdID",
// 		LoginIP:     "LoginIP",
// 		VisitorType: "realname",
// 	}

// 	userDetail := drivenadapters.UserInfo{
// 		UserName: "UserName",
// 		Roles:    []string{"Role"},
// 	}

// 	Convey("ListOperator", t, func() {
// 		Convey("Toekn Expired", func() {
// 			// token expired
// 			params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
// 			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
// 			paramsByte, _ := json.Marshal(params)
// 			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators", bytes.NewReader(paramsByte))
// 			resp := httptest.NewRecorder()
// 			engine.ServeHTTP(resp, req)
// 			assert.Equal(t, 401, resp.Code)
// 		})

// 		userInfo.Active = true
// 		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
// 		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

// 		Convey("Page Type Err", func() {
// 			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators?page=page", http.NoBody)
// 			resp := httptest.NewRecorder()
// 			engine.ServeHTTP(resp, req)
// 			assert.Equal(t, 400, resp.Code)
// 		})

// 		Convey("Limit Type Err", func() {
// 			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators?page=0&limit=limit", http.NoBody)
// 			resp := httptest.NewRecorder()
// 			engine.ServeHTTP(resp, req)
// 			assert.Equal(t, 400, resp.Code)
// 		})

// 		Convey("Param Unmarshal Err", func() {
// 			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
// 				return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"params": []string{"title: Invalid type, expected string"}})
// 			})
// 			defer patch.Reset()
// 			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators?page=0&limit=20", http.NoBody)
// 			resp := httptest.NewRecorder()
// 			engine.ServeHTTP(resp, req)
// 			assert.Equal(t, 400, resp.Code)
// 		})

// 		Convey("ListComboOperator Err", func() {
// 			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
// 				return nil
// 			})
// 			defer patch.Reset()
// 			dep.mgnt.EXPECT().ListComboOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&mgnt.ComboOperatorList{}, fmt.Errorf("InternalErr"))
// 			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators?page=0&limit=20", http.NoBody)
// 			resp := httptest.NewRecorder()
// 			engine.ServeHTTP(resp, req)
// 			assert.Equal(t, 500, resp.Code)
// 		})

// 		Convey("Success", func() {
// 			patch := ApplyFunc(common.JSONSchemaValidV2, func(ctx context.Context, data []byte, path string) error {
// 				return nil
// 			})
// 			defer patch.Reset()
// 			mockRes := &mgnt.ComboOperatorList{
// 				Ops: []*mgnt.ComboOperatorItem{
// 					{
// 						OperatorID:   "5cc63ac1-e518-454e-b2b9-808ad4585cb0",
// 						OperatorName: "ceshi",
// 						Version:      "0d5410a3-647b-43cf-bc59-aa1cc6b76b4f",
// 						Description:  "description",
// 						OperatorType: "basic",
// 						Category:     "other_category",
// 						Status:       "published",
// 						DagID:        "",
// 						CreatorName:  "a",
// 						CreatedAt:    1745892343890787655,
// 						UpdaterName:  "a",
// 						UpdatedAt:    1745892343890787655,
// 					},
// 				},
// 				Page:  0,
// 				Limit: 10,
// 				Total: 1,
// 			}
// 			dep.mgnt.EXPECT().ListComboOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(mockRes, nil)
// 			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators?page=0&limit=20&name=ceshi", http.NoBody)
// 			resp := httptest.NewRecorder()
// 			engine.ServeHTTP(resp, req)
// 			assert.Equal(t, 200, resp.Code)
// 			var res mgnt.ComboOperatorList
// 			_ = json.Unmarshal(resp.Body.Bytes(), &res)
// 			assert.Equal(t, res.Ops[0].OperatorID, mockRes.Ops[0].OperatorID)
// 		})
// 	})
// }

func TestExport(t *testing.T) {
	dep, engine := MockRestHandlerRouter(t)

	Convey("Export", t, func() {
		Convey("ExportOperator Error", func() {
			dep.mgnt.EXPECT().ExportOperator(gomock.Any(), gomock.Any()).Times(1).Return(mgnt.ExportOperator{}, fmt.Errorf(" export err"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators/configs/export?id=583490894695543018", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("ExportOperator Success", func() {
			exportRes := mgnt.ExportOperator{
				Configs: []*mgnt.OperatorImportExportItem{
					{
						ID:          "583489364093350122",
						Title:       "qq",
						Description: "",
						Steps:       []entity.Step{},
						Category:    "default",
						OutPuts:     []*entity.OutPut{},
						OperatorID:  "eb853234-9722-4bf8-8f16-c596a8c38e14",
						IsRoot:      true,
					},
				},
				OperatorIDs: []string{"eb853234-9722-4bf8-8f16-c596a8c38e14"},
			}
			dep.mgnt.EXPECT().ExportOperator(gomock.Any(), gomock.Any()).Times(1).Return(exportRes, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/operators/configs/export?id=583490894695543018", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 200, resp.Code)
		})
	})
}

func TestImport(t *testing.T) {
	dep, engine := MockRestHandlerRouter(t)

	params := mgnt.ImportOperatorReq{
		Mode: "create",
		Configs: []mgnt.OperatorImportExportItem{
			{
				ID:          "583489364093350122",
				Title:       "qq",
				Description: "qq",
				Steps: []entity.Step{
					{
						ID:       "0",
						Title:    "",
						Operator: "@trigger/form",
						Parameters: map[string]interface{}{
							"fields": []interface{}{
								map[string]interface{}{
									"key":  "doc_id",
									"type": "string",
									"name": "abc",
								},
							},
						},
					},
					{
						ID:         "1",
						Operator:   "@operator/0a2d4bf8-6386-4b62-bedb-2a0b0ac973ab",
						DataSource: &entity.DataSource{},
						Parameters: map[string]interface{}{},
						Cron:       "",
						Branches:   []entity.Branch{},
						Steps:      []entity.Step{},
					},
				},
				Category: "default",
				OutPuts: []*entity.OutPut{
					{
						Key:  "key",
						Name: "abc",
						Type: "string",
					},
				},
				OperatorID: "eb853234-9722-4bf8-8f16-c596a8c38e14",
				IsRoot:     true,
			},
		},
	}

	Convey("Import", t, func() {
		Convey("X-User Error", func() {
			params := mgnt.ImportOperatorReq{}
			b, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/configs/import", bytes.NewReader(b))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 403, resp.Code)
		})

		path := "../../schema/base/import_operator.json"
		patch := ApplyGlobalVar(&importOperatorSchema, path)
		defer patch.Reset()
		Convey("Bad Params -  Mode Invalid", func() {
			params := mgnt.ImportOperatorReq{
				Mode:    "qqq",
				Configs: []mgnt.OperatorImportExportItem{},
			}
			b, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/configs/import", bytes.NewReader(b))
			req.Header.Add("X-User", "4fa5fafe-e751-11ef-b014-dac047ec7bab")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Bad Params - Config Is Empty", func() {
			params := mgnt.ImportOperatorReq{
				Mode:    "create",
				Configs: []mgnt.OperatorImportExportItem{},
			}
			b, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/configs/import", bytes.NewReader(b))
			req.Header.Add("X-User", "4fa5fafe-e751-11ef-b014-dac047ec7bab")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Import Error", func() {
			dep.mgnt.EXPECT().ImportOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("import error"))
			b, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/configs/import", bytes.NewReader(b))
			req.Header.Add("X-User", "4fa5fafe-e751-11ef-b014-dac047ec7bab")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Import Success", func() {
			dep.mgnt.EXPECT().ImportOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
			b, _ := json.Marshal(params)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/operators/configs/import", bytes.NewReader(b))
			req.Header.Add("X-User", "4fa5fafe-e751-11ef-b014-dac047ec7bab")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 201, resp.Code)
		})
	})
}
