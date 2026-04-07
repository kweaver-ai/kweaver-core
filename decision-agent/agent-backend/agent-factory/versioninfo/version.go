package versioninfo

import (
	"runtime"

	"github.com/kweaver-ai/kweaver-go-lib/audit"
)

var (
	ServerName    string = "agent-factory"
	ServerVersion string = "1.0.0"
	LanguageGo    string = "go"
	GoVersion     string = runtime.Version()
	GoArch        string = runtime.GOARCH
)

func init() {
	audit.DEFAULT_AUDIT_LOG_FROM = audit.AuditLogFrom{
		Package: "DataAgent",
		Service: audit.AuditLogFromService{
			Name: "agent-factory",
		},
	}
}
