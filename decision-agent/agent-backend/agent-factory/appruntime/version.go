package appruntime

import (
	"runtime"
)

var (
	ServerName    string = "agent-factory"
	ServerVersion string = "1.0.0"
	LanguageGo    string = "go"
	GoVersion     string = runtime.Version()
	GoArch        string = runtime.GOARCH
)

var TraceInstrumentationName = "Opentelemetry@1.39.0/exporter"
