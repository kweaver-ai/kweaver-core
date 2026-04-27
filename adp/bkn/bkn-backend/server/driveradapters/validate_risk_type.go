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

	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

// ValidateRiskTypes 校验风险类创建请求
func ValidateRiskTypes(ctx context.Context, knID string, riskTypes []*interfaces.RiskType) error {
	tmpNameMap := make(map[string]any)
	idMap := make(map[string]any)
	for i := range riskTypes {
		riskType := riskTypes[i]
		if riskType.ModuleType != "" && riskType.ModuleType != interfaces.MODULE_TYPE_RISK_TYPE {
			return rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_InvalidParameter_ModuleType).
				WithErrorDetails("Risk type module type is not 'risk_type'")
		}

		rtID := riskType.RTID
		if _, ok := idMap[rtID]; !ok || rtID == "" {
			idMap[rtID] = nil
		} else {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_Duplicated_IDInFile).
				WithDescription(map[string]any{"riskTypeID": rtID}).
				WithErrorDetails(fmt.Sprintf("RiskType ID '%s' already exists in the request body", rtID))
		}

		err := ValidateRiskType(ctx, riskType)
		if err != nil {
			return err
		}

		if _, ok := tmpNameMap[riskType.RTName]; !ok {
			tmpNameMap[riskType.RTName] = nil
		} else {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_Duplicated_Name).
				WithDescription(map[string]any{"riskTypeName": riskType.RTName}).
				WithErrorDetails(fmt.Sprintf("RiskType name '%s' already exists in the request body", riskType.RTName))
		}

		riskType.KNID = knID
	}
	return nil
}

// ValidateRiskType 校验单个风险类
func ValidateRiskType(ctx context.Context, riskType *interfaces.RiskType) error {
	err := validateID(ctx, riskType.RTID)
	if err != nil {
		return err
	}

	riskType.RTName = strings.TrimSpace(riskType.RTName)
	err = validateObjectName(ctx, riskType.RTName, interfaces.MODULE_TYPE_RISK_TYPE)
	if err != nil {
		return err
	}

	if err = ValidateTags(ctx, riskType.Tags); err != nil {
		return err
	}
	riskType.Tags = libCommon.TagSliceTransform(riskType.Tags)

	return nil
}
