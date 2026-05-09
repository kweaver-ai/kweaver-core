package apidocs

import (
	"embed"
	"io/fs"
	"net/http"
)

var (
	//go:embed assets
	embeddedAssets embed.FS

	//go:embed assets/agent-factory.json
	AgentFactoryJSON []byte

	//go:embed assets/agent-factory.yaml
	AgentFactoryYAML []byte

	//go:embed assets/agent-factory.html
	AgentFactoryHTML []byte

	//go:embed assets/agent-factory-redoc.html
	AgentFactoryRedocHTML []byte

	//go:embed assets/favicon.png
	AgentFactoryFaviconPNG []byte
)

// UIAssetsFileSystem 返回可用于 HTTP 提供本地 UI 资源的只读文件系统。
func UIAssetsFileSystem() http.FileSystem {
	subFS, err := fs.Sub(embeddedAssets, "assets/ui")
	if err != nil {
		panic(err)
	}

	return http.FS(subFS)
}
