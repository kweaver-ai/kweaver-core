package observability

import (
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/observability"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics/mock_observability"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

type MockDependency struct {
	observability *mock_observability.MockObservabilityHandler
	hydraAdmin    *mock_drivenadapters.MockHydraAdmin
	userMgnt      *mock_drivenadapters.MockUserManagement
}

// NewMockDependency 初始化Mock 依赖服务
func NewMockDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	return MockDependency{
		observability: mock_observability.NewMockObservabilityHandler(ctrl),
		hydraAdmin:    mock_drivenadapters.NewMockHydraAdmin(ctrl),
		userMgnt:      mock_drivenadapters.NewMockUserManagement(ctrl),
	}
}

func MockRestHandlerRouter(t *testing.T) (MockDependency, *gin.Engine) {
	InitErrorInfo()
	test := setGinMode()
	defer test()

	engine := gin.New()

	engine.Use(gin.Recovery())

	var h RESTHandler
	dep := NewMockDependency(t)
	group := engine.Group("/api/automation/v1")
	h = &observabilityRESTHandler{
		observability: dep.observability,
	}

	middleware.SetMiddlewareMock(dep.hydraAdmin, dep.userMgnt)
	h.RegisterAPI(group)

	return dep, engine
}

func InitErrorInfo() {
	i18n.InitI18nTranslator("../../" + common.MultiResourcePath)
	ierr.InitServiceName(common.ErrCodeServiceName)
}

func TestFullView(t *testing.T) {
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
		Roles:    []string{"super_admin"},
	}

	Convey("FullView", t, func() {

		Convey("Token Expired", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/full-view", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		path := "../../schema/base/observability.json"
		patch := ApplyGlobalVar(&observabilitySchema, path)
		defer patch.Reset()
		Convey("Params Invalid - Type Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/full-view?type=dag", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Params Invalid - Page Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/full-view?page=-1", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Params Invalid - Limit Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/full-view?page=0&limit=101", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Full View Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			dep.observability.EXPECT().FullView(gomock.Any(), gomock.Any(), gomock.Any()).Return(observability.FullViewRes{}, fmt.Errorf("Full View Error"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/full-view?page=0&limit=10&type=data-flow", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Success", func() {
			mockRes := observability.FullViewRes{
				Basic: observability.BasicInfo{
					DagCnt:   10,
					Cron:     0,
					Event:    0,
					Manually: 0,
				},
				Run: observability.RunInfo{
					SuccessCnt: 100,
					FailedCnt:  90,
					TotalCnt:   200,
				},
			}
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			dep.observability.EXPECT().FullView(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRes, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/full-view?page=0&limit=10&type=data-flow", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var res observability.FullViewRes
			_ = json.Unmarshal(message, &res)
			assert.Equal(t, 200, resp.Code)
			So(res, ShouldResemble, mockRes)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestRunTime(t *testing.T) {
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
		Roles:    []string{"super_admin"},
	}

	Convey("RunTime", t, func() {
		Convey("Token Expired", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/runtime-view", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		path := "../../schema/base/observability.json"
		patch := ApplyGlobalVar(&observabilitySchema, path)
		defer patch.Reset()
		Convey("Params Invalid - Type Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/runtime-view?type=dag", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})
		Convey("Params Invalid - Page Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/runtime-view?page=-1", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Params Invalid - Limit Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/runtime-view?page=0&limit=101", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Runtime View Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			dep.observability.EXPECT().RuntimeView(gomock.Any(), gomock.Any(), gomock.Any()).Return(observability.RuntimeViewRes{}, fmt.Errorf("Full View Error"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/runtime-view?page=0&limit=10&type=data-flow", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Success", func() {
			mockRes := observability.RuntimeViewRes{
				Datas: []*observability.RuntimeViewItem{
					{
						ID:   "1",
						Name: "test",
						Metric: &observability.RuntimeViewMetric{
							FailedRate: 0,
						},
						StatusSummary: &drivenadapters.StatusCnt{
							Total:    10,
							Success:  9,
							Failed:   1,
							Blocked:  0,
							Canceled: 0,
							Running:  0,
							Init:     0,
						},
					},
				},
				Total: 1,
				Page:  0,
				Limit: 10,
			}
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			dep.observability.EXPECT().RuntimeView(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRes, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/runtime-view?page=0&limit=10&type=data-flow", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var res observability.RuntimeViewRes
			_ = json.Unmarshal(message, &res)
			assert.Equal(t, 200, resp.Code)
			So(res, ShouldResemble, mockRes)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})

}

func TestRecent(t *testing.T) {
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
		Roles:    []string{"super_admin"},
	}

	Convey("Recent", t, func() {
		Convey("Token Expired", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/recent", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		path := "../../schema/base/observability.json"
		patch := ApplyGlobalVar(&observabilitySchema, path)
		defer patch.Reset()
		Convey("Recent View Error", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			dep.observability.EXPECT().RecentRunView(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*observability.RuntimeViewItem{}, fmt.Errorf("Recent View Error"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/recent?trigger=cron", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Success", func() {
			mockRes := []*observability.RuntimeViewItem{
				{
					ID:   "1",
					Name: "test",
					Metric: &observability.RuntimeViewMetric{
						FailedRate: 0,
					},
					StatusSummary: &drivenadapters.StatusCnt{
						Total:    10,
						Success:  9,
						Failed:   1,
						Blocked:  0,
						Canceled: 0,
						Running:  0,
						Init:     0,
					},
				},
			}

			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
			dep.observability.EXPECT().RecentRunView(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRes, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/recent?trigger=cron", http.NoBody)
			req.Header.Set("X-Business-Domain", "test")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var res []*observability.RuntimeViewItem
			_ = json.Unmarshal(message, &res)
			assert.Equal(t, 200, resp.Code)
			So(res, ShouldResemble, mockRes)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})

}

func TestIsVisible(t *testing.T) {
	dep, engine := MockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:      true,
		UserID:      "UserID",
		UdID:        "UdID",
		LoginIP:     "LoginIP",
		VisitorType: "realname",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"super_admin"},
	}

	Convey("Visible", t, func() {
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.All()).Times(1).Return(userDetail, nil)
		dep.observability.EXPECT().IsVisible(gomock.Any()).Times(1).Return(true)
		req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/observability/visible", http.NoBody)
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
	})

}
