package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/auth"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics"
	"go.uber.org/mock/gomock"
)

func initCommonLog() commonLog.Logger {
	if commonLog.NewLogger() != nil {
		return commonLog.NewLogger()
	}
	logout := "1"
	logDir := "/var/log/contentAutoMation/ut"
	logName := "contentAutoMation.log"
	commonLog.InitLogger(logout, logDir, logName)
	return commonLog.NewLogger()
}

func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

type MockDependency struct {
	auth           *mock_logics.MockAuthHandler
	hydra          *mock_drivenadapters.MockHydraPublic
	hydraAdmin     *mock_drivenadapters.MockHydraAdmin
	userManagement *mock_drivenadapters.MockUserManagement
}

func NewDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return MockDependency{
		auth:           mock_logics.NewMockAuthHandler(ctrl),
		hydra:          mock_drivenadapters.NewMockHydraPublic(ctrl),
		hydraAdmin:     mock_drivenadapters.NewMockHydraAdmin(ctrl),
		userManagement: mock_drivenadapters.NewMockUserManagement(ctrl),
	}
}

func mockRestHandlerRouter(t *testing.T) (MockDependency, *gin.Engine) {
	test := setGinMode()
	defer test()
	engine := gin.New()

	// 设置 visitor 对象
	engine.Use(gin.Recovery())

	var h RESTHandler
	dep := NewDependency(t)
	middleware.SetMiddlewareMock(dep.hydraAdmin, dep.userManagement)
	group := engine.Group("/api/automation/v1")
	h = &restHandler{
		auth:       dep.auth,
		hydra:      dep.hydra,
		hydraAdmin: dep.hydraAdmin,
		logger:     initCommonLog(),
	}
	h.RegisterAPI(group)
	return dep, engine
}

func TestHydraCallback(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)

	// code invalid
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/oauth2/callback?code=&state=qwe", http.NoBody)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/oauth2/callback?code=code&state=", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// DecodeString err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/oauth2/callback?code=code&state=:/!", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// RequestToken internal err
	dep.auth.EXPECT().RequestToken(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal err"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/oauth2/callback?code=code&state=L2FueXNoYXJlL3poLWNuL2Rpcg==", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// RequestToken success
	dep.auth.EXPECT().RequestToken(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/oauth2/callback?code=code&state=L2FueXNoYXJlL3poLWNuL2Rpcg==", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code)
}

func TestCheckAuth(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	path := "../../schema/base/auth.json"
	patch := ApplyGlobalVar(&authSchema, path)
	defer patch.Reset()

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/oauth2/auth", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	userInfo.Active = true
	// auth
	var res = &auth.CheckAuthRes{
		Status: true,
		URL:    "",
	}
	params := map[string]interface{}{"redirect_uri": "http://10.4.107.97"}
	paramsByte, _ := json.Marshal(params)
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	dep.userManagement.EXPECT().GetUserInfo(gomock.Any()).Return(drivenadapters.UserInfo{
		UserName: "test",
		Roles:    []string{"admin"},
	}, nil)
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/oauth2/auth", bytes.NewReader(paramsByte))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = &auth.CheckAuthRes{}
	_ = json.Unmarshal(message, &apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult, res)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}
