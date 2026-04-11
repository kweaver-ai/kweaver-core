// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knfindskills implements the find_skills skill recall service.
package knfindskills

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/localize"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type findSkillsServiceImpl struct {
	logger        interfaces.Logger
	config        *config.Config
	ontologyQuery interfaces.DrivenOntologyQuery
	bknBackend    interfaces.BknBackendAccess
	coordinator   *recallCoordinator
}

var (
	fsOnce               sync.Once
	findSkillsServiceInst interfaces.IFindSkillsService
)

// NewFindSkillsService creates a singleton FindSkillsService.
func NewFindSkillsService() interfaces.IFindSkillsService {
	fsOnce.Do(func() {
		cfg := config.NewConfigLoader()
		oq := drivenadapters.NewOntologyQueryAccess()
		bkn := drivenadapters.NewBknBackendAccess()
		findSkillsServiceInst = &findSkillsServiceImpl{
			logger:        cfg.GetLogger(),
			config:        cfg,
			ontologyQuery: oq,
			bknBackend:    bkn,
			coordinator: &recallCoordinator{
				logger:        cfg.GetLogger(),
				config:        &cfg.FindSkills,
				ontologyQuery: oq,
				bknBackend:    bkn,
			},
		}
	})
	return findSkillsServiceInst
}

// NewFindSkillsServiceWith creates a FindSkillsService with injected dependencies (for testing).
func NewFindSkillsServiceWith(
	logger interfaces.Logger,
	cfg *config.Config,
	oq interfaces.DrivenOntologyQuery,
	bkn interfaces.BknBackendAccess,
) interfaces.IFindSkillsService {
	return &findSkillsServiceImpl{
		logger:        logger,
		config:        cfg,
		ontologyQuery: oq,
		bknBackend:    bkn,
		coordinator: &recallCoordinator{
			logger:        logger,
			config:        &cfg.FindSkills,
			ontologyQuery: oq,
			bknBackend:    bkn,
		},
	}
}

// FindSkills is the main entry point for skill recall.
func (s *findSkillsServiceImpl) FindSkills(ctx context.Context, req *interfaces.FindSkillsReq) (*interfaces.FindSkillsResp, error) {
	var err error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	fsCfg := &s.config.FindSkills

	// 1. Normalize & detect recall mode
	mode, err := NormalizeAndDetectMode(req, fsCfg)
	if err != nil {
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}

	s.logger.WithContext(ctx).Infof("[FindSkills] kn_id=%s, mode=%d, object_type_id=%s, instance_count=%d, has_skill_query=%v",
		req.KnID, mode, req.ObjectTypeID, len(req.InstanceIdentities), req.SkillQuery != "")

	// 2. Build skill_query condition (needs skills ObjectType metadata)
	var skillQueryCond *interfaces.KnCondition
	if req.SkillQuery != "" {
		skillsObjTypes, getErr := s.bknBackend.GetObjectTypeDetail(ctx, req.KnID, []string{fsCfg.SkillsObjectTypeID}, true)
		if getErr != nil {
			err = fmt.Errorf("skill_query requires skills ObjectType metadata but GetObjectTypeDetail failed: %w", getErr)
			return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, err.Error())
		}
		if len(skillsObjTypes) == 0 {
			err = fmt.Errorf("skill_query requires skills ObjectType (id=%s) but none found in kn_id=%s", fsCfg.SkillsObjectTypeID, req.KnID)
			return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, err.Error())
		}
		skillQueryCond = BuildSkillQueryCondition(req.SkillQuery, skillsObjTypes[0], req.TopK)
	}

	// 3. Apply total timeout
	totalTimeoutMs := fsCfg.TotalTimeoutMs
	if totalTimeoutMs <= 0 {
		totalTimeoutMs = 10000
	}
	recallCtx, cancel := context.WithTimeout(ctx, time.Duration(totalTimeoutMs)*time.Millisecond)
	defer cancel()

	// 4. Execute recall based on mode
	var matches []interfaces.SkillMatch
	var emptyHint interfaces.EmptyResultHint
	switch mode {
	case interfaces.RecallModeNetwork:
		matches, emptyHint, err = s.coordinator.recallNetwork(recallCtx, req, skillQueryCond)
	case interfaces.RecallModeObjectType:
		matches, emptyHint, err = s.coordinator.recallObjectType(recallCtx, req, skillQueryCond)
	case interfaces.RecallModeInstance:
		matches, emptyHint, err = s.coordinator.recallInstance(recallCtx, req, skillQueryCond)
	default:
		err = fmt.Errorf("unknown recall mode: %d", mode)
	}

	if err != nil {
		s.logger.WithContext(ctx).Errorf("[FindSkills] recall failed: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, err.Error())
	}

	// 5. Assemble result
	resp := Assemble(matches, req.TopK)

	// 6. Generate empty-result message
	if len(resp.Entries) == 0 {
		msgKey := resolveEmptyResultMessageKey(emptyHint, mode, req.SkillQuery != "")
		resp.Message = translateMessage(ctx, msgKey)
	}

	s.logger.WithContext(ctx).Infof("[FindSkills] returning %d skills for kn_id=%s", len(resp.Entries), req.KnID)
	return resp, nil
}

func resolveEmptyResultMessageKey(hint interfaces.EmptyResultHint, mode interfaces.RecallMode, hasSkillQuery bool) string {
	if hint != interfaces.HintNone {
		return string(hint)
	}
	if hasSkillQuery {
		return "find_skills.skill_query_no_match"
	}
	switch mode {
	case interfaces.RecallModeNetwork:
		return "find_skills.network_no_skills"
	case interfaces.RecallModeObjectType:
		return "find_skills.object_type_no_match"
	case interfaces.RecallModeInstance:
		return "find_skills.instance_no_match"
	default:
		return "find_skills.network_no_skills"
	}
}

func translateMessage(ctx context.Context, msgKey string) string {
	lang := common.GetLanguageFromCtx(ctx)
	langKey := strings.ReplaceAll(lang, "-", "_")
	tr := localize.NewI18nTranslator(langKey)
	return tr.Trans(msgKey)
}
