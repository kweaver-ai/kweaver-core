// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func defaultFSConfig() *config.FindSkillsConfig {
	return &config.FindSkillsConfig{
		DefaultTopK:        10,
		MaxTopK:            20,
		SkillsObjectTypeID: "skills",
	}
}

func TestNormalizeAndDetectMode_NetworkMode(t *testing.T) {
	req := &interfaces.FindSkillsReq{KnID: "kn1"}
	mode, err := NormalizeAndDetectMode(req, defaultFSConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != interfaces.RecallModeNetwork {
		t.Errorf("expected RecallModeNetwork(1), got %d", mode)
	}
	if req.TopK != 10 {
		t.Errorf("expected default TopK=10, got %d", req.TopK)
	}
}

func TestNormalizeAndDetectMode_ObjectTypeMode(t *testing.T) {
	req := &interfaces.FindSkillsReq{KnID: "kn1", ObjectTypeID: "contract"}
	mode, err := NormalizeAndDetectMode(req, defaultFSConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != interfaces.RecallModeObjectType {
		t.Errorf("expected RecallModeObjectType(2), got %d", mode)
	}
}

func TestNormalizeAndDetectMode_InstanceMode(t *testing.T) {
	req := &interfaces.FindSkillsReq{
		KnID:               "kn1",
		ObjectTypeID:       "contract",
		InstanceIdentities: []map[string]interface{}{{"id": "C-001"}},
	}
	mode, err := NormalizeAndDetectMode(req, defaultFSConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != interfaces.RecallModeInstance {
		t.Errorf("expected RecallModeInstance(3), got %d", mode)
	}
}

func TestNormalizeAndDetectMode_InstanceWithoutObjectType(t *testing.T) {
	req := &interfaces.FindSkillsReq{
		KnID:               "kn1",
		InstanceIdentities: []map[string]interface{}{{"id": "C-001"}},
	}
	_, err := NormalizeAndDetectMode(req, defaultFSConfig())
	if err == nil {
		t.Fatal("expected error when instance_identities present but object_type_id missing")
	}
}

func TestNormalizeAndDetectMode_TopKClamping(t *testing.T) {
	cfg := defaultFSConfig()

	// TopK=0 -> default
	req := &interfaces.FindSkillsReq{KnID: "kn1", TopK: 0}
	_, _ = NormalizeAndDetectMode(req, cfg)
	if req.TopK != 10 {
		t.Errorf("expected TopK=10, got %d", req.TopK)
	}

	// TopK > MaxTopK -> clamped
	req = &interfaces.FindSkillsReq{KnID: "kn1", TopK: 50}
	_, _ = NormalizeAndDetectMode(req, cfg)
	if req.TopK != 20 {
		t.Errorf("expected TopK=20, got %d", req.TopK)
	}
}
