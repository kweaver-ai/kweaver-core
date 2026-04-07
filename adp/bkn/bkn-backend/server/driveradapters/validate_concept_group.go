// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"strings"

	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"

	"bkn-backend/interfaces"
)

// 指标模型必要创建参数的非空校验。bool 为 dsl 语句中是否使用了 top_hits 的标识。
func ValidateConceptGroup(ctx context.Context, cg *interfaces.ConceptGroup) error {
	// 校验id的合法性
	err := validateID(ctx, cg.CGID)
	if err != nil {
		return err
	}

	// 校验名称合法性
	// 去掉模型名称的前后空格
	cg.CGName = strings.TrimSpace(cg.CGName)
	err = validateObjectName(ctx, cg.CGName, interfaces.MODULE_TYPE_CONCEPT_GROUP)
	if err != nil {
		return err
	}

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, cg.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	cg.Tags = libCommon.TagSliceTransform(cg.Tags)

	// 校验comment合法性
	err = validateObjectComment(ctx, cg.Comment)
	if err != nil {
		return err
	}

	return nil
}
