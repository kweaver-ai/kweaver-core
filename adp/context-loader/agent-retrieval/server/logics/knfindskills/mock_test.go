// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// testLogger is a minimal logger for tests
type testLogger struct{}

func (l *testLogger) WithContext(ctx context.Context) interfaces.Logger { return l }
func (l *testLogger) Info(v ...interface{})                             {}
func (l *testLogger) Debug(v ...interface{})                            {}
func (l *testLogger) Warn(v ...interface{})                             {}
func (l *testLogger) Error(v ...interface{})                            {}
func (l *testLogger) Infof(format string, args ...interface{})          {}
func (l *testLogger) Debugf(format string, args ...interface{})         {}
func (l *testLogger) Warnf(format string, args ...interface{})          {}
func (l *testLogger) Errorf(format string, args ...interface{})         {}

// testBknBackend is a configurable mock for BknBackendAccess
type testBknBackend struct {
	searchRelationTypesFunc func(ctx context.Context, query *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error)
	getObjectTypeDetailFunc func(ctx context.Context, knID string, otIds []string, includeDetail bool) ([]*interfaces.ObjectType, error)
}

func (m *testBknBackend) SearchRelationTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
	if m.searchRelationTypesFunc != nil {
		return m.searchRelationTypesFunc(ctx, query)
	}
	return &interfaces.RelationTypeConcepts{}, nil
}

func (m *testBknBackend) GetObjectTypeDetail(ctx context.Context, knID string, otIds []string, includeDetail bool) ([]*interfaces.ObjectType, error) {
	if m.getObjectTypeDetailFunc != nil {
		return m.getObjectTypeDetailFunc(ctx, knID, otIds, includeDetail)
	}
	return nil, nil
}

func (m *testBknBackend) GetKnowledgeNetworkDetail(ctx context.Context, knID string) (*interfaces.KnowledgeNetworkDetail, error) {
	return nil, nil
}
func (m *testBknBackend) SearchObjectTypes(ctx context.Context, req *interfaces.QueryConceptsReq) (*interfaces.ObjectTypeConcepts, error) {
	return nil, nil
}
func (m *testBknBackend) GetRelationTypeDetail(ctx context.Context, knID string, rtIDs []string, includeDetail bool) ([]*interfaces.RelationType, error) {
	return nil, nil
}
func (m *testBknBackend) SearchActionTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (*interfaces.ActionTypeConcepts, error) {
	return nil, nil
}
func (m *testBknBackend) GetActionTypeDetail(ctx context.Context, knID string, atIDs []string, includeDetail bool) ([]*interfaces.ActionType, error) {
	return nil, nil
}
func (m *testBknBackend) CreateFullBuildOntologyJob(ctx context.Context, knID string, req *interfaces.CreateFullBuildOntologyJobReq) (*interfaces.CreateJobResp, error) {
	return nil, nil
}
func (m *testBknBackend) ListOntologyJobs(ctx context.Context, knID string, req *interfaces.ListOntologyJobsReq) (*interfaces.ListOntologyJobsResp, error) {
	return nil, nil
}

// testOntologyQuery is a configurable mock for DrivenOntologyQuery
type testOntologyQuery struct {
	queryObjectInstancesFunc   func(ctx context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error)
	queryInstanceSubgraphFunc  func(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error)
	subgraphCallCount          int
}

func (m *testOntologyQuery) QueryObjectInstances(ctx context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
	if m.queryObjectInstancesFunc != nil {
		return m.queryObjectInstancesFunc(ctx, req)
	}
	return &interfaces.QueryObjectInstancesResp{}, nil
}

func (m *testOntologyQuery) QueryInstanceSubgraph(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
	m.subgraphCallCount++
	if m.queryInstanceSubgraphFunc != nil {
		return m.queryInstanceSubgraphFunc(ctx, req)
	}
	return &interfaces.QueryInstanceSubgraphResp{}, nil
}

func (m *testOntologyQuery) QueryLogicProperties(ctx context.Context, req *interfaces.QueryLogicPropertiesReq) (*interfaces.QueryLogicPropertiesResp, error) {
	return nil, nil
}

func (m *testOntologyQuery) QueryActions(ctx context.Context, req *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
	return nil, nil
}

func makeSkillInstances(count int) []any {
	data := make([]any, count)
	for i := 0; i < count; i++ {
		data[i] = map[string]any{
			"skill_id":    fmt.Sprintf("skill_%d", i),
			"name":        fmt.Sprintf("技能_%d", i),
			"description": fmt.Sprintf("描述_%d", i),
			"_score":      0.9 - float64(i)*0.1,
		}
	}
	return data
}

// makeSubgraphEntries builds a mock subgraph response matching the real API structure:
//
//	PathEntries.entries -> ObjectSubGraphResponse[] ->
//	  objects: map[objectID -> ObjectInfoInSubgraph{object_type_id, properties}]
func makeSubgraphEntries(objectTypeID string, skillProps ...map[string]interface{}) []interface{} {
	objects := map[string]interface{}{}
	for i, props := range skillProps {
		objID := fmt.Sprintf("%s-obj-%d", objectTypeID, i)
		if sid, ok := props["skill_id"]; ok {
			objID = fmt.Sprintf("%s-%v", objectTypeID, sid)
		}
		objects[objID] = map[string]interface{}{
			"id":               objID,
			"object_type_id":   objectTypeID,
			"object_type_name": objectTypeID,
			"display":          props["name"],
			"properties":       props,
		}
	}
	return []interface{}{
		map[string]interface{}{
			"objects":        objects,
			"relation_paths": []interface{}{},
			"total_count":    len(skillProps),
		},
	}
}
