package capierr

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func NewCustom400Err(ctx context.Context, customErrCode string, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusBadRequest,
		customErrCode).
		WithErrorDetails(detail)

	return
}

func NewCustom401Err(ctx context.Context, customErrCode string, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusUnauthorized,
		customErrCode).
		WithErrorDetails(detail)

	return
}

func NewCustom403Err(ctx context.Context, customErrCode string, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusForbidden,
		customErrCode).
		WithErrorDetails(detail)

	return
}

func NewCustom404Err(ctx context.Context, customErrCode string, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusNotFound,
		customErrCode).
		WithErrorDetails(detail)

	return
}

func NewCustom405Err(ctx context.Context, customErrCode string, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusMethodNotAllowed,
		customErrCode).
		WithErrorDetails(detail)

	return
}

func NewCustom409Err(ctx context.Context, customErrCode string, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusConflict,
		customErrCode).
		WithErrorDetails(detail)

	return
}

func NewCustom500Err(ctx context.Context, customErrCode string, detail interface{}) (httpErr *rest.HTTPError) {
	httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError,
		customErrCode).
		WithErrorDetails(detail)

	return
}
