package static

import "embed"

//go:embed *
var StaticFiles embed.FS
