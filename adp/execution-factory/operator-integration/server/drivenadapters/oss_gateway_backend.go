package drivenadapters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// OSSGatewayBackendClient OSS 网关后端客户端
type ossGatewayBackendClient struct {
	httpClient      interfaces.HTTPClient
	client          *http.Client
	baseURL         string
	storageID       string
	internalRequest bool
	expires         int64
	storageMu       sync.RWMutex
	refreshDefault  bool
	refreshOnce     sync.Once
	stopCh          chan struct{}
	logger          interfaces.Logger
	isReady         bool
}

type gatewayAuthResponse struct {
	Data gatewayAuthRequest `json:"data"`
}

type gatewayAuthRequest struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	FormField map[string]string `json:"form_field"`
	Body      string            `json:"body"`
}

type storageListResponse struct {
	Count int64         `json:"count"`
	Data  []storageItem `json:"data"`
}

type storageItem struct {
	StorageID string `json:"storage_id"`
	IsDefault bool   `json:"is_default"`
	IsEnabled bool   `json:"is_enabled"`
}

const ossGatewayStorageRefreshInterval = 5 * time.Minute

var (
	ossOnce   = sync.Once{}
	ossClient interfaces.OSSGatewayBackendClient
)

// NewOSSGatewayBackendClient 创建 OSS 网关后端客户端
func NewOSSGatewayBackendClient() interfaces.OSSGatewayBackendClient {
	ossOnce.Do(func() {
		cfg := config.NewConfigLoader().OSSGatewayBackendConfig
		expires := cfg.Expires
		if expires <= 0 {
			expires = 3600 // 单位（秒）
		}
		client := &ossGatewayBackendClient{
			httpClient:      rest.NewHTTPClient(),
			client:          rest.NewRawHTTPClient(),
			baseURL:         fmt.Sprintf("%s://%s:%d/api/v1", cfg.PrivateProtocol, cfg.PrivateHost, cfg.PrivatePort),
			storageID:       cfg.StorageID,
			internalRequest: cfg.InternalRequest,
			expires:         expires,
			refreshDefault:  cfg.StorageID == "",
			stopCh:          make(chan struct{}),
			logger:          config.NewConfigLoader().GetLogger(),
		}
		if err := client.initStorageID(context.Background()); err == nil {
			client.startStorageRefresh()
			client.isReady = true
		} else {
			client.isReady = false
			client.logger.Errorf("init storage id failed, baseURL: %s, err: %v", client.baseURL, err)
		}
		ossClient = client
	})
	return ossClient
}

// IsReady 检查 OSS 网关后端客户端是否就绪
func (c *ossGatewayBackendClient) IsReady() bool {
	return c.isReady
}

func (c *ossGatewayBackendClient) Close() error {
	close(c.stopCh)
	return nil
}

// UploadFile 单个文件上传
func (c *ossGatewayBackendClient) UploadFile(ctx context.Context, object *interfaces.OssObject, content []byte) (err error) {
	src := fmt.Sprintf("%s/upload/%s/%s", c.baseURL, object.StorageID, url.PathEscape(object.StorageKey))
	query := url.Values{
		"request_method":   []string{http.MethodPut},
		"expires":          []string{fmt.Sprintf("%d", c.expires)},
		"internal_request": []string{fmt.Sprintf("%t", c.internalRequest)},
	}
	headers := common.GetHeaderFromCtx(ctx)
	respCode, respBody, err := c.httpClient.GetNoUnmarshal(ctx, src, query, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("upload file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("upload file failed, respCode: %d, respData: %v", respCode, respBody))
		return err
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("upload file failed, respCode: %d, respData: %v", respCode, respBody)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("upload file failed, respCode: %d, respData: %v", respCode, respBody))
		return err
	}
	authResp := &gatewayAuthResponse{}
	err = utils.StringToObject(string(respBody), authResp)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("upload file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("upload file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err))
		return err
	}
	if authResp.Data.Method == "" || authResp.Data.URL == "" {
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOSSGatewayFailed,
			fmt.Sprintf("upload file failed, invalid oss gateway response: method=%s url=%s", authResp.Data.Method, authResp.Data.URL))
		return
	}
	return c.doSignedRequest(ctx, authResp.Data.Method, authResp.Data.URL, authResp.Data.Headers, bytes.NewReader(content))
}

// DownloadFile 下载文件
func (c *ossGatewayBackendClient) DownloadFile(ctx context.Context, object *interfaces.OssObject) (data []byte, err error) {
	authReq, err := c.getDownloadInfo(ctx, object)
	if err != nil {
		return nil, err
	}
	resp, err := c.doSignedRequestRaw(ctx, authReq.Method, authReq.URL, authReq.Headers, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return
}

// DeleteFile 删除文件
func (c *ossGatewayBackendClient) DeleteFile(ctx context.Context, object *interfaces.OssObject) (err error) {
	src := fmt.Sprintf("%s/delete/%s/%s", c.baseURL, object.StorageID, url.PathEscape(object.StorageKey))
	headers := common.GetHeaderFromCtx(ctx)
	query := url.Values{
		"expires":          []string{fmt.Sprintf("%d", c.expires)},
		"internal_request": []string{fmt.Sprintf("%t", c.internalRequest)},
	}
	respCode, respBody, err := c.httpClient.GetNoUnmarshal(ctx, src, query, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("delete file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("delete file failed, respCode: %d, respData: %v", respCode, respBody))
		return
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("delete file failed, respCode: %d, respData: %v", respCode, respBody)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("delete file failed, respCode: %d, respData: %v", respCode, respBody))
		return
	}
	authResp := &gatewayAuthResponse{}
	err = utils.StringToObject(string(respBody), authResp)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("delete file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("delete file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err))
		return
	}
	if authResp.Data.Method == "" || authResp.Data.URL == "" {
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOSSGatewayFailed,
			fmt.Sprintf("delete file failed, invalid oss gateway response: method=%s url=%s", authResp.Data.Method, authResp.Data.URL))
		return
	}
	return c.doSignedRequest(ctx, authResp.Data.Method, authResp.Data.URL, authResp.Data.Headers, nil)
}

func (c *ossGatewayBackendClient) getDownloadInfo(ctx context.Context, object *interfaces.OssObject) (resp *gatewayAuthRequest, err error) {
	src := fmt.Sprintf("%s/download/%s/%s", c.baseURL, object.StorageID, url.PathEscape(object.StorageKey))
	headers := common.GetHeaderFromCtx(ctx)
	query := url.Values{
		"expires":          []string{fmt.Sprintf("%d", c.expires)},
		"internal_request": []string{fmt.Sprintf("%t", c.internalRequest)},
	}
	respCode, respBody, err := c.httpClient.GetNoUnmarshal(ctx, src, query, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("download file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("download file failed, respCode: %d, respData: %v", respCode, respBody))
		return
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("download file failed, respCode: %d, respData: %v", respCode, respBody)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("download file failed, respCode: %d, respData: %v", respCode, respBody))
		return
	}
	authResp := &gatewayAuthResponse{}
	err = utils.StringToObject(string(respBody), authResp)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("download file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err)
		err = errors.NewHTTPError(ctx, respCode, errors.ErrExtOSSGatewayFailed, fmt.Sprintf("download file failed, respCode: %d, respData: %v, err: %v", respCode, respBody, err))
		return
	}
	if authResp.Data.Method == "" || authResp.Data.URL == "" {
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOSSGatewayFailed,
			fmt.Sprintf("download file failed, invalid oss gateway response: method=%s url=%s", authResp.Data.Method, authResp.Data.URL))
		return
	}
	resp = &gatewayAuthRequest{
		Method:  authResp.Data.Method,
		URL:     authResp.Data.URL,
		Headers: authResp.Data.Headers,
	}
	return
}

// GetDownloadURL 获取文件下载URL
func (c *ossGatewayBackendClient) GetDownloadURL(ctx context.Context, object *interfaces.OssObject) (url string, err error) {
	resp, err := c.getDownloadInfo(ctx, object)
	if err != nil {
		return
	}
	return resp.URL, nil
}

func (c *ossGatewayBackendClient) CurrentStorageID(ctx context.Context) (string, error) {
	return c.currentStorageID(ctx)
}

func (c *ossGatewayBackendClient) initStorageID(ctx context.Context) error {
	c.storageMu.RLock()
	storageID := c.storageID
	c.storageMu.RUnlock()
	if storageID != "" {
		return nil
	}
	storageID, err := c.GetDefaultStorageID(ctx)
	if err != nil {
		return err
	}
	c.storageMu.Lock()
	c.storageID = storageID
	c.storageMu.Unlock()
	return nil
}

func (c *ossGatewayBackendClient) currentStorageID(ctx context.Context) (string, error) {
	c.storageMu.RLock()
	defer c.storageMu.RUnlock()
	if c.storageID == "" {
		return "", errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "oss gateway storage_id is empty")
	}
	return c.storageID, nil
}

// GetDefaultStorageID 获取默认存储ID
func (c *ossGatewayBackendClient) GetDefaultStorageID(ctx context.Context) (string, error) {
	src := fmt.Sprintf("%s/storages?enabled=true&is_default=true", c.baseURL)
	headers := common.GetHeaderFromCtx(ctx)
	respCode, respBody, err := c.httpClient.GetNoUnmarshal(ctx, src, nil, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("GetDefaultStorageID failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOSSGatewayFailed, map[string]any{
			"error": err.Error(),
		})
		return "", err
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("GetDefaultStorageID failed, unexpected status code: %d, response: %s", respCode, string(respBody))
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOSSGatewayFailed, map[string]any{
			"error":     fmt.Sprintf("unexpected status code: %d", respCode),
			"response":  string(respBody),
			"http_code": respCode,
		})
		return "", err
	}
	payload := &storageListResponse{}
	err = utils.StringToObject(string(respBody), payload)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("GetDefaultStorageID failed, StringToObject failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOSSGatewayFailed, map[string]any{
			"error": err.Error(),
		})
		return "", err
	}
	if len(payload.Data) == 0 {
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOSSGatewayDefaultStorageNotFound,
			fmt.Sprintf("default oss gateway storage not found, response: %s", string(respBody)))
		return "", err
	}
	for _, item := range payload.Data {
		if item.IsEnabled && item.IsDefault {
			return item.StorageID, nil
		}
	}
	return "", errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("default oss gateway storage not found, response: %s", string(respBody)))
}

func (c *ossGatewayBackendClient) startStorageRefresh() {
	if !c.refreshDefault || ossGatewayStorageRefreshInterval <= 0 {
		return
	}
	c.refreshOnce.Do(func() {
		ticker := time.NewTicker(ossGatewayStorageRefreshInterval)
		c.logger.Infof("start oss gateway storage refresh, interval=%s", ossGatewayStorageRefreshInterval.String())
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					storageID, err := c.GetDefaultStorageID(context.Background())
					if err != nil {
						continue
					}
					c.storageMu.Lock()
					c.storageID = storageID
					c.storageMu.Unlock()
					c.logger.Debugf("oss gateway storage refresh success, storage_id=%s", storageID)
				case <-c.stopCh:
					c.logger.Infof("stop oss gateway storage refresh")
					return
				}
			}
		}()
	})
}

func (c *ossGatewayBackendClient) doSignedRequest(ctx context.Context, method, reqURL string, headers map[string]string, body io.Reader) error {
	resp, err := c.doSignedRequestRaw(ctx, method, reqURL, headers, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// doSignedRequestRaw 发送已签名的HTTP请求
func (c *ossGatewayBackendClient) doSignedRequestRaw(ctx context.Context, method, reqURL string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer func() { _ = resp.Body.Close() }()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("signed request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody))))
	}
	return resp, nil
}
