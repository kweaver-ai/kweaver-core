package crest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

func TestReplyError2(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		createError    func(*gin.Context) error
		wantStatusCode int
	}{
		{
			name: "普通错误",
			createError: func(ctx *gin.Context) error {
				return errors.New("普通错误")
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "HTTP_400错误",
			createError: func(ctx *gin.Context) error {
				return rest.NewHTTPError(ctx.Request.Context(), http.StatusBadRequest, rest.PublicError_BadRequest)
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "HTTP_401错误",
			createError: func(ctx *gin.Context) error {
				return rest.NewHTTPError(ctx.Request.Context(), http.StatusUnauthorized, rest.PublicError_Unauthorized)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "HTTP_404错误",
			createError: func(ctx *gin.Context) error {
				return rest.NewHTTPError(ctx.Request.Context(), http.StatusNotFound, rest.PublicError_NotFound)
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "HTTP_500错误",
			createError: func(ctx *gin.Context) error {
				return rest.NewHTTPError(ctx.Request.Context(), http.StatusInternalServerError, rest.PublicError_InternalServerError)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}

			err := tt.createError(c)
			ReplyError2(c, err)

			assert.Equal(t, tt.wantStatusCode, w.Code, "Status code should match")
		})
	}
}

func TestGetRestHttpErr(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil错误",
			err:  nil,
			want: false,
		},
		{
			name: "普通错误",
			err:  errors.New("普通错误"),
			want: false,
		},
		{
			name: "HTTP_400错误",
			err:  rest.NewHTTPError(context.Background(), http.StatusBadRequest, rest.PublicError_BadRequest),
			want: true,
		},
		{
			name: "HTTP_401错误",
			err:  rest.NewHTTPError(context.Background(), http.StatusUnauthorized, rest.PublicError_Unauthorized),
			want: true,
		},
		{
			name: "HTTP_404错误",
			err:  rest.NewHTTPError(context.Background(), http.StatusNotFound, rest.PublicError_NotFound),
			want: true,
		},
		{
			name: "HTTP_500错误",
			err:  rest.NewHTTPError(context.Background(), http.StatusInternalServerError, rest.PublicError_InternalServerError),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			restErr, ok := GetRestHttpErr(tt.err)

			assert.Equal(t, tt.want, ok, "GetRestHttpErr should return expected bool")

			if tt.want {
				assert.NotNil(t, restErr, "GetRestHttpErr should return non-nil error")
			} else {
				assert.Nil(t, restErr, "GetRestHttpErr should return nil error")
			}
		})
	}
}
