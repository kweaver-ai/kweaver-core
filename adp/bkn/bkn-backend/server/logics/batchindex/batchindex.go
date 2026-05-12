// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package batchindex builds request-scoped BatchIDIndex values from KN / concept group payloads.
package batchindex

import (
	"encoding/json"
	"fmt"

	"bkn-backend/interfaces"
)

// NewBatchIDIndex constructs an empty index; empty branch defaults to MAIN_BRANCH.
func NewBatchIDIndex(knID, branch string) *interfaces.BatchIDIndex {
	if branch == "" {
		branch = interfaces.MAIN_BRANCH
	}
	return &interfaces.BatchIDIndex{
		KNID:            knID,
		Branch:          branch,
		ObjectTypes:     make(map[string]*interfaces.ObjectType),
		RelationTypeIDs: make(map[string]struct{}),
		ActionTypeIDs:   make(map[string]struct{}),
		RiskTypeIDs:     make(map[string]struct{}),
		ConceptGroupIDs: make(map[string]struct{}),
		Metrics:         make(map[string]*interfaces.MetricDefinition),
	}
}

// EnsureObjectTypePropertyMap fills PropertyMap from DataProperties when unset, for mapping rule validation.
func EnsureObjectTypePropertyMap(ot *interfaces.ObjectType) {
	if ot == nil {
		return
	}
	if len(ot.PropertyMap) > 0 {
		return
	}
	if len(ot.DataProperties) == 0 {
		return
	}
	ot.PropertyMap = make(map[string]string, len(ot.DataProperties))
	for _, prop := range ot.DataProperties {
		if prop == nil {
			continue
		}
		ot.PropertyMap[prop.Name] = prop.DisplayName
	}
}

// MergeBatchIndex merges src into dst (mutates dst). Returns an error when the same object type ID has conflicting payloads.
func MergeBatchIndex(dst, src *interfaces.BatchIDIndex) error {
	if src == nil || dst == nil {
		return nil
	}
	for id, ot := range src.ObjectTypes {
		if err := addObjectType(dst, id, ot); err != nil {
			return err
		}
	}
	for id := range src.RelationTypeIDs {
		dst.RelationTypeIDs[id] = struct{}{}
	}
	for id := range src.ActionTypeIDs {
		dst.ActionTypeIDs[id] = struct{}{}
	}
	for id := range src.RiskTypeIDs {
		dst.RiskTypeIDs[id] = struct{}{}
	}
	for id := range src.ConceptGroupIDs {
		dst.ConceptGroupIDs[id] = struct{}{}
	}
	for id, m := range src.Metrics {
		if err := addMetric(dst, id, m); err != nil {
			return err
		}
	}
	return nil
}

// CollectKNFromPayload collects a BatchIDIndex from a full KN (including nested concept group buckets).
func CollectKNFromPayload(kn *interfaces.KN) (*interfaces.BatchIDIndex, error) {
	if kn == nil {
		return NewBatchIDIndex("", interfaces.MAIN_BRANCH), nil
	}
	branch := kn.Branch
	if branch == "" {
		branch = interfaces.MAIN_BRANCH
	}
	idx := NewBatchIDIndex(kn.KNID, branch)
	for _, ot := range kn.ObjectTypes {
		if err := ingestObjectType(idx, ot); err != nil {
			return nil, err
		}
	}
	for _, rt := range kn.RelationTypes {
		ingestRelationType(idx, rt)
	}
	for _, at := range kn.ActionTypes {
		ingestActionType(idx, at)
	}
	for _, rt := range kn.RiskTypes {
		ingestRiskType(idx, rt)
	}
	for _, cg := range kn.ConceptGroups {
		if err := ingestConceptGroup(idx, cg); err != nil {
			return nil, err
		}
	}
	for _, m := range kn.Metrics {
		if err := ingestMetric(idx, m); err != nil {
			return nil, err
		}
	}
	return idx, nil
}

// CollectFromConceptGroups collects an index from concept groups only (may nest object/relation/action).
func CollectFromConceptGroups(knID, branch string, groups []*interfaces.ConceptGroup) (*interfaces.BatchIDIndex, error) {
	if branch == "" {
		branch = interfaces.MAIN_BRANCH
	}
	idx := NewBatchIDIndex(knID, branch)
	for _, cg := range groups {
		if err := ingestConceptGroup(idx, cg); err != nil {
			return nil, err
		}
	}
	return idx, nil
}

func ingestConceptGroup(b *interfaces.BatchIDIndex, cg *interfaces.ConceptGroup) error {
	if cg == nil {
		return nil
	}
	if cg.CGID != "" {
		b.ConceptGroupIDs[cg.CGID] = struct{}{}
	}
	for _, ot := range cg.ObjectTypes {
		if err := ingestObjectType(b, ot); err != nil {
			return err
		}
	}
	for _, rt := range cg.RelationTypes {
		ingestRelationType(b, rt)
	}
	for _, at := range cg.ActionTypes {
		ingestActionType(b, at)
	}
	return nil
}

func ingestObjectType(b *interfaces.BatchIDIndex, ot *interfaces.ObjectType) error {
	if ot == nil {
		return nil
	}
	if err := addObjectType(b, ot.OTID, ot); err != nil {
		return err
	}
	for _, cg := range ot.ConceptGroups {
		if cg != nil && cg.CGID != "" {
			b.ConceptGroupIDs[cg.CGID] = struct{}{}
		}
	}
	return nil
}

func ingestRelationType(b *interfaces.BatchIDIndex, rt *interfaces.RelationType) {
	if rt == nil || rt.RTID == "" {
		return
	}
	b.RelationTypeIDs[rt.RTID] = struct{}{}
}

func ingestActionType(b *interfaces.BatchIDIndex, at *interfaces.ActionType) {
	if at == nil || at.ATID == "" {
		return
	}
	b.ActionTypeIDs[at.ATID] = struct{}{}
}

func ingestMetric(b *interfaces.BatchIDIndex, m *interfaces.MetricDefinition) error {
	if m == nil {
		return nil
	}
	return addMetric(b, m.ID, m)
}

func addMetric(b *interfaces.BatchIDIndex, id string, m *interfaces.MetricDefinition) error {
	if id == "" || m == nil {
		return nil
	}
	if existing, ok := b.Metrics[id]; ok {
		if !metricBatchPayloadEqual(existing, m) {
			return fmt.Errorf("conflicting metric definitions for id %q", id)
		}
		return nil
	}
	b.Metrics[id] = m
	return nil
}

func metricBatchPayloadEqual(a, b *interfaces.MetricDefinition) bool {
	if a == nil || b == nil {
		return a == b
	}
	aj, e1 := json.Marshal(a)
	bj, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && string(aj) == string(bj)
}

func ingestRiskType(b *interfaces.BatchIDIndex, rt *interfaces.RiskType) {
	if rt == nil || rt.RTID == "" {
		return
	}
	b.RiskTypeIDs[rt.RTID] = struct{}{}
}

func addObjectType(b *interfaces.BatchIDIndex, id string, ot *interfaces.ObjectType) error {
	if id == "" || ot == nil {
		return nil
	}
	if existing, ok := b.ObjectTypes[id]; ok {
		if !objectTypeBatchPayloadEqual(existing, ot) {
			return fmt.Errorf("conflicting object type definitions for id %q", id)
		}
		return nil
	}
	b.ObjectTypes[id] = ot
	return nil
}

func objectTypeBatchPayloadEqual(a, b *interfaces.ObjectType) bool {
	if a == nil || b == nil {
		return a == b
	}
	aj, e1 := json.Marshal(a.ObjectTypeWithKeyField)
	bj, e2 := json.Marshal(b.ObjectTypeWithKeyField)
	return e1 == nil && e2 == nil && string(aj) == string(bj)
}

// HasObjectTypeID reports whether the batch declares the object type ID (including nested payloads).
func HasObjectTypeID(id string, b *interfaces.BatchIDIndex) bool {
	if b == nil || id == "" {
		return false
	}
	_, ok := b.ObjectTypes[id]
	return ok
}

// HasConceptGroupID reports whether the batch declares the concept group ID.
func HasConceptGroupID(id string, b *interfaces.BatchIDIndex) bool {
	if b == nil || id == "" {
		return false
	}
	_, ok := b.ConceptGroupIDs[id]
	return ok
}

// HasRiskTypeID reports whether the batch declares the risk type ID.
func HasRiskTypeID(id string, b *interfaces.BatchIDIndex) bool {
	if b == nil || id == "" {
		return false
	}
	_, ok := b.RiskTypeIDs[id]
	return ok
}
