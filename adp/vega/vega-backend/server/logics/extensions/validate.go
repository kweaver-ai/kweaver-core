// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package extensions 提供实体级 / 字段级 extensions 校验（Issue #382 方案 B）
package extensions

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

const (
	MaxEntityExtensionPairs   = 64
	MaxPropertyExtensionPairs = 32
	MaxExtensionKeyLen        = 128
	MaxExtensionValueLen      = 512
	MaxExtensionFilterPairs   = 5
	ReservedKeyPrefix         = "vega_"
)

// ValidateEntityExtensionsMap 校验根级 extensions：扁平 string→string、条数与长度、保留前缀。
func ValidateEntityExtensionsMap(ctx context.Context, m map[string]string) error {
	if len(m) > MaxEntityExtensionPairs {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_QuotaExceeded).
			WithErrorDetails(fmt.Sprintf("extensions 最多 %d 条", MaxEntityExtensionPairs))
	}
	for k, v := range m {
		if err := validateOnePair(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}

// ValidatePropertyExtensionsMap 校验 Property.extensions（配额略紧）。
func ValidatePropertyExtensionsMap(ctx context.Context, m map[string]string) error {
	if len(m) > MaxPropertyExtensionPairs {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_PropertyQuotaExceeded).
			WithErrorDetails(fmt.Sprintf("单字段 extensions 最多 %d 条", MaxPropertyExtensionPairs))
	}
	for k, v := range m {
		if err := validateOnePair(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}

// ValidateSchemaPropertiesExtensions 校验 schema_definition 中各 Property 的 extensions。
func ValidateSchemaPropertiesExtensions(ctx context.Context, props []*interfaces.Property) error {
	for _, p := range props {
		if p == nil || len(p.Extensions) == 0 {
			continue
		}
		if err := ValidatePropertyExtensionsMap(ctx, p.Extensions); err != nil {
			return err
		}
	}
	return nil
}

func validateOnePair(ctx context.Context, k, v string) error {
	if k == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_InvalidFormat).
			WithErrorDetails("extensions 的 key 不能为空")
	}
	if len(k) > MaxExtensionKeyLen {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_InvalidFormat).
			WithErrorDetails(fmt.Sprintf("extensions key 长度不能超过 %d", MaxExtensionKeyLen))
	}
	if strings.HasPrefix(strings.ToLower(k), ReservedKeyPrefix) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_ReservedKey).
			WithErrorDetails("extensions key 不能使用保留前缀 vega_")
	}
	if len(v) > MaxExtensionValueLen {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_InvalidFormat).
			WithErrorDetails(fmt.Sprintf("extensions value 长度不能超过 %d", MaxExtensionValueLen))
	}
	return nil
}

// ValidateExtensionQueryPairs 校验列表筛选 extension_key / extension_value 成对且数量上限。
func ValidateExtensionQueryPairs(ctx context.Context, keys, values []string) error {
	if len(keys) == 0 && len(values) == 0 {
		return nil
	}
	if len(keys) != len(values) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_MismatchedQueryPairs).
			WithErrorDetails("extension_key 与 extension_value 必须成对且数量一致")
	}
	if len(keys) > MaxExtensionFilterPairs {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_TooManyFilterPairs).
			WithErrorDetails(fmt.Sprintf("扩展筛选条件最多 %d 组", MaxExtensionFilterPairs))
	}
	for i := range keys {
		if keys[i] == "" || values[i] == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Extensions_MismatchedQueryPairs).
				WithErrorDetails("extension_key 与 extension_value 不能为空")
		}
	}
	return nil
}
