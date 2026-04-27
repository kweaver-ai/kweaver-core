package handler

import (
	"net/url"
	"oss-gateway/internal/service"
	"oss-gateway/pkg/adapter"
	"oss-gateway/pkg/response"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PartInfo = adapter.PartInfo

type UploadHandler struct {
	service service.URLService
}

func NewUploadHandler(service service.URLService) *UploadHandler {
	return &UploadHandler{service: service}
}

func (h *UploadHandler) GetUploadURL(c *gin.Context) {
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

	requestMethod := c.DefaultQuery("request_method", "PUT")
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

	presignedURL, err := h.service.GetUploadURL(c.Request.Context(), storageID, decodedKey, requestMethod, expires, internalRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, presignedURL)
}

func (h *UploadHandler) InitMultipartUpload(c *gin.Context) {
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

	sizeStr := c.Query("size")
	if sizeStr == "" {
		response.InvalidParam(c, "size is required")
		return
	}

	fileSize, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		response.InvalidParam(c, "invalid size")
		return
	}

	result, err := h.service.InitMultipartUpload(c.Request.Context(), storageID, decodedKey, fileSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UploadHandler) GetUploadPartURLs(c *gin.Context) {
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

	var req struct {
		UploadID        string `json:"upload_id" binding:"required"`
		PartID          []int  `json:"part_id" binding:"required"`
		InternalRequest bool   `json:"internal_request"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	expiresStr := c.Query("expires")
	var expires int64
	if expiresStr != "" {
		expires, err = strconv.ParseInt(expiresStr, 10, 64)
		if err != nil {
			response.InvalidParam(c, "invalid expires")
			return
		}
	}

	urls, err := h.service.GetUploadPartURLs(c.Request.Context(), storageID, decodedKey, req.UploadID, req.PartID, expires, req.InternalRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	result := make(map[string]interface{})
	for partNum, url := range urls {
		result[strconv.Itoa(partNum)] = url
	}

	response.Success(c, map[string]interface{}{
		"authrequest": result,
	})
}

func (h *UploadHandler) CompleteMultipartUpload(c *gin.Context) {
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

	uploadID := c.Query("upload_id")
	if uploadID == "" {
		response.InvalidParam(c, "upload_id is required")
		return
	}

	var etagMap map[string]string
	if err := c.ShouldBindJSON(&etagMap); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var parts []PartInfo
	for partNumStr, etag := range etagMap {
		partNum, err := strconv.Atoi(partNumStr)
		if err != nil {
			continue
		}
		parts = append(parts, PartInfo{
			PartNumber: partNum,
			ETag:       etag,
		})
	}

	// Sort parts by PartNumber to ensure ascending order
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	presignedURL, err := h.service.CompleteMultipartUpload(c.Request.Context(), storageID, decodedKey, uploadID, parts)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	responseData := map[string]interface{}{
		"method":  presignedURL.Method,
		"url":     presignedURL.URL,
		"headers": presignedURL.Headers,
	}

	if presignedURL.Body != "" {
		responseData["request_body"] = presignedURL.Body
	}

	response.Success(c, responseData)
}
