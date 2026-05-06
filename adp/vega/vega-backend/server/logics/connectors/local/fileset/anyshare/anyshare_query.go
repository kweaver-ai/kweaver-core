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
	"vega-backend/interfaces"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

// ExecuteQuery executes a query on the fileset.
func (c *AnyShareConnector) ExecuteQuery(ctx context.Context, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {
	// Extract docID from resource
	docID := ""
	if resource.SourceMetadata != nil {
		if id, ok := resource.SourceMetadata["id"].(string); ok && id != "" {
			docID = id
		}
	}
	if docID == "" {
		docID = resource.SourceIdentifier
	}

	// Extract keyword, dimension and custom from filter condition
	// Note: dimension, custom and condition are AND related
	// - dimension: allowed keywords are [basename, content, summary]
	// - custom: allowed keywords are [created_at, modified_at, tags]
	// - condition: allowed keywords are [extension, created_by, modified_by]
	keyword := ""
	dimension := []string{}               // default dimension
	custom := []map[string]interface{}{}  // custom conditions for file-search
	condition := map[string]interface{}{} // condition for file-search
	model := ""
	// Validate filter condition and extract keyword, custom, dimension, condition and model
	if params.ActualFilterCond != nil {
		filterResult, err := c.ConvertFilterCondition(ctx, params.ActualFilterCond)
		if err != nil {
			return nil, fmt.Errorf("failed to process filter condition: %w", err)
		}
		keyword = filterResult.Keyword
		custom = filterResult.Custom
		dimension = filterResult.Dimension
		condition = filterResult.Condition
		model = filterResult.Model
	}

	// Process sort parameters
	sortParams, err := processSortParams(params.Sort)
	if err != nil {
		return nil, fmt.Errorf("failed to process sort params: %w", err)
	}

	// Validate output fields
	if len(params.OutputFields) > 0 {
		if err := validateOutputFields(params.OutputFields); err != nil {
			return nil, fmt.Errorf("failed to validate output fields: %w", err)
		}
	}

	// Call SearchFiles to get results
	files, err := c.SearchFiles(ctx, docID, keyword, dimension, model, custom, condition, params.Limit, params.Offset, sortParams, params.OutputFields)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)

	}

	return &interfaces.QueryResult{
		Rows: files,
	}, nil
}

// SearchFiles searches files based on query parameters.
func (c *AnyShareConnector) SearchFiles(ctx context.Context, docID, keyword string, dimension []string, model string, custom []map[string]interface{}, condition map[string]interface{}, rows, start int, sort map[string]interface{}, outputFields []string) ([]map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	if docID == "" {
		return nil, fmt.Errorf("empty doc id")
	}
	u := fmt.Sprintf("%s/api/ecosearch/v1/file-search", c.baseURL)
	payload := map[string]interface{}{
		"keyword":      keyword,
		"range":        []string{fmt.Sprintf("%s/*", docID)},
		"rows":         rows,
		"start":        start,
		"type":         "doc",
		"quick_search": true,
	}

	// Add model if model is specified
	if model != "" {
		payload["model"] = model
	}

	// Add dimension if dimension is specified
	if len(dimension) > 0 {
		payload["dimension"] = dimension
	}
	// Add condition if condition is specified
	if len(condition) > 0 {
		payload["condition"] = condition
	}
	// Add custom if custom conditions are specified
	if len(custom) > 0 {
		payload["custom"] = custom
	}
	// Add sort if sort parameters are specified
	if len(sort) > 0 {
		payload["sort"] = sort
	}

	logger.Debugf("AnyShare file search params: %v", payload)

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
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("file-search http %d: %s", resp.StatusCode, truncateForLog(raw))
	}

	var result struct {
		Files []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("file-search decode: %w", err)
	}

	// Process output fields if specified
	if len(outputFields) > 0 {
		result.Files = filterOutputFields(result.Files, outputFields)
	}

	return result.Files, nil
}

// processSortParams processes sort parameters and converts them to the format required by /file-search API.
// The sort parameter is an object with two properties:
// - field: allowed values are "created_at", "modified_at"
// - sort_type: allowed values are "asc" and "desc"
// Note: Only the first sort field is used, as the API supports single field sorting
func processSortParams(sort []*interfaces.SortField) (map[string]interface{}, error) {
	if len(sort) == 0 {
		return nil, nil
	}

	// Define allowed fields for sorting
	allowedFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	// Define allowed sort types
	allowedSortTypes := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	// Check if more than one sort field is specified
	if len(sort) > 1 {
		return nil, fmt.Errorf("only one sort field is allowed")
	}

	// Use the first sort field
	s := sort[0]
	if s == nil {
		return nil, nil
	}

	// Validate field
	if !allowedFields[s.Field] {
		return nil, fmt.Errorf("invalid sort field: %s, allowed fields are: created_at, modified_at", s.Field)
	}

	// Validate sort type
	sortType := "asc" // default sort type
	if s.Direction != "" {
		if !allowedSortTypes[s.Direction] {
			return nil, fmt.Errorf("invalid sort type: %s, allowed types are: asc, desc", s.Direction)
		}
		sortType = s.Direction
	}

	// Return sort object
	return map[string]interface{}{
		"field":     s.Field,
		"sort_type": sortType,
	}, nil
}

// validateOutputFields validates the output fields.
// If a field in outputFields is not in the allowed fields, it will return an error.
func validateOutputFields(outputFields []string) error {
	// Define allowed fields for output
	allowedFields := map[string]bool{
		"*":              true,
		"doc_id":         true,
		"basename":       true,
		"source":         true,
		"parent_path":    true,
		"doc_lib_type":   true,
		"extension":      true,
		"summary":        true,
		"created_by":     true,
		"modified_by":    true,
		"created_at":     true,
		"modified_at":    true,
		"size":           true,
		"security_level": true,
		"doc_type":       true,
		"content":        true,
		"tags":           true,
		"title":          true,
		"embedded_image": true,
		"only_display":   true,
		"score":          true,
	}

	// Validate all output fields
	for _, field := range outputFields {
		if !allowedFields[field] {
			return fmt.Errorf("invalid output field: %s, allowed fields are: *, doc_id, basename, source, parent_path, doc_lib_type, extension, summary, created_by, modified_by, created_at, modified_at, size, security_level, doc_type, content, tags, title, embedded_image, only_display, score", field)
		}
	}
	return nil
}

// filterOutputFields filters the output fields based on the specified fields.
// Only returns the fields that are in the outputFields list.
func filterOutputFields(files []map[string]any, outputFields []string) []map[string]any {
	if len(files) == 0 || len(outputFields) == 0 {
		return files
	}

	// Check if outputFields contains "*", if so, return all fields without filtering
	for _, field := range outputFields {
		if field == "*" {
			return files
		}
	}

	// Filter each file to only include the specified output fields
	filteredFiles := make([]map[string]any, len(files))
	for i, file := range files {
		filteredFile := make(map[string]any)
		for _, field := range outputFields {
			if value, ok := file[field]; ok {
				filteredFile[field] = value
			}
		}
		filteredFiles[i] = filteredFile
	}

	return filteredFiles
}
