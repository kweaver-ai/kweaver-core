package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/xeipuuv/gojsonschema"
)

// BindAndValid binds and validates data
// func BindAndValid(c *gin.Context, form interface{}) error {
// 	err := c.ShouldBindJSON(form)
// 	return err
// }

// BindAndValid binds and validates data
func BindAndValid(r io.Reader, schemaPath string, out interface{}) error {
	data, _ := io.ReadAll(r)

	err := JSONSchemaValid(data, schemaPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &out)
	if err != nil {
		return errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()})
	}
	return nil
}

// JSONSchemaValid params schema validate
func JSONSchemaValid(data []byte, path string) error {
	var filePath = fmt.Sprintf("file://%s", path)
	_, err := os.Stat(path)
	if err != nil {
		apath, _ := os.Getwd()
		filePath = fmt.Sprintf("file://%s/schema/%s", apath, path)
	}

	schemaLoader := gojsonschema.NewReferenceLoader(filePath)
	documentLoader := gojsonschema.NewStringLoader(string(data))
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"validate": err.Error()})
	}

	if !result.Valid() {
		detail := make([]interface{}, len(result.Errors()))
		for i, desc := range result.Errors() {
			detail[i] = desc.String()
		}
		return errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": detail})
	}

	return nil
}

// JSONSchemaValidV2 适配新错误码规范
func JSONSchemaValidV2(ctx context.Context, data []byte, path string) error {
	var filePath = fmt.Sprintf("file://%s", path)
	_, err := os.Stat(path)
	if err != nil {
		apath, _ := os.Getwd()
		filePath = fmt.Sprintf("file://%s/schema/%s", apath, path)
	}

	schemaLoader := gojsonschema.NewReferenceLoader(filePath)
	documentLoader := gojsonschema.NewStringLoader(string(data))
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"validate": err.Error()})
	}

	if !result.Valid() {
		detail := make([]interface{}, len(result.Errors()))
		for i, desc := range result.Errors() {
			detail[i] = desc.String()
		}
		return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"params": detail})
	}

	return nil
}

// ParseQuery 解析查询参数
func ParseQuery(c *gin.Context) (map[string]interface{}, error) {
	var (
		page  = c.DefaultQuery("page", "0")
		limit = c.DefaultQuery("limit", "20")
	)
	pageInt, err := strconv.ParseInt(page, 0, 64)

	if err != nil {
		return nil, ierr.NewPublicRestError(c.Request.Context(),
			ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"params": []string{"page: Invalid type, expected integer"}})
	}

	limitInt, err := strconv.ParseInt(limit, 0, 64)

	if err != nil {
		return nil, ierr.NewPublicRestError(c.Request.Context(),
			ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"params": []string{"limit: Invalid type, expected integer"}})
	}

	var query = map[string]interface{}{
		"page":  pageInt,
		"limit": limitInt,
	}

	queryParams := c.Request.URL.Query()
	for key, values := range queryParams {
		if key != "page" && key != "limit" {
			if len(values) > 0 {
				query[key] = values[0]
			}
		}
	}

	return query, nil
}
