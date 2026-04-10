package conversationhandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iportdriver/iportdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type convTestLogger struct{}

func (convTestLogger) Infof(string, ...interface{})  {}
func (convTestLogger) Infoln(...interface{})         {}
func (convTestLogger) Debugf(string, ...interface{}) {}
func (convTestLogger) Debugln(...interface{})        {}
func (convTestLogger) Errorf(string, ...interface{}) {}
func (convTestLogger) Errorln(...interface{})        {}
func (convTestLogger) Warnf(string, ...interface{})  {}
func (convTestLogger) Warnln(...interface{})         {}
func (convTestLogger) Panicf(string, ...interface{}) {}
func (convTestLogger) Panicln(...interface{})        {}
func (convTestLogger) Fatalf(string, ...interface{}) {}
func (convTestLogger) Fatalln(...interface{})        {}

func newConversationCtx(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	c.Request = req

	return c, recorder
}

func setConversationVisitor(c *gin.Context, userID string, userType rest.VisitorType, withReqCtx bool) {
	visitor := &rest.Visitor{ID: userID, Type: userType}
	c.Set(cenum.VisitUserInfoCtxKey.String(), visitor)

	if withReqCtx {
		ctx := context.WithValue(c.Request.Context(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
		c.Request = c.Request.WithContext(ctx)
	}
}

func hasRoute(routes []gin.RouteInfo, method, path string) bool {
	for _, r := range routes {
		if r.Method == method && r.Path == path {
			return true
		}
	}

	return false
}

func TestConversationHandler_RegPubRouter(t *testing.T) {
	t.Parallel()

	h := &conversationHTTPHandler{}
	r := gin.New()
	h.RegPubRouter(r.Group("/v1"))
	routes := r.Routes()

	assert.True(t, hasRoute(routes, http.MethodGet, "/v1/app/:app_key/conversation"))
	assert.True(t, hasRoute(routes, http.MethodGet, "/v1/app/:app_key/conversation/:id"))
	assert.True(t, hasRoute(routes, http.MethodPut, "/v1/app/:app_key/conversation/:id"))
	assert.True(t, hasRoute(routes, http.MethodDelete, "/v1/app/:app_key/conversation/:id"))
	assert.True(t, hasRoute(routes, http.MethodDelete, "/v1/app/:app_key/conversation"))
	assert.True(t, hasRoute(routes, http.MethodPost, "/v1/app/:app_key/conversation"))
	assert.True(t, hasRoute(routes, http.MethodPut, "/v1/app/:app_key/conversation/:id/mark_read"))
}

func TestConversationHandler_List(t *testing.T) {
	t.Parallel()

	t.Run("invalid page", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.URL.RawQuery = "page=bad&size=10"

		h.List(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("invalid size", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.URL.RawQuery = "page=1&size=bad"

		h.List(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("page=0 should return 400", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.URL.RawQuery = "page=0&size=10"

		h.List(c)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("negative page should return 400", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.URL.RawQuery = "page=-1&size=10"

		h.List(c)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("size exceeds max should return 400", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.URL.RawQuery = "page=1&size=1001"

		h.List(c)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, int64(0), errors.New("list failed"))

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.URL.RawQuery = "page=1&size=10"
		setConversationVisitor(c, "u1", rest.VisitorType_App, true)

		h.List(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req conversationreq.ListReq) (conversationresp.ListConversationResp, int64, error) {
			assert.Equal(t, "app-1", req.AgentAPPKey)
			assert.Equal(t, "u1", req.UserId)
			assert.Equal(t, 1, req.Page)
			assert.Equal(t, 10, req.Size)

			return conversationresp.ListConversationResp{}, int64(0), nil
		})

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.URL.RawQuery = "page=1&size=10"
		setConversationVisitor(c, "u1", rest.VisitorType_App, true)

		h.List(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestConversationHandler_Detail(t *testing.T) {
	t.Parallel()

	t.Run("id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		h.Detail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Detail(gomock.Any(), "conv-1").Return(conversationresp.ConversationDetail{}, errors.New("detail failed"))

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.Detail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Detail(gomock.Any(), "conv-1").Return(conversationresp.ConversationDetail{}, nil)

		c, recorder := newConversationCtx(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.Detail(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestConversationHandler_Update(t *testing.T) {
	t.Parallel()

	t.Run("bind error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodPut, "/", "{")
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.Update(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

		c, recorder := newConversationCtx(http.MethodPut, "/", `{"title":"new title"}`)
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.Update(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		c, recorder := newConversationCtx(http.MethodPut, "/", `{"title":"new title"}`)
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.Update(c)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestConversationHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodDelete, "/", "")
		h.Delete(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Delete(gomock.Any(), "conv-1").Return(errors.New("delete failed"))

		c, recorder := newConversationCtx(http.MethodDelete, "/", "")
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.Delete(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Delete(gomock.Any(), "conv-1").Return(nil)

		c, recorder := newConversationCtx(http.MethodDelete, "/", "")
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.Delete(c)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestConversationHandler_DeleteByAPPKey(t *testing.T) {
	t.Parallel()

	t.Run("app key empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodDelete, "/", "")
		h.DeleteByAPPKey(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().DeleteByAppKey(gomock.Any(), "app-1").Return(errors.New("delete by app key failed"))

		c, recorder := newConversationCtx(http.MethodDelete, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		h.DeleteByAPPKey(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().DeleteByAppKey(gomock.Any(), "app-1").Return(nil)

		c, recorder := newConversationCtx(http.MethodDelete, "/", "")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		h.DeleteByAPPKey(c)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestConversationHandler_Init(t *testing.T) {
	t.Parallel()

	t.Run("app key empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodPost, "/", `{"title":"hello"}`)
		setConversationVisitor(c, "u1", rest.VisitorType_App, true)

		h.Init(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodPost, "/", "{")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setConversationVisitor(c, "u1", rest.VisitorType_App, true)

		h.Init(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Init(gomock.Any(), gomock.Any()).Return(conversationresp.InitConversationResp{}, errors.New("init failed"))

		c, recorder := newConversationCtx(http.MethodPost, "/", `{"title":"hello"}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setConversationVisitor(c, "u1", rest.VisitorType_App, true)

		h.Init(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().Init(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req conversationreq.InitReq) (conversationresp.InitConversationResp, error) {
			assert.Equal(t, "app-1", req.AgentAPPKey)
			assert.Equal(t, "u1", req.UserID)
			assert.Equal(t, "v2", req.ExecutorVersion)

			return conversationresp.InitConversationResp{ID: "conv-1"}, nil
		})

		c, recorder := newConversationCtx(http.MethodPost, "/", `{"title":"hello world"}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setConversationVisitor(c, "u1", rest.VisitorType_App, true)

		h.Init(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestConversationHandler_MarkRead(t *testing.T) {
	t.Parallel()

	t.Run("id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodPut, "/", `{"latest_read_index":1}`)
		h.MarkRead(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		c, recorder := newConversationCtx(http.MethodPut, "/", "{")
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.MarkRead(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().MarkRead(gomock.Any(), "conv-1", 1).Return(errors.New("mark read failed"))

		c, recorder := newConversationCtx(http.MethodPut, "/", `{"latest_read_index":1}`)
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.MarkRead(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &conversationHTTPHandler{conversationSvc: mockSvc, logger: convTestLogger{}}

		mockSvc.EXPECT().MarkRead(gomock.Any(), "conv-1", 1).Return(nil)

		c, recorder := newConversationCtx(http.MethodPut, "/", `{"latest_read_index":1}`)
		c.Params = gin.Params{{Key: "id", Value: "conv-1"}}
		h.MarkRead(c)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}
