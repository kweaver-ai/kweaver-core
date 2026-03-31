// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"

	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// 业务知识网络必要创建参数的非空校验。bool 为 dsl 语句中是否使用了 top_hits 的标识。
func ValidateKN(ctx context.Context, kn *interfaces.KN) error {
	// 校验id的合法性
	err := validateID(ctx, kn.KNID)
	if err != nil {
		return err
	}

	// 校验名称合法性
	// 去掉模型名称的前后空格
	kn.KNName = strings.TrimSpace(kn.KNName)
	err = validateObjectName(ctx, kn.KNName, interfaces.MODULE_TYPE_KN)
	if err != nil {
		return err
	}

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, kn.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	kn.Tags = libCommon.TagSliceTransform(kn.Tags)

	// 校验comment合法性
	err = validateObjectComment(ctx, kn.Comment)
	if err != nil {
		return err
	}

	// 校验分支非空
	if kn.Branch == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_NullParameter_Branch).
			WithErrorDetails("branch must be set")
	}

	return nil
}

// 路径查询的参数校验
func ValidateRelationTypePathsQuery(ctx context.Context, query *interfaces.RelationTypePathsBaseOnSource) error {
	// 起点对象类非空
	if query.SourceObjecTypeId == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_NullParameter_SourceObjectTypeId)
	}

	// 方向非空
	if query.Direction == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_NullParameter_Direction)
	}

	// 方向有效性
	if !interfaces.DIRECTION_MAP[query.Direction] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter_Direction).
			WithErrorDetails(fmt.Sprintf("当前支持的方向有: forward, backward, bidirectional. 请求的方向为: %s", query.Direction))
	}

	// 路径长度不超过3
	if query.PathLength > 3 || query.PathLength < 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter_PathLength).
			WithErrorDetails(fmt.Sprintf("路径长度不超过3, 请求的路径长度为%d", query.PathLength))
	}

	return nil
}
