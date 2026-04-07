// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

func validateID(ctx context.Context, ID string) error {
	if ID == "" {
		return nil
	}

	// 非内置视图校验数据视图 id，只包含小写英文字母和数字和下划线(_)和连字符(-)，且不能以下划线开头，不能超过40个字符
	re := regexp2.MustCompile(interfaces.RegexPattern_NonBuiltin_ViewID, regexp2.RE2)
	match, err := re.MatchString(ID)
	if err != nil || !match {
		errDetails := `The ID can contain only lowercase letters, digits and underscores(_),
			it cannot start with underscores and cannot exceed 40 characters`
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_ID).
			WithErrorDetails(errDetails)
	}

	return nil
}

// 名称合法性校验
func validateName(ctx context.Context, name string) error {
	if name == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Name)
	}

	if utf8.RuneCountInString(name) > interfaces.NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Name).
			WithErrorDetails(fmt.Sprintf("The length of the name %v exceeds %v", name, interfaces.NAME_MAX_LENGTH))
	}

	return nil
}

// tags 的合法性校验
func ValidateTags(ctx context.Context, Tags []string) error {
	if len(Tags) > interfaces.TAGS_MAX_NUMBER {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tag).
			WithErrorDetails(fmt.Sprintf("The number of tags exceeds %v", interfaces.TAGS_MAX_NUMBER))
	}

	for _, tag := range Tags {
		err := validateTag(ctx, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// 数据标签名称合法性校验
func validateTag(ctx context.Context, tag string) error {
	// 去除tag的左右空格
	tag = strings.Trim(tag, " ")

	if tag == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tag)
		// .WithErrorDetails("Data tag name is null")
	}

	if utf8.RuneCountInString(tag) > interfaces.TAG_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tag).
			WithErrorDetails(fmt.Sprintf("The length of the tag name exceeds %d", interfaces.TAG_MAX_LENGTH))
	}

	if isInvalid := strings.ContainsAny(tag, interfaces.TAG_INVALID_CHARACTER); isInvalid {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tag).
			WithErrorDetails(fmt.Sprintf("Tag name contains special characters, such as %s", interfaces.TAG_INVALID_CHARACTER))
	}

	return nil
}

// 备注合法性校验
func validateDescription(ctx context.Context, description string) error {
	if utf8.RuneCountInString(description) > interfaces.DESCRIPTION_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Description).
			WithErrorDetails(fmt.Sprintf("The length of the description exceeds %v", interfaces.DESCRIPTION_MAX_LENGTH))
	}
	return nil
}

// 分页参数合法性校验
func validatePaginationQueryParams(ctx context.Context, offset, limit, sort, direction string,
	supportedSortTypes map[string]string) (interfaces.PaginationQueryParams, error) {
	pageParams := interfaces.PaginationQueryParams{}

	off, err := strconv.Atoi(offset)
	if err != nil {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Offset).
			WithErrorDetails(err.Error())
	}

	if off < interfaces.MIN_OFFSET {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Offset).
			WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET))
	}

	lim, err := strconv.Atoi(limit)
	if err != nil {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Limit).
			WithErrorDetails(err.Error())
	}

	if !(limit == interfaces.NO_LIMIT || (lim >= interfaces.MIN_LIMIT && lim <= interfaces.MAX_LIMIT)) {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("The number per page does not equal %s is not in the range of [%d,%d]", interfaces.NO_LIMIT, interfaces.MIN_LIMIT, interfaces.MAX_LIMIT))
	}

	_, ok := supportedSortTypes[sort]
	if !ok {
		types := make([]string, 0)
		for t := range supportedSortTypes {
			types = append(types, t)
		}
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Sort).
			WithErrorDetails(fmt.Sprintf("Wrong sort type, does not belong to any item in set %v ", types))
	}

	if direction != interfaces.DESC_DIRECTION && direction != interfaces.ASC_DIRECTION {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Direction).
			WithErrorDetails("The sort direction is not desc or asc")
	}

	return interfaces.PaginationQueryParams{
		Offset:    off,
		Limit:     lim,
		Sort:      supportedSortTypes[sort],
		Direction: direction,
	}, nil
}

// ConnectorConfig 合法性校验
func validateConnectorConfig(ctx context.Context, cfg interfaces.ConnectorConfig) error {
	// Check for duplicate elements in databases
	if dbValue, exists := cfg["databases"]; exists {
		if dbArray, ok := dbValue.([]any); ok {
			if err := checkDuplicateElements(ctx, dbArray, "databases"); err != nil {
				return err
			}
		}
	}

	// Check for duplicate elements in schemas
	if schemaValue, exists := cfg["schemas"]; exists {
		if schemaArray, ok := schemaValue.([]any); ok {
			if err := checkDuplicateElements(ctx, schemaArray, "schemas"); err != nil {
				return err
			}
		}
	}

	return nil
}

// 检查数组中是否存在重复元素
func checkDuplicateElements(ctx context.Context, arr []any, fieldName string) error {
	seen := make(map[string]bool)
	for _, item := range arr {
		strItem := fmt.Sprintf("%v", item)
		if seen[strItem] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ConnectorConfig).
				WithErrorDetails(fmt.Sprintf("duplicate element found in '%s': %s", fieldName, strItem))
		}
		seen[strItem] = true
	}
	return nil
}
