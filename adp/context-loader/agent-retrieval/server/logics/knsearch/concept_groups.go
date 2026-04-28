// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import "strings"

// normalizeConceptGroups trims whitespace, drops empty IDs, and de-duplicates the
// caller-provided concept group list while preserving its first-seen order. It is
// the single source of truth for concept-group normalization across SearchSchema,
// KnSearch and direct BKN search request builders.
//
// concept_groups is the only SearchSchema knob that actually constrains BKN-side
// recall (object/relation/action/metric types). When the caller provides a
// non-empty list, ContextLoader skips local in-memory filtering of a full
// network export and instead delegates filtering to BKN's typed search APIs by
// passing the normalized list through QueryConceptsReq.ConceptGroups.
func normalizeConceptGroups(groups []string) []string {
	if len(groups) == 0 {
		return nil
	}
	out := make([]string, 0, len(groups))
	seen := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		out = append(out, group)
	}
	return out
}
