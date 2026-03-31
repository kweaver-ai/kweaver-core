// Package errors parse error
package errors

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	liberrors "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
)

func GetLanguageFromRequest(c *gin.Context) (string, bool) {
	language := c.Request.Header.Get("X-Language")
	language = strings.ToLower(language)
	language = strings.Trim(language, " ")

	if language == "" {
		return "", false
	}

	if language == "zh_tw" || language == "zh-tw" {
		return Languages[1], true
	}

	if strings.HasPrefix(language, "zh") {
		return Languages[0], true
	}

	return Languages[2], true
}

// ReplyError parse error
func ReplyError(c *gin.Context, err error) {
	switch iError := err.(type) {
	case *IError:
		language, ok := GetLanguageFromRequest(c)

		if ok {
			errMsg, ok := ErrorsMsg[iError.MainCode][language]
			if ok {
				iError.Description = errMsg[0]
				iError.Solution = errMsg[1]
			}
		}

		switch iError.MainCode {
		case InternalError:
			c.JSON(http.StatusInternalServerError, err)
			return
		case InvalidParameter, DuplicatedName, FileTypeNotSupported, FileSizeExceed, DuplicatedAdmin:
			c.JSON(http.StatusBadRequest, err)
			return
		case UnAuthorization:
			c.JSON(http.StatusUnauthorized, err)
			return
		case TaskNotFound, DagInsNotFound:
			c.JSON(http.StatusNotFound, err)
			return
		case NoPermission, OperationDenied, Forbidden:
			c.JSON(http.StatusForbidden, err)
			return
		case UnAvailable:
			c.JSON(http.StatusServiceUnavailable, err)
			return
		default:
			httpCode, ok := ErrorsHttpCode[iError.MainCode]
			if ok {
				c.JSON(httpCode, err)
				return
			}
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	case ExHTTPError:
		resp, _ := ExHTTPErrorParser(err)
		c.JSON(http.StatusInternalServerError, NewIError(InternalError, ErrorDepencyService, resp))
		return
	case liberrors.ExHTTPError:
		resp, _ := liberrors.ExHTTPErrorParser(err)
		c.JSON(http.StatusInternalServerError, NewIError(InternalError, ErrorDepencyService, resp))
		return
	case *liberrors.RestError:
		// 新错误码适配，带国际化Description操作
		liberrors.ReplyError(c, err)
		return
	default:
		c.JSON(http.StatusInternalServerError, NewIError(InternalError, "", nil))
		return
	}
}

// ParseError parse error
func ParseError(err error) error {
	switch err.(type) {
	case *IError:
		return err
	case ExHTTPError:
		resp, _ := ExHTTPErrorParser(err)
		return NewIError(InternalError, ErrorDepencyService, resp)
	case nil:
		return nil
	default:
		return NewIError(InternalError, "", err.Error())
	}
}
