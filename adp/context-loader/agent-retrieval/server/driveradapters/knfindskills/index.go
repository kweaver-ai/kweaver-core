// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knfindskills provides the HTTP handler for the find_skills skill recall endpoint.
package knfindskills

import (
	"net/http"
	"sync"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	logicsFS "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knfindskills"
)

// KnFindSkillsHandler HTTP handler interface for find_skills
type KnFindSkillsHandler interface {
	FindSkills(c *gin.Context)
}

type knFindSkillsHandler struct {
	Logger            interfaces.Logger
	FindSkillsService interfaces.IFindSkillsService
}

var (
	fsHandlerOnce sync.Once
	fsHandler     KnFindSkillsHandler
)

// NewKnFindSkillsHandler creates a singleton KnFindSkillsHandler
func NewKnFindSkillsHandler() KnFindSkillsHandler {
	fsHandlerOnce.Do(func() {
		conf := config.NewConfigLoader()
		fsHandler = &knFindSkillsHandler{
			Logger:            conf.GetLogger(),
			FindSkillsService: logicsFS.NewFindSkillsService(),
		}
	})
	return fsHandler
}

// FindSkills handles POST /kn/find_skills
func (h *knFindSkillsHandler) FindSkills(c *gin.Context) {
	var err error
	req := &interfaces.FindSkillsReq{}

	if err = c.ShouldBindHeader(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	if err = c.ShouldBindJSON(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	if err = defaults.Set(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	resp, err := h.FindSkillsService.FindSkills(c.Request.Context(), req)
	if err != nil {
		h.Logger.Errorf("[KnFindSkillsHandler#FindSkills] FindSkills failed, err: %v", err)
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, resp)
}
