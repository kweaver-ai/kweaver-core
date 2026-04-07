package drivenadapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

// 响应结构定义

// ApiResponse 统一响应包裹
type ApiResponse[T any] struct {
	Data T `json:"data"`
}

// ListResponse 列表响应
type ListResponse[T any] struct {
	Count int `json:"count"`
	Data  []T `json:"data"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	StorageID        string `json:"storage_id"`
	StorageName      string `json:"storage_name"`
	VendorType       string `json:"vendor_type"`
	Endpoint         string `json:"endpoint"`
	BucketName       string `json:"bucket_name"`
	Region           string `json:"region"`
	IsDefault        bool   `json:"is_default"`
	IsEnabled        bool   `json:"is_enabled"`
	InternalEndpoint string `json:"internal_endpoint"`
}

// PresignedRequest 预签名请求信息
type PresignedRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	FormField   map[string]string `json:"form_field,omitempty"`
	Body        string            `json:"body,omitempty"`
	RequestBody string            `json:"request_body,omitempty"` // 完成分片上传时返回的 XML body
}

// InitMultiUploadResponse 分片上传初始化响应
type InitMultiUploadResponse struct {
	UploadID string `json:"upload_id"`
	PartSize int64  `json:"part_size"`
	Key      string `json:"key"`
}

// UploadPartResponse 分片上传URL响应
type UploadPartResponse struct {
	AuthRequest map[string]PresignedRequest `json:"authrequest"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Solution    string `json:"solution"`
	Cause       string `json:"cause"`
}

// 错误定义
var (
	ErrStorageNotFound  = errors.New("storage not found")
	ErrObjectNotFound   = errors.New("object not found")
	ErrInvalidParam     = errors.New("invalid parameter")
	ErrInternalError    = errors.New("internal server error")
	ErrServiceNotReady  = errors.New("service not ready")
	ErrNoAvailableStore = errors.New("no available storage")
)

// mapErrorResponse 映射错误响应
func mapErrorResponse(errResp ErrorResponse) error {
	switch errResp.Code {
	case "404031101":
		return fmt.Errorf("%w: %s", ErrStorageNotFound, errResp.Description)
	case "404031100":
		return fmt.Errorf("%w: %s", ErrObjectNotFound, errResp.Description)
	case "400031101":
		return fmt.Errorf("%w: %s", ErrInvalidParam, errResp.Description)
	case "500031100":
		return fmt.Errorf("%w: %s", ErrInternalError, errResp.Description)
	case "503031100":
		return fmt.Errorf("%w: %s", ErrServiceNotReady, errResp.Description)
	default:
		return fmt.Errorf("oss gateway error: %s - %s", errResp.Code, errResp.Message)
	}
}

// encodeKey URL编码对象Key
func encodeKey(key string) string {
	return url.PathEscape(key)
}

// ossGatewayBackend OssGateway Backend 实现
type ossGatewayBackend struct {
	address string
	client  otelHttp.HTTPClient
}

var (
	backendOnce sync.Once
	backendOg   OssGateWay
)

// NewOssGatewayBackend 创建 OssGateway Backend 实例
func NewOssGatewayBackend() OssGateWay {
	backendOnce.Do(func() {
		config := common.NewConfig()
		backendOg = &ossGatewayBackend{
			address: fmt.Sprintf("http://%s:%d", config.OssGatewayBackend.Host, config.OssGatewayBackend.Port),
			client:  NewOtelHTTPClient(),
		}
	})
	return backendOg
}

// parseApiResponse 解析API响应
func (og *ossGatewayBackend) parseApiResponse(ctx context.Context, respParam interface{}, target interface{}) error {
	data, ok := respParam.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response format: expected map[string]interface{}")
	}

	// 检查是否有错误
	if _, exists := data["code"]; exists {
		var errResp ErrorResponse
		utils.ParseInterface(respParam, &errResp)
		return mapErrorResponse(errResp)
	}

	// 解析data字段
	if dataField, exists := data["data"]; exists {
		utils.ParseInterface(dataField, target)
		return nil
	}

	// 直接解析（兼容非包裹格式）
	utils.ParseInterface(respParam, target)
	return nil
}

// parseListResponse 解析列表响应
func (og *ossGatewayBackend) parseListResponse(ctx context.Context, respParam interface{}) (*ListResponse[StorageInfo], error) {
	var listResp ListResponse[StorageInfo]
	data, ok := respParam.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: expected map[string]interface{}")
	}

	// 检查是否有错误
	if _, exists := data["code"]; exists {
		var errResp ErrorResponse
		utils.ParseInterface(respParam, &errResp)
		return nil, mapErrorResponse(errResp)
	}

	utils.ParseInterface(respParam, &listResp)
	return &listResp, nil
}

// GetAvaildOSS 获取可用对象存储ID
func (og *ossGatewayBackend) GetAvaildOSS(ctx context.Context) (string, error) {
	target := fmt.Sprintf("%s/api/v1/storages?enabled=true", og.address)

	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetAvaildOSS] request failed: %s %s", target, err.Error())
		return "", err
	}

	listResp, err := og.parseListResponse(ctx, respParam)
	if err != nil {
		return "", err
	}

	if listResp.Count == 0 || len(listResp.Data) == 0 {
		return "", ErrNoAvailableStore
	}

	// 优先返回默认存储
	for _, storage := range listResp.Data {
		if storage.IsDefault {
			return storage.StorageID, nil
		}
	}

	// 无默认存储，返回第一个可用存储
	return listResp.Data[0].StorageID, nil
}

// SimpleUpload 小文件上传
func (og *ossGatewayBackend) SimpleUpload(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader) error {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/upload/%s/%s?request_method=PUT&internal_request=%t", og.address, ossID, encodedKey, internalRequest)

	var resp ApiResponse[PresignedRequest]
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] get presigned url failed: %s %s", target, err.Error())
		return err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] parse response failed: %s", err.Error())
		return err
	}

	if resp.Data.URL == "" {
		return fmt.Errorf("invalid response: missing URL")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] read file failed: %s", err.Error())
		return err
	}

	_, _, err = og.client.OSSClient(ctx, resp.Data.URL, resp.Data.Method, resp.Data.Headers, &data)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] upload to oss failed: %s", err.Error())
		return err
	}

	return nil
}

// UploadFile 上传文件（自动选择简单上传或分片上传）
func (og *ossGatewayBackend) UploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error {
	if size <= 20*1024*1024 {
		return og.SimpleUpload(ctx, ossID, key, internalRequest, file)
	}
	return og.multiUploadFile(ctx, ossID, key, internalRequest, file, size)
}

// multiUploadFile 分片上传
func (og *ossGatewayBackend) multiUploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error {
	var (
		partMinSize int64 = 20 * 1024 * 1024
		partMaxSize int64 = 20 * 1024 * 1024
		partMaxNum  int64 = 10000
		partSize    int64
		partCount   int64
		fileSize    = size
		eTags       = make(map[string]string, 0)
	)

	// 计算分片大小和数量
	for {
		partSize += partMinSize
		if partSize > partMaxSize {
			return errors.New("file too long")
		}
		partCount = fileSize / partSize
		if fileSize == 0 || fileSize%partSize != 0 {
			partCount++
		}
		if partCount <= partMaxNum {
			break
		}
	}

	// 初始化分片上传
	uploadInfo, err := og.initMultiUpload(ctx, ossID, key, internalRequest, size)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[multiUploadFile] init multi upload failed: %s", err.Error())
		return err
	}

	// 使用服务器返回的分片大小
	if uploadInfo.PartSize > 0 {
		partSize = uploadInfo.PartSize
		partCount = fileSize / partSize
		if fileSize == 0 || fileSize%partSize != 0 {
			partCount++
		}
	}

	partFile := make([]byte, partSize)
	for i := int64(1); i <= partCount; i++ {
		var partFileSize int
		partFileSize, err = file.Read(partFile)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("[multiUploadFile] read file error: %s", err.Error())
			return err
		}

		eTag, err := og.uploadPart(ctx, ossID, key, uploadInfo.UploadID, i, partFile[:partFileSize], internalRequest)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("[multiUploadFile] upload part failed: %s", err.Error())
			return err
		}

		strPartID := strconv.FormatInt(i, 10)
		eTags[strPartID] = eTag
	}

	// 完成分片上传
	if err := og.completeMultiUpload(ctx, ossID, key, uploadInfo.UploadID, eTags, internalRequest); err != nil {
		traceLog.WithContext(ctx).Warnf("[multiUploadFile] complete multi upload failed: %s", err.Error())
		return err
	}

	return nil
}

// initMultiUpload 初始化分片上传
func (og *ossGatewayBackend) initMultiUpload(ctx context.Context, ossID, key string, internalRequest bool, size int64) (*InitMultiUploadResponse, error) {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/initmultiupload/%s/%s?size=%d&internal_request=%t", og.address, ossID, encodedKey, size, internalRequest)

	var resp ApiResponse[InitMultiUploadResponse]
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[initMultiUpload] request failed: %s %s", target, err.Error())
		return nil, err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// uploadPart 上传分片
func (og *ossGatewayBackend) uploadPart(ctx context.Context, ossID, key, uploadID string, partID int64, partFile []byte, internalRequest bool) (string, error) {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/uploadpart/%s/%s", og.address, ossID, encodedKey)

	// 构造请求体
	reqBody := map[string]interface{}{
		"upload_id":        uploadID,
		"part_id":          []int64{partID},
		"internal_request": internalRequest,
	}

	var resp ApiResponse[UploadPartResponse]
	_, respParam, err := og.client.Post(ctx, target, nil, reqBody)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[uploadPart] get presigned url failed: %s %s", target, err.Error())
		return "", err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return "", err
	}

	// 获取对应分片的预签名请求
	strPartID := strconv.FormatInt(partID, 10)
	presignedReq, ok := resp.Data.AuthRequest[strPartID]
	if !ok {
		return "", fmt.Errorf("presigned url not found for part %d", partID)
	}

	// 上传分片
	respHeaders, _, err := og.client.OSSClient(ctx, presignedReq.URL, presignedReq.Method, presignedReq.Headers, &partFile)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[uploadPart] upload to oss failed: %s", err.Error())
		return "", err
	}

	// 获取ETag (保留双引号，OSS完成分片上传需要)
	eTag := respHeaders.Get("Etag")
	if eTag == "" {
		return "", fmt.Errorf("missing ETag in response")
	}
	// 确保ETag包含双引号
	if !strings.HasPrefix(eTag, "\"") {
		eTag = "\"" + eTag + "\""
	}
	return eTag, nil
}

// completeMultiUpload 完成分片上传
func (og *ossGatewayBackend) completeMultiUpload(ctx context.Context, ossID, key, uploadID string, eTags map[string]string, internalRequest bool) error {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/completeupload/%s/%s?upload_id=%s&internal_request=%t", og.address, ossID, encodedKey, uploadID, internalRequest)

	var resp ApiResponse[PresignedRequest]
	_, respParam, err := og.client.Post(ctx, target, nil, eTags)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[completeMultiUpload] get complete url failed: %s %s", target, err.Error())
		return err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return err
	}

	// 执行完成上传请求
	// 优先使用 request_body (完成分片上传返回)，否则使用 body
	requestBody := resp.Data.RequestBody
	if requestBody == "" {
		requestBody = resp.Data.Body
	}
	bodyByte := []byte(requestBody)
	_, _, err = og.client.OSSClient(ctx, resp.Data.URL, resp.Data.Method, resp.Data.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[completeMultiUpload] complete upload failed: %s", err.Error())
		return err
	}

	return nil
}

// DownloadFile 下载文件到缓冲区
func (og *ossGatewayBackend) DownloadFile(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) ([]byte, error) {
	var (
		i        int64  = 1
		start    int64  = 0
		end      int64  = 4194303
		partSize int64  = 4194304
		retry    int    = 0
		buff     []byte = make([]byte, 0)
	)

	// 获取文件大小
	fileSize, err := og.GetObjectMeta(ctx, ossID, key, internalRequest, opts...)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DownloadFile] get object meta failed: %s", err.Error())
		return buff, err
	}

	// 下载次数
	_downloadTime := float64(fileSize) / float64(partSize)
	downloadTime := int64(1)
	if _downloadTime > 1 {
		downloadTime = int64(_downloadTime) + 1
	}

	for i <= downloadTime {
		data, isLoss, err := og.downloadFileByFrag(ctx, ossID, key, internalRequest, start, end, partSize, fileSize, opts...)
		if err != nil {
			return buff, err
		}

		if isLoss {
			retry++
			if retry == 3 {
				return buff, errors.New("[DownloadFile] fragment download file byte loss")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		retry = 0
		buff = append(buff, data...)
		start = end + 1
		end = end + partSize

		if i == downloadTime-1 {
			end = fileSize
		}
		i++
	}

	if int64(len(buff)) != fileSize {
		traceLog.WithContext(ctx).Warnf("file may be broken, filesize: %d, download size: %d", fileSize, len(buff))
		return buff, fmt.Errorf("[DownloadFile] file may be broken, filesize: %d, download size: %d", fileSize, len(buff))
	}

	return buff, nil
}

// downloadFileByFrag 分片下载
func (og *ossGatewayBackend) downloadFileByFrag(ctx context.Context, ossID, key string, internalRequest bool, start, end, partSize, fileSize int64, opts ...OssOpt) ([]byte, bool, error) {
	var buff = make([]byte, 0)
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/download/%s/%s?internal_request=%t", og.address, ossID, encodedKey, internalRequest)

	var resp ApiResponse[PresignedRequest]
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[downloadFileByFrag] get presigned url failed: %s %s", target, err.Error())
		return buff, true, err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return buff, true, err
	}

	// 文件小于4M时，下载原本大小
	if fileSize <= end {
		end = fileSize - 1
	}

	// 分片范围
	if resp.Data.Headers == nil {
		resp.Data.Headers = make(map[string]string)
	}
	resp.Data.Headers["Range"] = fmt.Sprintf("bytes=%v-%v", start, end)

	bodyByte := []byte(resp.Data.Body)
	_, buff, err = og.client.OSSClient(ctx, resp.Data.URL, resp.Data.Method, resp.Data.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[downloadFileByFrag] download from oss failed: %s", err.Error())
		return buff, true, err
	}

	// 检验下载的分片文件是否完整
	isLoss := utils.IsByteLoss(len(buff), start, end, partSize, fileSize)
	if isLoss {
		traceLog.WithContext(ctx).Warnf("[downloadFileByFrag] download fragment file incomplete: start:%d-end:%d", start, end)
		return buff, isLoss, nil
	}

	return buff, isLoss, nil
}

// DownloadFile2Local 下载文件到本地
func (og *ossGatewayBackend) DownloadFile2Local(ctx context.Context, ossID, key string, internalRequest bool, filePath string, opts ...OssOpt) (int64, error) {
	data, err := og.DownloadFile(ctx, ossID, key, internalRequest, opts...)
	fileSize := int64(len(data))
	if err != nil {
		return fileSize, err
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DownloadFile2Local] create file failed: %s", err.Error())
		return fileSize, err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DownloadFile2Local] write file failed: %s", err.Error())
		return fileSize, err
	}

	return fileSize, nil
}

// DeleteFile 删除文件
func (og *ossGatewayBackend) DeleteFile(ctx context.Context, ossID, key string, internalRequest bool) error {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/delete/%s/%s?internal_request=%t", og.address, ossID, encodedKey, internalRequest)

	var resp ApiResponse[PresignedRequest]
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DeleteFile] get presigned url failed: %s %s", target, err.Error())
		return err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return err
	}

	bodyByte := []byte(resp.Data.Body)
	_, _, err = og.client.OSSClient(ctx, resp.Data.URL, resp.Data.Method, resp.Data.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DeleteFile] delete from oss failed: %s", err.Error())
		return err
	}

	return nil
}

// GetDownloadURL 获取下载URL
func (og *ossGatewayBackend) GetDownloadURL(ctx context.Context, ossID, key string, expires int64, internalRequest bool, opts ...OssOpt) (string, error) {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/download/%s/%s?internal_request=%t", og.address, ossID, encodedKey, internalRequest)
	if expires > 0 {
		target += fmt.Sprintf("&expires=%d", expires)
	}

	var resp ApiResponse[PresignedRequest]
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetDownloadURL] get presigned url failed: %s %s", target, err.Error())
		return "", err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return "", err
	}

	return resp.Data.URL, nil
}

// GetObjectMeta 获取对象元数据
func (og *ossGatewayBackend) GetObjectMeta(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) (int64, error) {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/head/%s/%s?internal_request=%t", og.address, ossID, encodedKey, internalRequest)

	var resp ApiResponse[PresignedRequest]
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetObjectMeta] get presigned url failed: %s %s", target, err.Error())
		return -1, err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return -1, err
	}

	bodyByte := []byte(resp.Data.Body)
	respHeader, _, err := og.client.OSSClient(ctx, resp.Data.URL, resp.Data.Method, resp.Data.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetObjectMeta] head request failed: %s", err.Error())
		return -1, err
	}

	var contentLength int64 = -1
	if comma := respHeader.Get("Content-Length"); len(comma) != 0 {
		contentLength, _ = strconv.ParseInt(comma, 10, 64)
	}

	return contentLength, nil
}

// GetUploadReq 获取上传请求信息
func (og *ossGatewayBackend) GetUploadReq(ctx context.Context, ossID, key string, expires int64, internalRequest bool) (*UploadRequest, error) {
	encodedKey := encodeKey(key)
	target := fmt.Sprintf("%s/api/v1/upload/%s/%s?request_method=PUT&internal_request=%t", og.address, ossID, encodedKey, internalRequest)
	if expires > 0 {
		target += fmt.Sprintf("&expires=%d", expires)
	}

	var resp ApiResponse[PresignedRequest]
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetUploadReq] get presigned url failed: %s %s", target, err.Error())
		return nil, err
	}

	if err := og.parseApiResponse(ctx, respParam, &resp.Data); err != nil {
		return nil, err
	}

	if resp.Data.URL == "" {
		return nil, fmt.Errorf("invalid response: missing URL")
	}

	return &UploadRequest{
		Method:  resp.Data.Method,
		URL:     resp.Data.URL,
		Headers: resp.Data.Headers,
		Expires: expires,
	}, nil
}

// NewReader 创建Reader
func (og *ossGatewayBackend) NewReader(ossID string, ossKey string, opts ...OssOpt) *Reader {
	return &Reader{
		og:     og,
		ossID:  ossID,
		ossKey: ossKey,
		opts:   opts,
		cache:  map[string]string{},
	}
}
