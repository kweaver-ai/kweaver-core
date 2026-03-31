package drivenadapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
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

const DefaultOSSNotFound = float64(404031002)

// RequestBody oss接口requestbody信息
type (
	RequestBody struct {
		Method    string            `json:"method,omitempty"`
		URL       string            `json:"url,omitempty"`
		Headers   map[string]string `json:"headers,omitempty"`
		FormFiled map[string]string `json:"form_field,omitempty"`
		Body      string            `json:"request_body,omitempty"`
	}

	MultiUploadInfo struct {
		PartSize string `json:"part_size,omitempty"`
		UploaID  string `json:"upload_id,omitempty"`
	}

	OAuth struct {
		ClientID     string `json:"client_id,omitempty"`
		ClientSecret string `json:"client_secret,omitempty"`
	}

	Address struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	}

	OSSInfo struct {
		Default bool   `json:"default,omitempty"`
		Enabled bool   `json:"enabled,omitempty"`
		OSSID   string `json:"id,omitempty"`
		Name    string `json:"name,omitempty"`
	}

	// UploadRequest 上传请求信息，用于先触发后上传场景
	UploadRequest struct {
		Method  string            `json:"method,omitempty"`  // HTTP 方法，如 PUT
		URL     string            `json:"url,omitempty"`     // 上传目标 URL
		Headers map[string]string `json:"headers,omitempty"` // 请求头
		Expires int64             `json:"expires,omitempty"` // 过期时间（秒）
	}
)

type OssOpt struct {
	StoragePrifix *bool
}

// OssGateWay oss接口
type OssGateWay interface {
	UploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error
	SimpleUpload(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader) error
	GetAvaildOSS(ctx context.Context) (string, error)
	DownloadFile2Local(ctx context.Context, ossID, key string, internalRequest bool, filePath string, opts ...OssOpt) (int64, error)
	DownloadFile(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) ([]byte, error)
	DeleteFile(ctx context.Context, ossID, key string, internalRequest bool) error
	GetDownloadURL(ctx context.Context, ossID, key string, expires int64, internalRequest bool, opts ...OssOpt) (string, error)
	GetObjectMeta(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) (int64, error)
	NewReader(ossID string, ossKey string, opts ...OssOpt) *Reader
	// GetUploadReq 获取上传请求信息，用于先触发后上传场景
	// 返回可供前端直接上传的目标信息（method, url, headers）
	GetUploadReq(ctx context.Context, ossID, key string, expires int64, internalRequest bool) (*UploadRequest, error)
}

type ossGatetway struct {
	address string
	client  otelHttp.HTTPClient
}

var (
	OgOnce sync.Once
	og     OssGateWay
)

// NewOssGateWay 创建oss服务
func NewOssGateWay() OssGateWay {
	OgOnce.Do(func() {
		config := common.NewConfig()

		switch config.Server.StorageBackend {
		case common.StorageBackendS3:
			og = NewOssGatewayS3()
		case common.StorageBackendOssGateway:
			og = &ossGatetway{
				address: fmt.Sprintf("http://%s:%s", config.OssGateWay.PrivateHost, config.OssGateWay.PrivatePort),
				client:  NewOtelHTTPClient(),
			}
		default: // StorageBackendOssGatewayBackend 或空（默认）
			og = NewOssGatewayBackend()
		}
	})
	return og
}

func (og *ossGatetway) appendOptQuery(url string, opts ...OssOpt) string {
	if len(opts) > 0 && opts[0].StoragePrifix != nil {
		return fmt.Sprintf("%s&storage_prefix=%v", url, *(opts[0].StoragePrifix))
	}

	return url
}

// UploadFile 上传文件
func (og *ossGatetway) UploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error {
	// 普通上传
	if size <= 20*1024*1024 {
		if err := og.SimpleUpload(ctx, ossID, key, internalRequest, file); err != nil {
			return err
		}
	} else {
		// 分片上传
		if err := og.MultiUploadFile(ctx, ossID, key, internalRequest, file, size); err != nil {
			return err
		}
	}
	return nil
}

// SimpleUpload 小文件上传
func (og *ossGatetway) SimpleUpload(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader) error {
	target := fmt.Sprintf("%s/api/ossgateway/v1/upload/%s/%s?request_method=PUT&internal_request=%t", og.address, ossID, key, internalRequest)
	requestBody := &RequestBody{}
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] get object requestbody failed, detail: %s %s", target, err.Error())
		return err
	}

	utils.ParseInterface(respParam, &requestBody)

	// Validate that we got a valid request body with URL
	if requestBody.URL == "" {
		err := fmt.Errorf("invalid response from OSS gateway: missing URL")
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] %s", err.Error())
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] read file err, detail: %s", err.Error())
		return err
	}
	_, _, err = og.client.OSSClient(ctx, requestBody.URL, requestBody.Method, requestBody.Headers, &data)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[SimpleUpload] send request err, detail: %s", err.Error())
		return err
	}
	return nil
}

// MultiUploadFile 分片上传
func (og *ossGatetway) MultiUploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error {
	var (
		partMinSize int64 = 20 * 1024 * 1024
		partMaxSize int64 = 20 * 1024 * 1024
		partMaxNum  int64 = 10000
		partSize    int64
		partCount   int64
		fileSize    = size
		eTags       = make(map[string]string, 0)
	)

	// 获得分片总数
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

	// 获取开始上传文件信息
	uploadInfo, err := og.InitMultiUpload(ctx, ossID, key, internalRequest, size)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[MultiUploadFile] InitMultiUpload, detail: %s", err.Error())
		return err
	}
	partFile := make([]byte, partSize)
	for i := int64(1); i <= partCount; i++ {
		var partFileSize int
		partFileSize, err = file.Read(partFile)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("[MultiUploadFile] read file error, detail: %s", err.Error())
			return err
		}
		eTag, err := og.UploadFragFile(ctx, ossID, key, uploadInfo.UploaID, i, 1, partFile[:partFileSize], internalRequest)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("[MultiUploadFile] UploadFragFile, detail: %s", err.Error())
			return err
		}

		// 存储etag信息
		strPartID := strconv.FormatInt(i, 10)
		eTags[strPartID] = eTag
	}
	// 完成文件上传协议
	if err := og.CompleteMultiUpload(ctx, ossID, key, uploadInfo.UploaID, eTags, internalRequest); err != nil {
		traceLog.WithContext(ctx).Warnf("[MultiUploadFile] CompleteMultiUpload, detail: %s", err.Error())
		return err
	}
	return nil
}

// InitMultiUpload 获取开始上传文件协议信息
func (og *ossGatetway) InitMultiUpload(ctx context.Context, ossID, key string, internalRequest bool, size int64) (*MultiUploadInfo, error) {
	target := fmt.Sprintf("%s/api/ossgateway/v1/initmultiupload/%s/%s?size=%d&internal_request=%t", og.address, ossID, key, size, internalRequest)
	// 获取开始上传协议，分块大小和上传id
	uploadInfo := &MultiUploadInfo{}
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[InitMultiUpload] get object requestbody failed, detail: %s %s", target, err.Error())
		return uploadInfo, err
	}

	utils.ParseInterface(respParam, &uploadInfo)
	return uploadInfo, nil
}

// UploadFragFile 上传分片文件
func (og *ossGatetway) UploadFragFile(ctx context.Context, ossID, key, uploadID string, partID, partNum int64, partFile []byte, internalRequest bool) (string, error) {
	target := fmt.Sprintf("%s/api/ossgateway/v1/uploadpart/%s/%s?part_id=%d&part_num=%d&upload_id=%s&internal_request=%t", og.address, ossID, key, partID, partNum, uploadID, internalRequest)
	// 获取开始上传协议，分块大小和上传id
	out := make(map[string]RequestBody, partNum)
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[UploadFragFile] get upload info failed, detail: %s %s", target, err.Error())
		return "", err
	}

	utils.ParseInterface(respParam, &out)
	// 开始分片上传
	strPartID := strconv.FormatInt(partID, 10)
	requestBody := out[strPartID]
	respHeaders, _, err := og.client.OSSClient(ctx, requestBody.URL, requestBody.Method, requestBody.Headers, &partFile)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[UploadFragFile] put file to oss failed, detail: %s %s", requestBody.URL, err.Error())
		return "", err
	}

	// 获取Etag 作为完成文件上传协议参数
	eTag := respHeaders.Get("Etag")
	// 不去除 \" 会导致上传文件协议出错
	return strings.Trim(eTag, "\""), nil
}

// CompleteMultiUpload 完成上传文件协议
func (og *ossGatetway) CompleteMultiUpload(ctx context.Context, ossID, key, uploadID string, eTags map[string]string, internalRequest bool) error {
	target := fmt.Sprintf("%s/api/ossgateway/v1/completeupload/%s/%s?upload_id=%s&internal_request=%t", og.address, ossID, key, uploadID, internalRequest)
	requestBody := &RequestBody{}
	_, respParam, err := og.client.Post(ctx, target, nil, eTags)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[CompleteMultiUpload] get complete multiupload info failed, detail: %s %s", target, err.Error())
		return err
	}

	utils.ParseInterface(respParam, &requestBody)
	bodyByte := []byte(requestBody.Body)
	_, _, err = og.client.OSSClient(ctx, requestBody.URL, requestBody.Method, requestBody.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[CompleteMultiUpload] complete multiupload failed, detail: %s %s", target, err.Error())
		return err
	}

	return nil
}

// GetObjectMeta 获取对象元数据
func (og *ossGatetway) GetObjectMeta(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) (int64, error) {
	target := fmt.Sprintf("%s/api/ossgateway/v1/head/%s/%s?internal_request=%t", og.address, ossID, key, internalRequest)
	target = og.appendOptQuery(target, opts...)
	requestBody := &RequestBody{}
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetObjectMeta] get object requestbody failed, detail: %s %s", target, err.Error())
		return -1, err
	}

	utils.ParseInterface(respParam, &requestBody)
	bodyByte := []byte(requestBody.Body)
	// 请求元数据接口
	respHeader, _, err := og.client.OSSClient(ctx, requestBody.URL, requestBody.Method, requestBody.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetObjectMeta] get object meta failed, detail: %s %s", target, err.Error())
		return -1, err
	}

	var contentLength int64 = -1
	if comma := respHeader.Get("Content-Length"); len(comma) != 0 {
		contentLength, _ = strconv.ParseInt(comma, 10, 64)
	}
	return contentLength, nil
}

// DownloadFile2Local 下载文件到本地
func (og *ossGatetway) DownloadFile2Local(ctx context.Context, ossID, key string, internalRequest bool, filePath string, opts ...OssOpt) (int64, error) {
	data, err := og.DownloadFile(ctx, ossID, key, internalRequest, opts...)
	fileSize := int64(len(data))
	if err != nil {
		return fileSize, err
	}
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DownloadFile2Local] file download failed when create tmp file, detail: %s", err.Error())
		return fileSize, err
	}
	_, err = f.Write(data)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DownloadFile2Local] write download bytes to file fail, detail: %s", err.Error())
		return fileSize, err
	}
	return fileSize, nil
}

// DownloadFile2Local 下载文件到缓冲区
func (og *ossGatetway) DownloadFile(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) ([]byte, error) {
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
		traceLog.WithContext(ctx).Warnf("[DownloadFile] get object meta info failed, detail: %s", err.Error())
		return buff, err
	}

	//下载次数
	_downloadTime, _ := strconv.ParseFloat(fmt.Sprintf("%.8f", float64(fileSize)/float64(partSize)), 64)
	downloadTime := math.Ceil(_downloadTime)

	for i <= int64(downloadTime) {
		data, isLoss, err := og.DownloadFileByFrag(ctx, ossID, key, internalRequest, start, end, partSize, fileSize, opts...)
		if err != nil {
			// 下载过程出错，直接返回
			return buff, err
		}
		// 判断字节存在缺失，重新下载，重试次数3次
		if isLoss {
			retry++
			if retry == 3 {
				return buff, errors.New("[DownloadFile] fragment download file byte loos")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 下载成功重置重试次数
		retry = 0
		// 添加下载内容进缓冲区
		buff = append(buff, data...)
		start = end + 1
		end = end + partSize

		// 最后一次下载
		if i == int64(downloadTime)-1 {
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

func (og *ossGatetway) GetDownloadURL(ctx context.Context, ossID, key string, expires int64, internalRequest bool, opts ...OssOpt) (string, error) {
	target := fmt.Sprintf("%s/api/ossgateway/v1/download/%s/%s?type=query_string&internal_request=%t", og.address, ossID, key, internalRequest)
	target = og.appendOptQuery(target, opts...)
	if expires != 0 {
		target += fmt.Sprintf("&Expires=%d", expires)
	}
	requestBody := &RequestBody{}
	_, respParam, err := og.client.Get(ctx, target, nil)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetDownloadURL] get downlaod url failed, detail: %s %s", target, err.Error())
		return "", err
	}

	utils.ParseInterface(respParam, &requestBody)
	return requestBody.URL, nil
}

// DownloadFileByFrag 分片下载指定字节文件
func (og *ossGatetway) DownloadFileByFrag(ctx context.Context, ossID, key string, internalRequest bool, start, end, partSize, fileSize int64, opts ...OssOpt) ([]byte, bool, error) {
	var buff = make([]byte, 0)
	target := fmt.Sprintf("%s/api/ossgateway/v1/download/%s/%s?internal_request=%t", og.address, ossID, key, internalRequest)
	target = og.appendOptQuery(target, opts...)
	requestBody := &RequestBody{}

	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DownloadFileByFrag] get downlaod fragment info failed, detail: %s %s", target, err.Error())
		return buff, true, err
	}

	//文件小于4M时，下载原本大小
	if fileSize <= end {
		end = fileSize - 1
	}

	utils.ParseInterface(respParam, &requestBody)
	bodyByte := []byte(requestBody.Body)

	// 分片范围
	requestBody.Headers["Range"] = fmt.Sprintf("bytes=%v-%v", start, end)
	_, buff, err = og.client.OSSClient(ctx, requestBody.URL, requestBody.Method, requestBody.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DownloadFileByFrag] download frag file failed, detail: %s %s", target, err.Error())
		return buff, true, err
	}

	// 检验下载的分片文件是否完整
	isLoss := utils.IsByteLoss(len(buff), start, end, partSize, fileSize)
	if isLoss {
		traceLog.WithContext(ctx).Warnf("[DownloadFileByFrag] download fragment file fail, detail: start:%d-end:%d bytes not completion", start, end)
		return buff, isLoss, nil
	}
	return buff, isLoss, nil
}

// getDefaultOSS 获取默认对象存储信息id
func (og *ossGatetway) getDefaultOSS(ctx context.Context) (string, error) {
	var ossID string
	target := fmt.Sprintf("%s/api/ossgateway/v1/default-storage", og.address)
	res := make(map[string]interface{}, 0)
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetDefaultOSS] get default oss failed, detail: %s %s", target, err.Error())
		return ossID, err
	}

	utils.ParseInterface(respParam, &res)

	// Safely extract storage_id with type assertion
	if storageID, ok := res["storage_id"]; ok {
		if ossID, ok = storageID.(string); ok {
			return ossID, nil
		}
		return "", fmt.Errorf("storage_id is not a string type")
	}

	return "", fmt.Errorf("storage_id not found in response")
}

// getObjectStorageInfos 获取本站点的对象存储信息
func (og *ossGatetway) getObjectStorageInfos(ctx context.Context, bizType string) ([]*OSSInfo, error) {
	var ossList = make([]*OSSInfo, 0)
	var target string
	if bizType != "" {
		target = fmt.Sprintf("%s/api/ossgateway/v1/objectstorageinfo?app=%v", og.address, bizType)
	} else {
		target = fmt.Sprintf("%s/api/ossgateway/v1/objectstorageinfo", og.address)
	}
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetObjectStorageInfos] get local oss failed, detail: %s %s", target, err.Error())
		return ossList, err
	}
	utils.ParseInterface(respParam, &ossList)
	return ossList, nil
}

// GetAvaildOSS 获取可用对象存储id
func (og *ossGatetway) GetAvaildOSS(ctx context.Context) (string, error) {
	defaultOSS, err := og.getDefaultOSS(ctx)
	if err != nil {
		parsedBody, _err := ExHTTPErrorParser(err)
		if _err != nil || parsedBody["code"] != DefaultOSSNotFound {
			return "", err
		}
	}

	if defaultOSS != "" {
		return defaultOSS, nil
	}

	// 获取本站点可用对象存储列表
	ossList, err := og.getObjectStorageInfos(ctx, "as")
	if err != nil {
		return "", err
	}
	if len(ossList) == 0 {
		return "", fmt.Errorf("no available oss")
	}
	for _, oss := range ossList {
		if oss.Enabled {
			return oss.OSSID, nil
		}
	}
	return "", fmt.Errorf("no available oss")
}

func (og *ossGatetway) DeleteFile(ctx context.Context, ossID, key string, internalRequest bool) error {
	target := fmt.Sprintf("%s/api/ossgateway/v1/delete/%s/%s?internal_request=%t", og.address, ossID, key, internalRequest)
	requestBody := &RequestBody{}

	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DeleteFile] get delete info failed, detail: %s %s", target, err.Error())
		return err
	}

	utils.ParseInterface(respParam, &requestBody)
	bodyByte := []byte(requestBody.Body)

	_, _, err = og.client.OSSClient(ctx, requestBody.URL, requestBody.Method, requestBody.Headers, &bodyByte)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[DeleteFile] delete file failed, detail: %s %s", target, err.Error())
		return err
	}

	return nil
}

// GetUploadReq 获取上传请求信息，用于先触发后上传场景
func (og *ossGatetway) GetUploadReq(ctx context.Context, ossID, key string, expires int64, internalRequest bool) (*UploadRequest, error) {
	target := fmt.Sprintf("%s/api/ossgateway/v1/upload/%s/%s?request_method=PUT&internal_request=%t", og.address, ossID, key, internalRequest)
	if expires > 0 {
		target += fmt.Sprintf("&Expires=%d", expires)
	}

	requestBody := &RequestBody{}
	_, respParam, err := og.client.Get(ctx, target, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetUploadReq] get upload request failed, detail: %s %s", target, err.Error())
		return nil, err
	}

	utils.ParseInterface(respParam, &requestBody)

	if requestBody.URL == "" {
		err := fmt.Errorf("invalid response from OSS gateway: missing URL")
		traceLog.WithContext(ctx).Warnf("[GetUploadReq] %s", err.Error())
		return nil, err
	}

	return &UploadRequest{
		Method:  requestBody.Method,
		URL:     requestBody.URL,
		Headers: requestBody.Headers,
		Expires: expires,
	}, nil
}

func (og *ossGatetway) NewReader(ossID string, ossKey string, opts ...OssOpt) *Reader {
	return &Reader{
		opts:   opts,
		og:     og,
		ossID:  ossID,
		ossKey: ossKey,
		cache:  map[string]string{},
	}
}

type Reader struct {
	opts   []OssOpt
	og     OssGateWay
	ossID  string
	ossKey string
	cache  map[string]string
}

func (r *Reader) Url(ctx context.Context) (string, error) {
	if url, ok := r.cache["url"]; ok {
		return url, nil
	}

	url, err := r.og.GetDownloadURL(ctx, r.ossID, r.ossKey, 0, true, r.opts...)

	if err != nil {
		return "", err
	}

	r.cache["url"] = url
	return url, nil
}

func (r *Reader) Text(ctx context.Context) (string, error) {
	if text, ok := r.cache["text"]; ok {
		return text, nil
	}

	data, err := r.og.DownloadFile(ctx, r.ossID, r.ossKey, true, r.opts...)

	if err != nil {
		return "", err
	}

	text := string(data)
	r.cache["text"] = text
	return text, nil
}

func (r *Reader) Json(ctx context.Context) (any, error) {
	var result any

	text, err := r.Text(ctx)

	if err != nil {
		return nil, err
	}

	if text == "" {
		return nil, nil
	}

	err = json.Unmarshal([]byte(text), &result)

	if err != nil {
		return nil, nil
	}

	return result, nil
}
