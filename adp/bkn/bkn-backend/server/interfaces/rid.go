// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// RidKind defines valid resource kinds for rid properties
const (
	RID_KIND_SKILL = "skill"
)

// ValidRidKinds contains all valid resource kinds
var ValidRidKinds = map[string]bool{
	RID_KIND_SKILL: true,
}

// IsValidRidKind checks if the given kind is a valid rid resource kind
func IsValidRidKind(kind string) bool {
	return ValidRidKinds[kind]
}
