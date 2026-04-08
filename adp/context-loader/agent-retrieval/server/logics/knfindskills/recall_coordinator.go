// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"context"
	"fmt"
	"sync"
	"time"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// recallCoordinator orchestrates multi-path skill recall
type recallCoordinator struct {
	logger        interfaces.Logger
	config        *config.FindSkillsConfig
	ontologyQuery interfaces.DrivenOntologyQuery
	bknBackend    interfaces.BknBackendAccess
}

// recallNetwork handles Mode 1: network-level recall.
// Returns skills only when the skills ObjectType has NO relation to any other ObjectType.
func (rc *recallCoordinator) recallNetwork(
	ctx context.Context,
	req *interfaces.FindSkillsReq,
	skillQueryCond *interfaces.KnCondition,
) ([]interfaces.SkillMatch, error) {
	var err error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	hasRelation, err := rc.skillsHaveAnyRelation(ctx, req.KnID)
	if err != nil {
		return nil, fmt.Errorf("check skills relation: %w", err)
	}
	if hasRelation {
		rc.logger.WithContext(ctx).Infof("[FindSkills] skills ObjectType has relations, network-level recall returns empty")
		return nil, nil
	}

	oqReq := &interfaces.QueryObjectInstancesReq{
		KnID:       req.KnID,
		OtID:       rc.config.SkillsObjectTypeID,
		Cond:       skillQueryCond,
		Limit:      req.TopK,
		Properties: []string{"skill_id", "name", "description"},
	}

	resp, err := rc.ontologyQuery.QueryObjectInstances(ctx, oqReq)
	if err != nil {
		return nil, fmt.Errorf("QueryObjectInstances(skills): %w", err)
	}

	return extractSkillMatchesFromInstances(resp.Data, "network", 10), nil
}

// recallObjectType handles Mode 2: object-type-level recall.
func (rc *recallCoordinator) recallObjectType(
	ctx context.Context,
	req *interfaces.FindSkillsReq,
	skillQueryCond *interfaces.KnCondition,
) ([]interfaces.SkillMatch, error) {
	var err error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	rt, err := rc.findRelationType(ctx, req.KnID, req.ObjectTypeID)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		rc.logger.WithContext(ctx).Warnf("[FindSkills] no RelationType between %s and skills, returning empty", req.ObjectTypeID)
		return nil, nil
	}

	subReq := BuildSubgraphRequest(req.KnID, req.ObjectTypeID, rt, nil, skillQueryCond, req.TopK, rc.config.SkillsObjectTypeID)
	resp, err := rc.ontologyQuery.QueryInstanceSubgraph(ctx, subReq)
	if err != nil {
		return nil, fmt.Errorf("QueryInstanceSubgraph(object_type): %w", err)
	}

	return extractSkillMatchesFromSubgraph(resp, rc.config.SkillsObjectTypeID, "object_type", 50), nil
}

// recallInstance handles Mode 3: instance-level recall.
// Runs object-type-level and instance-level recalls concurrently.
func (rc *recallCoordinator) recallInstance(
	ctx context.Context,
	req *interfaces.FindSkillsReq,
	skillQueryCond *interfaces.KnCondition,
) ([]interfaces.SkillMatch, error) {
	var spanErr error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, spanErr)

	rt, err := rc.findRelationType(ctx, req.KnID, req.ObjectTypeID)
	if err != nil {
		spanErr = err
		return nil, err
	}
	if rt == nil {
		rc.logger.WithContext(ctx).Warnf("[FindSkills] no RelationType between %s and skills, returning empty", req.ObjectTypeID)
		return nil, nil
	}

	timeoutMs := rc.config.RecallTimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}
	recallCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	var (
		mu             sync.Mutex
		allMatches     []interfaces.SkillMatch
		otLevelErr     error
		instLevelErr   error
		wg             sync.WaitGroup
	)

	// Object-type-level path
	wg.Add(1)
	go func() {
		defer wg.Done()
		subReq := BuildSubgraphRequest(req.KnID, req.ObjectTypeID, rt, nil, skillQueryCond, req.TopK, rc.config.SkillsObjectTypeID)
		resp, err := rc.ontologyQuery.QueryInstanceSubgraph(recallCtx, subReq)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			otLevelErr = err
			rc.logger.WithContext(ctx).Warnf("[FindSkills] object_type-level recall failed: %v", err)
			return
		}
		allMatches = append(allMatches, extractSkillMatchesFromSubgraph(resp, rc.config.SkillsObjectTypeID, "object_type", 50)...)
	}()

	// Instance-level path
	wg.Add(1)
	go func() {
		defer wg.Done()
		instCond := buildInstanceFilterCondition(req.InstanceIdentities)
		subReq := BuildSubgraphRequest(req.KnID, req.ObjectTypeID, rt, instCond, skillQueryCond, req.TopK, rc.config.SkillsObjectTypeID)
		resp, err := rc.ontologyQuery.QueryInstanceSubgraph(recallCtx, subReq)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			instLevelErr = err
			rc.logger.WithContext(ctx).Warnf("[FindSkills] instance-level recall failed: %v", err)
			return
		}
		allMatches = append(allMatches, extractSkillMatchesFromSubgraph(resp, rc.config.SkillsObjectTypeID, "object_selector", 100)...)
	}()

	wg.Wait()

	if otLevelErr != nil && instLevelErr != nil {
		spanErr = fmt.Errorf("both recall paths failed: ot=%v, inst=%v", otLevelErr, instLevelErr)
		return nil, spanErr
	}

	return allMatches, nil
}

// skillsHaveAnyRelation checks if the skills ObjectType participates in any RelationType.
func (rc *recallCoordinator) skillsHaveAnyRelation(ctx context.Context, knID string) (bool, error) {
	query := &interfaces.QueryConceptsReq{
		KnID:  knID,
		Limit: 1,
		Cond: &interfaces.KnCondition{
			Operation: interfaces.KnOperationTypeOr,
			SubConditions: []*interfaces.KnCondition{
				{
					Field:     "source_object_type_id",
					Operation: interfaces.KnOperationTypeEqual,
					Value:     rc.config.SkillsObjectTypeID,
					ValueFrom: interfaces.CondValueFromConst,
				},
				{
					Field:     "target_object_type_id",
					Operation: interfaces.KnOperationTypeEqual,
					Value:     rc.config.SkillsObjectTypeID,
					ValueFrom: interfaces.CondValueFromConst,
				},
			},
		},
	}
	result, err := rc.bknBackend.SearchRelationTypes(ctx, query)
	if err != nil {
		return false, err
	}
	return len(result.Entries) > 0, nil
}

// findRelationType looks up the RelationType between objectTypeID and skills.
// Returns nil (no error) when no relation exists.
func (rc *recallCoordinator) findRelationType(ctx context.Context, knID, objectTypeID string) (*interfaces.RelationType, error) {
	query := &interfaces.QueryConceptsReq{
		KnID:  knID,
		Limit: 1,
		Cond: &interfaces.KnCondition{
			Operation: interfaces.KnOperationTypeOr,
			SubConditions: []*interfaces.KnCondition{
				{
					Operation: interfaces.KnOperationTypeAnd,
					SubConditions: []*interfaces.KnCondition{
						{
							Field:     "source_object_type_id",
							Operation: interfaces.KnOperationTypeEqual,
							Value:     objectTypeID,
							ValueFrom: interfaces.CondValueFromConst,
						},
						{
							Field:     "target_object_type_id",
							Operation: interfaces.KnOperationTypeEqual,
							Value:     rc.config.SkillsObjectTypeID,
							ValueFrom: interfaces.CondValueFromConst,
						},
					},
				},
				{
					Operation: interfaces.KnOperationTypeAnd,
					SubConditions: []*interfaces.KnCondition{
						{
							Field:     "source_object_type_id",
							Operation: interfaces.KnOperationTypeEqual,
							Value:     rc.config.SkillsObjectTypeID,
							ValueFrom: interfaces.CondValueFromConst,
						},
						{
							Field:     "target_object_type_id",
							Operation: interfaces.KnOperationTypeEqual,
							Value:     objectTypeID,
							ValueFrom: interfaces.CondValueFromConst,
						},
					},
				},
			},
		},
	}
	result, err := rc.bknBackend.SearchRelationTypes(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SearchRelationTypes(%s<->skills): %w", objectTypeID, err)
	}
	if len(result.Entries) == 0 {
		return nil, nil
	}
	return result.Entries[0], nil
}

// buildInstanceFilterCondition converts instance_identities to a KnCondition.
// Each identity map becomes an AND group; multiple identities are OR-combined.
func buildInstanceFilterCondition(identities []map[string]interface{}) *interfaces.KnCondition {
	if len(identities) == 0 {
		return nil
	}

	var orSubs []*interfaces.KnCondition
	for _, identity := range identities {
		if len(identity) == 0 {
			continue
		}
		var andSubs []*interfaces.KnCondition
		for k, v := range identity {
			andSubs = append(andSubs, &interfaces.KnCondition{
				Field:     k,
				Operation: interfaces.KnOperationTypeEqual,
				Value:     v,
				ValueFrom: interfaces.CondValueFromConst,
			})
		}
		if len(andSubs) == 1 {
			orSubs = append(orSubs, andSubs[0])
		} else if len(andSubs) > 1 {
			orSubs = append(orSubs, &interfaces.KnCondition{
				Operation:     interfaces.KnOperationTypeAnd,
				SubConditions: andSubs,
			})
		}
	}

	if len(orSubs) == 0 {
		return nil
	}
	if len(orSubs) == 1 {
		return orSubs[0]
	}
	return &interfaces.KnCondition{
		Operation:     interfaces.KnOperationTypeOr,
		SubConditions: orSubs,
	}
}

// extractSkillMatchesFromInstances extracts SkillMatch from QueryObjectInstancesResp data.
func extractSkillMatchesFromInstances(data []any, scope string, priority int) []interfaces.SkillMatch {
	var matches []interfaces.SkillMatch
	for _, item := range data {
		dataMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := interfaces.SkillMatch{
			SkillID:      stringFromMap(dataMap, "skill_id"),
			Name:         stringFromMap(dataMap, "name"),
			Description:  stringFromMap(dataMap, "description"),
			MatchedScope: scope,
			Priority:     priority,
			Score:        float64FromMap(dataMap, "_score"),
		}
		if m.SkillID == "" {
			continue
		}
		matches = append(matches, m)
	}
	return matches
}

// extractSkillMatchesFromSubgraph extracts SkillMatch from QueryInstanceSubgraphResp.
//
// Response structure per API spec (PathEntries):
//
//	{
//	  "entries": [                          // ObjectSubGraphResponse[]
//	    {
//	      "objects": {                      // map[objectID -> ObjectInfoInSubgraph]
//	        "skills-skill_review": {
//	          "id":               "skills-skill_review",
//	          "object_type_id":   "skills",
//	          "object_type_name": "skills",
//	          "properties": {
//	            "skill_id":    "skill_review",
//	            "name":        "合同审查",
//	            "description": "..."
//	          }
//	        }
//	      },
//	      "relation_paths": [...],
//	      "total_count": 1
//	    }
//	  ]
//	}
func extractSkillMatchesFromSubgraph(resp *interfaces.QueryInstanceSubgraphResp, skillsOTID, scope string, priority int) []interfaces.SkillMatch {
	if resp == nil || resp.Entries == nil {
		return nil
	}

	var matches []interfaces.SkillMatch

	entries, ok := resp.Entries.([]interface{})
	if !ok {
		return nil
	}

	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		matches = append(matches, extractFromSubgraphEntry(entryMap, skillsOTID, scope, priority)...)
	}

	return matches
}

// extractFromSubgraphEntry extracts skills from a single ObjectSubGraphResponse entry.
// It iterates the "objects" map and filters by object_type_id == skillsOTID,
// then reads skill metadata from the "properties" sub-map.
func extractFromSubgraphEntry(entry map[string]interface{}, skillsOTID, scope string, priority int) []interfaces.SkillMatch {
	var matches []interfaces.SkillMatch

	objectsRaw, ok := entry["objects"]
	if !ok {
		return nil
	}
	objectsMap, ok := objectsRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	for _, objRaw := range objectsMap {
		objMap, ok := objRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if stringFromMap(objMap, "object_type_id") != skillsOTID {
			continue
		}

		props := mapFromMap(objMap, "properties")

		skillID := stringFromMap(props, "skill_id")
		if skillID == "" {
			continue
		}

		m := interfaces.SkillMatch{
			SkillID:      skillID,
			Name:         stringFromMap(props, "name"),
			Description:  stringFromMap(props, "description"),
			MatchedScope: scope,
			Priority:     priority,
			Score:        float64FromMap(props, "_score"),
		}
		if m.Name == "" {
			m.Name = stringFromMap(objMap, "display")
		}

		matches = append(matches, m)
	}

	return matches
}

func mapFromMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if sub, ok := v.(map[string]interface{}); ok {
			return sub
		}
	}
	return m
}

func stringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func float64FromMap(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}
