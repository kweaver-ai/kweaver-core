package capierr

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func New400Err(ctx context.Context, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusBadRequest,
		rest.PublicError_BadRequest).
		WithErrorDetails(detail)

	return
}

func New401Err(ctx context.Context, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusUnauthorized,
		rest.PublicError_Unauthorized).
		WithErrorDetails(detail)

	return
}

func New403Err(ctx context.Context, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusForbidden,
		rest.PublicError_Forbidden).
		WithErrorDetails(detail)

	return
}

func New404Err(ctx context.Context, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusNotFound,
		rest.PublicError_NotFound).
		WithErrorDetails(detail)

	return
}

func New405Err(ctx context.Context, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusMethodNotAllowed,
		rest.PublicError_MethodNotAllowed).
		WithErrorDetails(detail)

	return
}

func New409Err(ctx context.Context, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusConflict,
		rest.PublicError_Conflict).
		WithErrorDetails(detail)

	return
}

func New500Err(ctx context.Context, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError,
		rest.PublicError_InternalServerError).
		WithErrorDetails(detail)

	return
}
