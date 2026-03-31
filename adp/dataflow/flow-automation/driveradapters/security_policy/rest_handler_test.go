package security_policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.uber.org/mock/gomock"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics"
)

// MockLogger 实现 commonLog.Logger 接口
type MockLogger struct{}

func (m *MockLogger) Debugf(format string, args ...interface{}) {}
func (m *MockLogger) Infof(format string, args ...interface{})  {}
func (m *MockLogger) Warnf(format string, args ...interface{})  {}
func (m *MockLogger) Errorf(format string, args ...interface{}) {}
func (m *MockLogger) Fatalf(format string, args ...interface{}) {}
func (m *MockLogger) Debugln(args ...interface{})               {}
func (m *MockLogger) Infoln(args ...interface{})                {}
func (m *MockLogger) Warnln(args ...interface{})                {}
func (m *MockLogger) Errorln(args ...interface{})               {}
func (m *MockLogger) Fatalln(args ...interface{})               {}
func (m *MockLogger) Panicln(args ...interface{})               {}
func (m *MockLogger) Panicf(format string, args ...interface{}) {}
func (m *MockLogger) Tracef(format string, args ...interface{}) {}
func (m *MockLogger) Traceln(args ...interface{})               {}

type MockDependency struct {
	mgnt       *mock_logics.MockMgntHandler
	hydra      *mock_drivenadapters.MockHydraPublic
	hydraAdmin *mock_drivenadapters.MockHydraAdmin
	userMgnt   *mock_drivenadapters.MockUserManagement
}

func NewMockDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	return MockDependency{
		mgnt:       mock_logics.NewMockMgntHandler(ctrl),
		hydra:      mock_drivenadapters.NewMockHydraPublic(ctrl),
		hydraAdmin: mock_drivenadapters.NewMockHydraAdmin(ctrl),
		userMgnt:   mock_drivenadapters.NewMockUserManagement(ctrl),
	}
}

func mockPrivateRestHandlerRouter(t *testing.T) (MockDependency, *gin.Engine) {
	test := setGinMode()
	defer test()
	engine := gin.New()
	engine.Use(gin.Recovery())

	dep := NewMockDependency(t)
	group := engine.Group("/api/automation/v1/security-policy")

	handler := &SecurityPolicyHandler{
		mgnt:       dep.mgnt,
		hydra:      dep.hydra,
		hydraAdmin: dep.hydraAdmin,
		userMgnt:   dep.userMgnt,
	}
	handler.RegisterPrivateAPI(group)

	return dep, engine
}

func mockPublicRestHandlerRouter(t *testing.T) (MockDependency, *gin.Engine) {

	test := setGinMode()
	defer test()
	engine := gin.New()
	engine.Use(gin.Recovery())

	dep := NewMockDependency(t)
	group := engine.Group("/api/automation/v1/security-policy")

	handler := &SecurityPolicyHandler{
		mgnt:       dep.mgnt,
		hydra:      dep.hydra,
		hydraAdmin: dep.hydraAdmin,
		userMgnt:   dep.userMgnt,
	}

	middleware.SetMiddlewareMock(dep.hydraAdmin, dep.userMgnt)
	handler.RegisterAPI(group)

	return dep, engine
}

func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

func TestSecurityPolicyFlow(t *testing.T) {
	dep, engine := mockPrivateRestHandlerRouter(t)

	tests := []struct {
		name       string
		method     string
		path       string
		flowID     string
		setupMock  func()
		wantStatus int
	}{
		{
			name:   "Get Flow Form - Not Found",
			method: http.MethodGet,
			path:   "/api/automation/v1/security-policy/flows/%s/form",
			flowID: "nonexistent",
			setupMock: func() {
				dep.mgnt.EXPECT().
					GetSecurityPolicyFlowByID(gomock.Any(), "nonexistent").
					Return(mgnt.Flow{}, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"id": "nonexistent"}))
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "Get Flow Form - Success Empty Fields",
			method: http.MethodGet,
			path:   "/api/automation/v1/security-policy/flows/%s/form",
			flowID: "existing-flow",
			setupMock: func() {
				dep.mgnt.EXPECT().
					GetSecurityPolicyFlowByID(gomock.Any(), "existing-flow").
					Return(mgnt.Flow{
						ID:    "existing-flow",
						Steps: []entity.Step{},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Get Flow Form - Success With Fields",
			method: http.MethodGet,
			path:   "/api/automation/v1/security-policy/flows/%s/form",
			flowID: "flow-with-fields",
			setupMock: func() {
				fields := primitive.A{
					map[string]interface{}{
						"key":      "field1",
						"name":     "Field 1",
						"type":     "string",
						"required": true,
					},
				}
				dep.mgnt.EXPECT().
					GetSecurityPolicyFlowByID(gomock.Any(), "flow-with-fields").
					Return(mgnt.Flow{
						ID: "flow-with-fields",
						Steps: []entity.Step{
							{
								Parameters: map[string]interface{}{
									"fields": fields,
								},
							},
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Delete Flow - Success",
			method: http.MethodDelete,
			path:   "/api/automation/v1/security-policy/flows/%s",
			flowID: "flow-to-delete",
			setupMock: func() {
				dep.mgnt.EXPECT().
					DeleteSecurityPolicyFlow(gomock.Any(), "flow-to-delete", nil).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "Delete Flow - Not Found",
			method: http.MethodDelete,
			path:   "/api/automation/v1/security-policy/flows/%s",
			flowID: "nonexistent",
			setupMock: func() {
				dep.mgnt.EXPECT().
					DeleteSecurityPolicyFlow(gomock.Any(), "nonexistent", nil).
					Return(ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"id": "nonexistent"}))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			resp := httptest.NewRecorder()
			path := fmt.Sprintf(tt.path, tt.flowID)
			req := httptest.NewRequest(tt.method, path, nil)
			engine.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
		})
	}
}

func TestSecurityPolicyHandler_createProc(t *testing.T) {
	dep, engine := mockPrivateRestHandlerRouter(t)

	// Mock schema validation
	path := "../../schema/security-policy/proc-params.json"
	patch := ApplyGlobalVar(&procParamsSchema, path)
	defer patch.Reset()

	tests := []struct {
		name       string
		reqBody    interface{}
		setupMock  func()
		wantStatus int
		wantBody   gin.H
	}{
		{
			name: "success",
			reqBody: mgnt.ProcParams{
				FlowID: "flow123",
				Values: map[string]interface{}{
					"key": "value",
				},
				Source: map[string]interface{}{
					"type": "file",
					"id":   "file123",
					"name": "example.txt",
					"rev":  "1.0",
					"size": 1024,
					"path": "/path/to/example.txt",
				},
				UserID: "test-user",
			},
			setupMock: func() {
				dep.mgnt.EXPECT().
					StartSecurityPolicyFlowProc(gomock.Any(), gomock.Any()).
					Return("proc123", nil)
			},
			wantStatus: http.StatusCreated,
			wantBody: gin.H{
				"id":     "proc123",
				"status": "init",
			},
		},
		{
			name: "invalid schema",
			reqBody: map[string]interface{}{
				"invalid": "params",
			},
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			resp := httptest.NewRecorder()
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/security-policy/procs", bytes.NewReader(body))
			engine.ServeHTTP(resp, req)
			fmt.Println(resp.Body.String())

			assert.Equal(t, tt.wantStatus, resp.Code)
			if tt.wantBody != nil {
				var got gin.H
				err := json.Unmarshal(resp.Body.Bytes(), &got)
				// assert.NoError(t, err)
				assert.Equal(t, nil, err)
				assert.Equal(t, tt.wantBody, got)
			}
		})
	}
}

func TestSecurityPolicyHandler_StopProc(t *testing.T) {
	dep, engine := mockPrivateRestHandlerRouter(t)

	// Mock schema validation
	path := "../../schema/security-policy/update-proc-status.json"
	patch := ApplyGlobalVar(&updateProcStatusSchema, path)
	defer patch.Reset()

	tests := []struct {
		name       string
		procID     string
		reqBody    interface{}
		setupMock  func()
		wantStatus int
	}{
		{
			name:   "success",
			procID: "proc123",
			reqBody: map[string]interface{}{
				"status": "canceled",
			},
			setupMock: func() {
				dep.mgnt.EXPECT().
					StopSecurityPolicyFlowProc(gomock.Any(), "proc123", nil).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "not found",
			procID: "nonexistent",
			reqBody: map[string]interface{}{
				"status": "canceled",
			},
			setupMock: func() {
				dep.mgnt.EXPECT().
					StopSecurityPolicyFlowProc(gomock.Any(), "nonexistent", nil).
					Return(ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"id": "nonexistent"}))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/automation/v1/security-policy/procs/%s/status", tt.procID), bytes.NewReader([]byte("{}")))
			engine.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
		})
	}
}

func TestSecurityPolicyPublicAPI(t *testing.T) {
	dep, engine := mockPublicRestHandlerRouter(t)
	// Mock schema validation
	_flowStepsSchema := "../../schema/security-policy/flow-steps.json"
	_updateProcStatusSchema := "../../schema/security-policy/update-proc-status.json"

	patch := ApplyGlobalVar(&flowStepsSchema, _flowStepsSchema)
	defer patch.Reset()
	patch = ApplyGlobalVar(&updateProcStatusSchema, _updateProcStatusSchema)
	defer patch.Reset()

	mockUser := drivenadapters.TokenIntrospectInfo{
		UserID:  "test-user",
		Active:  true,
		UdID:    "test-user",
		LoginIP: "127.0.0.1",
	}

	tests := []struct {
		name       string
		method     string
		path       string
		reqBody    interface{}
		setupMock  func()
		wantStatus int
		wantBody   gin.H
	}{
		{
			name:   "Update Flow Steps - Success",
			method: http.MethodPut,
			path:   "/api/automation/v1/security-policy/flows/flow123/steps",
			reqBody: []entity.Step{
				{
					ID:    "step1",
					Title: "Step 1",
				},
				{
					ID:    "step2",
					Title: "Step 2",
				},
			},
			setupMock: func() {
				// Mock the user info call
				dep.userMgnt.EXPECT().
					GetUserInfo("test-user").
					Return(drivenadapters.UserInfo{
						UserID:   "test-user",
						UserName: "testuser",
						Roles:    []string{"super_admin"},
					}, nil)
				dep.mgnt.EXPECT().
					UpdateSecurityPolicyFlow(gomock.Any(), "flow123", gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "Delete Flow - Success",
			method: http.MethodDelete,
			path:   "/api/automation/v1/security-policy/flows/flow123",
			setupMock: func() {
				// Mock the user info call
				dep.userMgnt.EXPECT().
					GetUserInfo("test-user").
					Return(drivenadapters.UserInfo{
						UserID:   "test-user",
						UserName: "testuser",
						Roles:    []string{"super_admin"},
					}, nil)
				dep.mgnt.EXPECT().
					DeleteSecurityPolicyFlow(gomock.Any(), "flow123", gomock.Any()).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "Get Flow By ID - Success",
			method: http.MethodGet,
			path:   "/api/automation/v1/security-policy/flows/flow123",
			setupMock: func() {
				// Mock the user info call
				dep.userMgnt.EXPECT().
					GetUserInfo("test-user").
					Return(drivenadapters.UserInfo{
						UserID:   "test-user",
						UserName: "testuser",
						Roles:    []string{"super_admin"},
					}, nil)
				dep.mgnt.EXPECT().
					GetSecurityPolicyFlowByID(gomock.Any(), "flow123").
					Return(mgnt.Flow{
						ID:         "flow123",
						PolicyType: "upload",
						Steps: []entity.Step{
							{
								ID:    "step1",
								Title: "Step 1",
							},
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Stop Proc By Client - Success",
			method: http.MethodPut,
			path:   "/api/automation/v1/security-policy/procs/proc123/status",
			reqBody: map[string]interface{}{
				"status": "canceled",
			},
			setupMock: func() {
				// Mock the user info call
				dep.userMgnt.EXPECT().
					GetUserInfo("test-user").
					Return(drivenadapters.UserInfo{
						UserID:   "test-user",
						UserName: "testuser",
						Roles:    []string{"super_admin"},
					}, nil)

				dep.mgnt.EXPECT().
					StopSecurityPolicyFlowProc(gomock.Any(), "proc123", gomock.Any()).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup authentication mock
			dep.hydraAdmin.EXPECT().
				Introspect(gomock.Any(), gomock.Any()).
				Return(mockUser, nil)

			tt.setupMock()

			resp := httptest.NewRecorder()
			var req *http.Request

			if tt.reqBody != nil {
				body, _ := json.Marshal(tt.reqBody)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			// Add Authorization header
			req.Header.Set("Authorization", "Bearer test-token")

			engine.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
			if tt.wantBody != nil {
				var got gin.H
				err := json.Unmarshal(resp.Body.Bytes(), &got)
				assert.Equal(t, nil, err)
				assert.Equal(t, tt.wantBody, got)
			}
		})
	}
}

func TestSecurityPolicyPublicAPI_CreateFlow(t *testing.T) {

	dep, engine := mockPublicRestHandlerRouter(t)
	// Mock schema validation
	_createFlowParamsSchema := "../../schema/security-policy/create-flow-params.json"

	patch := ApplyGlobalVar(&createFlowParamsSchema, _createFlowParamsSchema)
	defer patch.Reset()

	mockUser := drivenadapters.TokenIntrospectInfo{
		UserID:  "test-user",
		Active:  true,
		UdID:    "test-user",
		LoginIP: "127.0.0.1",
	}

	tests := []struct {
		name       string
		method     string
		path       string
		reqBody    interface{}
		setupMock  func()
		wantStatus int
		wantBody   gin.H
	}{
		{
			name:   "Create Flow - Success",
			method: http.MethodPost,
			path:   "/api/automation/v1/security-policy/flows",
			reqBody: mgnt.CreateFlowParams{
				PolicyType: "upload",
				Steps: []entity.Step{
					{
						ID:       "step1",
						Title:    "Step 1",
						Operator: "and",
						Parameters: map[string]interface{}{
							"fields": []interface{}{},
						},
					},
					{
						ID:       "step2",
						Title:    "Step 2",
						Operator: "and",
						Parameters: map[string]interface{}{
							"fields": []interface{}{},
						},
					},
				},
			},
			setupMock: func() {
				// Setup authentication mock
				dep.hydraAdmin.EXPECT().
					Introspect(gomock.Any(), gomock.Any()).
					Return(mockUser, nil)
				// Mock the user info call
				dep.userMgnt.EXPECT().
					GetUserInfo("test-user").
					Return(drivenadapters.UserInfo{
						UserID:   "test-user",
						UserName: "testuser",
						Roles:    []string{"super_admin"},
					}, nil)

				dep.mgnt.EXPECT().
					CreateSecurityPolicyFlow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("new-flow-id", nil)
			},
			wantStatus: http.StatusCreated,
			wantBody: gin.H{
				"id": "new-flow-id",
			},
		},
		{
			name:   "Create Flow - Invalid Request Body",
			method: http.MethodPost,
			path:   "/api/automation/v1/security-policy/flows",
			reqBody: map[string]interface{}{
				"invalid": "params",
			},
			setupMock: func() {
				// Setup authentication mock
				dep.hydraAdmin.EXPECT().
					Introspect(gomock.Any(), gomock.Any()).
					Return(mockUser, nil)
				// Mock the user info call
				dep.userMgnt.EXPECT().
					GetUserInfo("test-user").
					Return(drivenadapters.UserInfo{
						UserID:   "test-user",
						UserName: "testuser",
						Roles:    []string{"super_admin"},
					}, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "Create Flow - Unauthorized User",
			method: http.MethodPost,
			path:   "/api/automation/v1/security-policy/flows",
			reqBody: mgnt.CreateFlowParams{
				PolicyType: "upload",
				Steps:      []entity.Step{},
			},
			setupMock: func() {
				_mockUser := drivenadapters.TokenIntrospectInfo{
					UserID:  "test-user",
					Active:  false,
					UdID:    "test-user",
					LoginIP: "127.0.0.1",
				}
				dep.hydraAdmin.EXPECT().
					Introspect(gomock.Any(), gomock.Any()).
					Return(_mockUser, nil)

			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "Create Flow - Internal Error",
			method: http.MethodPost,
			path:   "/api/automation/v1/security-policy/flows",
			reqBody: mgnt.CreateFlowParams{
				PolicyType: "upload",
				Steps: []entity.Step{
					{
						ID:       "step1",
						Title:    "Step 1",
						Operator: "and",
						Parameters: map[string]interface{}{
							"fields": []interface{}{},
						},
					},
					{
						ID:       "step2",
						Title:    "Step 2",
						Operator: "and",
						Parameters: map[string]interface{}{
							"fields": []interface{}{},
						},
					},
				},
			},
			setupMock: func() {
				// Setup authentication mock
				dep.hydraAdmin.EXPECT().
					Introspect(gomock.Any(), gomock.Any()).
					Return(mockUser, nil)
				// Mock the user info call
				dep.userMgnt.EXPECT().
					GetUserInfo("test-user").
					Return(drivenadapters.UserInfo{
						UserID:   "test-user",
						UserName: "testuser",
						Roles:    []string{"super_admin"},
					}, nil)

				dep.mgnt.EXPECT().
					CreateSecurityPolicyFlow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", ierrors.NewIError(ierrors.InternalError, "database error", nil))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.setupMock()

			resp := httptest.NewRecorder()
			var req *http.Request

			if tt.reqBody != nil {
				body, _ := json.Marshal(tt.reqBody)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			// Add Authorization header
			req.Header.Set("Authorization", "Bearer test-token")

			engine.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
			if tt.wantBody != nil {
				var got gin.H
				err := json.Unmarshal(resp.Body.Bytes(), &got)
				assert.Equal(t, nil, err)
				assert.Equal(t, tt.wantBody, got)
			}
		})
	}
}
