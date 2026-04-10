// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// NormalizeAndDetectMode validates the request and determines the recall mode.
// Returns an error (400) if instance_identities is present but object_type_id is missing.
func NormalizeAndDetectMode(req *interfaces.FindSkillsReq, cfg *config.FindSkillsConfig) (interfaces.RecallMode, error) {
	if len(req.InstanceIdentities) > 0 && req.ObjectTypeID == "" {
		return 0, fmt.Errorf("%d:instance_identities requires object_type_id", http.StatusBadRequest)
	}

	if req.TopK <= 0 {
		req.TopK = cfg.DefaultTopK
	}
	if req.TopK > cfg.MaxTopK {
		req.TopK = cfg.MaxTopK
	}

	if req.ObjectTypeID != "" && len(req.InstanceIdentities) > 0 {
		return interfaces.RecallModeInstance, nil
	}
	if req.ObjectTypeID != "" {
		return interfaces.RecallModeObjectType, nil
	}
	return interfaces.RecallModeNetwork, nil
}
