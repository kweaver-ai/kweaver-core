// Package errors parse error
package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReplyError parse error
func ReplyError(c *gin.Context, err error) {
	switch RError := err.(type) {
	case *RestError:
		if RError.HTTPCode <= 0 {
			replyByMaincode(c, RError.MainCode, err)
		} else {
			replyByHttpCode(c, RError.HTTPCode, err)
		}
	}
}

func replyByHttpCode(c *gin.Context, httpCode int, err error) {
	if http.StatusText(httpCode) != "" {
		c.JSON(httpCode, err)
		return
	}
	c.JSON(http.StatusInternalServerError, err)
}

func replyByMaincode(c *gin.Context, mainCode string, err error) {
	switch mainCode {
	case PErrorBadRequest:
		c.JSON(http.StatusBadRequest, err)
		return
	case PErrorUnauthorized:
		c.JSON(http.StatusUnauthorized, err)
		return
	case PErrorForbidden:
		c.JSON(http.StatusForbidden, err)
		return
	case PErrorNotFound:
		c.JSON(http.StatusNotFound, err)
		return
	case PErrorMethodNotAllowed:
		c.JSON(http.StatusMethodNotAllowed, err)
		return
	case PErrorConflict:
		c.JSON(http.StatusConflict, err)
		return
	case PErrorInternalServerError:
		c.JSON(http.StatusInternalServerError, err)
		return
	case PErrorNotImplemented:
		c.JSON(http.StatusNotImplemented, err)
		return
	case PErrorServiceUnavailable:
		c.JSON(http.StatusServiceUnavailable, err)
		return
	case PErrorLoopDetected:
		c.JSON(http.StatusLoopDetected, err)
		return
	default:
		c.JSON(http.StatusInternalServerError, err)
		return
	}
}
