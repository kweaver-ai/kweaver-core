// Package rest 响应处理
// @file rest.go
// @description: 响应处理
package rest

import (
	"errors"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	validatorv "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	errorwrap "github.com/pkg/errors"
)

const (
	ContentTypeKey  = "Content-Type"
	ContentTypeJSON = "application/json"
)

// ReplyOK 响应成功
func ReplyOK(c *gin.Context, statusCode int, body interface{}) {
	var (
		bodyStr string
		err     error
	)

	if body != nil {
		bodyStr, err = sonic.MarshalString(body)
		if err != nil {
			logger.DefaultLogger().Errorf("marshal body error: %v", err)
			statusCode = http.StatusInternalServerError
			ctx := c.Request.Context()
			bodyStr = myErr.DefaultHTTPError(ctx, statusCode, err.Error()).Error()
		}
	}

	c.Writer.Header().Set(ContentTypeKey, ContentTypeJSON)
	c.String(statusCode, bodyStr)
}

// ReplyError 响应错误
func ReplyError(c *gin.Context, err error) {
	if err != nil {
		errWithStack := errorwrap.WithStack(err)
		logger.DefaultLogger().Debug("Error:", errWithStack.Error())
	}
	var httpCode int
	ctx := c.Request.Context()
	var body string
	switch e := err.(type) {
	case *ExHTTPError:
		httpCode = e.HTTPCode
		body = e.Error()
	default:
		httpError := &myErr.HTTPError{}
		vErr := make(validator.ValidationErrors, 0)
		if errors.As(err, &httpError) {
			httpCode = httpError.HTTPCode
			body = err.Error()
		} else if errors.As(err, &vErr) {
			httpCode = http.StatusBadRequest
			if len(vErr) > 0 {
				extCode := validatorv.TagToErrorType[vErr[0].Tag()]
				body = myErr.NewHTTPError(ctx, http.StatusBadRequest, extCode, vErr[0].Error()).Error()
			} else {
				body = myErr.DefaultHTTPError(ctx, httpCode, err.Error()).Error()
			}
		} else {
			httpCode = http.StatusInternalServerError
			body = myErr.DefaultHTTPError(ctx, httpCode, err.Error()).Error()
		}
	}
	c.Writer.Header().Set(ContentTypeKey, ContentTypeJSON)
	c.String(httpCode, body)
}

// ExHTTPError 依赖服务的错误码
type ExHTTPError struct {
	HTTPCode int
	Body     []byte
}

func (e *ExHTTPError) Error() string {
	return string(e.Body)
}

// ReplyWithExecutionMode 根据执行模式处理响应
func ReplyWithExecutionMode(c *gin.Context, resp interface{}, err error) {
	// 判断流式响应模式
	ctx := c.Request.Context()
	executionMode := common.GetExecutionModeFromCtx(ctx)
	streamingMode, _ := common.GetStreamingModeFromCtx(ctx)
	switch executionMode {
	case interfaces.ExecutionModeStream:
		switch streamingMode {
		case interfaces.StreamingModeSSE:
			if err == nil {
				return
			}
			switch e := err.(type) {
			case *ExHTTPError:
				c.SSEvent("error", e)
			case *myErr.HTTPError:
				c.SSEvent("error", e)
			default:
				err = myErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
				c.SSEvent("error", e)
			}
			return
		case interfaces.StreamingModeHTTP:
			if err != nil {
				ReplyError(c, err)
				return
			}
		}
	case interfaces.ExecutionModeAsync, interfaces.ExecutionModeSync:
		if err != nil {
			ReplyError(c, err)
			return
		}
		ReplyOK(c, http.StatusOK, resp)
	}
}
