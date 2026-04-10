// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

const skillsObjectTypeIDDefault = "skills"

// BuildSubgraphRequest constructs a QueryInstanceSubgraphReq for skills recall
// along the relation path between objectTypeID and the skills object type.
// The direction is auto-detected from relationType.SourceObjectTypeID / TargetObjectTypeID.
func BuildSubgraphRequest(
	knID string,
	objectTypeID string,
	relationType *interfaces.RelationType,
	instanceCondition *interfaces.KnCondition,
	skillQueryCondition *interfaces.KnCondition,
	topK int,
	skillsOTID string,
) *interfaces.QueryInstanceSubgraphReq {
	if skillsOTID == "" {
		skillsOTID = skillsObjectTypeIDDefault
	}

	isForward := relationType.SourceObjectTypeID == objectTypeID &&
		relationType.TargetObjectTypeID == skillsOTID

	var sourceOTID, targetOTID string
	if isForward {
		sourceOTID = objectTypeID
		targetOTID = skillsOTID
	} else {
		sourceOTID = skillsOTID
		targetOTID = objectTypeID
	}

	startObjType := buildObjectTypeOnPath(objectTypeID, instanceCondition, 0)
	skillsObjType := buildObjectTypeOnPath(skillsOTID, skillQueryCondition, topK)

	var objectTypes []map[string]interface{}
	if isForward {
		objectTypes = []map[string]interface{}{startObjType, skillsObjType}
	} else {
		objectTypes = []map[string]interface{}{skillsObjType, startObjType}
	}

	typeEdge := map[string]interface{}{
		"relation_type_id":       relationType.ID,
		"source_object_type_id":  sourceOTID,
		"target_object_type_id":  targetOTID,
	}

	path := map[string]interface{}{
		"object_types":   objectTypes,
		"relation_types": []map[string]interface{}{typeEdge},
	}

	return &interfaces.QueryInstanceSubgraphReq{
		KnID:              knID,
		RelationTypePaths: []map[string]interface{}{path},
	}
}

func buildObjectTypeOnPath(otID string, condition *interfaces.KnCondition, limit int) map[string]interface{} {
	m := map[string]interface{}{
		"id": otID,
	}
	if condition != nil {
		m["condition"] = condition
	}
	if limit > 0 {
		m["limit"] = limit
	}
	return m
}
