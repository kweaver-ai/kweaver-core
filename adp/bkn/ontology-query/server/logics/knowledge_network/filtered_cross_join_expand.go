// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knowledge_network

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
)

const defaultFilteredCrossJoinMaxExpand = 10000

func (kns *knowledgeNetworkService) filteredCrossJoinMaxExpand() int {
	n := kns.appSetting.ServerSetting.FilteredCrossJoinMaxEdgeExpand
	if n <= 0 {
		return defaultFilteredCrossJoinMaxExpand
	}
	return n
}

// expandFilteredCrossJoin builds next-hop objects for relation type filtered_cross_join:
// Current-side batch is filtered in memory with EvaluateInstanceAgainstCondition (avoids second object-store query
// and keeps relation-side semantics aligned with app-layer condition evaluation).
// Next-side instances: relation mapping condition AND path-node ActualCondition, via GetObjectsByObjectTypeID.
// Pairs are emitted in stable order (current ObjectID asc, then next row order) until maxEdgeExpand (silent truncate).
func (kns *knowledgeNetworkService) expandFilteredCrossJoin(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource,
	batch []interfaces.LevelObject,
	edge *interfaces.TypeEdge,
	nextTypeMeta interfaces.ObjectTypeWithKeyField,
	isForward bool,
	rules interfaces.FilteredCrossJoinMapping,
) (map[string]interfaces.Objects, error) {

	maxQ := kns.filteredCrossJoinMaxExpand()
	var filterCurrent *cond.CondCfg
	var queryNext *cond.CondCfg
	if isForward {
		filterCurrent = rules.SourceCondition
		queryNext = rules.TargetCondition
	} else {
		filterCurrent = rules.TargetCondition
		queryNext = rules.SourceCondition
	}

	filtered := make([]interfaces.LevelObject, 0, len(batch))
	for _, lo := range batch {
		if lo.ObjectType == nil {
			continue
		}
		ok, err := logics.EvaluateInstanceAgainstCondition(ctx, lo.ObjectData, filterCurrent, lo.ObjectType)
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("filtered_cross_join evaluate current instance: %s", err.Error()))
		}
		if ok {
			filtered = append(filtered, lo)
		}
	}
	if len(filtered) == 0 {
		return nil, nil
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ObjectID < filtered[j].ObjectID
	})

	var nextObjectTypeID string
	if isForward {
		nextObjectTypeID = edge.RelationType.TargetObjectTypeID
	} else {
		nextObjectTypeID = edge.RelationType.SourceObjectTypeID
	}

	nextObjectQuery := &interfaces.ObjectQueryBaseOnObjectType{
		KNID:         query.KNID,
		Branch:       query.Branch,
		ObjectTypeID: nextObjectTypeID,
		PageQuery: interfaces.PageQuery{
			NeedTotal: false,
			Limit:     maxQ,
		},
		CommonQueryParameters: interfaces.CommonQueryParameters{
			IncludeTypeInfo:    true,
			IncludeLogicParams: query.IncludeLogicParams,
			IgnoringStore:      query.IgnoringStore,
		},
	}

	// Next hop: FCJ mapping condition AND path object-type condition (e.g. A→B→C with filter on C).
	nextObjectQuery.ActualCondition = queryNext
	if nextTypeMeta.ActualCondition != nil {
		nextObjectQuery.ActualCondition = &cond.CondCfg{
			Operation: cond.OperationAnd,
			SubConds:  []*cond.CondCfg{nextObjectQuery.ActualCondition, nextTypeMeta.ActualCondition},
		}
	}

	ptrPropMap := logics.TransferPropsToPropMap(nextTypeMeta.DataProperties)
	propMap := make(map[string]cond.DataProperty, len(ptrPropMap))
	for k, v := range ptrPropMap {
		if v != nil {
			propMap[k] = *v
		}
	}
	fullNextOT := interfaces.ObjectType{ObjectTypeWithKeyField: nextTypeMeta}
	nextObjectQuery.Sort = logics.BuildIndexSort(fullNextOT, propMap)

	nextObjects, err := kns.ots.GetObjectsByObjectTypeID(ctx, nextObjectQuery)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interfaces.Objects)
	pairCount := 0
	for _, s := range filtered {
		if pairCount >= maxQ {
			break
		}
		for _, trow := range nextObjects.Datas {
			if pairCount >= maxQ {
				break
			}
			objs := result[s.ObjectID]
			if objs.Datas == nil {
				objs = interfaces.Objects{
					Datas:      []map[string]any{},
					ObjectType: nextObjects.ObjectType,
					TotalCount: 0,
				}
			}
			objs.Datas = append(objs.Datas, trow)
			objs.TotalCount++
			result[s.ObjectID] = objs
			pairCount++
		}
	}

	return result, nil
}

// matchFilteredCrossJoinRelations pairs source and target instances using filtered_cross_join rules with the same
// ordering and silent truncation as expandFilteredCrossJoin (for SearchSubgraphByObjects).
func (kns *knowledgeNetworkService) matchFilteredCrossJoinRelations(ctx context.Context,
	sourceObjects []interfaces.LevelObject,
	targetObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge,
) ([]interfaces.Relation, error) {

	rules, ok := edge.RelationType.MappingRules.(interfaces.FilteredCrossJoinMapping)
	if !ok {
		return nil, nil
	}

	maxQ := kns.filteredCrossJoinMaxExpand()
	src := filterAndSortLevelObjectsByCond(ctx, sourceObjects, rules.SourceCondition)
	tgt := filterAndSortLevelObjectsByCond(ctx, targetObjects, rules.TargetCondition)
	if len(src) == 0 || len(tgt) == 0 {
		return nil, nil
	}

	relations := make([]interfaces.Relation, 0)
	pairCount := 0
	for _, s := range src {
		if pairCount >= maxQ {
			break
		}
		for _, t := range tgt {
			if pairCount >= maxQ {
				break
			}
			relations = append(relations, interfaces.Relation{
				RelationTypeId:   edge.RelationTypeId,
				RelationTypeName: edge.RelationType.RTName,
				SourceObjectId:   s.ObjectID,
				TargetObjectId:   t.ObjectID,
			})
			pairCount++
		}
	}
	return relations, nil
}

func filterAndSortLevelObjectsByCond(ctx context.Context, objs []interfaces.LevelObject, c *cond.CondCfg) []interfaces.LevelObject {
	out := make([]interfaces.LevelObject, 0, len(objs))
	for _, lo := range objs {
		if lo.ObjectType == nil {
			continue
		}
		ok, err := logics.EvaluateInstanceAgainstCondition(ctx, lo.ObjectData, c, lo.ObjectType)
		if err != nil || !ok {
			continue
		}
		out = append(out, lo)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ObjectID < out[j].ObjectID })
	return out
}
