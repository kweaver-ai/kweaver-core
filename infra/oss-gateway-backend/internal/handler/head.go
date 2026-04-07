package handler

import (
	"net/url"
	"oss-gateway/internal/service"
	"oss-gateway/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HeadHandler struct {
	service service.URLService
}

func NewHeadHandler(service service.URLService) *HeadHandler {
	return &HeadHandler{service: service}
}

func (h *HeadHandler) GetHeadURL(c *gin.Context) {
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

	presignedURL, err := h.service.GetHeadURL(c.Request.Context(), storageID, decodedKey, expires, internalRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, presignedURL)
}

func (h *HeadHandler) BatchGetHeadURL(c *gin.Context) {
	storageID := c.Param("storageId")

	var req struct {
		Keys            []string `json:"keys" binding:"required"`
		InternalRequest bool     `json:"internal_request"`
		Expires         int64    `json:"expires"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if len(req.Keys) > 100 {
		response.TooManyKeys(c, 100)
		return
	}

	urls, err := h.service.BatchGetHeadURL(c.Request.Context(), storageID, req.Keys, req.Expires, req.InternalRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, urls)
}
