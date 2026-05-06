// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package remote provides HTTP-based remote connector implementations.
package remote

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

const (
	defaultDialTimeout    = 5 * time.Second
	defaultRequestTimeout = 30 * time.Second
	maxIdleConns          = 10
	idleConnTimeout       = 5 * time.Minute
)

// Client HTTP 客户端封装，用于与远程 connector 服务通信
type Client struct {
	httpClient *http.Client
}

// NewClient 创建新的 HTTP 客户端
func NewClient() *Client {
	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConns,
		IdleConnTimeout:     idleConnTimeout,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   defaultRequestTimeout,
		},
	}
}

// Request 发送 HTTP 请求
func (c *Client) Request(ctx context.Context, method, url string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := sonic.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Errorf("Remote connector request failed: status=%d, body=%s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get 发送 GET 请求
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, url, nil)
}

// Post 发送 POST 请求
func (c *Client) Post(ctx context.Context, url string, body any) ([]byte, error) {
	return c.Request(ctx, http.MethodPost, url, body)
}

// Delete 发送 DELETE 请求
func (c *Client) Delete(ctx context.Context, url string) ([]byte, error) {
	return c.Request(ctx, http.MethodDelete, url, nil)
}

// ============================================
// Remote Connector API 请求/响应结构
// ============================================

// CreateConnectionRequest 创建连接请求
type CreateConnectionRequest struct {
	Type     string         `json:"type"`
	Host     string         `json:"host"`
	Port     int            `json:"port"`
	Database string         `json:"database"`
	Username string         `json:"username"`
	Password string         `json:"password"`
	Options  map[string]any `json:"options,omitempty"`
}

// CreateConnectionResponse 创建连接响应
type CreateConnectionResponse struct {
	ConnectionID string `json:"connection_id"`
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
}

// ColumnMeta 列元数据
type ColumnMeta struct {
	Name        string `json:"name"`
	NativeType  string `json:"native_type"`
	VegaType    string `json:"vega_type"`
	Nullable    bool   `json:"nullable"`
	Description string `json:"description"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	Query string `json:"query"`
	Args  []any  `json:"args,omitempty"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Total   int64            `json:"total"`
}

// IndexMetaResponse 索引元数据响应
type IndexMetaResponse struct {
	Name      string               `json:"name"`
	DocCount  int64                `json:"doc_count"`
	StoreSize string               `json:"store_size"`
	Mapping   map[string]FieldMeta `json:"mapping"`
	Settings  map[string]any       `json:"settings"`
}

// FieldMeta 字段元数据
type FieldMeta struct {
	Type       string `json:"type"`
	Analyzer   string `json:"analyzer,omitempty"`
	Searchable bool   `json:"searchable"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query map[string]any `json:"query"`
	From  int            `json:"from"`
	Size  int            `json:"size"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Hits  []map[string]any `json:"hits"`
	Total int64            `json:"total"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ConnectorInfoResponse connector 信息响应
type ConnectorInfoResponse struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}
