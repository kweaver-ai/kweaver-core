package handler

import (
	"net/url"
	"oss-gateway/internal/service"
	"oss-gateway/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteHandler struct {
	service service.URLService
}

func NewDeleteHandler(service service.URLService) *DeleteHandler {
	return &DeleteHandler{service: service}
}

func (h *DeleteHandler) GetDeleteURL(c *gin.Context) {
	storageID := c.Param("storageId")
	objectKey := c.Param("key")

	// Remove leading slash from wildcard parameter
	if len(objectKey) > 0 && objectKey[0] == '/' {
		objectKey = objectKey[1:]
	}

	decodedKey, err := url.PathUnescape(objectKey)
	if err != nil {
		response.InvalidParam(c, "invalid object key")
		return
	}

	expiresStr := c.Query("expires")
	internalRequest := c.DefaultQuery("internal_request", "false") == "true"

	var expires int64
	if expiresStr != "" {
		expires, err = strconv.ParseInt(expiresStr, 10, 64)
		if err != nil {
			response.InvalidParam(c, "invalid expires")
			return
		}
	}

	presignedURL, err := h.service.GetDeleteURL(c.Request.Context(), storageID, decodedKey, expires, internalRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, presignedURL)
}
