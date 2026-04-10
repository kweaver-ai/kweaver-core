// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package anyshare implements the AnyShare fileset connector.
package anyshare

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

const (
	authTypeToken       = 1
	authTypeAppSecret   = 2
	docLibTypeKnowledge = 1
	// docLibTypeDocument = 2 — list all entry doc libs when not filtering by knowledge type

	httpTimeout = 60 * time.Second

	PORT_MIN = 1
	// PORT_MAX 有效端口最大值
	PORT_MAX = 65535
)

type anyshareConfig struct {
	Protocol   string   `mapstructure:"protocol"`
	Host       string   `mapstructure:"host"`
	Port       int      `mapstructure:"port"`
	AuthType   int      `mapstructure:"auth_type"`
	Token      string   `mapstructure:"token"`
	AppID      string   `mapstructure:"app_id"`
	AppSecret  string   `mapstructure:"app_secret"`
	DocLibType int      `mapstructure:"doc_lib_type"`
	Paths      []string `mapstructure:"paths"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
}

type userInfoDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type docLibDTO struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Attr       int         `json:"attr"`
	Rev        string      `json:"rev"`
	CreatedAt  string      `json:"created_at"`
	ModifiedAt string      `json:"modified_at"`
	CreatedBy  userInfoDTO `json:"created_by"`
	ModifiedBy userInfoDTO `json:"modified_by"`
}

type docItemDetailDTO struct {
	ObjectId               string         `json:"object_id"`
	Name                   string         `json:"name"`
	Rev                    string         `json:"rev"`
	Size                   int64          `json:"size"`
	StorageName            string         `json:"storage_name"`
	SecurityClassification int            `json:"security_classification"`
	CreatedAt              string         `json:"created_at"`
	ModifiedAt             string         `json:"modified_at"`
	CreatedBy              userInfoDTO    `json:"created_by"`
	ModifiedBy             userInfoDTO    `json:"modified_by"`
	CustomMetadata         map[string]any `json:"custom_metadata"`
	DocLib                 docLibDTO      `json:"doc_lib"`
	Path                   string         `json:"path"`
	DocID                  string         `json:"doc_id"`
	Type                   string         `json:"type"`
}

type pathInfoDTO struct {
	DocID       string `json:"docid"`
	Name        string `json:"name"`
	Rev         string `json:"rev"`
	Size        int64  `json:"size"`
	ClientMTime int64  `json:"client_mtime"`
	Modified    int64  `json:"modified"`
}

// AnyShareConnector implements FilesetConnector for AnyShare 7.x document APIs.
type AnyShareConnector struct {
	enabled    bool
	config     *anyshareConfig
	connected  bool
	httpClient *http.Client
	baseURL    string
	authHeader string
}

// NewAnyShareConnector returns the builder for the anyshare connector type.
func NewAnyShareConnector() connectors.FilesetConnector {
	return &AnyShareConnector{}
}

// GetType returns the data source type.
func (c *AnyShareConnector) GetType() string {
	return "anyshare"
}

// GetName returns the connector name.
func (c *AnyShareConnector) GetName() string {
	return "anyshare"
}

// GetMode returns the connector mode.
func (c *AnyShareConnector) GetMode() string {
	return interfaces.ConnectorModeLocal
}

// GetCategory returns the connector category.
func (c *AnyShareConnector) GetCategory() string {
	return interfaces.ConnectorCategoryFileset
}

// GetEnabled returns the enabled status.
func (c *AnyShareConnector) GetEnabled() bool {
	return c.enabled
}

// SetEnabled sets the enabled status.
func (c *AnyShareConnector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// GetSensitiveFields returns sensitive field names.
func (c *AnyShareConnector) GetSensitiveFields() []string {
	return []string{"token", "app_secret"}
}

// GetFieldConfig returns connector form fields.
func (c *AnyShareConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return map[string]interfaces.ConnectorFieldConfig{
		"protocol":     {Name: "协议", Type: "string", Description: "http 或 https", Required: true, Encrypted: false},
		"host":         {Name: "主机地址", Type: "string", Description: "AnyShare 服务主机", Required: true, Encrypted: false},
		"port":         {Name: "端口", Type: "integer", Description: "服务端口", Required: true, Encrypted: false},
		"auth_type":    {Name: "认证方式", Type: "integer", Description: "1=访问令牌 Token，2=AppID/AppSecret", Required: true, Encrypted: false},
		"token":        {Name: "访问令牌", Type: "string", Description: "auth_type=1 时必填", Required: false, Encrypted: true},
		"app_id":       {Name: "应用账户 ID", Type: "string", Description: "auth_type=2 时必填", Required: false, Encrypted: false},
		"app_secret":   {Name: "应用密钥", Type: "string", Description: "auth_type=2 时必填", Required: false, Encrypted: true},
		"doc_lib_type": {Name: "文档库类型", Type: "integer", Description: "1=知识库", Required: true, Encrypted: false},
		"paths":        {Name: "路径列表", Type: "array", Description: "可选；按文档库名称路径解析起点，空则按文档库类型枚举", Required: false, Encrypted: false},
	}
}

// New creates a configured connector instance.
func (c *AnyShareConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	var ac anyshareConfig
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &ac,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, fmt.Errorf("decode anyshare config: %w", err)
	}
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode anyshare config: %w", err)
	}
	if ac.Protocol != "http" && ac.Protocol != "https" {
		return nil, fmt.Errorf("anyshare protocol must be http or https")
	}
	if ac.Host == "" || ac.Port <= 0 {
		return nil, fmt.Errorf("anyshare host and port are required")
	}
	// 验证端口号范围
	if ac.Port < PORT_MIN || ac.Port > PORT_MAX {
		return nil, fmt.Errorf("port %d is out of valid range (%d-%d)", ac.Port, PORT_MIN, PORT_MAX)
	}

	if ac.AuthType != authTypeToken && ac.AuthType != authTypeAppSecret {
		return nil, fmt.Errorf("anyshare auth_type must be 1 (token) or 2 (app credentials)")
	}
	switch ac.AuthType {
	case authTypeToken:
		if ac.Token == "" {
			return nil, fmt.Errorf("anyshare token is required when auth_type=1")
		}
	case authTypeAppSecret:
		if ac.AppID == "" || ac.AppSecret == "" {
			return nil, fmt.Errorf("anyshare app_id and app_secret are required when auth_type=2")
		}
	}
	if ac.DocLibType != docLibTypeKnowledge {
		return nil, fmt.Errorf("anyshare doc_lib_type must be 1 (knowledge)")
	}

	// 检查数组中是否存在重复元素
	seen := make(map[string]bool)
	for _, path := range ac.Paths {
		if seen[path] {
			return nil, fmt.Errorf("duplicate element found in 'paths': %s", path)
		}
		seen[path] = true
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	}
	nc := &AnyShareConnector{
		config: &ac,
		httpClient: &http.Client{
			Timeout:   httpTimeout,
			Transport: tr,
		},
		baseURL: fmt.Sprintf("%s://%s:%d", ac.Protocol, ac.Host, ac.Port),
	}
	return nc, nil
}

// Connect authenticates and marks the connector ready.
func (c *AnyShareConnector) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}
	if err := c.authenticate(ctx); err != nil {
		return err
	}
	c.connected = true
	return nil
}

func (c *AnyShareConnector) authenticate(ctx context.Context) error {
	switch c.config.AuthType {
	case authTypeToken:
		c.authHeader = normalizeBearer(c.config.Token)
	case authTypeAppSecret:
		token, err := c.fetchOAuthToken(ctx)
		if err != nil {
			return err
		}
		c.authHeader = normalizeBearer(token)
	default:
		return fmt.Errorf("unsupported auth_type %d", c.config.AuthType)
	}
	return nil
}

func normalizeBearer(raw string) string {
	if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return raw
	}
	return "Bearer " + raw
}

func (c *AnyShareConnector) fetchOAuthToken(ctx context.Context) (string, error) {
	u := fmt.Sprintf("%s/oauth2/token", c.baseURL)
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "all")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.config.AppID, c.config.AppSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth2 token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("oauth2 token decode: %w", err)
	}
	if tr.Error != "" || tr.AccessToken == "" {
		return "", fmt.Errorf("oauth2 token error: %s (status=%d body=%s)", tr.Error, resp.StatusCode, truncateForLog(body))
	}
	return tr.AccessToken, nil
}

func truncateForLog(b []byte) string {
	const max = 512
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}

// Close releases runtime state.
func (c *AnyShareConnector) Close(ctx context.Context) error {
	c.connected = false
	c.authHeader = ""
	return nil
}

// Ping checks connectivity with a lightweight API call.
func (c *AnyShareConnector) Ping(ctx context.Context) error {
	return c.TestConnection(ctx)
}

// TestConnection verifies credentials against AnyShare.
func (c *AnyShareConnector) TestConnection(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	// 如果配置了 paths，遍历检查每个路径是否能获取到 docid
	if len(c.config.Paths) > 0 {
		for _, p := range c.config.Paths {
			docInfo, err := c.getDocIDByPath(ctx, p)
			if err != nil {
				return fmt.Errorf("failed to validate path %q: %w", p, err)
			}
			// 检查路径是否指向文件（size != -1 表示文件）
			if docInfo.Size != -1 {
				return fmt.Errorf("path %q must be a directory, not a file", p)
			}
		}
		return nil
	}
	// 没有配置 paths 时，通过 getEntryDocLib 判断
	_, err := c.getEntryDocLib(ctx)
	return err
}

// GetMetadata returns basic catalog metadata.
func (c *AnyShareConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return map[string]any{
		"connector":        "anyshare",
		"host":             c.config.Host,
		"port":             c.config.Port,
		"doc_lib_type":     c.config.DocLibType,
		"protocol":         c.config.Protocol,
		"paths_configured": len(c.config.Paths) > 0,
	}, nil
}

// buildFilesetMeta 创建 FilesetMeta 对象的辅助函数
func buildFilesetMeta(id, name, displayPath, itemType, createdAt, modifiedAt string, createdBy, modifiedBy userInfoDTO, rev string) *interfaces.FilesetMeta {
	return &interfaces.FilesetMeta{
		ID:          id,
		Name:        name,
		DisplayPath: displayPath,
		SourceMetadata: map[string]any{
			"id":          id,
			"name":        name,
			"type":        itemType,
			"created_at":  createdAt,
			"modified_at": modifiedAt,
			"created_by":  createdBy,
			"modified_by": modifiedBy,
			"rev":         rev,
		},
	}
}

// ListFilesets discovers one level of files/folders per configured roots (see design doc).
func (c *AnyShareConnector) ListFilesets(ctx context.Context) ([]*interfaces.FilesetMeta, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	var out []*interfaces.FilesetMeta

	if len(c.config.Paths) == 0 {
		// 获取知识库列表（仅第一层）
		libs, err := c.getEntryDocLib(ctx)
		if err != nil {
			return nil, err
		}

		// 将知识库信息转换为FilesetMeta返回
		for _, lib := range libs {
			meta := buildFilesetMeta(lib.ID, lib.Name, lib.Name, lib.Type, lib.CreatedAt, lib.ModifiedAt, lib.CreatedBy, lib.ModifiedBy, lib.Rev)
			out = append(out, meta)
		}
		return out, nil
	}

	// 遍历每个路径
	for _, p := range c.config.Paths {

		// 根据路径获取doc_id
		docInfo, err := c.getDocIDByPath(ctx, p)
		if err != nil {
			return nil, err
		}

		// 检查路径是否指向文件（size != -1 表示文件）
		if docInfo.Size != -1 {
			return nil, fmt.Errorf("path %q must be a directory, not a file", p)
		}

		// 获取文件夹详细信息
		detail, err := c.getDocItemDetail(ctx, docInfo.DocID)
		if err != nil {
			return nil, err
		}

		// 将文件夹本身作为一个 resource
		meta := buildFilesetMeta(detail.DocID, detail.Name, detail.Path, detail.Type, detail.CreatedAt, detail.ModifiedAt, detail.CreatedBy, detail.ModifiedBy, detail.Rev)
		out = append(out, meta)
	}

	return out, nil
}

// SearchFiles searches files based on query parameters.
func (c *AnyShareConnector) SearchFiles(ctx context.Context, docID, keyword string, rows, start int) ([]map[string]any, int64, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, 0, err
	}
	if docID == "" {
		return nil, 0, fmt.Errorf("empty doc id")
	}
	u := fmt.Sprintf("%s/api/ecosearch/v1/file-search", c.baseURL)
	payload := map[string]interface{}{
		"keyword":      keyword,
		"model":        "phrase",
		"range":        []string{fmt.Sprintf("%s/*", docID)},
		"rows":         rows,
		"start":        start,
		"type":         "doc",
		"quick_search": true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf("file-search http %d: %s", resp.StatusCode, truncateForLog(raw))
	}

	var result struct {
		Files []map[string]any `json:"files"`
		Hits  int64            `json:"hits"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, 0, fmt.Errorf("file-search decode: %w", err)
	}

	return result.Files, result.Hits, nil
}

// GetObjectDownloadInfo returns download info for a given file or folder.
func (c *AnyShareConnector) GetObjectDownloadInfo(ctx context.Context, resourceName, docID string) (map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	if docID == "" {
		return nil, fmt.Errorf("empty doc id")
	}
	u := fmt.Sprintf("%s/api/open-doc/v1/file-download", c.baseURL)
	// 当前 docID 为 gns://... 的形式，name为下载文件时的打包名称
	payload := map[string]interface{}{
		"doc": []map[string]string{
			{"id": docID},
		},
		"name": resourceName,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("file-download http %d: %s", resp.StatusCode, truncateForLog(raw))
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("file-download decode: %w", err)
	}
	if msg, ok := m["message"].(string); ok && msg != "" {
		return nil, fmt.Errorf("file-download: %s", msg)
	}
	return m, nil
}

func (c *AnyShareConnector) getEntryDocLib(ctx context.Context) ([]docLibDTO, error) {
	q := url.Values{}
	q.Set("sort", "doc_lib_name")
	q.Set("direction", "asc")
	if c.config.DocLibType == docLibTypeKnowledge {
		q.Set("type", "knowledge_doc_lib")
	}
	u := fmt.Sprintf("%s/api/document/v1/entry-doc-lib?%s", c.baseURL, q.Encode())
	var libs []docLibDTO
	if err := c.getJSON(ctx, u, &libs); err != nil {
		return nil, err
	}
	return libs, nil
}

func (c *AnyShareConnector) getDocIDByPath(ctx context.Context, namepath string) (pathInfoDTO, error) {
	u := fmt.Sprintf("%s/api/efast/v1/file/getinfobypath", c.baseURL)
	body, err := json.Marshal(map[string]string{"namepath": namepath})

	var info pathInfoDTO
	if err != nil {
		return info, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return info, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return info, fmt.Errorf("getinfobypath http %d: %s", resp.StatusCode, truncateForLog(raw))
	}
	if err := json.Unmarshal(raw, &info); err != nil {
		return info, err
	}
	if info.DocID == "" {
		return info, fmt.Errorf("getinfobypath: empty docid for path %q", namepath)
	}
	return info, nil
}

func (c *AnyShareConnector) getDocItemDetail(ctx context.Context, docID string) (*docItemDetailDTO, error) {
	// docID格式为 gns://.../.../object_id，需要提取object_id
	parts := strings.Split(docID, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid docID format: %s", docID)
	}
	objectID := parts[len(parts)-1]

	u := fmt.Sprintf("%s/api/efast/v2/items/%s/all", c.baseURL, objectID)

	var detail docItemDetailDTO
	if err := c.getJSON(ctx, u, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (c *AnyShareConnector) getJSON(ctx context.Context, reqURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.authHeader)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GET %s: http %d %s", reqURL, resp.StatusCode, truncateForLog(raw))
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
