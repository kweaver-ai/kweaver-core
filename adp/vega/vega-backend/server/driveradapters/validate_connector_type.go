// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

func ValidateConnectorTypeReq(ctx context.Context, req *interfaces.ConnectorTypeReq) error {
	if err := ValidateConnectorMode(ctx, req.Mode); err != nil {
		return err
	}
	if err := ValidateConnectorCategory(ctx, req.Category); err != nil {
		return err
	}

	// Remote mode requires endpoint
	if req.Mode == interfaces.ConnectorModeRemote && req.Endpoint == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Endpoint).
			WithErrorDetails("Remote connector requires endpoint URL")
	}

	return nil
}

func ValidateConnectorMode(ctx context.Context, mode string) error {
	if mode == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Mode)
	}
	switch mode {
	case interfaces.ConnectorModeLocal:
	case interfaces.ConnectorModeRemote:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Mode)
	}
	return nil
}

func ValidateConnectorCategory(ctx context.Context, category string) error {
	if category == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Category)
	}
	switch category {
	case interfaces.ConnectorCategoryTable:
	case interfaces.ConnectorCategoryIndex:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_ConnectorType_InvalidParameter_Category)
	}
	return nil
}
