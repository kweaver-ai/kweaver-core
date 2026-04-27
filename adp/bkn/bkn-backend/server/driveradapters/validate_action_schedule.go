// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

// ValidateActionScheduleCreate validates the create schedule request
func ValidateActionScheduleCreate(ctx context.Context, req *interfaces.ActionScheduleCreateRequest) error {
	// Validate name
	if req.Name == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("name is required")
	}
	if len(req.Name) > 100 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("name must be less than 100 characters")
	}

	// Validate action_type_id
	if req.ActionTypeID == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("action_type_id is required")
	}

	// Validate cron_expression
	if req.CronExpression == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("cron_expression is required")
	}

	// Validate _instance_identities
	if len(req.InstanceIdentities) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("_instance_identities is required and cannot be empty")
	}

	// Validate status if provided
	if req.Status != "" && req.Status != interfaces.ScheduleStatusActive && req.Status != interfaces.ScheduleStatusInactive {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidStatus).
			WithErrorDetails("status must be 'active' or 'inactive'")
	}

	return nil
}

// ValidateActionScheduleUpdate validates the update schedule request
func ValidateActionScheduleUpdate(ctx context.Context, req *interfaces.ActionScheduleUpdateRequest) error {
	// Validate name if provided
	if req.Name != "" && len(req.Name) > 100 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("name must be less than 100 characters")
	}

	// At least one field should be provided
	if req.Name == "" && req.CronExpression == "" && req.InstanceIdentities == nil && req.DynamicParams == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("at least one field must be provided for update")
	}

	return nil
}
