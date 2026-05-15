// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/extensions"
)

func ValidateCatalogRequest(ctx context.Context, req *interfaces.CatalogRequest) error {
	if err := validateName(ctx, req.Name); err != nil {
		return err
	}
	if err := ValidateTags(ctx, req.Tags); err != nil {
		return err
	}
	if err := validateDescription(ctx, req.Description); err != nil {
		return err
	}
	if err := validateConnectorConfig(ctx, req.ConnectorCfg); err != nil {
		return err
	}
	if req.Extensions != nil {
		if err := extensions.ValidateEntityExtensionsMap(ctx, *req.Extensions); err != nil {
			return err
		}
	}
	return nil
}

func ValidateCatalogListQueryParams(ctx context.Context, params interfaces.CatalogsQueryParams) error {
	if err := validateCatalogTypeQueryParam(ctx, params.Type); err != nil {
		return err
	}
	if err := validateCatalogHealthCheckStatusQueryParam(ctx, params.HealthCheckStatus); err != nil {
		return err
	}
	if err := extensions.ValidateExtensionQueryPairs(ctx, params.ExtensionKeys, params.ExtensionValues); err != nil {
		return err
	}
	return nil
}

func validateCatalogTypeQueryParam(ctx context.Context, typ string) error {
	if typ == "" {
		return nil
	}

	switch typ {
	case interfaces.CatalogTypePhysical,
		interfaces.CatalogTypeLogical:
		return nil
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_Type).
			WithErrorDetails(fmt.Sprintf("invalid type: %s", typ))
	}
}

func validateCatalogHealthCheckStatusQueryParam(ctx context.Context, status string) error {
	if status == "" {
		return nil
	}

	switch status {
	case interfaces.CatalogHealthStatusHealthy,
		interfaces.CatalogHealthStatusDegraded,
		interfaces.CatalogHealthStatusUnhealthy,
		interfaces.CatalogHealthStatusOffline,
		interfaces.CatalogHealthStatusDisabled:
		return nil
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("invalid health_check_status: %s", status))
	}
}
