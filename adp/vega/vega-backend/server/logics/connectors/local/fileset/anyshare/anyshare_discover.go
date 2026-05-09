// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package anyshare implements the AnyShare fileset connector.
package anyshare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"vega-backend/interfaces"
)

// ListFilesets discovers one level of files/folders per configured roots (see design doc).
func (c *AnyShareConnector) ListFilesets(ctx context.Context) ([]*interfaces.FilesetMeta, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	var out []*interfaces.FilesetMeta

	if len(c.config.Paths) == 0 {
		// 获取文档库列表（仅第一层）
		libs, err := c.getEntryDocLib(ctx)
		if err != nil {
			return nil, err
		}

		// 将知识库信息转换为FilesetMeta返回
		for _, lib := range libs {
			meta := buildFilesetMeta(lib.ID, lib.Name, lib.Name, lib.Type,
				lib.CreatedAt, lib.ModifiedAt, lib.CreatedBy, lib.ModifiedBy, lib.Rev)
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
		detail, err := c.getDocItemDetail(ctx, docInfo.DocID, "all")
		if err != nil {
			return nil, err
		}

		// 将文件夹本身作为一个 resource
		meta := buildFilesetMeta(detail.DocID, detail.Name, detail.Path, detail.Type,
			detail.CreatedAt, detail.ModifiedAt, detail.CreatedBy, detail.ModifiedBy, detail.Rev)
		out = append(out, meta)
	}

	return out, nil
}

// buildFilesetMeta 创建 FilesetMeta 对象的辅助函数
func buildFilesetMeta(id, name, displayPath, itemType, createdAt, modifiedAt string,
	createdBy, modifiedBy userInfoDTO, rev string) *interfaces.FilesetMeta {
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
		Columns: []interfaces.FilesetColumnMeta{
			{Name: "doc_id", Type: "string", OrigType: "string"},
			{Name: "basename", Type: "string", OrigType: "string"},
			{Name: "source", Type: "string", OrigType: "string"},
			{Name: "parent_path", Type: "string", OrigType: "string"},
			{Name: "doc_lib_type", Type: "string", OrigType: "string"},
			{Name: "extension", Type: "string", OrigType: "string"},
			{Name: "summary", Type: "string", OrigType: "string"},
			{Name: "created_by", Type: "string", OrigType: "string"},
			{Name: "modified_by", Type: "string", OrigType: "string"},
			{Name: "created_at", Type: "integer", OrigType: "integer"},
			{Name: "modified_at", Type: "integer", OrigType: "integer"},
			{Name: "size", Type: "integer", OrigType: "integer"},
			{Name: "security_level", Type: "integer", OrigType: "integer"},
			{Name: "doc_type", Type: "string", OrigType: "string"},
			{Name: "content", Type: "string", OrigType: "string"},
			{Name: "tags", Type: "string", OrigType: "string"},
			{Name: "title", Type: "string", OrigType: "string"},
			{Name: "embedded_image", Type: "string", OrigType: "string"},
			{Name: "only_display", Type: "boolean", OrigType: "boolean"},
			{Name: "score", Type: "integer", OrigType: "integer"},
		},
	}
}

func (c *AnyShareConnector) getEntryDocLib(ctx context.Context) ([]entryDocLibDTO, error) {
	q := url.Values{}
	q.Set("sort", "doc_lib_name")
	q.Set("direction", "asc")
	switch c.config.DocLibType {
	case docLibTypeKnowledge:
		q.Set("type", "knowledge_doc_lib")
	case docLibTypeDocument:
		q.Set("type", "custom_doc_lib")
		q.Set("subtype_id", customDocLibSubTypeDocumentId)
	}
	u := fmt.Sprintf("%s/api/document/v1/entry-doc-lib?%s", c.baseURL, q.Encode())
	var libs []entryDocLibDTO
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
	defer func() { _ = resp.Body.Close() }()
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

func (c *AnyShareConnector) getDocItemDetail(ctx context.Context, docID string, field string) (*docItemDetailDTO, error) {
	// docID格式为 gns://.../.../object_id，需要提取object_id
	parts := strings.Split(docID, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid docID format: %s", docID)
	}
	objectID := parts[len(parts)-1]

	u := fmt.Sprintf("%s/api/efast/v2/items/%s/%s", c.baseURL, objectID, field)

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
	defer func() { _ = resp.Body.Close() }()
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
