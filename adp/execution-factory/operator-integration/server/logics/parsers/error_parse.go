package parsers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
)

const (
	errorTypeLoadFailed       = "OpenAPILoadFailed"       // OpenAPI规范加载失败
	errorTypeValidationFailed = "OpenAPIValidationFailed" // OpenAPI规范验证失败
	elementTypeLen            = 2
)

var (
	parameterRegex = regexp.MustCompile(`parameter\s*["']([^"']+)["']`)
	responseRegex  = regexp.MustCompile(`response\s*["']([^"']+)["']`)
	schemaRegex    = regexp.MustCompile(`schema\s*["']([^"']+)["']`)
	operationRegex = regexp.MustCompile(`operation\s*["']([^"']+)["']`)
	fieldRegex     = regexp.MustCompile(`field\s*["']([^"']+)["']`)
)

// parseOpenAPILoadError 解析OpenAPI加载错误
func parseOpenAPILoadError(ctx context.Context, originalErr error) *errors.HTTPError {
	if originalErr == nil {
		return nil
	}
	errorCode, errorParams, errorDetails := extractErrorInfo(originalErr, errorTypeLoadFailed)
	return errors.NewHTTPError(ctx, http.StatusBadRequest, errorCode, errorDetails, errorParams...)
}

// parseOpenAPIValidationError 解析OpenAPI验证错误
func parseOpenAPIValidationError(ctx context.Context, originalErr error) *errors.HTTPError {
	if originalErr == nil {
		return nil
	}
	errorCode, errorParams, errorDetails := extractErrorInfo(originalErr, errorTypeValidationFailed)
	return errors.NewHTTPError(ctx, http.StatusBadRequest, errorCode, errorDetails, errorParams...)
}

// extractErrorInfo 提取错误信息，返回错误码、参数和错误详情
func extractErrorInfo(err error, errorType string) (errors.ErrorCode, []interface{}, interface{}) {
	// 检查是否是MultiError类型
	if multiErr, ok := err.(*openapi3.MultiError); ok {
		return handleMultiError(multiErr, errorType)
	}

	// 单个错误处理
	return handleSingleError(err, errorType)
}

// handleMultiError 处理MultiError类型，返回最外层错误码，其他错误在详情中补充
func handleMultiError(multiErr *openapi3.MultiError, errorType string) (errors.ErrorCode, []interface{}, interface{}) {
	var (
		mainErrorCode errors.ErrorCode
		mainParams    []interface{}
		errorDetails  []string
	)

	// 遍历所有子错误
	for i, subErr := range *multiErr {
		if subErr != nil {
			subErrorCode, subParams, subErrorDetails := handleSingleError(subErr, errorType)

			// 第一个错误作为主错误码
			if i == 0 {
				mainErrorCode = subErrorCode
				mainParams = subParams
			}

			// 所有错误都添加到详情中
			errorDetails = append(errorDetails, fmt.Sprintf("错误%d: %s", i+1, subErrorDetails))
		}
	}

	// 如果没有找到主错误码，使用默认错误码
	if mainErrorCode == "" {
		mainErrorCode = getDefaultErrorCode(errorType)
	}

	return mainErrorCode, mainParams, errorDetails
}

// getDefaultErrorCode 获取默认错误码
func getDefaultErrorCode(errorType string) errors.ErrorCode {
	switch errorType {
	case errorTypeLoadFailed:
		return errors.ErrExtOpenAPISyntaxInvalid
	case errorTypeValidationFailed:
		return errors.ErrExtOpenAPIInvalidSpecification
	default:
		return errors.ErrExtOpenAPISyntaxInvalid
	}
}

// handleSingleError 处理单个错误
func handleSingleError(err error, errorType string) (errors.ErrorCode, []interface{}, interface{}) {
	errStr := err.Error()

	// 根据错误类型获取对应的错误码、参数和详细错误信息
	var errorCode errors.ErrorCode
	var errorParams []interface{}
	errorDetails := errStr

	switch errorType {
	case errorTypeLoadFailed:
		errorCode = errors.ErrExtOpenAPISyntaxInvalid
	case errorTypeValidationFailed:
		errorCode, errorParams = getValidationErrorCodeAndParams(errStr)
	default:
		errorCode, errorParams = getGenericErrorCodeAndParams(errStr)
	}
	return errorCode, errorParams, errorDetails
}

// extractElement 从错误信息中提取具体元素
func extractElement(errStr string, regex *regexp.Regexp) string {
	matches := regex.FindStringSubmatch(errStr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getGenericErrorCodeAndParams 获取通用错误码和参数
func getGenericErrorCodeAndParams(errStr string) (errors.ErrorCode, []interface{}) {
	field := extractElement(errStr, fieldRegex)

	if strings.Contains(errStr, "required") && field != "" {
		return errors.ErrExtOpenAPIInvalidSpecificationRequired, []interface{}{field}
	}
	if strings.Contains(errStr, "invalid") && field != "" {
		return errors.ErrExtOpenAPIInvalidSpecificationInvalid, []interface{}{field}
	}
	if strings.Contains(errStr, "missing") && field != "" {
		return errors.ErrExtOpenAPIInvalidSpecificationMissing, []interface{}{field}
	}
	if strings.Contains(errStr, "duplicate") && field != "" {
		return errors.ErrExtOpenAPIInvalidSpecificationDuplicate, []interface{}{field}
	}

	return errors.ErrExtOpenAPIInvalidSpecificationOperation, nil
}

// getValidationErrorCodeAndParams 获取验证阶段的错误码和参数（简化版本）
func getValidationErrorCodeAndParams(errStr string) (errors.ErrorCode, []interface{}) {
	if strings.Contains(errStr, "invalid components") {
		return errors.ErrExtOpenAPIInvalidComponent, nil
	}

	// 处理 "invalid info" 错误
	if strings.Contains(errStr, "invalid info") {
		return errors.ErrExtOpenAPIInvalidSpecificationRequired, []interface{}{"info"}
	}

	// 处理 "must be an object" 错误
	if strings.Contains(errStr, "must be an object") {
		// 提取具体的元素名称
		if strings.Contains(errStr, "info") {
			return errors.ErrExtOpenAPIInvalidSpecificationRequired, []interface{}{"info"}
		}
		if strings.Contains(errStr, "components") {
			return errors.ErrExtOpenAPIInvalidComponent, nil
		}
		if strings.Contains(errStr, "path") {
			return errors.ErrExtOpenAPIInvalidPath, nil
		}

		// 通用处理
		return errors.ErrExtOpenAPIInvalidSpecificationRequired, nil
	}

	// 处理 "value of openapi must be a non-empty string" 错误
	if strings.Contains(errStr, "value of openapi must be a non-empty string") {
		return errors.ErrExtOpenAPIInvalidSpecificationRequired, []interface{}{"openapi"}
	}
	// 1. Schema相关错误（最高优先级）
	if strings.Contains(errStr, "schema") {
		schema := extractElement(errStr, schemaRegex)
		if schema != "" {
			if strings.Contains(errStr, "ref") {
				return errors.ErrExtOpenAPIInvalidSchemaRef, []interface{}{schema}
			}
			if strings.Contains(errStr, "type") {
				return errors.ErrExtOpenAPIInvalidSchemaType, []interface{}{schema}
			}
			return errors.ErrExtOpenAPIInvalidSchemaType, []interface{}{schema}
		}
		return errors.ErrExtOpenAPIInvalidSchemaValue, nil
	}

	// 2. 参数相关错误
	if strings.Contains(errStr, "parameter") {
		parameter := extractElement(errStr, parameterRegex)
		if parameter != "" {
			if strings.Contains(errStr, "required") {
				return errors.ErrExtOpenAPIInvalidParameterRequired, []interface{}{parameter}
			}
			if strings.Contains(errStr, "schema") {
				return errors.ErrExtOpenAPIInvalidParameterSchema, []interface{}{parameter}
			}
			return errors.ErrExtOpenAPIInvalidParameterDefinition, []interface{}{parameter}
		}
		return errors.ErrExtOpenAPIInvalidParameterValue, nil
	}

	// 3. 响应相关错误
	if strings.Contains(errStr, "response") {
		response := extractElement(errStr, responseRegex)
		if response != "" {
			if strings.Contains(errStr, "required") {
				return errors.ErrExtOpenAPIInvalidResponseRequired, []interface{}{response}
			}
			return errors.ErrExtOpenAPIInvalidResponseDefinition, []interface{}{response}
		}
		return errors.ErrExtOpenAPIInvalidResponseSchema, nil
	}

	// 4. 路径相关错误
	if strings.Contains(errStr, "path") {
		return errors.ErrExtOpenAPIInvalidPath, nil
	}

	// 5. 操作相关错误
	if strings.Contains(errStr, "operation") {
		operation := extractElement(errStr, operationRegex)
		if operation != "" {
			return errors.ErrExtOpenAPIInvalidSpecificationOperation, []interface{}{operation}
		}
		return errors.ErrExtOpenAPIInvalidSpecificationOperation, nil
	}

	// 6. 通用验证错误
	field := extractElement(errStr, fieldRegex)
	if field != "" {
		if strings.Contains(errStr, "required") {
			return errors.ErrExtOpenAPIInvalidSpecificationRequired, []interface{}{field}
		}
		if strings.Contains(errStr, "missing") {
			return errors.ErrExtOpenAPIInvalidSpecificationMissing, []interface{}{field}
		}
		if strings.Contains(errStr, "invalid") {
			return errors.ErrExtOpenAPIInvalidSpecificationInvalid, []interface{}{field}
		}
		if strings.Contains(errStr, "duplicate") {
			return errors.ErrExtOpenAPIInvalidSpecificationDuplicate, []interface{}{field}
		}
	}

	// 7. 默认错误
	return errors.ErrExtOpenAPIInvalidSpecification, nil
}
