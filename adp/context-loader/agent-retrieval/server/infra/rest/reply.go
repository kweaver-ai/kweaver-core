// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package rest 响应处理
// @file rest.go
// @description: 响应处理
package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
	errorwrap "github.com/pkg/errors"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	myErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/localize"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/logger"
	validatorv "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/validator"
)

const (
	ContentTypeKey  = "Content-Type"
	ContentTypeJSON = "application/json"
)

// ReplyOK 响应成功；根据 context 中的 response_format 决定输出 JSON 或 TOON
func ReplyOK(c *gin.Context, statusCode int, body interface{}) {
	format := FormatJSON
	if v, ok := common.GetResponseFormatFromCtx(c.Request.Context()); ok {
		if f, ok := v.(ResponseFormat); ok {
			format = f
		}
	}

	contentType := ContentTypeJSON
	var bodyBytes []byte
	var err error
	if body != nil {
		contentType, bodyBytes, err = MarshalResponse(format, body)
		if err != nil {
			logger.DefaultLogger().Errorf("marshal body error: %v", err)
			statusCode = http.StatusInternalServerError
			ctx := c.Request.Context()
			bodyStr := myErr.DefaultHTTPError(ctx, statusCode, err.Error()).Error()
			c.Writer.Header().Set(ContentTypeKey, ContentTypeJSON)
			c.String(statusCode, bodyStr)
			return
		}
	}

	if len(bodyBytes) > 0 {
		c.Data(statusCode, contentType, bodyBytes)
	} else {
		c.Writer.Header().Set(ContentTypeKey, contentType)
		c.String(statusCode, "")
	}
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
				// 生成友好的多语言错误详情信息（大模型可理解）
				friendlyDetails := formatValidatorErrorDetails(ctx, vErr[0])
				body = myErr.NewHTTPError(ctx, http.StatusBadRequest, extCode, friendlyDetails).Error()
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

// formatValidatorErrorDetails 格式化 validator 错误详情信息，生成大模型可理解的多语言友好错误信息
func formatValidatorErrorDetails(ctx context.Context, err validator.FieldError) string {
	// 获取语言设置（格式：zh-CN 或 en-US）
	lang := common.GetLanguageFromCtx(ctx)
	// 将 zh-CN 转换为 zh_CN 格式（国际化系统使用的格式）
	langKey := strings.ReplaceAll(lang, "-", "_")
	tr := localize.NewI18nTranslator(langKey)

	fieldName := err.Field()
	tag := err.Tag()
	param := err.Param()
	currentValue := ""
	if err.Value() != nil {
		currentValue = fmt.Sprintf("%v", err.Value())
	}

	// 根据不同的验证标签生成友好的多语言错误信息
	var templateKey string
	var formatArgs []interface{}

	switch tag {
	case "min", "gte", "gt":
		if currentValue != "" {
			templateKey = "desc.ValidationDetailMin"
			formatArgs = []interface{}{fieldName, currentValue, param, param}
		} else {
			templateKey = "desc.ValidationDetailMinNoValue"
			formatArgs = []interface{}{fieldName, param}
		}
	case "max", "lte", "lt":
		if currentValue != "" {
			templateKey = "desc.ValidationDetailMax"
			formatArgs = []interface{}{fieldName, currentValue, param, param}
		} else {
			templateKey = "desc.ValidationDetailMaxNoValue"
			formatArgs = []interface{}{fieldName, param}
		}
	case "required":
		templateKey = "desc.ValidationDetailRequired"
		formatArgs = []interface{}{fieldName}
	case "oneof":
		templateKey = "desc.ValidationDetailOneof"
		// 格式化选项列表：中文使用"、"，英文使用", "
		options := strings.ReplaceAll(param, " ", "、")
		if langKey == "en_US" {
			options = strings.ReplaceAll(param, " ", ", ")
		}
		formatArgs = []interface{}{fieldName, options}
	default:
		// 其他验证标签，使用通用格式
		if currentValue != "" {
			templateKey = "desc.ValidationDetailUnknown"
			formatArgs = []interface{}{fieldName, currentValue, tag, param}
		} else {
			templateKey = "desc.ValidationDetailUnknownNoValue"
			formatArgs = []interface{}{fieldName, tag, param}
		}
	}

	// 获取翻译模板并格式化
	template := tr.Trans(templateKey)
	return fmt.Sprintf(template, formatArgs...)
}
