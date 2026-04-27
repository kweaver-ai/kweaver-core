// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/interfaces"
)

func Test_ValidateActionScheduleCreate(t *testing.T) {
	Convey("Test ValidateActionScheduleCreate\n", t, func() {
		ctx := context.Background()

		valid := &interfaces.ActionScheduleCreateRequest{
			Name:               "schedule1",
			ActionTypeID:       "at1",
			CronExpression:     "0 * * * *",
			InstanceIdentities: []map[string]any{{"id": "i1"}},
		}

		Convey("Success with valid request\n", func() {
			err := ValidateActionScheduleCreate(ctx, valid)
			So(err, ShouldBeNil)
		})

		Convey("Success with active status\n", func() {
			req := *valid
			req.Status = interfaces.ScheduleStatusActive
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldBeNil)
		})

		Convey("Success with inactive status\n", func() {
			req := *valid
			req.Status = interfaces.ScheduleStatusInactive
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldBeNil)
		})

		Convey("Failed when name is empty\n", func() {
			req := *valid
			req.Name = ""
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when name exceeds 100 characters\n", func() {
			req := *valid
			req.Name = strings.Repeat("a", 101)
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when action_type_id is empty\n", func() {
			req := *valid
			req.ActionTypeID = ""
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when cron_expression is empty\n", func() {
			req := *valid
			req.CronExpression = ""
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when instance_identities is empty\n", func() {
			req := *valid
			req.InstanceIdentities = []map[string]any{}
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when status is invalid\n", func() {
			req := *valid
			req.Status = "unknown"
			err := ValidateActionScheduleCreate(ctx, &req)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidateActionScheduleUpdate(t *testing.T) {
	Convey("Test ValidateActionScheduleUpdate\n", t, func() {
		ctx := context.Background()

		Convey("Success when name is provided\n", func() {
			req := &interfaces.ActionScheduleUpdateRequest{Name: "new-name"}
			err := ValidateActionScheduleUpdate(ctx, req)
			So(err, ShouldBeNil)
		})

		Convey("Success when cron_expression is provided\n", func() {
			req := &interfaces.ActionScheduleUpdateRequest{CronExpression: "0 0 * * *"}
			err := ValidateActionScheduleUpdate(ctx, req)
			So(err, ShouldBeNil)
		})

		Convey("Success when instance_identities is provided\n", func() {
			req := &interfaces.ActionScheduleUpdateRequest{
				InstanceIdentities: []map[string]any{{"id": "i1"}},
			}
			err := ValidateActionScheduleUpdate(ctx, req)
			So(err, ShouldBeNil)
		})

		Convey("Failed when name exceeds 100 characters\n", func() {
			req := &interfaces.ActionScheduleUpdateRequest{Name: strings.Repeat("b", 101)}
			err := ValidateActionScheduleUpdate(ctx, req)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when all fields are empty\n", func() {
			req := &interfaces.ActionScheduleUpdateRequest{}
			err := ValidateActionScheduleUpdate(ctx, req)
			So(err, ShouldNotBeNil)
		})
	})
}
