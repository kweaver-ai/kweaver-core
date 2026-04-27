// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func Test_ValidateJob(t *testing.T) {
	Convey("Test ValidateJob\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid job\n", func() {
			job := &interfaces.JobInfo{
				Name:    "job1",
				JobType: interfaces.JobTypeFull,
			}
			err := ValidateJob(ctx, job)
			So(err, ShouldBeNil)
		})

		Convey("Failed with empty name\n", func() {
			job := &interfaces.JobInfo{
				Name:    "",
				JobType: interfaces.JobTypeFull,
			}
			err := ValidateJob(ctx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty concept ID in config\n", func() {
			job := &interfaces.JobInfo{
				Name:    "job1",
				JobType: interfaces.JobTypeFull,
				JobConceptConfig: []interfaces.ConceptConfig{
					{
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
						ConceptID:   "",
					},
				},
			}
			err := ValidateJob(ctx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid concept type in config\n", func() {
			job := &interfaces.JobInfo{
				Name:    "job1",
				JobType: interfaces.JobTypeFull,
				JobConceptConfig: []interfaces.ConceptConfig{
					{
						ConceptType: "invalid_type",
						ConceptID:   "id1",
					},
				},
			}
			err := ValidateJob(ctx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid job type\n", func() {
			job := &interfaces.JobInfo{
				Name:    "job1",
				JobType: interfaces.JobType("invalid"),
			}
			err := ValidateJob(ctx, job)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidateJobType(t *testing.T) {
	Convey("Test ValidateJobType\n", t, func() {
		ctx := context.Background()

		Convey("Success with full type\n", func() {
			err := ValidateJobType(ctx, interfaces.JobTypeFull)
			So(err, ShouldBeNil)
		})

		Convey("Success with incremental type\n", func() {
			err := ValidateJobType(ctx, interfaces.JobTypeIncremental)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid type\n", func() {
			err := ValidateJobType(ctx, interfaces.JobType("invalid"))
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InvalidParameter_JobType)
		})
	})
}

func Test_ValidateJobState(t *testing.T) {
	Convey("Test ValidateJobState\n", t, func() {
		ctx := context.Background()

		validStates := []interfaces.JobState{
			interfaces.JobStatePending,
			interfaces.JobStateRunning,
			interfaces.JobStateCompleted,
			interfaces.JobStateCanceled,
			interfaces.JobStateFailed,
		}

		for _, state := range validStates {
			Convey("Success with "+string(state)+"\n", func() {
				err := ValidateJobState(ctx, state)
				So(err, ShouldBeNil)
			})
		}

		Convey("Failed with invalid state\n", func() {
			err := ValidateJobState(ctx, interfaces.JobState("invalid"))
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InvalidParameter_JobState)
		})
	})
}

func Test_ValidateTaskState(t *testing.T) {
	Convey("Test ValidateTaskState\n", t, func() {
		ctx := context.Background()

		validStates := []interfaces.TaskState{
			interfaces.TaskStatePending,
			interfaces.TaskStateRunning,
			interfaces.TaskStateCompleted,
			interfaces.TaskStateCanceled,
			interfaces.TaskStateFailed,
		}

		for _, state := range validStates {
			Convey("Success with "+string(state)+"\n", func() {
				err := ValidateTaskState(ctx, state)
				So(err, ShouldBeNil)
			})
		}

		Convey("Failed with invalid state\n", func() {
			err := ValidateTaskState(ctx, interfaces.TaskState("invalid"))
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InvalidParameter_TaskState)
		})
	})
}

func Test_ValidateConceptType(t *testing.T) {
	Convey("Test ValidateConceptType\n", t, func() {
		ctx := context.Background()

		Convey("Success with object_type\n", func() {
			err := ValidateConceptType(ctx, interfaces.MODULE_TYPE_OBJECT_TYPE)
			So(err, ShouldBeNil)
		})

		Convey("Success with relation_type\n", func() {
			err := ValidateConceptType(ctx, interfaces.MODULE_TYPE_RELATION_TYPE)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid type\n", func() {
			err := ValidateConceptType(ctx, "invalid_type")
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InvalidParameter_ConceptType)
		})
	})
}
