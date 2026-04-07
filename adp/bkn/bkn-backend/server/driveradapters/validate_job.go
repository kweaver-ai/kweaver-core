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

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func ValidateJob(ctx context.Context, jobInfo *interfaces.JobInfo) error {

	// 校验名称合法性
	// 去掉名称的前后空格
	jobInfo.Name = strings.TrimSpace(jobInfo.Name)
	err := validateObjectName(ctx, jobInfo.Name, interfaces.MODULE_TYPE_JOB)
	if err != nil {
		return err
	}

	// 如果job的branch为空，则赋值 main
	if jobInfo.Branch == "" {
		jobInfo.Branch = interfaces.MAIN_BRANCH
	}

	err = ValidateJobType(ctx, jobInfo.JobType)
	if err != nil {
		return err
	}

	if jobInfo.JobConceptConfig != nil || len(jobInfo.JobConceptConfig) != 0 {
		for _, conceptConfig := range jobInfo.JobConceptConfig {
			err = ValidateConceptType(ctx, conceptConfig.ConceptType)
			if err != nil {
				return err
			}

			if conceptConfig.ConceptID == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_Job_InvalidParameter_JobConceptConfig).
					WithErrorDetails("The job concept config must contain concept_id")
			}
		}
	}
	return nil
}

func ValidateJobType(ctx context.Context, jobType interfaces.JobType) error {
	switch jobType {
	case interfaces.JobTypeFull:
	case interfaces.JobTypeIncremental:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_Job_InvalidParameter_JobType).
			WithErrorDetails(fmt.Sprintf("The job_type value can only be 'full', 'incremental', but got: %s", jobType))
	}

	return nil
}

func ValidateJobState(ctx context.Context, jobState interfaces.JobState) error {
	switch jobState {
	case interfaces.JobStatePending:
	case interfaces.JobStateRunning:
	case interfaces.JobStateCompleted:
	case interfaces.JobStateCanceled:
	case interfaces.JobStateFailed:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_Job_InvalidParameter_JobState).
			WithErrorDetails(fmt.Sprintf("The job state value can be 'pending', 'running', 'completed', 'canceled', 'failed', but got: %s", jobState))
	}
	return nil
}

func ValidateTaskState(ctx context.Context, taskState interfaces.TaskState) error {
	switch taskState {
	case interfaces.TaskStatePending:
	case interfaces.TaskStateRunning:
	case interfaces.TaskStateCompleted:
	case interfaces.TaskStateCanceled:
	case interfaces.TaskStateFailed:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_Job_InvalidParameter_TaskState).
			WithErrorDetails(fmt.Sprintf("The task state value can be 'pending', 'running', 'completed', 'canceled', 'failed', but got: %s", taskState))
	}
	return nil
}

func ValidateConceptType(ctx context.Context, conceptType string) error {
	switch conceptType {
	case interfaces.MODULE_TYPE_OBJECT_TYPE:
	case interfaces.MODULE_TYPE_RELATION_TYPE:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_Job_InvalidParameter_ConceptType).
			WithErrorDetails(fmt.Sprintf("The concept_type value can be 'object_type' or 'relation_type', but got: %s", conceptType))
	}
	return nil
}
