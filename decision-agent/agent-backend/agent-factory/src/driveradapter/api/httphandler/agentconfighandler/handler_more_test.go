package v3agentconfighandler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
)

// ==================== Create — private API bind error ====================

func TestCreate_BindError_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, _ := newACTestCtx(http.MethodPost, "/v3/agent", `invalid json`)
	setACInternalAPI(c)

	h.Create(c)

	assert.True(t, len(c.Errors) > 0, "should have errors")
}

// ==================== Update — deeper paths ====================

func TestUpdate_EmptyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, recorder := newACTestCtx(http.MethodPut, "/v3/agent/", "")
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: ""}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUpdate_BindError_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, _ := newACTestCtx(http.MethodPut, "/v3/agent/agent-1", `invalid json`)
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}

	h.Update(c)

	assert.True(t, len(c.Errors) > 0, "should have errors")
}

// ==================== Delete — deeper paths ====================

func TestDelete_EmptyID_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, recorder := newACTestCtx(http.MethodDelete, "/v3/agent/", "")
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: ""}}

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDelete_ServiceError_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	mockSvc.EXPECT().Delete(gomock.Any(), "agent-1", gomock.Any(), true).Return(
		auditlogdto.AgentDeleteAuditLogInfo{}, errors.New("delete failed"))

	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, _ := newACTestCtx(http.MethodDelete, "/v3/agent/agent-1", "")
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}

	h.Delete(c)

	assert.True(t, len(c.Errors) > 0, "should have errors")
}

func TestDelete_Success_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	mockSvc.EXPECT().Delete(gomock.Any(), "agent-1", gomock.Any(), true).Return(
		auditlogdto.AgentDeleteAuditLogInfo{Name: "my-agent"}, nil)

	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, _ := newACTestCtx(http.MethodDelete, "/v3/agent/agent-1", "")
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}

	h.Delete(c)

	assert.Empty(t, c.Errors, "should have no errors")
}

// ==================== Copy — deeper paths ====================

func TestCopy_EmptyAgentID_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, _ := newACTestCtx(http.MethodPost, "/v3/agent//copy", "")
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: ""}}

	h.Copy(c)

	assert.True(t, len(c.Errors) > 0, "should have errors")
}

func TestCopy_ServiceError_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	mockSvc.EXPECT().Copy(gomock.Any(), "agent-1", gomock.Any()).Return(
		nil, auditlogdto.AgentCopyAuditLogInfo{}, errors.New("copy failed"))

	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, _ := newACTestCtx(http.MethodPost, "/v3/agent/agent-1/copy", "")
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}

	h.Copy(c)

	assert.True(t, len(c.Errors) > 0, "should have errors")
}

func TestCopy_Success_PrivateAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	mockSvc.EXPECT().Copy(gomock.Any(), "agent-1", gomock.Any()).Return(
		&agentconfigresp.CopyResp{}, auditlogdto.AgentCopyAuditLogInfo{ID: "new-id", Name: "copy"}, nil)

	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, recorder := newACTestCtx(http.MethodPost, "/v3/agent/agent-1/copy", "")
	setACInternalAPI(c)
	c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}

	h.Copy(c)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

// ==================== BatchFields — all paths ====================

func TestBatchFields_BindError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	c, _ := newACTestCtx(http.MethodPost, "/v3/agent-fields", `invalid json`)

	h.BatchFields(c)

	assert.True(t, len(c.Errors) > 0, "should have errors")
}

func TestBatchFields_ValidateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	// Empty agent_ids → validate fails
	body := `{"agent_ids":[],"fields":["name"]}`
	c, _ := newACTestCtx(http.MethodPost, "/v3/agent-fields", body)

	h.BatchFields(c)

	assert.True(t, len(c.Errors) > 0, "should have validation errors")
}

func TestBatchFields_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	mockSvc.EXPECT().BatchFields(gomock.Any(), gomock.Any()).Return(nil, errors.New("batch failed"))

	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	body := `{"agent_ids":["a1","a2"],"fields":["name"]}`
	c, _ := newACTestCtx(http.MethodPost, "/v3/agent-fields", body)

	h.BatchFields(c)

	assert.True(t, len(c.Errors) > 0, "should have errors")
}

func TestBatchFields_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := v3portdrivermock.NewMockIDataAgentConfigSvc(ctrl)
	mockSvc.EXPECT().BatchFields(gomock.Any(), gomock.Any()).Return(&agentconfigresp.BatchFieldsResp{}, nil)

	h := &daConfHTTPHandler{daConfSvc: mockSvc, logger: acTestLogger{}}

	body := `{"agent_ids":["a1","a2"],"fields":["name"]}`
	c, recorder := newACTestCtx(http.MethodPost, "/v3/agent-fields", body)

	h.BatchFields(c)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
