// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package version

import (
	"runtime"

	"github.com/kweaver-ai/kweaver-go-lib/audit"
)

var (
	ServerName    string = "ontology-query"
	ServerVersion string = "6.0.0"
	LanguageGo    string = "go"
	GoVersion     string = runtime.Version()
	GoArch        string = runtime.GOARCH
)

func init() {
	audit.DEFAULT_AUDIT_LOG_FROM = audit.AuditLogFrom{
		Package: "OntologyEngine",
		Service: audit.AuditLogFromService{
			Name: "ontology-query",
		},
	}
}
