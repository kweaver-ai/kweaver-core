package mgnt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/tests/mock_logics"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

type listDagsResult struct {
	Total int64                 `json:"total"`
	Page  int64                 `json:"page"`
	Limit int64                 `json:"limit"`
	Dags  []*mgnt.DagSimpleInfo `json:"dags"`
}

type listDagsWithCreatorResult struct {
	Total int64                 `json:"total"`
	Page  int64                 `json:"page"`
	Limit int64                 `json:"limit"`
	Dags  []*mgnt.DagSimpleInfo `json:"dags"`
}

type DagInsRunList struct {
	Total    int64                      `json:"total"`
	Page     int64                      `json:"page"`
	Limit    int64                      `json:"limit"`
	Results  []*mgnt.DagInstanceRunInfo `json:"results"`
	Progress *mgnt.Progress             `json:"progress"`
}

func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

type MockDependency struct {
	mgnt       *mock_logics.MockMgntHandler
	hydra      *mock_drivenadapters.MockHydraPublic
	hydraAdmin *mock_drivenadapters.MockHydraAdmin
	policy     *mock_logics.MockHandler
	userMgnt   *mock_drivenadapters.MockUserManagement
	efast      *mock_drivenadapters.MockEfast
	coderunner *mock_drivenadapters.MockCodeRunner
}

func NewDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return MockDependency{
		mgnt:       mock_logics.NewMockMgntHandler(ctrl),
		hydra:      mock_drivenadapters.NewMockHydraPublic(ctrl),
		hydraAdmin: mock_drivenadapters.NewMockHydraAdmin(ctrl),
		policy:     mock_logics.NewMockHandler(ctrl),
		userMgnt:   mock_drivenadapters.NewMockUserManagement(ctrl),
		efast:      mock_drivenadapters.NewMockEfast(ctrl),
		coderunner: mock_drivenadapters.NewMockCodeRunner(ctrl),
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
	group := engine.Group("/api/automation/v1")
	h = &restHandler{
		mgnt:       dep.mgnt,
		hydra:      dep.hydra,
		hydraAdmin: dep.hydraAdmin,
		config:     common.NewConfig(),
		policy:     dep.policy,
		userMgnt:   dep.userMgnt,
		efast:      dep.efast,
		coderunner: dep.coderunner,
	}
	middleware.SetMiddlewareMock(dep.hydraAdmin, dep.userMgnt)
	h.RegisterAPI(group)
	h.RegisterPrivateAPI(group)

	group2 := engine.Group("/api/automation/v2")
	h2 := &restHandler{
		mgnt:       dep.mgnt,
		hydra:      dep.hydra,
		hydraAdmin: dep.hydraAdmin,
		config:     common.NewConfig(),
		policy:     dep.policy,
		userMgnt:   dep.userMgnt,
		efast:      dep.efast,
	}
	h2.RegisterAPIv2(group2)
	return dep, engine
}

func TestCreate(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	resp := httptest.NewRecorder()
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
	path := "../../schema/base/create.json"
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
	// token expired
	params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	paramsByte, _ := json.Marshal(params)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dag", bytes.NewReader(paramsByte))
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	patch := ApplyGlobalVar(&createSchema, path)
	defer patch.Reset()

	// invalid params
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/dag", bytes.NewReader(paramsByte))
	req.Header.Set("X-Business-Domain", "bd_public")
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// create error
	params = map[string]interface{}{"title": "add15", "description": "", "status": "normal", "steps": []interface{}{
		map[string]interface{}{
			"id":       "0",
			"title":    "新建目录触发执行",
			"operator": "@anyshare-trigger/copy-folder",
			"parameters": map[string]interface{}{
				"docid": "gns://9A8C9277947D4898A350427C768A194C/79666B59BC2D4D26ABCF0B94FDFC8166",
			},
		},
		map[string]interface{}{
			"id":       "1",
			"title":    "复制文件",
			"operator": "@anyshare/folder/copy",
			"parameters": map[string]interface{}{
				"docid": "gns://9A8C9277947D4898A350427C768A194C/79666B59BC2D4D26ABCF0B94FDFC8166",
			},
		},
	}}
	resp = httptest.NewRecorder()
	paramsByte, _ = json.Marshal(params)
	dep.mgnt.EXPECT().CreateDag(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("", errors.New("internal error"))
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/dag", bytes.NewReader(paramsByte))
	req.Header.Set("X-Business-Domain", "bd_public")
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// create success
	resp = httptest.NewRecorder()
	paramsByte, _ = json.Marshal(params)
	dep.mgnt.EXPECT().CreateDag(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("id", nil)
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/dag", bytes.NewReader(paramsByte))
	req.Header.Set("X-Business-Domain", "bd_public")
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var tmpBody map[string]string
	_ = json.Unmarshal(message, &tmpBody)
	assert.Equal(t, 201, resp.Code)
	assert.Equal(t, "id", tmpBody["id"])
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestUpdate(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	resp := httptest.NewRecorder()
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/update.json"
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag/dagid", nil)
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// invalid params update title error
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	patch := ApplyGlobalVar(&updateSchema, path)
	defer patch.Reset()
	params := map[string]interface{}{"title": "add15:", "description": "", "status": "normal"}
	paramsByte, _ := json.Marshal(params)
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag/dagid", bytes.NewReader(paramsByte))
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// invalid parameter status err
	params = map[string]interface{}{"title": "add15", "description": "", "status": "status"}
	paramsByte, _ = json.Marshal(params)
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag/dagid", bytes.NewReader(paramsByte))
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// invalid parameter steps err
	params = map[string]interface{}{"title": "add15", "description": "", "status": "normal", "steps": []interface{}{}}
	paramsByte, _ = json.Marshal(params)
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag/dagid", bytes.NewReader(paramsByte))
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// update err
	params = map[string]interface{}{"title": "add15", "description": "", "status": "normal", "steps": []interface{}{
		map[string]interface{}{
			"id":       "0",
			"title":    "新建目录触发执行",
			"operator": "@anyshare-trigger/copy-folder",
			"parameters": map[string]interface{}{
				"docid": "gns://9A8C9277947D4898A350427C768A194C/79666B59BC2D4D26ABCF0B94FDFC8166",
			},
		},
		map[string]interface{}{
			"id":       "1",
			"title":    "复制文件",
			"operator": "@anyshare/folder/copy",
			"parameters": map[string]interface{}{
				"docid": "gns://9A8C9277947D4898A350427C768A194C/79666B59BC2D4D26ABCF0B94FDFC8166",
			},
		},
	}}
	paramsByte, _ = json.Marshal(params)
	resp = httptest.NewRecorder()
	dep.mgnt.EXPECT().UpdateDag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal error"))
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag/dagid", bytes.NewReader(paramsByte))
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// update success
	params = map[string]interface{}{"title": "add15", "description": "", "status": "normal", "steps": []interface{}{
		map[string]interface{}{
			"id":       "0",
			"title":    "新建目录触发执行",
			"operator": "@anyshare-trigger/copy-folder",
			"parameters": map[string]interface{}{
				"docid": "gns://9A8C9277947D4898A350427C768A194C/79666B59BC2D4D26ABCF0B94FDFC8166",
			},
		},
		map[string]interface{}{
			"id":       "1",
			"title":    "复制文件",
			"operator": "@anyshare/folder/copy",
			"parameters": map[string]interface{}{
				"docid": "gns://9A8C9277947D4898A350427C768A194C/79666B59BC2D4D26ABCF0B94FDFC8166",
			},
		},
	}}
	paramsByte, _ = json.Marshal(params)
	resp = httptest.NewRecorder()
	dep.mgnt.EXPECT().UpdateDag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag/dagid", bytes.NewReader(paramsByte))
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 204, resp.Code)
}

func TestGetDagByID(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	resp := httptest.NewRecorder()
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagid", nil)
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// get dag by id internal error
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	dep.mgnt.EXPECT().GetDagByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("internal err"))
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagid", http.NoBody)
	req.Header.Set("X-Business-Domain", "bd_public")
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// get dag by id success
	// get dag by id internal error
	userInfo.Active = true
	var dagInfo = &mgnt.DagInfo{
		ID:          "ID",
		Title:       "Title",
		Description: "Description",
		Status:      "Status",
		Steps:       []entity.Step{},
		CreatedAt:   0,
		UpdatedAt:   0,
	}
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.mgnt.EXPECT().GetDagByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dagInfo, nil)
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagid", http.NoBody)
	req.Header.Set("X-Business-Domain", "bd_public")
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = &mgnt.DagInfo{}
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult, dagInfo)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestDeleteDagByID(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/automation/v1/dag/dagid", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// get dag by id internal error
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	dep.mgnt.EXPECT().DeleteDagByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal err"))
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/automation/v1/dag/dagid", http.NoBody)
	req.Header.Set("X-Business-Domain", "bd_public")
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// delete dag by id success
	dep.mgnt.EXPECT().DeleteDagByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/automation/v1/dag/dagid", http.NoBody)
	req.Header.Set("X-Business-Domain", "bd_public")
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 204, resp.Code)
}

func TestListDags(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/listtask.json"
	patch := ApplyGlobalVar(&listDagSchema, path)
	defer patch.Reset()
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// params page type err
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags?page=qwe", nil)
	req.Header.Set("X-Business-Domain", "bd_public")
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params limit type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags?page=20&limit=qwe", nil)
	req.Header.Set("X-Business-Domain", "bd_public")
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// listdags internal err
	dep.mgnt.EXPECT().ListDag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("internal err"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags?page=20&limit=1", nil)
	req.Header.Set("X-Business-Domain", "bd_public")
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// listdags success
	var dags = []*mgnt.DagSimpleInfo{
		{
			ID:        "ID",
			Title:     "Title",
			Actions:   []string{"a", "b", "c"},
			CreatedAt: 0,
			UpdatedAt: 0,
			Status:    "",
		}}
	var mockReslt = &listDagsResult{
		Total: 10,
		Page:  20,
		Limit: 1,
		Dags:  dags,
	}
	dep.mgnt.EXPECT().ListDag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags, int64(10), nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags?page=20&limit=1", nil)
	req.Header.Set("X-Business-Domain", "bd_public")
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = &listDagsResult{}
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult.Dags, mockReslt.Dags)
	assert.Equal(t, apiResult.Limit, mockReslt.Limit)
	assert.Equal(t, apiResult.Page, mockReslt.Page)
	assert.Equal(t, apiResult.Total, mockReslt.Total)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestRunInstance(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// runinstance internal err
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	dep.mgnt.EXPECT().RunInstance(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal error"))
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance/dagId", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// runinstance success
	dep.mgnt.EXPECT().RunInstance(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance/dagId", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

func TestRunFormInstance(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
	path := "../../schema/base/forminstance.json"
	patch := ApplyGlobalVar(&formInstanceSchema, path)
	defer patch.Reset()

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// runinstance internal err
	userInfo.Active = true
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"test": "test",
		},
	}
	paramsByte, _ := json.Marshal(data)

	// dep.hydraAdmin.EXPECT().Introspect(gomock.All(),gomock.All()).AnyTimes().Return(userInfo, nil)
	// dep.mgnt.EXPECT().RunFormInstance(gomock.Any(),gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal error"))
	// req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance-form/dagId", nil)
	// resp = httptest.NewRecorder()
	// engine.ServeHTTP(resp, req)
	// assert.Equal(t, 400, resp.Code)

	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	dep.mgnt.EXPECT().RunFormInstance(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("dagInsID", errors.New("internal error"))
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance-form/dagId", bytes.NewReader(paramsByte))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// runinstance success
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.mgnt.EXPECT().RunFormInstance(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("dagInsID", nil)
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance-form/dagId", bytes.NewReader(paramsByte))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

func TestCancleRunningInstance(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/cancletask.json"
	patch := ApplyGlobalVar(&cancletaskSchema, path)
	defer patch.Reset()
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/run-instance/instanceId", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// params status invalid
	userInfo.Active = true
	params := map[string]interface{}{"status": "asd"}
	paramsByte, _ := json.Marshal(params)
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/run-instance/instanceId", bytes.NewReader(paramsByte))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params status invalid type
	params = map[string]interface{}{"status": 123}
	paramsByte, _ = json.Marshal(params)
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/run-instance/instanceId", bytes.NewReader(paramsByte))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// cancleRunningInstance internal error
	params = map[string]interface{}{"status": common.CanceledStatus}
	paramsByte, _ = json.Marshal(params)
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.mgnt.EXPECT().CancelRunningInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal err"))
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/run-instance/instanceId", bytes.NewReader(paramsByte))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	params = map[string]interface{}{"status": common.CanceledStatus}
	paramsByte, _ = json.Marshal(params)
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.mgnt.EXPECT().CancelRunningInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/run-instance/instanceId", bytes.NewReader(paramsByte))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 201, resp.Code)
}

func TestDagInsRunList(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/runtasklist.json"
	patch := ApplyGlobalVar(&dagRunListSchema, path)
	defer patch.Reset()
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// path params empty
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag//results", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params page type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results?page=asd", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params
	// params limit type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results?page=0&limit=qwe", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params sortBy type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results?page=0&limit=20&sortby=asd", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params order type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results?page=0&limit=20&sortby=started_at&order=qwe", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params type type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results?page=0&limit=20&sortby=started_at&order=asc&type=asd", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// ListDagInstance internal err
	dep.mgnt.EXPECT().ListDagInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("internal error"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results?page=0&limit=20&sortby=started_at&order=asc&type=success", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// ListDagInstance success
	var dagInsRunList = &mgnt.DagInstanceRunList{
		DagInstanceRunInfo: []*mgnt.DagInstanceRunInfo{
			{
				ID:        "ID",
				Status:    "Status",
				StartedAt: 0,
				EndedAt:   0,
			}},
		Progress: &mgnt.Progress{
			Total:   10,
			Success: 5,
			Failed:  5,
		},
	}
	var mockResult = &DagInsRunList{
		Total:    10,
		Page:     0,
		Limit:    20,
		Results:  dagInsRunList.DagInstanceRunInfo,
		Progress: dagInsRunList.Progress,
	}
	dep.mgnt.EXPECT().ListDagInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(dagInsRunList, int64(10), nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/results?page=0&limit=20&sortby=started_at&order=asc&type=success", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = &DagInsRunList{}
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, *apiResult, *mockResult)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestListTaskInstance(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/result/resultId", http.NoBody)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// path params dagId empty
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag//result/resultId", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// ListTaskInstance internal err
	dep.mgnt.EXPECT().ListTaskInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("internal error"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/result/resultId", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// ListTaskInstance success
	var taskInsRunList = []*mgnt.TaskInstanceRunInfo{
		{
			ID:        "ID",
			Operator:  "Operator",
			StartedAt: 0,
			Status:    "success",
			Inputs:    map[string]interface{}{"Inputs": "Inputs"},
			Outputs:   map[string]interface{}{"Outputs": "outputs"},
		}}
	dep.mgnt.EXPECT().ListTaskInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(taskInsRunList, int64(1), nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/result/resultId", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = make([]*mgnt.TaskInstanceRunInfo, 0)
	_ = json.Unmarshal(message, &apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult, taskInsRunList)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestGetSuggestDagName(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/suggestname/name", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

	// GetSuggestDagName error
	dep.mgnt.EXPECT().GetSuggestDagName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("internal err"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/suggestname/a", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// GetSuggestDagName success
	var suggestName = "a"
	dep.mgnt.EXPECT().GetSuggestDagName(gomock.Any(), gomock.Any(), gomock.Any()).Return(suggestName, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/suggestname/a", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = map[string]string{}
	_ = json.Unmarshal(message, &apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult["name"], suggestName)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestRunWithDoc(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
	path := "../../schema/base/run-with-doc.json"
	patch := ApplyGlobalVar(&runWithDocSchema, path)
	defer patch.Reset()

	// token expired
	{
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags", nil)
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		assert.Equal(t, 401, resp.Code)
	}

	userInfo.Active = true

	{
		data := map[string]interface{}{
			"docid": "gns://AE8A905C83F540BE89DB3F1F85F5C6C7/2E9940724AB14C1692A2A0B65E1EAAE0",
			"data":  map[string]interface{}{},
		}
		paramsByte, _ := json.Marshal(data)

		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
		dep.mgnt.EXPECT().RunInstanceWithDoc(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal error"))
		req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance-with-doc/dagId", bytes.NewReader(paramsByte))
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%v", string(body))
		assert.Equal(t, 500, resp.Code)
	}

	{
		data := map[string]interface{}{
			"docid": "gns://AE8A905C83F540BE89DB3F1F85F5C6C7/2E9940724AB14C1692A2A0B65E1EAAE0",
			"data":  map[string]interface{}{},
		}
		paramsByte, _ := json.Marshal(data)
		// runinstance success
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.mgnt.EXPECT().RunInstanceWithDoc(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
		req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance-with-doc/dagId", bytes.NewReader(paramsByte))
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%v", string(body))
		assert.Equal(t, 200, resp.Code)
	}

	// token alive
	// runinstance internal err
	{
		data := map[string]interface{}{
			"data": map[string]interface{}{},
		}
		paramsByte, _ := json.Marshal(data)

		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.mgnt.EXPECT().RunInstanceWithDoc(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal error"))
		req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/run-instance-with-doc/dagId", bytes.NewReader(paramsByte))
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%v", string(body))
		assert.Equal(t, 400, resp.Code)
	}

}

type DagSimpleInfoWithCreator struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Actions   []string `json:"actions"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
	Status    string   `json:"status"`
	Creator   string   `json:"creator"`
}

func TestListDocumentDags(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/list-document-dag.json"
	patch := ApplyGlobalVar(&listDocumentDagSchema, path)
	defer patch.Reset()
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
	dep.efast.EXPECT().GetDocMsg(gomock.Any(), gomock.Any()).AnyTimes().Return(&drivenadapters.DocAttr{
		Name: "Fake File",
		Size: 0,
	}, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/document-dags", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// params page type err
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/document-dags?page=qwe", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params limit type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/document-dags?page=20&limit=qwe", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// listdags internal err
	dep.mgnt.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("internal err"))
	dep.userMgnt.EXPECT().GetNameByAccessorIDs(gomock.Any()).Times(1).Return(map[string]string{"UserID": "UserName"}, nil)
	dep.userMgnt.EXPECT().GetUserAccessorIDs(gomock.Any()).Times(1).Return([]string{"UserID"}, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/document-dags?docid=gns%3A%2F%2FAE8A905C83F540BE89DB3F1F85F5C6C7%2F2E9940724AB14C1692A2A0B65E1EAAE0&page=20&limit=1", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// listdags success
	var dags = []*mgnt.DagSimpleInfo{
		{
			ID:        "ID",
			Title:     "Title",
			Actions:   []string{"a", "b", "c"},
			CreatedAt: 0,
			UpdatedAt: 0,
			Status:    "",
			UserID:    "UserID",
		}}

	var mockReslt = &listDagsWithCreatorResult{
		Total: 10,
		Page:  20,
		Limit: 1,
		Dags: []*mgnt.DagSimpleInfo{
			{

				ID:        "ID",
				Title:     "Title",
				Actions:   []string{"a", "b", "c"},
				CreatedAt: 0,
				UpdatedAt: 0,
				Status:    "",
			},
		},
	}

	dep.mgnt.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags, int64(10), nil)
	dep.userMgnt.EXPECT().GetNameByAccessorIDs(gomock.Any()).Times(1).Return(map[string]string{"UserID": "UserName"}, nil)
	dep.userMgnt.EXPECT().GetUserAccessorIDs(gomock.Any()).Times(1).Return([]string{"UserID"}, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/document-dags?docid=gns%3A%2F%2FAE8A905C83F540BE89DB3F1F85F5C6C7%2F2E9940724AB14C1692A2A0B65E1EAAE0&page=20&limit=1", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)

	fmt.Printf("%v", string(message))
	var apiResult = &listDagsWithCreatorResult{}
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult.Limit, mockReslt.Limit)
	assert.Equal(t, apiResult.Page, mockReslt.Page)
	assert.Equal(t, apiResult.Total, mockReslt.Total)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestListRelatedDags(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/list-document-dag.json"
	patch := ApplyGlobalVar(&listDocumentDagSchema, path)
	defer patch.Reset()
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
	dep.efast.EXPECT().GetDocMsg(gomock.Any(), gomock.Any()).AnyTimes().Return(&drivenadapters.DocAttr{
		Name: "Fake File",
		Size: 0,
	}, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/related-dags", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// params page type err
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/related-dags?page=qwe", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params limit type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/related-dags?page=20&limit=qwe", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// listdags internal err
	dep.mgnt.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("internal err"))
	dep.userMgnt.EXPECT().GetNameByAccessorIDs(gomock.Any()).Times(1).Return(map[string]string{"UserID": "UserName"}, nil)
	dep.userMgnt.EXPECT().GetUserAccessorIDs(gomock.Any()).Times(1).Return([]string{"UserID"}, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/related-dags?docid=gns%3A%2F%2FAE8A905C83F540BE89DB3F1F85F5C6C7%2F2E9940724AB14C1692A2A0B65E1EAAE0&page=20&limit=1", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// listdags success
	var dags = []*mgnt.DagSimpleInfo{
		{
			ID:        "ID",
			Title:     "Title",
			Actions:   []string{"a", "b", "c"},
			CreatedAt: 0,
			UpdatedAt: 0,
			Status:    "",
			UserID:    "UserID",
		}}

	var mockReslt = &listDagsWithCreatorResult{
		Total: 10,
		Page:  20,
		Limit: 1,
		Dags: []*mgnt.DagSimpleInfo{
			{

				ID:        "ID",
				Title:     "Title",
				Actions:   []string{"a", "b", "c"},
				CreatedAt: 0,
				UpdatedAt: 0,
				Status:    "",
			},
		},
	}

	dep.mgnt.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags, int64(10), nil)
	dep.userMgnt.EXPECT().GetNameByAccessorIDs(gomock.Any()).Times(1).Return(map[string]string{"UserID": "UserName"}, nil)
	dep.userMgnt.EXPECT().GetUserAccessorIDs(gomock.Any()).Times(1).Return([]string{"UserID"}, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/related-dags?docid=gns%3A%2F%2FAE8A905C83F540BE89DB3F1F85F5C6C7%2F2E9940724AB14C1692A2A0B65E1EAAE0&page=20&limit=1", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)

	fmt.Printf("%v", string(message))
	var apiResult = &listDagsWithCreatorResult{}
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult.Limit, mockReslt.Limit)
	assert.Equal(t, apiResult.Page, mockReslt.Page)
	assert.Equal(t, apiResult.Total, mockReslt.Total)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestCallAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dep, engine := mockRestHandlerRouter(t)

	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:      true,
		UserID:      "UserID",
		UdID:        "UdID",
		LoginIP:     "LoginIP",
		VisitorType: "realname",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

	// Test data
	agentKey := "testAgentKey"
	inputs := map[string]interface{}{
		"input1": "value1",
		"input2": "value2",
	}
	options := &drivenadapters.CallAgentOptions{
		Stream: false,
	}
	expectedResponse := &drivenadapters.CallAgentRes{
		Answer: map[string]interface{}{
			"result": "hello",
		},
	}

	// Mock expectations
	dep.mgnt.EXPECT().CallAgent(gomock.Any(), agentKey, inputs, options, gomock.Any()).Times(1).Return(expectedResponse, nil, nil)

	// Create request
	body, _ := json.Marshal(inputs)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/agent/"+agentKey, bytes.NewReader(body))
	resp := httptest.NewRecorder()

	// Serve request
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)

	// Assertions
	var apiResult drivenadapters.CallAgentRes
	_ = json.Unmarshal(message, &apiResult)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedResponse, apiResult)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestGetDagInstanceCount(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/count", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// path params empty
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag//count", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// GetDagInstanceCount internal err
	dep.mgnt.EXPECT().GetDagInstanceCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(int64(0), errors.New("internal error"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/count?type=success", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// GetDagInstanceCount success
	dep.mgnt.EXPECT().GetDagInstanceCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(10), nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v1/dag/dagId/count?type=success&start_time=1716489600&end_time=1716576000", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = &DagInsRunList{}
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestListDagInstanceV2(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/runtasklist.json"
	patch := ApplyGlobalVar(&dagRunListSchema, path)
	defer patch.Reset()
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// path params empty
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag//results", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params page type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results?page=asd", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params
	// params limit type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results?page=0&limit=qwe", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params sortBy type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results?page=0&limit=20&sortby=asd", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params order type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results?page=0&limit=20&sortby=started_at&order=qwe", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params type type err
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results?page=0&limit=20&sortby=started_at&order=asc&type=asd", http.NoBody)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// ListDagInstance internal err
	dep.mgnt.EXPECT().ListDagInstanceV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("internal error"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results?page=0&limit=20&sortby=started_at&order=asc&type=success", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// ListDagInstance success
	var dagInsRunList = []*mgnt.DagInstanceRunInfo{
		{
			ID:        "ID",
			Status:    "Status",
			StartedAt: 0,
			EndedAt:   0,
		},
	}

	dep.mgnt.EXPECT().ListDagInstanceV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(dagInsRunList, int64(1), nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/results?page=0&limit=20&sortby=started_at&order=asc&type=success&start_time=1716489600&end_time=1716576000", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = &DagInsRunList{}
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestRetryDagInstance(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag-instance/dagInsId/retry", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// path params empty
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag-instance//retry", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// RetryDagInstance internal err
	dep.mgnt.EXPECT().RetryDagInstance(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("internal error"))
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag-instance/dagInsId/retry", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	dep.mgnt.EXPECT().RetryDagInstance(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	req = httptest.NewRequest(http.MethodPut, "/api/automation/v1/dag-instance/dagInsId/retry", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 202, resp.Code)
}

type taskInsRunListRes struct {
	Page    int64                       `json:"page"`
	Limit   int64                       `json:"limit"`
	Total   int64                       `json:"total"`
	Results []*mgnt.TaskInstanceRunInfo `json:"results"`
}

func TestListTaskInstanceV2(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/list-task-instance-v2.json"
	patch := ApplyGlobalVar(&listTaskInstanceV2Schema, path)
	defer patch.Reset()

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/result/resultId", http.NoBody)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// path params dagId empty
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag//result/resultId", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// ListTaskInstance internal err
	dep.mgnt.EXPECT().ListTaskInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("internal error"))
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/result/resultId", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// ListTaskInstance success

	var res = taskInsRunListRes{
		Page:  0,
		Limit: 20,
		Total: 1,
		Results: []*mgnt.TaskInstanceRunInfo{
			{
				ID:        "ID",
				Operator:  "Operator",
				StartedAt: 0,
				Status:    "success",
				Inputs:    map[string]interface{}{"Inputs": "Inputs"},
				Outputs:   map[string]interface{}{"Outputs": "outputs"},
			}},
	}
	dep.mgnt.EXPECT().ListTaskInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(res.Results, res.Total, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/automation/v2/dag/dagId/result/resultId?limit=20&page=0", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = new(taskInsRunListRes)
	_ = json.Unmarshal(message, apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, *apiResult, res)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestBatchListDag(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	path := "../../schema/base/batchlistdag.json"
	patch := ApplyGlobalVar(&batchGetDagSchema, path)
	defer patch.Reset()
	dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)

	// token expired
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/v2/dags/name", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)

	// token alive
	// params page type err
	userInfo.Active = true
	dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
	dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v2/dags/name", nil)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// params limit type err
	input := map[string]interface{}{
		"method": "POST",
	}

	body, _ := json.Marshal(input)
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v2/dags/name", bytes.NewBuffer(body))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)

	// listdags internal err
	input = map[string]interface{}{
		"method":  "GET",
		"dag_ids": []string{"568692340517061413"},
	}

	body, _ = json.Marshal(input)
	dep.mgnt.EXPECT().BatchGetDag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("internal err"))
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v2/dags/name", bytes.NewBuffer(body))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)

	// listdags success
	var dags = []*mgnt.DagInfoOption{
		{
			ID:   "568692340517061413",
			Name: "name",
		}}
	dep.mgnt.EXPECT().BatchGetDag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags, nil)
	req = httptest.NewRequest(http.MethodPost, "/api/automation/v2/dags/name", bytes.NewBuffer(body))
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	result := resp.Result()
	message, _ := io.ReadAll(result.Body)
	var apiResult = []*mgnt.DagInfoOption{}
	_ = json.Unmarshal(message, &apiResult)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, apiResult[0].Name, "name")
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}

func TestGetActions(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  false,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}

	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	Convey("GetActions", t, func() {

		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		Convey("Token Expired", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).Times(1).Return(userInfo, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/actions", nil)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 401, resp.Code)
		})

		userInfo.Active = true
		Convey("Get Actions Failed", func() {
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
			dep.mgnt.EXPECT().ListActions(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("internal error"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/actions", nil)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Get Actions Success", func() {
			actions := []map[string]interface{}{
				{
					"name": common.InternalToolPy3Opt,
					"config": map[string]interface{}{
						"enable": true,
					},
				},
			}
			dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
			dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)
			dep.mgnt.EXPECT().ListActions(gomock.Any(), gomock.Any()).Times(1).Return(actions, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/actions", nil)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult []map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, len(apiResult), len(actions))
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestRunCode(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
		TokenID:  "Bearer test-token",
	}
	Convey("RunCode", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		path := "../../schema/base/runcode.json"
		patch := ApplyGlobalVar(&runcodeSchema, path)
		defer patch.Reset()
		Convey("Invalid JSON Schema", func() {
			reqBody := `{"code": "print('hello')", "input_params": [{}]`
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/pycode/run-by-params", strings.NewReader(reqBody))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Run Code Failed", func() {
			reqBody := `{"code": "print('hello')", "input_params": [{}], "output_params": [{}]}`
			dep.coderunner.EXPECT().RunPyCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("execution failed"))
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/pycode/run-by-params", strings.NewReader(reqBody))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Run Code Success", func() {
			reqBody := `{"code": "print('hello')", "input_params": [{}], "output_params": [{}]}`
			expectedResult := []map[string]interface{}{{"output": "hello"}}
			dep.coderunner.EXPECT().RunPyCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(expectedResult, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/pycode/run-by-params", strings.NewReader(reqBody))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult []map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, len(apiResult), len(expectedResult))
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestContinueBlockedInstance(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("ContinueBlockedInstance", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("Continue Instance Failed", func() {
			dep.mgnt.EXPECT().ContinueBlockInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("continue failed"))
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/task/test-instance", nil)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Continue Instance Success", func() {
			dep.mgnt.EXPECT().ContinueBlockInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/task/test-instance", nil)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 201, resp.Code)
		})
	})
}

func TestUpdateTaskResults(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("UpdateTaskResults", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("Invalid JSON", func() {
			reqBody := `invalid json`
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/task/test-task/results", strings.NewReader(reqBody))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Update Task Results Failed", func() {
			reqBody := `{"result": "success"}`
			dep.mgnt.EXPECT().UpdateTaskResults(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("update failed"))
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/task/test-task/results", strings.NewReader(reqBody))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Update Task Results Success", func() {
			reqBody := `{"result": "success"}`
			dep.mgnt.EXPECT().UpdateTaskResults(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
			req := httptest.NewRequest(http.MethodPut, "/api/automation/v1/task/test-task/results", strings.NewReader(reqBody))
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 204, resp.Code)
		})
	})
}

func TestGetDagTriggerConfig(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("GetDagTriggerConfig", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("Get Trigger Config Failed", func() {
			dep.mgnt.EXPECT().GetDagTriggerConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(mgnt.TriggerConfig{}, errors.New("get config failed"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/task-instance/test-task/trigger-config", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Get Trigger Config Success", func() {
			expectedConfig := mgnt.TriggerConfig{
				ID:       "1",
				Operator: common.InternalToolPy,
				Params:   map[string]interface{}{"test": "test"},
				Result:   map[string]interface{}{"res": "test"},
			}
			dep.mgnt.EXPECT().GetDagTriggerConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(expectedConfig, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/task-instance/test-task/trigger-config", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult mgnt.TriggerConfig
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, expectedConfig.Operator, apiResult.Operator)
			assert.Equal(t, expectedConfig.ID, apiResult.ID)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestGetAgents(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("GetAgents", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("Get Agents Failed", func() {
			dep.mgnt.EXPECT().GetAgents(gomock.Any()).Times(1).Return(nil, errors.New("get agents failed"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/agents", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Get Agents Success", func() {
			expectedAgents := []*rds.AgentModel{
				{
					ID:      0,
					Name:    "name",
					AgentID: "id",
					Version: "vers",
				},
			}
			dep.mgnt.EXPECT().GetAgents(gomock.Any()).Times(1).Return(expectedAgents, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/agents", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult []*rds.AgentModel
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, len(apiResult), len(expectedAgents))
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestListDagsWithPerm(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("ListDagsWithPerm", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		path := "../../schema/base/listtask.json"
		patch := ApplyGlobalVar(&listDagSchema, path)
		defer patch.Reset()

		Convey("Invalid Page Parameter", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v2/dags?page=invalid", http.NoBody)
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Invalid Limit Parameter", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v2/dags?limit=invalid", http.NoBody)
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("List Dags Failed", func() {
			dep.mgnt.EXPECT().ListDagV2(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, int64(0), errors.New("list failed"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v2/dags?page=0&limit=20", http.NoBody)
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("List Dags Success", func() {
			expectedDags := []*mgnt.DagSimpleInfo{
				{
					ID:     "dag1",
					Title:  "Test Dag 1",
					Status: "normal",
				},
				{
					ID:     "dag2",
					Title:  "Test Dag 2",
					Status: "stopped",
				},
			}
			dep.mgnt.EXPECT().ListDagV2(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(expectedDags, int64(2), nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v2/dags?page=0&limit=20", nil)
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			dags := apiResult["dags"].([]interface{})
			assert.Equal(t, len(dags), len(expectedDags))
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestSingleDebug(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("SingleDebug", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		path := "../../schema/base/single_debug.json"
		patch := ApplyGlobalVar(&singleDeBugSchema, path)
		defer patch.Reset()

		Convey("Invalid JSON", func() {
			reqBody := `invalid json`
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/single-debug", strings.NewReader(reqBody))
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Single Debug Failed", func() {
			reqBody := `{"id":"1", "operator": "test-operator", "parameters": {}}`
			dep.mgnt.EXPECT().SingleDeBug(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("", errors.New("debug failed"))
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/single-debug", strings.NewReader(reqBody))
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Single Debug Success", func() {
			reqBody := `{"id":"1", "operator": "test-operator", "parameters": {}}`
			expectedID := "debug-instance-id"
			dep.mgnt.EXPECT().SingleDeBug(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(expectedID, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/single-debug", strings.NewReader(reqBody))
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, expectedID, apiResult["id"])
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestFullDebug(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	reqBody := `{"steps": [
        {
            "id": "0",
            "title": "移动文件触发执行",
            "operator": "@anyshare-trigger/move-file",
            "parameters": {
                "docid": "gns://9A8C9277947D4898A350427C768A194C/73103941E98249D0805DB0481799357D",
                "inherit": false
            }
        },
        {
            "id": "1",
            "title": "复制目录",
            "operator": "@anyshare/file/copy",
            "parameters": {
                "destparent": "gns://9A8C9277947D4898A350427C768A194C/73103941E98249D0805DB0481799357D",
                "docid": "{{__0.id}}",
                "ondup": 2
            }
        }
    ]}`

	Convey("FullDebug", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		path := "../../schema/base/full_debug.json"
		patch := ApplyGlobalVar(&fullDeBugSchema, path)
		defer patch.Reset()

		Convey("Invalid JSON", func() {
			reqBody := `invalid json`
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/full-debug", strings.NewReader(reqBody))
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("Full Debug Failed", func() {
			dep.mgnt.EXPECT().FullDebug(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return("", "", errors.New("debug failed"))
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/full-debug", strings.NewReader(reqBody))
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Full Debug Success", func() {
			expectedID := "debug-dag-id"
			expectedInstID := "debug-instance-id"
			dep.mgnt.EXPECT().FullDebug(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(expectedID, expectedInstID, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/dags/full-debug", strings.NewReader(reqBody))
			req.Header.Set("X-Business-Domain", "bd_public")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, expectedID, apiResult["id"])
			assert.Equal(t, expectedInstID, apiResult["inst_id"])
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestDebugDagsResult(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("DebugDagsResult", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("Get Debug Result Failed", func() {
			dep.mgnt.EXPECT().SingleDeBugResult(gomock.Any(), gomock.Any()).Times(1).Return(entity.TaskInstanceStatusFailed, nil, errors.New("get result failed"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/single-debug/result?id=test-id", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("Debug Result Still Running", func() {
			dep.mgnt.EXPECT().SingleDeBugResult(gomock.Any(), gomock.Any()).Times(1).Return(entity.TaskInstanceStatusRunning, nil, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/single-debug/result?id=test-id", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 202, resp.Code)
		})

		Convey("Debug Result Success", func() {
			expectedStatus := entity.TaskInstanceStatusSuccess
			expectedContents := map[string]interface{}{"output": "debug result"}
			dep.mgnt.EXPECT().SingleDeBugResult(gomock.Any(), gomock.Any()).Times(1).Return(expectedStatus, expectedContents, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/single-debug/result?id=test-id", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, expectedStatus.ToString(), apiResult["status"])
			assert.Equal(t, expectedContents, apiResult["result"])
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})

		Convey("Debug Result Failed", func() {
			expectedStatus := entity.TaskInstanceStatusFailed
			expectedContents := map[string]interface{}{"error": "debug failed"}
			dep.mgnt.EXPECT().SingleDeBugResult(gomock.Any(), gomock.Any()).Times(1).Return(expectedStatus, expectedContents, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/dags/single-debug/result?id=test-id", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			assert.Equal(t, expectedStatus.ToString(), apiResult["status"])
			assert.Equal(t, expectedContents, apiResult["result"])
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestListModelBindDags(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}
	Convey("ListModelBindDags", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("List Model Bind Dags Failed", func() {
			dep.mgnt.EXPECT().ListModelBindDags(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("list failed"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/models/test-model/dags", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("List Model Bind Dags Success", func() {
			expectedDags := []*mgnt.DagSimpleInfo{
				{
					ID:     "dag1",
					Title:  "Model Bound Dag 1",
					Status: "normal",
				},
				{
					ID:     "dag2",
					Title:  "Model Bound Dag 2",
					Status: "stopped",
				},
			}
			dep.mgnt.EXPECT().ListModelBindDags(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(expectedDags, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/models/test-model/dags", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			dags := apiResult["dags"].([]interface{})
			assert.Equal(t, len(dags), len(expectedDags))
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})

		Convey("List Model Bind Dags Empty", func() {
			expectedDags := []*mgnt.DagSimpleInfo{}
			dep.mgnt.EXPECT().ListModelBindDags(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(expectedDags, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/models/test-model/dags", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			dags := apiResult["dags"].([]interface{})
			assert.Equal(t, len(dags), 0)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})
	})
}

func TestHistoryData(t *testing.T) {
	dep, engine := mockRestHandlerRouter(t)
	userInfo := drivenadapters.TokenIntrospectInfo{
		Active:  true,
		UserID:  "UserID",
		UdID:    "UdID",
		LoginIP: "LoginIP",
	}
	userDetail := drivenadapters.UserInfo{
		UserName: "UserName",
		Roles:    []string{"Role"},
	}

	Convey("HistoryData", t, func() {
		dep.policy.EXPECT().CheckStatus().AnyTimes().Return(true, nil)
		dep.hydraAdmin.EXPECT().Introspect(gomock.All(), gomock.All()).AnyTimes().Return(userInfo, nil)
		dep.userMgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(userDetail, nil)

		Convey("History Data Failed", func() {
			dep.mgnt.EXPECT().ListHistoryData(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(mgnt.HistoryDataResp{}, errors.New("list failed"))
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/history-dags", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 500, resp.Code)
		})

		Convey("History Data Success with default parameters", func() {
			expectedData := mgnt.HistoryDataResp{
				Total: 2,
				Items: []*mgnt.HistoryDataItem{
					{
						ID:   "data1",
						Type: "data-flow",
					},
					{
						ID:   "data2",
						Type: "data-flow",
					},
				},
			}
			dep.mgnt.EXPECT().ListHistoryData(gomock.Any(), int64(0), int64(20)).Times(1).Return(expectedData, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/history-dags", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})

		Convey("History Data Success with custom parameters", func() {
			expectedData := mgnt.HistoryDataResp{
				Total: 1,
				Items: []*mgnt.HistoryDataItem{
					{
						ID:   "data1",
						Type: "data-flow",
					},
				},
			}
			dep.mgnt.EXPECT().ListHistoryData(gomock.Any(), int64(5), int64(50)).Times(1).Return(expectedData, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/history-dags?page=5&limit=50", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})

		Convey("History Data Empty", func() {
			expectedData := mgnt.HistoryDataResp{
				Total: 0,
				Items: []*mgnt.HistoryDataItem{},
			}
			dep.mgnt.EXPECT().ListHistoryData(gomock.Any(), int64(0), int64(20)).Times(1).Return(expectedData, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/history-dags", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			result := resp.Result()
			message, _ := io.ReadAll(result.Body)
			var apiResult map[string]interface{}
			_ = json.Unmarshal(message, &apiResult)
			assert.Equal(t, 200, resp.Code)
			if err := result.Body.Close(); err != nil {
				assert.Equal(t, err, nil)
			}
		})

		Convey("History Data with invalid page parameter", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/history-dags?page=invalid", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})

		Convey("History Data with invalid limit parameter", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/automation/v1/history-dags?limit=invalid", http.NoBody)
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assert.Equal(t, 400, resp.Code)
		})
	})
}
