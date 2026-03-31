package callback

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics"
	"go.uber.org/mock/gomock"
)

type stubDocumentConverter struct {
	result     map[string]any
	err        error
	lastReq    *dependency.GotenbergCallbackRequest
	bodyBuffer []byte
}

func (s *stubDocumentConverter) ExtractFullText(ctx context.Context, docID string) (map[string]any, error) {
	return nil, nil
}

func (s *stubDocumentConverter) ConvertToPDF(ctx context.Context, taskID, docID string) error {
	return nil
}

func (s *stubDocumentConverter) HandleGotenbergCallback(ctx context.Context, req *dependency.GotenbergCallbackRequest) (map[string]any, error) {
	s.lastReq = req
	if req != nil && req.Body != nil {
		s.bodyBuffer, _ = io.ReadAll(req.Body)
		req.Body = bytes.NewReader(s.bodyBuffer)
	}
	return s.result, s.err
}

func (s *stubDocumentConverter) ResolveFlowFile(context.Context, string) (*dependency.ResolvedFlowFile, error) {
	return nil, nil
}

func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

func newTestEngine(t *testing.T, mgntHandler *mock_logics.MockMgntHandler, converter dependency.DocumentConverter) *gin.Engine {
	t.Helper()
	engine := gin.New()
	engine.Use(gin.Recovery())
	group := engine.Group("/api/automation/v1")

	var h RESTHandler = &restHandler{
		mgnt:         mgntHandler,
		docConverter: converter,
	}
	h.RegisterPrivateAPI(group)

	return engine
}

func TestGotenbergCallbackSuccess(t *testing.T) {
	restore := setGinMode()
	defer restore()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgntHandler := mock_logics.NewMockMgntHandler(ctrl)
	converter := &stubDocumentConverter{
		result: map[string]any{"doc_id": "pdf-1"},
	}
	engine := newTestEngine(t, mgntHandler, converter)

	mgntHandler.EXPECT().
		ContinueBlockInstances(gomock.Any(), []string{"task-1"}, map[string]any{"doc_id": "pdf-1"}, entity.TaskInstanceStatusSuccess).
		Times(1).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/gotenberg/callback/success", bytes.NewReader([]byte("%PDF-test")))
	req.Header.Set("X-Task-ID", "task-1")
	req.Header.Set("X-Source-Doc-ID", "dfs://123")
	req.Header.Set("X-Result-File-Name", "result.pdf")
	req.Header.Set("Content-Type", "application/pdf")

	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "task-1", converter.lastReq.TaskID)
	assert.Equal(t, "dfs://123", converter.lastReq.DocID)
	assert.Equal(t, "result.pdf", converter.lastReq.FileName)
	assert.Equal(t, "application/pdf", converter.lastReq.ContentType)
	assert.Equal(t, "%PDF-test", string(converter.bodyBuffer))
}

func TestGotenbergCallbackConverterFailureContinuesTaskAsFailed(t *testing.T) {
	restore := setGinMode()
	defer restore()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgntHandler := mock_logics.NewMockMgntHandler(ctrl)
	converter := &stubDocumentConverter{
		err: errors.New("convert failed"),
	}
	engine := newTestEngine(t, mgntHandler, converter)

	mgntHandler.EXPECT().
		ContinueBlockInstances(gomock.Any(), []string{"task-2"}, map[string]any{
			"code":    "gotenberg_callback_failed",
			"message": "convert failed",
		}, entity.TaskInstanceStatusFailed).
		Times(1).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/gotenberg/callback/success", bytes.NewReader([]byte("bad-body")))
	req.Header.Set("X-Task-ID", "task-2")
	req.Header.Set("X-Source-Doc-ID", "dfs://456")
	req.Header.Set("Content-Type", "application/pdf")

	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestGotenbergCallbackErrorRouteContinuesTaskAsFailed(t *testing.T) {
	restore := setGinMode()
	defer restore()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgntHandler := mock_logics.NewMockMgntHandler(ctrl)
	converter := &stubDocumentConverter{}
	engine := newTestEngine(t, mgntHandler, converter)

	mgntHandler.EXPECT().
		ContinueBlockInstances(gomock.Any(), []string{"task-3"}, map[string]any{
			"code":    "gotenberg_callback_failed",
			"message": "gotenberg upstream error",
		}, entity.TaskInstanceStatusFailed).
		Times(1).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/automation/v1/gotenberg/callback/error", bytes.NewReader([]byte("gotenberg upstream error")))
	req.Header.Set("X-Task-ID", "task-3")
	req.Header.Set("Content-Type", "text/plain")

	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, converter.lastReq, nil)
}
