package versions

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/versions"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics/mock_dag_versions"
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
	dagVersions *mock_dag_versions.MockDagVersionService
	hydraAdmin  *mock_drivenadapters.MockHydraAdmin
	userMgnt    *mock_drivenadapters.MockUserManagement
}

// NewMockDependency 初始化Mock 依赖服务
func NewMockDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	return MockDependency{
		dagVersions: mock_dag_versions.NewMockDagVersionService(ctrl),
		hydraAdmin:  mock_drivenadapters.NewMockHydraAdmin(ctrl),
		userMgnt:    mock_drivenadapters.NewMockUserManagement(ctrl),
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
	h = &DagVersionHandler{
		dagVersions: dep.dagVersions,
	}

	middleware.SetMiddlewareMock(dep.hydraAdmin, dep.userMgnt)
	h.RegisterAPI(group)

	return dep, engine
}

func InitErrorInfo() {
	i18n.InitI18nTranslator("../../" + common.MultiResourcePath)
	ierr.InitServiceName(common.ErrCodeServiceName)
}

func TestList(t *testing.T) {
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

	mockVersions := []versions.DagVersionSimple{
		{
			ID:        "1",
			Version:   "v1.0.0",
			VersionID: "1",
			ChangeLog: "",
			UserID:    "",
			UserName:  "",
			CreatedAt: 0,
		},
	}

	Convey("ListDagVersions", t, func() {

		Convey("Token Expired", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/dagId/versions", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		Convey("List Dag Versions Error - Dag Not Found", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(userDetail, nil)
			dep.dagVersions.EXPECT().ListDagVersions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, ierr.NewPublicRestError(context.Background(), ierr.PErrorNotFound, "", nil))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/dagId/versions", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 404, resp.Code)
		})

		Convey("List Dag Versions Success", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(userDetail, nil)
			dep.dagVersions.EXPECT().ListDagVersions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(mockVersions, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/dagId/versions", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var res []versions.DagVersionSimple
			_ = json.Unmarshal(message, &res)
			defer result.Body.Close()
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, res[0].ID, "1")
		})
	})
}

func TestGetNextVersion(t *testing.T) {
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

	Convey("ListDagVersions", t, func() {

		Convey("Token Expired", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/dagId/versions", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		Convey("Get Next Version Error - Dag Not Found", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(userDetail, nil)
			dep.dagVersions.EXPECT().GetNextVersion(gomock.Any(), gomock.Any()).Times(1).Return("", ierr.NewPublicRestError(context.Background(), ierr.PErrorNotFound, "", nil))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/dagId/versions/next", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 404, resp.Code)
		})

		Convey("Get Next Version Success", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(userDetail, nil)
			dep.dagVersions.EXPECT().GetNextVersion(gomock.Any(), gomock.Any()).Times(1).Return("v1.0.0", nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/dagId/versions/next", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var res map[string]string
			_ = json.Unmarshal(message, &res)
			defer result.Body.Close()
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, res["version"], "v1.0.0")
		})
	})
}

func TestRevert(t *testing.T) {
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

	Convey("Revert", t, func() {

		Convey("Token Expired", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/dagId/versions/versionID/rollback", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		path := "../../schema/base/reverttoversion.json"
		patch := ApplyGlobalVar(&revertToVersionSchema, path)
		defer patch.Reset()
		Convey("Params Invalid - Type Error", func() {
			params := map[string]interface{}{
				"version": "va.b.c",
			}
			data, _ := json.Marshal(params)
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(userDetail, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/dagId/versions/versionID/rollback", bytes.NewBuffer(data))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("RevertToVersion Error - Dag Not Found", func() {
			params := map[string]interface{}{
				"version": "v1.0.0",
			}
			data, _ := json.Marshal(params)
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(userDetail, nil)
			dep.dagVersions.EXPECT().RevertToVersion(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("", ierr.NewPublicRestError(context.Background(), ierr.PErrorNotFound, "", nil))
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/dagId/versions/versionID/rollback", bytes.NewBuffer(data))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 404, resp.Code)
		})

		Convey("RevertToVersion Success	", func() {
			params := map[string]interface{}{
				"version": "v1.0.0",
			}
			data, _ := json.Marshal(params)
			dep.hydraAdmin.EXPECT().Introspect(gomock.Any(), gomock.Any()).Times(1).Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(userDetail, nil)
			dep.dagVersions.EXPECT().RevertToVersion(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("554202956002553400", nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/dagId/versions/versionID/rollback", bytes.NewBuffer(data))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var res map[string]string
			_ = json.Unmarshal(message, &res)
			defer result.Body.Close()
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, res["version_id"], "554202956002553400")
		})
	})
}
