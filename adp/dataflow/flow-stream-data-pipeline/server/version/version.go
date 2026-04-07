package version

import (
	"runtime"
)

var (
	ServerName    string = "flow-stream-data-pipeline"
	ServerVersion string = "6.0.3"
	LanguageGo    string = "go"
	GoVersion     string = runtime.Version()
	GoArch        string = runtime.GOARCH
)
