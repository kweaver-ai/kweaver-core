// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knontologyjob 提供本体构建任务管理的 HTTP 处理器
package knontologyjob

import (
	"net/http"
	"sync"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// KnOntologyJobHandler Ontology job management handler
type KnOntologyJobHandler interface {
	FullBuildOntology(c *gin.Context)
	GetFullOntologyBuildingStatus(c *gin.Context)
}

type knOntologyJobHandler struct {
	Logger           interfaces.Logger
	BknBackendAccess interfaces.BknBackendAccess
}

// DefaultJobListLimit 默认查询任务列表的数量限制
const DefaultJobListLimit = 50

var (
	kojOnce    sync.Once
	kojHandler KnOntologyJobHandler
)

// NewKnOntologyJobHandler Create new KnOntologyJobHandler
func NewKnOntologyJobHandler() KnOntologyJobHandler {
	kojOnce.Do(func() {
		conf := config.NewConfigLoader()
		kojHandler = &knOntologyJobHandler{
			Logger:           conf.GetLogger(),
			BknBackendAccess: drivenadapters.NewBknBackendAccess(),
		}
	})
	return kojHandler
}

// FullBuildOntology Create a full ontology build job
// POST /api/agent-retrieval/in/v1/kn/full_build_ontology
func (h *knOntologyJobHandler) FullBuildOntology(c *gin.Context) {
	var err error

	// Bind request from JSON body including kn_id
	reqBody := struct {
		KnID string `json:"kn_id" binding:"required"`
		interfaces.CreateFullBuildOntologyJobReq
	}{}

	if err = c.ShouldBindJSON(&reqBody); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// Create request object
	req := &interfaces.CreateFullBuildOntologyJobReq{
		Name: reqBody.Name,
	}

	// Set default values
	if err = defaults.Set(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// Validate request
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// Call BKN backend to create job
	resp, err := h.BknBackendAccess.CreateFullBuildOntologyJob(c.Request.Context(), reqBody.KnID, req)
	if err != nil {
		h.Logger.Errorf("[KnOntologyJobHandler#FullBuildOntology] CreateFullBuildOntologyJob failed, knId: %s, err: %v", reqBody.KnID, err)
		rest.ReplyError(c, err)
		return
	}

	// Return success response
	rest.ReplyOK(c, http.StatusCreated, resp)
}

// GetFullOntologyBuildingStatus Get ontology build job status
// GET /api/agent-retrieval/in/v1/kn/full_ontology_building_status
func (h *knOntologyJobHandler) GetFullOntologyBuildingStatus(c *gin.Context) {
	var err error

	// Get kn_id from query
	knID := c.Query("kn_id")
	if knID == "" {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "kn_id is required")
		rest.ReplyError(c, err)
		return
	}

	// Create list request to get latest 50 jobs
	req := &interfaces.ListOntologyJobsReq{
		JobType:   interfaces.OntologyJobTypeFull,
		Limit:     DefaultJobListLimit,
		Direction: "desc", // Descending by create_time
	}

	// Call BKN backend to list jobs
	resp, err := h.BknBackendAccess.ListOntologyJobs(c.Request.Context(), knID, req)
	if err != nil {
		h.Logger.Errorf("[KnOntologyJobHandler#GetFullOntologyBuildingStatus] ListOntologyJobs failed, knId: %s, err: %v", knID, err)
		rest.ReplyError(c, err)
		return
	}

	// Determine overall state based on the latest 50 jobs
	overallState := interfaces.OntologyJobStateCompleted
	stateDetail := "All latest 50 jobs are completed"

	for _, job := range resp.Entries {
		if job.State == interfaces.OntologyJobStateRunning {
			overallState = interfaces.OntologyJobStateRunning
			stateDetail = "Some jobs are still running"
			break
		}
	}

	// Build simplified response
	response := map[string]interface{}{
		"kn_id":        knID,
		"state":        overallState,
		"state_detail": stateDetail,
	}

	// Return the simplified response
	rest.ReplyOK(c, http.StatusOK, response)
}
