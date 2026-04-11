package chelper

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBizDomainIDFromGinHeader(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		wantID       string
		wantIsExist  bool
		wantErr      bool
		errContains  string
	}{
		{
			name:         "nil context",
			setupContext: nil,
			wantErr:      true,
			errContains:  "c is nil",
		},
		{
			name: "biz domain ID in header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXBizDomainID.String(), "domain-123")
			},
			wantID:      "domain-123",
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "no biz domain ID header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
			},
			wantID:      "",
			wantIsExist: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var c *gin.Context
			if tt.setupContext != nil {
				c, _ = gin.CreateTestContext(httptest.NewRecorder())
				tt.setupContext(c)
			}

			bizDomainID, isExist, err := GetBizDomainIDFromGinHeader(c)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantID, bizDomainID)
				assert.Equal(t, tt.wantIsExist, isExist)
			}
		})
	}
}

func TestGetBizDomainIDFromCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantID    string
		wantPanic bool
	}{
		{
			name:      "nil context panics",
			ctx:       nil,
			wantID:    "",
			wantPanic: true,
		},
		{
			name:      "context without biz domain ID",
			ctx:       context.Background(),
			wantID:    "",
			wantPanic: false,
		},
		{
			name:      "context with biz domain ID",
			ctx:       context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "domain-456"), //nolint:staticcheck
			wantID:    "domain-456",
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantPanic {
				assert.Panics(t, func() {
					GetBizDomainIDFromCtx(tt.ctx)
				})
			} else {
				bizDomainID := GetBizDomainIDFromCtx(tt.ctx)
				assert.Equal(t, tt.wantID, bizDomainID)
			}
		})
	}

	t.Run("context with wrong type panics", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), 123) //nolint:staticcheck // SA1029

		assert.Panics(t, func() {
			GetBizDomainIDFromCtx(ctx)
		})
	})
}

func TestGetUserIDFromGinContext(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		wantID       string
		wantErr      bool
		errContains  string
	}{
		{
			name: "visitor info in context",
			setupContext: func(c *gin.Context) {
				visitor := &rest.Visitor{
					ID: "user-123",
				}
				c.Set(cenum.VisitUserInfoCtxKey.String(), visitor)
			},
			wantID:  "user-123",
			wantErr: false,
		},
		{
			name: "no visitor info in context",
			setupContext: func(c *gin.Context) {
				// Do nothing
			},
			wantID:      "",
			wantErr:     true,
			errContains: "ctx_vistor_info not found",
		},
		{
			name: "invalid type in context",
			setupContext: func(c *gin.Context) {
				c.Set(cenum.VisitUserInfoCtxKey.String(), "not a visitor")
			},
			wantID:      "",
			wantErr:     true,
			errContains: "invalid 'ctx_vistor_info' context value type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			userID, err := GetUserIDFromGinContext(c)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantID, userID)
			}
		})
	}
}

func TestGetVisitorFromCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantID    string
		wantNil   bool
		wantPanic bool
	}{
		{
			name:      "nil context panics",
			ctx:       nil,
			wantPanic: true,
		},
		{
			name:    "context without visitor",
			ctx:     context.Background(),
			wantNil: true,
		},
		{
			name: "context with visitor",
			ctx: context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ //nolint:staticcheck
				ID:      "visitor-123",
				TokenID: "token-456",
			}),
			wantID:  "visitor-123",
			wantNil: false,
		},
		{
			name:      "context with wrong type panics",
			ctx:       context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), "not a visitor"), //nolint:staticcheck
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantPanic {
				assert.Panics(t, func() {
					GetVisitorFromCtx(tt.ctx)
				})
			} else {
				visitor := GetVisitorFromCtx(tt.ctx)
				if tt.wantNil {
					assert.Nil(t, visitor)
				} else {
					assert.NotNil(t, visitor)
					assert.Equal(t, tt.wantID, visitor.ID)
				}
			}
		})
	}
}

func TestGetUserIDFromCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		ctx    context.Context
		wantID string
	}{
		{
			name:   "context without visitor",
			ctx:    context.Background(),
			wantID: "",
		},
		{
			name: "context with visitor",
			ctx: context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ //nolint:staticcheck // SA1029
				ID: "user-789",
			}),
			wantID: "user-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID := GetUserIDFromCtx(tt.ctx)
			assert.Equal(t, tt.wantID, userID)
		})
	}
}

func TestGetUserTokenFromCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantToken string
	}{
		{
			name:      "context without visitor",
			ctx:       context.Background(),
			wantToken: "",
		},
		{
			name: "context with visitor",
			ctx: context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ //nolint:staticcheck // SA1029
				TokenID: "token-123",
			}),
			wantToken: "token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token := GetUserTokenFromCtx(tt.ctx)
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

func TestGetTraceIDFromCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         context.Context
		wantTraceID string
		wantPanic   bool
	}{
		{
			name:      "nil context panics",
			ctx:       nil,
			wantPanic: true,
		},
		{
			name:        "context without trace ID",
			ctx:         context.Background(),
			wantTraceID: "",
		},
		{
			name:        "context with trace ID",
			ctx:         context.WithValue(context.Background(), cenum.TraceIDCtxKey.String(), "trace-123"), //nolint:staticcheck // SA1029
			wantTraceID: "trace-123",
		},
		{
			name:      "context with wrong type panics",
			ctx:       context.WithValue(context.Background(), cenum.TraceIDCtxKey.String(), 123), //nolint:staticcheck // SA1029
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantPanic {
				assert.Panics(t, func() {
					GetTraceIDFromCtx(tt.ctx)
				})
			} else {
				traceID := GetTraceIDFromCtx(tt.ctx)
				assert.Equal(t, tt.wantTraceID, traceID)
			}
		})
	}
}

func TestGetVisitLanguageCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantLang  rest.Language
		wantPanic bool
	}{
		{
			name:      "nil context panics",
			ctx:       nil,
			wantPanic: true,
		},
		{
			name:      "context without language panics (GConfig not initialized)",
			ctx:       context.Background(),
			wantPanic: true, // GConfig.GetDefaultLanguage() panics when GConfig is nil
		},
		{
			name:      "context with simplified chinese",
			ctx:       context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese), //nolint:staticcheck // SA1029
			wantLang:  rest.SimplifiedChinese,
			wantPanic: false,
		},
		{
			name:      "context with american english",
			ctx:       context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.AmericanEnglish), //nolint:staticcheck // SA1029
			wantLang:  rest.AmericanEnglish,
			wantPanic: false,
		},
		{
			name:      "context with wrong type panics",
			ctx:       context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), 123), //nolint:staticcheck // SA1029
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantPanic {
				assert.Panics(t, func() {
					GetVisitLanguageCtx(tt.ctx)
				})
			} else {
				lang := GetVisitLanguageCtx(tt.ctx)
				assert.Equal(t, tt.wantLang, lang)
			}
		})
	}
}
