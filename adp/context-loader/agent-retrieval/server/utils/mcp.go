// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import "fmt"

func GenerateMCPKey(mcpID string, mcpVersion int) string {
	return fmt.Sprintf("%s-%d", mcpID, mcpVersion)
}

func GenerateMCPServerVersion(mcpVersion int) string {
	return fmt.Sprintf("%d.0.0", mcpVersion)
}
