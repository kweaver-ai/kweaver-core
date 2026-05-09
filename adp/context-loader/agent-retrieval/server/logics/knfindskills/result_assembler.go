// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"sort"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// Assemble deduplicates by skill_id (keeping the highest priority),
// sorts by Priority desc -> Score desc, and trims to topK.
func Assemble(matches []interfaces.SkillMatch, topK int) *interfaces.FindSkillsResp {
	if len(matches) == 0 {
		return &interfaces.FindSkillsResp{Entries: []*interfaces.SkillItem{}}
	}

	deduped := dedup(matches)

	sort.SliceStable(deduped, func(i, j int) bool {
		if deduped[i].Priority != deduped[j].Priority {
			return deduped[i].Priority > deduped[j].Priority
		}
		if deduped[i].Score != deduped[j].Score {
			return deduped[i].Score > deduped[j].Score
		}
		return deduped[i].SkillID < deduped[j].SkillID
	})

	if topK > 0 && len(deduped) > topK {
		deduped = deduped[:topK]
	}

	entries := make([]*interfaces.SkillItem, 0, len(deduped))
	for i := range deduped {
		entries = append(entries, &interfaces.SkillItem{
			SkillID:     deduped[i].SkillID,
			Name:        deduped[i].Name,
			Description: deduped[i].Description,
		})
	}

	return &interfaces.FindSkillsResp{Entries: entries}
}

func dedup(matches []interfaces.SkillMatch) []interfaces.SkillMatch {
	best := make(map[string]*interfaces.SkillMatch, len(matches))
	for i := range matches {
		m := &matches[i]
		if m.SkillID == "" {
			continue
		}
		existing, ok := best[m.SkillID]
		if !ok || m.Priority > existing.Priority ||
			(m.Priority == existing.Priority && m.Score > existing.Score) {
			best[m.SkillID] = m
		}
	}

	keys := make([]string, 0, len(best))
	for k := range best {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]interfaces.SkillMatch, 0, len(best))
	for _, k := range keys {
		out = append(out, *best[k])
	}
	return out
}
