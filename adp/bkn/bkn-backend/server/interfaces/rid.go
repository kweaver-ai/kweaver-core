// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "sort"

// RidKind defines valid resource kinds for rid properties
const (
	RID_KIND_SKILL    = "skill"
	RID_KIND_TOOL     = "tool"
	RID_KIND_OPERATOR = "operator"
	RID_KIND_AGENT    = "agent"
)

// ValidRidKinds contains all valid resource kinds
var ValidRidKinds = map[string]bool{
	RID_KIND_SKILL:    true,
	RID_KIND_TOOL:     true,
	RID_KIND_OPERATOR: true,
	RID_KIND_AGENT:    true,
}

// IsValidRidKind checks if the given kind is a valid rid resource kind
func IsValidRidKind(kind string) bool {
	return ValidRidKinds[kind]
}

// ValidRidKindList returns valid rid kinds in a stable order.
func ValidRidKindList() []string {
	kinds := make([]string, 0, len(ValidRidKinds))
	for kind := range ValidRidKinds {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}
