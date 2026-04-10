// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func TestAssemble_Empty(t *testing.T) {
	resp := Assemble(nil, 10)
	if len(resp.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(resp.Entries))
	}
}

func TestAssemble_DedupKeepsHigherPriority(t *testing.T) {
	matches := []interfaces.SkillMatch{
		{SkillID: "s1", Name: "A", Priority: 50, Score: 0.8, MatchedScope: "object_type"},
		{SkillID: "s1", Name: "A", Priority: 100, Score: 0.6, MatchedScope: "object_selector"},
	}
	resp := Assemble(matches, 10)
	if len(resp.Entries) != 1 {
		t.Fatalf("expected 1 entry after dedup, got %d", len(resp.Entries))
	}
	if resp.Entries[0].SkillID != "s1" {
		t.Errorf("expected s1, got %s", resp.Entries[0].SkillID)
	}
}

func TestAssemble_SortByPriorityThenScore(t *testing.T) {
	matches := []interfaces.SkillMatch{
		{SkillID: "s1", Name: "A", Priority: 50, Score: 0.9},
		{SkillID: "s2", Name: "B", Priority: 100, Score: 0.5},
		{SkillID: "s3", Name: "C", Priority: 100, Score: 0.8},
	}
	resp := Assemble(matches, 10)
	if len(resp.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(resp.Entries))
	}
	// Priority 100 first, higher score first within same priority
	if resp.Entries[0].SkillID != "s3" {
		t.Errorf("expected s3 first (pri=100, score=0.8), got %s", resp.Entries[0].SkillID)
	}
	if resp.Entries[1].SkillID != "s2" {
		t.Errorf("expected s2 second (pri=100, score=0.5), got %s", resp.Entries[1].SkillID)
	}
	if resp.Entries[2].SkillID != "s1" {
		t.Errorf("expected s1 third (pri=50), got %s", resp.Entries[2].SkillID)
	}
}

func TestAssemble_TopKTrim(t *testing.T) {
	matches := []interfaces.SkillMatch{
		{SkillID: "s1", Name: "A", Priority: 100, Score: 0.9},
		{SkillID: "s2", Name: "B", Priority: 90, Score: 0.8},
		{SkillID: "s3", Name: "C", Priority: 80, Score: 0.7},
	}
	resp := Assemble(matches, 2)
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries after topK trim, got %d", len(resp.Entries))
	}
}

func TestAssemble_StableSortSamePriorityAndScore(t *testing.T) {
	matches := []interfaces.SkillMatch{
		{SkillID: "s_charlie", Name: "C", Priority: 50, Score: 0.5},
		{SkillID: "s_alpha", Name: "A", Priority: 50, Score: 0.5},
		{SkillID: "s_bravo", Name: "B", Priority: 50, Score: 0.5},
	}
	for run := 0; run < 20; run++ {
		resp := Assemble(matches, 10)
		if len(resp.Entries) != 3 {
			t.Fatalf("run %d: expected 3 entries, got %d", run, len(resp.Entries))
		}
		if resp.Entries[0].SkillID != "s_alpha" {
			t.Errorf("run %d: expected s_alpha first, got %s", run, resp.Entries[0].SkillID)
		}
		if resp.Entries[1].SkillID != "s_bravo" {
			t.Errorf("run %d: expected s_bravo second, got %s", run, resp.Entries[1].SkillID)
		}
		if resp.Entries[2].SkillID != "s_charlie" {
			t.Errorf("run %d: expected s_charlie third, got %s", run, resp.Entries[2].SkillID)
		}
	}
}

func TestAssemble_SkipEmptySkillID(t *testing.T) {
	matches := []interfaces.SkillMatch{
		{SkillID: "", Name: "no-id", Priority: 100},
		{SkillID: "s1", Name: "valid", Priority: 50},
	}
	resp := Assemble(matches, 10)
	if len(resp.Entries) != 1 {
		t.Fatalf("expected 1 entry (skipping empty skill_id), got %d", len(resp.Entries))
	}
	if resp.Entries[0].SkillID != "s1" {
		t.Errorf("expected s1, got %s", resp.Entries[0].SkillID)
	}
}
